package yandex_dns_inbound_endpoint_iam_binding_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	dns "github.com/yandex-cloud/go-genproto/yandex/cloud/dns/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
	dnsv1sdk "github.com/yandex-cloud/go-sdk/services/dns/v1"
	vpcv1sdk "github.com/yandex-cloud/go-sdk/services/vpc/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers/iam"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
)

const timeout = 15 * time.Minute

var endpointName = test.GenerateNameForResource(10)

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestAccDNSInboundEndpoint_basicIamMember(t *testing.T) {
	var (
		endpoint dns.DnsInboundEndpoint
		userID   = "allUsers"
		role     = "dns.viewer"

		ctx, cancel = context.WithTimeout(context.Background(), timeout)
	)

	defer cancel()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckDnsInboundEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDnsInboundEndpointVPCOnly(endpointName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCNetworkExists("yandex_vpc_network.test", &vpc.Network{}),
				),
			},
			{
				Config: testAccDnsInboundEndpointWithIAMMember_basic(endpointName, role, userID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDnsInboundEndpointExists("yandex_dns_inbound_endpoint.foobar", &endpoint),
					iam.TestAccCheckIamBindingEqualsMembers(ctx, func() iam.BindingsGetter {
						cfg := test.AccProvider.(*yandex_framework.Provider).GetConfig()
						return dnsv1sdk.NewDnsInboundEndpointClient(cfg.SDKv2)
					}, &endpoint, role, []string{"system:" + userID}),
				),
			},
		},
	})
}

func testAccCheckDnsInboundEndpointExists(n string, endpoint *dns.DnsInboundEndpoint) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

		found, err := dnsv1sdk.NewDnsInboundEndpointClient(config.SDKv2).Get(context.Background(), &dns.GetDnsInboundEndpointRequest{
			DnsInboundEndpointId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("DNS Inbound Endpoint not found")
		}

		*endpoint = *found

		return nil
	}
}

func testAccDnsInboundEndpointVPCOnly(name string) string {
	return fmt.Sprintf(`
resource "yandex_vpc_network" "test" {
		name = "%[1]s"
}

resource "yandex_vpc_subnet" "test" {
		name           = "%[1]s"
		zone           = "ru-central1-a"
		network_id     = yandex_vpc_network.test.id
		v4_cidr_blocks = ["10.0.0.0/24"]
}
`, name)
}

func testAccCheckVPCNetworkExists(n string, network *vpc.Network) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

		found, err := vpcv1sdk.NewNetworkClient(config.SDKv2).Get(context.Background(), &vpc.GetNetworkRequest{
			NetworkId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Network not found")
		}

		*network = *found
		return nil
	}
}

//revive:disable:var-naming
func testAccDnsInboundEndpointWithIAMMember_basic(name, role, userID string) string {
	return fmt.Sprintf(`
resource "yandex_vpc_network" "test" {
  name = "%[1]s"
}

resource "yandex_vpc_subnet" "test" {
  name           = "%[1]s"
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.test.id
  v4_cidr_blocks = ["10.0.0.0/24"]
}

resource "yandex_vpc_address" "test" {
  name        = "%[1]s"
  description = "internal address for DNS inbound endpoint"

  internal_ipv4_address {
    subnet_id = yandex_vpc_subnet.test.id
  }
  deletion_protection = false
}

resource "yandex_dns_inbound_endpoint" "foobar" {
  name       = "%[1]s"
  network_id = yandex_vpc_network.test.id
  address_id = yandex_vpc_address.test.id
}

resource "yandex_dns_inbound_endpoint_iam_binding" "test-endpoint-binding" {
  role                    = "%[2]s"
  members                 = ["system:%[3]s"]
  dns_inbound_endpoint_id = yandex_dns_inbound_endpoint.foobar.id
}
`, name, role, userID)
}

func testAccCheckDnsInboundEndpointDestroy(s *terraform.State) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_dns_inbound_endpoint" {
			continue
		}

		_, err := dnsv1sdk.NewDnsInboundEndpointClient(config.SDKv2).Get(context.Background(), &dns.GetDnsInboundEndpointRequest{
			DnsInboundEndpointId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("DNS Inbound Endpoint still exists")
		}
	}

	return nil
}
