package yandex_dns_inbound_endpoint_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func TestAccDataSourceDnsInboundEndpoint_byID(t *testing.T) {
	endpointName := acctest.RandomWithPrefix("tf-dns-inbound-endpoint")
	folderID := test.GetExampleFolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckDnsInboundEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceDnsInboundEndpointConfig(endpointName, folderID, true),
				Check: resource.ComposeTestCheckFunc(
					test.AccCheckResourceIDField("data.yandex_dns_inbound_endpoint.test", "dns_inbound_endpoint_id"),
					resource.TestCheckResourceAttr("data.yandex_dns_inbound_endpoint.test", "name", endpointName),
					resource.TestCheckResourceAttrSet("data.yandex_dns_inbound_endpoint.test", "id"),
					resource.TestCheckResourceAttr("data.yandex_dns_inbound_endpoint.test", "folder_id", folderID),
					resource.TestCheckResourceAttrSet("data.yandex_dns_inbound_endpoint.test", "created_at"),
					resource.TestCheckResourceAttrSet("data.yandex_dns_inbound_endpoint.test", "address"),
					resource.TestCheckResourceAttrSet("data.yandex_dns_inbound_endpoint.test", "address_id"),
					resource.TestCheckResourceAttrSet("data.yandex_dns_inbound_endpoint.test", "network_id"),
				),
			},
		},
	})
}

func testAccDataSourceDnsInboundEndpointConfig(name, folderID string, useID bool) string {
	if useID {
		return testAccDataSourceDnsInboundEndpointResourceConfig(name, folderID) + dnsInboundEndpointDataByIDConfig
	}

	return testAccDataSourceDnsInboundEndpointResourceConfig(name, folderID)
}

func testAccDataSourceDnsInboundEndpointResourceConfig(name, folderID string) string {
	return fmt.Sprintf(`
resource "yandex_vpc_network" "network1" {
  name = "%[1]s-network"
}

resource "yandex_vpc_subnet" "subnet1" {
  name           = "%[1]s-subnet"
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.network1.id
  v4_cidr_blocks = ["192.168.0.0/24"]
}

resource "yandex_vpc_address" "addr1" {
  name        = "%[1]s-addr"
  description = "internal address for DNS inbound endpoint"

  internal_ipv4_address {
    subnet_id = yandex_vpc_subnet.subnet1.id
  }
  deletion_protection = false
}

resource "yandex_dns_inbound_endpoint" "test" {
  folder_id  = "%[2]s"
  name       = "%[1]s"
  network_id = yandex_vpc_network.network1.id
  address_id = yandex_vpc_address.addr1.id
}
`, name, folderID)
}

const dnsInboundEndpointDataByIDConfig = `
data "yandex_dns_inbound_endpoint" "test" {
  dns_inbound_endpoint_id = "${yandex_dns_inbound_endpoint.test.id}"
}
`
