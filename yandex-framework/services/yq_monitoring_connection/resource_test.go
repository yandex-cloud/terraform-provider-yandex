package yq_monitoring_connection_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func TestAccYQMonitoringConnectionBasic(t *testing.T) {
	connectionName := fmt.Sprintf("my-conn-%s", acctest.RandString(5))
	connectionResourceName := "my-connection"
	existingConnectionResourceName := fmt.Sprintf("yandex_yq_monitoring_connection.%s", connectionResourceName)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy: func(s *terraform.State) error {
			return test.TestYandexYQAllConnectionsDestroyed(s, "yandex_yq_monitoring_connection")
		},
		Steps: []resource.TestStep{
			{
				Config: testAccYQMonitoringConnectionConfig(connectionName, connectionResourceName),
				Check: resource.ComposeTestCheckFunc(
					test.TestAccYQConnectionExists(connectionName, existingConnectionResourceName),
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
	folderID := os.Getenv("YC_FOLDER_ID")
	return fmt.Sprintf(`
	resource "yandex_yq_monitoring_connection" "%s" {
        name = "%s"
		description = "my_desc"
		folder_id = "%s"
    }`,
		connectionResourceName,
		connectionName,
		folderID,
	)
}
