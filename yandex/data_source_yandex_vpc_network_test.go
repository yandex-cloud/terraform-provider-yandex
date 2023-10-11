package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
)

func TestAccDataSourceVPCNetwork_byID(t *testing.T) {
	t.Parallel()

	networkName := acctest.RandomWithPrefix("tf-network")
	networkDesc := "Description for test"
	folderID := getExampleFolderID()

	var network vpc.Network

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVPCNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVPCNetworkConfig(networkName, networkDesc, true),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceVPCNetworkExists("data.yandex_vpc_network.bar", &network),
					testAccCheckResourceIDField("data.yandex_vpc_network.bar", "network_id"),
					resource.TestCheckResourceAttr("data.yandex_vpc_network.bar", "name", networkName),
					resource.TestCheckResourceAttr("data.yandex_vpc_network.bar", "description", networkDesc),
					resource.TestCheckResourceAttr("data.yandex_vpc_network.bar", "folder_id", folderID),
					resource.TestCheckResourceAttr("data.yandex_vpc_network.bar", "subnet_ids.#", "0"),
					testAccCheckCreatedAtAttr("data.yandex_vpc_network.bar"),
				),
			},
		},
	})
}

func TestAccDataSourceVPCNetwork_byName(t *testing.T) {
	t.Parallel()

	networkName := acctest.RandomWithPrefix("tf-network")
	networkDesc := "Description for test"
	folderID := getExampleFolderID()

	var network vpc.Network

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVPCNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVPCNetworkConfig(networkName, networkDesc, false),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceVPCNetworkExists("data.yandex_vpc_network.bar", &network),
					testAccCheckResourceIDField("data.yandex_vpc_network.bar", "network_id"),
					resource.TestCheckResourceAttr("data.yandex_vpc_network.bar", "name", networkName),
					resource.TestCheckResourceAttr("data.yandex_vpc_network.bar", "description", networkDesc),
					resource.TestCheckResourceAttr("data.yandex_vpc_network.bar", "folder_id", folderID),
					resource.TestCheckResourceAttr("data.yandex_vpc_network.bar", "subnet_ids.#", "0"),
					testAccCheckCreatedAtAttr("data.yandex_vpc_network.bar"),
				),
			},
		},
	})
}

func TestAccDataSourceVPCNetworkWithSubnets(t *testing.T) {
	t.Parallel()

	networkName := acctest.RandomWithPrefix("tf-network")
	networkDesc := "Test network with Subnets"
	folderID := getExampleFolderID()

	var network vpc.Network

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVPCNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVPCNetworkConfigWithSubnets(networkName, networkDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceVPCNetworkExists("yandex_vpc_network.foo", &network),
				),
			},
			{
				Config: testAccDataSourceVPCNetworkConfigWithSubnetsWithDataSource(networkName, networkDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceVPCNetworkExists("data.yandex_vpc_network.bar", &network),
					resource.TestCheckResourceAttr("data.yandex_vpc_network.bar", "name", networkName),
					resource.TestCheckResourceAttr("data.yandex_vpc_network.bar", "description", networkDesc),
					resource.TestCheckResourceAttr("data.yandex_vpc_network.bar", "folder_id", folderID),
					resource.TestCheckResourceAttr("data.yandex_vpc_network.bar", "subnet_ids.#", "2"),
					testAccCheckCreatedAtAttr("data.yandex_vpc_network.bar"),
				),
			},
		},
	})
}

func testAccDataSourceVPCNetworkExists(n string, network *vpc.Network) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ds, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if ds.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.VPC().Network().Get(context.Background(), &vpc.GetNetworkRequest{
			NetworkId: ds.Primary.ID,
		})

		if err != nil {
			return err
		}

		if found.Id != ds.Primary.ID {
			return fmt.Errorf("Network not found")
		}

		*network = *found

		return nil
	}
}

func testAccDataSourceVPCNetworkConfig(name, desc string, useID bool) string {
	if useID {
		return testAccDataSourceVPCNetworkResourceConfig(name, desc) + vpcNetworkDataByIDConfig
	}

	return testAccDataSourceVPCNetworkResourceConfig(name, desc) + vpcNetworkDataByNameConfig
}

//revive:disable:var-naming
func testAccDataSourceVPCNetworkResourceConfig(name, desc string) string {
	return fmt.Sprintf(`
resource "yandex_vpc_network" "foo" {
  name        = "%s"
  description = "%s"
}
`, name, desc)
}

const vpcNetworkDataByIDConfig = `
data "yandex_vpc_network" "bar" {
  network_id = "${yandex_vpc_network.foo.id}"
}
`

const vpcNetworkDataByNameConfig = `
data "yandex_vpc_network" "bar" {
  name = "${yandex_vpc_network.foo.name}"
}
`

func testAccDataSourceVPCNetworkConfigWithSubnets(name, desc string) string {
	return fmt.Sprintf(`
resource "yandex_vpc_network" "foo" {
  name        = "%s"
  description = "%s"
}

resource "yandex_vpc_subnet" "bar1" {
  network_id     = "${yandex_vpc_network.foo.id}"
  v4_cidr_blocks = ["172.16.1.0/24"]
}

resource "yandex_vpc_subnet" "bar2" {
  network_id     = "${yandex_vpc_network.foo.id}"
  v4_cidr_blocks = ["172.16.2.0/24"]
}
`, name, desc)
}

func testAccDataSourceVPCNetworkConfigWithSubnetsWithDataSource(name, desc string) string {
	return fmt.Sprintf(`
data "yandex_vpc_network" "bar" {
  network_id = "${yandex_vpc_network.foo.id}"
}

resource "yandex_vpc_network" "foo" {
  name        = "%s"
  description = "%s"
}

resource "yandex_vpc_subnet" "bar1" {
  network_id     = "${yandex_vpc_network.foo.id}"
  v4_cidr_blocks = ["172.16.1.0/24"]
}

resource "yandex_vpc_subnet" "bar2" {
  network_id     = "${yandex_vpc_network.foo.id}"
  v4_cidr_blocks = ["172.16.2.0/24"]
}
`, name, desc)
}
