package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
)

func testAccVPCAddressBasic(name string, zone string) string {
	return fmt.Sprintf(`
resource "yandex_dns_zone" "zone1" {
  description = "desc"
  zone        = "%[2]s"
  public      = true
}

resource "yandex_vpc_address" "addr1" {
  name        = "%[1]s"
  description = "desc"

  labels = {
    tf-label    = "tf-label-value"
    empty-label = ""
  }

  external_ipv4_address {
    zone_id                  = "ru-central1-d"
    ddos_protection_provider = "qrator"
  }
  deletion_protection = true

  dns_record {
     dns_zone_id = yandex_dns_zone.zone1.id
     fqdn     = "some.fqdn"
  }
}
`, name, zone)
}

func testAccVPCAddressUpdate(name string, zone string) string {
	return fmt.Sprintf(`
resource "yandex_dns_zone" "zone1" {
  description = "desc"
  zone        = "%[2]s"
  public      = true
}

resource "yandex_vpc_address" "addr1" {
  name        = "%[1]s"
  description = "new desc"

  labels = {
    new-label = "new"
  }

  external_ipv4_address {
    zone_id                  = "ru-central1-d"
  }

  dns_record {
     dns_zone_id = yandex_dns_zone.zone1.id
     fqdn     = "other.fqdn"
     ptr      = true
  }
}
`, name, zone)
}

func testAccVPCAddressRecreate(name string) string {
	return fmt.Sprintf(`
resource "yandex_vpc_address" "addr1" {
  name        = "%s"
  description = "new desc"

  labels = {
    new-label = "new"
  }

  external_ipv4_address {
    zone_id                  = "ru-central1-d"
    ddos_protection_provider = "qrator"
  }
  deletion_protection = false
}
`, name)
}

func TestAccVPCAddress_basic(t *testing.T) {
	t.Parallel()

	var address vpc.Address
	addressName := acctest.RandomWithPrefix("tf-address")
	dnsZone := acctest.RandomWithPrefix("zone") + ".dnstest.test."
	updatedAddressName := acctest.RandomWithPrefix("tf-address")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVPCAddressDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCAddressBasic(addressName, dnsZone),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCAddressExists("yandex_vpc_address.addr1", &address),
					resource.TestCheckResourceAttrSet("yandex_vpc_address.addr1", "folder_id"),
					resource.TestCheckResourceAttr("yandex_vpc_address.addr1", "name", addressName),
					resource.TestCheckResourceAttr("yandex_vpc_address.addr1", "description", "desc"),
					resource.TestCheckResourceAttr("yandex_vpc_address.addr1", "external_ipv4_address.#", "1"),
					resource.TestCheckResourceAttr(
						"yandex_vpc_address.addr1", "external_ipv4_address.0.zone_id", "ru-central1-d",
					),
					resource.TestCheckResourceAttr(
						"yandex_vpc_address.addr1", "external_ipv4_address.0.ddos_protection_provider", "qrator",
					),
					resource.TestCheckResourceAttr("yandex_vpc_address.addr1", "reserved", "true"),
					resource.TestCheckResourceAttr("yandex_vpc_address.addr1", "used", "false"),
					resource.TestCheckResourceAttr("yandex_vpc_address.addr1", "deletion_protection", "true"),
					testAccCheckVPCAddressContainsLabel(&address, "tf-label", "tf-label-value"),
					testAccCheckVPCAddressContainsLabel(&address, "empty-label", ""),
					testAccCheckCreatedAtAttr("yandex_vpc_address.addr1"),
					resource.TestCheckResourceAttr("yandex_vpc_address.addr1", "dns_record.#", "1"),
					resource.TestCheckResourceAttr("yandex_vpc_address.addr1", "dns_record.0.fqdn", "some.fqdn"),
				),
			},
			{
				Config: testAccVPCAddressUpdate(updatedAddressName, dnsZone),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCAddressExists("yandex_vpc_address.addr1", &address),
					resource.TestCheckResourceAttrSet("yandex_vpc_address.addr1", "folder_id"),
					resource.TestCheckResourceAttr("yandex_vpc_address.addr1", "name", updatedAddressName),
					resource.TestCheckResourceAttr("yandex_vpc_address.addr1", "description", "new desc"),
					resource.TestCheckResourceAttr("yandex_vpc_address.addr1", "external_ipv4_address.#", "1"),
					resource.TestCheckResourceAttr(
						"yandex_vpc_address.addr1", "external_ipv4_address.0.zone_id", "ru-central1-d",
					),
					resource.TestCheckResourceAttr("yandex_vpc_address.addr1", "reserved", "true"),
					resource.TestCheckResourceAttr("yandex_vpc_address.addr1", "used", "false"),
					resource.TestCheckResourceAttr("yandex_vpc_address.addr1", "deletion_protection", "true"),
					testAccCheckVPCAddressContainsLabelNotFound(&address, "tf-label"),
					testAccCheckVPCAddressContainsLabelNotFound(&address, "empty-label"),
					testAccCheckVPCAddressContainsLabel(&address, "new-label", "new"),
					testAccCheckCreatedAtAttr("yandex_vpc_address.addr1"),
					resource.TestCheckResourceAttr("yandex_vpc_address.addr1", "dns_record.#", "1"),
					resource.TestCheckResourceAttr("yandex_vpc_address.addr1", "dns_record.0.fqdn", "other.fqdn"),
					resource.TestCheckResourceAttr("yandex_vpc_address.addr1", "dns_record.0.ptr", "true"),
				),
			},
			{
				Config: testAccVPCAddressRecreate(updatedAddressName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCAddressRecreated("yandex_vpc_address.addr1", address.GetId()),
					resource.TestCheckResourceAttr(
						"yandex_vpc_address.addr1", "external_ipv4_address.0.zone_id", "ru-central1-d",
					),
					resource.TestCheckResourceAttr(
						"yandex_vpc_address.addr1", "external_ipv4_address.0.ddos_protection_provider", "qrator",
					),
					resource.TestCheckResourceAttr("yandex_vpc_address.addr1", "deletion_protection", "false"),
					resource.TestCheckResourceAttr("yandex_vpc_address.addr1", "dns_record.#", "0"),
				),
			},
			{
				ResourceName:      "yandex_vpc_address.addr1",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckVPCAddressRecreated(name string, id string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == id {
			return fmt.Errorf("VPC Address is not recreated")
		}

		return nil
	}
}

func testAccCheckVPCAddressDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_vpc_address" {
			continue
		}

		_, err := config.sdk.VPC().Address().Get(context.Background(), &vpc.GetAddressRequest{
			AddressId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("address still exists")
		}
	}

	return nil
}

func testAccCheckVPCAddressExists(name string, address *vpc.Address) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.VPC().Address().Get(context.Background(), &vpc.GetAddressRequest{
			AddressId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("address not found")
		}

		//goland:noinspection GoVetCopyLock
		*address = *found

		return nil
	}
}

func testAccCheckVPCAddressContainsLabel(address *vpc.Address, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		v, ok := address.Labels[key]
		if !ok {
			return fmt.Errorf("expected label with key '%s' not found", key)
		}
		if v != value {
			return fmt.Errorf("incorrect label value for key '%s': expected '%s' but found '%s'", key, value, v)
		}
		return nil
	}
}

func testAccCheckVPCAddressContainsLabelNotFound(address *vpc.Address, key string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if _, ok := address.Labels[key]; ok {
			return fmt.Errorf("expected label with key '%s' found", key)
		}
		return nil
	}
}
