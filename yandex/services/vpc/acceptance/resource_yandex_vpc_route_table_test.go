package acceptance

import (
	"context"
	"fmt"
	"testing"

	statecheck "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance/statecheck"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
	vpcsdk "github.com/yandex-cloud/go-sdk/services/vpc/v1"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/vpc/testfixtures"
)

func TestAccVPCRouteTable_basic(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	var routeTable1 vpc.RouteTable
	var routeTable2 vpc.RouteTable

	networkName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	routeTable1Name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	routeTable2Name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testAccCheckVPCRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteTable_basic(networkName, routeTable1Name, routeTable2Name),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckVPCRouteTableExists(
						"yandex_vpc_route_table.rt-a", &routeTable1),
					suite.testAccCheckVPCRouteTableExists(
						"yandex_vpc_route_table.rt-b", &routeTable2),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-a", "static_route.0.destination_prefix", "10.0.0.0/16"),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-a", "static_route.0.next_hop_address", "10.0.0.10"),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-b", "static_route.0.destination_prefix", "10.1.0.0/16"),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-b", "static_route.0.next_hop_address", "10.1.0.10"),
					statecheck.CheckCreatedAtAttr("yandex_vpc_route_table.rt-a"),
					statecheck.CheckCreatedAtAttr("yandex_vpc_route_table.rt-b"),
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
	suite := testSuite{acceptance.New(t)}
	var network vpc.Network
	var gateway vpc.Gateway
	var routeTable1 vpc.RouteTable
	var routeTable2 vpc.RouteTable

	networkName := acctest.RandomWithPrefix("tf-network")
	gatewayName := acctest.RandomWithPrefix("tf-gateway")
	routeTable1Name := acctest.RandomWithPrefix("tf-route-table-a")
	routeTable2Name := acctest.RandomWithPrefix("tf-route-table-b")
	updatedRouteTable1Name := routeTable1Name + "-update"
	updatedRouteTable2Name := routeTable2Name + "-update"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testAccCheckVPCRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteTable_basic(networkName, routeTable1Name, routeTable2Name),
				Check: resource.ComposeTestCheckFunc(
					testfixtures.CheckNetworkExists(suite.Config, "yandex_vpc_network.foo", &network),

					suite.testAccCheckVPCRouteTableExists("yandex_vpc_route_table.rt-a", &routeTable1),
					resource.TestCheckResourceAttrPtr("yandex_vpc_route_table.rt-a", "network_id", &network.Id),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-a", "name", routeTable1Name),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-a", "description", "description for route table A"),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-a", "static_route.0.destination_prefix", "10.0.0.0/16"),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-a", "static_route.0.next_hop_address", "10.0.0.10"),
					testAccCheckVPCRouteTableContainsLabel(&routeTable1, "tf-label", "tf-label-value-a"),
					testAccCheckVPCRouteTableContainsLabel(&routeTable1, "empty-label", ""),
					statecheck.CheckCreatedAtAttr("yandex_vpc_route_table.rt-a"),

					suite.testAccCheckVPCRouteTableExists("yandex_vpc_route_table.rt-b", &routeTable2),
					resource.TestCheckResourceAttrPtr("yandex_vpc_route_table.rt-b", "network_id", &network.Id),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-b", "name", routeTable2Name),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-b", "description", "description for route table B"),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-b", "static_route.0.destination_prefix", "10.1.0.0/16"),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-b", "static_route.0.next_hop_address", "10.1.0.10"),
					testAccCheckVPCRouteTableContainsLabel(&routeTable2, "tf-label", "tf-label-value-b"),
					testAccCheckVPCRouteTableContainsLabel(&routeTable2, "empty-label", ""),
					statecheck.CheckCreatedAtAttr("yandex_vpc_route_table.rt-b"),
				),
			},
			{
				Config: testAccVPCRouteTable_update(networkName, gatewayName, updatedRouteTable1Name, updatedRouteTable2Name),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckVPCGatewayExists("yandex_vpc_gateway.bar", &gateway),
					suite.testAccCheckVPCRouteTableExists("yandex_vpc_route_table.rt-a", &routeTable1),
					resource.TestCheckResourceAttrPtr("yandex_vpc_route_table.rt-a", "network_id", &network.Id),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-a", "name", updatedRouteTable1Name),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-a", "static_route.0.destination_prefix", "10.100.0.0/16"),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-a", "static_route.0.next_hop_address", "192.168.11.11"),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-a", "static_route.1.destination_prefix", "10.101.0.0/16"),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-a", "static_route.1.next_hop_address", "192.168.11.13"),
					testAccCheckVPCRouteTableContainsLabel(&routeTable1, "empty-label", "oh-look-theres-a-label-now"),
					testAccCheckVPCRouteTableContainsLabel(&routeTable1, "new-field", "only-shows-up-when-updated"),

					suite.testAccCheckVPCRouteTableExists("yandex_vpc_route_table.rt-b", &routeTable2),
					resource.TestCheckResourceAttrPtr("yandex_vpc_route_table.rt-b", "network_id", &network.Id),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-b", "name", updatedRouteTable2Name),
					// reverse route order because TypeSet...
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-b", "static_route.1.destination_prefix", "10.101.0.0/16"),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-b", "static_route.1.next_hop_address", "192.168.22.22"),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-b", "static_route.0.destination_prefix", "0.0.0.0/0"),
					resource.TestCheckResourceAttrPtr("yandex_vpc_route_table.rt-b", "static_route.0.gateway_id", &gateway.Id),
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

