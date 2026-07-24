package yandex

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/dns/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAccDNSRecordSet_basic(t *testing.T) {
	t.Parallel()

	var rs dns.RecordSet
	zoneName := acctest.RandomWithPrefix("tf-dns-zone")
	fqdn := acctest.RandomWithPrefix("tf-test") + ".dnstest.test."

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDNSRecordSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDNSRecordSetBasic(zoneName, fqdn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDNSRecordSetExists("yandex_dns_recordset.rs1", &rs),
					resource.TestCheckResourceAttr("yandex_dns_recordset.rs1", "type", "A"),
					resource.TestCheckResourceAttr("yandex_dns_recordset.rs1", "name", "srv."+fqdn),
					resource.TestCheckResourceAttr("yandex_dns_recordset.rs1", "ttl", "200"),
					resource.TestCheckResourceAttr("yandex_dns_recordset.rs1", "description", "rs1 description 1"),
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

func TestAccDNSRecordSet_short(t *testing.T) {
	t.Parallel()

	var rs dns.RecordSet
	zoneName := acctest.RandomWithPrefix("tf-dns-zone")
	fqdn := acctest.RandomWithPrefix("tf-test") + ".dnstest.test."

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDNSRecordSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDNSRecordSetBasicShort(zoneName, fqdn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDNSRecordSetExists("yandex_dns_recordset.rs1", &rs),
					resource.TestCheckResourceAttr("yandex_dns_recordset.rs1", "type", "A"),
					resource.TestCheckResourceAttr("yandex_dns_recordset.rs1", "name", "srv"),
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

func TestAccDNSRecordSet_zoneChange(t *testing.T) {
	t.Parallel()

	var rs dns.RecordSet
	zoneName := acctest.RandomWithPrefix("tf-dns-zone")
	fqdn := acctest.RandomWithPrefix("tf-test") + ".dnstest.test."

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDNSRecordSetDestroy,
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
	fqdn := acctest.RandomWithPrefix("tf-test") + ".dnstest.test."

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDNSRecordSetDestroy,
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
				Config: testAccDNSRecordSetBasic4(zoneName, fqdn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDNSRecordSetExists("yandex_dns_recordset.rs1", &rs),
					resource.TestCheckResourceAttr("yandex_dns_recordset.rs1", "type", "A"),
					resource.TestCheckResourceAttr("yandex_dns_recordset.rs1", "name", "srv."+fqdn),
					resource.TestCheckResourceAttr("yandex_dns_recordset.rs1", "description", "rs1 description 2"),
					resource.TestCheckResourceAttr("yandex_dns_recordset.rs1", "ttl", "200"),
					testAccCheckDnsRecordsetData(&rs, "192.168.0.1", false),
					testAccCheckDnsRecordsetData(&rs, "192.168.0.2", false),
					testAccCheckDnsRecordsetData(&rs, "192.168.0.3", true),
				),
			},
			{
				Config: testAccDNSRecordSetBasic3(zoneName, fqdn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDNSRecordSetExists("yandex_dns_recordset.rs1", &rs),
					resource.TestCheckResourceAttr("yandex_dns_recordset.rs1", "type", "CNAME"),
					resource.TestCheckResourceAttr("yandex_dns_recordset.rs1", "name", "srv2."+fqdn),
					resource.TestCheckResourceAttr("yandex_dns_recordset.rs1", "description", ""),
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
  zone_id     = yandex_dns_zone.zone1.id
  name        = "srv.%[2]s"
  type        = "A"
  description = "rs1 description 1"
  ttl         = 200
  data        = ["192.168.0.1", "192.168.0.2"]
}
`, name, fqdn)
}

func testAccDNSRecordSetBasicShort(name, fqdn string) string {
	return fmt.Sprintf(`
resource "yandex_dns_zone" "zone1" {
  name        = "%[1]s"
  description = "desc"
  zone        = "%[2]s"
}

resource "yandex_dns_recordset" "rs1" {
  zone_id = yandex_dns_zone.zone1.id
  name    = "srv"
  type    = "A"
  ttl     = 200
  data    = ["192.168.0.1"]
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

func testAccDNSRecordSetBasic4(name, fqdn string) string {
	return fmt.Sprintf(`
resource "yandex_dns_zone" "zone1" {
  name        = "%[1]s"
  description = "desc"
  zone        = "%[2]s"
}

resource "yandex_dns_recordset" "rs1" {
  zone_id     = yandex_dns_zone.zone1.id
  name        = "srv.%[2]s"
  type        = "A"
  description = "rs1 description 2"
  ttl         = 200
  data        = ["192.168.0.3"]
}

resource "yandex_dns_recordset" "rs2" {
  zone_id     = yandex_dns_zone.zone1.id
  name        = "srv3"
  type        = "A"
  description = "rs2 description 1"
  ttl         = 200
  data        = ["192.168.0.10"]
}
`, name, fqdn)
}

func testAccCheckDNSRecordSetDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)
	sdk := getSDK(config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_dns_recordset" {
			continue
		}

		_, err := sdk.DNS().DnsZone().GetRecordSet(context.Background(), &dns.GetDnsZoneRecordSetRequest{
			DnsZoneId: rs.Primary.Attributes["zone_id"],
			Name:      rs.Primary.Attributes["name"],
			Type:      rs.Primary.Attributes["type"],
		})

		if err == nil {
			return fmt.Errorf("DNS RecordSet still exists: %s/%s/%s",
				rs.Primary.Attributes["zone_id"],
				rs.Primary.Attributes["name"],
				rs.Primary.Attributes["type"])
		}
		if status.Code(err) != codes.NotFound {
			return fmt.Errorf("failed to check DNS RecordSet destruction %s/%s/%s: %w",
				rs.Primary.Attributes["zone_id"],
				rs.Primary.Attributes["name"],
				rs.Primary.Attributes["type"],
				err)
		}
	}

	return nil
}

// longDKIMTXTValue is a DKIM TXT value > 255 bytes; the Yandex DNS API will return it
// split into two quoted character-strings separated by " ".
const longDKIMTXTValue = `"v=DKIM1; k=rsa; p=MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA2a2rwplBQLF29amygykEMmYz0+Kcj3bKBp29OoVXFBFAQFbnScEBInGGaVFMBMPEBiMrJNmvRQKMn/YFTfDH9MWbEPHFY2XnHUHZQjKuHUkXqMpjIBXNBLxkNVRKEXBRGGBOEIhLBaKfM5N1LN7O9S8GkMbH3YUgBOK0NOZC0r1MFLgLXuJDVXg+GhFJfNIwzYHqHFLhRCrEmGIzAGxFmiqkjp8VQiMerGRilHolTuVFNkfVs/t1tTRTgGzXAx7ClZHOuKX/k9U/KEIlN1VEWQP4lcLqMqZ0i5GxrRFyOPqBBFbJKsRr40D5EQFKUEJ"`

// longDKIMTXTValue2 differs from longDKIMTXTValue by one character (QLF→QlF) to exercise update/delete semantics.
const longDKIMTXTValue2 = `"v=DKIM1; k=rsa; p=MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA2a2rwplBQlF29amygykEMmYz0+Kcj3bKBp29OoVXFBFAQFbnScEBInGGaVFMBMPEBiMrJNmvRQKMn/YFTfDH9MWbEPHFY2XnHUHZQjKuHUkXqMpjIBXNBLxkNVRKEXBRGGBOEIhLBaKfM5N1LN7O9S8GkMbH3YUgBOK0NOZC0r1MFLgLXuJDVXg+GhFJfNIwzYHqHFLhRCrEmGIzAGxFmiqkjp8VQiMerGRilHolTuVFNkfVs/t1tTRTgGzXAx7ClZHOuKX/k9U/KEIlN1VEWQP4lcLqMqZ0i5GxrRFyOPqBBFbJKsRr40D5EQFKUEJ"`

func testAccDNSRecordSetLongTXT(dnsZoneName, dnsZoneDescription, recordName, txtValue string) string {
	return fmt.Sprintf(`
resource "yandex_dns_zone" "zone1" {
  name        = %[1]q
  description = %[2]q
  zone        = "%[1]s.dnstest.test."
}

resource "yandex_dns_recordset" "rs1" {
  zone_id = yandex_dns_zone.zone1.id
  name    = %[3]q
  type    = "TXT"
  ttl     = 300
  data    = [%[4]q]
}
`, dnsZoneName, dnsZoneDescription, recordName, txtValue)
}

func TestAccDNSRecordSet_longTXT(t *testing.T) {
	t.Parallel()

	var rs1, rs2 dns.RecordSet

	dnsZoneName := fmt.Sprintf("zone-txt-%s", acctest.RandString(10))
	dnsZoneDescription := "Test zone for long TXT record"
	recordName := fmt.Sprintf("_domainkey.test%s.%s.dnstest.test.", acctest.RandString(6), dnsZoneName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDNSRecordSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDNSRecordSetLongTXT(dnsZoneName, dnsZoneDescription, recordName, longDKIMTXTValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDNSRecordSetExists("yandex_dns_recordset.rs1", &rs1),
					// Assert raw API returned a split value (contains `" "` boundary).
					func(s *terraform.State) error {
						if len(rs1.Data) == 0 {
							return fmt.Errorf("expected API to return split TXT data, got no values")
						}
						if !strings.Contains(rs1.Data[0], "\" \"") {
							return fmt.Errorf("expected API to return split TXT data (containing '\" \"'), got: %q", rs1.Data[0])
						}
						return nil
					},
				),
			},
			{
				Config:             testAccDNSRecordSetLongTXT(dnsZoneName, dnsZoneDescription, recordName, longDKIMTXTValue),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				Config: testAccDNSRecordSetLongTXT(dnsZoneName, dnsZoneDescription, recordName, longDKIMTXTValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDNSRecordSetExists("yandex_dns_recordset.rs1", &rs2),
				),
			},
			{
				Config:             testAccDNSRecordSetLongTXT(dnsZoneName, dnsZoneDescription, recordName, longDKIMTXTValue2),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				ResourceName:            "yandex_dns_recordset.rs1",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"data"},
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					if len(states) != 1 {
						return fmt.Errorf("expected one imported DNS RecordSet state, got %d", len(states))
					}

					for key, value := range states[0].Attributes {
						if strings.HasPrefix(key, "data.") && key != "data.#" && canonicalizeTXTRecordValue(value) == longDKIMTXTValue2 {
							return nil
						}
					}

					return fmt.Errorf("expected imported TXT data equivalent to %q, got %#v", longDKIMTXTValue2, states[0].Attributes)
				},
			},
		},
	})
}
