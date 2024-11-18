package yandex

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/functions/v1"
)

const functionDataSource = "data.yandex_function.test-function"

func TestAccDataSourceYandexFunction_byID(t *testing.T) {
	t.Parallel()

	var function functions.Function
	functionName := acctest.RandomWithPrefix("tf-function")
	functionDesc := acctest.RandomWithPrefix("tf-function-desc")
	zipFilename := "test-fixtures/serverless/main.zip"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexFunctionByID(functionName, functionDesc, zipFilename),
				Check: resource.ComposeTestCheckFunc(
					testYandexFunctionExists(functionDataSource, &function),
					resource.TestCheckResourceAttrSet(functionDataSource, "function_id"),
					resource.TestCheckResourceAttr(functionDataSource, "name", functionName),
					resource.TestCheckResourceAttr(functionDataSource, "description", functionDesc),
					resource.TestCheckResourceAttrSet(functionDataSource, "folder_id"),
					testAccCheckCreatedAtAttr(functionDataSource),
				),
			},
		},
	})
}

func TestAccDataSourceYandexFunction_byName(t *testing.T) {
	t.Parallel()

	var function functions.Function
	functionName := acctest.RandomWithPrefix("tf-function")
	functionDesc := acctest.RandomWithPrefix("tf-function-desc")
	zipFilename := "test-fixtures/serverless/main.zip"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexFunctionByName(functionName, functionDesc, zipFilename),
				Check: resource.ComposeTestCheckFunc(
					testYandexFunctionExists(functionDataSource, &function),
					resource.TestCheckResourceAttrSet(functionDataSource, "function_id"),
					resource.TestCheckResourceAttr(functionDataSource, "name", functionName),
					resource.TestCheckResourceAttr(functionDataSource, "description", functionDesc),
					resource.TestCheckResourceAttrSet(functionDataSource, "folder_id"),
					testAccCheckCreatedAtAttr(functionDataSource),
				),
			},
		},
	})
}

func TestAccDataSourceYandexFunction_noVersion(t *testing.T) {
	t.Parallel()

	var function functions.Function
	tfName := "test-function"
	resourcePath := "yandex_function." + tfName
	dataSourcePath := "data.yandex_function." + tfName

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: func() string {
					sb := &strings.Builder{}
					testWriteResourceYandexFunction(
						sb,
						tfName,
						acctest.RandomWithPrefix("tf-function"),
						"user_hash",
						128,
						"main",
						"python37",
						"test-fixtures/serverless/main.zip",
						testResourceYandexFunctionOptionFactory.WithDescription(acctest.RandomWithPrefix("tf-function-desc")),
						testResourceYandexFunctionOptionFactory.WithExecutionTimeout("-3"),
					)
					testWriteDataSourceYandexFunction(
						sb,
						tfName,
						testDataSourceYandexFunctionOptionFactory.WithFunctionID("${"+resourcePath+".id}"),
					)
					return sb.String()
				}(),
				Check: resource.ComposeTestCheckFunc(
					// function exists
					testYandexFunctionExists(resourcePath, &function),
					// function version not exists
					testYandexFunctionNoVersionsExists(resourcePath),
					// all function attributes are set
					resource.TestCheckResourceAttrPtr(dataSourcePath, "function_id", &function.Id),
					resource.TestCheckResourceAttrPtr(dataSourcePath, "name", &function.Name),
					resource.TestCheckResourceAttrPtr(dataSourcePath, "description", &function.Description),
					resource.TestCheckResourceAttrPtr(dataSourcePath, "folder_id", &function.FolderId),
					testAccCheckCreatedAtAttr(dataSourcePath),
					// all version attributes are not set
					resource.TestCheckNoResourceAttr(dataSourcePath, "runtime"),
					resource.TestCheckNoResourceAttr(dataSourcePath, "entrypoint"),
					resource.TestCheckNoResourceAttr(dataSourcePath, "memory"),
					resource.TestCheckNoResourceAttr(dataSourcePath, "execution_timeout"),
					resource.TestCheckNoResourceAttr(dataSourcePath, "service_account_id"),
					resource.TestCheckNoResourceAttr(dataSourcePath, "environment"),
					resource.TestCheckResourceAttr(dataSourcePath, "tags.%", "0"),
					resource.TestCheckNoResourceAttr(dataSourcePath, "secrets"),
					resource.TestCheckNoResourceAttr(dataSourcePath, "storage_mounts"),
					resource.TestCheckNoResourceAttr(dataSourcePath, "mounts"),
					resource.TestCheckNoResourceAttr(dataSourcePath, "version"),
					resource.TestCheckNoResourceAttr(dataSourcePath, "connectivity"),
					resource.TestCheckNoResourceAttr(dataSourcePath, "async_invocation"),
					resource.TestCheckNoResourceAttr(dataSourcePath, "log_options"),
					resource.TestCheckNoResourceAttr(dataSourcePath, "tmpfs_size"),
					resource.TestCheckNoResourceAttr(dataSourcePath, "concurrency"),
				),
			},
		},
	})
}

