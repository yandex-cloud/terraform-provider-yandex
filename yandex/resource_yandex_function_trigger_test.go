package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/functions/v1"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	iot "github.com/yandex-cloud/go-genproto/yandex/cloud/iot/devices/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/triggers/v1"
)

const triggerResource = "yandex_function_trigger.test-trigger"

func init() {
	resource.AddTestSweepers("yandex_function_trigger", &resource.Sweeper{
		Name: "yandex_function_trigger",
		F:    testSweepFunctionTrigger,
	})
}

func testSweepFunctionTrigger(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &triggers.ListTriggersRequest{FolderId: conf.FolderID}
	it := conf.sdk.Serverless().Triggers().Trigger().TriggerIterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepFunctionTrigger(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Function Trigger %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepFunctionTrigger(conf *Config, id string) bool {
	return sweepWithRetry(sweepFunctionTriggerOnce, conf, "Function Trigger", id)
}

func sweepFunctionTriggerOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexFunctionDefaultTimeout)
	defer cancel()

	op, err := conf.sdk.Serverless().Triggers().Trigger().Delete(ctx, &triggers.DeleteTriggerRequest{
		TriggerId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}

func TestAccYandexFunctionTrigger_basic(t *testing.T) {
	t.Parallel()

	var trigger triggers.Trigger
	triggerName := acctest.RandomWithPrefix("tf-trigger")
	triggerDesc := acctest.RandomWithPrefix("tf-trigger-desc")
	labelKey := acctest.RandomWithPrefix("tf-trigger-label")
	labelValue := acctest.RandomWithPrefix("tf-trigger-label-value")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexFunctionTriggerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexFunctionTriggerBasic(triggerName, triggerDesc, labelKey, labelValue),
				Check: resource.ComposeTestCheckFunc(
					testYandexFunctionTriggerExists(triggerResource, &trigger),
					resource.TestCheckResourceAttr(triggerResource, "name", triggerName),
					resource.TestCheckResourceAttr(triggerResource, "description", triggerDesc),
					resource.TestCheckResourceAttrSet(triggerResource, "function.0.id"),
					resource.TestCheckResourceAttrSet(triggerResource, "folder_id"),
					resource.TestCheckResourceAttrSet(triggerResource, "timer.0.cron_expression"),
					testYandexFunctionTriggerContainsLabel(&trigger, labelKey, labelValue),
					testAccCheckCreatedAtAttr(triggerResource),
				),
			},
			functionTriggerImportTestStep(),
		},
	})
}

func TestAccYandexFunctionTrigger_update(t *testing.T) {
	t.Parallel()

	var trigger triggers.Trigger
	triggerName := acctest.RandomWithPrefix("tf-trigger")
	triggerDesc := acctest.RandomWithPrefix("tf-trigger-desc")
	labelKey := acctest.RandomWithPrefix("tf-trigger-label")
	labelValue := acctest.RandomWithPrefix("tf-trigger-label-value")

	triggerNameUpdated := acctest.RandomWithPrefix("tf-trigger")
	triggerDescUpdated := acctest.RandomWithPrefix("tf-trigger-desc")
	labelKeyUpdated := acctest.RandomWithPrefix("tf-trigger-label")
	labelValueUpdated := acctest.RandomWithPrefix("tf-trigger-label-value")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexFunctionTriggerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexFunctionTriggerBasic(triggerName, triggerDesc, labelKey, labelValue),
				Check: resource.ComposeTestCheckFunc(
					testYandexFunctionTriggerExists(triggerResource, &trigger),
					resource.TestCheckResourceAttr(triggerResource, "name", triggerName),
					resource.TestCheckResourceAttr(triggerResource, "description", triggerDesc),
					resource.TestCheckResourceAttrSet(triggerResource, "timer.0.cron_expression"),
					resource.TestCheckResourceAttrSet(triggerResource, "function.0.id"),
					testYandexFunctionTriggerContainsLabel(&trigger, labelKey, labelValue),
					testAccCheckCreatedAtAttr(triggerResource),
				),
			},
			functionTriggerImportTestStep(),
			{
				Config: testYandexFunctionTriggerBasic(triggerNameUpdated, triggerDescUpdated, labelKeyUpdated, labelValueUpdated),
				Check: resource.ComposeTestCheckFunc(
					testYandexFunctionTriggerExists(triggerResource, &trigger),
					resource.TestCheckResourceAttr(triggerResource, "name", triggerNameUpdated),
					resource.TestCheckResourceAttr(triggerResource, "description", triggerDescUpdated),
					testYandexFunctionTriggerContainsLabel(&trigger, labelKeyUpdated, labelValueUpdated),
					testAccCheckCreatedAtAttr(triggerResource),
				),
			},
			functionTriggerImportTestStep(),
		},
	})
}

