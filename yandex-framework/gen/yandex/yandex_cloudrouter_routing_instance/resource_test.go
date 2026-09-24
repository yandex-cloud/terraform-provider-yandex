package yandex_cloudrouter_routing_instance_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	cloudrouter "github.com/yandex-cloud/go-genproto/yandex/cloud/cloudrouter/v1"
	cloudroutersdk "github.com/yandex-cloud/go-sdk/services/cloudrouter/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex"
	yandexframework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestAccCloudRouterRoutingInstance_basic(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-cloudrouter")
	provider := yandexframework.NewFrameworkProvider().(*yandexframework.Provider)
	config := func(attributes string) string {
		return fmt.Sprintf(`
resource "yandex_vpc_network" "test" {
  name = %q
}

resource "yandex_vpc_subnet" "test" {
  name           = %q
  zone           = %q
  network_id     = yandex_vpc_network.test.id
  v4_cidr_blocks = ["10.123.45.0/24"]
}

resource "yandex_cloudrouter_routing_instance" "test" {
  name      = %q
  folder_id = %q
  %s
}

data "yandex_cloudrouter_routing_instance" "test" {
  routing_instance_id = yandex_cloudrouter_routing_instance.test.id
}
`, name, name, os.Getenv("YC_ZONE"), name, os.Getenv("YC_FOLDER_ID"), attributes)
	}
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			for _, key := range []string{"YC_FOLDER_ID", "YC_ZONE"} {
				if os.Getenv(key) == "" {
					t.Fatalf("%s must be set for acceptance tests", key)
				}
			}
			if os.Getenv("YC_TOKEN") == "" && os.Getenv("YC_SERVICE_ACCOUNT_KEY_FILE") == "" {
				t.Fatal("YC_TOKEN or YC_SERVICE_ACCOUNT_KEY_FILE must be set for acceptance tests")
			}
		},
		ProtoV6ProviderFactories: routingInstanceAcceptanceFactories(provider),
		CheckDestroy: func(state *terraform.State) error {
			for _, rs := range state.RootModule().Resources {
				if rs.Type != "yandex_cloudrouter_routing_instance" {
					continue
				}
				_, err := cloudroutersdk.NewRoutingInstanceClient(provider.GetConfig().SDKv2).Get(context.Background(),
					&cloudrouter.GetRoutingInstanceRequest{RoutingInstanceId: rs.Primary.ID})
				if status.Code(err) != codes.NotFound {
					return fmt.Errorf("routing instance %s: expected NotFound after destroy, got %v", rs.Primary.ID, err)
				}
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: config(`
  description = "initial description"
  labels = { stage = "initial" }
  vpc_info = []
  cic_private_connection_info = []
`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(routingInstanceResourceName, "name", name),
					resource.TestCheckResourceAttr(routingInstanceResourceName, "folder_id", os.Getenv("YC_FOLDER_ID")),
					resource.TestCheckResourceAttrSet(routingInstanceResourceName, "created_at"),
					resource.TestCheckResourceAttrPair(routingInstanceResourceName, "id", "data.yandex_cloudrouter_routing_instance.test", "id"),
				),
			},
			{ResourceName: routingInstanceResourceName, ImportState: true, ImportStateVerify: true},
			{
				Config: config(`
  description = "updated description"
  labels = { stage = "updated" }
  vpc_info = [{
    vpc_network_id = yandex_vpc_network.test.id
    az_infos = [{ manual_info = {
      az_id = yandex_vpc_subnet.test.zone
      prefixes = yandex_vpc_subnet.test.v4_cidr_blocks
    } }]
  }]
`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(routingInstanceResourceName, "description", "updated description"),
					resource.TestCheckResourceAttr(routingInstanceResourceName, "labels.stage", "updated"),
					resource.TestCheckResourceAttr(routingInstanceResourceName, "vpc_info.#", "1"),
					resource.TestCheckResourceAttr("data.yandex_cloudrouter_routing_instance.test", "vpc_info.#", "1"),
				),
			},
			{ResourceName: routingInstanceResourceName, ImportState: true, ImportStateVerify: true},
			{
				Config: config(`
  description = ""
  labels = {}
  vpc_info = []
`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(routingInstanceResourceName, "description", ""),
					resource.TestCheckResourceAttr(routingInstanceResourceName, "labels.%", "0"),
					resource.TestCheckResourceAttr(routingInstanceResourceName, "vpc_info.#", "0"),
				),
			},
		},
	})
}

func routingInstanceAcceptanceFactories(provider *yandexframework.Provider) map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"yandex": func() (tfprotov6.ProviderServer, error) {
			ctx := context.Background()
			sdkProvider, err := tf5to6server.UpgradeServer(ctx, yandex.NewSDKProvider().GRPCProvider)
			if err != nil {
				return nil, err
			}
			mux, err := tf6muxserver.NewMuxServer(ctx, providerserver.NewProtocol6(provider), func() tfprotov6.ProviderServer { return sdkProvider })
			if err != nil {
				return nil, err
			}
			return mux.ProviderServer(), nil
		},
	}
}
