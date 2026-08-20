package organizationmanager_mfa_enforcement_audience_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	organizationmanager "github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1"
	organizationmanagersdk "github.com/yandex-cloud/go-sdk/services/organizationmanager/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestAccOrganizationManagerMfaEnforcementAudienceCreate(t *testing.T) {
	organizationId := test.GetExampleOrganizationID()
	suffix := acctest.RandString(10)
	subjectConfig := testAccMfaAudienceSubject(organizationId, suffix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: subjectConfig + testAccMfaEnforcement(organizationId) + testAccMfaEnforcementAudience("yandex_organizationmanager_idp_user.subject1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMfaEnforcementWithAudienceExists("yandex_organizationmanager_mfa_enforcement.foo", "yandex_organizationmanager_mfa_enforcement_audience.bar"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_mfa_enforcement.foo", "organization_id", organizationId),
					resource.TestCheckResourceAttrPair("yandex_organizationmanager_mfa_enforcement_audience.bar", "subject_id", "yandex_organizationmanager_idp_user.subject1", "user_id"),
				),
			},
		},
	})
}

func TestAccOrganizationManagerMfaEnforcementAudienceRecreateForNewSubjectId(t *testing.T) {
	organizationId := test.GetExampleOrganizationID()
	suffix := acctest.RandString(10)
	subjectsConfig := testAccMfaAudienceSubject(organizationId, suffix) + testAccMfaAudienceSecondSubject(suffix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: subjectsConfig + testAccMfaEnforcement(organizationId) + testAccMfaEnforcementAudience("yandex_organizationmanager_idp_user.subject1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMfaEnforcementWithAudienceExists("yandex_organizationmanager_mfa_enforcement.foo", "yandex_organizationmanager_mfa_enforcement_audience.bar"),
					resource.TestCheckResourceAttrPair("yandex_organizationmanager_mfa_enforcement_audience.bar", "subject_id", "yandex_organizationmanager_idp_user.subject1", "user_id"),
				),
			},
			{
				Config: subjectsConfig + testAccMfaEnforcement(organizationId) + testAccMfaEnforcementAudience("yandex_organizationmanager_idp_user.subject2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMfaEnforcementWithAudienceExists("yandex_organizationmanager_mfa_enforcement.foo", "yandex_organizationmanager_mfa_enforcement_audience.bar"),
					resource.TestCheckResourceAttrPair("yandex_organizationmanager_mfa_enforcement_audience.bar", "subject_id", "yandex_organizationmanager_idp_user.subject2", "user_id"),
				),
			},
		},
	})
}

func TestAccOrganizationManagerMfaEnforcementAudienceDelete(t *testing.T) {
	organizationId := test.GetExampleOrganizationID()
	suffix := acctest.RandString(10)
	subjectConfig := testAccMfaAudienceSubject(organizationId, suffix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: subjectConfig + testAccMfaEnforcement(organizationId) + testAccMfaEnforcementAudience("yandex_organizationmanager_idp_user.subject1"),
				Check:  testAccCheckMfaEnforcementWithAudienceExists("yandex_organizationmanager_mfa_enforcement.foo", "yandex_organizationmanager_mfa_enforcement_audience.bar"),
			},
			{
				Config: subjectConfig + testAccMfaEnforcement(organizationId),
				Check:  testAccCheckMfaEnforcementWithAudienceExists("yandex_organizationmanager_mfa_enforcement.foo"),
			},
		},
	})
}

func testAccMfaEnforcement(organizationId string) string {
	return fmt.Sprintf(`
resource "yandex_organizationmanager_mfa_enforcement" "foo" {
	name            = "test-mfa-enforcement-name"
	organization_id = "%s"
	acr_id 		    = "any-mfa"
	ttl 		    = "5m0s"
	status 		    = "MFA_ENFORCEMENT_STATUS_ACTIVE"
	enroll_window   = "5h0m0s"
}
`, organizationId)
}

func testAccMfaEnforcementAudience(subjectResource string) string {
	return fmt.Sprintf(`
resource "yandex_organizationmanager_mfa_enforcement_audience" "bar" {
	mfa_enforcement_id = yandex_organizationmanager_mfa_enforcement.foo.id
	subject_id = %s.user_id
}
`, subjectResource)
}

