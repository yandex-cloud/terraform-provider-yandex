package yandex_organizationmanager_idp_application_saml_signature_certificate_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	saml "github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1/idp/application/saml"
	samlsdk "github.com/yandex-cloud/go-sdk/services/organizationmanager/v1/idp/application/saml"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/internal/provider/framework"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	signatureCertificateSweepPageSize      = 1000
	signatureCertificateSweepDeleteTimeout = 15 * time.Minute
	signatureCertificateNamePrefix         = "tf-acc-test-saml-sig-cert"
)

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
