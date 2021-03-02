package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/dns/v1"
)

func TestIpv6AddressDiffSuppress(t *testing.T) {
	cases := map[string]struct {
		Old, New       string
		ShouldSuppress bool
	}{
		"compact form should suppress diff": {
			Old:            "2a03:b0c0:1:e0::29b:8001",
			New:            "2a03:b0c0:0001:00e0:0000:0000:029b:8001",
			ShouldSuppress: true,
		},
		"different address should not suppress diff": {
			Old:            "2a03:b0c0:1:e00::29b:8001",
			New:            "2a03:b0c0:0001:00e0:0000:0000:029b:8001",
			ShouldSuppress: false,
		},
		"compact form should suppress diff 2": {
			Old:            "fd12:3456:789a:1:0:0:0:1",
			New:            "fd12:3456:789a:1::1",
			ShouldSuppress: true,
		},
	}

	for tn, tc := range cases {
		shouldSuppress := ipv6DiffSuppressor("", tc.Old, tc.New, nil)
		if shouldSuppress != tc.ShouldSuppress {
			t.Errorf("%s: expected %t", tn, tc.ShouldSuppress)
		}
	}
}

func TestAccDNSRecordSet_basic(t *testing.T) {
	t.Parallel()

	var rs dns.RecordSet
	zoneName := acctest.RandomWithPrefix("tf-dns-zone")
	fqdn := acctest.RandomWithPrefix("tf-test") + ".example.com."

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVPCAddressDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDNSRecordSetBasic(zoneName, fqdn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDNSRecordSetExists("yandex_dns_recordset.rs1", &rs),
					resource.TestCheckResourceAttr("yandex_dns_recordset.rs1", "type", "A"),
					resource.TestCheckResourceAttr("yandex_dns_recordset.rs1", "name", "srv."+fqdn),
					resource.TestCheckResourceAttr("yandex_dns_recordset.rs1", "ttl", "200"),
					testAccCheckDnsRecordsetData(&rs, "192.168.0.1", true),
					testAccCheckDnsRecordsetData(&rs, "192.168.0.2", true),
				),
			},
			{
				ResourceName:      "yandex_dns_recordset.rs1",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDNSRecordSet_zoneChange(t *testing.T) {
	t.Parallel()

	var rs dns.RecordSet
	zoneName := acctest.RandomWithPrefix("tf-dns-zone")
	fqdn := acctest.RandomWithPrefix("tf-test") + ".example.com."

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVPCAddressDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDNSRecordSetBasic(zoneName, fqdn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDNSRecordSetExists("yandex_dns_recordset.rs1", &rs),
					resource.TestCheckResourceAttr("yandex_dns_recordset.rs1", "type", "A"),
					resource.TestCheckResourceAttr("yandex_dns_recordset.rs1", "name", "srv."+fqdn),
					resource.TestCheckResourceAttr("yandex_dns_recordset.rs1", "ttl", "200"),
					testAccCheckDnsRecordsetData(&rs, "192.168.0.1", true),
				),
			}, {
				Config: testAccDNSRecordSetBasic2(zoneName, fqdn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDNSRecordSetExists("yandex_dns_recordset.rs1", &rs),
					resource.TestCheckResourceAttr("yandex_dns_recordset.rs1", "type", "A"),
					resource.TestCheckResourceAttr("yandex_dns_recordset.rs1", "name", "srv2."+fqdn),
					resource.TestCheckResourceAttr("yandex_dns_recordset.rs1", "ttl", "200"),
					testAccCheckDnsRecordsetData(&rs, "192.168.0.1", true),
				),
			},
			{
				ResourceName:      "yandex_dns_recordset.rs1",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDNSRecordSet_update(t *testing.T) {
	t.Parallel()

	var rs dns.RecordSet
	zoneName := acctest.RandomWithPrefix("tf-dns-zone")
	fqdn := acctest.RandomWithPrefix("tf-test") + ".example.com."

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVPCAddressDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDNSRecordSetBasic(zoneName, fqdn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDNSRecordSetExists("yandex_dns_recordset.rs1", &rs),
					resource.TestCheckResourceAttr("yandex_dns_recordset.rs1", "type", "A"),
					resource.TestCheckResourceAttr("yandex_dns_recordset.rs1", "name", "srv."+fqdn),
					resource.TestCheckResourceAttr("yandex_dns_recordset.rs1", "ttl", "200"),
					testAccCheckDnsRecordsetData(&rs, "192.168.0.1", true),
					testAccCheckDnsRecordsetData(&rs, "192.168.0.2", true),
				),
			}, {
				Config: testAccDNSRecordSetBasic3(zoneName, fqdn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDNSRecordSetExists("yandex_dns_recordset.rs1", &rs),
					resource.TestCheckResourceAttr("yandex_dns_recordset.rs1", "type", "CNAME"),
					resource.TestCheckResourceAttr("yandex_dns_recordset.rs1", "name", "srv2."+fqdn),
					resource.TestCheckResourceAttr("yandex_dns_recordset.rs1", "ttl", "300"),
					testAccCheckDnsRecordsetData(&rs, "srv."+fqdn, true),
					testAccCheckDnsRecordsetData(&rs, "192.168.0.1", false),
					testAccCheckDnsRecordsetData(&rs, "192.168.0.2", false),
				),
			},
			{
				ResourceName: "yandex_dns_recordset.rs1",
				ImportState:  true,
			},
		},
	})
}

