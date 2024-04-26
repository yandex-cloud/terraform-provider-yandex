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

func TestAccDataSourceVPCSubnet_basic(t *testing.T) {
	t.Parallel()

	subnetName1 := acctest.RandomWithPrefix("tf-subnet-1")
	subnetName2 := acctest.RandomWithPrefix("tf-subnet-2")
	subnetDesc1 := "Description for test subnet #1"
	subnetDesc2 := "Description for test subnet #2"

	folderID := getExampleFolderID()
	var network vpc.Network

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckVPCNetworkDestroy,
			testAccCheckVPCSubnetDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVPCSubnetConfig_basic(subnetName1, subnetDesc1, subnetName2, subnetDesc2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCNetworkExists("yandex_vpc_network.foo", &network),

					testAccDataSourceVPCSubnetExists("data.yandex_vpc_subnet.bar1"),
					testAccDataSourceVPCSubnetExists("data.yandex_vpc_subnet.bar2"),

					testAccCheckResourceIDField("data.yandex_vpc_subnet.bar1", "subnet_id"),
					resource.TestCheckResourceAttr("data.yandex_vpc_subnet.bar1", "name", subnetName1),
					resource.TestCheckResourceAttr("data.yandex_vpc_subnet.bar1", "description", subnetDesc1),
					resource.TestCheckResourceAttr("data.yandex_vpc_subnet.bar1", "folder_id", folderID),
					resource.TestCheckResourceAttr("data.yandex_vpc_subnet.bar1", "zone", "ru-central1-b"),
					resource.TestCheckResourceAttr("data.yandex_vpc_subnet.bar1", "v4_cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr("data.yandex_vpc_subnet.bar1", "v4_cidr_blocks.0", "172.16.1.0/24"),
					resource.TestCheckResourceAttrSet("data.yandex_vpc_subnet.bar1", "network_id"),
					testAccCheckCreatedAtAttr("data.yandex_vpc_subnet.bar1"),

					testAccCheckResourceIDField("data.yandex_vpc_subnet.bar2", "subnet_id"),
					resource.TestCheckResourceAttr("data.yandex_vpc_subnet.bar2", "name", subnetName2),
					resource.TestCheckResourceAttr("data.yandex_vpc_subnet.bar2", "description", subnetDesc2),
					resource.TestCheckResourceAttr("data.yandex_vpc_subnet.bar2", "folder_id", folderID),
					resource.TestCheckResourceAttr("data.yandex_vpc_subnet.bar2", "zone", "ru-central1-d"),
					resource.TestCheckResourceAttr("data.yandex_vpc_subnet.bar2", "v4_cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr("data.yandex_vpc_subnet.bar2", "v4_cidr_blocks.0", "172.16.2.0/24"),
					resource.TestCheckResourceAttrSet("data.yandex_vpc_subnet.bar2", "network_id"),
					testAccCheckCreatedAtAttr("data.yandex_vpc_subnet.bar2"),
				),
			},
		},
	})
}
func TestAccDataSourceVPCSubnet_withRouteTable(t *testing.T) {
	t.Parallel()

	subnetName := acctest.RandomWithPrefix("tf-subnet")
	subnetDesc := "Description for test subnet"

	folderID := getExampleFolderID()
	var subnet vpc.Subnet

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckVPCNetworkDestroy,
			testAccCheckVPCRouteTableDestroy,
			testAccCheckVPCSubnetDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVPCSubnetConfig_basicRouteTable(subnetName, subnetDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCSubnetExists("yandex_vpc_subnet.foo", &subnet),

					testAccDataSourceVPCSubnetExists("data.yandex_vpc_subnet.bar"),

					testAccCheckResourceIDField("data.yandex_vpc_subnet.bar", "subnet_id"),
					resource.TestCheckResourceAttr("data.yandex_vpc_subnet.bar", "name", subnetName),
					resource.TestCheckResourceAttr("data.yandex_vpc_subnet.bar", "description", subnetDesc),
					resource.TestCheckResourceAttr("data.yandex_vpc_subnet.bar", "folder_id", folderID),
					resource.TestCheckResourceAttr("data.yandex_vpc_subnet.bar", "zone", "ru-central1-b"),
					resource.TestCheckResourceAttr("data.yandex_vpc_subnet.bar", "v4_cidr_blocks.#", "1"),
					resource.TestCheckResourceAttr("data.yandex_vpc_subnet.bar", "v4_cidr_blocks.0", "172.16.1.0/24"),
					resource.TestCheckResourceAttrPtr("data.yandex_vpc_subnet.bar", "route_table_id", &subnet.RouteTableId),
					resource.TestCheckResourceAttrPtr("data.yandex_vpc_subnet.bar", "network_id", &subnet.NetworkId),
					testAccCheckCreatedAtAttr("data.yandex_vpc_subnet.bar"),
				),
			},
		},
	})
}

