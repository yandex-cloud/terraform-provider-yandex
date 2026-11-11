package acceptance

import (
	"fmt"
	"testing"

	functionfixtures "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/serverless/functions/testfixtures"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/functions/v1"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
)

func importFunctionIDFunc(function *functions.Function, role string) func(*terraform.State) (string, error) {
	return func(s *terraform.State) (string, error) {
		return function.Id + "," + role, nil
	}
}

func TestAccFunctionIamBinding(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	var function functions.Function
	functionName := acctest.RandomWithPrefix("tf-function")
	zipFilename := "../../../../test-fixtures/serverless/main.zip"

	userID := "allUsers"
	role := "serverless.functions.invoker"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		Steps: []resource.TestStep{
			{
				Config: testAccFunctionIamBinding_basic(functionName, zipFilename, role, userID),
				Check: resource.ComposeTestCheckFunc(
					functionfixtures.CheckExists(suite.Config, functionResource, &function),
					suite.testAccCheckFunctionIam(functionResource, role, []string{"system:" + userID}),
				),
			},
			{
				ResourceName:                         "yandex_function_iam_binding.foo",
				ImportStateIdFunc:                    importFunctionIDFunc(&function, role),
				ImportState:                          true,
				ImportStateVerifyIdentifierAttribute: "function_id",
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
