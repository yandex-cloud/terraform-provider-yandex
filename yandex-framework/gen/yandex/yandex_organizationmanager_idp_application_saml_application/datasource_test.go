package yandex_organizationmanager_idp_application_saml_application_test

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

func TestAccDataSourceOrganizationManagerIdpApplicationSamlApplication_byID(t *testing.T) {
	appName := acctest.RandomWithPrefix("tf-acc-test-saml-app")
	organizationID := test.GetExampleOrganizationID()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckIdpSamlApplicationDataSourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIdpSamlApplicationConfig(appName, organizationID, true),
				Check: resource.ComposeTestCheckFunc(
					test.AccCheckResourceIDField("data.yandex_organizationmanager_idp_application_saml_application.source", "application_id"),
					resource.TestCheckResourceAttrSet("data.yandex_organizationmanager_idp_application_saml_application.source", "id"),
					resource.TestCheckResourceAttr("data.yandex_organizationmanager_idp_application_saml_application.source", "name", appName),
					resource.TestCheckResourceAttr("data.yandex_organizationmanager_idp_application_saml_application.source", "organization_id", organizationID),
					resource.TestCheckResourceAttr("data.yandex_organizationmanager_idp_application_saml_application.source", "description", "Test SAML application"),
					resource.TestCheckResourceAttrSet("data.yandex_organizationmanager_idp_application_saml_application.source", "application_id"),
					resource.TestCheckResourceAttrSet("data.yandex_organizationmanager_idp_application_saml_application.source", "status"),
					resource.TestCheckResourceAttrSet("data.yandex_organizationmanager_idp_application_saml_application.source", "created_at"),
					resource.TestCheckResourceAttrSet("data.yandex_organizationmanager_idp_application_saml_application.source", "updated_at"),
					resource.TestCheckResourceAttr("data.yandex_organizationmanager_idp_application_saml_application.source", "service_provider.entity_id", "https://example.com/saml/metadata"),
					resource.TestCheckResourceAttr("data.yandex_organizationmanager_idp_application_saml_application.source", "service_provider.acs_urls.0.url", "https://example.com/saml/acs"),
					resource.TestCheckResourceAttr("data.yandex_organizationmanager_idp_application_saml_application.source", "service_provider.slo_urls.0.url", "https://example.com/saml/slo"),
					resource.TestCheckResourceAttr("data.yandex_organizationmanager_idp_application_saml_application.source", "service_provider.slo_urls.0.protocol_binding", "HTTP_POST"),
					resource.TestCheckResourceAttr("data.yandex_organizationmanager_idp_application_saml_application.source", "attribute_mapping.name_id.format", "EMAIL"),
					resource.TestCheckResourceAttr("data.yandex_organizationmanager_idp_application_saml_application.source", "attribute_mapping.attributes.0.name", "email"),
					resource.TestCheckResourceAttr("data.yandex_organizationmanager_idp_application_saml_application.source", "attribute_mapping.attributes.0.value", "SubjectClaims.email"),
					resource.TestCheckResourceAttr("data.yandex_organizationmanager_idp_application_saml_application.source", "security_settings.signature_mode", "ASSERTIONS"),
					resource.TestCheckResourceAttr("data.yandex_organizationmanager_idp_application_saml_application.source", "group_claims_settings.group_attribute_name", "groups"),
					resource.TestCheckResourceAttr("data.yandex_organizationmanager_idp_application_saml_application.source", "group_claims_settings.group_distribution_type", "ASSIGNED_GROUPS"),
					resource.TestCheckResourceAttr("data.yandex_organizationmanager_idp_application_saml_application.source", "labels.env", "test"),
					resource.TestCheckResourceAttr("data.yandex_organizationmanager_idp_application_saml_application.source", "labels.app", "saml"),
					test.AccCheckCreatedAtAttr("data.yandex_organizationmanager_idp_application_saml_application.source"),
				),
			},
		},
	})
}

func testAccDataSourceIdpSamlApplicationConfig(appName, organizationID string, useID bool) string {
	if useID {
		return testAccDataSourceIdpSamlApplicationResourceConfig(appName, organizationID) + idpSamlApplicationDataByIDConfig
	}

	return testAccDataSourceIdpSamlApplicationResourceConfig(appName, organizationID)
}

func testAccDataSourceIdpSamlApplicationResourceConfig(appName, organizationID string) string {
	return fmt.Sprintf(`
resource "yandex_organizationmanager_idp_application_saml_application" "foobar" {
  name           = "%s"
  organization_id = "%s"
  description    = "Test SAML application"

  service_provider = {
    entity_id = "https://example.com/saml/metadata"

    acs_urls = [{
      url = "https://example.com/saml/acs"
    }]

    slo_urls = [{
      url              = "https://example.com/saml/slo"
      protocol_binding = "HTTP_POST"
    }]
  }

  attribute_mapping = {
    name_id = {
      format = "EMAIL"
    }

    attributes = [{
      name  = "email"
      value = "SubjectClaims.email"
    }]
  }

  security_settings = {
    signature_mode = "ASSERTIONS"
  }

  group_claims_settings = {
    group_attribute_name   = "groups"
    group_distribution_type = "ASSIGNED_GROUPS"
  }

  labels = {
    env = "test"
    app = "saml"
  }
}
`, appName, organizationID)
}

const idpSamlApplicationDataByIDConfig = `
data "yandex_organizationmanager_idp_application_saml_application" "source" {
  application_id = yandex_organizationmanager_idp_application_saml_application.foobar.application_id
}
`

func testAccCheckIdpSamlApplicationDataSourceDestroy(s *terraform.State) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type == "yandex_organizationmanager_idp_application_saml_application" {
			_, err := samlsdk.NewApplicationClient(config.SDKv2).Get(context.Background(), &saml.GetApplicationRequest{
				ApplicationId: rs.Primary.ID,
			})

			if err != nil {
				if grpcStatus, ok := status.FromError(err); ok && grpcStatus != nil && grpcStatus.Code() == codes.NotFound {
					continue
				} else if ok {
					return fmt.Errorf("Error while requesting Yandex Cloud: grpc code error : %d, http message error: %s", grpcStatus.Code(), grpcStatus.Message())
				}
				return fmt.Errorf("IdpSamlApplication still exists")
			}
			return fmt.Errorf("IdpSamlApplication still exists")
		}
	}

	return nil
}
