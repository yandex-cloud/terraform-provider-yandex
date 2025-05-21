package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccYandexYQObjectStorageConnection_basic(t *testing.T) {
	connectionResourceName := "my-connection"
	existingConnectionResourceName := fmt.Sprintf("yandex_yq_object_storage_connection.%s", connectionResourceName)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		// CheckDestroy: testYandexYDBDatabaseServerlessDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccYQObjectStorageConnectionConfig(connectionResourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccYQObjectStorageConnectionExists(existingConnectionResourceName),
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

func testAccYQObjectStorageConnectionConfig(connectionResourceName string) string {
	return fmt.Sprintf(`
	resource "yandex_yq_object_storage_connection" "%s" {
        name = "my_cnn_name"
        bucket = "my_bucket"
    }`,
		connectionResourceName,
	)
}

func testAccYQObjectStorageConnectionExists(connectionResourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		prs, ok := s.RootModule().Resources[connectionResourceName]
		if !ok {
			return fmt.Errorf("not found: %s, r: %v", connectionResourceName, s.RootModule().Resources)
		}
		if prs.Primary.ID == "" {
			return fmt.Errorf("%s", "no ID for connection is set")
		}
		return nil
	}
}
