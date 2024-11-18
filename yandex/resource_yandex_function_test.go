package yandex

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/functions/v1"
)

const functionResource = "yandex_function.test-function"

func init() {
	resource.AddTestSweepers("yandex_function", &resource.Sweeper{
		Name: "yandex_function",
		F:    testSweepFunction,
		Dependencies: []string{
			"yandex_function_trigger",
		},
	})
}

func testSweepFunction(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &functions.ListFunctionsRequest{FolderId: conf.FolderID}
	it := conf.sdk.Serverless().Functions().Function().FunctionIterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepFunction(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Function %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepFunction(conf *Config, id string) bool {
	return sweepWithRetry(sweepFunctionOnce, conf, "Function", id)
}

func sweepFunctionOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexFunctionDefaultTimeout)
	defer cancel()

	op, err := conf.sdk.Serverless().Functions().Function().Delete(ctx, &functions.DeleteFunctionRequest{
		FunctionId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}

func TestAccYandexFunction_basic(t *testing.T) {
	t.Parallel()

	var function functions.Function
	functionName := acctest.RandomWithPrefix("tf-function")
	functionDesc := acctest.RandomWithPrefix("tf-function-desc")
	labelKey := acctest.RandomWithPrefix("tf-function-label")
	labelValue := acctest.RandomWithPrefix("tf-function-label-value")

	zipFilename := "test-fixtures/serverless/main.zip"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexFunctionDestroy,
		Steps: []resource.TestStep{
			basicYandexFunctionTestStep(functionName, functionDesc, labelKey, labelValue, zipFilename, &function),
			functionImportTestStep(),
		},
	})
}

func TestAccYandexFunction_update(t *testing.T) {
	t.Parallel()

	var function functions.Function
	functionName := acctest.RandomWithPrefix("tf-function")
	functionDesc := acctest.RandomWithPrefix("tf-function-desc")
	labelKey := acctest.RandomWithPrefix("tf-function-label")
	labelValue := acctest.RandomWithPrefix("tf-function-label-value")

	functionNameUpdated := acctest.RandomWithPrefix("tf-function-updated")
	functionDescUpdated := acctest.RandomWithPrefix("tf-function-desc-updated")
	labelKeyUpdated := acctest.RandomWithPrefix("tf-function-label-updated")
	labelValueUpdated := acctest.RandomWithPrefix("tf-function-label-value-updated")

	zipFilename := "test-fixtures/serverless/main.zip"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexFunctionDestroy,
		Steps: []resource.TestStep{
			basicYandexFunctionTestStep(functionName, functionDesc, labelKey, labelValue, zipFilename, &function),
			functionImportTestStep(),
			basicYandexFunctionTestStep(functionNameUpdated, functionDescUpdated, labelKeyUpdated, labelValueUpdated, zipFilename, &function),
			functionImportTestStep(),
		},
	})
}

func TestAccYandexFunction_updateAfterVersionCreateError(t *testing.T) {
	t.Parallel()

	var functionFirstApply functions.Function
	var functionSecondApply functions.Function
	var version *functions.Version
	resourceName := "test-function"
	resourcePath := "yandex_function." + resourceName
	functionName := acctest.RandomWithPrefix("tf-function")

	newConfig := func(executionTimeout string) string {
		sb := &strings.Builder{}
		testWriteResourceYandexFunction(
			sb,
			resourceName,
			functionName,
			"user_hash",
			128,
			"main",
			"python37",
			"test-fixtures/serverless/main.zip",
			testResourceYandexFunctionOptionFactory.WithExecutionTimeout(executionTimeout),
		)
		return sb.String()
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: newConfig("-3"),
				Check: resource.ComposeTestCheckFunc(
					testYandexFunctionExists(resourcePath, &functionFirstApply),
					testYandexFunctionNoVersionsExists(resourcePath),
				),
			},
			{
				Config: newConfig("3"),
				Check: resource.ComposeTestCheckFunc(
					testYandexFunctionExists(resourcePath, &functionSecondApply),
					func(*terraform.State) error {
						if functionFirstApply.GetId() != functionSecondApply.GetId() {
							return fmt.Errorf("Must not create new function")
						}
						return nil
					},
					testYandexFunctionVersionExists(resourcePath, &version),
				),
			},
		},
	})
}

