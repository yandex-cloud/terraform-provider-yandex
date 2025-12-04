package yandex_organizationmanager_idp_application_saml_signature_certificate_test

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
	signatureCertificateSweepPageSize      = 1000
	signatureCertificateSweepDeleteTimeout = 15 * time.Minute
	signatureCertificateNamePrefix         = "tf-acc-test-saml-sig-cert"
)

func init() {
	resource.AddTestSweepers("yandex_organizationmanager_idp_application_saml_signature_certificate", &resource.Sweeper{
		Name:         "yandex_organizationmanager_idp_application_saml_signature_certificate",
		F:            testSweepIdpSamlSignatureCertificate,
		Dependencies: []string{},
	})
}

// TestMain - add sweepers flag to the go test command so we can run sweepers.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestAccOrganizationManagerIdpApplicationSamlSignatureCertificate_basic(t *testing.T) {
	appName := acctest.RandomWithPrefix("tf-acc-test-saml-app")
	certName := acctest.RandomWithPrefix(signatureCertificateNamePrefix)
	organizationID := test.GetExampleOrganizationID()
	resourceName := "yandex_organizationmanager_idp_application_saml_signature_certificate.foobar"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckIdpSamlSignatureCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIdpSamlSignatureCertificateConfig(appName, certName, "initial description", organizationID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIdpSamlSignatureCertificateExists(resourceName, certName, "initial description"),
					test.AccCheckCreatedAtAttr(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", certName),
					resource.TestCheckResourceAttr(resourceName, "description", "initial description"),
					resource.TestCheckResourceAttr(resourceName, "status", saml.SignatureCertificate_INACTIVE.String()),
					resource.TestCheckResourceAttrSet(resourceName, "signature_certificate_id"),
					resource.TestCheckResourceAttrSet(resourceName, "fingerprint"),
					resource.TestCheckResourceAttrSet(resourceName, "application_id"),
				),
			},
		},
	})
}

func TestAccOrganizationManagerIdpApplicationSamlSignatureCertificate_update(t *testing.T) {
	appName := acctest.RandomWithPrefix("tf-acc-test-saml-app")
	initialName := acctest.RandomWithPrefix(signatureCertificateNamePrefix)
	updatedName := acctest.RandomWithPrefix(signatureCertificateNamePrefix)
	organizationID := test.GetExampleOrganizationID()
	resourceName := "yandex_organizationmanager_idp_application_saml_signature_certificate.foobar"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckIdpSamlSignatureCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIdpSamlSignatureCertificateConfig(appName, initialName, "initial description", organizationID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIdpSamlSignatureCertificateExists(resourceName, initialName, "initial description"),
					resource.TestCheckResourceAttr(resourceName, "name", initialName),
					resource.TestCheckResourceAttr(resourceName, "description", "initial description"),
				),
			},
			{
				Config: testAccIdpSamlSignatureCertificateConfig(appName, updatedName, "updated description", organizationID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIdpSamlSignatureCertificateExists(resourceName, updatedName, "updated description"),
					resource.TestCheckResourceAttr(resourceName, "name", updatedName),
					resource.TestCheckResourceAttr(resourceName, "description", "updated description"),
					resource.TestCheckResourceAttr(resourceName, "status", saml.SignatureCertificate_INACTIVE.String()),
					resource.TestCheckResourceAttrSet(resourceName, "fingerprint"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"data"},
			},
		},
	})
}

func testAccCheckIdpSamlSignatureCertificateDestroy(s *terraform.State) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_organizationmanager_idp_application_saml_signature_certificate" {
			continue
		}

		_, err := samlsdk.NewSignatureCertificateClient(config.SDKv2).Get(context.Background(), &saml.GetSignatureCertificateRequest{
			SignatureCertificateId: rs.Primary.ID,
		})
		if err != nil {
			if grpcStatus, ok := status.FromError(err); ok && grpcStatus != nil && grpcStatus.Code() == codes.NotFound {
				continue
			} else if ok {
				return fmt.Errorf("Error while requesting Yandex Cloud: grpc code error : %d, http message error: %s", grpcStatus.Code(), grpcStatus.Message())
			}
			return fmt.Errorf("Idp SAML signature certificate still exists")
		}
		return fmt.Errorf("Idp SAML signature certificate still exists")
	}

	return nil
}

