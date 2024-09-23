package yandex

import (
	"context"
	"fmt"
	"testing"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1/privatelink"
)

func init() {
	resource.AddTestSweepers("yandex_vpc_private_endpoint", &resource.Sweeper{
		Name:         "yandex_vpc_private_endpoint",
		F:            testSweepVPCPrivateEndpoints,
		Dependencies: []string{},
	})
}

func testSweepVPCPrivateEndpoints(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &privatelink.ListPrivateEndpointsRequest{
		Container: &privatelink.ListPrivateEndpointsRequest_FolderId{
			FolderId: conf.FolderID,
		},
	}
	it := conf.sdk.VPCPrivateLink().PrivateEndpoint().PrivateEndpointIterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepVPCPrivateEndpoint(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep VPC Private Endpoint %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepVPCPrivateEndpoint(conf *Config, id string) bool {
	return sweepWithRetry(sweepVPCPrivateEndpointOnce, conf, "VPC Private Endpoint", id)
}

func sweepVPCPrivateEndpointOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexVPCPrivateEndpointDefaultTimeout)
	defer cancel()

	op, err := conf.sdk.VPCPrivateLink().PrivateEndpoint().Delete(ctx, &privatelink.DeletePrivateEndpointRequest{
		PrivateEndpointId: id,
	})

	return handleSweepOperation(ctx, conf, op, err)
}

func TestAccVPCPrivateEndpoint_Basic(t *testing.T) {
	t.Parallel()

	networkName := acctest.RandomWithPrefix("tf-network")
	subnetName := acctest.RandomWithPrefix("tf-subnet")
	peName := acctest.RandomWithPrefix("tf-private-endpoint")
	peNewName := peName + "-updated"

	var pe privatelink.PrivateEndpoint

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVPCPrivateEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPrivateEndpointConfigBasic(networkName, subnetName, peName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCPrivateEndpointExists("yandex_vpc_private_endpoint.pe", &pe),
					resource.TestCheckResourceAttr("yandex_vpc_private_endpoint.pe", "name", peName),
					resource.TestCheckResourceAttrSet("yandex_vpc_private_endpoint.pe", "folder_id"),
					testAccCheckCreatedAtAttr("yandex_vpc_private_endpoint.pe"),
					resource.TestCheckResourceAttr("yandex_vpc_private_endpoint.pe", "dns_options.#", "1"),
					resource.TestCheckResourceAttr("yandex_vpc_private_endpoint.pe", "dns_options.0.private_dns_records_enabled", "false"),
					resource.TestCheckResourceAttr("yandex_vpc_private_endpoint.pe", "endpoint_address.#", "1"),
					resource.TestCheckResourceAttrSet("yandex_vpc_private_endpoint.pe", "endpoint_address.0.subnet_id"),
					resource.TestCheckResourceAttrSet("yandex_vpc_private_endpoint.pe", "endpoint_address.0.address"),
					resource.TestCheckResourceAttrSet("yandex_vpc_private_endpoint.pe", "endpoint_address.0.address_id"),
					testAccCheckVPCPrivateEndpointContainsLabel(&pe, "tf-label", "tf-label-value"),
					testAccCheckVPCPrivateEndpointContainsLabel(&pe, "empty-label", ""),
				),
			},
			{
				Config: testAccVPCPrivateEndpointConfigDnsOptions(networkName, subnetName, peNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCPrivateEndpointExists("yandex_vpc_private_endpoint.pe", &pe),
					resource.TestCheckResourceAttr("yandex_vpc_private_endpoint.pe", "name", peNewName),
					resource.TestCheckResourceAttrSet("yandex_vpc_private_endpoint.pe", "folder_id"),
					testAccCheckCreatedAtAttr("yandex_vpc_private_endpoint.pe"),
					resource.TestCheckResourceAttr("yandex_vpc_private_endpoint.pe", "dns_options.#", "1"),
					resource.TestCheckResourceAttr("yandex_vpc_private_endpoint.pe", "dns_options.0.private_dns_records_enabled", "true"),
					resource.TestCheckResourceAttr("yandex_vpc_private_endpoint.pe", "endpoint_address.#", "1"),
					resource.TestCheckResourceAttrSet("yandex_vpc_private_endpoint.pe", "endpoint_address.0.subnet_id"),
					resource.TestCheckResourceAttrSet("yandex_vpc_private_endpoint.pe", "endpoint_address.0.address"),
					resource.TestCheckResourceAttrSet("yandex_vpc_private_endpoint.pe", "endpoint_address.0.address_id"),
				),
			},
		},
	})
}

