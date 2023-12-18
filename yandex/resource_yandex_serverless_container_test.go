package yandex

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/c2h5oh/datasize"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/containers/v1"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const serverlessContainerResource = "yandex_serverless_container.test-container"
const serverlessContainerServiceAccountResource = "yandex_iam_service_account.test-account"
const serverlessContainerTestImage1 = "cr.yandex/yc/demo/coi:v1"
const serverlessContainerTestDigest1 = "sha256:e1d772fa8795adac847a2410c87d0d2e2d38fa02f118cab8c0b5fe1fb95c47f3"
const serverlessContainerTestImage2 = "cr.yandex/yc/demo/nginx-hostname:cli"

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
		ResourceName:      serverlessContainerResource,
		ImportState:       true,
		ImportStateVerify: true,
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
	params.concurrency = acctest.RandIntRange(0, 3)
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

	params.storageMount = testStorageMountParameters{
		storageMountPointPath: "/mount/point/path",
		storageMountBucket:    acctest.RandomWithPrefix("tf-function-test-bucket"),
		storageMountPrefix:    "tf-container-path",
		storageMountReadOnly:  false,
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
	params.concurrency = params.concurrency + 1
	paramsUpdated.imageURL = serverlessContainerTestImage2
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

	paramsUpdated.storageMount = testStorageMountParameters{
		storageMountPointPath: "/mount/point/path/updated",
		storageMountBucket:    acctest.RandomWithPrefix("tf-function-test-bucket-updated"),
		storageMountPrefix:    "tf-container-path-updated",
		storageMountReadOnly:  true,
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
				resource.TestCheckResourceAttrSet(serverlessContainerResource, "secrets.0.id"),
				resource.TestCheckResourceAttrSet(serverlessContainerResource, "secrets.0.version_id"),
				resource.TestCheckResourceAttr(serverlessContainerResource, "secrets.0.key", params.secret.secretKey),
				resource.TestCheckResourceAttr(serverlessContainerResource, "secrets.0.environment_variable", params.secret.secretEnvVar),
				resource.TestCheckResourceAttr(serverlessContainerResource, "storage_mounts.0.mount_point_path", params.storageMount.storageMountPointPath),
				resource.TestCheckResourceAttr(serverlessContainerResource, "storage_mounts.0.bucket", params.storageMount.storageMountBucket),
				resource.TestCheckResourceAttr(serverlessContainerResource, "storage_mounts.0.prefix", params.storageMount.storageMountPrefix),
				resource.TestCheckResourceAttr(serverlessContainerResource, "storage_mounts.0.read_only", fmt.Sprint(params.storageMount.storageMountReadOnly)),
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
	name             string
	desc             string
	labelKey         string
	labelValue       string
	memory           int
	cores            int
	coreFraction     int
	executionTimeout string
	concurrency      int
	imageURL         string
	workDir          string
	command          string
	argument         string
	envVarKey        string
	envVarValue      string
	serviceAccount   string
	secret           testSecretParameters
	storageMount     testStorageMountParameters
	logOptions       testLogOptions
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
  service_account_id = "${yandex_iam_service_account.test-account.id}"
  depends_on = [
	yandex_resourcemanager_folder_iam_member.payload-viewer
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
  folder_id = yandex_iam_service_account.test-account.folder_id
  role      = "storage.editor"
  member    = "serviceAccount:${yandex_iam_service_account.test-account.id}"
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
  folder_id = yandex_lockbox_secret.secret.folder_id
  role      = "lockbox.payloadViewer"
  member    = "serviceAccount:${yandex_iam_service_account.test-account.id}"
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
