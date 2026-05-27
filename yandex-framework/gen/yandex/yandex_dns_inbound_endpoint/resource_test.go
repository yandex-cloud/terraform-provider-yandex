package yandex_dns_inbound_endpoint_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	dns "github.com/yandex-cloud/go-genproto/yandex/cloud/dns/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
	dnsv1sdk "github.com/yandex-cloud/go-sdk/services/dns/v1"
	vpcv1sdk "github.com/yandex-cloud/go-sdk/services/vpc/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestAccDNSInboundEndpoint_basic(t *testing.T) {
	var (
		folderID = test.GetExampleFolderID()
		name     = fmt.Sprintf("tf-test-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckDnsInboundEndpointDestroy,
		Steps: []resource.TestStep{
			{
				// Step 1: Create VPC network and subnet first
				Config: testAccDnsInboundEndpointVPCOnly(folderID, name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCNetworkExists("yandex_vpc_network.network1", &vpc.Network{}),
				),
			},
			{
				// Step 2: Create address and DNS inbound endpoint
				Config: testAccDnsInboundEndpointBasic(folderID, name, "test description", "label-value"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDnsInboundEndpointExists("yandex_dns_inbound_endpoint.test"),
					resource.TestCheckResourceAttr("yandex_dns_inbound_endpoint.test", "name", name),
					resource.TestCheckResourceAttr("yandex_dns_inbound_endpoint.test", "description", "test description"),
					resource.TestCheckResourceAttr("yandex_dns_inbound_endpoint.test", "folder_id", folderID),
					resource.TestCheckResourceAttr("yandex_dns_inbound_endpoint.test", "labels.test-label", "label-value"),
					resource.TestCheckResourceAttrSet("yandex_dns_inbound_endpoint.test", "id"),
					resource.TestCheckResourceAttrSet("yandex_dns_inbound_endpoint.test", "created_at"),
					resource.TestCheckResourceAttrSet("yandex_dns_inbound_endpoint.test", "address"),
					resource.TestCheckResourceAttrSet("yandex_vpc_address.addr1", "id"),
					resource.TestCheckResourceAttrSet("yandex_vpc_address.addr1", "internal_ipv4_address.0.address"),
				),
			},
			{
				ResourceName:      "yandex_dns_inbound_endpoint.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDNSInboundEndpoint_update(t *testing.T) {
	var (
		folderID = test.GetExampleFolderID()
		name     = fmt.Sprintf("tf-test-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckDnsInboundEndpointDestroy,
		Steps: []resource.TestStep{
			{
				// Step 1: Create VPC network and subnet first
				// This allows time for DNS service to learn about the new VPC network
				Config: testAccDnsInboundEndpointVPCOnly(folderID, name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCNetworkExists("yandex_vpc_network.network1", &vpc.Network{}),
				),
			},
			{
				// Step 2: Create DNS inbound endpoint with initial values
				Config: testAccDnsInboundEndpointBasic(folderID, name, "initial description", "initial-value"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDnsInboundEndpointExists("yandex_dns_inbound_endpoint.test"),
					resource.TestCheckResourceAttr("yandex_dns_inbound_endpoint.test", "description", "initial description"),
					resource.TestCheckResourceAttr("yandex_dns_inbound_endpoint.test", "labels.test-label", "initial-value"),
				),
			},
			{
				Config: testAccDnsInboundEndpointUpdated(folderID, name, "updated description", "updated-value"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDnsInboundEndpointExists("yandex_dns_inbound_endpoint.test"),
					resource.TestCheckResourceAttr("yandex_dns_inbound_endpoint.test", "description", "updated description"),
					resource.TestCheckResourceAttr("yandex_dns_inbound_endpoint.test", "labels.test-label", "updated-value"),
				),
			},
			{
				ResourceName:      "yandex_dns_inbound_endpoint.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDNSInboundEndpoint_recreateWithAddressChange(t *testing.T) {
	var (
		folderID = test.GetExampleFolderID()
		name     = fmt.Sprintf("tf-test-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckDnsInboundEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDnsInboundEndpointWithTwoSubnets(folderID, name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCNetworkExists("yandex_vpc_network.network1", &vpc.Network{}),
				),
			},
			{
				Config: testAccDnsInboundEndpointWithFirstAddress(folderID, name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDnsInboundEndpointExists("yandex_dns_inbound_endpoint.test"),
					resource.TestCheckResourceAttr("yandex_dns_inbound_endpoint.test", "name", name),
					resource.TestCheckResourceAttrSet("yandex_vpc_address.addr1", "id"),
					resource.TestCheckResourceAttrSet("yandex_vpc_address.addr1", "internal_ipv4_address.0.address"),
				),
			},
			{
				Config: testAccDnsInboundEndpointWithSecondAddress(folderID, name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDnsInboundEndpointExists("yandex_dns_inbound_endpoint.test"),
					resource.TestCheckResourceAttr("yandex_dns_inbound_endpoint.test", "name", name),
					resource.TestCheckResourceAttrSet("yandex_vpc_address.addr2", "id"),
					resource.TestCheckResourceAttrSet("yandex_vpc_address.addr2", "internal_ipv4_address.0.address"),
				),
			},
			{
				ResourceName:      "yandex_dns_inbound_endpoint.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDNSInboundEndpoint_deletionProtection(t *testing.T) {
	var (
		folderID = test.GetExampleFolderID()
		name     = fmt.Sprintf("tf-test-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckDnsInboundEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDnsInboundEndpointVPCOnly(folderID, name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCNetworkExists("yandex_vpc_network.network1", &vpc.Network{}),
				),
			},
			{
				Config: testAccDnsInboundEndpointDeletionProtection(folderID, name, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDnsInboundEndpointExists("yandex_dns_inbound_endpoint.test"),
					resource.TestCheckResourceAttr("yandex_dns_inbound_endpoint.test", "deletion_protection", "true"),
				),
			},
			{
				Config: testAccDnsInboundEndpointDeletionProtection(folderID, name, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDnsInboundEndpointExists("yandex_dns_inbound_endpoint.test"),
					resource.TestCheckResourceAttr("yandex_dns_inbound_endpoint.test", "deletion_protection", "false"),
				),
			},
			{
				ResourceName:      "yandex_dns_inbound_endpoint.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDnsInboundEndpointDestroy(s *terraform.State) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_dns_inbound_endpoint" {
			continue
		}

		_, err := dnsv1sdk.NewDnsInboundEndpointClient(config.SDKv2).Get(
			context.Background(),
			&dns.GetDnsInboundEndpointRequest{DnsInboundEndpointId: rs.Primary.ID},
		)

		if err == nil {
			return fmt.Errorf("DNS Inbound Endpoint %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckDnsInboundEndpointExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set for %s", resourceName)
		}

		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

		found, err := dnsv1sdk.NewDnsInboundEndpointClient(config.SDKv2).Get(
			context.Background(),
			&dns.GetDnsInboundEndpointRequest{DnsInboundEndpointId: rs.Primary.ID},
		)
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("DNS Inbound Endpoint %s not found", resourceName)
		}

		return nil
	}
}

func testAccDnsInboundEndpointBasic(folderID, name, description, labelValue string) string {
	return fmt.Sprintf(`
resource "yandex_vpc_network" "network1" {
  name = "%[1]s-network"
}

resource "yandex_vpc_subnet" "subnet1" {
  name           = "%[1]s-subnet"
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.network1.id
  v4_cidr_blocks = ["192.168.0.0/24"]
}

resource "yandex_vpc_address" "addr1" {
  name        = "%[1]s-addr"
  description = "internal address for DNS inbound endpoint"

  labels = {
    test-label = "test-value"
  }

  internal_ipv4_address {
    subnet_id = yandex_vpc_subnet.subnet1.id
  }
  deletion_protection = false
}

resource "yandex_dns_inbound_endpoint" "test" {
  folder_id   = "%[2]s"
  name        = "%[1]s"
  description = "%[3]s"
  network_id  = yandex_vpc_network.network1.id
  address_id  = yandex_vpc_address.addr1.id

  labels = {
    test-label = "%[4]s"
  }
}
`, name, folderID, description, labelValue)
}

func testAccDnsInboundEndpointUpdated(folderID, name, description, labelValue string) string {
	return fmt.Sprintf(`
resource "yandex_vpc_network" "network1" {
  name = "%[1]s-network"
}

resource "yandex_vpc_subnet" "subnet1" {
  name           = "%[1]s-subnet"
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.network1.id
  v4_cidr_blocks = ["192.168.0.0/24"]
}

resource "yandex_vpc_address" "addr1" {
  name        = "%[1]s-addr"
  description = "internal address for DNS inbound endpoint"

  labels = {
    test-label = "test-value"
  }

  internal_ipv4_address {
    subnet_id = yandex_vpc_subnet.subnet1.id
  }
  deletion_protection = false
}

resource "yandex_dns_inbound_endpoint" "test" {
  folder_id   = "%[2]s"
  name        = "%[1]s"
  description = "%[3]s"
  network_id  = yandex_vpc_network.network1.id
  address_id  = yandex_vpc_address.addr1.id

  labels = {
    test-label = "%[4]s"
  }
}
`, name, folderID, description, labelValue)
}

func testAccDnsInboundEndpointWithFirstAddress(folderID, name string) string {
	return fmt.Sprintf(`
resource "yandex_vpc_network" "network1" {
  name = "%[1]s-network"
}

resource "yandex_vpc_subnet" "subnet1" {
  name           = "%[1]s-subnet1"
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.network1.id
  v4_cidr_blocks = ["192.168.0.0/24"]
}

resource "yandex_vpc_address" "addr1" {
  name        = "%[1]s-addr1"
  description = "first internal address for DNS inbound endpoint"

  internal_ipv4_address {
    subnet_id = yandex_vpc_subnet.subnet1.id
  }
  deletion_protection = false
}

resource "yandex_dns_inbound_endpoint" "test" {
  folder_id  = "%[2]s"
  name       = "%[1]s"
  network_id = yandex_vpc_network.network1.id
  address_id = yandex_vpc_address.addr1.id
}
`, name, folderID)
}

func testAccDnsInboundEndpointWithSecondAddress(folderID, name string) string {
	return fmt.Sprintf(`
resource "yandex_vpc_network" "network1" {
  name = "%[1]s-network"
}

resource "yandex_vpc_subnet" "subnet1" {
  name           = "%[1]s-subnet1"
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.network1.id
  v4_cidr_blocks = ["192.168.0.0/24"]
}

resource "yandex_vpc_subnet" "subnet2" {
  name           = "%[1]s-subnet2"
  zone           = "ru-central1-b"
  network_id     = yandex_vpc_network.network1.id
  v4_cidr_blocks = ["192.168.1.0/24"]
}

resource "yandex_vpc_address" "addr1" {
  name        = "%[1]s-addr1"
  description = "first internal address for DNS inbound endpoint"

  internal_ipv4_address {
    subnet_id = yandex_vpc_subnet.subnet1.id
  }
  deletion_protection = false
}

resource "yandex_vpc_address" "addr2" {
  name        = "%[1]s-addr2"
  description = "second internal address for DNS inbound endpoint"

  internal_ipv4_address {
    subnet_id = yandex_vpc_subnet.subnet2.id
  }
  deletion_protection = false
}

resource "yandex_dns_inbound_endpoint" "test" {
  folder_id  = "%[2]s"
  name       = "%[1]s"
  network_id = yandex_vpc_network.network1.id
  address_id = yandex_vpc_address.addr2.id
}
`, name, folderID)
}

func testAccDnsInboundEndpointDeletionProtection(folderID, name string, deletionProtection bool) string {
	return fmt.Sprintf(`
resource "yandex_vpc_network" "network1" {
  name = "%[1]s-network"
}

resource "yandex_vpc_subnet" "subnet1" {
  name           = "%[1]s-subnet"
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.network1.id
  v4_cidr_blocks = ["192.168.0.0/24"]
}

resource "yandex_vpc_address" "addr1" {
  name        = "%[1]s-addr"
  description = "internal address for DNS inbound endpoint"

  internal_ipv4_address {
    subnet_id = yandex_vpc_subnet.subnet1.id
  }
  deletion_protection = false
}

resource "yandex_dns_inbound_endpoint" "test" {
  folder_id           = "%[2]s"
  name                = "%[1]s"
  network_id          = yandex_vpc_network.network1.id
  address_id          = yandex_vpc_address.addr1.id
  deletion_protection = %[3]t
}
`, name, folderID, deletionProtection)
}

func testAccDnsInboundEndpointVPCOnly(folderID, name string) string {
	return fmt.Sprintf(`
resource "yandex_vpc_network" "network1" {
 name = "%[1]s-network"
}

resource "yandex_vpc_subnet" "subnet1" {
 name           = "%[1]s-subnet"
 zone           = "ru-central1-a"
 network_id     = yandex_vpc_network.network1.id
 v4_cidr_blocks = ["192.168.0.0/24"]
}
`, name)
}

func testAccDnsInboundEndpointWithTwoSubnets(folderID, name string) string {
	return fmt.Sprintf(`
resource "yandex_vpc_network" "network1" {
  name = "%[1]s-network"
}

resource "yandex_vpc_subnet" "subnet1" {
  name           = "%[1]s-subnet1"
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.network1.id
  v4_cidr_blocks = ["192.168.0.0/24"]
}

resource "yandex_vpc_subnet" "subnet2" {
  name           = "%[1]s-subnet2"
  zone           = "ru-central1-b"
  network_id     = yandex_vpc_network.network1.id
  v4_cidr_blocks = ["192.168.1.0/24"]
}
`, name)
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

		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

		found, err := vpcv1sdk.NewNetworkClient(config.SDKv2).Get(context.Background(), &vpc.GetNetworkRequest{
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
