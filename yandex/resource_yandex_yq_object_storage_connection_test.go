package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/ydb-platform/ydb-go-genproto/draft/protos/Ydb_FederatedQuery"
)

func TestAccYandexYQObjectStorageConnection_basic(t *testing.T) {
	connectionName := "my-conn"
	connectionResourceName := "my-connection"
	existingConnectionResourceName := fmt.Sprintf("yandex_yq_object_storage_connection.%s", connectionResourceName)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		// CheckDestroy: testYandexYDBDatabaseServerlessDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccYQObjectStorageConnectionConfig(connectionName, connectionResourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccYQObjectStorageConnectionExists(connectionName, existingConnectionResourceName),
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
        bucket = "my_bucket"
    }`,
		connectionResourceName,
		connectionName,
	)
}

func testAccYQObjectStorageConnectionExists(connectionName string, connectionResourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		prs, ok := s.RootModule().Resources[connectionResourceName]
		if !ok {
			return fmt.Errorf("not found: %s, r: %v", connectionResourceName, s.RootModule().Resources)
		}
		if prs.Primary.ID == "" {
			return fmt.Errorf("%s", "no ID for connection is set")
		}

		config := testAccProvider.Meta().(*Config)
		req := &Ydb_FederatedQuery.DescribeConnectionRequest{
			ConnectionId: prs.Primary.ID,
		}

		response, err := config.yqSdk.Client().DescribeConnection(context.Background(), req)
		if err != nil {
			return err
		}

		actualConnectionName := response.Connection.Content.Name
		if actualConnectionName != connectionName {
			return fmt.Errorf("invalid connection name %s, expected %s", actualConnectionName, connectionName)
		}

		return nil
	}
}