func TestAccDataSourceYandexFunction_full(t *testing.T) {
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
			mountPoint: "mp-name-2",
			mountMode:  "rw",
		},
		ephemeralDiskSizeGB:      5,
		ephemeralDiskBlockSizeKB: 4,
	}
	params.objectStorageMounts = testObjectStorageParameters{
		testMountParameters: testMountParameters{
			mountPoint: "mp-name-3",
			mountMode:  "ro",
		},
		objectStorageBucket: bucket,
		objectStoragePrefix: "tf-function-path",
	}
	params.zipFilename = "test-fixtures/serverless/main.zip"
	params.maxAsyncRetries = "2"
	params.logOptions = testLogOptions{
		disabled: false,
		minLevel: "WARN",
	}
	params.tmpfsSize = "0"
	params.concurrency = "2"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexFunctionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexFunctionDataSource(params),
				Check: resource.ComposeTestCheckFunc(
					testYandexFunctionExists(functionDataSource, &function),
					resource.TestCheckResourceAttr(functionDataSource, "name", params.name),
					resource.TestCheckResourceAttr(functionDataSource, "description", params.desc),
					resource.TestCheckResourceAttrSet(functionDataSource, "folder_id"),
					testYandexFunctionContainsLabel(&function, params.labelKey, params.labelValue),
					resource.TestCheckResourceAttr(functionDataSource, "runtime", params.runtime),
					resource.TestCheckResourceAttr(functionDataSource, "memory", params.memory),
					resource.TestCheckResourceAttr(functionDataSource, "execution_timeout", params.executionTimeout),
					resource.TestCheckResourceAttrSet(functionDataSource, "service_account_id"),
					testYandexFunctionContainsEnv(functionResource, params.envKey, params.envValue),
					testYandexFunctionContainsTag(functionDataSource, params.tags),
					resource.TestCheckResourceAttrSet(functionDataSource, "version"),
					resource.TestCheckResourceAttrSet(functionDataSource, "image_size"),
					resource.TestCheckResourceAttrSet(functionDataSource, "secrets.0.id"),
					resource.TestCheckResourceAttrSet(functionDataSource, "secrets.0.version_id"),
					resource.TestCheckResourceAttr(functionDataSource, "secrets.0.key", params.secret.secretKey),
					resource.TestCheckResourceAttr(functionDataSource, "secrets.0.environment_variable", params.secret.secretEnvVar),

					resource.TestCheckResourceAttr(functionDataSource, "storage_mounts.#", "2"),

					resource.TestCheckResourceAttr(functionDataSource, "storage_mounts.1.mount_point_name", params.storageMount.storageMountPointName),
					resource.TestCheckResourceAttr(functionDataSource, "storage_mounts.1.bucket", params.storageMount.storageMountBucket),
					resource.TestCheckResourceAttr(functionDataSource, "storage_mounts.1.prefix", params.storageMount.storageMountPrefix),
					resource.TestCheckResourceAttr(functionDataSource, "storage_mounts.1.read_only", fmt.Sprint(params.storageMount.storageMountReadOnly)),

					resource.TestCheckResourceAttr(functionDataSource, "storage_mounts.0.mount_point_name", params.objectStorageMounts.mountPoint),
					resource.TestCheckResourceAttr(functionDataSource, "storage_mounts.0.bucket", params.objectStorageMounts.objectStorageBucket),
					resource.TestCheckResourceAttr(functionDataSource, "storage_mounts.0.prefix", params.objectStorageMounts.objectStoragePrefix),
					resource.TestCheckResourceAttr(functionDataSource, "storage_mounts.0.read_only", modeStringToBool(params.objectStorageMounts.mountMode)),

					resource.TestCheckResourceAttr(functionDataSource, "mounts.#", "3"),

					resource.TestCheckResourceAttr(functionDataSource, "mounts.0.name", params.ephemeralDiskMounts.mountPoint),
					resource.TestCheckResourceAttr(functionDataSource, "mounts.0.mode", params.ephemeralDiskMounts.mountMode),
					resource.TestCheckResourceAttr(functionDataSource, "mounts.0.ephemeral_disk.0.size_gb", strconv.Itoa(params.ephemeralDiskMounts.ephemeralDiskSizeGB)),
					resource.TestCheckResourceAttr(functionDataSource, "mounts.0.ephemeral_disk.0.block_size_kb", strconv.Itoa(params.ephemeralDiskMounts.ephemeralDiskBlockSizeKB)),

					resource.TestCheckResourceAttr(functionDataSource, "mounts.1.name", params.objectStorageMounts.mountPoint),
					resource.TestCheckResourceAttr(functionDataSource, "mounts.1.mode", params.objectStorageMounts.mountMode),
					resource.TestCheckResourceAttr(functionDataSource, "mounts.1.object_storage.0.bucket", params.objectStorageMounts.objectStorageBucket),
					resource.TestCheckResourceAttr(functionDataSource, "mounts.1.object_storage.0.prefix", params.objectStorageMounts.objectStoragePrefix),

					resource.TestCheckResourceAttr(functionDataSource, "mounts.2.name", params.storageMount.storageMountPointName),
					resource.TestCheckResourceAttr(functionDataSource, "mounts.2.mode", modeBoolToString(params.storageMount.storageMountReadOnly)),
					resource.TestCheckResourceAttr(functionDataSource, "mounts.2.object_storage.0.bucket", params.storageMount.storageMountBucket),
					resource.TestCheckResourceAttr(functionDataSource, "mounts.2.object_storage.0.prefix", params.storageMount.storageMountPrefix),

					resource.TestCheckResourceAttr(functionDataSource, "async_invocation.0.retries_count", params.maxAsyncRetries),
					resource.TestCheckResourceAttr(functionResource, "log_options.0.disabled", fmt.Sprint(params.logOptions.disabled)),
					resource.TestCheckResourceAttr(functionResource, "log_options.0.min_level", params.logOptions.minLevel),
					resource.TestCheckResourceAttrSet(functionResource, "log_options.0.log_group_id"),
					resource.TestCheckResourceAttr(functionDataSource, "tmpfs_size", params.tmpfsSize),
					resource.TestCheckResourceAttr(functionDataSource, "concurrency", params.concurrency),
					testAccCheckCreatedAtAttr(functionDataSource),
				),
			},
		},
	})
}

