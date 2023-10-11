package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/functions/v1"
)

func importFunctionIDFunc(function *functions.Function, role string) func(*terraform.State) (string, error) {
	return func(s *terraform.State) (string, error) {
		return function.Id + " " + role, nil
	}
}

func TestAccFunctionIamBinding(t *testing.T) {
	var function functions.Function
	functionName := acctest.RandomWithPrefix("tf-function")
	zipFilename := "test-fixtures/serverless/main.zip"

	userID := "allUsers"
	role := "serverless.functions.invoker"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionIamBinding_basic(functionName, zipFilename, role, userID),
				Check: resource.ComposeTestCheckFunc(
					testYandexFunctionExists(functionResource, &function),
					testAccCheckFunctionIam(functionResource, role, []string{"system:" + userID}),
				),
			},
			{
				ResourceName:      "yandex_function_iam_binding.foo",
				ImportStateIdFunc: importFunctionIDFunc(&function, role),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

//revive:disable:var-naming
func testAccFunctionIamBinding_basic(funcName, zipFile, role, userID string) string {
	return fmt.Sprintf(`
resource "yandex_function" "test-function" {
  name       = "%s"
  user_hash  = "user_hash"
  runtime    = "python37"
  entrypoint = "main"
  memory     = "128"
  content {
    zip_filename = "%s"
  }
}

resource "yandex_function_iam_binding" "foo" {
  function_id = yandex_function.test-function.id
  role        = "%s"
  members     = ["system:%s"]
}
`, funcName, zipFile, role, userID)
}