func TestAccYandexFunctionTrigger_iot(t *testing.T) {
	t.Parallel()

	var trigger triggers.Trigger
	triggerName := acctest.RandomWithPrefix("tf-trigger")
	registryName := acctest.RandomWithPrefix("tf-registry")
	deviceName := acctest.RandomWithPrefix("tf-device")

	var registry iot.Registry
	var device iot.Device

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexFunctionTriggerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexFunctionTriggerIoT(registryName, deviceName, triggerName),
				Check: resource.ComposeTestCheckFunc(
					testYandexFunctionTriggerExists(triggerResource, &trigger),
					testYandexIoTCoreDeviceExists(iotRegistryResourceForDevices, iotDeviceResource, &registry, &device),
					resource.TestCheckResourceAttr(triggerResource, "name", triggerName),
					resource.TestCheckResourceAttrSet(triggerResource, "function.0.id"),
					testCheckResourceAttrByPointer(triggerResource, "iot.0.registry_id", &registry.Id),
					testCheckResourceAttrByPointer(triggerResource, "iot.0.device_id", &device.Id),
					resource.TestCheckResourceAttrSet(triggerResource, "iot.0.topic"),
					testAccCheckCreatedAtAttr(triggerResource),
				),
			},
			functionTriggerImportTestStep(),
		},
	})
}

func TestAccYandexFunctionTrigger_message(t *testing.T) {
	t.Skip("TODO: Use the test in manual mode or when YMQ will be implemented for provider")

	t.Parallel()

	var trigger triggers.Trigger
	triggerName := acctest.RandomWithPrefix("tf-trigger")

	queueID := "<use your queueID>"
	serviceAccount := acctest.RandomWithPrefix("tf-service-account")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexFunctionTriggerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexFunctionTriggerMessageQueue(triggerName, queueID, serviceAccount),
				Check: resource.ComposeTestCheckFunc(
					testYandexFunctionTriggerExists(triggerResource, &trigger),
					resource.TestCheckResourceAttr(triggerResource, "name", triggerName),
					resource.TestCheckResourceAttrSet(triggerResource, "function.0.id"),
					resource.TestCheckResourceAttr(triggerResource, "message_queue.0.queue_id", queueID),
					resource.TestCheckResourceAttrSet(triggerResource, "message_queue.0.service_account_id"),
					resource.TestCheckResourceAttr(triggerResource, "message_queue.0.batch_cutoff", "10"),
					resource.TestCheckResourceAttr(triggerResource, "message_queue.0.batch_size", "3"),
					resource.TestCheckResourceAttr(triggerResource, "message_queue.0.visibility_timeout", "3"),
					testAccCheckCreatedAtAttr(triggerResource),
				),
			},
			functionTriggerImportTestStep(),
		},
	})
}

