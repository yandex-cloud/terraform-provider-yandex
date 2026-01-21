package yandex_organizationmanager_idp_userpool_domain_test

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	idp "github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1/idp"
	idpsdk "github.com/yandex-cloud/go-sdk/services/organizationmanager/v1/idp"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
)

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}
func TestAccOrganizationManagerIdpUserpoolDomain_basic(t *testing.T) {
	userpoolName := acctest.RandomWithPrefix("tf-acc-test-userpool")
	organizationID := test.GetExampleOrganizationID()
	testSubdomain := acctest.RandomWithPrefix("tf-acc-test-subdomain")
	testDomain := acctest.RandomWithPrefix("tf-acc-test") + ".example.com"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckIdpUserpoolDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIdpUserpoolDomain_basic(userpoolName, organizationID, testSubdomain, testDomain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIdpUserpoolDomainExists("yandex_organizationmanager_idp_userpool_domain.foobar"),
					test.AccCheckCreatedAtAttr("yandex_organizationmanager_idp_userpool_domain.foobar"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_userpool_domain.foobar", "domain", testDomain),
					resource.TestCheckResourceAttrSet("yandex_organizationmanager_idp_userpool_domain.foobar", "userpool_id"),
				),
			},
			{
				ResourceName:      "yandex_organizationmanager_idp_userpool_domain.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
func testAccCheckIdpUserpoolDomainDestroy(s *terraform.State) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_organizationmanager_idp_userpool_domain" {
			continue
		}
		userpoolID := rs.Primary.Attributes["userpool_id"]
		domain := rs.Primary.Attributes["domain"]
		if userpoolID == "" || domain == "" {
			continue
		}
		_, err := idpsdk.NewUserpoolClient(config.SDKv2).GetDomain(context.Background(), &idp.GetUserpoolDomainRequest{
			UserpoolId: userpoolID,
			Domain:     domain,
		})
		if err != nil {
			if grpcStatus, ok := status.FromError(err); ok && grpcStatus != nil && grpcStatus.Code() == codes.NotFound {
				return nil
			} else if ok {
				return fmt.Errorf("Error while requesting Yandex Cloud: grpc code error : %d, http message error: %s", grpcStatus.Code(), grpcStatus.Message())
			}
			return fmt.Errorf("IdpUserpoolDomain still exists")
		}
	}
	return nil
}
func testAccCheckIdpUserpoolDomainExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}
		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()
		userpoolID := rs.Primary.Attributes["userpool_id"]
		domain := rs.Primary.Attributes["domain"]
		if userpoolID == "" || domain == "" {
			return fmt.Errorf("userpool_id or domain is not set")
		}
		found, err := idpsdk.NewUserpoolClient(config.SDKv2).GetDomain(context.Background(), &idp.GetUserpoolDomainRequest{
			UserpoolId: userpoolID,
			Domain:     domain,
		})
		if err != nil {
			return err
		}
		if found.GetDomain() != domain {
			return fmt.Errorf("IdpUserpoolDomain %s not found", n)
		}
		return nil
	}
}
func testAccIdpUserpoolDomain_basic(userpoolName, organizationID, defaultSubdomain, domain string) string {
	return fmt.Sprintf(`
resource "yandex_organizationmanager_idp_userpool" "foobar" {
  name              = "%s"
  organization_id   = "%s"
  default_subdomain = "%s"
}
resource "yandex_organizationmanager_idp_userpool_domain" "foobar" {
  userpool_id = yandex_organizationmanager_idp_userpool.foobar.userpool_id
  domain      = "%s"
}
`, userpoolName, organizationID, defaultSubdomain, domain)
}