func TestAccVPCRouteTable_updateGatewayToNextHopAddress(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	var network vpc.Network
	var gateway vpc.Gateway
	var routeTable1 vpc.RouteTable
	var routeTable2 vpc.RouteTable

	networkName := acctest.RandomWithPrefix("tf-network")
	gatewayName := acctest.RandomWithPrefix("tf-gateway")
	routeTable1Name := acctest.RandomWithPrefix("tf-route-table-a")
	routeTable2Name := acctest.RandomWithPrefix("tf-route-table-b")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testAccCheckVPCRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteTable_update(networkName, gatewayName, routeTable1Name, routeTable2Name),
				Check: resource.ComposeTestCheckFunc(
					testfixtures.CheckNetworkExists(suite.Config, "yandex_vpc_network.foo", &network),
					suite.testAccCheckVPCGatewayExists("yandex_vpc_gateway.bar", &gateway),

					suite.testAccCheckVPCRouteTableExists("yandex_vpc_route_table.rt-a", &routeTable1),
					resource.TestCheckResourceAttrPtr("yandex_vpc_route_table.rt-a", "network_id", &network.Id),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-a", "name", routeTable1Name),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-a", "static_route.0.destination_prefix", "10.100.0.0/16"),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-a", "static_route.0.next_hop_address", "192.168.11.11"),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-a", "static_route.1.destination_prefix", "10.101.0.0/16"),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-a", "static_route.1.next_hop_address", "192.168.11.13"),
					testAccCheckVPCRouteTableContainsLabel(&routeTable1, "empty-label", "oh-look-theres-a-label-now"),
					testAccCheckVPCRouteTableContainsLabel(&routeTable1, "new-field", "only-shows-up-when-updated"),

					suite.testAccCheckVPCRouteTableExists("yandex_vpc_route_table.rt-b", &routeTable2),
					resource.TestCheckResourceAttrPtr("yandex_vpc_route_table.rt-b", "network_id", &network.Id),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-b", "name", routeTable2Name),
					// reverse route order because TypeSet...
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-b", "static_route.1.destination_prefix", "10.101.0.0/16"),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-b", "static_route.1.next_hop_address", "192.168.22.22"),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-b", "static_route.0.destination_prefix", "0.0.0.0/0"),
					resource.TestCheckResourceAttrPtr("yandex_vpc_route_table.rt-b", "static_route.0.gateway_id", &gateway.Id),
					testAccCheckVPCRouteTableContainsLabel(&routeTable2, "empty-label", "oh-look-theres-a-label-now"),
					testAccCheckVPCRouteTableContainsLabel(&routeTable2, "new-field", "only-shows-up-when-updated"),
				),
			},
			{
				// change gateway_id to next_hop_address
				Config: testAccVPCRouteTable_updateGatewayToNextHopAddress(networkName, gatewayName, routeTable1Name, routeTable2Name),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckVPCGatewayExists("yandex_vpc_gateway.bar", &gateway),
					suite.testAccCheckVPCRouteTableExists("yandex_vpc_route_table.rt-a", &routeTable1),
					resource.TestCheckResourceAttrPtr("yandex_vpc_route_table.rt-a", "network_id", &network.Id),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-a", "name", routeTable1Name),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-a", "static_route.0.destination_prefix", "10.100.0.0/16"),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-a", "static_route.0.next_hop_address", "192.168.11.11"),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-a", "static_route.1.destination_prefix", "10.101.0.0/16"),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-a", "static_route.1.next_hop_address", "192.168.11.13"),
					testAccCheckVPCRouteTableContainsLabel(&routeTable1, "empty-label", "oh-look-theres-a-label-now"),
					testAccCheckVPCRouteTableContainsLabel(&routeTable1, "new-field", "only-shows-up-when-updated"),

					suite.testAccCheckVPCRouteTableExists("yandex_vpc_route_table.rt-b", &routeTable2),
					resource.TestCheckResourceAttrPtr("yandex_vpc_route_table.rt-b", "network_id", &network.Id),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-b", "name", routeTable2Name),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-b", "static_route.0.destination_prefix", "0.0.0.0/0"),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-b", "static_route.0.next_hop_address", "192.168.22.33"),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-b", "static_route.1.destination_prefix", "10.101.0.0/16"),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-b", "static_route.1.next_hop_address", "192.168.22.22"),
					testAccCheckVPCRouteTableContainsLabel(&routeTable2, "empty-label", "oh-look-theres-a-label-now"),
					testAccCheckVPCRouteTableContainsLabel(&routeTable2, "new-field", "only-shows-up-when-updated"),
				),
			},
			{
				// change next_hop_address back to gateway_id
				Config: testAccVPCRouteTable_update(networkName, gatewayName, routeTable1Name, routeTable2Name),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckVPCGatewayExists("yandex_vpc_gateway.bar", &gateway),
					suite.testAccCheckVPCRouteTableExists("yandex_vpc_route_table.rt-a", &routeTable1),
					resource.TestCheckResourceAttrPtr("yandex_vpc_route_table.rt-a", "network_id", &network.Id),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-a", "name", routeTable1Name),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-a", "static_route.0.destination_prefix", "10.100.0.0/16"),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-a", "static_route.0.next_hop_address", "192.168.11.11"),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-a", "static_route.1.destination_prefix", "10.101.0.0/16"),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-a", "static_route.1.next_hop_address", "192.168.11.13"),
					testAccCheckVPCRouteTableContainsLabel(&routeTable1, "empty-label", "oh-look-theres-a-label-now"),
					testAccCheckVPCRouteTableContainsLabel(&routeTable1, "new-field", "only-shows-up-when-updated"),

					suite.testAccCheckVPCRouteTableExists("yandex_vpc_route_table.rt-b", &routeTable2),
					resource.TestCheckResourceAttrPtr("yandex_vpc_route_table.rt-b", "network_id", &network.Id),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-b", "name", routeTable2Name),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-b", "static_route.0.destination_prefix", "0.0.0.0/0"),
					resource.TestCheckResourceAttrPtr("yandex_vpc_route_table.rt-b", "static_route.0.gateway_id", &gateway.Id),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-b", "static_route.1.destination_prefix", "10.101.0.0/16"),
					resource.TestCheckResourceAttr("yandex_vpc_route_table.rt-b", "static_route.1.next_hop_address", "192.168.22.22"),
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
				ResourceName:            "yandex_vpc_route_table.rt-b",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"static_route.1.gateway_id"},
			},
		},
	})
}

