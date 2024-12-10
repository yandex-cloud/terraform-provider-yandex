package yandex

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/c2h5oh/datasize"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/containers/v1"
)

const serverlessContainerResource = "yandex_serverless_container.test-container"
const serverlessContainerServiceAccountResource = "yandex_iam_service_account.test-account"
const serverlessContainerTestImage1 = "cr.yandex/yc/demo/coi:v1"
const serverlessContainerTestDigest1 = "sha256:e1d772fa8795adac847a2410c87d0d2e2d38fa02f118cab8c0b5fe1fb95c47f3"
const serverlessContainerTestImage2 = "cr.yandex/yc/demo/nginx-hostname:cli"
const serverlessContainerTestImage3 = "cr.yandex/mirror/library/hello-world"

func init() {
	resource.AddTestSweepers("yandex_serverless_container", &resource.Sweeper{
		Name: "yandex_serverless_container",
		F:    testSweepServerlessContainer,
		Dependencies: []string{
			"yandex_iam_service_account",
		},
	})
}

func testSweepServerlessContainer(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &containers.ListContainersRequest{FolderId: conf.FolderID}
	it := conf.sdk.Serverless().Containers().Container().ContainerIterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepServerlessContainer(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Serverless Container %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepServerlessContainer(conf *Config, id string) bool {
	return sweepWithRetry(sweepServerlessContainerOnce, conf, "Serverless Container", id)
}

func sweepServerlessContainerOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexServerlessContainerDefaultTimeout)
	defer cancel()

	op, err := conf.sdk.Serverless().Containers().Container().Delete(ctx, &containers.DeleteContainerRequest{
		ContainerId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}

func TestAccYandexServerlessContainer_basic(t *testing.T) {
	t.Parallel()

	var container containers.Container
	var revision containers.Revision
	containerName := acctest.RandomWithPrefix("tf-container")
	containerDesc := acctest.RandomWithPrefix("tf-container-desc")
	memory := (1 + acctest.RandIntRange(1, 4)) * 128

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexServerlessContainerDestroy,
		Steps: []resource.TestStep{
			basicYandexServerlessContainerTestStep(containerName, containerDesc, memory, serverlessContainerTestImage1, &container, &revision, true),
			serverlessContainerImportTestStep(),
		},
	})
}

func TestAccYandexServerlessContainer_update(t *testing.T) {
	t.Parallel()

	var container containers.Container
	var revision containers.Revision
	containerName := acctest.RandomWithPrefix("tf-container")
	containerDesc := acctest.RandomWithPrefix("tf-container-desc")
	memory := (1 + acctest.RandIntRange(1, 3)) * 128

	containerNameUpdated := acctest.RandomWithPrefix("tf-container-updated")
	containerDescUpdated := acctest.RandomWithPrefix("tf-container-desc-updated")
	memoryUpdated := (4 + acctest.RandIntRange(4, 6)) * 128

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexServerlessContainerDestroy,
		Steps: []resource.TestStep{
			// create container
			basicYandexServerlessContainerTestStep(containerName, containerDesc, memory, serverlessContainerTestImage1, &container, &revision, true),
			serverlessContainerImportTestStep(),
			// update container
			basicYandexServerlessContainerTestStep(containerNameUpdated, containerDescUpdated, memory, serverlessContainerTestImage1, &container, &revision, false),
			serverlessContainerImportTestStep(),
			// update revision
			basicYandexServerlessContainerTestStep(containerNameUpdated, containerDescUpdated, memoryUpdated, serverlessContainerTestImage2, &container, &revision, true),
			serverlessContainerImportTestStep(),
			// update container & revision
			basicYandexServerlessContainerTestStep(containerName, containerDesc, memory, serverlessContainerTestImage1, &container, &revision, true),
			serverlessContainerImportTestStep(),
		},
	})
}

func serverlessContainerImportTestStep() resource.TestStep {
	return resource.TestStep{
		ResourceName:            serverlessContainerResource,
		ImportState:             true,
		ImportStateVerify:       true,
		ImportStateVerifyIgnore: []string{"storage_mounts"},
	}
}

func basicYandexServerlessContainerTestStep(containerName string, containerDesc string, memory int, image string, container *containers.Container, revision *containers.Revision, revisionChanged bool) resource.TestStep {
	var newRevision containers.Revision
	return resource.TestStep{
		Config: testYandexServerlessContainerBasic(containerName, containerDesc, memory, image),
		Check: resource.ComposeTestCheckFunc(
			// container
			testYandexServerlessContainerExists(serverlessContainerResource, container),
			resource.TestCheckResourceAttr(serverlessContainerResource, "name", containerName),
			resource.TestCheckResourceAttr(serverlessContainerResource, "description", containerDesc),
			// revision
			resource.TestCheckResourceAttrSet(serverlessContainerResource, "revision_id"),
			testYandexServerlessContainerRevisionExists(serverlessContainerResource, &newRevision),
			testYandexServerlessContainerRevisionMemory(&newRevision, memory),
			testYandexServerlessContainerRevisionChanged(revision, &newRevision, revisionChanged),
			// metadata
			resource.TestCheckResourceAttrSet(serverlessContainerResource, "folder_id"),
			resource.TestCheckResourceAttrSet(serverlessContainerResource, "url"),
			testAccCheckCreatedAtAttr(serverlessContainerResource),
		),
	}
}

