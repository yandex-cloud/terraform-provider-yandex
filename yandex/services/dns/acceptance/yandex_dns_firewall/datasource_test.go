package yandex_dns_firewall_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func TestAccDataSourceDNSFirewall_byID(t *testing.T) {
	var (
		folderID   = testhelpers.GetExampleFolderID()
		name       = fmt.Sprintf("tf-test-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
		labelValue = acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testhelpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: testhelpers.AccProviderFactories,
		CheckDestroy:             testAccCheckDnsFirewallDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceDnsFirewallByID(folderID, name, labelValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDnsFirewallExists("yandex_dns_firewall.test"),
					resource.TestCheckResourceAttrSet("data.yandex_dns_firewall.source", "id"),
					resource.TestCheckResourceAttrSet("data.yandex_dns_firewall.source", "dns_firewall_id"),
					resource.TestCheckResourceAttr("data.yandex_dns_firewall.source", "name", name),
					resource.TestCheckResourceAttr("data.yandex_dns_firewall.source", "folder_id", folderID),
					resource.TestCheckResourceAttr("data.yandex_dns_firewall.source", "enabled", "true"),
					resource.TestCheckResourceAttr("data.yandex_dns_firewall.source", "labels.test-label", labelValue),
					resource.TestCheckResourceAttrSet("data.yandex_dns_firewall.source", "created_at"),
				),
			},
		},
	})
}

func TestAccDataSourceDNSFirewall_withFqdns(t *testing.T) {
	var (
		folderID = testhelpers.GetExampleFolderID()
		name     = fmt.Sprintf("tf-test-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testhelpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: testhelpers.AccProviderFactories,
		CheckDestroy:             testAccCheckDnsFirewallDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceDnsFirewallWithFqdns(folderID, name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDnsFirewallExists("yandex_dns_firewall.test"),
					resource.TestCheckTypeSetElemAttr("data.yandex_dns_firewall.source", "whitelist_fqdns.*", "allowed.example.com."),
					resource.TestCheckTypeSetElemAttr("data.yandex_dns_firewall.source", "blacklist_fqdns.*", "blocked.example.com."),
				),
			},
		},
	})
}

func TestAccDataSourceDNSFirewall_withResourceConfig(t *testing.T) {
	var (
		folderID = testhelpers.GetExampleFolderID()
		name     = fmt.Sprintf("tf-test-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testhelpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: testhelpers.AccProviderFactories,
		CheckDestroy:             testAccCheckDnsFirewallDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceDnsFirewallWithResourceConfig(folderID, name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDnsFirewallExists("yandex_dns_firewall.test"),
					resource.TestCheckResourceAttrSet("data.yandex_dns_firewall.source", "id"),
					resource.TestCheckResourceAttr("data.yandex_dns_firewall.source", "resource_config.type", "NETWORK"),
				),
			},
		},
	})
}

func testAccDataSourceDnsFirewallByID(folderID, name, labelValue string) string {
	return fmt.Sprintf(`
resource "yandex_dns_firewall" "test" {
  folder_id   = "%s"
  name        = "%s"
  enabled     = true

  labels = {
    test-label = "%s"
  }
}

data "yandex_dns_firewall" "source" {
  dns_firewall_id = yandex_dns_firewall.test.id
}
`, folderID, name, labelValue)
}

func testAccDataSourceDnsFirewallWithResourceConfig(folderID, name string) string {
	return fmt.Sprintf(`
resource "yandex_dns_firewall" "test" {
  folder_id       = "%s"
  name            = "%s"
  enabled         = true
  whitelist_fqdns = ["*.allowed.example.com."]
  blacklist_fqdns = ["blocked.example.com."]

  resource_config = {
    type = "NETWORK"
  }
}

data "yandex_dns_firewall" "source" {
  dns_firewall_id = yandex_dns_firewall.test.id
}
`, folderID, name)
}

func testAccDataSourceDnsFirewallWithFqdns(folderID, name string) string {
	return fmt.Sprintf(`
resource "yandex_dns_firewall" "test" {
  folder_id       = "%s"
  name            = "%s"
  enabled         = true
  whitelist_fqdns = ["allowed.example.com."]
  blacklist_fqdns = ["blocked.example.com."]
}

data "yandex_dns_firewall" "source" {
  dns_firewall_id = yandex_dns_firewall.test.id
}
`, folderID, name)
}
