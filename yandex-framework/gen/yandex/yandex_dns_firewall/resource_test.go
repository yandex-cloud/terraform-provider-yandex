package yandex_dns_firewall_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	dns "github.com/yandex-cloud/go-genproto/yandex/cloud/dns/v1"
	dnsv1sdk "github.com/yandex-cloud/go-sdk/services/dns/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const yandexDnsFirewallDefaultTimeout = 5 * time.Minute

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func init() {
	resource.AddTestSweepers("yandex_dns_firewall", &resource.Sweeper{
		Name:         "yandex_dns_firewall",
		F:            testSweepDnsFirewall,
		Dependencies: []string{},
	})
}

func testSweepDnsFirewall(_ string) error {
	conf, err := testhelpers.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := dnsv1sdk.NewDnsFirewallClient(conf.SDKv2).List(context.Background(), &dns.ListDnsFirewallsRequest{
		FolderId: conf.ProviderState.FolderID.ValueString(),
	})
	if err != nil {
		return fmt.Errorf("error listing DNS firewalls: %s", err)
	}

	result := &multierror.Error{}
	for _, fw := range resp.DnsFirewalls {
		if !sweepDnsFirewall(conf, fw.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep DNS Firewall %q", fw.Id))
		}
	}

	return result.ErrorOrNil()
}

func sweepDnsFirewall(conf *provider_config.Config, id string) bool {
	return testhelpers.SweepWithRetry(sweepDnsFirewallOnce, conf, "yandex_dns_firewall", id)
}

func sweepDnsFirewallOnce(conf *provider_config.Config, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), yandexDnsFirewallDefaultTimeout)
	defer cancel()

	op, err := dnsv1sdk.NewDnsFirewallClient(conf.SDKv2).Delete(ctx, &dns.DeleteDnsFirewallRequest{
		DnsFirewallId: id,
	})
	if err != nil {
		return err
	}
	_, err = op.Wait(ctx)
	return err
}

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
