package organizationmanager_idp_application_saml_application_assignment_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	saml "github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1/idp/application/saml"
	samlsdk "github.com/yandex-cloud/go-sdk/services/organizationmanager/v1/idp/application/saml"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestAccOrganizationManagerIdpApplicationSamlApplicationAssignmentCreate(t *testing.T) {
	organizationId := test.GetExampleOrganizationID()
	subjectId := test.GetExampleUserID1()
	appName := acctest.RandomWithPrefix("tf-acc-test-saml-app")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSamlApplication(organizationId, appName) + testAccSamlApplicationAssignment(subjectId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSamlApplicationWithAssignmentExists("yandex_organizationmanager_idp_application_saml_application.foo", "yandex_organizationmanager_idp_application_saml_application_assignment.bar"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_application_saml_application.foo", "organization_id", organizationId),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_application_saml_application_assignment.bar", "subject_id", subjectId),
				),
			},
		},
	})
}

func TestAccOrganizationManagerIdpApplicationSamlApplicationAssignmentRecreateForNewSubjectId(t *testing.T) {
	organizationId := test.GetExampleOrganizationID()
	subjectId1 := test.GetExampleUserID1()
	subjectId2 := test.GetExampleUserID2()
	appName := acctest.RandomWithPrefix("tf-acc-test-saml-app")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSamlApplication(organizationId, appName) + testAccSamlApplicationAssignment(subjectId1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSamlApplicationWithAssignmentExists("yandex_organizationmanager_idp_application_saml_application.foo", "yandex_organizationmanager_idp_application_saml_application_assignment.bar"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_application_saml_application_assignment.bar", "subject_id", subjectId1),
				),
			},
			{
				Config: testAccSamlApplication(organizationId, appName) + testAccSamlApplicationAssignment(subjectId2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSamlApplicationWithAssignmentExists("yandex_organizationmanager_idp_application_saml_application.foo", "yandex_organizationmanager_idp_application_saml_application_assignment.bar"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_application_saml_application_assignment.bar", "subject_id", subjectId2),
				),
			},
		},
	})
}

func TestAccOrganizationManagerIdpApplicationSamlApplicationAssignmentDelete(t *testing.T) {
	organizationId := test.GetExampleOrganizationID()
	subjectId := test.GetExampleUserID1()
	appName := acctest.RandomWithPrefix("tf-acc-test-saml-app")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSamlApplication(organizationId, appName) + testAccSamlApplicationAssignment(subjectId),
				Check:  testAccCheckSamlApplicationWithAssignmentExists("yandex_organizationmanager_idp_application_saml_application.foo", "yandex_organizationmanager_idp_application_saml_application_assignment.bar"),
			},
			{
				Config: testAccSamlApplication(organizationId, appName),
				Check:  testAccCheckSamlApplicationWithAssignmentExists("yandex_organizationmanager_idp_application_saml_application.foo"),
			},
		},
	})
}

func testAccSamlApplication(organizationId, appName string) string {
	return fmt.Sprintf(`
resource "yandex_organizationmanager_idp_application_saml_application" "foo" {
  organization_id = "%s"
  name            = "%s"
  service_provider = {
    entity_id = "https://example.com/saml/metadata"
    acs_urls = [{
      url   = "https://example.com/saml/acs"
      index = 0
    }]
  }
  attribute_mapping = {
    name_id = {
      format = "EMAIL"
    }
  }
}
`, organizationId, appName)
}

func testAccSamlApplicationAssignment(subjectId string) string {
	return fmt.Sprintf(`
resource "yandex_organizationmanager_idp_application_saml_application_assignment" "bar" {
	application_id = yandex_organizationmanager_idp_application_saml_application.foo.application_id
	subject_id = "%s"
}
`, subjectId)
}

func testAccCheckSamlApplicationWithAssignmentExists(application string, assignments ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()
		client := samlsdk.NewApplicationClient(config.SDKv2)
		applicationRS, err := resourceState(s, application)
		if err != nil {
			return err
		}
		applicationId := applicationRS.Primary.Attributes["application_id"]
		if applicationId == "" {
			applicationId = applicationRS.Primary.ID
		}
		resp1, err := client.Get(context.Background(), &saml.GetApplicationRequest{
			ApplicationId: applicationId,
		})
		if err != nil {
			return err
		}
		if resp1.Id != applicationId {
			return fmt.Errorf("SAML application %s not found", application)
		}
		resp2, err := client.ListAssignments(context.Background(), &saml.ListAssignmentsRequest{
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
			return fmt.Errorf("invalid SAML application's assignments: expected '%s', got '%s'", expected, got)
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
				return fmt.Errorf("assignment '%s' for SAML application '%s' not found", subjectId, applicationId)
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
		if rs.Type == "yandex_organizationmanager_idp_application_saml_application" {
			_, err = samlsdk.NewApplicationClient(config.SDKv2).Get(context.Background(), &saml.GetApplicationRequest{
				ApplicationId: rs.Primary.Attributes["application_id"],
			})
		}
		if err != nil {
			if grpcStatus, ok := status.FromError(err); ok && grpcStatus != nil && grpcStatus.Code() == codes.NotFound {
				return nil
			} else if ok {
				return fmt.Errorf("error while requesting Yandex Cloud: grpc code error : %d, http message error: %s", grpcStatus.Code(), grpcStatus.Message())
			}
			return fmt.Errorf("SAML application '%s' still exists", rs.Primary.ID)
		}
	}

	return nil
}
