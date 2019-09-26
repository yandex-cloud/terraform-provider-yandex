package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
)

func TestAccVPCNetwork_basic(t *testing.T) {
	t.Parallel()

	var network vpc.Network
	networkName := acctest.RandomWithPrefix("tf-network")
	networkDesc := "Network description for test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVPCNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetwork_basic(networkName, networkDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCNetworkExists("yandex_vpc_network.foo", &network),
					resource.TestCheckResourceAttr("yandex_vpc_network.foo", "name", networkName),
					resource.TestCheckResourceAttr("yandex_vpc_network.foo", "description", networkDesc),
					resource.TestCheckResourceAttr("yandex_vpc_network.foo", "subnet_ids.#", "0"),
					resource.TestCheckResourceAttrSet("yandex_vpc_network.foo", "folder_id"),
					testAccCheckVPCNetworkContainsLabel(&network, "tf-label", "tf-label-value"),
					testAccCheckVPCNetworkContainsLabel(&network, "empty-label", ""),
					testAccCheckCreatedAtAttr("yandex_vpc_network.foo"),
				),
			},
			{
				ResourceName:      "yandex_vpc_network.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCNetwork_update(t *testing.T) {
	t.Parallel()

	var network vpc.Network
	networkName := acctest.RandomWithPrefix("tf-network")
	networkDesc := "Network description for test"
	updatedNetworkName := networkName + "-update"
	updatedNetworkDesc := networkDesc + " with update"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVPCNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetwork_basic(networkName, networkDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCNetworkExists("yandex_vpc_network.foo", &network),
					resource.TestCheckResourceAttr("yandex_vpc_network.foo", "name", networkName),
					resource.TestCheckResourceAttr("yandex_vpc_network.foo", "description", networkDesc),
					resource.TestCheckResourceAttr("yandex_vpc_network.foo", "subnet_ids.#", "0"),
					resource.TestCheckResourceAttrSet("yandex_vpc_network.foo", "folder_id"),
					testAccCheckCreatedAtAttr("yandex_vpc_network.foo"),
				),
			},
			{
				Config: testAccVPCNetwork_update(updatedNetworkName, updatedNetworkDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCNetworkExists("yandex_vpc_network.foo", &network),
					resource.TestCheckResourceAttr("yandex_vpc_network.foo", "name", updatedNetworkName),
					resource.TestCheckResourceAttr("yandex_vpc_network.foo", "description", updatedNetworkDesc),
					resource.TestCheckResourceAttr("yandex_vpc_network.foo", "subnet_ids.#", "0"),
					resource.TestCheckResourceAttrSet("yandex_vpc_network.foo", "folder_id"),
					testAccCheckCreatedAtAttr("yandex_vpc_network.foo"),
				),
			},
			{
				ResourceName:      "yandex_vpc_network.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
func TestAccVPCNetwork_addSubnets(t *testing.T) {
	t.Parallel()

	var network vpc.Network
	networkName := acctest.RandomWithPrefix("tf-network")
	networkDesc := "Network description for test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVPCNetworkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetwork_basic(networkName, networkDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCNetworkExists("yandex_vpc_network.foo", &network),
					resource.TestCheckResourceAttr("yandex_vpc_network.foo", "name", networkName),
					resource.TestCheckResourceAttr("yandex_vpc_network.foo", "description", networkDesc),
					resource.TestCheckResourceAttr("yandex_vpc_network.foo", "subnet_ids.#", "0"),
					resource.TestCheckResourceAttrSet("yandex_vpc_network.foo", "folder_id"),
					testAccCheckCreatedAtAttr("yandex_vpc_network.foo"),
				),
			},
			// Add two subnets to test network. The list of network subnets is read before two new subnets
			// created. So changes in 'subnet_ids' are checked in next step.
			{
				Config: testAccVPCNetwork_addSubnets(networkName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCNetworkExists("yandex_vpc_network.foo", &network),
					resource.TestCheckResourceAttr("yandex_vpc_network.foo", "name", networkName),
					resource.TestCheckResourceAttr("yandex_vpc_network.foo", "description", ""),
					resource.TestCheckResourceAttr("yandex_vpc_network.foo", "subnet_ids.#", "0"),
				),
			},
			{
				Config: testAccVPCNetwork_addSubnets(networkName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCNetworkExists("yandex_vpc_network.foo", &network),
					resource.TestCheckResourceAttr("yandex_vpc_network.foo", "subnet_ids.#", "2"),
				),
			},
			{
				ResourceName:      "yandex_vpc_network.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckVPCNetworkDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_vpc_network" {
			continue
		}

		_, err := config.sdk.VPC().Network().Get(context.Background(), &vpc.GetNetworkRequest{
			NetworkId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("Network still exists")
		}
	}

	return nil
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

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.VPC().Network().Get(context.Background(), &vpc.GetNetworkRequest{
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

func testAccCheckVPCNetworkContainsLabel(network *vpc.Network, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		v, ok := network.Labels[key]
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
func testAccVPCNetwork_basic(name, description string) string {
	return fmt.Sprintf(`
resource "yandex_vpc_network" "foo" {
  name        = "%s"
  description = "%s"

  labels = {
    tf-label    = "tf-label-value"
    empty-label = ""
  }
}
`, name, description)
}

func testAccVPCNetwork_update(name, description string) string {
	return fmt.Sprintf(`
resource "yandex_vpc_network" "foo" {
  name        = "%s"
  description = "%s"

  labels = {
    empty-label = "oh-look-theres-a-label-now"
    new-field   = "only-shows-up-when-updated"
  }
}
`, name, description)
}

func testAccVPCNetwork_addSubnets(name, description string) string {
	return fmt.Sprintf(`
resource "yandex_vpc_network" "foo" {
  name        = "%s"
  description = "%s"

  labels = {
    empty-label = "oh-look-theres-a-label-now"
    new-field   = "only-shows-up-when-updated"
  }
}

resource "yandex_vpc_subnet" "bar1" {
  network_id     = "${yandex_vpc_network.foo.id}"
  v4_cidr_blocks = ["172.16.1.0/24"]
}

resource "yandex_vpc_subnet" "bar2" {
  network_id     = "${yandex_vpc_network.foo.id}"
  v4_cidr_blocks = ["172.16.2.0/24"]
}
`, name, description)
}