func TestAccYandexFunctionTrigger_object(t *testing.T) {
	t.Parallel()

	var trigger triggers.Trigger
	triggerName := acctest.RandomWithPrefix("tf-trigger")
	bucket := acctest.RandomWithPrefix("tf-bucket")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexFunctionTriggerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexFunctionTriggerObjectStorage(triggerName, bucket),
				Check: resource.ComposeTestCheckFunc(
					testYandexFunctionTriggerExists(triggerResource, &trigger),
					resource.TestCheckResourceAttr(triggerResource, "name", triggerName),
					resource.TestCheckResourceAttrSet(triggerResource, "function.0.id"),
					resource.TestCheckResourceAttrSet(triggerResource, "object_storage.0.bucket_id"),
					resource.TestCheckResourceAttr(triggerResource, "object_storage.0.prefix", "prefix"),
					resource.TestCheckResourceAttr(triggerResource, "object_storage.0.suffix", "suffix"),
					resource.TestCheckResourceAttr(triggerResource, "object_storage.0.create", "true"),
					resource.TestCheckResourceAttr(triggerResource, "object_storage.0.update", "true"),
					resource.TestCheckResourceAttr(triggerResource, "object_storage.0.delete", "true"),
					testAccCheckCreatedAtAttr(triggerResource),
				),
			},
			functionTriggerImportTestStep(),
		},
	})
}

func TestAccYandexFunctionTrigger_loggroup(t *testing.T) {
	t.Parallel()

	trigger := &triggers.Trigger{}
	triggerName := acctest.RandomWithPrefix("tf-trigger")
	logSrcFn := &functions.Function{}
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexFunctionTriggerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexFunctionTriggerLogGroup(triggerName, 5, 100),
				Check: resource.ComposeTestCheckFunc(
					testYandexFunctionTriggerExists(triggerResource, trigger),
					resource.TestCheckResourceAttr(triggerResource, "name", triggerName),
					resource.TestCheckResourceAttrSet(triggerResource, "function.0.id"),
					testYandexFunctionExists("yandex_function.log-src-fn", logSrcFn),
					resource.TestCheckResourceAttr(triggerResource, "log_group.0.batch_cutoff", "5"),
					resource.TestCheckResourceAttr(triggerResource, "log_group.0.batch_size", "100"),
					testAccCheckCreatedAtAttr(triggerResource),
					testTriggerLogGroupEqFunctionGroup(logSrcFn, trigger),
				),
			},
			functionTriggerImportTestStep(),
		},
	})
}

func testTriggerLogGroupEqFunctionGroup(fn *functions.Function, trigger *triggers.Trigger) resource.TestCheckFunc {
	return func(*terraform.State) error {
		lgs := trigger.Rule.GetCloudLogs().LogGroupId
		if len(lgs) != 1 {
			return fmt.Errorf("trigger expected to have one log group, found: %v", lgs)
		}
		if triggerLG := lgs[0]; triggerLG != fn.LogGroupId {
			return fmt.Errorf("trigger expected to have function's log group: %s, but got: %s", fn.LogGroupId, triggerLG)
		}
		return nil
	}
}

func TestAccYandexFunctionTrigger_logging(t *testing.T) {
	t.Parallel()

	trigger := &triggers.Trigger{}
	triggerName := acctest.RandomWithPrefix("tf-trigger")
	logSrcFn := &functions.Function{}
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexFunctionTriggerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexFunctionTriggerLogging(triggerName, 5, 100),
				Check: resource.ComposeTestCheckFunc(
					testYandexFunctionTriggerExists(triggerResource, trigger),
					resource.TestCheckResourceAttr(triggerResource, "name", triggerName),
					resource.TestCheckResourceAttrSet(triggerResource, "function.0.id"),
					testYandexFunctionExists("yandex_function.log-src-fn", logSrcFn),
					resource.TestCheckResourceAttr(triggerResource, "logging.0.batch_cutoff", "5"),
					resource.TestCheckResourceAttr(triggerResource, "logging.0.batch_size", "100"),
					testAccCheckCreatedAtAttr(triggerResource),
				),
			},
			functionTriggerImportTestStep(),
		},
	})
}

func testYandexFunctionTriggerDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_function_trigger" {
			continue
		}

		_, err := testGetFunctionTriggerByID(config, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Trigger still exists")
		}
	}

	return nil
}

func functionTriggerImportTestStep() resource.TestStep {
	return resource.TestStep{
		ResourceName:      "yandex_function_trigger.test-trigger",
		ImportState:       true,
		ImportStateVerify: true,
	}
}

func testYandexFunctionTriggerExists(name string, trigger *triggers.Trigger) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := testGetFunctionTriggerByID(config, rs.Primary.ID)
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Trigger not found")
		}

		*trigger = *found
		return nil
	}
}

func testCheckResourceAttrByPointer(name string, key string, value *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		f := resource.TestCheckResourceAttr(name, key, *value)
		return f(s)
	}
}

func testGetFunctionTriggerByID(config *Config, ID string) (*triggers.Trigger, error) {
	req := triggers.GetTriggerRequest{
		TriggerId: ID,
	}

	return config.sdk.Serverless().Triggers().Trigger().Get(context.Background(), &req)
}

func testYandexFunctionTriggerContainsLabel(trigger *triggers.Trigger, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		v, ok := trigger.Labels[key]
		if !ok {
			return fmt.Errorf("Expected label with key '%s' not found", key)
		}
		if v != value {
			return fmt.Errorf("Incorrect label value for key '%s': expected '%s' but found '%s'", key, value, v)
		}
		return nil
	}
}

func testYandexFunctionTriggerBasic(name string, desc string, labelKey string, labelValue string) string {
	return fmt.Sprintf(`
resource "yandex_iam_service_account" "test-account" {
  name = "%s-acc"
}

resource "yandex_resourcemanager_folder_iam_member" "test_account" {
  folder_id   = "%s"
  member      = "serviceAccount:${yandex_iam_service_account.test-account.id}"
  role        = "editor"
  sleep_after = 30
}

resource "yandex_function" "tf-test" {
  name       = "%s-func"
  user_hash  = "user_hash"
  runtime    = "python37"
  entrypoint = "main"
  memory     = "128"
  content {
    zip_filename = "test-fixtures/serverless/main.zip"
  }
  service_account_id = yandex_iam_service_account.test-account.id
  depends_on         = [yandex_resourcemanager_folder_iam_member.test_account]
}

resource "yandex_function_trigger" "test-trigger" {
  name        = "%s"
  description = "%s"
  labels = {
    %s          = "%s"
    empty-label = ""
  }
  timer {
    cron_expression = "* * * * ? *"
  }
  function {
    id                 = yandex_function.tf-test.id
    service_account_id = yandex_iam_service_account.test-account.id
  }
}
	`, name, getExampleFolderID(), name, name, desc, labelKey, labelValue)
}