func TestAccYandexFunction_full(t *testing.T) {
	t.Parallel()

	var function functions.Function
	params := testYandexFunctionParameters{}
	params.name = acctest.RandomWithPrefix("tf-function")
	params.desc = acctest.RandomWithPrefix("tf-function-desc")
	params.labelKey = acctest.RandomWithPrefix("tf-function-label")
	params.labelValue = acctest.RandomWithPrefix("tf-function-label-value")
	params.userHash = acctest.RandomWithPrefix("tf-function-hash")
	params.runtime = "python37"
	params.memory = "128"
	params.executionTimeout = "10"
	params.serviceAccount = acctest.RandomWithPrefix("tf-service-account")
	params.envKey = "tf_function_env"
	params.envValue = "tf_function_env_value"
	params.tags = acctest.RandomWithPrefix("tf-function-tag")
	params.secret = testSecretParameters{
		secretName:   "tf-function-secret-name",
		secretKey:    "tf-function-secret-key",
		secretEnvVar: "TF_FUNCTION_ENV_KEY",
		secretValue:  "tf-function-secret-value",
	}
	bucket := acctest.RandomWithPrefix("tf-function-test-bucket")
	params.storageMount = testStorageMountParameters{
		storageMountPointName: "mp-name",
		storageMountBucket:    bucket,
		storageMountPrefix:    "tf-function-path",
		storageMountReadOnly:  false,
	}
	params.ephemeralDiskMounts = testEphemeralDiskParameters{
		testMountParameters: testMountParameters{
			mountPoint: "mp-name-1",
			mountMode:  "rw",
		},
		ephemeralDiskSizeGB:      5,
		ephemeralDiskBlockSizeKB: 4,
	}
	params.objectStorageMounts = testObjectStorageParameters{
		testMountParameters: testMountParameters{
			mountPoint: "mp-name-2",
			mountMode:  "ro",
		},
		objectStorageBucket: bucket,
		objectStoragePrefix: "tf-function-path",
	}
	params.zipFilename = "test-fixtures/serverless/main.zip"
	params.maxAsyncRetries = "2"
	params.logOptions = testLogOptions{
		disabled: false,
		minLevel: "ERROR",
	}
	params.tmpfsSize = "0"
	params.concurrency = "2"

	paramsUpdated := testYandexFunctionParameters{}
	paramsUpdated.name = acctest.RandomWithPrefix("tf-function-updated")
	paramsUpdated.desc = acctest.RandomWithPrefix("tf-function-desc-updated")
	paramsUpdated.labelKey = acctest.RandomWithPrefix("tf-function-label-updated")
	paramsUpdated.labelValue = acctest.RandomWithPrefix("tf-function-label-value-updated")
	paramsUpdated.userHash = acctest.RandomWithPrefix("tf-function-hash-updated")
	paramsUpdated.runtime = "python27"
	paramsUpdated.memory = "2048"
	paramsUpdated.executionTimeout = "11"
	paramsUpdated.serviceAccount = acctest.RandomWithPrefix("tf-service-account")
	paramsUpdated.envKey = "tf_function_env_updated"
	paramsUpdated.envValue = "tf_function_env_value_updated"
	paramsUpdated.tags = acctest.RandomWithPrefix("tf-function-tag-updated")
	paramsUpdated.secret = testSecretParameters{
		secretName:   "tf-function-secret-name",
		secretKey:    "tf-function-secret-key-updated",
		secretEnvVar: "TF_FUNCTION_ENV_KEY_UPDATED",
		secretValue:  "tf-function-secret-value",
	}
	bucket = acctest.RandomWithPrefix("tf-function-test-bucket")
	paramsUpdated.storageMount = testStorageMountParameters{
		storageMountPointName: "mp-name-updated",
		storageMountBucket:    bucket,
		storageMountPrefix:    "tf-function-path",
		storageMountReadOnly:  false,
	}
	paramsUpdated.ephemeralDiskMounts = testEphemeralDiskParameters{
		testMountParameters: testMountParameters{
			mountPoint: "mp-name-1-updated",
			mountMode:  "rw",
		},
		ephemeralDiskSizeGB:      6,
		ephemeralDiskBlockSizeKB: 4,
	}
	paramsUpdated.objectStorageMounts = testObjectStorageParameters{
		testMountParameters: testMountParameters{
			mountPoint: "mp-name-2-updated",
			mountMode:  "ro",
		},
		objectStorageBucket: bucket,
		objectStoragePrefix: "tf-function-path",
	}
	paramsUpdated.zipFilename = "test-fixtures/serverless/main.zip"
	paramsUpdated.maxAsyncRetries = "3"
	paramsUpdated.logOptions = testLogOptions{
		disabled: false,
		minLevel: "WARN",
	}
	paramsUpdated.tmpfsSize = "1024"
	paramsUpdated.concurrency = "3"

	testConfigFunc := func(params testYandexFunctionParameters) resource.TestStep {
		return resource.TestStep{
			Config: testYandexFunctionFull(params),
			Check: resource.ComposeTestCheckFunc(
				testYandexFunctionExists(functionResource, &function),
				resource.TestCheckResourceAttr(functionResource, "name", params.name),
				resource.TestCheckResourceAttr(functionResource, "description", params.desc),
				resource.TestCheckResourceAttrSet(functionResource, "folder_id"),
				testYandexFunctionContainsLabel(&function, params.labelKey, params.labelValue),
				resource.TestCheckResourceAttr(functionResource, "user_hash", params.userHash),
				resource.TestCheckResourceAttr(functionResource, "runtime", params.runtime),
				resource.TestCheckResourceAttr(functionResource, "memory", params.memory),
				resource.TestCheckResourceAttr(functionResource, "execution_timeout", params.executionTimeout),
				resource.TestCheckResourceAttrSet(functionResource, "service_account_id"),
				testYandexFunctionContainsEnv(functionResource, params.envKey, params.envValue),
				testYandexFunctionContainsTag(functionResource, params.tags),
				resource.TestCheckResourceAttrSet(functionResource, "version"),
				resource.TestCheckResourceAttrSet(functionResource, "image_size"),
				resource.TestCheckResourceAttrSet(functionResource, "secrets.0.id"),
				resource.TestCheckResourceAttrSet(functionResource, "secrets.0.version_id"),
				resource.TestCheckResourceAttr(functionResource, "secrets.0.key", params.secret.secretKey),
				resource.TestCheckResourceAttr(functionResource, "secrets.0.environment_variable", params.secret.secretEnvVar),

				resource.TestCheckResourceAttr(functionResource, "mounts.#", "3"),

				resource.TestCheckResourceAttr(functionResource, "mounts.0.name", params.ephemeralDiskMounts.mountPoint),
				resource.TestCheckResourceAttr(functionResource, "mounts.0.mode", params.ephemeralDiskMounts.mountMode),
				resource.TestCheckResourceAttr(functionResource, "mounts.0.ephemeral_disk.0.size_gb", strconv.Itoa(params.ephemeralDiskMounts.ephemeralDiskSizeGB)),
				resource.TestCheckResourceAttr(functionResource, "mounts.0.ephemeral_disk.0.block_size_kb", strconv.Itoa(params.ephemeralDiskMounts.ephemeralDiskBlockSizeKB)),

				resource.TestCheckResourceAttr(functionResource, "mounts.1.name", params.objectStorageMounts.mountPoint),
				resource.TestCheckResourceAttr(functionResource, "mounts.1.mode", params.objectStorageMounts.mountMode),
				resource.TestCheckResourceAttr(functionResource, "mounts.1.object_storage.0.bucket", params.objectStorageMounts.objectStorageBucket),
				resource.TestCheckResourceAttr(functionResource, "mounts.1.object_storage.0.prefix", params.objectStorageMounts.objectStoragePrefix),

				resource.TestCheckResourceAttr(functionResource, "mounts.2.name", params.storageMount.storageMountPointName),
				resource.TestCheckResourceAttr(functionResource, "mounts.2.mode", modeBoolToString(params.storageMount.storageMountReadOnly)),
				resource.TestCheckResourceAttr(functionResource, "mounts.2.object_storage.0.bucket", params.storageMount.storageMountBucket),
				resource.TestCheckResourceAttr(functionResource, "mounts.2.object_storage.0.prefix", params.storageMount.storageMountPrefix),

				resource.TestCheckResourceAttr(functionResource, "async_invocation.0.retries_count", params.maxAsyncRetries),
				resource.TestCheckResourceAttr(functionResource, "log_options.0.disabled", fmt.Sprint(params.logOptions.disabled)),
				resource.TestCheckResourceAttr(functionResource, "log_options.0.min_level", params.logOptions.minLevel),
				resource.TestCheckResourceAttrSet(functionResource, "log_options.0.log_group_id"),
				resource.TestCheckResourceAttr(functionResource, "tmpfs_size", params.tmpfsSize),
				resource.TestCheckResourceAttr(functionResource, "concurrency", params.concurrency),
				testAccCheckCreatedAtAttr(functionResource),
			),
		}
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexFunctionDestroy,
		Steps: []resource.TestStep{
			testConfigFunc(params),
			functionImportTestStep(),
			testConfigFunc(paramsUpdated),
			functionImportTestStep(),
		},
	})
}

