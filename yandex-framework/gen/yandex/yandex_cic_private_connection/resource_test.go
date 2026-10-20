package yandex_cic_private_connection_test

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
	cic "github.com/yandex-cloud/go-genproto/yandex/cloud/cic/v1"
	cicsdk "github.com/yandex-cloud/go-sdk/services/cic/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex"
	yandexframework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAccCICPrivateConnection_basic(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-cic")
	provider := yandexframework.NewFrameworkProvider().(*yandexframework.Provider)
	config := func(attributes string) string {
		return fmt.Sprintf(`
resource "yandex_cic_private_connection" "test" {
  name = %q
  folder_id = %q
  trunk_connection_id = %q
  vlan_id = %s
  ipv4_peering = {
    peering_subnet = %q
    cloud_ip = %q
    peer_ip = %q
    peer_bgp_asn = %s
  }
  %s
}

data "yandex_cic_private_connection" "test" {
  private_connection_id = yandex_cic_private_connection.test.id
}

data "yandex_cic_trunk_connection" "test" {
  trunk_connection_id = %q
}

data "yandex_cic_point_of_presence" "test" {
  point_of_presence_id = data.yandex_cic_trunk_connection.test.point_of_presence_id
}
`, name, os.Getenv("YC_FOLDER_ID"), os.Getenv("YC_CIC_TRUNK_ID"), os.Getenv("YC_CIC_VLAN_ID"),
			os.Getenv("YC_CIC_PEERING_SUBNET"), os.Getenv("YC_CIC_CLOUD_IP"), os.Getenv("YC_CIC_PEER_IP"),
			os.Getenv("YC_CIC_PEER_ASN"), attributes, os.Getenv("YC_CIC_TRUNK_ID"))
	}
	initial := config(`
  description = "Terraform acceptance test"
  labels = { stage = "initial" }
  ipv4_static_routes = []
`)
	updated := config(fmt.Sprintf(`
  description = "Updated Terraform acceptance test"
  labels = { stage = "updated" }
  ipv4_static_routes = [{ prefix = %q }]
`, os.Getenv("YC_CIC_STATIC_PREFIX")))
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			for _, key := range []string{"YC_FOLDER_ID", "YC_CIC_TRUNK_ID", "YC_CIC_VLAN_ID", "YC_CIC_PEERING_SUBNET", "YC_CIC_CLOUD_IP", "YC_CIC_PEER_IP", "YC_CIC_PEER_ASN", "YC_CIC_STATIC_PREFIX"} {
				if os.Getenv(key) == "" {
					t.Fatalf("%s must be set for acceptance tests", key)
				}
			}
			if os.Getenv("YC_TOKEN") == "" && os.Getenv("YC_SERVICE_ACCOUNT_KEY_FILE") == "" {
				t.Fatal("YC_TOKEN or YC_SERVICE_ACCOUNT_KEY_FILE must be set for acceptance tests")
			}
		},
		ProtoV6ProviderFactories: privateConnectionAcceptanceFactories(provider),
		CheckDestroy: func(state *terraform.State) error {
			for _, rs := range state.RootModule().Resources {
				if rs.Type != "yandex_cic_private_connection" {
					continue
				}
				_, err := cicsdk.NewPrivateConnectionClient(provider.GetConfig().SDKv2).Get(context.Background(),
					&cic.GetPrivateConnectionRequest{PrivateConnectionId: rs.Primary.ID})
				if status.Code(err) != codes.NotFound {
					return fmt.Errorf("private connection %s: expected NotFound after destroy, got %v", rs.Primary.ID, err)
				}
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: initial,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(privateConnectionResourceName, "name", name),
					resource.TestCheckResourceAttr(privateConnectionResourceName, "folder_id", os.Getenv("YC_FOLDER_ID")),
					resource.TestCheckResourceAttr(privateConnectionResourceName, "vlan_id", os.Getenv("YC_CIC_VLAN_ID")),
					resource.TestCheckResourceAttr(privateConnectionResourceName, "status", "ACTIVE"),
					resource.TestCheckResourceAttrPair(privateConnectionResourceName, "id", "data.yandex_cic_private_connection.test", "id"),
				),
			},
			{Config: initial, PlanOnly: true},
			{ResourceName: privateConnectionResourceName, ImportState: true, ImportStateVerify: true},
			{
				Config: updated,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(privateConnectionResourceName, "labels.stage", "updated"),
					resource.TestCheckResourceAttr(privateConnectionResourceName, "ipv4_static_routes.#", "1"),
					resource.TestCheckResourceAttr("data.yandex_cic_private_connection.test", "ipv4_static_routes.#", "1"),
				),
			},
			{Config: updated, PlanOnly: true},
			{ResourceName: privateConnectionResourceName, ImportState: true, ImportStateVerify: true},
			{
				Config: initial,
				Check:  resource.TestCheckResourceAttr(privateConnectionResourceName, "ipv4_static_routes.#", "0"),
			},
		},
	})
}

func privateConnectionAcceptanceFactories(provider *yandexframework.Provider) map[string]func() (tfprotov6.ProviderServer, error) {
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
