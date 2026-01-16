package organizationmanager_idp_application_oauth_application_assignment_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	oauth "github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1/idp/application/oauth"
	oauthsdk "github.com/yandex-cloud/go-sdk/services/organizationmanager/v1/idp/application/oauth"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const oauthClientFolderID = "b1g6bop87aoiekbkko82"

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestAccOrganizationManagerIdpApplicationOauthApplicationAssignmentCreate(t *testing.T) {
	organizationId := test.GetExampleOrganizationID()
	subjectId := test.GetExampleUserID1()
	appName := acctest.RandomWithPrefix("tf-acc-test-oauth-app")
	clientName := acctest.RandomWithPrefix("tf-acc-test-oauth-client")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOauthApplication(organizationId, appName, clientName) + testAccOauthApplicationAssignment(subjectId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOauthApplicationWithAssignmentExists("yandex_organizationmanager_idp_application_oauth_application.foo", "yandex_organizationmanager_idp_application_oauth_application_assignment.bar"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_application_oauth_application.foo", "organization_id", organizationId),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_application_oauth_application_assignment.bar", "subject_id", subjectId),
				),
			},
		},
	})
}

func TestAccOrganizationManagerIdpApplicationOauthApplicationAssignmentRecreateForNewSubjectId(t *testing.T) {
	organizationId := test.GetExampleOrganizationID()
	subjectId1 := test.GetExampleUserID1()
	subjectId2 := test.GetExampleUserID2()
	appName := acctest.RandomWithPrefix("tf-acc-test-oauth-app")
	clientName := acctest.RandomWithPrefix("tf-acc-test-oauth-client")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOauthApplication(organizationId, appName, clientName) + testAccOauthApplicationAssignment(subjectId1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOauthApplicationWithAssignmentExists("yandex_organizationmanager_idp_application_oauth_application.foo", "yandex_organizationmanager_idp_application_oauth_application_assignment.bar"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_application_oauth_application_assignment.bar", "subject_id", subjectId1),
				),
			},
			{
				Config: testAccOauthApplication(organizationId, appName, clientName) + testAccOauthApplicationAssignment(subjectId2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOauthApplicationWithAssignmentExists("yandex_organizationmanager_idp_application_oauth_application.foo", "yandex_organizationmanager_idp_application_oauth_application_assignment.bar"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_application_oauth_application_assignment.bar", "subject_id", subjectId2),
				),
			},
		},
	})
}

func TestAccOrganizationManagerIdpApplicationOauthApplicationAssignmentDelete(t *testing.T) {
	organizationId := test.GetExampleOrganizationID()
	subjectId := test.GetExampleUserID1()
	appName := acctest.RandomWithPrefix("tf-acc-test-oauth-app")
	clientName := acctest.RandomWithPrefix("tf-acc-test-oauth-client")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOauthApplication(organizationId, appName, clientName) + testAccOauthApplicationAssignment(subjectId),
				Check:  testAccCheckOauthApplicationWithAssignmentExists("yandex_organizationmanager_idp_application_oauth_application.foo", "yandex_organizationmanager_idp_application_oauth_application_assignment.bar"),
			},
			{
				Config: testAccOauthApplication(organizationId, appName, clientName),
				Check:  testAccCheckOauthApplicationWithAssignmentExists("yandex_organizationmanager_idp_application_oauth_application.foo"),
			},
		},
	})
}

func testAccOauthApplication(organizationId, appName, clientName string) string {
	return fmt.Sprintf(`
resource "yandex_iam_oauth_client" "test_client" {
  name      = "%s"
  folder_id = "%s"
  scopes    = ["iam"]
}

resource "yandex_organizationmanager_idp_application_oauth_application" "foo" {
  organization_id = "%s"
  name            = "%s"
  client_grant = {
    client_id         = yandex_iam_oauth_client.test_client.id
    authorized_scopes  = ["openid", "profile", "email"]
  }
  group_claims_settings = {
    group_distribution_type = "ALL_GROUPS"
  }
}
`, clientName, oauthClientFolderID, organizationId, appName)
}

func testAccOauthApplicationAssignment(subjectId string) string {
	return fmt.Sprintf(`
resource "yandex_organizationmanager_idp_application_oauth_application_assignment" "bar" {
	application_id = yandex_organizationmanager_idp_application_oauth_application.foo.application_id
	subject_id = "%s"
}
`, subjectId)
}

func testAccCheckOauthApplicationWithAssignmentExists(application string, assignments ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()
		client := oauthsdk.NewApplicationClient(config.SDKv2)
		applicationRS, err := resourceState(s, application)
		if err != nil {
			return err
		}
		applicationId := applicationRS.Primary.Attributes["application_id"]
		if applicationId == "" {
			applicationId = applicationRS.Primary.ID
		}
		resp1, err := client.Get(context.Background(), &oauth.GetApplicationRequest{
			ApplicationId: applicationId,
		})
		if err != nil {
			return err
		}
		if resp1.Id != applicationId {
			return fmt.Errorf("OAuth application %s not found", application)
		}
		resp2, err := client.ListAssignments(context.Background(), &oauth.ListAssignmentsRequest{
			ApplicationId: applicationId,
			PageSize:      100,
		})
		if err != nil {
			return err
		}
		if len(assignments) != len(resp2.Assignments) {
			expected := ""
			for _, a := range resp2.Assignments {
				if expected != "" {
					expected += ", "
				}
				expected += a.SubjectId
			}
			got := ""
			for _, assignment := range assignments {
				assignmentRS, err := resourceState(s, assignment)
				if err != nil {
					return err
				}
				subjectId := assignmentRS.Primary.Attributes["subject_id"]
				if got != "" {
					got += ", "
				}
				got += subjectId
			}
			return fmt.Errorf("invalid OAuth application's assignments: expected '%s', got '%s'", expected, got)
		}
		for _, assignment := range assignments {
			assignmentRS, err := resourceState(s, assignment)
			if err != nil {
				return err
			}
			if assignmentRS.Primary.Attributes["application_id"] != applicationId {
				return fmt.Errorf("invalid application id in assignment: expected '%s', got '%s'", applicationId, assignmentRS.Primary.Attributes["application_id"])
			}
			subjectId := assignmentRS.Primary.Attributes["subject_id"]
			found := false
			for _, a := range resp2.Assignments {
				if a.SubjectId == subjectId {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("assignment '%s' for OAuth application '%s' not found", subjectId, applicationId)
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
		if rs.Type == "yandex_organizationmanager_idp_application_oauth_application" {
			_, err = oauthsdk.NewApplicationClient(config.SDKv2).Get(context.Background(), &oauth.GetApplicationRequest{
				ApplicationId: rs.Primary.Attributes["application_id"],
			})
		}
		if err != nil {
			if grpcStatus, ok := status.FromError(err); ok && grpcStatus != nil && grpcStatus.Code() == codes.NotFound {
				return nil
			} else if ok {
				return fmt.Errorf("error while requesting Yandex Cloud: grpc code error : %d, http message error: %s", grpcStatus.Code(), grpcStatus.Message())
			}
			return fmt.Errorf("OAuth application '%s' still exists", rs.Primary.ID)
		}
	}

	return nil
}