func TestAccYandexServerlessContainer_updateAfterRevisionDeployError(t *testing.T) {
	t.Parallel()

	var containerFirstApply containers.Container
	var containerSecondApply containers.Container
	var revision containers.Revision
	resourceName := "test-container"
	resourcePath := "yandex_serverless_container." + resourceName
	containerName := acctest.RandomWithPrefix("tf-container")

	newConfig := func(options ...testResourceYandexServerlessContainerOption) string {
		sb := &strings.Builder{}
		testWriteResourceYandexServerlessContainer(
			sb,
			resourceName,
			containerName,
			128,
			serverlessContainerTestImage1,
			options...,
		)
		return sb.String()
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: newConfig(
					testResourceYandexServerlessContainerOptionFactory.WithServiceAccountID("non-existent"),
				),
				Check: resource.ComposeTestCheckFunc(
					testYandexServerlessContainerExists(resourcePath, &containerFirstApply),
					testYandexServerlessContainerRevisionNotExists(resourcePath),
				),
			},
			{
				Config: newConfig(),
				Check: resource.ComposeTestCheckFunc(
					testYandexServerlessContainerExists(resourcePath, &containerSecondApply),
					func(*terraform.State) error {
						if containerFirstApply.GetId() != containerSecondApply.GetId() {
							return fmt.Errorf("Must not create new container")
						}
						return nil
					},
					testYandexServerlessContainerRevisionExists(resourcePath, &revision),
				),
			},
		},
	})
}

