package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/dns/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
)

func TestAccDNSZone_basic(t *testing.T) {
	t.Parallel()

	var zone dns.DnsZone
	var net1, net2, net3 vpc.Network
	zoneName := acctest.RandomWithPrefix("tf-dns-zone")
	fqdn := acctest.RandomWithPrefix("tf-test") + ".dnstest.test."

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDnsZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDNSZoneBasic(zoneName, fqdn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDNSZoneExists("yandex_dns_zone.zone1", &zone),
					testAccCheckVPCNetworkExists("yandex_vpc_network.net1", &net1),
					testAccCheckVPCNetworkExists("yandex_vpc_network.net2", &net2),
					testAccCheckVPCNetworkExists("yandex_vpc_network.net3", &net3),
					resource.TestCheckResourceAttrSet("yandex_dns_zone.zone1", "folder_id"),
					resource.TestCheckResourceAttr("yandex_dns_zone.zone1", "zone", fqdn),
					resource.TestCheckResourceAttr("yandex_dns_zone.zone1", "name", zoneName),
					resource.TestCheckResourceAttr("yandex_dns_zone.zone1", "description", "desc"),
					testAccCheckDnsZoneLabel(&zone, "tf-label", "tf-label-value"),
					testAccCheckDnsZoneLabel(&zone, "empty-label", ""),
					testAccCheckDnsZoneNetwork(&zone, &net1, true),
					testAccCheckDnsZoneNetwork(&zone, &net2, true),
				),
			},
			{
				ResourceName:      "yandex_dns_zone.zone1",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDNSZone_update(t *testing.T) {
	t.Parallel()

	var zone dns.DnsZone
	var net1, net2, net3 vpc.Network
	zoneName := acctest.RandomWithPrefix("tf-dns-zone")
	fqdn := acctest.RandomWithPrefix("tf-test") + ".dnstest.test."

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDnsZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDNSZoneBasic(zoneName, fqdn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDNSZoneExists("yandex_dns_zone.zone1", &zone),
					testAccCheckVPCNetworkExists("yandex_vpc_network.net1", &net1),
					testAccCheckVPCNetworkExists("yandex_vpc_network.net2", &net2),
					testAccCheckVPCNetworkExists("yandex_vpc_network.net3", &net3),
					resource.TestCheckResourceAttrSet("yandex_dns_zone.zone1", "folder_id"),
					resource.TestCheckResourceAttr("yandex_dns_zone.zone1", "zone", fqdn),
					resource.TestCheckResourceAttr("yandex_dns_zone.zone1", "name", zoneName),
					resource.TestCheckResourceAttr("yandex_dns_zone.zone1", "description", "desc"),
					testAccCheckDnsZoneLabel(&zone, "tf-label", "tf-label-value"),
					testAccCheckDnsZoneLabel(&zone, "empty-label", ""),
					testAccCheckDnsZoneNetwork(&zone, &net1, true),
					testAccCheckDnsZoneNetwork(&zone, &net2, true),
				),
			},
			{
				Config: testAccDNSZoneUpdate1(zoneName, fqdn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDNSZoneExists("yandex_dns_zone.zone1", &zone),
					testAccCheckVPCNetworkExists("yandex_vpc_network.net1", &net1),
					testAccCheckVPCNetworkExists("yandex_vpc_network.net2", &net2),
					testAccCheckVPCNetworkExists("yandex_vpc_network.net3", &net3),
					resource.TestCheckResourceAttrSet("yandex_dns_zone.zone1", "folder_id"),
					resource.TestCheckResourceAttr("yandex_dns_zone.zone1", "zone", fqdn),
					resource.TestCheckResourceAttr("yandex_dns_zone.zone1", "name", zoneName),
					resource.TestCheckResourceAttr("yandex_dns_zone.zone1", "description", "desc1"),
					testAccCheckDnsZoneLabel(&zone, "tf-label", "tf-label-value1"),
					testAccCheckDnsZoneNetwork(&zone, &net1, true),
					testAccCheckDnsZoneNetwork(&zone, &net2, false),
					testAccCheckDnsZoneNetwork(&zone, &net3, true),
				),
			},
			{
				ResourceName:      "yandex_dns_zone.zone1",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDNSZoneVisibility_update(t *testing.T) {
	t.Parallel()

	var zone dns.DnsZone
	var net1 vpc.Network
	var rs dns.RecordSet
	zoneName := acctest.RandomWithPrefix("tf-dns-zone")
	fqdn := acctest.RandomWithPrefix("tf-test") + ".dnstest.test."

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDnsZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDNSZoneBasicPublic(zoneName, fqdn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDNSZoneExists("yandex_dns_zone.zone1", &zone),
					testAccCheckVPCNetworkExists("yandex_vpc_network.net1", &net1),
					testAccCheckDNSRecordSetExists("yandex_dns_recordset.rs1", &rs),
					testAccCheckDnsZoneIsPublic(&zone),
				),
			},
			{
				Config: testAccDNSZoneBasicPrivate(zoneName, fqdn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDNSZoneExists("yandex_dns_zone.zone1", &zone),
					testAccCheckVPCNetworkExists("yandex_vpc_network.net1", &net1),
					testAccCheckDNSRecordSetExists("yandex_dns_recordset.rs1", &rs),
					testAccCheckDnsZoneIsPrivate(&zone),
					testAccCheckDnsZoneNetwork(&zone, &net1, true),
				),
			},
			{
				Config: testAccDNSZoneBasicPublicPrivate(zoneName, fqdn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDNSZoneExists("yandex_dns_zone.zone1", &zone),
					testAccCheckVPCNetworkExists("yandex_vpc_network.net1", &net1),
					testAccCheckDNSRecordSetExists("yandex_dns_recordset.rs1", &rs),
					testAccCheckDnsZoneIsPublicPrivate(&zone),
					testAccCheckDnsZoneNetwork(&zone, &net1, true),
				),
			},
			{
				Config: testAccDNSZoneBasicPrivate(zoneName, fqdn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDNSZoneExists("yandex_dns_zone.zone1", &zone),
					testAccCheckVPCNetworkExists("yandex_vpc_network.net1", &net1),
					testAccCheckDNSRecordSetExists("yandex_dns_recordset.rs1", &rs),
					testAccCheckDnsZoneIsPrivate(&zone),
					testAccCheckDnsZoneNetwork(&zone, &net1, true),
				),
			},
			{
				Config: testAccDNSZoneBasicPublic(zoneName, fqdn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDNSZoneExists("yandex_dns_zone.zone1", &zone),
					testAccCheckVPCNetworkExists("yandex_vpc_network.net1", &net1),
					testAccCheckDNSRecordSetExists("yandex_dns_recordset.rs1", &rs),
					testAccCheckDnsZoneIsPublic(&zone),
				),
			},
			{
				Config: testAccDNSZoneBasicPublicPrivate(zoneName, fqdn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDNSZoneExists("yandex_dns_zone.zone1", &zone),
					testAccCheckVPCNetworkExists("yandex_vpc_network.net1", &net1),
					testAccCheckDNSRecordSetExists("yandex_dns_recordset.rs1", &rs),
					testAccCheckDnsZoneIsPublicPrivate(&zone),
					testAccCheckDnsZoneNetwork(&zone, &net1, true),
				),
			},
			{
				ResourceName:      "yandex_dns_zone.zone1",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDNSZone_deletionProtection(t *testing.T) {
	t.Parallel()

	var zone dns.DnsZone
	zoneName := acctest.RandomWithPrefix("tf-dns-zone")
	fqdn := acctest.RandomWithPrefix("tf-test") + ".dnstest.test."

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDnsZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDNSZoneDeletionProtectionOn(zoneName, fqdn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDNSZoneExists("yandex_dns_zone.zone1", &zone),
					resource.TestCheckResourceAttr("yandex_dns_zone.zone1", "deletion_protection", "true"),
				),
			},
			{
				Config: testAccDNSZoneDeletionProtectionOff(zoneName, fqdn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDNSZoneExists("yandex_dns_zone.zone1", &zone),
					resource.TestCheckResourceAttr("yandex_dns_zone.zone1", "deletion_protection", "false"),
				),
			},
			{
				ResourceName:      "yandex_dns_zone.zone1",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDNSZoneExists(name string, zone *dns.DnsZone) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		sdk := getSDK(testAccProvider.Meta().(*Config))
		found, err := sdk.DNS().DnsZone().Get(context.Background(), &dns.GetDnsZoneRequest{
			DnsZoneId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("address not found")
		}

		//goland:noinspection GoVetCopyLock
		*zone = *found

		return nil
	}
}

func testAccCheckDnsZoneLabel(z *dns.DnsZone, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if z.Labels == nil {
			return fmt.Errorf("no labels found on dns zone %s", z.Name)
		}

		if v, ok := z.Labels[key]; ok {
			if v != value {
				return fmt.Errorf("expected value '%s' but found value '%s' for label '%s' on dns zone %s", value, v, key, z.Name)
			}
		} else {
			return fmt.Errorf("no label found with key %s on dns zone %s", key, z.Name)
		}

		return nil
	}
}

func testAccCheckDnsZoneNetwork(z *dns.DnsZone, network *vpc.Network, isPresent bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if z.PrivateVisibility == nil {
			return fmt.Errorf("no private visibility in dns zone %s", z.Name)
		}

		var found bool
		for _, n := range z.PrivateVisibility.NetworkIds {
			if n == network.Id {
				found = true
				break
			}
		}

		if found != isPresent {
			return fmt.Errorf("invalid presence for network with id \"%s\" in dns zone %s", network.Id, z.Name)
		}

		return nil
	}
}

func testAccCheckDnsZoneIsPublic(z *dns.DnsZone) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if z.PrivateVisibility != nil {
			return fmt.Errorf("public dns zone %s has private visibility", z.Name)
		}

		return nil
	}
}

