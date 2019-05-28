package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
)

func TestAccVPCSubnet_basic(t *testing.T) {
	t.Parallel()

	var subnet1 vpc.Subnet
	var subnet2 vpc.Subnet

	networkName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	subnet1Name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	subnet2Name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVPCSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnet_basic(networkName, subnet1Name, subnet2Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCSubnetExists(
						"yandex_vpc_subnet.subnet-a", &subnet1),
					testAccCheckVPCSubnetExists(
						"yandex_vpc_subnet.subnet-b", &subnet2),
					testAccCheckCreatedAtAttr("yandex_vpc_subnet.subnet-a"),
					testAccCheckCreatedAtAttr("yandex_vpc_subnet.subnet-b"),
				),
			},
			{
				ResourceName:      "yandex_vpc_subnet.subnet-a",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "yandex_vpc_subnet.subnet-b",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCSubnet_update(t *testing.T) {
	t.Parallel()

	var network vpc.Network
	var subnet1 vpc.Subnet
	var subnet2 vpc.Subnet

	networkName := acctest.RandomWithPrefix("tf-network")
	subnet1Name := acctest.RandomWithPrefix("tf-subnet-a")
	subnet2Name := acctest.RandomWithPrefix("tf-subnet-b")
	updatedSubnet1Name := subnet1Name + "-update"
	updatedSubnet2Name := subnet2Name + "-update"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVPCSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnet_basic(networkName, subnet1Name, subnet2Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCNetworkExists("yandex_vpc_network.foo", &network),

					testAccCheckVPCSubnetExists("yandex_vpc_subnet.subnet-a", &subnet1),
					resource.TestCheckResourceAttrPtr("yandex_vpc_subnet.subnet-a", "network_id", &network.Id),
					resource.TestCheckResourceAttr("yandex_vpc_subnet.subnet-a", "name", subnet1Name),
					resource.TestCheckResourceAttr("yandex_vpc_subnet.subnet-a", "description", "description for subnet-a"),
					resource.TestCheckResourceAttr("yandex_vpc_subnet.subnet-a", "zone", "ru-central1-a"),
					resource.TestCheckResourceAttr("yandex_vpc_subnet.subnet-a", "v4_cidr_blocks.0", "10.0.0.0/16"),
					testAccCheckVPCSubnetContainsLabel(&subnet1, "tf-label", "tf-label-value-a"),
					testAccCheckVPCSubnetContainsLabel(&subnet1, "empty-label", ""),
					testAccCheckCreatedAtAttr("yandex_vpc_subnet.subnet-a"),

					testAccCheckVPCSubnetExists("yandex_vpc_subnet.subnet-b", &subnet2),
					resource.TestCheckResourceAttrPtr("yandex_vpc_subnet.subnet-b", "network_id", &network.Id),
					resource.TestCheckResourceAttr("yandex_vpc_subnet.subnet-b", "name", subnet2Name),
					resource.TestCheckResourceAttr("yandex_vpc_subnet.subnet-b", "description", "description for subnet-b"),
					resource.TestCheckResourceAttr("yandex_vpc_subnet.subnet-b", "zone", "ru-central1-b"),
					resource.TestCheckResourceAttr("yandex_vpc_subnet.subnet-b", "v4_cidr_blocks.0", "10.1.0.0/16"),
					testAccCheckVPCSubnetContainsLabel(&subnet2, "tf-label", "tf-label-value-b"),
					testAccCheckVPCSubnetContainsLabel(&subnet2, "empty-label", ""),
					testAccCheckCreatedAtAttr("yandex_vpc_subnet.subnet-b"),
				),
			},
			{
				Config: testAccVPCSubnet_update(networkName, updatedSubnet1Name, updatedSubnet2Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCSubnetExists("yandex_vpc_subnet.subnet-a", &subnet1),
					resource.TestCheckResourceAttrPtr("yandex_vpc_subnet.subnet-a", "network_id", &network.Id),
					resource.TestCheckResourceAttr("yandex_vpc_subnet.subnet-a", "name", updatedSubnet1Name),
					resource.TestCheckResourceAttr("yandex_vpc_subnet.subnet-a", "v4_cidr_blocks.0", "10.100.0.0/16"),
					testAccCheckVPCSubnetContainsLabel(&subnet1, "empty-label", "oh-look-theres-a-label-now"),
					testAccCheckVPCSubnetContainsLabel(&subnet1, "new-field", "only-shows-up-when-updated"),

					testAccCheckVPCSubnetExists("yandex_vpc_subnet.subnet-b", &subnet2),
					resource.TestCheckResourceAttrPtr("yandex_vpc_subnet.subnet-b", "network_id", &network.Id),
					resource.TestCheckResourceAttr("yandex_vpc_subnet.subnet-b", "name", updatedSubnet2Name),
					resource.TestCheckResourceAttr("yandex_vpc_subnet.subnet-b", "v4_cidr_blocks.0", "10.101.0.0/16"),
					testAccCheckVPCSubnetContainsLabel(&subnet2, "empty-label", "oh-look-theres-a-label-now"),
					testAccCheckVPCSubnetContainsLabel(&subnet2, "new-field", "only-shows-up-when-updated"),
				),
			},
			{
				ResourceName:      "yandex_vpc_subnet.subnet-a",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "yandex_vpc_subnet.subnet-b",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCSubnet_withRouteTable(t *testing.T) {
	t.Parallel()

	var network vpc.Network
	var subnet vpc.Subnet

	networkName := acctest.RandomWithPrefix("tf-network")
	subnet1Name := acctest.RandomWithPrefix("tf-subnet-a")
	updatedSubnet1Name := subnet1Name + "-update"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVPCSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnet_withoutRouteTable(networkName, subnet1Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCNetworkExists("yandex_vpc_network.foo", &network),

					testAccCheckVPCSubnetExists("yandex_vpc_subnet.subnet-a", &subnet),
					resource.TestCheckResourceAttrPtr("yandex_vpc_subnet.subnet-a", "network_id", &network.Id),
					resource.TestCheckResourceAttr("yandex_vpc_subnet.subnet-a", "name", subnet1Name),
					resource.TestCheckResourceAttr("yandex_vpc_subnet.subnet-a", "description", "description for subnet-a"),
					resource.TestCheckResourceAttr("yandex_vpc_subnet.subnet-a", "zone", "ru-central1-a"),
					resource.TestCheckResourceAttr("yandex_vpc_subnet.subnet-a", "route_table_id", ""),
					resource.TestCheckResourceAttr("yandex_vpc_subnet.subnet-a", "v4_cidr_blocks.0", "10.0.0.0/16"),
					testAccCheckVPCSubnetContainsLabel(&subnet, "tf-label", "tf-label-value-a"),
					testAccCheckVPCSubnetContainsLabel(&subnet, "empty-label", ""),
					testAccCheckCreatedAtAttr("yandex_vpc_subnet.subnet-a"),
				),
			},
			{
				Config: testAccVPCSubnet_withRouteTable(networkName, updatedSubnet1Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCSubnetExists("yandex_vpc_subnet.subnet-a", &subnet),
					resource.TestCheckResourceAttrPtr("yandex_vpc_subnet.subnet-a", "network_id", &network.Id),
					resource.TestCheckResourceAttrPtr("yandex_vpc_subnet.subnet-a", "route_table_id", &subnet.RouteTableId),
					resource.TestCheckResourceAttr("yandex_vpc_subnet.subnet-a", "name", updatedSubnet1Name),
					resource.TestCheckResourceAttr("yandex_vpc_subnet.subnet-a", "v4_cidr_blocks.0", "10.100.0.0/16"),
					testAccCheckVPCSubnetContainsLabel(&subnet, "empty-label", "oh-look-theres-a-label-now"),
					testAccCheckVPCSubnetContainsLabel(&subnet, "new-field", "only-shows-up-when-updated"),
				),
			},
			{
				ResourceName:      "yandex_vpc_subnet.subnet-a",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCSubnet_basicV6(t *testing.T) {
	t.Skip("waiting ipv6 support in subnets")
	t.Parallel()

	var subnet1 vpc.Subnet
	var subnet2 vpc.Subnet

	cnName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	subnet1Name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	subnet2Name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVPCSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCSubnet_basicV6(cnName, subnet1Name, subnet2Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCSubnetExists(
						"yandex_vpc_subnet.subnet-a", &subnet1),
					testAccCheckVPCSubnetExists(
						"yandex_vpc_subnet.subnet-b", &subnet2),
					testAccCheckCreatedAtAttr("yandex_vpc_subnet.subnet-a"),
					testAccCheckCreatedAtAttr("yandex_vpc_subnet.subnet-b"),
				),
			},
			{
				ResourceName:      "yandex_vpc_subnet.subnet-a",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "yandex_vpc_subnet.subnet-b",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckVPCSubnetDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_vpc_subnet" {
			continue
		}

		_, err := config.sdk.VPC().Subnet().Get(context.Background(), &vpc.GetSubnetRequest{
			SubnetId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("Subnet still exists")
		}
	}

	return nil
}

func testAccCheckVPCSubnetExists(name string, subnet *vpc.Subnet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.VPC().Subnet().Get(context.Background(), &vpc.GetSubnetRequest{
			SubnetId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Subnet not found")
		}

		*subnet = *found

		return nil
	}
}

func testAccCheckVPCSubnetContainsLabel(subnet *vpc.Subnet, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		v, ok := subnet.Labels[key]
		if !ok {
			return fmt.Errorf("Expected label with key '%s' not found", key)
		}
		if v != value {
			return fmt.Errorf("Incorrect label value for key '%s': expected '%s' but found '%s'", key, value, v)
		}
		return nil
	}
}

//revive:disable:var-naming
func testAccVPCSubnet_basic(networkName, subnet1Name, subnet2Name string) string {
	return fmt.Sprintf(`
resource "yandex_vpc_network" "foo" {
  name = "%s"
}

resource "yandex_vpc_subnet" "subnet-a" {
  name           = "%s"
  description    = "description for subnet-a"
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.foo.id}"
  v4_cidr_blocks = ["10.0.0.0/16"]

  labels = {
    tf-label    = "tf-label-value-a"
    empty-label = ""
  }
}

resource "yandex_vpc_subnet" "subnet-b" {
  name           = "%s"
  description    = "description for subnet-b"
  zone           = "ru-central1-b"
  network_id     = "${yandex_vpc_network.foo.id}"
  v4_cidr_blocks = ["10.1.0.0/16"]

  labels = {
    tf-label    = "tf-label-value-b"
    empty-label = ""
  }
}
`, networkName, subnet1Name, subnet2Name)
}

func testAccVPCSubnet_basicV6(networkName, subnet1Name, subnet2Name string) string {
	return fmt.Sprintf(`
resource "yandex_vpc_network" "foo" {
  name = "%s"
}

resource "yandex_vpc_subnet" "subnet-a" {
  name           = "%s"
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.foo.id}"
  v4_cidr_blocks = ["10.0.0.0/16"]
  v6_cidr_blocks = ["fda9:8765:4321:1::/64"]
}

resource "yandex_vpc_subnet" "subnet-b" {
  name           = "%s"
  zone           = "ru-central1-b"
  network_id     = "${yandex_vpc_network.foo.id}"
  v4_cidr_blocks = ["10.1.0.0/16"]
  v6_cidr_blocks = ["fda9:8765:4321:2::/64"]
}
`, networkName, subnet1Name, subnet2Name)
}

func testAccVPCSubnet_update(networkName, subnet1Name, subnet2Name string) string {
	return fmt.Sprintf(`
resource "yandex_vpc_network" "foo" {
  name = "%s"
}

resource "yandex_vpc_subnet" "subnet-a" {
  name           = "%s"
  description    = "description with update for subnet-a"
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.foo.id}"
  v4_cidr_blocks = ["10.100.0.0/16"]

  labels = {
    empty-label = "oh-look-theres-a-label-now"
    new-field   = "only-shows-up-when-updated"
  }
}

resource "yandex_vpc_subnet" "subnet-b" {
  name           = "%s"
  description    = "description with update for subnet-b"
  zone           = "ru-central1-b"
  network_id     = "${yandex_vpc_network.foo.id}"
  v4_cidr_blocks = ["10.101.0.0/16"]

  labels = {
    empty-label = "oh-look-theres-a-label-now"
    new-field   = "only-shows-up-when-updated"
  }
}
`, networkName, subnet1Name, subnet2Name)
}

func testAccVPCSubnet_withoutRouteTable(networkName, subnet1Name string) string {
	return fmt.Sprintf(`
resource "yandex_vpc_network" "foo" {
  name = "%s"
}

resource "yandex_vpc_subnet" "subnet-a" {
  name           = "%s"
  description    = "description for subnet-a"
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.foo.id}"
  v4_cidr_blocks = ["10.0.0.0/16"]

  labels = {
    tf-label    = "tf-label-value-a"
    empty-label = ""
  }
}

resource "yandex_vpc_route_table" "rt-a" {
  network_id = "${yandex_vpc_network.foo.id}"

  static_route {
    destination_prefix = "172.16.10.0/24"
    next_hop_address   = "10.0.0.172"
  }
}
`, networkName, subnet1Name)
}

func testAccVPCSubnet_withRouteTable(networkName, subnet1Name string) string {
	return fmt.Sprintf(`
resource "yandex_vpc_network" "foo" {
  name = "%s"
}

resource "yandex_vpc_subnet" "subnet-a" {
  name           = "%s"
  description    = "description with update for subnet-a"
  zone           = "ru-central1-a"
  network_id     = "${yandex_vpc_network.foo.id}"
  route_table_id = "${yandex_vpc_route_table.rt-a.id}"
  v4_cidr_blocks = ["10.100.0.0/16"]

  labels = {
    empty-label = "oh-look-theres-a-label-now"
    new-field   = "only-shows-up-when-updated"
  }
}

resource "yandex_vpc_route_table" "rt-a" {
  network_id = "${yandex_vpc_network.foo.id}"

  static_route {
    destination_prefix = "172.16.10.0/24"
    next_hop_address   = "10.0.0.172"
  }
}
`, networkName, subnet1Name)
}
