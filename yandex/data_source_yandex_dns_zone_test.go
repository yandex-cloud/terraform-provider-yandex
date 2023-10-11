package yandex

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceDNSZone_byID(t *testing.T) {
	t.Parallel()

	zoneName := acctest.RandomWithPrefix("tf-dns-zone")
	fqdn := acctest.RandomWithPrefix("tf-test") + ".dnstest.test."

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDnsZoneDestroy,
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

	instanceAttrsToTest = append(instanceAttrsToTest, baseAttrsToTest...)
	return testAttrsCheck(datasourceName, resourceName, instanceAttrsToTest)
}
