package yandex

import (
	"context"
	"fmt"
	"testing"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
)

func init() {
	resource.AddTestSweepers("yandex_vpc_network", &resource.Sweeper{
		Name: "yandex_vpc_network",
		F:    testSweepVPCNetworks,
		Dependencies: []string{
			"yandex_vpc_subnet",
			"yandex_vpc_route_table",
			"yandex_vpc_security_group",
		},
	})
}

func testSweepVPCNetworks(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &vpc.ListNetworksRequest{FolderId: conf.FolderID}
	it := conf.sdk.VPC().Network().NetworkIterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepVPCNetwork(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep VPC network %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepVPCNetwork(conf *Config, id string) bool {
	return sweepWithRetry(sweepVPCNetworkOnce, conf, "VPC Network", id)
}

func sweepVPCNetworkOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexVPCNetworkDefaultTimeout)
	defer cancel()

	req := &vpc.ListNetworkSubnetsRequest{NetworkId: id}
	subIt := conf.sdk.VPC().Network().NetworkSubnetsIterator(conf.Context(), req)
	for subIt.Next() {
		subID := subIt.Value().GetId()
		err := sweepVPCSubnetOnce(conf, subID)
		if err != nil {
			return err
		}
	}

	op, err := conf.sdk.VPC().Network().Delete(ctx, &vpc.DeleteNetworkRequest{
		NetworkId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}

// NOTE(dxan): function may return non-empty string and non-nil error. Example:
// Resource is successfully created, but wait fails: the function returns id and wait error
func createVPCNetworkForSweeper(conf *Config) (string, error) {
	ctx, cancel := conf.ContextWithTimeout(yandexVPCNetworkDefaultTimeout)
	defer cancel()
	op, err := conf.sdk.WrapOperation(conf.sdk.VPC().Network().Create(ctx, &vpc.CreateNetworkRequest{
		FolderId:    conf.FolderID,
		Name:        acctest.RandomWithPrefix("sweeper"),
		Description: "created by sweeper",
	}))
	if err != nil {
		return "", fmt.Errorf("failed to create network: %v", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return "", fmt.Errorf("failed to get metadata from create network operation: %v", err)
	}

	md, ok := protoMetadata.(*vpc.CreateNetworkMetadata)
	if !ok {
		return "", fmt.Errorf("failed to get Network ID from create operation metadata")
	}
	debugLog("Network '%s' was created, waiting for complete operation '%s'", md.GetNetworkId(), op.Id())

	err = op.Wait(ctx)
	if err != nil {
		return "", fmt.Errorf("error while waiting for create subnet operation: %v", err)
	}

	return md.NetworkId, nil
}

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
					resource.TestCheckResourceAttrSet("yandex_vpc_network.foo", "default_security_group_id"),
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
