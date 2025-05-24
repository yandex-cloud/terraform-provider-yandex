package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccYandexYQObjectStorageConnectionBasic(t *testing.T) {
	connectionName := fmt.Sprintf("my-conn-%s", acctest.RandString(5))
	connectionResourceName := "my-connection"
	existingConnectionResourceName := fmt.Sprintf("yandex_yq_object_storage_connection.%s", connectionResourceName)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(s *terraform.State) error {
			return testYandexYQAllConnectionsDestroyed(s, "yandex_yq_object_storage_connection")
		},
		Steps: []resource.TestStep{
			{
				Config: testAccYQObjectStorageConnectionConfig(connectionName, connectionResourceName),
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

func testAccYQObjectStorageConnectionConfig(connectionName string, connectionResourceName string) string {
	return fmt.Sprintf(`
	resource "yandex_yq_object_storage_connection" "%s" {
        name = "%s"
		description = "my_desc"
        bucket = "my_bucket"
    }`,
		connectionResourceName,
		connectionName,
	)
}
