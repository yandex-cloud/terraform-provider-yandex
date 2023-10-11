package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDataSourceYandexMessageQueue_basic(t *testing.T) {
	var randInt int = acctest.RandInt()
	resourceName := "yandex_message_queue.test"
	datasourceName := "data.yandex_message_queue.by_name"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceYandexMessageQueueConfig(randInt),
				Check:  testAccDataSourceYandexMessageQueueCheck(datasourceName, resourceName),
			},
		},
	})
}

func testAccDataSourceYandexMessageQueueCheck(datasourceName, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[datasourceName]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", datasourceName)
		}

		sqsQueueRs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", resourceName)
		}

		attrNames := []string{
			"arn",
			"name",
		}

		for _, attrName := range attrNames {
			if rs.Primary.Attributes[attrName] != sqsQueueRs.Primary.Attributes[attrName] {
				return fmt.Errorf(
					"%s is %s; want %s",
					attrName,
					rs.Primary.Attributes[attrName],
					sqsQueueRs.Primary.Attributes[attrName],
				)
			}
		}

		return nil
	}
}

func testAccDataSourceYandexMessageQueueConfig(randInt int) string {
	return fmt.Sprintf(`
resource "yandex_message_queue" "wrong" {
  name = "%[1]d_wrong"

  access_key = yandex_iam_service_account_static_access_key.sa-key.access_key
  secret_key = yandex_iam_service_account_static_access_key.sa-key.secret_key
}

resource "yandex_message_queue" "test" {
  name = "%[1]d"

  access_key = yandex_iam_service_account_static_access_key.sa-key.access_key
  secret_key = yandex_iam_service_account_static_access_key.sa-key.secret_key
}

data "yandex_message_queue" "by_name" {
  name = "${yandex_message_queue.test.name}"

  access_key = yandex_iam_service_account_static_access_key.sa-key.access_key
  secret_key = yandex_iam_service_account_static_access_key.sa-key.secret_key
}
`, randInt) + testAccCommonIamDependenciesEditorConfig(randInt)
}
