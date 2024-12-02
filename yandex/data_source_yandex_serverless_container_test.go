package yandex

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/containers/v1"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const serverlessContainerDataSource = "data.yandex_serverless_container.test-container"

func TestAccDataSourceYandexServerlessContainer_byID(t *testing.T) {
	t.Parallel()

	var container containers.Container
	containerName := acctest.RandomWithPrefix("tf-container")
	containerDesc := acctest.RandomWithPrefix("tf-container-desc")
	memory := (1 + acctest.RandIntRange(1, 3)) * 128

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexServerlessContainerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexServerlessContainerByID(containerName, containerDesc, memory, serverlessContainerTestImage1),
				Check: resource.ComposeTestCheckFunc(
					testYandexServerlessContainerExists(serverlessContainerDataSource, &container),
					resource.TestCheckResourceAttrSet(serverlessContainerDataSource, "container_id"),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "name", containerName),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "description", containerDesc),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "memory", strconv.Itoa(memory)),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "image.0.url", serverlessContainerTestImage1),
					resource.TestCheckResourceAttrSet(serverlessContainerDataSource, "folder_id"),
					testAccCheckCreatedAtAttr(serverlessContainerDataSource),
				),
			},
		},
	})
}

func TestAccDataSourceYandexServerlessContainer_byName(t *testing.T) {
	t.Parallel()

	var container containers.Container
	containerName := acctest.RandomWithPrefix("tf-container")
	containerDesc := acctest.RandomWithPrefix("tf-container-desc")
	memory := (1 + acctest.RandIntRange(1, 3)) * 128

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexServerlessContainerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexServerlessContainerByName(containerName, containerDesc, memory, serverlessContainerTestImage1),
				Check: resource.ComposeTestCheckFunc(
					testYandexServerlessContainerExists(serverlessContainerDataSource, &container),
					resource.TestCheckResourceAttrSet(serverlessContainerDataSource, "container_id"),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "name", containerName),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "description", containerDesc),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "memory", strconv.Itoa(memory)),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "image.0.url", serverlessContainerTestImage1),
					resource.TestCheckResourceAttrSet(serverlessContainerDataSource, "folder_id"),
					testAccCheckCreatedAtAttr(serverlessContainerDataSource),
				),
			},
		},
	})
}

func TestAccDataSourceYandexServerlessContainer_noRevision(t *testing.T) {
	t.Parallel()

	var container containers.Container
	tfName := "test-container"
	resourcePath := "yandex_serverless_container." + tfName
	dataSourcePath := "data.yandex_serverless_container." + tfName

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexServerlessContainerDestroy,
		Steps: []resource.TestStep{
			{
				Config: func() string {
					sb := &strings.Builder{}
					testWriteResourceYandexServerlessContainer(
						sb,
						tfName,
						acctest.RandomWithPrefix("tf-container"),
						128,
						serverlessContainerTestImage1,
						testResourceYandexServerlessContainerOptionFactory.WithDescription(acctest.RandomWithPrefix("tf-function-desc")),
						testResourceYandexServerlessContainerOptionFactory.WithServiceAccountID("non-existent"), // prevent creation of revision
					)
					testWriteDataSourceYandexServerlessContainer(
						sb,
						tfName,
						testDataSourceYandexServerlessContainerOptionFactory.WithContainerID("${"+resourcePath+".id}"),
					)
					return sb.String()
				}(),
				Check: resource.ComposeTestCheckFunc(
					// container exists
					testYandexServerlessContainerExists(resourcePath, &container),
					// container revision not exists
					testYandexServerlessContainerRevisionNotExists(resourcePath),
					// all container attributes are set
					resource.TestCheckResourceAttrPtr(serverlessContainerDataSource, "container_id", &container.Id),
					resource.TestCheckResourceAttrPtr(serverlessContainerDataSource, "name", &container.Name),
					resource.TestCheckResourceAttrPtr(serverlessContainerDataSource, "folder_id", &container.FolderId),
					resource.TestCheckResourceAttrPtr(serverlessContainerDataSource, "description", &container.Description),
					resource.TestCheckResourceAttrPtr(serverlessContainerDataSource, "url", &container.Url),
					testAccCheckCreatedAtAttr(serverlessContainerDataSource),
					// all revision attributes are not set
					resource.TestCheckNoResourceAttr(dataSourcePath, "revision_id"),
					resource.TestCheckNoResourceAttr(dataSourcePath, "memory"),
					resource.TestCheckNoResourceAttr(dataSourcePath, "cores"),
					resource.TestCheckNoResourceAttr(dataSourcePath, "core_fraction"),
					resource.TestCheckNoResourceAttr(dataSourcePath, "execution_timeout"),
					resource.TestCheckNoResourceAttr(dataSourcePath, "concurrency"),
					resource.TestCheckNoResourceAttr(dataSourcePath, "service_account_id"),
					resource.TestCheckNoResourceAttr(dataSourcePath, "secrets"),
					resource.TestCheckNoResourceAttr(dataSourcePath, "storage_mounts"),
					resource.TestCheckNoResourceAttr(dataSourcePath, "mounts"),
					resource.TestCheckNoResourceAttr(dataSourcePath, "image"),
					resource.TestCheckNoResourceAttr(dataSourcePath, "connectivity"),
					resource.TestCheckNoResourceAttr(dataSourcePath, "log_options"),
				),
			},
		},
	})
}

