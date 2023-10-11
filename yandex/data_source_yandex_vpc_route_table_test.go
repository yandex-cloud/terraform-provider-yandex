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

func TestAccDataSourceVPCRouteTable(t *testing.T) {
	t.Parallel()

	routeTableName1 := acctest.RandomWithPrefix("tf-route-table")
	routeTableDesc1 := "Description for test route table"

	folderID := getExampleFolderID()
	var network vpc.Network

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckVPCNetworkDestroy,
			testAccCheckVPCRouteTableDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVPCRouteTableConfig(routeTableName1, routeTableDesc1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCNetworkExists("yandex_vpc_network.foo", &network),

					testAccDataSourceVPCRouteTableExists("data.yandex_vpc_route_table.bar1"),

					testAccCheckResourceIDField("data.yandex_vpc_route_table.bar1", "route_table_id"),
					resource.TestCheckResourceAttr("data.yandex_vpc_route_table.bar1", "name", routeTableName1),
					resource.TestCheckResourceAttr("data.yandex_vpc_route_table.bar1", "description", routeTableDesc1),
					resource.TestCheckResourceAttr("data.yandex_vpc_route_table.bar1", "folder_id", folderID),
					resource.TestCheckResourceAttr("data.yandex_vpc_route_table.bar1", "static_route.0.destination_prefix", "192.168.20.0/24"),
					resource.TestCheckResourceAttr("data.yandex_vpc_route_table.bar1", "static_route.0.next_hop_address", "192.168.22.22"),
					resource.TestCheckResourceAttrSet("data.yandex_vpc_route_table.bar1", "network_id"),
					testAccCheckCreatedAtAttr("data.yandex_vpc_route_table.bar1"),
				),
			},
		},
	})
}

func testAccDataSourceVPCRouteTableExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ds, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if ds.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.VPC().RouteTable().Get(context.Background(), &vpc.GetRouteTableRequest{
			RouteTableId: ds.Primary.ID,
		})

		if err != nil {
			return err
		}

		if found.Id != ds.Primary.ID {
			return fmt.Errorf("Route table not found")
		}

		return nil
	}
}

//revive:disable:var-naming
func testAccDataSourceVPCRouteTableConfig(name1, desc1 string) string {
	return fmt.Sprintf(`
data "yandex_vpc_route_table" "bar1" {
  route_table_id = "${yandex_vpc_route_table.foo1.id}"
}

resource "yandex_vpc_network" "foo" {
  name        = "%s"
  description = "description for test"
}

resource "yandex_vpc_route_table" "foo1" {
  name        = "%s"
  network_id  = "${yandex_vpc_network.foo.id}"
  description = "%s"

  static_route {
    destination_prefix = "192.168.20.0/24"
    next_hop_address   = "192.168.22.22"
  }
}
`, acctest.RandomWithPrefix("tf-network"), name1, desc1)
}