func testYandexFunctionTriggerIoT(regName, devName, name string) string {
	return fmt.Sprintf(`
resource "yandex_iam_service_account" "test-account" {
  name = "%s-acc"
}

resource "yandex_resourcemanager_folder_iam_member" "test_account" {
  folder_id   = "%s"
  member      = "serviceAccount:${yandex_iam_service_account.test-account.id}"
  role        = "editor"
  sleep_after = 30
}

resource "yandex_iot_core_registry" "test-registry" {
  name = "%s"
}

resource "yandex_iot_core_device" "test-device" {
  registry_id = yandex_iot_core_registry.test-registry.id
  name        = "%s"
}

resource "yandex_function" "tf-test" {
  name       = "%s-func"
  user_hash  = "user_hash"
  runtime    = "python37"
  entrypoint = "main"
  memory     = "128"
  content {
    zip_filename = "test-fixtures/serverless/main.zip"
  }
  service_account_id = yandex_iam_service_account.test-account.id
  depends_on         = [yandex_resourcemanager_folder_iam_member.test_account]
}

resource "yandex_message_queue" "queue" {
  name = "%s-tfotherqueuq"

  access_key = yandex_iam_service_account_static_access_key.sa-key.access_key
  secret_key = yandex_iam_service_account_static_access_key.sa-key.secret_key
}

resource "yandex_function_trigger" "test-trigger" {
  name = "%s"
  iot {
    registry_id = yandex_iot_core_registry.test-registry.id
    device_id   = yandex_iot_core_device.test-device.id
    topic       = join("/", ["$devices", yandex_iot_core_device.test-device.id, "events"])
  }
  function {
    id                 = yandex_function.tf-test.id
    service_account_id = yandex_iam_service_account.test-account.id
  }
  dlq {
    queue_id           = yandex_message_queue.queue.arn
    service_account_id = yandex_iam_service_account.test-account.id
  }
}
	`, name, getExampleFolderID(), regName, devName, name, name, name) + testAccCommonIamDependenciesEditorConfig(acctest.RandInt())
}

//nolint:unused
func testYandexFunctionTriggerMessageQueue(name, queueID, serviceAccountID string) string {
	return fmt.Sprintf(`
resource "yandex_iam_service_account" "test-account" {
  name = "%s-acc"
}

resource "yandex_resourcemanager_folder_iam_member" "test_account" {
  folder_id   = "%s"
  member      = "serviceAccount:${yandex_iam_service_account.test-account.id}"
  role        = "editor"
  sleep_after = 30
}

resource "yandex_function" "tf-test" {
  name       = "%s-func"
  user_hash  = "user_hash"
  runtime    = "python37"
  entrypoint = "main"
  memory     = "128"
  content {
    zip_filename = "test-fixtures/serverless/main.zip"
  }
  service_account_id = yandex_iam_service_account.test-account.id
  depends_on         = [yandex_resourcemanager_folder_iam_member.test_account]
}

resource "yandex_function_trigger" "test-trigger" {
  name = "%s"
  message_queue {
    queue_id           = "%s"
    service_account_id = yandex_iam_service_account.test-account.id
    batch_cutoff       = "10"
    batch_size         = "3"
    visibility_timeout = "3"
  }
  function {
    id                 = yandex_function.tf-test.id
    service_account_id = yandex_iam_service_account.test-account.id
  }
}

resource "yandex_iam_service_account" "test-account" {
  name = "%s"
}
	`, name, getExampleFolderID(), name, name, queueID, serviceAccountID)
}

func testYandexFunctionTriggerObjectStorage(name, bucket string) string {
	return fmt.Sprintf(`
resource "yandex_function" "tf-test" {
  name       = "%s-func"
  user_hash  = "user_hash"
  runtime    = "python37"
  entrypoint = "main"
  memory     = "128"
  content {
    zip_filename = "test-fixtures/serverless/main.zip"
  }
  service_account_id = yandex_iam_service_account.sa.id
  depends_on         = [yandex_resourcemanager_folder_iam_member.test_account]
}

resource "yandex_function_trigger" "test-trigger" {
  name = "%s"
  object_storage {
    bucket_id = yandex_storage_bucket.tf-test.id
    prefix    = "prefix"
    suffix    = "suffix"
    create    = true
    update    = true
    delete    = true
  }
  function {
    id                 = yandex_function.tf-test.id
    service_account_id = yandex_iam_service_account.sa.id
  }
}

resource "yandex_iam_service_account" "sa" {
  name = "test-sa-for-tf-test-bucket-%s"
}

resource "yandex_resourcemanager_folder_iam_member" "test_account" {
  folder_id   = "%s"
  member      = "serviceAccount:${yandex_iam_service_account.sa.id}"
  role        = "editor"
  sleep_after = 30
}

resource "yandex_iam_service_account_static_access_key" "sa-key" {
  service_account_id = yandex_iam_service_account.sa.id

  depends_on = [
    yandex_resourcemanager_folder_iam_member.test_account
  ]
}

resource "yandex_storage_bucket" "tf-test" {
  bucket = "%s"

  access_key = yandex_iam_service_account_static_access_key.sa-key.access_key
  secret_key = yandex_iam_service_account_static_access_key.sa-key.secret_key
}
	`, name, name, bucket, getExampleFolderID(), bucket)
}

