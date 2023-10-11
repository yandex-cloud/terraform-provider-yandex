package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/containers/v1"
)

func importServerlessContainerIDFunc(container *containers.Container, role string) func(*terraform.State) (string, error) {
	return func(s *terraform.State) (string, error) {
		return container.Id + " " + role, nil
	}
}

func TestAccServerlessContainerIamBinding(t *testing.T) {
	var container containers.Container
	containerName := acctest.RandomWithPrefix("tf-container")
	memory := (1 + acctest.RandIntRange(1, 4)) * 128

	userID := "allUsers"
	role := "serverless.containers.invoker"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccServerlessContainerIamBinding_basic(containerName, memory, serverlessContainerTestImage1, role, userID),
				Check: resource.ComposeTestCheckFunc(
					testYandexServerlessContainerExists(serverlessContainerResource, &container),
					testAccCheckServerlessContainerIam(serverlessContainerResource, role, []string{"system:" + userID}),
				),
			},
			{
				ResourceName:      "yandex_serverless_container_iam_binding.foo",
				ImportStateIdFunc: importServerlessContainerIDFunc(&container, role),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

//revive:disable:var-naming
func testAccServerlessContainerIamBinding_basic(cName string, memory int, url, role, userID string) string {
	return fmt.Sprintf(`
resource "yandex_serverless_container" "test-container" {
  name        = "%s"
  memory      = %d
  image {
    url = "%s"
  }
}

resource "yandex_serverless_container_iam_binding" "foo" {
  container_id = yandex_serverless_container.test-container.id
  role        = "%s"
  members     = ["system:%s"]
}
`, cName, memory, url, role, userID)
}
