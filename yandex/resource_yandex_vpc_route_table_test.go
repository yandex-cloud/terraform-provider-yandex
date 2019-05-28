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

func TestAccVPCRouteTable_basic(t *testing.T) {
	t.Parallel()

	var routeTable1 vpc.RouteTable
	var routeTable2 vpc.RouteTable

	networkName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	routeTable1Name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	routeTable2Name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVPCRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteTable_basic(networkName, routeTable1Name, routeTable2Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCRouteTableExists(
						"yandex_vpc_route_table.rt-a", &routeTable1),
					testAccCheckVPCRouteTableExists(
						"yandex_vpc_route_table.rt-b", &routeTable2),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-a", "static_route.85705717.destination_prefix", "10.0.0.0/16"),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-a", "static_route.85705717.next_hop_address", "10.0.0.10"),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-b", "static_route.3313653742.destination_prefix", "10.1.0.0/16"),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-b", "static_route.3313653742.next_hop_address", "10.1.0.10"),
					testAccCheckCreatedAtAttr("yandex_vpc_route_table.rt-a"),
					testAccCheckCreatedAtAttr("yandex_vpc_route_table.rt-b"),
				),
			},
			{
				ResourceName:      "yandex_vpc_route_table.rt-a",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "yandex_vpc_route_table.rt-b",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCRouteTable_update(t *testing.T) {
	t.Parallel()

	var network vpc.Network
	var routeTable1 vpc.RouteTable
	var routeTable2 vpc.RouteTable

	networkName := acctest.RandomWithPrefix("tf-network")
	routeTable1Name := acctest.RandomWithPrefix("tf-route-table-a")
	routeTable2Name := acctest.RandomWithPrefix("tf-route-table-b")
	updatedRouteTable1Name := routeTable1Name + "-update"
	updatedRouteTable2Name := routeTable2Name + "-update"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVPCRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteTable_basic(networkName, routeTable1Name, routeTable2Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCNetworkExists("yandex_vpc_network.foo", &network),

					testAccCheckVPCRouteTableExists("yandex_vpc_route_table.rt-a", &routeTable1),
					resource.TestCheckResourceAttrPtr("yandex_vpc_route_table.rt-a", "network_id", &network.Id),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-a", "name", routeTable1Name),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-a", "description", "description for route table A"),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-a", "static_route.85705717.destination_prefix", "10.0.0.0/16"),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-a", "static_route.85705717.next_hop_address", "10.0.0.10"),
					testAccCheckVPCRouteTableContainsLabel(&routeTable1, "tf-label", "tf-label-value-a"),
					testAccCheckVPCRouteTableContainsLabel(&routeTable1, "empty-label", ""),
					testAccCheckCreatedAtAttr("yandex_vpc_route_table.rt-a"),

					testAccCheckVPCRouteTableExists("yandex_vpc_route_table.rt-b", &routeTable2),
					resource.TestCheckResourceAttrPtr("yandex_vpc_route_table.rt-b", "network_id", &network.Id),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-b", "name", routeTable2Name),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-b", "description", "description for route table B"),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-b", "static_route.3313653742.destination_prefix", "10.1.0.0/16"),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-b", "static_route.3313653742.next_hop_address", "10.1.0.10"),
					testAccCheckVPCRouteTableContainsLabel(&routeTable2, "tf-label", "tf-label-value-b"),
					testAccCheckVPCRouteTableContainsLabel(&routeTable2, "empty-label", ""),
					testAccCheckCreatedAtAttr("yandex_vpc_route_table.rt-b"),
				),
			},
			{
				Config: testAccVPCRouteTable_update(networkName, updatedRouteTable1Name, updatedRouteTable2Name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCRouteTableExists("yandex_vpc_route_table.rt-a", &routeTable1),
					resource.TestCheckResourceAttrPtr("yandex_vpc_route_table.rt-a", "network_id", &network.Id),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-a", "name", updatedRouteTable1Name),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-a", "static_route.489959240.destination_prefix", "10.100.0.0/16"),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-a", "static_route.489959240.next_hop_address", "192.168.11.11"),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-a", "static_route.2381258344.destination_prefix", "10.101.0.0/16"),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-a", "static_route.2381258344.next_hop_address", "192.168.11.13"),
					testAccCheckVPCRouteTableContainsLabel(&routeTable1, "empty-label", "oh-look-theres-a-label-now"),
					testAccCheckVPCRouteTableContainsLabel(&routeTable1, "new-field", "only-shows-up-when-updated"),

					testAccCheckVPCRouteTableExists("yandex_vpc_route_table.rt-b", &routeTable2),
					resource.TestCheckResourceAttrPtr("yandex_vpc_route_table.rt-b", "network_id", &network.Id),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-b", "name", updatedRouteTable2Name),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-b", "static_route.2095193886.destination_prefix", "10.101.0.0/16"),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-b", "static_route.2095193886.next_hop_address", "192.168.22.22"),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-b", "static_route.385763903.destination_prefix", "10.102.0.0/16"),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-b", "static_route.385763903.next_hop_address", "192.168.22.24"),
					testAccCheckVPCRouteTableContainsLabel(&routeTable2, "empty-label", "oh-look-theres-a-label-now"),
					testAccCheckVPCRouteTableContainsLabel(&routeTable2, "new-field", "only-shows-up-when-updated"),
				),
			},
			{
				ResourceName:      "yandex_vpc_route_table.rt-a",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "yandex_vpc_route_table.rt-b",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckVPCRouteTableExists(name string, routeTable *vpc.RouteTable) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.VPC().RouteTable().Get(context.Background(), &vpc.GetRouteTableRequest{
			RouteTableId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Route table not found")
		}

		*routeTable = *found

		return nil
	}
}