func TestAccYandexFunction_logOptions(t *testing.T) {
	t.Parallel()

	folderID := os.Getenv("YC_FOLDER_ID")
	var logGroupID string
	var function functions.Function
	var version *functions.Version
	resourceName := "test-function"
	resourcePath := "yandex_function." + resourceName
	functionName := acctest.RandomWithPrefix("tf-function-log-options")

	newConfig := func(extraOptions ...testResourceYandexFunctionOption) string {
		sb := &strings.Builder{}
		testWriteResourceYandexFunction(
			sb,
			resourceName,
			functionName,
			"user_hash",
			128,
			"main",
			"python37",
			"test-fixtures/serverless/main.zip",
			extraOptions...,
		)
		sb.WriteString(`resource "yandex_logging_group" "logging-group" {` + "\n")
		sb.WriteString(`}` + "\n")
		return sb.String()
	}

	importStep := func(extraChecks ...resource.TestCheckFunc) resource.TestStep {
		return resource.TestStep{
			ResourceName:      resourcePath,
			ImportState:       true,
			ImportStateVerify: true,
			ImportStateVerifyIgnore: []string{
				"content", "package", "image_size", "user_hash", "storage_mounts",
			},
			Check: resource.ComposeTestCheckFunc(extraChecks...),
		}
	}

	applyFunctionNoLogOptions := resource.TestStep{
		Config: newConfig(),
		Check: resource.ComposeTestCheckFunc(
			testYandexFunctionExists(resourcePath, &function),
			testYandexFunctionVersionExists(resourcePath, &version),
			resource.TestCheckResourceAttr(resourcePath, "log_options.#", "0"),
			testYandexFunctionVersionLogOptions(&version, &functions.LogOptions{
				Destination: &functions.LogOptions_FolderId{
					FolderId: folderID,
				},
			}),
		),
	}

	importFunctionNoLogOptions := importStep(
		resource.TestCheckResourceAttr(resourcePath, "log_options.#", "0"),
	)

	applyFunctionLogOptionsDisabled := resource.TestStep{
		Config: newConfig(
			testResourceYandexFunctionOptionFactory.WithLogOptions(
				true,
				"",
				"",
				"",
			),
		),
		Check: resource.ComposeTestCheckFunc(
			testYandexFunctionExists(resourcePath, &function),
			testYandexFunctionVersionExists(resourcePath, &version),
			resource.TestCheckResourceAttr(resourcePath, "log_options.#", "1"),
			resource.TestCheckResourceAttr(resourcePath, "log_options.0.disabled", "true"),
			resource.TestCheckResourceAttr(resourcePath, "log_options.0.log_group_id", ""),
			resource.TestCheckResourceAttr(resourcePath, "log_options.0.folder_id", ""),
			resource.TestCheckResourceAttr(resourcePath, "log_options.0.min_level", ""),
			testYandexFunctionVersionLogOptions(&version, &functions.LogOptions{
				Disabled: true,
				Destination: &functions.LogOptions_FolderId{
					FolderId: folderID,
				},
			}),
		),
	}

	importFunctionLogOptionsDisabled := importStep(
		resource.TestCheckResourceAttr(resourcePath, "log_options.#", "1"),
		resource.TestCheckResourceAttr(resourcePath, "log_options.0.disabled", "true"),
		resource.TestCheckResourceAttr(resourcePath, "log_options.0.log_group_id", ""),
		resource.TestCheckResourceAttr(resourcePath, "log_options.0.folder_id", ""),
		resource.TestCheckResourceAttr(resourcePath, "log_options.0.min_level", ""),
	)

	applyFunctionLogOptionsFolderID := resource.TestStep{
		Config: newConfig(
			testResourceYandexFunctionOptionFactory.WithLogOptions(
				false,
				folderID,
				"",
				"",
			),
		),
		Check: resource.ComposeTestCheckFunc(
			testYandexFunctionExists(resourcePath, &function),
			testYandexFunctionVersionExists(resourcePath, &version),
			resource.TestCheckResourceAttr(resourcePath, "log_options.#", "1"),
			resource.TestCheckResourceAttr(resourcePath, "log_options.0.disabled", "false"),
			resource.TestCheckResourceAttr(resourcePath, "log_options.0.log_group_id", ""),
			resource.TestCheckResourceAttr(resourcePath, "log_options.0.folder_id", folderID),
			resource.TestCheckResourceAttr(resourcePath, "log_options.0.min_level", ""),
			testYandexFunctionVersionLogOptions(&version, &functions.LogOptions{
				Destination: &functions.LogOptions_FolderId{
					FolderId: folderID,
				},
			}),
		),
	}

	var logOptionsWithLogGroupID *functions.LogOptions
	applyFunctionLogOptionsLogGroupID := resource.TestStep{
		Config: newConfig(
			testResourceYandexFunctionOptionFactory.WithLogOptions(
				false,
				"",
				"${yandex_logging_group.logging-group.id}",
				"",
			),
		),
		Check: resource.ComposeTestCheckFunc(
			testYandexFunctionExists(resourcePath, &function),
			testYandexFunctionVersionExists(resourcePath, &version),
			func(s *terraform.State) error {
				rs, ok := s.RootModule().Resources["yandex_logging_group.logging-group"]
				if !ok {
					return fmt.Errorf("Not found: yandex_logging_group.logging-group")
				}
				if rs.Primary.ID == "" {
					return fmt.Errorf("No ID is set")
				}
				logGroupID = rs.Primary.ID
				logOptionsWithLogGroupID = &functions.LogOptions{
					Destination: &functions.LogOptions_LogGroupId{
						LogGroupId: logGroupID,
					},
				}
				return nil
			},
			resource.TestCheckResourceAttr(resourcePath, "log_options.#", "1"),
			resource.TestCheckResourceAttr(resourcePath, "log_options.0.disabled", "false"),
			resource.TestCheckResourceAttrPtr(resourcePath, "log_options.0.log_group_id", &logGroupID),
			resource.TestCheckResourceAttr(resourcePath, "log_options.0.folder_id", ""),
			resource.TestCheckResourceAttr(resourcePath, "log_options.0.min_level", ""),
			testYandexFunctionVersionLogOptionsPtr(&version, &logOptionsWithLogGroupID),
		),
	}

	importFunctionLogOptionsLogGroupID := importStep(
		resource.TestCheckResourceAttr(resourcePath, "log_options.#", "1"),
		resource.TestCheckResourceAttr(resourcePath, "log_options.0.disabled", "false"),
		resource.TestCheckResourceAttrPtr(resourcePath, "log_options.0.log_group_id", &logGroupID),
		resource.TestCheckResourceAttr(resourcePath, "log_options.0.folder_id", ""),
		resource.TestCheckResourceAttr(resourcePath, "log_options.0.min_level", ""),
	)

	applyFunctionLogOptionsMinLevel := resource.TestStep{
		Config: newConfig(
			testResourceYandexFunctionOptionFactory.WithLogOptions(
				false,
				"",
				"",
				"ERROR"),
		),
		Check: resource.ComposeTestCheckFunc(
			testYandexFunctionExists(resourcePath, &function),
			testYandexFunctionVersionExists(resourcePath, &version),
			resource.TestCheckResourceAttr(resourcePath, "log_options.#", "1"),
			resource.TestCheckResourceAttr(resourcePath, "log_options.0.disabled", "false"),
			resource.TestCheckResourceAttr(resourcePath, "log_options.0.log_group_id", ""),
			resource.TestCheckResourceAttr(resourcePath, "log_options.0.folder_id", ""),
			resource.TestCheckResourceAttr(resourcePath, "log_options.0.min_level", "ERROR"),
			testYandexFunctionVersionLogOptions(&version, &functions.LogOptions{
				Destination: &functions.LogOptions_FolderId{
					FolderId: folderID,
				},
				MinLevel: logging.LogLevel_ERROR,
			}),
		),
	}

	importFunctionLogOptionsMinLevel := importStep(
		resource.TestCheckResourceAttr(resourcePath, "log_options.#", "1"),
		resource.TestCheckResourceAttr(resourcePath, "log_options.0.disabled", "true"),
		resource.TestCheckResourceAttr(resourcePath, "log_options.0.log_group_id", ""),
		resource.TestCheckResourceAttr(resourcePath, "log_options.0.folder_id", ""),
		resource.TestCheckResourceAttr(resourcePath, "log_options.0.min_level", "ERROR"),
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testYandexFunctionDestroy,
		Steps: []resource.TestStep{
			applyFunctionNoLogOptions,
			importFunctionNoLogOptions,
			applyFunctionLogOptionsDisabled,
			importFunctionLogOptionsDisabled,
			applyFunctionLogOptionsFolderID,
			// Can not verify import with folder id - acceptance tests designed to run within single folder,
			// therefore created function version log_options's destination will be the same as default.
			applyFunctionLogOptionsLogGroupID,
			importFunctionLogOptionsLogGroupID,
			applyFunctionLogOptionsMinLevel,
			importFunctionLogOptionsMinLevel,
			// Apply of config without log_options will return state to the beginning.
			applyFunctionNoLogOptions,
			importFunctionNoLogOptions,
		},
	})
}

