package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/triggers/v1"
)

const triggerDataSource = "data.yandex_function_trigger.test-trigger"

func TestAccDataSourceYandexFunctionTrigger_byID(t *testing.T) {
	t.Parallel()

	var trigger triggers.Trigger
	triggerName := acctest.RandomWithPrefix("tf-trigger")
	triggerDesc := acctest.RandomWithPrefix("tf-trigger-desc")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexFunctionTriggerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexFunctionTriggerByID(triggerName, triggerDesc),
				Check: resource.ComposeTestCheckFunc(
					testYandexFunctionTriggerExists(triggerDataSource, &trigger),
					resource.TestCheckResourceAttrSet(triggerDataSource, "trigger_id"),
					resource.TestCheckResourceAttr(triggerDataSource, "name", triggerName),
					resource.TestCheckResourceAttr(triggerDataSource, "description", triggerDesc),
					resource.TestCheckResourceAttrSet(triggerDataSource, "function.0.id"),
					resource.TestCheckResourceAttrSet(triggerDataSource, "folder_id"),
					resource.TestCheckResourceAttrSet(triggerDataSource, "timer.0.cron_expression"),
					testAccCheckCreatedAtAttr(triggerDataSource),
				),
			},
		},
	})
}

func TestAccDataSourceYandexFunctionTrigger_byName(t *testing.T) {
	t.Parallel()

	var trigger triggers.Trigger
	triggerName := acctest.RandomWithPrefix("tf-trigger")
	triggerDesc := acctest.RandomWithPrefix("tf-trigger-desc")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexFunctionTriggerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexFunctionTriggerByName(triggerName, triggerDesc),
				Check: resource.ComposeTestCheckFunc(
					testYandexFunctionTriggerExists(triggerDataSource, &trigger),
					resource.TestCheckResourceAttrSet(triggerDataSource, "trigger_id"),
					resource.TestCheckResourceAttr(triggerDataSource, "name", triggerName),
					resource.TestCheckResourceAttr(triggerDataSource, "description", triggerDesc),
					resource.TestCheckResourceAttrSet(triggerDataSource, "function.0.id"),
					resource.TestCheckResourceAttrSet(triggerDataSource, "folder_id"),
					resource.TestCheckResourceAttrSet(triggerDataSource, "timer.0.cron_expression"),
					testAccCheckCreatedAtAttr(triggerDataSource),
				),
			},
		},
	})
}

func testYandexFunctionTriggerByID(name string, desc string) string {
	return fmt.Sprintf(`
data "yandex_function_trigger" "test-trigger" {
  trigger_id = yandex_function_trigger.test-trigger.id
}

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
  timer {
    cron_expression = "* * * * ? *"
  }
  function {
    id                 = yandex_function.tf-test.id
    service_account_id = yandex_iam_service_account.test-account.id
  }
}
	`, name, getExampleFolderID(), name, name, desc)
}

func testYandexFunctionTriggerByName(name string, desc string) string {
	return fmt.Sprintf(`
data "yandex_function_trigger" "test-trigger" {
  name = yandex_function_trigger.test-trigger.name
}

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
  timer {
    cron_expression = "* * * * ? *"
  }
  function {
    id                 = yandex_function.tf-test.id
    service_account_id = yandex_iam_service_account.test-account.id
  }
}
	`, name, getExampleFolderID(), name, name, desc)
}
