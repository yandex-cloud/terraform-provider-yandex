package yandex_organizationmanager_idp_application_oauth_application_test

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

func TestAccDataSourceOrganizationManagerIdpApplicationOauthApplication_byID(t *testing.T) {
	appName := acctest.RandomWithPrefix("tf-acc-test-oauth-app")
	organizationID := test.GetExampleOrganizationID()
	clientName := acctest.RandomWithPrefix("tf-acc-test-oauth-client")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckIdpApplicationOauthApplicationDataSourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIdpApplicationOauthApplicationConfig(appName, organizationID, clientName, oauthClientFolderID, true),
				Check: resource.ComposeTestCheckFunc(
					test.AccCheckResourceIDField("data.yandex_organizationmanager_idp_application_oauth_application.source", "application_id"),
					resource.TestCheckResourceAttr("data.yandex_organizationmanager_idp_application_oauth_application.source", "name", appName),
					resource.TestCheckResourceAttrSet("data.yandex_organizationmanager_idp_application_oauth_application.source", "id"),
					resource.TestCheckResourceAttr("data.yandex_organizationmanager_idp_application_oauth_application.source", "organization_id", organizationID),
					resource.TestCheckResourceAttr("data.yandex_organizationmanager_idp_application_oauth_application.source", "description", "Test OAuth application description"),
					resource.TestCheckResourceAttr("data.yandex_organizationmanager_idp_application_oauth_application.source", "labels.env", "test"),
					resource.TestCheckResourceAttr("data.yandex_organizationmanager_idp_application_oauth_application.source", "labels.app", "test-app"),
					resource.TestCheckResourceAttrSet("data.yandex_organizationmanager_idp_application_oauth_application.source", "client_grant.client_id"),
					resource.TestCheckResourceAttr("data.yandex_organizationmanager_idp_application_oauth_application.source", "client_grant.authorized_scopes.#", "3"),
					resource.TestCheckTypeSetElemAttr("data.yandex_organizationmanager_idp_application_oauth_application.source", "client_grant.authorized_scopes.*", "openid"),
					resource.TestCheckTypeSetElemAttr("data.yandex_organizationmanager_idp_application_oauth_application.source", "client_grant.authorized_scopes.*", "profile"),
					resource.TestCheckTypeSetElemAttr("data.yandex_organizationmanager_idp_application_oauth_application.source", "client_grant.authorized_scopes.*", "email"),
					resource.TestCheckResourceAttrSet("data.yandex_organizationmanager_idp_application_oauth_application.source", "group_claims_settings.group_distribution_type"),
					resource.TestCheckResourceAttrSet("data.yandex_organizationmanager_idp_application_oauth_application.source", "status"),
					test.AccCheckCreatedAtAttr("data.yandex_organizationmanager_idp_application_oauth_application.source"),
				),
			},
		},
	})
}

func testAccDataSourceIdpApplicationOauthApplicationConfig(appName, organizationID, clientName, folderID string, useID bool) string {
	if useID {
		return testAccDataSourceIdpApplicationOauthApplicationResourceConfig(appName, organizationID, clientName, folderID) + idpApplicationOauthApplicationDataByIDConfig
	}

	return testAccDataSourceIdpApplicationOauthApplicationResourceConfig(appName, organizationID, clientName, folderID)
}

func testAccDataSourceIdpApplicationOauthApplicationResourceConfig(appName, organizationID, clientName, folderID string) string {
	return fmt.Sprintf(`
resource "yandex_iam_oauth_client" "test_client" {
  name       = "%s"
  folder_id  = "%s"
  scopes     = ["iam"]
}

resource "yandex_organizationmanager_idp_application_oauth_application" "foobar" {
  organization_id = "%s"
  name            = "%s"
  description     = "Test OAuth application description"

  client_grant = {
    client_id         = yandex_iam_oauth_client.test_client.id
    authorized_scopes = ["openid", "profile", "email"]
  }

  group_claims_settings = {
    group_distribution_type = "ALL_GROUPS"
  }

  labels = {
    env = "test"
    app = "test-app"
  }
}
`, clientName, folderID, organizationID, appName)
}

const idpApplicationOauthApplicationDataByIDConfig = `
data "yandex_organizationmanager_idp_application_oauth_application" "source" {
  application_id = "${yandex_organizationmanager_idp_application_oauth_application.foobar.application_id}"
}
`

func testAccCheckIdpApplicationOauthApplicationDataSourceDestroy(s *terraform.State) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_organizationmanager_idp_application_oauth_application" {
			continue
		}

		_, err := oauthsdk.NewApplicationClient(config.SDKv2).Get(context.Background(), &oauth.GetApplicationRequest{
			ApplicationId: rs.Primary.ID,
		})

		if err != nil {
			if grpcStatus, ok := status.FromError(err); ok && grpcStatus != nil && grpcStatus.Code() == codes.NotFound {
				return nil
			} else if ok {
				return fmt.Errorf("Error while requesting Yandex Cloud: grpc code error : %d, http message error: %s", grpcStatus.Code(), grpcStatus.Message())
			}
			return fmt.Errorf("IdpApplicationOauthApplication still exists")
		}
	}

	return nil
}