func TestAccYandexServerlessContainer_full(t *testing.T) {
	t.Parallel()

	var container containers.Container
	var revision containers.Revision
	params := testYandexServerlessContainerParameters{}
	params.name = acctest.RandomWithPrefix("tf-container")
	params.desc = acctest.RandomWithPrefix("tf-container-desc")
	params.labelKey = acctest.RandomWithPrefix("tf-container-label")
	params.labelValue = acctest.RandomWithPrefix("tf-container-label-value")
	params.memory = (1 + acctest.RandIntRange(1, 3)) * 128
	params.cores = 1
	params.coreFraction = 100
	params.executionTimeout = strconv.FormatInt(int64(1+acctest.RandIntRange(1, 10)), 10) + "s"
	params.concurrency = acctest.RandIntRange(1, 3) + 1
	params.runtime = "http"
	params.imageURL = serverlessContainerTestImage1
	params.workDir = acctest.RandomWithPrefix("tf-container-work-dir")
	params.command = acctest.RandomWithPrefix("tf-container-command")
	params.argument = acctest.RandomWithPrefix("tf-container-argument")
	params.envVarKey = "env_var_key"
	params.envVarValue = acctest.RandomWithPrefix("tf-container-env-value")
	params.serviceAccount = acctest.RandomWithPrefix("tf-container-sa")
	params.secret = testSecretParameters{
		secretName:   "tf-container-secret-name",
		secretKey:    "tf-container-secret-key",
		secretEnvVar: "TF_CONTAINER_ENV_KEY",
		secretValue:  "tf-container-secret-value",
	}
	bucket := acctest.RandomWithPrefix("tf-function-test-bucket")
	params.storageMount = testStorageMountParameters{
		storageMountPointPath: "/mount/point/a",
		storageMountBucket:    bucket,
		storageMountPrefix:    "tf-container-path",
		storageMountReadOnly:  false,
	}
	params.ephemeralDiskMounts = testEphemeralDiskParameters{
		testMountParameters: testMountParameters{
			mountPoint: "/mount/point/b",
			mountMode:  "rw",
		},
		ephemeralDiskSizeGB:      5,
		ephemeralDiskBlockSizeKB: 4,
	}
	params.objectStorageMounts = testObjectStorageParameters{
		testMountParameters: testMountParameters{
			mountPoint: "/mount/point/c",
			mountMode:  "ro",
		},
		objectStorageBucket: bucket,
		objectStoragePrefix: "tf-function-path",
	}

	params.logOptions = testLogOptions{
		disabled: false,
		minLevel: "ERROR",
	}

	paramsUpdated := testYandexServerlessContainerParameters{}
	paramsUpdated.name = acctest.RandomWithPrefix("tf-container-updated")
	paramsUpdated.desc = acctest.RandomWithPrefix("tf-container-desc-updated")
	paramsUpdated.labelKey = acctest.RandomWithPrefix("tf-container-label-updated")
	paramsUpdated.labelValue = acctest.RandomWithPrefix("tf-container-label-value-updated")
	paramsUpdated.memory = (4 + acctest.RandIntRange(4, 6)) * 128
	paramsUpdated.cores = 1
	paramsUpdated.coreFraction = 100
	paramsUpdated.executionTimeout = strconv.FormatInt(int64(11+acctest.RandIntRange(11, 20)), 10) + "s"
	paramsUpdated.concurrency = 1
	paramsUpdated.runtime = "task"
	paramsUpdated.imageURL = serverlessContainerTestImage3
	paramsUpdated.workDir = acctest.RandomWithPrefix("tf-container-work-dir-updated")
	paramsUpdated.command = acctest.RandomWithPrefix("tf-container-command-updated")
	paramsUpdated.argument = acctest.RandomWithPrefix("tf-container-argument-updated")
	paramsUpdated.envVarKey = "env_var_key"
	paramsUpdated.envVarValue = acctest.RandomWithPrefix("tf-container-env-value-updated")
	paramsUpdated.serviceAccount = acctest.RandomWithPrefix("tf-container-sa-updated")
	paramsUpdated.secret = testSecretParameters{
		secretName:   "tf-container-secret-name-updated",
		secretKey:    "tf-container-secret-key-updated",
		secretEnvVar: "TF_CONTAINER_ENV_KEY_UPDATED",
		secretValue:  "tf-container-secret-value-updated",
	}

	bucket = acctest.RandomWithPrefix("tf-function-test-bucket-updated")
	paramsUpdated.storageMount = testStorageMountParameters{
		storageMountPointPath: "/mount/point/a-a",
		storageMountBucket:    bucket,
		storageMountPrefix:    "tf-container-path-updated",
		storageMountReadOnly:  true,
	}
	paramsUpdated.ephemeralDiskMounts = testEphemeralDiskParameters{
		testMountParameters: testMountParameters{
			mountPoint: "/mount/point/b-b",
			mountMode:  "rw",
		},
		ephemeralDiskSizeGB:      10,
		ephemeralDiskBlockSizeKB: 4,
	}
	paramsUpdated.objectStorageMounts = testObjectStorageParameters{
		testMountParameters: testMountParameters{
			mountPoint: "/mount/point/c-c",
			mountMode:  "ro",
		},
		objectStorageBucket: bucket,
		objectStoragePrefix: "tf-function-path",
	}
	paramsUpdated.logOptions = testLogOptions{
		disabled: false,
		minLevel: "WARN",
	}

	testConfigFunc := func(params testYandexServerlessContainerParameters) resource.TestStep {
		return resource.TestStep{
			Config: testYandexServerlessContainerFull(params),
			Check: resource.ComposeTestCheckFunc(
				// container
				testYandexServerlessContainerExists(serverlessContainerResource, &container),
				testYandexServerlessContainerName(&container, params.name),
				testYandexServerlessContainerDescription(&container, params.desc),
				testYandexServerlessContainerContainsLabel(&container, params.labelKey, params.labelValue),
				// revision
				resource.TestCheckResourceAttrSet(serverlessContainerResource, "revision_id"),
				testYandexServerlessContainerRevisionExists(serverlessContainerResource, &revision),
				testYandexServerlessContainerRevisionMemory(&revision, params.memory),
				testYandexServerlessContainerRevisionCores(&revision, params.cores, params.coreFraction),
				testYandexServerlessContainerRevisionExecutionTimeout(&revision, params.executionTimeout),
				testYandexServerlessContainerRevisionConcurrency(&revision, params.concurrency),
				testYandexServerlessContainerRevisionImage(&revision, params),
				testYandexServerlessContainerRevisionServiceAccount(&revision, serverlessContainerServiceAccountResource),
				testYandexServerlessContainerRevisionRuntime(&revision, params.runtime),
				resource.TestCheckResourceAttrSet(serverlessContainerResource, "secrets.0.id"),
				resource.TestCheckResourceAttrSet(serverlessContainerResource, "secrets.0.version_id"),
				resource.TestCheckResourceAttr(serverlessContainerResource, "secrets.0.key", params.secret.secretKey),
				resource.TestCheckResourceAttr(serverlessContainerResource, "secrets.0.environment_variable", params.secret.secretEnvVar),

				resource.TestCheckResourceAttr(serverlessContainerResource, "mounts.#", "3"),

				resource.TestCheckResourceAttr(serverlessContainerResource, "mounts.0.mount_point_path", params.ephemeralDiskMounts.mountPoint),
				resource.TestCheckResourceAttr(serverlessContainerResource, "mounts.0.mode", params.ephemeralDiskMounts.mountMode),
				resource.TestCheckResourceAttr(serverlessContainerResource, "mounts.0.ephemeral_disk.0.size_gb", strconv.Itoa(params.ephemeralDiskMounts.ephemeralDiskSizeGB)),
				resource.TestCheckResourceAttr(serverlessContainerResource, "mounts.0.ephemeral_disk.0.block_size_kb", strconv.Itoa(params.ephemeralDiskMounts.ephemeralDiskBlockSizeKB)),

				resource.TestCheckResourceAttr(serverlessContainerResource, "mounts.1.mount_point_path", params.objectStorageMounts.mountPoint),
				resource.TestCheckResourceAttr(serverlessContainerResource, "mounts.1.mode", params.objectStorageMounts.mountMode),
				resource.TestCheckResourceAttr(serverlessContainerResource, "mounts.1.object_storage.0.bucket", params.objectStorageMounts.objectStorageBucket),
				resource.TestCheckResourceAttr(serverlessContainerResource, "mounts.1.object_storage.0.prefix", params.objectStorageMounts.objectStoragePrefix),

				resource.TestCheckResourceAttr(serverlessContainerResource, "mounts.2.mount_point_path", params.storageMount.storageMountPointPath),
				resource.TestCheckResourceAttr(serverlessContainerResource, "mounts.2.mode", modeBoolToString(params.storageMount.storageMountReadOnly)),
				resource.TestCheckResourceAttr(serverlessContainerResource, "mounts.2.object_storage.0.bucket", params.storageMount.storageMountBucket),
				resource.TestCheckResourceAttr(serverlessContainerResource, "mounts.2.object_storage.0.prefix", params.storageMount.storageMountPrefix),

				resource.TestCheckResourceAttr(serverlessContainerResource, "log_options.0.disabled", fmt.Sprint(params.logOptions.disabled)),
				resource.TestCheckResourceAttr(serverlessContainerResource, "log_options.0.min_level", params.logOptions.minLevel),
				resource.TestCheckResourceAttrSet(serverlessContainerResource, "log_options.0.log_group_id"),
				// metadata
				resource.TestCheckResourceAttrSet(serverlessContainerResource, "folder_id"),
				resource.TestCheckResourceAttrSet(serverlessContainerResource, "url"),
				testAccCheckCreatedAtAttr(serverlessContainerResource),
			),
		}
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexServerlessContainerDestroy,
		Steps: []resource.TestStep{
			testConfigFunc(params),
			serverlessContainerImportTestStep(),
			testConfigFunc(paramsUpdated),
			serverlessContainerImportTestStep(),
		},
	})
}