func modeBoolToString(isReadOnly bool) string {
	if isReadOnly {
		return "ro"
	}
	return "rw"
}

func modeStringToBool(mode string) string {
	if mode == "ro" {
		return "true"
	}
	return "false"
}

func functionImportTestStep() resource.TestStep {
	return resource.TestStep{
		ResourceName:      "yandex_function.test-function",
		ImportState:       true,
		ImportStateVerify: true,
		ImportStateVerifyIgnore: []string{
			"content", "package", "image_size", "user_hash", "storage_mounts",
		},
	}
}

func basicYandexFunctionTestStep(functionName, functionDesc, labelKey, labelValue, zipFilename string, function *functions.Function) resource.TestStep {
	return resource.TestStep{
		Config: testYandexFunctionBasic(functionName, functionDesc, labelKey, labelValue, zipFilename),
		Check: resource.ComposeTestCheckFunc(
			testYandexFunctionExists(functionResource, function),
			resource.TestCheckResourceAttr(functionResource, "name", functionName),
			resource.TestCheckResourceAttr(functionResource, "description", functionDesc),
			resource.TestCheckResourceAttrSet(functionResource, "folder_id"),
			testYandexFunctionContainsLabel(function, labelKey, labelValue),
			testAccCheckCreatedAtAttr(functionResource),
		),
	}
}

func testYandexFunctionDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_function" {
			continue
		}

		_, err := testGetFunctionByID(config, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Function still exists")
		}
	}

	return nil
}

func testYandexFunctionExists(name string, function *functions.Function) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := testGetFunctionByID(config, rs.Primary.ID)
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Function not found")
		}

		*function = *found
		return nil
	}
}

func testGetFunctionByID(config *Config, ID string) (*functions.Function, error) {
	req := functions.GetFunctionRequest{
		FunctionId: ID,
	}

	return config.sdk.Serverless().Functions().Function().Get(context.Background(), &req)
}

func testYandexFunctionVersionExists(name string, versionPtr **functions.Version) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		primary := rs.Primary
		if primary == nil {
			return fmt.Errorf("Primary instance not found within resource %s", name)
		}

		versionID, ok := primary.Attributes["version"]
		if !ok || len(versionID) <= 0 {
			return fmt.Errorf(
				"Primary instance of resource %s does not cotain \"version\" attribute or it is empty string",
				name,
			)
		}

		config := testAccProvider.Meta().(*Config)
		version, err := testGetFunctionVersionByID(config, versionID)
		if err != nil {
			return err
		}

		if versionPtr != nil {
			*versionPtr = version
		}
		return nil
	}
}

func testGetFunctionVersionByID(config *Config, ID string) (*functions.Version, error) {
	req := functions.GetFunctionVersionRequest{
		FunctionVersionId: ID,
	}
	return config.sdk.Serverless().Functions().Function().GetVersion(context.Background(), &req)
}