func testAccCheckDnsZoneIsPrivate(z *dns.DnsZone) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if z.PublicVisibility != nil {
			return fmt.Errorf("private dns zone %s has public visibility", z.Name)
		}

		return nil
	}
}

func testAccCheckDnsZoneIsPublicPrivate(z *dns.DnsZone) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if z.PublicVisibility == nil || z.PrivateVisibility == nil {
			return fmt.Errorf("public-private dns zone %s do not have public ov private visibility", z.Name)
		}

		return nil
	}
}

func testAccDNSZoneBasic(name, fqdn string) string {
	return fmt.Sprintf(`
resource "yandex_vpc_network" "net1" {}
resource "yandex_vpc_network" "net2" {}
resource "yandex_vpc_network" "net3" {}

resource "yandex_dns_zone" "zone1" {
  name        = "%s"
  description = "desc"

  labels = {
    tf-label    = "tf-label-value"
    empty-label = ""
  }

  zone             = "%s"
  public           = true
  private_networks = [yandex_vpc_network.net1.id, yandex_vpc_network.net2.id]
}
`, name, fqdn)
}

func testAccDNSZoneDeletionProtectionOn(name, fqdn string) string {
	return fmt.Sprintf(`
resource "yandex_dns_zone" "zone1" {
  name        = "%s"
  description = "desc"

  labels = {
    tf-label    = "tf-label-value"
    empty-label = ""
  }

  zone             = "%s"
  public           = true

  deletion_protection = true
}
`, name, fqdn)
}

