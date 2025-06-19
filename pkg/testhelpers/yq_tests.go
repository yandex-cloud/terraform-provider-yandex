package testhelpers

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"github.com/ydb-platform/ydb-go-genproto/draft/protos/Ydb_FederatedQuery"
)

func testGetYQConnectionByID(config *config.Config, connectionId string) (*Ydb_FederatedQuery.DescribeConnectionResult, error) {
	req := &Ydb_FederatedQuery.DescribeConnectionRequest{
		ConnectionId: connectionId,
	}

	return config.YqSdk.Client().DescribeConnection(context.Background(), req)
}

func TestAccYQConnectionExists(connectionName string, connectionResourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		prs, ok := s.RootModule().Resources[connectionResourceName]
		if !ok {
			return fmt.Errorf("not found: %s, r: %v", connectionResourceName, s.RootModule().Resources)
		}
		if prs.Primary.ID == "" {
			return fmt.Errorf("%s", "no ID for connection is set")
		}

		config := AccProvider.(*yandex_framework.Provider).GetConfig()
		response, err := testGetYQConnectionByID(&config, prs.Primary.ID)
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

func TestYandexYQAllConnectionsDestroyed(s *terraform.State, resourceType string) error {
	config := AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != resourceType {
			continue
		}

		response, err := testGetYQConnectionByID(&config, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Connection with id %s still exists, resource type %s, , details: %v", rs.Primary.ID, resourceType, response)
		}
	}

	return nil
}

func testGetYQBindingByID(config *config.Config, bindingId string) (*Ydb_FederatedQuery.DescribeBindingResult, error) {
	req := &Ydb_FederatedQuery.DescribeBindingRequest{
		BindingId: bindingId,
	}

	return config.YqSdk.Client().DescribeBinding(context.Background(), req)
}

func TestAccYQBindingExists(bindingName string, bindingResourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		prs, ok := s.RootModule().Resources[bindingResourceName]
		if !ok {
			return fmt.Errorf("not found: %s, r: %v", bindingResourceName, s.RootModule().Resources)
		}
		if prs.Primary.ID == "" {
			return fmt.Errorf("%s", "no ID for binding is set")
		}

		config := AccProvider.(*yandex_framework.Provider).GetConfig()
		response, err := testGetYQBindingByID(&config, prs.Primary.ID)
		if err != nil {
			return err
		}

		actualBindingName := response.Binding.Content.Name
		if actualBindingName != bindingName {
			return fmt.Errorf("invalid binding name %s, expected %s", actualBindingName, bindingName)
		}

		return nil
	}
}

func TestYandexYQAllBindingsDestroyed(s *terraform.State, resourceType string) error {
	config := AccProvider.(*yandex_framework.Provider).GetConfig()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != resourceType {
			continue
		}

		response, err := testGetYQBindingByID(&config, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Binding with id %s still exists, resource type %s, , details: %v", rs.Primary.ID, resourceType, response)
		}
	}

	return nil
}