func testYandexFunctionByID(name string, desc string, zipFilename string) string {
	return fmt.Sprintf(`
data "yandex_function" "test-function" {
  function_id = "${yandex_function.test-function.id}"
}

resource "yandex_function" "test-function" {
  name        = "%s"
  description = "%s"
  user_hash   = "user_hash"
  runtime     = "python37"
  entrypoint  = "main"
  memory      = "128"
  content {
    zip_filename = "%s"
  }
}
	`, name, desc, zipFilename)
}

func testYandexFunctionByName(name string, desc string, zipFilename string) string {
	return fmt.Sprintf(`
data "yandex_function" "test-function" {
  name = "${yandex_function.test-function.name}"
}

resource "yandex_function" "test-function" {
  name        = "%s"
  description = "%s"
  user_hash   = "user_hash"
  runtime     = "python37"
  entrypoint  = "main"
  memory      = "128"
  content {
    zip_filename = "%s"
  }
}
	`, name, desc, zipFilename)
}

func testYandexFunctionDataSource(params testYandexFunctionParameters) string {
	return fmt.Sprintf(`
data "yandex_function" "test-function" {
  function_id = "${yandex_function.test-function.id}"
}

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

type testDataSourceYandexFunctionOptions struct {
	name       *string
	functionID *string
}

type testDataSourceYandexFunctionOption func(o *testDataSourceYandexFunctionOptions)

type testDataSourceYandexFunctionOptionFactoryImpl bool

const testDataSourceYandexFunctionOptionFactory = testDataSourceYandexFunctionOptionFactoryImpl(true)

func (testDataSourceYandexFunctionOptionFactoryImpl) WithName(name string) testDataSourceYandexFunctionOption {
	return func(o *testDataSourceYandexFunctionOptions) {
		o.name = &name
	}
}

func (testDataSourceYandexFunctionOptionFactoryImpl) WithFunctionID(functionID string) testDataSourceYandexFunctionOption {
	return func(o *testDataSourceYandexFunctionOptions) {
		o.functionID = &functionID
	}
}

func testWriteDataSourceYandexFunction(
	sb *strings.Builder,
	resourceName string,
	options ...testDataSourceYandexFunctionOption,
) {
	var o testDataSourceYandexFunctionOptions
	for _, option := range options {
		option(&o)
	}

	fprintfLn := func(sb *strings.Builder, format string, a ...any) {
		_, _ = fmt.Fprintf(sb, format, a...)
		sb.WriteRune('\n')
	}

	fprintfLn(sb, "data \"yandex_function\" \"%s\" {", resourceName)
	if name := o.name; name != nil {
		fprintfLn(sb, "  name = \"%s\"", *name)
	}
	if functionID := o.functionID; functionID != nil {
		fprintfLn(sb, "  function_id = \"%s\"", *functionID)
	}
	fprintfLn(sb, "}")
}