func testYandexFunctionNoVersionsExists(resourcePath string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourcePath]
		if !ok {
			return fmt.Errorf("Not found: %s", resourcePath)
		}

		primary := rs.Primary
		if primary == nil {
			return fmt.Errorf("Primary instance not found within resource %s", resourcePath)
		}

		functionID := primary.ID
		if functionID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		versions, err := testListFunctionVersionsByFunctionID(config, functionID)
		if err != nil {
			return fmt.Errorf("Error while getting Yandex Function versions: %s", err.Error())
		}

		if len(versions) > 0 {
			versionsIDs := make([]string, 0, len(versions))
			for _, version := range versions {
				versionsIDs = append(versionsIDs, version.GetId())
			}
			return fmt.Errorf("Function has version(s): %s, while expected it has none", strings.Join(versionsIDs, ", "))
		}

		return nil
	}
}

func testListFunctionVersionsByFunctionID(config *Config, functionID string) ([]*functions.Version, error) {
	req := functions.ListFunctionsVersionsRequest{
		Id: &functions.ListFunctionsVersionsRequest_FunctionId{FunctionId: functionID},
	}
	resp, err := config.sdk.Serverless().Functions().Function().ListVersions(context.Background(), &req)
	if err != nil {
		return nil, err
	}
	return resp.Versions, nil
}

func testYandexFunctionContainsLabel(function *functions.Function, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		v, ok := function.Labels[key]
		if !ok {
			return fmt.Errorf("Expected label with key '%s' not found", key)
		}
		if v != value {
			return fmt.Errorf("Incorrect label value for key '%s': expected '%s' but found '%s'", key, value, v)
		}
		return nil
	}
}

func testYandexFunctionContainsEnv(name string, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resources, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found environment: %s in %s", value, s.RootModule().Path)
		}

		for k, v := range resources.Primary.Attributes {
			if strings.HasPrefix(k, "environment") && strings.Contains(k, key) && v == value {
				return nil
			}
		}

		return fmt.Errorf("Not found environment: %s, value: %s in %s", key, value, s.RootModule().Path)
	}
}

func testYandexFunctionVersionLogOptions(
	versionPtr **functions.Version,
	expected *functions.LogOptions,
) resource.TestCheckFunc {
	return testYandexFunctionVersionLogOptionsPtr(versionPtr, &expected)
}

// Same as testYandexFunctionVersionLogOptions but receives pointer that can be updated while the test is running.
func testYandexFunctionVersionLogOptionsPtr(
	versionPtr **functions.Version,
	expectedPtr **functions.LogOptions,
) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		actual := (*versionPtr).GetLogOptions()
		expected := *expectedPtr
		if assert.ObjectsAreEqual(expected, actual) {
			return nil
		}
		return fmt.Errorf("Created Function Version log options not equal to expected:\n"+
			"\nExpected:\n%s\n"+
			"\nActual:\n%s\n",
			expected.String(),
			actual.String(),
		)
	}
}

func testYandexFunctionContainsTag(name, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resources, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found tags: %s in %s", value, s.RootModule().Path)
		}

		for k, v := range resources.Primary.Attributes {
			if strings.HasPrefix(k, "tags") && v == value {
				return nil
			}
		}

		return fmt.Errorf("Not found tags: %s in %s", value, s.RootModule().Path)
	}
}