func TestAccYandexServerlessContainer_logOptions(t *testing.T) {
	t.Parallel()

	folderID := os.Getenv("YC_FOLDER_ID")
	var container containers.Container
	var revision containers.Revision
	var logOptionsWithLogGroupID *containers.LogOptions
	var logGroupID string
	name := acctest.RandomWithPrefix("tf-serverless-container-log-options")
	resourceName := "test-container"
	resourcePath := "yandex_serverless_container." + resourceName

	newConfig := func(extraOptions ...testResourceYandexServerlessContainerOption) string {
		sb := &strings.Builder{}
		testWriteResourceYandexServerlessContainer(
			sb,
			resourceName,
			name,
			128,
			serverlessContainerTestImage1,
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

	applyServerlessContainerNoLogOptions := resource.TestStep{
		Config: newConfig(),
		Check: resource.ComposeTestCheckFunc(
			testYandexServerlessContainerExists(resourcePath, &container),
			testYandexServerlessContainerRevisionExists(resourcePath, &revision),
			resource.TestCheckResourceAttr(resourcePath, "log_options.#", "0"),
			testYandexServerlessContainerRevisionLogOptions(&revision, &containers.LogOptions{
				Destination: &containers.LogOptions_FolderId{
					FolderId: folderID,
				},
			}),
		),
	}

	importServerlessContainerNoLogOptions := importStep(
		resource.TestCheckResourceAttr(resourcePath, "log_options.#", "0"),
	)

	applyServerlessContainerLogOptionsDisabled := resource.TestStep{
		Config: newConfig(
			testResourceYandexServerlessContainerOptionFactory.WithLogOptions(
				true,
				"",
				"",
				"",
			),
		),
		Check: resource.ComposeTestCheckFunc(
			testYandexServerlessContainerExists(resourcePath, &container),
			testYandexServerlessContainerRevisionExists(resourcePath, &revision),
			resource.TestCheckResourceAttr(resourcePath, "log_options.#", "1"),
			resource.TestCheckResourceAttr(resourcePath, "log_options.0.disabled", "true"),
			resource.TestCheckResourceAttr(resourcePath, "log_options.0.log_group_id", ""),
			resource.TestCheckResourceAttr(resourcePath, "log_options.0.folder_id", ""),
			resource.TestCheckResourceAttr(resourcePath, "log_options.0.min_level", ""),
			testYandexServerlessContainerRevisionLogOptions(&revision, &containers.LogOptions{
				Disabled: true,
				Destination: &containers.LogOptions_FolderId{
					FolderId: folderID,
				},
			}),
		),
	}

	importServerlessContainerLogOptionsDisabled := importStep(
		resource.TestCheckResourceAttr(resourcePath, "log_options.#", "1"),
		resource.TestCheckResourceAttr(resourcePath, "log_options.0.disabled", "true"),
		resource.TestCheckResourceAttr(resourcePath, "log_options.0.log_group_id", ""),
		resource.TestCheckResourceAttr(resourcePath, "log_options.0.folder_id", ""),
		resource.TestCheckResourceAttr(resourcePath, "log_options.0.min_level", ""),
	)

	applyServerlessContainerLogOptionsFolderID := resource.TestStep{
		Config: newConfig(
			testResourceYandexServerlessContainerOptionFactory.WithLogOptions(
				false,
				folderID,
				"",
				"",
			),
		),
		Check: resource.ComposeTestCheckFunc(
			testYandexServerlessContainerExists(resourcePath, &container),
			testYandexServerlessContainerRevisionExists(resourcePath, &revision),
			resource.TestCheckResourceAttr(resourcePath, "log_options.#", "1"),
			resource.TestCheckResourceAttr(resourcePath, "log_options.0.disabled", "false"),
			resource.TestCheckResourceAttr(resourcePath, "log_options.0.log_group_id", ""),
			resource.TestCheckResourceAttr(resourcePath, "log_options.0.folder_id", folderID),
			resource.TestCheckResourceAttr(resourcePath, "log_options.0.min_level", ""),
			testYandexServerlessContainerRevisionLogOptions(&revision, &containers.LogOptions{
				Destination: &containers.LogOptions_FolderId{
					FolderId: folderID,
				},
			}),
		),
	}

	applyServerlessContainerLogOptionsLogGroupID := resource.TestStep{
		Config: newConfig(
			testResourceYandexServerlessContainerOptionFactory.WithLogOptions(
				false,
				"",
				"${yandex_logging_group.logging-group.id}",
				"",
			),
		),
		Check: resource.ComposeTestCheckFunc(
			testYandexServerlessContainerExists(resourcePath, &container),
			testYandexServerlessContainerRevisionExists(resourcePath, &revision),
			func(s *terraform.State) error {
				rs, ok := s.RootModule().Resources["yandex_logging_group.logging-group"]
				if !ok {
					return fmt.Errorf("Not found: %s", name)
				}
				if rs.Primary.ID == "" {
					return fmt.Errorf("No ID is set")
				}
				logGroupID = rs.Primary.ID
				logOptionsWithLogGroupID = &containers.LogOptions{
					Destination: &containers.LogOptions_LogGroupId{
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
			testYandexServerlessContainerRevisionLogOptionsPtr(&revision, &logOptionsWithLogGroupID),
		),
	}

	importServerlessContainerLogOptionsLogGroupID := importStep(
		resource.TestCheckResourceAttr(resourcePath, "log_options.#", "1"),
		resource.TestCheckResourceAttr(resourcePath, "log_options.0.disabled", "false"),
		resource.TestCheckResourceAttrPtr(resourcePath, "log_options.0.log_group_id", &logGroupID),
		resource.TestCheckResourceAttr(resourcePath, "log_options.0.folder_id", ""),
		resource.TestCheckResourceAttr(resourcePath, "log_options.0.min_level", ""),
	)

	applyServerlessContainerLogOptionsMinLevel := resource.TestStep{
		Config: newConfig(
			testResourceYandexServerlessContainerOptionFactory.WithLogOptions(
				false,
				"",
				"",
				"ERROR"),
		),
		Check: resource.ComposeTestCheckFunc(
			testYandexServerlessContainerExists(resourcePath, &container),
			testYandexServerlessContainerRevisionExists(resourcePath, &revision),
			resource.TestCheckResourceAttr(resourcePath, "log_options.#", "1"),
			resource.TestCheckResourceAttr(resourcePath, "log_options.0.disabled", "false"),
			resource.TestCheckResourceAttr(resourcePath, "log_options.0.log_group_id", ""),
			resource.TestCheckResourceAttr(resourcePath, "log_options.0.folder_id", ""),
			resource.TestCheckResourceAttr(resourcePath, "log_options.0.min_level", "ERROR"),
			testYandexServerlessContainerRevisionLogOptions(&revision, &containers.LogOptions{
				Destination: &containers.LogOptions_FolderId{
					FolderId: folderID,
				},
				MinLevel: logging.LogLevel_ERROR,
			}),
		),
	}

	importServerlessContainerLogOptionsMinLevel := importStep(
		resource.TestCheckResourceAttr(resourcePath, "log_options.#", "1"),
		resource.TestCheckResourceAttr(resourcePath, "log_options.0.disabled", "true"),
		resource.TestCheckResourceAttr(resourcePath, "log_options.0.log_group_id", ""),
		resource.TestCheckResourceAttr(resourcePath, "log_options.0.folder_id", ""),
		resource.TestCheckResourceAttr(resourcePath, "log_options.0.min_level", "ERROR"),
	)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testYandexServerlessContainerDestroy,
		Steps: []resource.TestStep{
			applyServerlessContainerNoLogOptions,
			importServerlessContainerNoLogOptions,
			applyServerlessContainerLogOptionsDisabled,
			importServerlessContainerLogOptionsDisabled,
			applyServerlessContainerLogOptionsFolderID,
			// Can not verify import with folder id - acceptance tests designed to run within single folder,
			// therefore created serverless container revision log_options's destination will be the same as default.
			applyServerlessContainerLogOptionsLogGroupID,
			importServerlessContainerLogOptionsLogGroupID,
			applyServerlessContainerLogOptionsMinLevel,
			importServerlessContainerLogOptionsMinLevel,
			// Apply of config without log_options will return state to the beginning.
			applyServerlessContainerNoLogOptions,
			importServerlessContainerNoLogOptions,
		},
	})
}

func testYandexServerlessContainerDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_serverless_container" {
			continue
		}

		_, err := testGetServerlessContainerByID(config, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Serverless Container still exists")
		}
	}

	return nil
}