func testAccMfaAudienceSubject(organizationID, suffix string) string {
	return fmt.Sprintf(`
resource "yandex_organizationmanager_idp_userpool" "audience" {
  name              = "tf-acc-test-userpool-mfa-%[1]s"
  organization_id   = "%[2]s"
  default_subdomain = "tf-acc-mfa-%[1]s"
}

resource "yandex_organizationmanager_idp_user" "subject1" {
  userpool_id = yandex_organizationmanager_idp_userpool.audience.userpool_id
  username    = "subject1@tf-acc-mfa-%[1]s.idp.yandexcloud.net"
  full_name   = "MFA Test Subject One"
  email       = "subject1-%[1]s@example.com"
  is_active   = true
  password_spec = {
    password = "MfaTest195!-%[1]s"
  }
}
`, suffix, organizationID)
}

func testAccMfaAudienceSecondSubject(suffix string) string {
	return fmt.Sprintf(`
resource "yandex_organizationmanager_idp_user" "subject2" {
  userpool_id = yandex_organizationmanager_idp_userpool.audience.userpool_id
  username    = "subject2@tf-acc-mfa-%[1]s.idp.yandexcloud.net"
  full_name   = "MFA Test Subject Two"
  email       = "subject2-%[1]s@example.com"
  is_active   = true
  password_spec = {
    password = "MfaTest195!-%[1]s"
  }
}
`, suffix)
}

func testAccCheckMfaEnforcementWithAudienceExists(mfaEnforcement string, audiences ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()
		client := organizationmanagersdk.NewMfaEnforcementClient(config.SDKv2)
		mfaEnforcementRS, err := resourceState(s, mfaEnforcement)
		if err != nil {
			return err
		}
		mfaEnforcementId := mfaEnforcementRS.Primary.ID
		resp1, err := client.Get(context.Background(), &organizationmanager.GetMfaEnforcementRequest{
			MfaEnforcementId: mfaEnforcementId,
		})
		if err != nil {
			return err
		}
		if resp1.Id != mfaEnforcementId {
			return fmt.Errorf("MFA enforcement %s not found", mfaEnforcement)
		}
		resp2, err := client.ListAudience(context.Background(), &organizationmanager.ListAudienceRequest{
			MfaEnforcementId: mfaEnforcementId,
		})
		if err != nil {
			return err
		}
		if len(audiences) != len(resp2.Subjects) {
			expected := ""
			for _, s := range resp2.Subjects {
				if expected != "" {
					expected += ", "
				}
				expected += s.Id
			}
			got := ""
			for _, audience := range audiences {
				audienceRS, err := resourceState(s, audience)
				if err != nil {
					return err
				}
				subjectId := audienceRS.Primary.Attributes["subject_id"]
				if got != "" {
					got += ", "
				}
				got += subjectId
			}
			return fmt.Errorf("invalid MFA enforcement's audience: expected '%s', got '%s'", expected, got)
		}
		for _, audience := range audiences {
			audienceRS, err := resourceState(s, audience)
			if err != nil {
				return err
			}
			if audienceRS.Primary.Attributes["mfa_enforcement_id"] != mfaEnforcementId {
				return fmt.Errorf("invalid MFA enforcement id in audience: expected '%s', got '%s'", mfaEnforcementId, audienceRS.Primary.Attributes["mfa_enforcement_id"])
			}
			subjectId := audienceRS.Primary.Attributes["subject_id"]
			found := false
			for _, s := range resp2.Subjects {
				if s.Id == subjectId {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("audience '%s' for MFA enforcement '%s' not found", subjectId, mfaEnforcementId)
			}
		}
		return nil
	}
}

func resourceState(s *terraform.State, resourceName string) (*terraform.ResourceState, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource '%s' not found", resourceName)
	}
	if rs.Primary.ID == "" {
		return nil, fmt.Errorf("no id is set for resource '%s'", resourceName)
	}
	return rs, nil
}

func testAccCheckDestroy(s *terraform.State) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()
	var err error
	for _, rs := range s.RootModule().Resources {
		if rs.Type == "yandex_organizationmanager_mfa_enforcement" {
			_, err = organizationmanagersdk.NewMfaEnforcementClient(config.SDKv2).Get(context.Background(), &organizationmanager.GetMfaEnforcementRequest{
				MfaEnforcementId: rs.Primary.ID,
			})
		}
		if err != nil {
			if grpcStatus, ok := status.FromError(err); ok && grpcStatus != nil && grpcStatus.Code() == codes.NotFound {
				return nil
			} else if ok {
				return fmt.Errorf("error while requesting Yandex Cloud: grpc code error : %d, http message error: %s", grpcStatus.Code(), grpcStatus.Message())
			}
			return fmt.Errorf("MFA enforcement '%s' still exists", rs.Primary.ID)
		}
	}

	return nil
}