func TestAccDataSourceVPCSubnet_withDhcpOptions(t *testing.T) {
	t.Parallel()

	subnetName := acctest.RandomWithPrefix("tf-subnet")
	subnetDesc := "Description for test subnet"

	var subnet vpc.Subnet

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckVPCNetworkDestroy,
			testAccCheckVPCRouteTableDestroy,
			testAccCheckVPCSubnetDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVPCSubnetConfig_basicDhcpOptions(subnetName, subnetDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCSubnetExists("yandex_vpc_subnet.foo", &subnet),
					testAccDataSourceVPCSubnetExists("data.yandex_vpc_subnet.bar"),
					testAccCheckResourceIDField("data.yandex_vpc_subnet.bar", "subnet_id"),
					resource.TestCheckResourceAttr("data.yandex_vpc_subnet.bar", "dhcp_options.0.domain_name", "example.com"),
					resource.TestCheckResourceAttr("data.yandex_vpc_subnet.bar", "dhcp_options.0.domain_name_servers.0", "1.1.1.1"),
					resource.TestCheckResourceAttr("data.yandex_vpc_subnet.bar", "dhcp_options.0.domain_name_servers.1", "8.8.8.8"),
					resource.TestCheckResourceAttr("data.yandex_vpc_subnet.bar", "dhcp_options.0.ntp_servers.0", "193.67.79.202"),
				),
			},
		},
	})
}

func testAccDataSourceVPCSubnetExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ds, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if ds.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.VPC().Subnet().Get(context.Background(), &vpc.GetSubnetRequest{
			SubnetId: ds.Primary.ID,
		})

		if err != nil {
			return err
		}

		if found.Id != ds.Primary.ID {
			return fmt.Errorf("Subnet not found")
		}

		return nil
	}
}

//revive:disable:var-naming
func testAccDataSourceVPCSubnetConfig_basic(name1, desc1, name2, desc2 string) string {
	return fmt.Sprintf(`
data "yandex_vpc_subnet" "bar1" {
  subnet_id = "${yandex_vpc_subnet.foo1.id}"
}

data "yandex_vpc_subnet" "bar2" {
  name = "${yandex_vpc_subnet.foo2.name}"
}

resource "yandex_vpc_network" "foo" {
  name        = "%s"
  description = "description for test"
}

resource "yandex_vpc_subnet" "foo1" {
  name           = "%s"
  network_id     = "${yandex_vpc_network.foo.id}"
  description    = "%s"
  v4_cidr_blocks = ["172.16.1.0/24"]
  zone           = "ru-central1-b"
}

resource "yandex_vpc_subnet" "foo2" {
  name           = "%s"
  network_id     = "${yandex_vpc_network.foo.id}"
  description    = "%s"
  v4_cidr_blocks = ["172.16.2.0/24"]
  zone           = "ru-central1-d"
}
`, acctest.RandomWithPrefix("tf-network"), name1, desc1, name2, desc2)
}

func testAccDataSourceVPCSubnetConfig_basicRouteTable(name1, desc1 string) string {
	return fmt.Sprintf(`
data "yandex_vpc_subnet" "bar" {
  subnet_id = "${yandex_vpc_subnet.foo.id}"
}

resource "yandex_vpc_network" "foo" {
  name        = "%s"
  description = "description for test"
}

resource "yandex_vpc_subnet" "foo" {
  name           = "%s"
  network_id     = "${yandex_vpc_network.foo.id}"
  route_table_id = "${yandex_vpc_route_table.foo.id}"
  description    = "%s"
  v4_cidr_blocks = ["172.16.1.0/24"]
  zone           = "ru-central1-b"
}

resource "yandex_vpc_route_table" "foo" {
  network_id = "${yandex_vpc_network.foo.id}"

  static_route {
    destination_prefix = "172.32.10.0/24"
    next_hop_address   = "172.16.1.32"
  }
}
`, acctest.RandomWithPrefix("tf-network"), name1, desc1)
}

func testAccDataSourceVPCSubnetConfig_basicDhcpOptions(name, desc string) string {
	return fmt.Sprintf(`
data "yandex_vpc_subnet" "bar" {
  subnet_id = "${yandex_vpc_subnet.foo.id}"
}

resource "yandex_vpc_network" "foo" {
  name        = "%s"
  description = "description for test"
}

resource "yandex_vpc_subnet" "foo" {
  name           = "%s"
  network_id     = "${yandex_vpc_network.foo.id}"
  description    = "%s"
  v4_cidr_blocks = ["172.16.1.0/24"]
  zone           = "ru-central1-b"

  dhcp_options {
    domain_name 		= "example.com"
    domain_name_servers = ["1.1.1.1", "8.8.8.8"]
    ntp_servers 		= ["193.67.79.202"]
  }
}
`, acctest.RandomWithPrefix("tf-network"), name, desc)
}