func testYandexServerlessContainerExists(name string, container *containers.Container) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := testGetServerlessContainerByID(config, rs.Primary.ID)
		if err != nil {
			return err
		}

		if found.GetId() != rs.Primary.ID {
			return fmt.Errorf("Serverless Container not found")
		}

		*container = *found
		return nil
	}
}

func testYandexServerlessContainerRevisionExists(name string, revision *containers.Revision) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		config := testAccProvider.Meta().(*Config)
		revisionID := rs.Primary.Attributes["revision_id"]

		found, err := testGetServerlessContainerRevisionByID(config, revisionID)
		if err != nil {
			return err
		}

		if found.GetId() != revisionID {
			return fmt.Errorf("Serverless Container Revision not found")
		}

		*revision = *found
		return nil
	}
}

func testYandexServerlessContainerRevisionNotExists(resourcePath string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourcePath]
		if !ok {
			return fmt.Errorf("Not found: %s", resourcePath)
		}

		primary := rs.Primary
		if primary == nil {
			return fmt.Errorf("Primary instance not found within resource %s", resourcePath)
		}

		containerID := primary.ID
		if containerID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		revisions, err := testListContainerRevisionsByContainerID(config, containerID)
		if err != nil {
			return fmt.Errorf("Error while getting Yandex Container Revisions: %s", err.Error())
		}

		if len(revisions) > 0 {
			revisionsIDs := make([]string, 0, len(revisions))
			for _, version := range revisions {
				revisionsIDs = append(revisionsIDs, version.GetId())
			}
			return fmt.Errorf("Container has revision(s): %s, while expected it has none", strings.Join(revisionsIDs, ", "))
		}

		return nil
	}
}

