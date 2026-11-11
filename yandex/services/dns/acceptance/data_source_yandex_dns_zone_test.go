package acceptance

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
	"github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance/statecheck"
)

func TestAccDataSourceDNSZone_byID(t *testing.T) {
	suite := acceptance.New(t)
	t.Parallel()

	zoneName := acctest.RandomWithPrefix("tf-dns-zone")
	fqdn := acctest.RandomWithPrefix("tf-test") + ".dnstest.test."

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: func(state *terraform.State) error { return testAccCheckDnsZoneDestroy(suite, state) },
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceDnsZoneConfig(zoneName, fqdn),
				Check:  testAccDataSourceDnsZoneCheck("data.yandex_dns_zone.bar", "yandex_dns_zone.zone1"),
			},
		},
	})
}

const computeDnsZoneDataByIDConfig = `
data "yandex_dns_zone" "bar" {
  dns_zone_id = yandex_dns_zone.zone1.id
}
`

func testAccDataSourceDnsZoneConfig(zoneName, fqdn string) string {
	return testAccDNSZoneBasic(zoneName, fqdn) + computeDnsZoneDataByIDConfig
}

func testAccDataSourceDnsZoneCheck(datasourceName string, resourceName string) resource.TestCheckFunc {
	instanceAttrsToTest := []string{
		"zone", "folder_id", "name", "description", "labels", "public", "private_networks",
	}

	return statecheck.MatchingAttributes(datasourceName, resourceName, instanceAttrsToTest)
}