func testYandexFunctionTriggerLogGroup(name string, batchCutoffSeconds, batchSize int) string {
	return fmt.Sprintf(`
resource "yandex_iam_service_account" "test-account" {
  name = "%s-acc"
}

resource "yandex_resourcemanager_folder_iam_member" "test_account" {
  folder_id   = "%s"
  member      = "serviceAccount:${yandex_iam_service_account.test-account.id}"
  role        = "editor"
  sleep_after = 30
}

resource "yandex_function" "tf-test" {
  name       = "%s-func"
  user_hash  = "user_hash"
  runtime    = "python37"
  entrypoint = "main"
  memory     = "128"
  content {
    zip_filename = "test-fixtures/serverless/main.zip"
  }
  service_account_id = yandex_iam_service_account.test-account.id
  depends_on         = [yandex_resourcemanager_folder_iam_member.test_account]
}

resource "yandex_function" "log-src-fn" {
  name       = "%s-log-src-func"
  user_hash  = "user_hash"
  runtime    = "python37"
  entrypoint = "main"
  memory     = "128"
  content {
    zip_filename = "test-fixtures/serverless/main.zip"
  }
}

resource "yandex_function_trigger" "test-trigger" {
  name = "%s"
  log_group {
    log_group_ids      = [yandex_function.log-src-fn.loggroup_id]
    batch_cutoff       = "%d"
    batch_size         = "%d"
  }
  function {
    id                 = yandex_function.tf-test.id
    service_account_id = yandex_iam_service_account.test-account.id
  }
}
	`, name, getExampleFolderID(), name, name, name, batchCutoffSeconds, batchSize)
}

func testYandexFunctionTriggerLogging(name string, batchCutoffSeconds, batchSize int) string {
	return fmt.Sprintf(`
resource "yandex_iam_service_account" "test-account" {
  name = "%s-acc"
}

resource "yandex_logging_group" "default-logging-group" {
	name = "default"
}

resource "yandex_resourcemanager_folder_iam_member" "test_account" {
  folder_id   = "%s"
  member      = "serviceAccount:${yandex_iam_service_account.test-account.id}"
  role        = "editor"
  sleep_after = 30
}

resource "yandex_function" "tf-test" {
  name       = "%s-func"
  user_hash  = "user_hash"
  runtime    = "python37"
  entrypoint = "main"
  memory     = "128"
  content {
    zip_filename = "test-fixtures/serverless/main.zip"
  }
  service_account_id = yandex_iam_service_account.test-account.id
  depends_on         = [yandex_resourcemanager_folder_iam_member.test_account]
}

resource "yandex_function" "logging-src-fn" {
  name       = "%s-logging-src-func"
  user_hash  = "user_hash"
  runtime    = "python37"
  entrypoint = "main"
  memory     = "128"
  content {
    zip_filename = "test-fixtures/serverless/main.zip"
  }
}

resource "yandex_function_trigger" "test-trigger" {
  name = "%s"
  logging {
    group_id = yandex_logging_group.default-logging-group.id
    batch_cutoff = "%d"
    batch_size   = "%d"
	resource_ids = [yandex_function.logging-src-fn.id]
	resource_types = ["serverless.function"]
	levels = ["info"]
  }
  function {
    id                 = yandex_function.tf-test.id
    service_account_id = yandex_iam_service_account.test-account.id
  }
}
	`, name, getExampleFolderID(), name, name, name, batchCutoffSeconds, batchSize)
}
