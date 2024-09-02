package yandex

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1/privatelink"
)

func TestAccDataSourceVPCPrivateEndpointByID(t *testing.T) {
	t.Parallel()

	networkName := acctest.RandomWithPrefix("tf-network")
	subnetName := acctest.RandomWithPrefix("tf-subnet")
	peName := acctest.RandomWithPrefix("tf-private-endpoint")

	var pe privatelink.PrivateEndpoint

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVPCPrivateEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVPCPrivateEndpointConfig(networkName, subnetName, peName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCPrivateEndpointExists("yandex_vpc_private_endpoint.pe", &pe),
					resource.TestCheckResourceAttr("data.yandex_vpc_private_endpoint.pe", "name", peName),
					resource.TestCheckResourceAttrSet("data.yandex_vpc_private_endpoint.pe", "folder_id"),
					testAccCheckCreatedAtAttr("data.yandex_vpc_private_endpoint.pe"),
					resource.TestCheckResourceAttr("data.yandex_vpc_private_endpoint.pe", "dns_options.#", "1"),
					resource.TestCheckResourceAttr("data.yandex_vpc_private_endpoint.pe", "dns_options.0.private_dns_records_enabled", "false"),
					resource.TestCheckResourceAttr("data.yandex_vpc_private_endpoint.pe", "endpoint_address.#", "1"),
					resource.TestCheckResourceAttrSet("data.yandex_vpc_private_endpoint.pe", "endpoint_address.0.subnet_id"),
					resource.TestCheckResourceAttrSet("data.yandex_vpc_private_endpoint.pe", "endpoint_address.0.address"),
					resource.TestCheckResourceAttrSet("data.yandex_vpc_private_endpoint.pe", "endpoint_address.0.address_id"),
				),
			},
		},
	})
}

func TestAccDataSourceVPCPrivateEndpointByName(t *testing.T) {
	t.Parallel()

	networkName := acctest.RandomWithPrefix("tf-network")
	subnetName := acctest.RandomWithPrefix("tf-subnet")
	peName := acctest.RandomWithPrefix("tf-private-endpoint")

	var pe privatelink.PrivateEndpoint

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVPCPrivateEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVPCPrivateEndpointConfig(networkName, subnetName, peName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCPrivateEndpointExists("yandex_vpc_private_endpoint.pe", &pe),
					resource.TestCheckResourceAttr("data.yandex_vpc_private_endpoint.pe", "name", peName),
					resource.TestCheckResourceAttrSet("data.yandex_vpc_private_endpoint.pe", "folder_id"),
					testAccCheckCreatedAtAttr("data.yandex_vpc_private_endpoint.pe"),
					resource.TestCheckResourceAttr("data.yandex_vpc_private_endpoint.pe", "dns_options.#", "1"),
					resource.TestCheckResourceAttr("data.yandex_vpc_private_endpoint.pe", "dns_options.0.private_dns_records_enabled", "false"),
					resource.TestCheckResourceAttr("data.yandex_vpc_private_endpoint.pe", "endpoint_address.#", "1"),
					resource.TestCheckResourceAttrSet("data.yandex_vpc_private_endpoint.pe", "endpoint_address.0.subnet_id"),
					resource.TestCheckResourceAttrSet("data.yandex_vpc_private_endpoint.pe", "endpoint_address.0.address"),
					resource.TestCheckResourceAttrSet("data.yandex_vpc_private_endpoint.pe", "endpoint_address.0.address_id"),
				),
			},
		},
	})
}

const vpcPrivateEndpointDataSourceByID = `
data "yandex_vpc_private_endpoint" "pe" {
  private_endpoint_id = "${yandex_vpc_private_endpoint.pe.id}"
}
`

const vpcPrivateEndpointDataSourceByName = `
data "yandex_vpc_private_endpoint" "pe" {
  name = "${yandex_vpc_private_endpoint.pe.name}"
}
`

func testAccDataSourceVPCPrivateEndpointConfig(networkName, subnetName, peName string, byName bool) string {
	spec := testAccVPCPrivateEndpointConfigBasic(networkName, subnetName, peName)
	if byName {
		return spec + vpcPrivateEndpointDataSourceByName
	}
	return spec + vpcPrivateEndpointDataSourceByID
}