func testListContainerRevisionsByContainerID(config *Config, containerID string) ([]*containers.Revision, error) {
	req := containers.ListContainersRevisionsRequest{
		Id: &containers.ListContainersRevisionsRequest_ContainerId{ContainerId: containerID},
	}
	resp, err := config.sdk.Serverless().Containers().Container().ListRevisions(context.Background(), &req)
	if err != nil {
		return nil, err
	}
	return resp.Revisions, nil
}

func testGetServerlessContainerByID(config *Config, ID string) (*containers.Container, error) {
	req := containers.GetContainerRequest{
		ContainerId: ID,
	}

	return config.sdk.Serverless().Containers().Container().Get(context.Background(), &req)
}

func testGetServerlessContainerRevisionByID(config *Config, ID string) (*containers.Revision, error) {
	req := containers.GetContainerRevisionRequest{
		ContainerRevisionId: ID,
	}

	return config.sdk.Serverless().Containers().Container().GetRevision(context.Background(), &req)
}

func testYandexServerlessContainerName(container *containers.Container, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if container.Name != name {
			return fmt.Errorf("Incorrect container name: expected '%s' but found '%s'", name, container.Name)
		}
		return nil
	}
}

func testYandexServerlessContainerDescription(container *containers.Container, description string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if container.Description != description {
			return fmt.Errorf("Incorrect container description: expected '%s' but found '%s'", description, container.Name)
		}
		return nil
	}
}

func testYandexServerlessContainerRevisionMemory(revision *containers.Revision, memory int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		expected := int64(int(datasize.MB.Bytes()) * memory)

		if expected != revision.Resources.Memory {
			return fmt.Errorf("Incorrect revision memory: expected '%d' but found '%d'", expected, revision.Resources.Memory)
		}
		return nil
	}
}

func testYandexServerlessContainerRevisionChanged(oldRevision *containers.Revision, newRevision *containers.Revision, revisionChanged bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if revisionChanged && oldRevision.GetId() == newRevision.GetId() {
			return fmt.Errorf("Missing revision update")
		}
		if !revisionChanged && oldRevision.GetId() != newRevision.GetId() {
			return fmt.Errorf("Unexpected revision update: expected revision '%s' but found '%s'", oldRevision.GetId(), newRevision.GetId())
		}

		*oldRevision = *newRevision
		return nil
	}
}

func testYandexServerlessContainerRevisionExecutionTimeout(revision *containers.Revision, executionTimeout string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		expected, _ := parseDuration(executionTimeout)

		if revision.ExecutionTimeout.AsDuration() != expected.AsDuration() {
			return fmt.Errorf("Incorrect execution timeout: expected '%s' but found '%s'",
				expected.AsDuration(), revision.ExecutionTimeout.AsDuration())
		}
		return nil
	}
}

func testYandexServerlessContainerRevisionConcurrency(revision *containers.Revision, concurrency int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if revision.Concurrency != int64(concurrency) {
			return fmt.Errorf("Incorrect concurrency: expected '%d' but found '%d'", concurrency, revision.Concurrency)
		}
		return nil
	}
}

func testYandexServerlessContainerRevisionCores(revision *containers.Revision, cores int, coreFraction int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if revision.Resources.Cores != int64(cores) {
			return fmt.Errorf("Incorrect cores: expected '%d' but found '%d'", cores, revision.Resources.Cores)
		}
		if revision.Resources.CoreFraction != int64(coreFraction) {
			return fmt.Errorf("Incorrect core fraction: expected '%d' but found '%d'", coreFraction, revision.Resources.CoreFraction)
		}
		return nil
	}
}

func testYandexServerlessContainerContainsLabel(container *containers.Container, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		v, ok := container.Labels[key]
		if !ok {
			return fmt.Errorf("Expected label with key '%s' not found", key)
		}
		if v != value {
			return fmt.Errorf("Incorrect label value for key '%s': expected '%s' but found '%s'", key, value, v)
		}
		return nil
	}
}

