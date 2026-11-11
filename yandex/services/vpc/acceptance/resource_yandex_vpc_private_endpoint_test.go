package acceptance

import (
	"context"
	"fmt"
	"testing"

	statecheck "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance/statecheck"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1/privatelink"
	privatelinksdk "github.com/yandex-cloud/go-sdk/services/vpc/v1/privatelink"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
)

func TestAccVPCPrivateEndpoint_Basic(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	networkName := acctest.RandomWithPrefix("tf-network")
	subnetName := acctest.RandomWithPrefix("tf-subnet")
	peName := acctest.RandomWithPrefix("tf-private-endpoint")
	peNewName := peName + "-updated"

	var pe privatelink.PrivateEndpoint

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testAccCheckVPCPrivateEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPrivateEndpointConfigBasic(networkName, subnetName, peName),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckVPCPrivateEndpointExists("yandex_vpc_private_endpoint.pe", &pe),
					resource.TestCheckResourceAttr("yandex_vpc_private_endpoint.pe", "name", peName),
					resource.TestCheckResourceAttrSet("yandex_vpc_private_endpoint.pe", "folder_id"),
					statecheck.CheckCreatedAtAttr("yandex_vpc_private_endpoint.pe"),
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
					suite.testAccCheckVPCPrivateEndpointExists("yandex_vpc_private_endpoint.pe", &pe),
					resource.TestCheckResourceAttr("yandex_vpc_private_endpoint.pe", "name", peNewName),
					resource.TestCheckResourceAttrSet("yandex_vpc_private_endpoint.pe", "folder_id"),
					statecheck.CheckCreatedAtAttr("yandex_vpc_private_endpoint.pe"),
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
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	networkName := acctest.RandomWithPrefix("tf-network")
	subnetName := acctest.RandomWithPrefix("tf-subnet")
	peName := acctest.RandomWithPrefix("tf-private-endpoint")

	var pe privatelink.PrivateEndpoint

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testAccCheckVPCPrivateEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPrivateEndpointConfigAddressSpec(networkName, subnetName, peName),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckVPCPrivateEndpointExists("yandex_vpc_private_endpoint.pe", &pe),
					resource.TestCheckResourceAttr("yandex_vpc_private_endpoint.pe", "name", peName),
					resource.TestCheckResourceAttrSet("yandex_vpc_private_endpoint.pe", "folder_id"),
					statecheck.CheckCreatedAtAttr("yandex_vpc_private_endpoint.pe"),
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

func (suite testSuite) testAccCheckVPCPrivateEndpointDestroy(s *terraform.State) error {
	config, configErr := suite.Config()
	if configErr != nil {
		return configErr
	}

	client := privatelinksdk.NewPrivateEndpointClient(config.SDK)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_vpc_private_endpoint" {
			continue
		}

		_, err := client.Get(context.Background(), &privatelink.GetPrivateEndpointRequest{
			PrivateEndpointId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("Private endpoint still exists")
		}
	}

	return nil
}

func (suite testSuite) testAccCheckVPCPrivateEndpointExists(name string, pe *privatelink.PrivateEndpoint) resource.TestCheckFunc {
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

		client := privatelinksdk.NewPrivateEndpointClient(config.SDK)

		found, err := client.Get(context.Background(), &privatelink.GetPrivateEndpointRequest{
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

func TestAccVPCPrivateEndpoint_DnsRecords(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	networkName := acctest.RandomWithPrefix("tf-network")
	subnetName := acctest.RandomWithPrefix("tf-subnet")
	peName := acctest.RandomWithPrefix("tf-private-endpoint")

	var pe privatelink.PrivateEndpoint

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testAccCheckVPCPrivateEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPrivateEndpointConfigDnsOptions(networkName, subnetName, peName),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckVPCPrivateEndpointExists("yandex_vpc_private_endpoint.pe", &pe),
					resource.TestCheckResourceAttr("yandex_vpc_private_endpoint.pe", "dns_options.#", "1"),
					resource.TestCheckResourceAttr("yandex_vpc_private_endpoint.pe", "dns_options.0.private_dns_records_enabled", "true"),
					resource.TestCheckResourceAttrSet("yandex_vpc_private_endpoint.pe", "dns_records.#"),
					resource.TestCheckResourceAttrSet("yandex_vpc_private_endpoint.pe", "dns_records.0.name"),
				),
			},
		},
	})
}

func TestAccVPCPrivateEndpoint_ServiceName(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	networkName := acctest.RandomWithPrefix("tf-network")
	subnetName := acctest.RandomWithPrefix("tf-subnet")
	peName := acctest.RandomWithPrefix("tf-private-endpoint")

	const serviceName = "yandex.cloud.storage"

	var pe privatelink.PrivateEndpoint

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testAccCheckVPCPrivateEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPrivateEndpointConfigServiceName(networkName, subnetName, peName, serviceName),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckVPCPrivateEndpointExists("yandex_vpc_private_endpoint.pe", &pe),
					resource.TestCheckResourceAttr("yandex_vpc_private_endpoint.pe", "name", peName),
					resource.TestCheckResourceAttrSet("yandex_vpc_private_endpoint.pe", "folder_id"),
					statecheck.CheckCreatedAtAttr("yandex_vpc_private_endpoint.pe"),
					resource.TestCheckResourceAttr("yandex_vpc_private_endpoint.pe", "service_name", serviceName),
					resource.TestCheckResourceAttr("yandex_vpc_private_endpoint.pe", "endpoint_address.#", "1"),
					resource.TestCheckResourceAttrSet("yandex_vpc_private_endpoint.pe", "endpoint_address.0.subnet_id"),
					resource.TestCheckResourceAttrSet("yandex_vpc_private_endpoint.pe", "endpoint_address.0.address"),
					resource.TestCheckResourceAttrSet("yandex_vpc_private_endpoint.pe", "endpoint_address.0.address_id"),
				),
			},
		},
	})
}

func testAccVPCPrivateEndpointConfigServiceName(networkName, subnetName, peName, serviceName string) string {
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
}

resource "yandex_vpc_private_endpoint" "pe" {
  name       = "%s"
  network_id = yandex_vpc_network.foo.id

  service_name = "%s"

  endpoint_address {
    subnet_id = yandex_vpc_subnet.subnet-a.id
  }
}
`, networkName, subnetName, peName, serviceName)
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
