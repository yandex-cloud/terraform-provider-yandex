package yandex_organizationmanager_idp_application_saml_application_test

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
	saml "github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1/idp/application/saml"
	samlsdk "github.com/yandex-cloud/go-sdk/services/organizationmanager/v1/idp/application/saml"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	samlAppSweepPageSize      = 1000
	samlAppSweepDeleteTimeout = 15 * time.Minute
	testResourceNamePrefix    = "tf-acc-test-saml-app"
)

func init() {
	resource.AddTestSweepers("yandex_organizationmanager_idp_application_saml_application", &resource.Sweeper{
		Name:         "yandex_organizationmanager_idp_application_saml_application",
		F:            testSweepIdpSamlApplication,
		Dependencies: []string{},
	})
}

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestAccOrganizationManagerIdpApplicationSamlApplication_basic(t *testing.T) {
	appName := acctest.RandomWithPrefix("tf-acc-test-saml-app")
	organizationID := test.GetExampleOrganizationID()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckIdpSamlApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIdpSamlApplication_basic(appName, organizationID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIdpSamlApplicationExists("yandex_organizationmanager_idp_application_saml_application.foobar"),
					test.AccCheckCreatedAtAttr("yandex_organizationmanager_idp_application_saml_application.foobar"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_application_saml_application.foobar", "name", appName),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_application_saml_application.foobar", "organization_id", organizationID),
					resource.TestCheckResourceAttrSet("yandex_organizationmanager_idp_application_saml_application.foobar", "application_id"),
					resource.TestCheckResourceAttrSet("yandex_organizationmanager_idp_application_saml_application.foobar", "id"),
				),
			},
		},
	})
}

func TestAccOrganizationManagerIdpApplicationSamlApplication_update(t *testing.T) {
	appName := acctest.RandomWithPrefix("tf-acc-test-saml-app")
	organizationID := test.GetExampleOrganizationID()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckIdpSamlApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIdpSamlApplication_full(appName, organizationID, "initial description", "test", "saml-initial", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIdpSamlApplicationExists("yandex_organizationmanager_idp_application_saml_application.foobar"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_application_saml_application.foobar", "name", appName),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_application_saml_application.foobar", "description", "initial description"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_application_saml_application.foobar", "service_provider.entity_id", "https://example-initial.com/saml/metadata"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_application_saml_application.foobar", "service_provider.acs_urls.0.url", "https://example-initial.com/saml/acs"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_application_saml_application.foobar", "service_provider.acs_urls.0.index", "0"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_application_saml_application.foobar", "service_provider.slo_urls.0.url", "https://example-initial.com/saml/slo"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_application_saml_application.foobar", "service_provider.slo_urls.0.protocol_binding", "HTTP_POST"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_application_saml_application.foobar", "attribute_mapping.name_id.format", "EMAIL"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_application_saml_application.foobar", "attribute_mapping.attributes.0.name", "email"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_application_saml_application.foobar", "attribute_mapping.attributes.0.value", "SubjectClaims.email"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_application_saml_application.foobar", "attribute_mapping.attributes.1.name", "firstName"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_application_saml_application.foobar", "attribute_mapping.attributes.1.value", "SubjectClaims.given_name"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_application_saml_application.foobar", "security_settings.signature_mode", "ASSERTIONS"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_application_saml_application.foobar", "group_claims_settings.group_attribute_name", "groups-initial"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_application_saml_application.foobar", "group_claims_settings.group_distribution_type", "ASSIGNED_GROUPS"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_application_saml_application.foobar", "labels.env", "test"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_application_saml_application.foobar", "labels.app", "saml-initial"),
				),
			},
			{
				Config: testAccIdpSamlApplication_full(appName+"-updated", organizationID, "updated description", "production", "saml-updated", `    version = "2"`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIdpSamlApplicationExists("yandex_organizationmanager_idp_application_saml_application.foobar"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_application_saml_application.foobar", "name", appName+"-updated"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_application_saml_application.foobar", "description", "updated description"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_application_saml_application.foobar", "labels.env", "production"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_application_saml_application.foobar", "labels.app", "saml-updated"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_application_saml_application.foobar", "labels.version", "2"),
				),
			},
			{
				ResourceName:            "yandex_organizationmanager_idp_application_saml_application.foobar",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}

func testAccCheckIdpSamlApplicationDestroy(s *terraform.State) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_organizationmanager_idp_application_saml_application" {
			continue
		}

		_, err := samlsdk.NewApplicationClient(config.SDKv2).Get(context.Background(), &saml.GetApplicationRequest{
			ApplicationId: rs.Primary.ID,
		})

		if err != nil {
			if grpcStatus, ok := status.FromError(err); ok && grpcStatus != nil && grpcStatus.Code() == codes.NotFound {
				return nil
			} else if ok {
				return fmt.Errorf("Error while requesting Yandex Cloud: grpc code error : %d, http message error: %s", grpcStatus.Code(), grpcStatus.Message())
			}
			return fmt.Errorf("IdpSamlApplication still exists")
		}
	}

	return nil
}

func testAccCheckIdpSamlApplicationExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

		found, err := samlsdk.NewApplicationClient(config.SDKv2).Get(context.Background(), &saml.GetApplicationRequest{
			ApplicationId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("IdpSamlApplication %s not found", n)
		}

		return nil
	}
}

func testAccIdpSamlApplication_basic(name, organizationID string) string {
	return fmt.Sprintf(`
resource "yandex_organizationmanager_idp_application_saml_application" "foobar" {
  name           = "%s"
  organization_id = "%s"
  
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
  }
}
`, name, organizationID)
}

func testAccIdpSamlApplication_full(name, organizationID, description, labelEnv, labelApp, labelsExtra string) string {
	return fmt.Sprintf(`
resource "yandex_organizationmanager_idp_application_saml_application" "foobar" {
  name           = "%s"
  organization_id = "%s"
  description    = "%s"
  
  service_provider = {
    entity_id = "https://example-initial.com/saml/metadata"
    
    acs_urls = [{
      url   = "https://example-initial.com/saml/acs"
      index = 0
    }]

    slo_urls = [{
      url              = "https://example-initial.com/saml/slo"
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
    }, {
      name  = "firstName"
      value = "SubjectClaims.given_name"
    }]
  }

  security_settings = {
    signature_mode = "ASSERTIONS"
  }

  group_claims_settings = {
    group_attribute_name   = "groups-initial"
    group_distribution_type = "ASSIGNED_GROUPS"
  }

  labels = {
    env = "%s"
    app = "%s"
%s
  }
}
`, name, organizationID, description, labelEnv, labelApp, labelsExtra)
}

func testSweepIdpSamlApplication(_ string) error {
	conf, err := test.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	organizationID := test.GetExampleOrganizationID()
	if organizationID == "" {
		log.Printf("[WARN] organization ID is not set, skipping saml application sweep")
		return nil
	}

	client := samlsdk.NewApplicationClient(conf.SDKv2)
	req := &saml.ListApplicationsRequest{
		OrganizationId: organizationID,
		PageSize:       samlAppSweepPageSize,
	}

	resp, err := client.List(context.Background(), req)
	if err != nil {
		return fmt.Errorf("error getting list of SAML applications: %s", err)
	}

	result := &multierror.Error{}

	// Sweep applications with test prefixes
	for _, app := range resp.Applications {
		if strings.HasPrefix(app.Name, testResourceNamePrefix) || strings.Contains(app.Name, "tf-acc-test-") {
			if !sweepIdpSamlApplication(conf, app.Id) {
				result = multierror.Append(result, fmt.Errorf("failed to sweep Idp SAML Application %q", app.Id))
			}
		}
	}

	// Handle pagination
	for resp.NextPageToken != "" {
		req.PageToken = resp.NextPageToken
		resp, err = client.List(context.Background(), req)
		if err != nil {
			result = multierror.Append(result, fmt.Errorf("error getting next page of SAML applications: %s", err))
			break
		}

		for _, app := range resp.Applications {
			if strings.HasPrefix(app.Name, testResourceNamePrefix) || strings.Contains(app.Name, "tf-acc-test-") {
				if !sweepIdpSamlApplication(conf, app.Id) {
					result = multierror.Append(result, fmt.Errorf("failed to sweep Idp SAML Application %q", app.Id))
				}
			}
		}
	}

	return result.ErrorOrNil()
}

func sweepIdpSamlApplication(conf *provider_config.Config, id string) bool {
	return test.SweepWithRetry(sweepIdpSamlApplicationOnce, conf, "Idp SAML Application", id)
}

func sweepIdpSamlApplicationOnce(conf *provider_config.Config, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), samlAppSweepDeleteTimeout)
	defer cancel()

	client := samlsdk.NewApplicationClient(conf.SDKv2)
	op, err := client.Delete(ctx, &saml.DeleteApplicationRequest{
		ApplicationId: id,
	})
	if err != nil {
		return err
	}

	_, err = op.Wait(ctx)
	return err
}