func testAccDNSZoneDeletionProtectionOff(name, fqdn string) string {
	return fmt.Sprintf(`
resource "yandex_dns_zone" "zone1" {
  name        = "%s"
  description = "desc"

  labels = {
    tf-label    = "tf-label-value"
    empty-label = ""
  }

  zone             = "%s"
  public           = true
}
`, name, fqdn)
}

func testAccDNSZoneBasicPublic(name, fqdn string) string {
	return fmt.Sprintf(`
resource "yandex_vpc_network" "net1" {}

resource "yandex_dns_zone" "zone1" {
  name        = "%[1]s"
  description = "desc"

  labels = {
    tf-label    = "tf-label-value"
    empty-label = ""
  }

  zone             = "%[2]s"
  public           = true
}

resource "yandex_dns_recordset" "rs1" {
  zone_id = yandex_dns_zone.zone1.id
  name    = "srv.%[2]s"
  type    = "A"
  ttl     = 200
  data    = ["10.1.0.1"]
}
`, name, fqdn)
}

func testAccDNSZoneBasicPrivate(name, fqdn string) string {
	return fmt.Sprintf(`
resource "yandex_vpc_network" "net1" {}

resource "yandex_dns_zone" "zone1" {
  name        = "%[1]s"
  description = "desc"

  labels = {
    tf-label    = "tf-label-value"
    empty-label = ""
  }

  zone             = "%[2]s"
  public           = false
  private_networks = [yandex_vpc_network.net1.id]
}

resource "yandex_dns_recordset" "rs1" {
  zone_id = yandex_dns_zone.zone1.id
  name    = "srv.%[2]s"
  type    = "A"
  ttl     = 200
  data    = ["10.1.0.1"]
}
`, name, fqdn)
}

func testAccDNSZoneBasicPublicPrivate(name, fqdn string) string {
	return fmt.Sprintf(`
resource "yandex_vpc_network" "net1" {}

resource "yandex_dns_zone" "zone1" {
  name        = "%[1]s"
  description = "desc"

  labels = {
    tf-label    = "tf-label-value"
    empty-label = ""
  }

  zone             = "%[2]s"
  public           = true
  private_networks = [yandex_vpc_network.net1.id]
}

resource "yandex_dns_recordset" "rs1" {
  zone_id = yandex_dns_zone.zone1.id
  name    = "srv.%[2]s"
  type    = "A"
  ttl     = 200
  data    = ["10.1.0.1"]
}
`, name, fqdn)
}

func testAccDNSZoneUpdate1(name, fqdn string) string {
	return fmt.Sprintf(`
resource "yandex_vpc_network" "net1" {}
resource "yandex_vpc_network" "net2" {}
resource "yandex_vpc_network" "net3" {}

resource "yandex_dns_zone" "zone1" {
  name        = "%s"
  description = "desc1"

  labels = {
    tf-label    = "tf-label-value1"
  }

  zone             = "%s"
  public           = true
  private_networks = [yandex_vpc_network.net1.id, yandex_vpc_network.net3.id]
}
`, name, fqdn)
}

func testAccCheckDnsZoneDestroy(s *terraform.State) error {
	sdk := getSDK(testAccProvider.Meta().(*Config))

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_dns_zone" {
			continue
		}

		_, err := sdk.DNS().DnsZone().Get(context.Background(), &dns.GetDnsZoneRequest{
			DnsZoneId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("Dns Zone still exists")
		}
	}

	return nil
}

func testAccCheckDnsZoneExists(name string, dnsZone *dns.DnsZone) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.DNS().DnsZone().Get(context.Background(), &dns.GetDnsZoneRequest{
			DnsZoneId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("DNS Zone not found")
		}

		*dnsZone = *found

		return nil
	}
}
