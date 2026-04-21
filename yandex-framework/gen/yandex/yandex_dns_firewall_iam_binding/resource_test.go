package yandex_dns_firewall_iam_binding_test

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
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers/iam"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	defaultTimeout = 5 * time.Minute

	dnsFirewallIAMBindingResourceType = "yandex_dns_firewall_iam_binding"
	dnsFirewallIAMBindingResourceFoo  = dnsFirewallIAMBindingResourceType + ".foo"
	dnsFirewallIAMBindingResourceBar  = dnsFirewallIAMBindingResourceType + ".bar"

	dnsFirewallIAMRoleUser   = "dns.firewallUser"
	dnsFirewallIAMRoleEditor = "dns.firewallEditor"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func init() {
	resource.AddTestSweepers("yandex_dns_firewall_iam_binding", &resource.Sweeper{
		Name:         "yandex_dns_firewall_iam_binding",
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
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
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

func TestAccDNSFirewallIamBinding_basic(t *testing.T) {
	var (
		firewall dns.DnsFirewall
		folderID = testhelpers.GetExampleFolderID()
		name     = fmt.Sprintf("tf-test-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
		ctx      = context.Background()
		role     = dnsFirewallIAMRoleUser
		userID   = "system:allAuthenticatedUsers"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testhelpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: testhelpers.AccProviderFactories,
		CheckDestroy:             testAccCheckDnsFirewallDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDnsFirewallIamBindingConfig(folderID, name, role, userID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDnsFirewallExistsWithObj("yandex_dns_firewall.test", &firewall),
					iam.TestAccCheckIamBindingContainsMembers(ctx, func() iam.BindingsGetter {
						cfg := testhelpers.AccProvider.(*yandex_framework.Provider).GetConfig()
						return dnsv1sdk.NewDnsFirewallClient(cfg.SDKv2)
					}, &firewall, role, []string{userID}),
				),
			},
			iam.IAMBindingImportTestStep(dnsFirewallIAMBindingResourceFoo, &firewall, role, "dns_firewall_id"),
		},
	})
}

func TestAccDNSFirewallIamBinding_multiple(t *testing.T) {
	var (
		firewall dns.DnsFirewall
		folderID = testhelpers.GetExampleFolderID()
		name     = fmt.Sprintf("tf-test-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
		ctx      = context.Background()
		roleFoo  = dnsFirewallIAMRoleUser
		roleBar  = dnsFirewallIAMRoleEditor
		userID   = "system:allAuthenticatedUsers"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testhelpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: testhelpers.AccProviderFactories,
		CheckDestroy:             testAccCheckDnsFirewallDestroy,
		Steps: []resource.TestStep{
			// Prepare firewall without bindings
			{
				Config: testAccDnsFirewallBasic(folderID, name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDnsFirewallExistsWithObj("yandex_dns_firewall.test", &firewall),
					iam.TestAccCheckIamBindingEmpty(ctx, func() iam.BindingsGetter {
						cfg := testhelpers.AccProvider.(*yandex_framework.Provider).GetConfig()
						return dnsv1sdk.NewDnsFirewallClient(cfg.SDKv2)
					}, &firewall, roleFoo),
				),
			},
			// One binding
			{
				Config: testAccDnsFirewallIamBindingConfig(folderID, name, roleFoo, userID),
				Check: iam.TestAccCheckIamBindingContainsMembers(ctx, func() iam.BindingsGetter {
					cfg := testhelpers.AccProvider.(*yandex_framework.Provider).GetConfig()
					return dnsv1sdk.NewDnsFirewallClient(cfg.SDKv2)
				}, &firewall, roleFoo, []string{userID}),
			},
			iam.IAMBindingImportTestStep(dnsFirewallIAMBindingResourceFoo, &firewall, roleFoo, "dns_firewall_id"),
			// Two bindings
			{
				Config: testAccDnsFirewallIamBindingMultipleConfig(folderID, name, roleFoo, roleBar, userID),
				Check: resource.ComposeTestCheckFunc(
					iam.TestAccCheckIamBindingContainsMembers(ctx, func() iam.BindingsGetter {
						cfg := testhelpers.AccProvider.(*yandex_framework.Provider).GetConfig()
						return dnsv1sdk.NewDnsFirewallClient(cfg.SDKv2)
					}, &firewall, roleFoo, []string{userID}),
					iam.TestAccCheckIamBindingContainsMembers(ctx, func() iam.BindingsGetter {
						cfg := testhelpers.AccProvider.(*yandex_framework.Provider).GetConfig()
						return dnsv1sdk.NewDnsFirewallClient(cfg.SDKv2)
					}, &firewall, roleBar, []string{userID}),
				),
			},
			iam.IAMBindingImportTestStep(dnsFirewallIAMBindingResourceFoo, &firewall, roleFoo, "dns_firewall_id"),
			iam.IAMBindingImportTestStep(dnsFirewallIAMBindingResourceBar, &firewall, roleBar, "dns_firewall_id"),
			// Remove all bindings
			{
				Config: testAccDnsFirewallBasic(folderID, name),
				Check: resource.ComposeTestCheckFunc(
					iam.TestAccCheckIamBindingEmpty(ctx, func() iam.BindingsGetter {
						cfg := testhelpers.AccProvider.(*yandex_framework.Provider).GetConfig()
						return dnsv1sdk.NewDnsFirewallClient(cfg.SDKv2)
					}, &firewall, roleFoo),
					iam.TestAccCheckIamBindingEmpty(ctx, func() iam.BindingsGetter {
						cfg := testhelpers.AccProvider.(*yandex_framework.Provider).GetConfig()
						return dnsv1sdk.NewDnsFirewallClient(cfg.SDKv2)
					}, &firewall, roleBar),
				),
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

func testAccCheckDnsFirewallExistsWithObj(resourceName string, firewall *dns.DnsFirewall) resource.TestCheckFunc {
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

		firewall.Id = found.Id
		return nil
	}
}

func testAccDnsFirewallBasic(folderID, name string) string {
	return fmt.Sprintf(`
resource "yandex_dns_firewall" "test" {
  folder_id = "%s"
  name      = "%s"
  enabled   = true
}
`, folderID, name)
}

func testAccDnsFirewallIamBindingConfig(folderID, name, role, userID string) string {
	return fmt.Sprintf(`
resource "yandex_dns_firewall" "test" {
  folder_id = "%s"
  name      = "%s"
  enabled   = true
}

resource "yandex_dns_firewall_iam_binding" "foo" {
  dns_firewall_id = yandex_dns_firewall.test.id
  role            = "%s"
  members         = ["%s"]
}
`, folderID, name, role, userID)
}

func testAccDnsFirewallIamBindingMultipleConfig(folderID, name, roleFoo, roleBar, userID string) string {
	return fmt.Sprintf(`
resource "yandex_dns_firewall" "test" {
  folder_id = "%s"
  name      = "%s"
  enabled   = true
}

resource "yandex_dns_firewall_iam_binding" "foo" {
  dns_firewall_id = yandex_dns_firewall.test.id
  role            = "%s"
  members         = ["%s"]
}

resource "yandex_dns_firewall_iam_binding" "bar" {
  dns_firewall_id = yandex_dns_firewall.test.id
  role            = "%s"
  members         = ["%s"]
}
`, folderID, name, roleFoo, userID, roleBar, userID)
}
