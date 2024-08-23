package yandex

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"testing"
)

const (
	iamServiceAgentData = "data.yandex_iam_service_agent.this_data"
)

func TestAccDataSourceIamServiceAgent(t *testing.T) {
	t.Parallel()
	cloudID := getExampleCloudID()
	serviceID := "connection-manager"
	microserviceID := "worker"
	agentServiceAccountID := "ajeh4a2l47inmf434s6e"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigIamServiceAgent(cloudID, serviceID, microserviceID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(iamServiceAgentData, "id", agentServiceAccountID),
					resource.TestCheckResourceAttr(iamServiceAgentData, "service_account_id", agentServiceAccountID),
					resource.TestCheckResourceAttr(iamServiceAgentData, "service_id", serviceID),
					resource.TestCheckResourceAttr(iamServiceAgentData, "microservice_id", microserviceID),
				),
			},
		},
	})
}

func testAccConfigIamServiceAgent(cloudID string, serviceID string, microserviceID string) string {
	return fmt.Sprintf(`
data "yandex_iam_service_agent" "this" {
  cloud_id        = "%s"
  service_id      = "%s"
  microservice_id = "%s"
}
`, cloudID, serviceID, microserviceID)
}