func testAccCheckVPCRouteTableContainsLabel(routeTable *vpc.RouteTable, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		v, ok := routeTable.Labels[key]
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
func testAccVPCRouteTable_basic(networkName, routeTable1Name, routeTable2Name string) string {
	return fmt.Sprintf(`
resource "yandex_vpc_network" "foo" {
  name = "%s"
}

resource "yandex_vpc_route_table" "rt-a" {
  name        = "%s"
  description = "description for route table A"
  network_id  = "${yandex_vpc_network.foo.id}"

  static_route {
    destination_prefix = "10.0.0.0/16"
    next_hop_address   = "10.0.0.10"
  }

  labels = {
    tf-label    = "tf-label-value-a"
    empty-label = ""
  }
}

resource "yandex_vpc_route_table" "rt-b" {
  name        = "%s"
  description = "description for route table B"
  network_id  = "${yandex_vpc_network.foo.id}"

  static_route {
    destination_prefix = "10.1.0.0/16"
    next_hop_address   = "10.1.0.10"
  }

  labels = {
    tf-label    = "tf-label-value-b"
    empty-label = ""
  }
}
`, networkName, routeTable1Name, routeTable2Name)
}

func testAccVPCRouteTable_update(networkName, routeTable1Name, routeTable2Name string) string {
	return fmt.Sprintf(`
resource "yandex_vpc_network" "foo" {
  name = "%s"
}

resource "yandex_vpc_route_table" "rt-a" {
  name        = "%s"
  description = "description with update for route table A"
  network_id  = "${yandex_vpc_network.foo.id}"

  static_route {
    destination_prefix = "10.100.0.0/16"
    next_hop_address   = "192.168.11.11"
  }

  static_route {
    destination_prefix = "10.101.0.0/16"
    next_hop_address   = "192.168.11.13"
  }

  labels = {
    empty-label = "oh-look-theres-a-label-now"
    new-field   = "only-shows-up-when-updated"
  }
}

resource "yandex_vpc_route_table" "rt-b" {
  name        = "%s"
  description = "description with update for route table B"
  network_id  = "${yandex_vpc_network.foo.id}"

  static_route {
    destination_prefix = "10.101.0.0/16"
    next_hop_address   = "192.168.22.22"
  }

  static_route {
    destination_prefix = "10.102.0.0/16"
    next_hop_address   = "192.168.22.24"
  }

  labels = {
    empty-label = "oh-look-theres-a-label-now"
    new-field   = "only-shows-up-when-updated"
  }
}
`, networkName, routeTable1Name, routeTable2Name)
}

func testAccCheckVPCRouteTableDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_vpc_route_table" {
			continue
		}

		_, err := config.sdk.VPC().RouteTable().Get(context.Background(), &vpc.GetRouteTableRequest{
			RouteTableId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("Route table still exists")
		}
	}

	return nil
}
