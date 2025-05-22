package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/ydb-platform/ydb-go-genproto/draft/protos/Ydb_FederatedQuery"
)

func TestAccYandexYQObjectStorageConnection_basic(t *testing.T) {
	connectionName := fmt.Sprintf("my-conn-%s", acctest.RandString(5))
	connectionResourceName := "my-connection"
	existingConnectionResourceName := fmt.Sprintf("yandex_yq_object_storage_connection.%s", connectionResourceName)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexYQConnectionAllDestroyed,
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

func testGetYQConnectionByID(config *Config, connectionId string) (*Ydb_FederatedQuery.DescribeConnectionResult, error) {
	req := &Ydb_FederatedQuery.DescribeConnectionRequest{
		ConnectionId: connectionId,
	}

	return config.yqSdk.Client().DescribeConnection(context.Background(), req)
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
		response, err := testGetYQConnectionByID(config, prs.Primary.ID)
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

func testYandexYQConnectionAllDestroyed(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_yq_object_storage_connection" {
			continue
		}

		response, err := testGetYQConnectionByID(config, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Connection with id %s still exists, details: %v", rs.Primary.ID, response)
		}
	}

	return nil
}