func testAccCheckDNSRecordSetExists(name string, rst *dns.RecordSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		dnsZoneId := rs.Primary.Attributes["zone_id"]
		dnsName := rs.Primary.Attributes["name"]
		dnsType := rs.Primary.Attributes["type"]

		sdk := getSDK(testAccProvider.Meta().(*Config))

		found, err := sdk.DNS().DnsZone().GetRecordSet(context.Background(), &dns.GetDnsZoneRecordSetRequest{
			DnsZoneId: dnsZoneId,
			Name:      dnsName,
			Type:      dnsType,
		})
		if err != nil {
			return err
		}

		//goland:noinspection GoVetCopyLock
		*rst = *found

		return nil
	}
}

func testAccCheckDnsRecordsetData(rs *dns.RecordSet, data string, isPresent bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var found bool
		for _, s := range rs.Data {
			if s == data {
				found = true
				break
			}
		}

		if found != isPresent {
			return fmt.Errorf("invalid presence for data %s in record set %s %s", data, rs.Type, rs.Name)
		}

		return nil
	}
}

func testAccDNSRecordSetBasic(name, fqdn string) string {
	return fmt.Sprintf(`
resource "yandex_dns_zone" "zone1" {
  name        = "%[1]s"
  description = "desc"
  zone        = "%[2]s"
}

resource "yandex_dns_recordset" "rs1" {
  zone_id = yandex_dns_zone.zone1.id
  name    = "srv.%[2]s"
  type    = "A"
  ttl     = 200
  data    = ["192.168.0.1", "192.168.0.2"]
}
`, name, fqdn)
}

func testAccDNSRecordSetBasic2(name, fqdn string) string {
	return fmt.Sprintf(`
resource "yandex_dns_zone" "zone1" {
  name        = "%[1]s"
  description = "desc"
  zone        = "%[2]s"
}

resource "yandex_dns_recordset" "rs1" {
  zone_id = yandex_dns_zone.zone1.id
  name    = "srv2.%[2]s"
  type    = "A"
  ttl     = 200
  data    = ["192.168.0.1", "192.168.0.2"]
}
`, name, fqdn)
}

func testAccDNSRecordSetBasic3(name, fqdn string) string {
	return fmt.Sprintf(`
resource "yandex_dns_zone" "zone1" {
  name        = "%[1]s"
  description = "desc"
  zone        = "%[2]s"
}

resource "yandex_dns_recordset" "rs1" {
  zone_id = yandex_dns_zone.zone1.id
  name    = "srv2.%[2]s"
  type    = "CNAME"
  ttl     = 300
  data    = ["srv.%[2]s"]
}
`, name, fqdn)
}
