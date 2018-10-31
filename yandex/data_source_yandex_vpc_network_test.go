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

func TestAccDataSourceVPCNetwork(t *testing.T) {
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
				Config: testAccDataSourceVPCNetworkConfig(networkName, networkDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceVPCNetworkExists("data.yandex_vpc_network.bar", &network),
					resource.TestCheckResourceAttr("data.yandex_vpc_network.bar", "name", networkName),
					resource.TestCheckResourceAttr("data.yandex_vpc_network.bar", "description", networkDesc),
					resource.TestCheckResourceAttr("data.yandex_vpc_network.bar", "folder_id", folderID),
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

//revive:disable:var-naming
func testAccDataSourceVPCNetworkConfig(name, desc string) string {
	return fmt.Sprintf(`
data "yandex_vpc_network" "bar" {
	network_id = "${yandex_vpc_network.foo.id}"
}

resource "yandex_vpc_network" "foo" {
	name        = "%s"
	description = "%s"
}`, name, desc)
}