func testYandexFunctionBasic(name string, desc string, labelKey string, labelValue string, zipFileName string) string {
	return fmt.Sprintf(`
resource "yandex_function" "test-function" {
  name        = "%s"
  description = "%s"
  labels = {
    %s          = "%s"
    empty-label = ""
  }
  user_hash  = "user_hash"
  runtime    = "python37"
  entrypoint = "main"
  memory     = "128"
  content {
    zip_filename = "%s"
  }
}
	`, name, desc, labelKey, labelValue, zipFileName)
}

type testYandexFunctionParameters struct {
	name                string
	desc                string
	labelKey            string
	labelValue          string
	userHash            string
	runtime             string
	memory              string
	executionTimeout    string
	serviceAccount      string
	envKey              string
	envValue            string
	tags                string
	secret              testSecretParameters
	storageMount        testStorageMountParameters
	ephemeralDiskMounts testEphemeralDiskParameters
	objectStorageMounts testObjectStorageParameters
	zipFilename         string
	maxAsyncRetries     string
	logOptions          testLogOptions
	tmpfsSize           string
	concurrency         string
}

type testSecretParameters struct {
	secretName   string
	secretKey    string
	secretEnvVar string
	secretValue  string
}

type testStorageMountParameters struct {
	storageMountPointName string
	storageMountPointPath string
	storageMountBucket    string
	storageMountPrefix    string
	storageMountReadOnly  bool
}

type testEphemeralDiskParameters struct {
	testMountParameters
	ephemeralDiskSizeGB      int
	ephemeralDiskBlockSizeKB int
}

type testObjectStorageParameters struct {
	testMountParameters
	objectStorageBucket string
	objectStoragePrefix string
}

type testMountParameters struct {
	mountPoint string
	mountMode  string
}

type testLogOptions struct {
	minLevel string
	disabled bool
}

func testYandexFunctionFull(params testYandexFunctionParameters) string {
	return fmt.Sprintf(`
resource "yandex_function" "test-function" {
  name        = "%s"
  description = "%s"
  labels = {
    %s          = "%s"
    empty-label = ""
  }
  user_hash          = "%s"
  runtime            = "%s"
  entrypoint         = "main"
  memory             = "%s"
  execution_timeout  = "%s"
  service_account_id = "${yandex_iam_service_account.test-account.id}"
  depends_on = [
	yandex_resourcemanager_folder_iam_member.payload-viewer
  ]
  environment = {
    %s = "%s"
  }
  tags = ["%s"]
  secrets {
    id = yandex_lockbox_secret.secret.id
    version_id = yandex_lockbox_secret_version.secret_version.id
    key = "%s"
    environment_variable = "%s"
  }
  storage_mounts {
    mount_point_name = "%s"
    bucket = yandex_storage_bucket.another-bucket.bucket
    prefix = "%s"
    read_only = %v
  }
  mounts {
  	name = %q
	mode = %q
	ephemeral_disk {
		size_gb = %d
	}
  }
  mounts {
  	name = %q
	mode = %q
	object_storage {
		bucket = yandex_storage_bucket.another-bucket.bucket
		prefix = %q
	}
  }
  content {
    zip_filename = "%s"
  }
  async_invocation {
	retries_count = "%s"
  }
  log_options {
  	disabled = "%t"
	log_group_id = yandex_logging_group.logging-group.id
	min_level = "%s"
  }
  tmpfs_size = "%s"
  concurrency = "%s"
}

resource "yandex_resourcemanager_folder_iam_member" "sa-editor" {
  folder_id   = yandex_iam_service_account.test-account.folder_id
  role        = "storage.editor"
  member      = "serviceAccount:${yandex_iam_service_account.test-account.id}"
  sleep_after = 30
}

resource "yandex_iam_service_account_static_access_key" "sa-static-key" {
  depends_on = [
	yandex_resourcemanager_folder_iam_member.sa-editor,
  ]
  service_account_id = yandex_iam_service_account.test-account.id
}

resource "yandex_storage_bucket" "another-bucket" {
  access_key = yandex_iam_service_account_static_access_key.sa-static-key.access_key
  secret_key = yandex_iam_service_account_static_access_key.sa-static-key.secret_key
  bucket = "%s"
}

resource "yandex_iam_service_account" "test-account" {
  name = "%s"
}

resource "yandex_resourcemanager_folder_iam_member" "payload-viewer" {
  folder_id   = yandex_lockbox_secret.secret.folder_id
  role        = "lockbox.payloadViewer"
  member      = "serviceAccount:${yandex_iam_service_account.test-account.id}"
  sleep_after = 30
}

resource "yandex_lockbox_secret" "secret" {
  name        = "%s"
}

resource "yandex_lockbox_secret_version" "secret_version" {
  secret_id = yandex_lockbox_secret.secret.id
  entries {
    key        = "%s"
    text_value = "%s"
  }
}

resource "yandex_logging_group" "logging-group" {
}
	`,
		params.name,
		params.desc,
		params.labelKey,
		params.labelValue,
		params.userHash,
		params.runtime,
		params.memory,
		params.executionTimeout,
		params.envKey,
		params.envValue,
		params.tags,
		params.secret.secretKey,
		params.secret.secretEnvVar,
		params.storageMount.storageMountPointName,
		params.storageMount.storageMountPrefix,
		params.storageMount.storageMountReadOnly,
		params.ephemeralDiskMounts.mountPoint,
		params.ephemeralDiskMounts.mountMode,
		params.ephemeralDiskMounts.ephemeralDiskSizeGB,
		params.objectStorageMounts.mountPoint,
		params.objectStorageMounts.mountMode,
		params.objectStorageMounts.objectStoragePrefix,
		params.zipFilename,
		params.maxAsyncRetries,
		params.logOptions.disabled,
		params.logOptions.minLevel,
		params.tmpfsSize,
		params.concurrency,
		params.storageMount.storageMountBucket,
		params.serviceAccount,
		params.secret.secretName,
		params.secret.secretKey,
		params.secret.secretValue)
}

