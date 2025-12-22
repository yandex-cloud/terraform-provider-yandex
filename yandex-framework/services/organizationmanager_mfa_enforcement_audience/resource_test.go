package organizationmanager_mfa_enforcement_audience_test

import (
	"context"
	"fmt"
	"testing"

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
	subjectId := test.GetExampleUserID1()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMfaEnforcement(organizationId) + testAccMfaEnforcementAudience(subjectId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMfaEnforcementWithAudienceExists("yandex_organizationmanager_mfa_enforcement.foo", "yandex_organizationmanager_mfa_enforcement_audience.bar"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_mfa_enforcement.foo", "organization_id", organizationId),
					resource.TestCheckResourceAttr("yandex_organizationmanager_mfa_enforcement_audience.bar", "subject_id", subjectId),
				),
			},
		},
	})
}

func TestAccOrganizationManagerMfaEnforcementAudienceRecreateForNewSubjectId(t *testing.T) {
	organizationId := test.GetExampleOrganizationID()
	subjectId1 := test.GetExampleUserID1()
	subjectId2 := test.GetExampleUserID2()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMfaEnforcement(organizationId) + testAccMfaEnforcementAudience(subjectId1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMfaEnforcementWithAudienceExists("yandex_organizationmanager_mfa_enforcement.foo", "yandex_organizationmanager_mfa_enforcement_audience.bar"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_mfa_enforcement_audience.bar", "subject_id", subjectId1),
				),
			},
			{
				Config: testAccMfaEnforcement(organizationId) + testAccMfaEnforcementAudience(subjectId2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMfaEnforcementWithAudienceExists("yandex_organizationmanager_mfa_enforcement.foo", "yandex_organizationmanager_mfa_enforcement_audience.bar"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_mfa_enforcement_audience.bar", "subject_id", subjectId2),
				),
			},
		},
	})
}

func TestAccOrganizationManagerMfaEnforcementAudienceDelete(t *testing.T) {
	organizationId := test.GetExampleOrganizationID()
	subjectId := test.GetExampleUserID1()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMfaEnforcement(organizationId) + testAccMfaEnforcementAudience(subjectId),
				Check:  testAccCheckMfaEnforcementWithAudienceExists("yandex_organizationmanager_mfa_enforcement.foo", "yandex_organizationmanager_mfa_enforcement_audience.bar"),
			},
			{
				Config: testAccMfaEnforcement(organizationId),
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
	ttl 		    = "5s"
	status 		    = "MFA_ENFORCEMENT_STATUS_ACTIVE"
	enroll_window   = "5h0m0s"
}
`, organizationId)
}

func testAccMfaEnforcementAudience(subjectId string) string {
	return fmt.Sprintf(`
resource "yandex_organizationmanager_mfa_enforcement_audience" "bar" {
	mfa_enforcement_id = yandex_organizationmanager_mfa_enforcement.foo.id
	subject_id = "%s"
}
`, subjectId)
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