func testYandexServerlessContainerRevisionImage(revision *containers.Revision, params testYandexServerlessContainerParameters) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if revision.GetImage().GetImageUrl() != params.imageURL {
			return fmt.Errorf("Incorrect image url: expected '%s' but found '%s'", params.imageURL, revision.GetImage().GetImageUrl())
		}
		if revision.GetImage().GetWorkingDir() != params.workDir {
			return fmt.Errorf("Incorrect work dir: expected '%s' but found '%s'", params.workDir, revision.GetImage().GetWorkingDir())
		}
		args := revision.GetImage().GetArgs().GetArgs()
		if len(args) != 1 {
			return fmt.Errorf("Incorrect amount of image arguments: expected '%d'", 1)
		}
		if args[0] != params.argument {
			return fmt.Errorf("Incorrect image argment: expected '%s' but found '%s'", params.argument, args[0])
		}
		commands := revision.GetImage().GetCommand().GetCommand()
		if len(commands) != 1 {
			return fmt.Errorf("Incorrect amount of image commands: expected '%d'", 1)
		}
		if commands[0] != params.command {
			return fmt.Errorf("Incorrect image command: expected '%s' but found '%s'", params.command, commands[0])
		}
		environments := revision.GetImage().GetEnvironment()
		if len(environments) != 1 {
			return fmt.Errorf("Incorrect amount of image environments: expected '%d'", 1)
		}
		if environments[params.envVarKey] != params.envVarValue {
			return fmt.Errorf("Incorrect image environment '%s': expected '%s' but found '%s'",
				params.envVarKey, params.envVarValue, environments[params.envVarKey])
		}
		return nil
	}
}

func testYandexServerlessContainerRevisionServiceAccount(revision *containers.Revision, serviceAccountResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		sa, ok := s.RootModule().Resources[serviceAccountResource]
		if !ok {
			return fmt.Errorf("Not found service account: %s", serverlessContainerResource)
		}
		if revision.ServiceAccountId != sa.Primary.ID {
			return fmt.Errorf("Incorrect service account: expected '%s' but found '%s'", sa.Primary.ID, revision.ServiceAccountId)
		}
		return nil
	}
}

func testYandexServerlessContainerRevisionRuntime(revision *containers.Revision, runtime string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rt := revision.GetRuntime()
		if rt == nil {
			return fmt.Errorf("Incorrect runtime: expected '%s' but found nil", runtime)
		}
		switch runtime {
		case "http":
			if rt.GetHttp() == nil {
				return fmt.Errorf("Incorrect runtime: expected 'http' but found '%s'", rt.String())
			}
		case "task":
			if rt.GetTask() == nil {
				return fmt.Errorf("Incorrect runtime: expected 'task' but found '%s'", rt.String())
			}
		}
		return nil
	}
}

func testYandexServerlessContainerRevisionLogOptions(
	revision *containers.Revision,
	expected *containers.LogOptions,
) resource.TestCheckFunc {
	return testYandexServerlessContainerRevisionLogOptionsPtr(revision, &expected)
}

// Same as testYandexServerlessContainerRevisionLogOptions but receives pointer that can be updated while the test is running.
func testYandexServerlessContainerRevisionLogOptionsPtr(
	revision *containers.Revision,
	expectedPtr **containers.LogOptions,
) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		actual := revision.GetLogOptions()
		expected := *expectedPtr
		if assert.ObjectsAreEqual(expected, actual) {
			return nil
		}
		return fmt.Errorf("Created Container Revision log options not equal to expected:\n"+
			"\nExpected:\n%s\n"+
			"\nActual:\n%s\n",
			expected.String(),
			actual.String(),
		)
	}
}

func testYandexServerlessContainerBasic(name string, desc string, memory int, image string) string {
	return fmt.Sprintf(`
resource "yandex_serverless_container" "test-container" {
  name        = "%s"
  description = "%s"
  memory      = %d
  image {
    url = "%s"
  }
}
	`, name, desc, memory, image)
}

type testYandexServerlessContainerParameters struct {
	name                string
	desc                string
	labelKey            string
	labelValue          string
	memory              int
	cores               int
	coreFraction        int
	executionTimeout    string
	concurrency         int
	runtime             string
	imageURL            string
	workDir             string
	command             string
	argument            string
	envVarKey           string
	envVarValue         string
	serviceAccount      string
	secret              testSecretParameters
	storageMount        testStorageMountParameters
	ephemeralDiskMounts testEphemeralDiskParameters
	objectStorageMounts testObjectStorageParameters
	logOptions          testLogOptions
}

