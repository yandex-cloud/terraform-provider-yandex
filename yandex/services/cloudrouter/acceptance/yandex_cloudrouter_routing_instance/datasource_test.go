package yandex_cloudrouter_routing_instance_test

import (
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	cloudrouter "github.com/yandex-cloud/go-genproto/yandex/cloud/cloudrouter/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestCloudRouterRoutingInstanceDataSource_byID(t *testing.T) {
	api := &routingInstanceAPIStub{instance: &cloudrouter.RoutingInstance{
		Id: "routing-instance-id", Name: "existing-router", FolderId: "folder-id", RegionId: "ru-central1",
		Status: cloudrouter.RoutingInstance_ACTIVE, CreatedAt: timestamppb.New(time.Date(2026, 9, 11, 0, 0, 0, 0, time.UTC)),
		Labels: map[string]string{"stage": "existing"},
		VpcInfo: []*cloudrouter.RoutingInstance_VpcInfo{{
			VpcNetworkId: "network-id",
			AzInfos: []*cloudrouter.RoutingInstance_VpcAzInfo{{ManualInfo: &cloudrouter.RoutingInstance_VpcManualInfo{
				AzId: "ru-central1-a", Prefixes: []string{"10.0.0.0/24"},
			}}},
		}},
		CicPrivateConnectionInfo: []*cloudrouter.RoutingInstance_CicPrivateConnectionInfo{{CicPrivateConnectionId: "private-connection-id"}},
	}}
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: routingInstanceTestFactories(t, api),
		Steps: []resource.TestStep{{
			Config: routingInstanceDataSourceTestConfig,
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("data.yandex_cloudrouter_routing_instance.test", "id", "routing-instance-id"),
				resource.TestCheckResourceAttr("data.yandex_cloudrouter_routing_instance.test", "name", "existing-router"),
				resource.TestCheckResourceAttr("data.yandex_cloudrouter_routing_instance.test", "folder_id", "folder-id"),
				resource.TestCheckResourceAttr("data.yandex_cloudrouter_routing_instance.test", "region_id", "ru-central1"),
				resource.TestCheckResourceAttr("data.yandex_cloudrouter_routing_instance.test", "status", "ACTIVE"),
				resource.TestCheckResourceAttr("data.yandex_cloudrouter_routing_instance.test", "labels.stage", "existing"),
				resource.TestCheckResourceAttr("data.yandex_cloudrouter_routing_instance.test", "vpc_info.#", "1"),
				resource.TestCheckTypeSetElemNestedAttrs("data.yandex_cloudrouter_routing_instance.test", "cic_private_connection_info.*", map[string]string{
					"cic_private_connection_id": "private-connection-id",
				}),
				api.checkMutations(0, 0),
			),
		}},
	})
}

func TestCloudRouterRoutingInstanceDataSource_readError(t *testing.T) {
	api := &routingInstanceAPIStub{readError: status.Error(codes.PermissionDenied, "read denied")}
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: routingInstanceTestFactories(t, api),
		Steps: []resource.TestStep{{
			Config:      routingInstanceDataSourceTestConfig,
			ExpectError: regexp.MustCompile("read denied"),
		}},
	})
}

func TestCloudRouterRoutingInstanceDataSource_notFound(t *testing.T) {
	api := &routingInstanceAPIStub{readError: status.Error(codes.NotFound, "routing instance not found")}
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: routingInstanceTestFactories(t, api),
		Steps: []resource.TestStep{{
			Config: routingInstanceDataSourceTestConfig,
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("data.yandex_cloudrouter_routing_instance.test", "routing_instance_id", "routing-instance-id"),
				resource.TestCheckNoResourceAttr("data.yandex_cloudrouter_routing_instance.test", "id"),
				resource.TestCheckNoResourceAttr("data.yandex_cloudrouter_routing_instance.test", "name"),
				resource.TestCheckNoResourceAttr("data.yandex_cloudrouter_routing_instance.test", "status"),
				api.checkMutations(0, 0),
			),
		}},
	})
}

const routingInstanceDataSourceTestConfig = `
data "yandex_cloudrouter_routing_instance" "test" {
  routing_instance_id = "routing-instance-id"
}
`