func testAccCheckIdpSamlSignatureCertificateExists(resourceName, expectedName, expectedDescription string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set for %s", resourceName)
		}

		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

		found, err := samlsdk.NewSignatureCertificateClient(config.SDKv2).Get(context.Background(), &saml.GetSignatureCertificateRequest{
			SignatureCertificateId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.GetName() != expectedName {
			return fmt.Errorf("expected name %q, got %q", expectedName, found.GetName())
		}

		if found.GetDescription() != expectedDescription {
			return fmt.Errorf("expected description %q, got %q", expectedDescription, found.GetDescription())
		}

		if expectedApplicationID, ok := rs.Primary.Attributes["application_id"]; ok && expectedApplicationID != "" {
			if found.GetApplicationId() != expectedApplicationID {
				return fmt.Errorf("expected application_id %q, got %q", expectedApplicationID, found.GetApplicationId())
			}
		}

		return nil
	}
}

func testAccIdpSamlSignatureCertificateConfig(appName, certificateName, description, organizationID string) string {
	return fmt.Sprintf(`
resource "yandex_organizationmanager_idp_application_saml_application" "foobar" {
  name            = "%s"
  organization_id = "%s"

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

  labels = {
    label = "label-value"
  }

  lifecycle {
    ignore_changes = [
      identity_provider_metadata,
      status,
      created_at,
      updated_at,
    ]
  }
}

resource "yandex_organizationmanager_idp_application_saml_signature_certificate" "foobar" {
  application_id = yandex_organizationmanager_idp_application_saml_application.foobar.application_id
  name           = "%s"
  description    = "%s"
}
`, appName, organizationID, certificateName, description)
}

func testSweepIdpSamlSignatureCertificate(_ string) error {
	conf, err := test.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting config for sweepers: %w", err)
	}

	organizationID := test.GetExampleOrganizationID()
	if organizationID == "" {
		log.Printf("[WARN] organization ID is not set, skipping signature certificate sweep")
		return nil
	}

	appClient := samlsdk.NewApplicationClient(conf.SDKv2)
	appReq := &saml.ListApplicationsRequest{
		OrganizationId: organizationID,
		PageSize:       signatureCertificateSweepPageSize,
	}

	result := &multierror.Error{}
	for {
		resp, err := appClient.List(context.Background(), appReq)
		if err != nil {
			return fmt.Errorf("error listing SAML applications: %w", err)
		}

		for _, app := range resp.GetApplications() {
			if err := sweepCertificatesForApplication(conf, app.GetId()); err != nil {
				result = multierror.Append(result, err)
			}
		}

		if resp.GetNextPageToken() == "" {
			break
		}
		appReq.PageToken = resp.GetNextPageToken()
	}

	return result.ErrorOrNil()
}

func sweepCertificatesForApplication(conf *provider_config.Config, applicationID string) error {
	client := samlsdk.NewSignatureCertificateClient(conf.SDKv2)
	req := &saml.ListSignatureCertificatesRequest{
		ApplicationId: applicationID,
		PageSize:      signatureCertificateSweepPageSize,
	}

	result := &multierror.Error{}
	for {
		resp, err := client.List(context.Background(), req)
		if err != nil {
			return fmt.Errorf("error listing signature certificates: %w", err)
		}

		for _, cert := range resp.GetSignatureCertificates() {
			if strings.HasPrefix(cert.GetName(), signatureCertificateNamePrefix) || strings.Contains(cert.GetName(), "tf-acc-test-") {
				if !test.SweepWithRetry(sweepIdpSamlSignatureCertificateOnce, conf, "Idp SAML signature certificate", cert.GetId()) {
					result = multierror.Append(result, fmt.Errorf("failed to sweep signature certificate %q", cert.GetId()))
				}
			}
		}

		if resp.GetNextPageToken() == "" {
			break
		}
		req.PageToken = resp.GetNextPageToken()
	}

	return result.ErrorOrNil()
}

func sweepIdpSamlSignatureCertificateOnce(conf *provider_config.Config, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), signatureCertificateSweepDeleteTimeout)
	defer cancel()

	client := samlsdk.NewSignatureCertificateClient(conf.SDKv2)
	op, err := client.Delete(ctx, &saml.DeleteSignatureCertificateRequest{
		SignatureCertificateId: id,
	})
	if err != nil {
		return err
	}

	_, err = op.Wait(ctx)
	return err
}
