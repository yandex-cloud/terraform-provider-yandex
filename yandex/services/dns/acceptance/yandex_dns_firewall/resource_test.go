package yandex_dns_firewall_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	dns "github.com/yandex-cloud/go-genproto/yandex/cloud/dns/v1"
	dnsv1sdk "github.com/yandex-cloud/go-sdk/services/dns/v1"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/internal/provider/framework"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestAccDNSFirewall_basic(t *testing.T) {
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
				Config: testAccDnsFirewallBasic(folderID, name, "test description", "label-value"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDnsFirewallExists("yandex_dns_firewall.test"),
					resource.TestCheckResourceAttr("yandex_dns_firewall.test", "name", name),
					resource.TestCheckResourceAttr("yandex_dns_firewall.test", "description", "test description"),
					resource.TestCheckResourceAttr("yandex_dns_firewall.test", "folder_id", folderID),
					resource.TestCheckResourceAttr("yandex_dns_firewall.test", "enabled", "true"),
					resource.TestCheckResourceAttr("yandex_dns_firewall.test", "labels.test-label", "label-value"),
					resource.TestCheckResourceAttrSet("yandex_dns_firewall.test", "id"),
					resource.TestCheckResourceAttrSet("yandex_dns_firewall.test", "created_at"),
				),
			},
			{
				ResourceName:      "yandex_dns_firewall.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDNSFirewall_update(t *testing.T) {
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
				Config: testAccDnsFirewallBasic(folderID, name, "initial description", "initial-value"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDnsFirewallExists("yandex_dns_firewall.test"),
					resource.TestCheckResourceAttr("yandex_dns_firewall.test", "description", "initial description"),
					resource.TestCheckResourceAttr("yandex_dns_firewall.test", "labels.test-label", "initial-value"),
					resource.TestCheckResourceAttr("yandex_dns_firewall.test", "enabled", "true"),
				),
			},
			{
				Config: testAccDnsFirewallUpdated(folderID, name, "updated description", "updated-value"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDnsFirewallExists("yandex_dns_firewall.test"),
					resource.TestCheckResourceAttr("yandex_dns_firewall.test", "description", "updated description"),
					resource.TestCheckResourceAttr("yandex_dns_firewall.test", "labels.test-label", "updated-value"),
					resource.TestCheckResourceAttr("yandex_dns_firewall.test", "enabled", "false"),
				),
			},
			{
				ResourceName:      "yandex_dns_firewall.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDNSFirewall_withFqdns(t *testing.T) {
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
				Config: testAccDnsFirewallWithFqdns(folderID, name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDnsFirewallExists("yandex_dns_firewall.test"),
					resource.TestCheckTypeSetElemAttr("yandex_dns_firewall.test", "whitelist_fqdns.*", "allowed.example.com."),
					resource.TestCheckTypeSetElemAttr("yandex_dns_firewall.test", "blacklist_fqdns.*", "blocked.example.com."),
				),
			},
			{
				ResourceName:      "yandex_dns_firewall.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDNSFirewall_withResourceConfig(t *testing.T) {
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
				Config: testAccDnsFirewallWithResourceConfig(folderID, name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDnsFirewallExists("yandex_dns_firewall.test"),
					resource.TestCheckResourceAttr("yandex_dns_firewall.test", "resource_config.type", "NETWORK"),
				),
			},
			{
				ResourceName:      "yandex_dns_firewall.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDNSFirewall_deletionProtection(t *testing.T) {
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
				Config: testAccDnsFirewallDeletionProtection(folderID, name, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDnsFirewallExists("yandex_dns_firewall.test"),
					resource.TestCheckResourceAttr("yandex_dns_firewall.test", "deletion_protection", "true"),
				),
			},
			{
				Config: testAccDnsFirewallDeletionProtection(folderID, name, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDnsFirewallExists("yandex_dns_firewall.test"),
					resource.TestCheckResourceAttr("yandex_dns_firewall.test", "deletion_protection", "false"),
				),
			},
			{
				ResourceName:      "yandex_dns_firewall.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDnsFirewallDestroy(s *terraform.State) error {
	config := testhelpers.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_dns_firewall" {
			continue
		}

		md := new(metadata.MD)
		_, err := dnsv1sdk.NewDnsFirewallClient(config.SDKv2).Get(
			context.Background(),
			&dns.GetDnsFirewallRequest{DnsFirewallId: rs.Primary.ID},
			grpc.Header(md),
		)

		if err == nil {
			return fmt.Errorf("DNS Firewall %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckDnsFirewallExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set for %s", resourceName)
		}

		config := testhelpers.AccProvider.(*yandex_framework.Provider).GetConfig()

		md := new(metadata.MD)
		found, err := dnsv1sdk.NewDnsFirewallClient(config.SDKv2).Get(
			context.Background(),
			&dns.GetDnsFirewallRequest{DnsFirewallId: rs.Primary.ID},
			grpc.Header(md),
		)
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("DNS Firewall %s not found", resourceName)
		}

		return nil
	}
}

func testAccDnsFirewallBasic(folderID, name, description, labelValue string) string {
	return fmt.Sprintf(`
resource "yandex_dns_firewall" "test" {
  folder_id   = "%s"
  name        = "%s"
  description = "%s"
  enabled     = true

  labels = {
    test-label = "%s"
  }
}
`, folderID, name, description, labelValue)
}

func testAccDnsFirewallUpdated(folderID, name, description, labelValue string) string {
	return fmt.Sprintf(`
resource "yandex_dns_firewall" "test" {
  folder_id   = "%s"
  name        = "%s"
  description = "%s"
  enabled     = false

  labels = {
    test-label = "%s"
  }
}
`, folderID, name, description, labelValue)
}

func testAccDnsFirewallWithFqdns(folderID, name string) string {
	return fmt.Sprintf(`
resource "yandex_dns_firewall" "test" {
  folder_id      = "%s"
  name           = "%s"
  enabled        = true
  whitelist_fqdns = ["allowed.example.com."]
  blacklist_fqdns = ["blocked.example.com."]
}
`, folderID, name)
}

func testAccDnsFirewallWithResourceConfig(folderID, name string) string {
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
`, folderID, name)
}

func testAccDnsFirewallDeletionProtection(folderID, name string, deletionProtection bool) string {
	return fmt.Sprintf(`
resource "yandex_dns_firewall" "test" {
  folder_id           = "%s"
  name                = "%s"
  enabled             = true
  deletion_protection = %t
}
`, folderID, name, deletionProtection)
}