func testYandexServerlessContainerFull(params testYandexServerlessContainerParameters) string {
	return fmt.Sprintf(`
resource "yandex_serverless_container" "test-container" {
  name        = "%s"
  description = "%s"
  labels = {
    %s          = "%s"
    empty-label = ""
  }
  memory             = %d
  cores              = %d
  core_fraction      = %d
  execution_timeout  = "%s"
  concurrency        = %d
  runtime {
    type = "%s"
  }
  service_account_id = "${yandex_iam_service_account.test-account.id}"
  depends_on = [
	yandex_resourcemanager_folder_iam_member.payload-viewer,
    yandex_resourcemanager_folder_iam_member.sa-editor
  ]
  secrets {
    id = yandex_lockbox_secret.secret.id
    version_id = yandex_lockbox_secret_version.secret_version.id
    key = "%s"
    environment_variable = "%s"
  }

  storage_mounts {
    mount_point_path = "%s"
    bucket = yandex_storage_bucket.another-bucket.bucket
    prefix = "%s"
    read_only = %v
  }
  mounts {
  	mount_point_path = %q
	mode = %q
	ephemeral_disk {
		size_gb = %d
	}
  }
  mounts {
  	mount_point_path = %q
	mode = %q
	object_storage {
		bucket = yandex_storage_bucket.another-bucket.bucket
		prefix = %q
	}
  }
  image {
    url         = "%s"
    work_dir    = "%s"
    command     = ["%s"]
    args        = ["%s"]
    environment = {
      %s = "%s"
    }
  }
  log_options {
  	disabled = "%t"
	log_group_id = yandex_logging_group.logging-group.id
	min_level = "%s"
  }
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
		params.memory,
		params.cores,
		params.coreFraction,
		params.executionTimeout,
		params.concurrency,
		params.runtime,
		params.secret.secretKey,
		params.secret.secretEnvVar,
		params.storageMount.storageMountPointPath,
		params.storageMount.storageMountPrefix,
		params.storageMount.storageMountReadOnly,
		params.ephemeralDiskMounts.mountPoint,
		params.ephemeralDiskMounts.mountMode,
		params.ephemeralDiskMounts.ephemeralDiskSizeGB,
		params.objectStorageMounts.mountPoint,
		params.objectStorageMounts.mountMode,
		params.objectStorageMounts.objectStoragePrefix,
		params.imageURL,
		params.workDir,
		params.command,
		params.argument,
		params.envVarKey,
		params.envVarValue,
		params.logOptions.disabled,
		params.logOptions.minLevel,
		params.storageMount.storageMountBucket,
		params.serviceAccount,
		params.secret.secretName,
		params.secret.secretKey,
		params.secret.secretValue)
}

type testResourceYandexServerlessContainerOptions struct {
	description      *string
	serviceAccountID *string
	image            *testResourceYandexServerlessContainerOptionsImage
	logOptions       *testResourceYandexServerlessContainerOptionsLogOptions
}

type testResourceYandexServerlessContainerOptionsImage struct {
	url string
}

type testResourceYandexServerlessContainerOptionsLogOptions struct {
	disabled   bool
	folderID   string
	LogGroupID string
	minLevel   string
}

type testResourceYandexServerlessContainerOption func(o *testResourceYandexServerlessContainerOptions)

type testResourceYandexServerlessContainerOptionFactoryImpl bool

const testResourceYandexServerlessContainerOptionFactory = testResourceYandexServerlessContainerOptionFactoryImpl(true)

func (testResourceYandexServerlessContainerOptionFactoryImpl) WithDescription(description string) testResourceYandexServerlessContainerOption {
	return func(o *testResourceYandexServerlessContainerOptions) {
		o.description = &description
	}
}

func (testResourceYandexServerlessContainerOptionFactoryImpl) WithServiceAccountID(serviceAccountID string) testResourceYandexServerlessContainerOption {
	return func(o *testResourceYandexServerlessContainerOptions) {
		o.serviceAccountID = &serviceAccountID
	}
}

func (testResourceYandexServerlessContainerOptionFactoryImpl) WithLogOptions(
	disabled bool,
	folderID string,
	LogGroupID string,
	minLevel string,
) testResourceYandexServerlessContainerOption {
	return func(o *testResourceYandexServerlessContainerOptions) {
		o.logOptions = &testResourceYandexServerlessContainerOptionsLogOptions{
			disabled:   disabled,
			folderID:   folderID,
			LogGroupID: LogGroupID,
			minLevel:   minLevel,
		}
	}
}

func testWriteResourceYandexServerlessContainer(
	sb *strings.Builder,
	resourceName string,
	containerName string,
	memoryMiB uint,
	imageURL string,
	options ...testResourceYandexServerlessContainerOption,
) {
	o := testResourceYandexServerlessContainerOptions{
		image: &testResourceYandexServerlessContainerOptionsImage{url: imageURL},
	}
	for _, option := range options {
		option(&o)
	}

	fprintfLn := func(sb *strings.Builder, format string, a ...any) {
		_, _ = fmt.Fprintf(sb, format, a...)
		sb.WriteRune('\n')
	}

	fprintfLn(sb, "resource \"yandex_serverless_container\" \"%s\" {", resourceName)
	fprintfLn(sb, "  name = \"%s\"", containerName)
	if description := o.description; description != nil {
		fprintfLn(sb, "  description = \"%s\"", *description)
	}
	fprintfLn(sb, "  memory = %d", memoryMiB)
	if serviceAccountID := o.serviceAccountID; serviceAccountID != nil {
		fprintfLn(sb, "  service_account_id = \"%s\"", *serviceAccountID)
	}
	fprintfLn(sb, "  image {")
	fprintfLn(sb, "    url = \"%s\"", o.image.url)
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