func TestAccVPCPrivateEndpoint_SpecificAddress(t *testing.T) {
	t.Parallel()

	networkName := acctest.RandomWithPrefix("tf-network")
	subnetName := acctest.RandomWithPrefix("tf-subnet")
	peName := acctest.RandomWithPrefix("tf-private-endpoint")

	var pe privatelink.PrivateEndpoint

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVPCPrivateEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPrivateEndpointConfigAddressSpec(networkName, subnetName, peName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCPrivateEndpointExists("yandex_vpc_private_endpoint.pe", &pe),
					resource.TestCheckResourceAttr("yandex_vpc_private_endpoint.pe", "name", peName),
					resource.TestCheckResourceAttrSet("yandex_vpc_private_endpoint.pe", "folder_id"),
					testAccCheckCreatedAtAttr("yandex_vpc_private_endpoint.pe"),
					resource.TestCheckResourceAttr("yandex_vpc_private_endpoint.pe", "dns_options.#", "1"),
					resource.TestCheckResourceAttrSet("yandex_vpc_private_endpoint.pe", "dns_options.0.private_dns_records_enabled"),
					resource.TestCheckResourceAttr("yandex_vpc_private_endpoint.pe", "endpoint_address.#", "1"),
					resource.TestCheckResourceAttrSet("yandex_vpc_private_endpoint.pe", "endpoint_address.0.subnet_id"),
					resource.TestCheckResourceAttrSet("yandex_vpc_private_endpoint.pe", "endpoint_address.0.address"),
					resource.TestCheckResourceAttrSet("yandex_vpc_private_endpoint.pe", "endpoint_address.0.address_id"),
				),
			},
		},
	})
}

func testAccCheckVPCPrivateEndpointContainsLabel(pe *privatelink.PrivateEndpoint, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		v, ok := pe.Labels[key]
		if !ok {
			return fmt.Errorf("expected label with key '%s' not found", key)
		}
		if v != value {
			return fmt.Errorf("incorrect label value for key '%s': expected '%s' but found '%s'", key, value, v)
		}
		return nil
	}
}

func testAccCheckVPCPrivateEndpointDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_vpc_private_endpoint" {
			continue
		}

		_, err := config.sdk.VPCPrivateLink().PrivateEndpoint().Get(context.Background(), &privatelink.GetPrivateEndpointRequest{
			PrivateEndpointId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("Private endpoint still exists")
		}
	}

	return nil
}

func testAccCheckVPCPrivateEndpointExists(name string, pe *privatelink.PrivateEndpoint) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.VPCPrivateLink().PrivateEndpoint().Get(context.Background(), &privatelink.GetPrivateEndpointRequest{
			PrivateEndpointId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Subnet not found")
		}

		*pe = *found

		return nil
	}
}

func testAccVPCPrivateEndpointConfigBasic(networkName, subnetName, peName string) string {
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

resource "yandex_vpc_private_endpoint" "pe" {
  name       = "%s"
  network_id = yandex_vpc_network.foo.id

  labels = {
    tf-label    = "tf-label-value"
    empty-label = ""
  }

  object_storage {}

  depends_on = [
    yandex_vpc_subnet.subnet-a
  ]
}
`, networkName, subnetName, peName)
}

func testAccVPCPrivateEndpointConfigDnsOptions(networkName, subnetName, peName string) string {
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

resource "yandex_vpc_private_endpoint" "pe" {
  name       = "%s"
  network_id = yandex_vpc_network.foo.id

  object_storage {}

  dns_options {
    private_dns_records_enabled = true
  }

  depends_on = [
    yandex_vpc_subnet.subnet-a
  ]
}
`, networkName, subnetName, peName)
}

func testAccVPCPrivateEndpointConfigAddressSpec(networkName, subnetName, peName string) string {
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

resource "yandex_vpc_private_endpoint" "pe" {
  name       = "%s"
  network_id = yandex_vpc_network.foo.id

  object_storage {}

  endpoint_address {
    subnet_id = yandex_vpc_subnet.subnet-a.id
  }

  dns_options {
    private_dns_records_enabled = true
  }
}
`, networkName, subnetName, peName)
}
