package yandex_organizationmanager_idp_application_oauth_application_test

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
	oauth "github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1/idp/application/oauth"
	oauthsdk "github.com/yandex-cloud/go-sdk/services/organizationmanager/v1/idp/application/oauth"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	applicationSweepPageSize      = 1000
	applicationSweepDeleteTimeout = 15 * time.Minute
	testResourceNamePrefix        = "tf-acc-test-oauth-app"
	oauthClientFolderID           = "b1g6bop87aoiekbkko82"
)

func init() {
	resource.AddTestSweepers("yandex_organizationmanager_idp_application_oauth_application", &resource.Sweeper{
		Name:         "yandex_organizationmanager_idp_application_oauth_application",
		F:            testSweepIdpApplicationOauthApplication,
		Dependencies: []string{},
	})
}

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestAccOrganizationManagerIdpApplicationOauthApplication_basic(t *testing.T) {
	appName := acctest.RandomWithPrefix("tf-acc-test-oauth-app")
	organizationID := test.GetExampleOrganizationID()
	clientName := acctest.RandomWithPrefix("tf-acc-test-oauth-client")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckIdpApplicationOauthApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIdpApplicationOauthApplication_full(appName, organizationID, clientName, oauthClientFolderID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIdpApplicationOauthApplicationExists("yandex_organizationmanager_idp_application_oauth_application.foobar"),
					test.AccCheckCreatedAtAttr("yandex_organizationmanager_idp_application_oauth_application.foobar"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_application_oauth_application.foobar", "name", appName),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_application_oauth_application.foobar", "organization_id", organizationID),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_application_oauth_application.foobar", "description", "Test OAuth application description"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_application_oauth_application.foobar", "labels.env", "test"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_application_oauth_application.foobar", "labels.app", "test-app"),
					resource.TestCheckResourceAttrSet("yandex_organizationmanager_idp_application_oauth_application.foobar", "application_id"),
					resource.TestCheckResourceAttrSet("yandex_organizationmanager_idp_application_oauth_application.foobar", "client_grant.client_id"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_application_oauth_application.foobar", "client_grant.authorized_scopes.#", "3"),
					resource.TestCheckTypeSetElemAttr("yandex_organizationmanager_idp_application_oauth_application.foobar", "client_grant.authorized_scopes.*", "openid"),
					resource.TestCheckTypeSetElemAttr("yandex_organizationmanager_idp_application_oauth_application.foobar", "client_grant.authorized_scopes.*", "profile"),
					resource.TestCheckTypeSetElemAttr("yandex_organizationmanager_idp_application_oauth_application.foobar", "client_grant.authorized_scopes.*", "email"),
					resource.TestCheckResourceAttrSet("yandex_organizationmanager_idp_application_oauth_application.foobar", "group_claims_settings.group_distribution_type"),
					resource.TestCheckResourceAttrSet("yandex_organizationmanager_idp_application_oauth_application.foobar", "status"),
				),
			},
		},
	})
}

