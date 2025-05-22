package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccYandexYQMonitoringConnectionBasic(t *testing.T) {
	connectionName := fmt.Sprintf("my-conn-%s", acctest.RandString(5))
	connectionResourceName := "my-connection"
	existingConnectionResourceName := fmt.Sprintf("yandex_yq_monitoring_connection.%s", connectionResourceName)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(s *terraform.State) error {
			return testYandexYQAllConnectionsDestroyed(s, "yandex_yq_monitoring_connection")
		},
		Steps: []resource.TestStep{
			{
				Config: testAccYQMonitoringConnectionConfig(connectionName, connectionResourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccYQConnectionExists(connectionName, existingConnectionResourceName),
				),
			},
			{
				ResourceName:      existingConnectionResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccYQMonitoringConnectionConfig(connectionName string, connectionResourceName string) string {
	return fmt.Sprintf(`
	resource "yandex_yq_monitoring_connection" "%s" {
        name = "%s"
        project = "my_project"
		cluster = "my_cluster"
    }`,
		connectionResourceName,
		connectionName,
	)
}