func TestAccDataSourceYandexServerlessContainer_full(t *testing.T) {
	t.Parallel()

	var container containers.Container
	params := testYandexServerlessContainerParameters{}
	params.name = acctest.RandomWithPrefix("tf-container")
	params.desc = acctest.RandomWithPrefix("tf-container-desc")
	params.labelKey = acctest.RandomWithPrefix("tf-container-label")
	params.labelValue = acctest.RandomWithPrefix("tf-container-label-value")
	params.memory = (1 + acctest.RandIntRange(1, 3)) * 128
	params.cores = 1
	params.coreFraction = 100
	params.executionTimeout = strconv.FormatInt(int64(1+acctest.RandIntRange(1, 10)), 10) + "s"
	params.concurrency = acctest.RandIntRange(1, 3)
	params.imageURL = serverlessContainerTestImage1
	params.runtime = "http"
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
		minLevel: "WARN",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexServerlessContainerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexServerlessContainerDataSource(params),
				Check: resource.ComposeTestCheckFunc(
					testYandexServerlessContainerExists(serverlessContainerDataSource, &container),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "name", params.name),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "description", params.desc),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "labels.%", "1"),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "labels."+params.labelKey, params.labelValue),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "memory", strconv.Itoa(params.memory)),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "cores", strconv.Itoa(params.cores)),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "core_fraction", strconv.Itoa(params.coreFraction)),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "execution_timeout", params.executionTimeout),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "concurrency", strconv.Itoa(params.concurrency)),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "image.0.url", params.imageURL),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "image.0.digest", serverlessContainerTestDigest1),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "image.0.work_dir", params.workDir),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "image.0.command.#", "1"),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "image.0.command.0", params.command),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "image.0.args.#", "1"),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "image.0.args.0", params.argument),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "image.0.environment.%", "1"),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "image.0.environment."+params.envVarKey, params.envVarValue),
					testYandexServerlessContainerServiceAccountAttr(serverlessContainerDataSource, serverlessContainerServiceAccountResource),
					resource.TestCheckResourceAttrSet(serverlessContainerDataSource, "secrets.0.id"),
					resource.TestCheckResourceAttrSet(serverlessContainerDataSource, "secrets.0.version_id"),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "secrets.0.key", params.secret.secretKey),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "secrets.0.environment_variable", params.secret.secretEnvVar),

					resource.TestCheckResourceAttr(serverlessContainerDataSource, "storage_mounts.#", "2"),

					resource.TestCheckResourceAttr(serverlessContainerDataSource, "storage_mounts.1.mount_point_path", params.storageMount.storageMountPointPath),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "storage_mounts.1.bucket", params.storageMount.storageMountBucket),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "storage_mounts.1.prefix", params.storageMount.storageMountPrefix),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "storage_mounts.1.read_only", fmt.Sprint(params.storageMount.storageMountReadOnly)),

					resource.TestCheckResourceAttr(serverlessContainerDataSource, "storage_mounts.0.mount_point_path", params.objectStorageMounts.mountPoint),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "storage_mounts.0.bucket", params.objectStorageMounts.objectStorageBucket),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "storage_mounts.0.prefix", params.objectStorageMounts.objectStoragePrefix),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "storage_mounts.0.read_only", modeStringToBool(params.objectStorageMounts.mountMode)),

					resource.TestCheckResourceAttr(serverlessContainerDataSource, "mounts.#", "3"),

					resource.TestCheckResourceAttr(serverlessContainerDataSource, "mounts.0.mount_point_path", params.ephemeralDiskMounts.mountPoint),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "mounts.0.mode", params.ephemeralDiskMounts.mountMode),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "mounts.0.ephemeral_disk.0.size_gb", strconv.Itoa(params.ephemeralDiskMounts.ephemeralDiskSizeGB)),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "mounts.0.ephemeral_disk.0.block_size_kb", strconv.Itoa(params.ephemeralDiskMounts.ephemeralDiskBlockSizeKB)),

					resource.TestCheckResourceAttr(serverlessContainerDataSource, "mounts.1.mount_point_path", params.objectStorageMounts.mountPoint),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "mounts.1.mode", params.objectStorageMounts.mountMode),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "mounts.1.object_storage.0.bucket", params.objectStorageMounts.objectStorageBucket),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "mounts.1.object_storage.0.prefix", params.objectStorageMounts.objectStoragePrefix),

					resource.TestCheckResourceAttr(serverlessContainerDataSource, "mounts.2.mount_point_path", params.storageMount.storageMountPointPath),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "mounts.2.mode", modeBoolToString(params.storageMount.storageMountReadOnly)),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "mounts.2.object_storage.0.bucket", params.storageMount.storageMountBucket),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "mounts.2.object_storage.0.prefix", params.storageMount.storageMountPrefix),

					resource.TestCheckResourceAttr(serverlessContainerDataSource, "runtime.#", "1"),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "log_options.0.disabled", fmt.Sprint(params.logOptions.disabled)), resource.TestCheckResourceAttr(serverlessContainerDataSource, "log_options.0.disabled", fmt.Sprint(params.logOptions.disabled)),
					resource.TestCheckResourceAttr(serverlessContainerDataSource, "log_options.0.min_level", params.logOptions.minLevel),
					resource.TestCheckResourceAttrSet(serverlessContainerDataSource, "log_options.0.log_group_id"),
					resource.TestCheckResourceAttrSet(serverlessContainerResource, "revision_id"),
					resource.TestCheckResourceAttrSet(serverlessContainerResource, "folder_id"),
					resource.TestCheckResourceAttrSet(serverlessContainerResource, "url"),
					testAccCheckCreatedAtAttr(serverlessContainerResource),
				),
			},
		},
	})
}