func (suite testSuite) testAccCheckVPCRouteTableExists(name string, routeTable *vpc.RouteTable) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}

		client := vpcsdk.NewRouteTableClient(config.SDK)

		found, err := client.Get(context.Background(), &vpc.GetRouteTableRequest{
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

func testAccVPCRouteTable_update(networkName, gatewayName, routeTable1Name, routeTable2Name string) string {
	return fmt.Sprintf(`
resource "yandex_vpc_network" "foo" {
  name = "%s"
}

resource "yandex_vpc_gateway" "bar" {
  name = "%s"
  shared_egress_gateway {}
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
    destination_prefix = "0.0.0.0/0"
    gateway_id         = "${yandex_vpc_gateway.bar.id}"
  }

  labels = {
    empty-label = "oh-look-theres-a-label-now"
    new-field   = "only-shows-up-when-updated"
  }
}
`, networkName, gatewayName, routeTable1Name, routeTable2Name)
}

func testAccVPCRouteTable_updateGatewayToNextHopAddress(networkName, gatewayName, routeTable1Name, routeTable2Name string) string {
	return fmt.Sprintf(`
resource "yandex_vpc_network" "foo" {
  name = "%s"
}

resource "yandex_vpc_gateway" "bar" {
  name = "%s"
  shared_egress_gateway {}
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
    destination_prefix = "0.0.0.0/0"
    next_hop_address   = "192.168.22.33"
  }

  labels = {
    empty-label = "oh-look-theres-a-label-now"
    new-field   = "only-shows-up-when-updated"
  }
}
`, networkName, gatewayName, routeTable1Name, routeTable2Name)
}

func (suite testSuite) testAccCheckVPCRouteTableDestroy(s *terraform.State) error {
	config, configErr := suite.Config()
	if configErr != nil {
		return configErr
	}

	client := vpcsdk.NewRouteTableClient(config.SDK)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_vpc_route_table" {
			continue
		}

		_, err := client.Get(context.Background(), &vpc.GetRouteTableRequest{
			RouteTableId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("Route table still exists")
		}
	}

	return nil
}
