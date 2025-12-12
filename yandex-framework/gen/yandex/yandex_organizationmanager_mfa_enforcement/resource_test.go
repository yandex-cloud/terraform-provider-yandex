package yandex_organizationmanager_mfa_enforcement_test

import (
	"context"
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	organizationmanager "github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1"
	organizationmanagersdk "github.com/yandex-cloud/go-sdk/services/organizationmanager/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	sweepPageSize          = 1000
	sweepDeleteTimeout     = 15 * time.Minute
	testResourceNamePrefix = "tf-acc-test-mfa-enforcement"
)

func init() {
	resource.AddTestSweepers("yandex_organizationmanager_mfa_enforcement", &resource.Sweeper{
		Name:         "yandex_organizationmanager_mfa_enforcement",
		F:            testSweepMfaEnforcement,
		Dependencies: []string{},
	})
}

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

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
					resource.TestCheckResourceAttr("yandex_organizationmanager_mfa_enforcement.foobar", "ttl", "5s"),
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
  ttl 		  	  = "5s"
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
  ttl 			  = "5s"
  status 		  = "%s"
  enroll_window   = "5h0m0s"
  description     = "%s"
}
`, name, organizationId, acrId, status, description)
}

func testSweepMfaEnforcement(_ string) error {
	conf, err := test.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	organizationId := test.GetExampleOrganizationID()
	if organizationId == "" {
		log.Printf("[WARN] organization id is not set, skipping mfa enforcement sweep")
		return nil
	}

	req := &organizationmanager.ListMfaEnforcementsRequest{
		OrganizationId: organizationId,
		PageSize:       sweepPageSize,
	}

	client := organizationmanagersdk.NewMfaEnforcementClient(conf.SDKv2)
	resp, err := client.List(context.Background(), req)
	if err != nil {
		return fmt.Errorf("error getting list of mfa enforcements: %s", err)
	}

	result := &multierror.Error{}
	for _, mfaEnforcement := range resp.MfaEnforcements {
		if strings.HasPrefix(mfaEnforcement.Name, testResourceNamePrefix) {
			if !sweepMfaEnforcement(conf, mfaEnforcement.Id) {
				result = multierror.Append(result, fmt.Errorf("failed to sweep mfa enforcement %q", mfaEnforcement.Id))
			}
		}
	}

	// Handle pagination if needed
	for resp.NextPageToken != "" {
		req.PageToken = resp.NextPageToken
		resp, err = client.List(context.Background(), req)
		if err != nil {
			result = multierror.Append(result, fmt.Errorf("error getting next page of mfa enforcements: %s", err))
			break
		}

		for _, mfaEnforcement := range resp.MfaEnforcements {
			if strings.HasPrefix(mfaEnforcement.Name, testResourceNamePrefix) {
				if !sweepMfaEnforcement(conf, mfaEnforcement.Id) {
					result = multierror.Append(result, fmt.Errorf("failed to sweep mfa enforcement %q", mfaEnforcement.Id))
				}
			}
		}
	}

	return result.ErrorOrNil()
}

func sweepMfaEnforcement(conf *provider_config.Config, id string) bool {
	return test.SweepWithRetry(sweepMfaEnforcementOnce, conf, "Mfa Enforcement", id)
}

func sweepMfaEnforcementOnce(conf *provider_config.Config, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), sweepDeleteTimeout)
	defer cancel()

	client := organizationmanagersdk.NewMfaEnforcementClient(conf.SDKv2)
	op, err := client.Delete(ctx, &organizationmanager.DeleteMfaEnforcementRequest{
		MfaEnforcementId: id,
	})
	if err != nil {
		return err
	}

	_, err = op.Wait(ctx)
	return err
}