func testYandexServerlessContainerServiceAccountAttr(name string, serviceAccountResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found serverless container: %s", name)
		}
		sa, ok := s.RootModule().Resources[serviceAccountResource]
		if !ok {
			return fmt.Errorf("Not found service account: %s", serviceAccountResource)
		}
		serviceAccountID := rs.Primary.Attributes["service_account_id"]
		if serviceAccountID != sa.Primary.ID {
			return fmt.Errorf("Incorrect service account id: expected '%s' but found '%s'", sa.Primary.ID, serviceAccountID)
		}
		return nil
	}
}

func testYandexServerlessContainerByID(name string, desc string, memory int, image string) string {
	return fmt.Sprintf(`
data "yandex_serverless_container" "test-container" {
  container_id = "${yandex_serverless_container.test-container.id}"
}

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

func testYandexServerlessContainerByName(name string, desc string, memory int, image string) string {
	return fmt.Sprintf(`
data "yandex_serverless_container" "test-container" {
  name = "${yandex_serverless_container.test-container.name}"
}

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

func testYandexServerlessContainerDataSource(params testYandexServerlessContainerParameters) string {
	return fmt.Sprintf(`
data "yandex_serverless_container" "test-container" {
  container_id = "${yandex_serverless_container.test-container.id}"
}

resource "yandex_serverless_container" "test-container" {
  name        = "%s"
  description = "%s"
  labels = {
    %s   = "%s"
  }
  memory             = %d
  cores              = %d
  core_fraction      = %d
  execution_timeout  = "%s"
  concurrency        = %d
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

type testDataSourceYandexServerlessContainerOptions struct {
	name        *string
	containerID *string
}

type testDataSourceYandexServerlessContainerOption func(o *testDataSourceYandexServerlessContainerOptions)

type testDataSourceYandexServerlessContainerOptionFactoryImpl bool

const testDataSourceYandexServerlessContainerOptionFactory = testDataSourceYandexServerlessContainerOptionFactoryImpl(true)

func (testDataSourceYandexServerlessContainerOptionFactoryImpl) WithName(name string) testDataSourceYandexServerlessContainerOption {
	return func(o *testDataSourceYandexServerlessContainerOptions) {
		o.name = &name
	}
}

func (testDataSourceYandexServerlessContainerOptionFactoryImpl) WithContainerID(containerID string) testDataSourceYandexServerlessContainerOption {
	return func(o *testDataSourceYandexServerlessContainerOptions) {
		o.containerID = &containerID
	}
}

func testWriteDataSourceYandexServerlessContainer(
	sb *strings.Builder,
	resourceName string,
	options ...testDataSourceYandexServerlessContainerOption,
) {
	var o testDataSourceYandexServerlessContainerOptions
	for _, option := range options {
		option(&o)
	}

	fprintfLn := func(sb *strings.Builder, format string, a ...any) {
		_, _ = fmt.Fprintf(sb, format, a...)
		sb.WriteRune('\n')
	}

	fprintfLn(sb, "data \"yandex_serverless_container\" \"%s\" {", resourceName)
	if name := o.name; name != nil {
		fprintfLn(sb, "  name = \"%s\"", *name)
	}
	if containerID := o.containerID; containerID != nil {
		fprintfLn(sb, "  container_id = \"%s\"", *containerID)
	}
	fprintfLn(sb, "}")
}
