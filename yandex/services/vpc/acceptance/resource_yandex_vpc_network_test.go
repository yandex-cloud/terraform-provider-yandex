package acceptance

import (
	"fmt"
	"testing"

	vpcfixtures "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/vpc/testfixtures"

	statecheck "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance/statecheck"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
)

func TestAccVPCNetwork_basic(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	var network vpc.Network
	networkName := acctest.RandomWithPrefix("tf-network")
	networkDesc := "Network description for test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: vpcfixtures.CheckNetworkDestroy(suite.Config),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetwork_basic(networkName, networkDesc),
				Check: resource.ComposeTestCheckFunc(
					vpcfixtures.CheckNetworkExists(suite.Config, "yandex_vpc_network.foo", &network),
					resource.TestCheckResourceAttr("yandex_vpc_network.foo", "name", networkName),
					resource.TestCheckResourceAttr("yandex_vpc_network.foo", "description", networkDesc),
					resource.TestCheckResourceAttr("yandex_vpc_network.foo", "subnet_ids.#", "0"),
					resource.TestCheckResourceAttrSet("yandex_vpc_network.foo", "folder_id"),
					resource.TestCheckResourceAttrSet("yandex_vpc_network.foo", "default_security_group_id"),
					testAccCheckVPCNetworkContainsLabel(&network, "tf-label", "tf-label-value"),
					testAccCheckVPCNetworkContainsLabel(&network, "empty-label", ""),
					statecheck.CheckCreatedAtAttr("yandex_vpc_network.foo"),
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
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	var network vpc.Network
	networkName := acctest.RandomWithPrefix("tf-network")
	networkDesc := "Network description for test"
	updatedNetworkName := networkName + "-update"
	updatedNetworkDesc := networkDesc + " with update"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: vpcfixtures.CheckNetworkDestroy(suite.Config),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetwork_basic(networkName, networkDesc),
				Check: resource.ComposeTestCheckFunc(
					vpcfixtures.CheckNetworkExists(suite.Config, "yandex_vpc_network.foo", &network),
					resource.TestCheckResourceAttr("yandex_vpc_network.foo", "name", networkName),
					resource.TestCheckResourceAttr("yandex_vpc_network.foo", "description", networkDesc),
					resource.TestCheckResourceAttr("yandex_vpc_network.foo", "subnet_ids.#", "0"),
					resource.TestCheckResourceAttrSet("yandex_vpc_network.foo", "folder_id"),
					statecheck.CheckCreatedAtAttr("yandex_vpc_network.foo"),
				),
			},
			{
				Config: testAccVPCNetwork_update(updatedNetworkName, updatedNetworkDesc),
				Check: resource.ComposeTestCheckFunc(
					vpcfixtures.CheckNetworkExists(suite.Config, "yandex_vpc_network.foo", &network),
					resource.TestCheckResourceAttr("yandex_vpc_network.foo", "name", updatedNetworkName),
					resource.TestCheckResourceAttr("yandex_vpc_network.foo", "description", updatedNetworkDesc),
					resource.TestCheckResourceAttr("yandex_vpc_network.foo", "subnet_ids.#", "0"),
					resource.TestCheckResourceAttrSet("yandex_vpc_network.foo", "folder_id"),
					statecheck.CheckCreatedAtAttr("yandex_vpc_network.foo"),
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
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	var network vpc.Network
	networkName := acctest.RandomWithPrefix("tf-network")
	networkDesc := "Network description for test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: vpcfixtures.CheckNetworkDestroy(suite.Config),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNetwork_basic(networkName, networkDesc),
				Check: resource.ComposeTestCheckFunc(
					vpcfixtures.CheckNetworkExists(suite.Config, "yandex_vpc_network.foo", &network),
					resource.TestCheckResourceAttr("yandex_vpc_network.foo", "name", networkName),
					resource.TestCheckResourceAttr("yandex_vpc_network.foo", "description", networkDesc),
					resource.TestCheckResourceAttr("yandex_vpc_network.foo", "subnet_ids.#", "0"),
					resource.TestCheckResourceAttrSet("yandex_vpc_network.foo", "folder_id"),
					statecheck.CheckCreatedAtAttr("yandex_vpc_network.foo"),
				),
			},
			// Add two subnets to test network. The list of network subnets is read before two new subnets
			// created. So changes in 'subnet_ids' are checked in next step.
			{
				Config: testAccVPCNetwork_addSubnets(networkName, ""),
				Check: resource.ComposeTestCheckFunc(
					vpcfixtures.CheckNetworkExists(suite.Config, "yandex_vpc_network.foo", &network),
					resource.TestCheckResourceAttr("yandex_vpc_network.foo", "name", networkName),
					resource.TestCheckResourceAttr("yandex_vpc_network.foo", "description", ""),
					resource.TestCheckResourceAttr("yandex_vpc_network.foo", "subnet_ids.#", "0"),
				),
			},
			{
				Config: testAccVPCNetwork_addSubnets(networkName, ""),
				Check: resource.ComposeTestCheckFunc(
					vpcfixtures.CheckNetworkExists(suite.Config, "yandex_vpc_network.foo", &network),
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
