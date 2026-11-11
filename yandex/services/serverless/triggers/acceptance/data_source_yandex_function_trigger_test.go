package acceptance

import (
	"fmt"
	"testing"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/acceptanceenv"

	statecheck "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance/statecheck"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/triggers/v1"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
)

const triggerDataSource = "data.yandex_function_trigger.test-trigger"

func TestAccDataSourceYandexFunctionTrigger_byID(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	var trigger triggers.Trigger
	triggerName := acctest.RandomWithPrefix("tf-trigger")
	triggerDesc := acctest.RandomWithPrefix("tf-trigger-desc")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testYandexFunctionTriggerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexFunctionTriggerByID(triggerName, triggerDesc),
				Check: resource.ComposeTestCheckFunc(
					suite.testYandexFunctionTriggerExists(triggerDataSource, &trigger),
					resource.TestCheckResourceAttrSet(triggerDataSource, "trigger_id"),
					resource.TestCheckResourceAttr(triggerDataSource, "name", triggerName),
					resource.TestCheckResourceAttr(triggerDataSource, "description", triggerDesc),
					resource.TestCheckResourceAttrSet(triggerDataSource, "function.0.id"),
					resource.TestCheckResourceAttrSet(triggerDataSource, "folder_id"),
					resource.TestCheckResourceAttrSet(triggerDataSource, "timer.0.cron_expression"),
					statecheck.CheckCreatedAtAttr(triggerDataSource),
				),
			},
		},
	})
}

func TestAccDataSourceYandexFunctionTrigger_byName(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	var trigger triggers.Trigger
	triggerName := acctest.RandomWithPrefix("tf-trigger")
	triggerDesc := acctest.RandomWithPrefix("tf-trigger-desc")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testYandexFunctionTriggerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexFunctionTriggerByName(triggerName, triggerDesc),
				Check: resource.ComposeTestCheckFunc(
					suite.testYandexFunctionTriggerExists(triggerDataSource, &trigger),
					resource.TestCheckResourceAttrSet(triggerDataSource, "trigger_id"),
					resource.TestCheckResourceAttr(triggerDataSource, "name", triggerName),
					resource.TestCheckResourceAttr(triggerDataSource, "description", triggerDesc),
					resource.TestCheckResourceAttrSet(triggerDataSource, "function.0.id"),
					resource.TestCheckResourceAttrSet(triggerDataSource, "folder_id"),
					resource.TestCheckResourceAttrSet(triggerDataSource, "timer.0.cron_expression"),
					statecheck.CheckCreatedAtAttr(triggerDataSource),
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
    zip_filename = "../../../../test-fixtures/serverless/main.zip"
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
	`, name, acceptanceenv.FolderID(), name, name, desc)
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
    zip_filename = "../../../../test-fixtures/serverless/main.zip"
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
	`, name, acceptanceenv.FolderID(), name, name, desc)
}