func TestAccOrganizationManagerIdpApplicationOauthApplication_update(t *testing.T) {
	appName := acctest.RandomWithPrefix("tf-acc-test-oauth-app")
	organizationID := test.GetExampleOrganizationID()
	clientName := acctest.RandomWithPrefix("tf-acc-test-oauth-client")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckIdpApplicationOauthApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIdpApplicationOauthApplication_full(appName, organizationID, clientName, oauthClientFolderID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIdpApplicationOauthApplicationExists("yandex_organizationmanager_idp_application_oauth_application.foobar"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_application_oauth_application.foobar", "name", appName),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_application_oauth_application.foobar", "description", "Test OAuth application description"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_application_oauth_application.foobar", "labels.env", "test"),
				),
			},
			{
				Config: testAccIdpApplicationOauthApplication_update(appName+"-updated", organizationID, clientName, oauthClientFolderID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIdpApplicationOauthApplicationExists("yandex_organizationmanager_idp_application_oauth_application.foobar"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_application_oauth_application.foobar", "name", appName+"-updated"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_application_oauth_application.foobar", "description", "Updated OAuth application description"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_application_oauth_application.foobar", "labels.env", "production"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_application_oauth_application.foobar", "labels.app", "updated-app"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_application_oauth_application.foobar", "labels.new-label", "new-value"),
				),
			},
			{
				ResourceName:            "yandex_organizationmanager_idp_application_oauth_application.foobar",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func testAccCheckIdpApplicationOauthApplicationDestroy(s *terraform.State) error {
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

func testAccCheckIdpApplicationOauthApplicationExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

		found, err := oauthsdk.NewApplicationClient(config.SDKv2).Get(context.Background(), &oauth.GetApplicationRequest{
			ApplicationId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("IdpApplicationOauthApplication %s not found", n)
		}

		return nil
	}
}

func testAccIdpApplicationOauthApplication_full(appName, organizationID, clientName, folderID string) string {
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

func testAccIdpApplicationOauthApplication_update(appName, organizationID, clientName, folderID string) string {
	return fmt.Sprintf(`
resource "yandex_iam_oauth_client" "test_client" {
  name       = "%s"
  folder_id  = "%s"
  scopes     = ["iam"]
}

resource "yandex_organizationmanager_idp_application_oauth_application" "foobar" {
  organization_id = "%s"
  name            = "%s"
  description     = "Updated OAuth application description"

  client_grant = {
    client_id         = yandex_iam_oauth_client.test_client.id
    authorized_scopes = ["openid", "profile", "email"]
  }

  group_claims_settings = {
    group_distribution_type = "ALL_GROUPS"
  }

  labels = {
    env       = "production"
    app       = "updated-app"
    new-label = "new-value"
  }
}
`, clientName, folderID, organizationID, appName)
}

func testSweepIdpApplicationOauthApplication(_ string) error {
	conf, err := test.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	organizationID := test.GetExampleOrganizationID()
	if organizationID == "" {
		log.Printf("[WARN] organization ID is not set, skipping OAuth application sweep")
		return nil
	}

	req := &oauth.ListApplicationsRequest{
		OrganizationId: organizationID,
		PageSize:       applicationSweepPageSize,
	}

	client := oauthsdk.NewApplicationClient(conf.SDKv2)
	resp, err := client.List(context.Background(), req)
	if err != nil {
		return fmt.Errorf("error getting list of OAuth applications: %s", err)
	}

	result := &multierror.Error{}
	for _, app := range resp.Applications {
		if strings.HasPrefix(app.Name, testResourceNamePrefix) {
			if !sweepIdpApplicationOauthApplication(conf, app.Id) {
				result = multierror.Append(result, fmt.Errorf("failed to sweep Idp Application OAuth Application %q", app.Id))
			}
		}
	}

	// Handle pagination if needed
	for resp.NextPageToken != "" {
		req.PageToken = resp.NextPageToken
		resp, err = client.List(context.Background(), req)
		if err != nil {
			result = multierror.Append(result, fmt.Errorf("error getting next page of OAuth applications: %s", err))
			break
		}

		for _, app := range resp.Applications {
			if strings.HasPrefix(app.Name, testResourceNamePrefix) {
				if !sweepIdpApplicationOauthApplication(conf, app.Id) {
					result = multierror.Append(result, fmt.Errorf("failed to sweep Idp Application OAuth Application %q", app.Id))
				}
			}
		}
	}

	return result.ErrorOrNil()
}

func sweepIdpApplicationOauthApplication(conf *provider_config.Config, id string) bool {
	return test.SweepWithRetry(sweepIdpApplicationOauthApplicationOnce, conf, "Idp Application OAuth Application", id)
}

func sweepIdpApplicationOauthApplicationOnce(conf *provider_config.Config, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), applicationSweepDeleteTimeout)
	defer cancel()

	client := oauthsdk.NewApplicationClient(conf.SDKv2)
	op, err := client.Delete(ctx, &oauth.DeleteApplicationRequest{
		ApplicationId: id,
	})
	if err != nil {
		return err
	}

	_, err = op.Wait(ctx)
	return err
}