type testResourceYandexFunctionOptions struct {
	description      *string
	executionTimeout *string
	logOptions       *testResourceYandexFunctionOptionsLogOptions
}

type testResourceYandexFunctionOptionsLogOptions struct {
	disabled   bool
	folderID   string
	LogGroupID string
	minLevel   string
}

type testResourceYandexFunctionOption func(o *testResourceYandexFunctionOptions)

type testResourceYandexFunctionOptionFactoryImpl bool

const testResourceYandexFunctionOptionFactory = testResourceYandexFunctionOptionFactoryImpl(true)

func (testResourceYandexFunctionOptionFactoryImpl) WithDescription(description string) testResourceYandexFunctionOption {
	return func(o *testResourceYandexFunctionOptions) {
		o.description = &description
	}
}

func (testResourceYandexFunctionOptionFactoryImpl) WithExecutionTimeout(executionTimeout string) testResourceYandexFunctionOption {
	return func(o *testResourceYandexFunctionOptions) {
		o.executionTimeout = &executionTimeout
	}
}

func (testResourceYandexFunctionOptionFactoryImpl) WithLogOptions(
	disabled bool,
	folderID string,
	LogGroupID string,
	minLevel string,
) testResourceYandexFunctionOption {
	return func(o *testResourceYandexFunctionOptions) {
		o.logOptions = &testResourceYandexFunctionOptionsLogOptions{
			disabled:   disabled,
			folderID:   folderID,
			LogGroupID: LogGroupID,
			minLevel:   minLevel,
		}
	}
}

func testWriteResourceYandexFunction(
	sb *strings.Builder,
	resourceName string,
	functionName string,
	userHash string,
	memoryMiB uint,
	entrypoint string,
	runtime string,
	zipFilename string,
	options ...testResourceYandexFunctionOption,
) {
	var o testResourceYandexFunctionOptions
	for _, option := range options {
		option(&o)
	}

	fprintfLn := func(sb *strings.Builder, format string, a ...any) {
		_, _ = fmt.Fprintf(sb, format, a...)
		sb.WriteRune('\n')
	}

	fprintfLn(sb, "resource \"yandex_function\" \"%s\" {", resourceName)
	fprintfLn(sb, "  name = \"%s\"", functionName)
	if description := o.description; description != nil {
		fprintfLn(sb, "  description = \"%s\"", *description)
	}
	fprintfLn(sb, "  user_hash = \"%s\"", userHash)
	fprintfLn(sb, "  runtime = \"%s\"", runtime)
	fprintfLn(sb, "  entrypoint = \"%s\"", entrypoint)
	fprintfLn(sb, "  memory = \"%d\"", memoryMiB)
	if executionTimeout := o.executionTimeout; executionTimeout != nil {
		fprintfLn(sb, "  execution_timeout = \"%s\"", *executionTimeout)
	}
	fprintfLn(sb, "  content {")
	fprintfLn(sb, "    zip_filename = \"%s\"", zipFilename)
	fprintfLn(sb, "  }")
	if logOptions := o.logOptions; logOptions != nil {
		fprintfLn(sb, "  log_options {")
		if logOptions.disabled {
			fprintfLn(sb, "    disabled = true")
		}
		if logGroupID := logOptions.LogGroupID; len(logGroupID) > 0 {
			fprintfLn(sb, "    log_group_id = \"%s\"", logGroupID)
		}
		if folderID := logOptions.folderID; len(folderID) > 0 {
			fprintfLn(sb, "    folder_id = \"%s\"", folderID)
		}
		if minLevel := logOptions.minLevel; len(minLevel) > 0 {
			fprintfLn(sb, "    min_level = \"%s\"", minLevel)
		}
		fprintfLn(sb, "  }")
	}
	fprintfLn(sb, "}")
}
