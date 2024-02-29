package yandex

import (
	"fmt"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/dns/v1"
	"reflect"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
)

const dnsZoneResource = "yandex_dns_zone.test-key"

func importDNSZoneIDFunc(dnsZone *dns.DnsZone, role string) func(*terraform.State) (string, error) {
	return func(s *terraform.State) (string, error) {
		return dnsZone.Id + " " + role, nil
	}
}

func TestAccDNSZoneIamBinding_basic(t *testing.T) {
	var dnsZone dns.DnsZone
	symmetricKeyName := acctest.RandomWithPrefix("tf-dns-zone")

	role := "viewer"
	userID := "system:allUsers"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDNSZoneIamBindingBasic(symmetricKeyName, role, userID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDnsZoneExists(dnsZoneResource, &dnsZone),
					testAccCheckDNSZoneIam(dnsZoneResource, role, []string{userID}),
				),
			},
			{
				ResourceName:      "yandex_dns_zone_iam_binding.viewer",
				ImportStateIdFunc: importDNSZoneIDFunc(&dnsZone, role),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDNSZoneIamBinding_remove(t *testing.T) {
	var dnsZone dns.DnsZone
	symmetricKeyName := acctest.RandomWithPrefix("tf-dns-zone")

	role := "viewer"
	userID := "system:allUsers"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			// Prepare data source
			{
				Config: testAccDNSZone(symmetricKeyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDnsZoneExists(dnsZoneResource, &dnsZone),
					testAccCheckDNSZoneEmptyIam(dnsZoneResource),
				),
			},
			// Apply IAM bindings
			{
				Config: testAccDNSZoneIamBindingBasic(symmetricKeyName, role, userID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDNSZoneIam(dnsZoneResource, role, []string{userID}),
				),
			},
			// Remove the bindings
			{
				Config: testAccDNSZone(symmetricKeyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDNSZoneEmptyIam(dnsZoneResource),
				),
			},
		},
	})
}

func testAccDNSZoneIamBindingBasic(dnsZoneId, role, userID string) string {
	return fmt.Sprintf(`
resource "yandex_dns_zone" "test-key" {
  name = "%s"
  zone = "t.e.s.t.z.o.n.e."
}

resource "yandex_dns_zone_iam_binding" "viewer" {
  dns_zone_id = yandex_dns_zone.test-key.id
  role        = "%s"
  members     = ["%s"]
}
`, dnsZoneId, role, userID)
}

func testAccDNSZone(symmetricKeyName string) string {
	return fmt.Sprintf(`
resource "yandex_dns_zone" "test-key" {
  name = "%s"
  zone = "t.e.s.t.z.o.n.e."
}
`, symmetricKeyName)
}

func testAccCheckDNSZoneEmptyIam(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		bindings, err := getDNSZoneResourceAccessBindings(s, resourceName)
		if err != nil {
			return err
		}

		if len(bindings) == 0 {
			return nil
		}

		return fmt.Errorf("Binding found but expected empty for %s", resourceName)
	}
}

func testAccCheckDNSZoneIam(resourceName, role string, members []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		bindings, err := getDNSZoneResourceAccessBindings(s, resourceName)
		if err != nil {
			return err
		}

		var roleMembers []string
		for _, binding := range bindings {
			if binding.RoleId == role {
				member := binding.Subject.Type + ":" + binding.Subject.Id
				roleMembers = append(roleMembers, member)
			}
		}
		sort.Strings(members)
		sort.Strings(roleMembers)

		if reflect.DeepEqual(members, roleMembers) {
			return nil
		}

		return fmt.Errorf("Binding found but expected members is %v, got %v", members, roleMembers)
	}
}

func getDNSZoneResourceAccessBindings(s *terraform.State, resourceName string) ([]*access.AccessBinding, error) {
	config := testAccProvider.Meta().(*Config)

	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("can't find %s in state", resourceName)
	}

	return getDnsZoneAccessBindings(config.Context(), config, rs.Primary.ID)
}
