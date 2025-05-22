package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/ydb-platform/ydb-go-genproto/draft/protos/Ydb_FederatedQuery"
)

func testGetYQConnectionByID(config *Config, connectionId string) (*Ydb_FederatedQuery.DescribeConnectionResult, error) {
	req := &Ydb_FederatedQuery.DescribeConnectionRequest{
		ConnectionId: connectionId,
	}

	return config.yqSdk.Client().DescribeConnection(context.Background(), req)
}

func testAccYQConnectionExists(connectionName string, connectionResourceName string) resource.TestCheckFunc {
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

func testYandexYQAllConnectionsDestroyed(s *terraform.State, resourceType string) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != resourceType {
			continue
		}

		response, err := testGetYQConnectionByID(config, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Connection with id %s still exists, resource type %s, , details: %v", rs.Primary.ID, resourceType, response)
		}
	}

	return nil
}
