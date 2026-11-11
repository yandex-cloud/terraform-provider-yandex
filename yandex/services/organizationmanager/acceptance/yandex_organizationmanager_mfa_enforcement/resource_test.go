package yandex_organizationmanager_mfa_enforcement_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	organizationmanager "github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1"
	organizationmanagersdk "github.com/yandex-cloud/go-sdk/services/organizationmanager/v1"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/internal/provider/framework"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAccOrganizationManagerMfaEnforcementCreate(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-test-mfa-enforcement")
	organizationId := test.GetExampleOrganizationID()
	acrId := "any-mfa"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMfaEnforcementDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMfaEnforcementCreate(name, organizationId, acrId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMfaEnforcementExists("yandex_organizationmanager_mfa_enforcement.foobar"),
					test.AccCheckCreatedAtAttr("yandex_organizationmanager_mfa_enforcement.foobar"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_mfa_enforcement.foobar", "name", name),
					resource.TestCheckResourceAttr("yandex_organizationmanager_mfa_enforcement.foobar", "ttl", "5m0s"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_mfa_enforcement.foobar", "status", "MFA_ENFORCEMENT_STATUS_ACTIVE"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_mfa_enforcement.foobar", "enroll_window", "5h0m0s"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_mfa_enforcement.foobar", "organization_id", organizationId),
					resource.TestCheckResourceAttr("yandex_organizationmanager_mfa_enforcement.foobar", "acr_id", acrId),
				),
			},
		},
	})
}

func TestAccOrganizationManagerMfaEnforcementUpdate(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-test-mfa-enforcement")
	organizationId := test.GetExampleOrganizationID()
	acrId := "any-mfa"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMfaEnforcementDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMfaEnforcementCreate(name, organizationId, acrId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMfaEnforcementExists("yandex_organizationmanager_mfa_enforcement.foobar"),
				),
			},
			{
				Config: testAccMfaEnforcementUpdate("new-name", organizationId, "phr", "new description", "MFA_ENFORCEMENT_STATUS_INACTIVE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMfaEnforcementExists("yandex_organizationmanager_mfa_enforcement.foobar"),
					test.AccCheckCreatedAtAttr("yandex_organizationmanager_mfa_enforcement.foobar"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_mfa_enforcement.foobar", "name", "new-name"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_mfa_enforcement.foobar", "description", "new description"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_mfa_enforcement.foobar", "acr_id", "phr"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_mfa_enforcement.foobar", "status", "MFA_ENFORCEMENT_STATUS_INACTIVE"),
				),
			},
			{
				ResourceName:            "yandex_organizationmanager_mfa_enforcement.foobar",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func testAccCheckMfaEnforcementDestroy(s *terraform.State) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_organizationmanager_mfa_enforcement" {
			continue
		}

		_, err := organizationmanagersdk.NewMfaEnforcementClient(config.SDKv2).Get(context.Background(), &organizationmanager.GetMfaEnforcementRequest{
			MfaEnforcementId: rs.Primary.ID,
		})

		if err != nil {
			if grpcStatus, ok := status.FromError(err); ok && grpcStatus != nil && grpcStatus.Code() == codes.NotFound {
				return nil
			} else if ok {
				return fmt.Errorf("Error while requesting Yandex Cloud: grpc code error : %d, http message error: %s", grpcStatus.Code(), grpcStatus.Message())
			}
			return fmt.Errorf("MfaEnforcement still exists")
		}
	}

	return nil
}

func testAccCheckMfaEnforcementExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No id is set")
		}

		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

		found, err := organizationmanagersdk.NewMfaEnforcementClient(config.SDKv2).Get(context.Background(), &organizationmanager.GetMfaEnforcementRequest{
			MfaEnforcementId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("MfaEnforcement %s not found", n)
		}

		return nil
	}
}

func testAccMfaEnforcementCreate(name, organizationId, acrId string) string {
	return fmt.Sprintf(`
resource "yandex_organizationmanager_mfa_enforcement" "foobar" {
  name            = "%s"
  organization_id = "%s"
  acr_id 		  = "%s"
  ttl 		  	  = "5m0s"
  status 		  = "MFA_ENFORCEMENT_STATUS_ACTIVE"
  enroll_window   = "5h0m0s"
}
`, name, organizationId, acrId)
}

func testAccMfaEnforcementUpdate(name, organizationId, acrId, description, status string) string {
	return fmt.Sprintf(`
resource "yandex_organizationmanager_mfa_enforcement" "foobar" {
  name            = "%s"
  organization_id = "%s"
  acr_id 	 	  = "%s"
  ttl 			  = "5m0s"
  status 		  = "%s"
  enroll_window   = "5h0m0s"
  description     = "%s"
}
`, name, organizationId, acrId, status, description)
}
