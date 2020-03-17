package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/triggers/v1"
)

const triggerDataSource = "data.yandex_function_trigger.test-trigger"

func TestDataSourceYandexFunctionTrigger_byID(t *testing.T) {
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

func TestDataSourceYandexFunctionTrigger_byName(t *testing.T) {
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
  trigger_id = "${yandex_function_trigger.test-trigger.id}"
}

resource "yandex_function_trigger" "test-trigger" {
  name        = "%s"
  description = "%s"
  timer {
    cron_expression = "* * * * ? *"
  }
  function {
    id = "tf-test"
  }
}
	`, name, desc)
}

func testYandexFunctionTriggerByName(name string, desc string) string {
	return fmt.Sprintf(`
data "yandex_function_trigger" "test-trigger" {
  name = "${yandex_function_trigger.test-trigger.name}"
}

resource "yandex_function_trigger" "test-trigger" {
  name        = "%s"
  description = "%s"
  timer {
    cron_expression = "* * * * ? *"
  }
  function {
    id = "tf-test"
  }
}
	`, name, desc)
}
