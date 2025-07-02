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
			return fmt.Errorf("no ID for connection is set")
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
			return fmt.Errorf("connection with id %s still exists, resource type %s, details: %v", rs.Primary.ID, resourceType, response)
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
			return fmt.Errorf("no ID for binding is set")
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
			return fmt.Errorf("binding with id %s still exists, resource type %s, details: %v", rs.Primary.ID, resourceType, response)
		}
	}

	return nil
}

func SweepAllConnections(t Ydb_FederatedQuery.ConnectionSetting_ConnectionType) error {
	conf, err := ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	if conf.YqSdk == nil {
		return fmt.Errorf("YQ SDK is not initialized")
	}

	protoReq := &Ydb_FederatedQuery.ListConnectionsRequest{
		Filter: &Ydb_FederatedQuery.ListConnectionsRequest_Filter{
			ConnectionType: t,
		},
		Limit: 100,
	}

	ctx := context.Background()

	for {
		resp, err := conf.YqSdk.Client().ListConnections(ctx, protoReq)
		if err != nil {
			return fmt.Errorf("error getting connections: %w", err)
		}

		for _, c := range resp.Connection {
			id := c.Meta.Id
			err := conf.YqSdk.Client().DeleteConnection(ctx, &Ydb_FederatedQuery.DeleteConnectionRequest{
				ConnectionId: id,
			})

			if err != nil {
				return err
			}
		}
		if len(resp.NextPageToken) == 0 {
			break
		}

		protoReq.PageToken = resp.NextPageToken
	}

	return nil
}

func SweepAllBindings(t Ydb_FederatedQuery.BindingSetting_BindingType) error {
	conf, err := ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	if conf.YqSdk == nil {
		return fmt.Errorf("YQ SDK is not initialized")
	}

	protoReq := &Ydb_FederatedQuery.ListBindingsRequest{
		Limit: 100,
	}

	ctx := context.Background()

	for {
		resp, err := conf.YqSdk.Client().ListBindings(ctx, protoReq)
		if err != nil {
			return fmt.Errorf("error getting bindings: %w", err)
		}

		for _, c := range resp.Binding {
			if c.Type != t {
				continue
			}

			id := c.Meta.Id
			err := conf.YqSdk.Client().DeleteBinding(ctx, &Ydb_FederatedQuery.DeleteBindingRequest{
				BindingId: id,
			})

			if err != nil {
				return err
			}
		}
		if len(resp.NextPageToken) == 0 {
			break
		}

		protoReq.PageToken = resp.NextPageToken
	}

	return nil
}
