package sweep

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	saml "github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1/idp/application/saml"
	samlsdk "github.com/yandex-cloud/go-sdk/services/organizationmanager/v1/idp/application/saml"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/internal/provider/framework/config"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

const (
	signatureCertificateSweepPageSize      = 1000
	signatureCertificateSweepDeleteTimeout = 15 * time.Minute
	signatureCertificateNamePrefix         = "tf-acc-test-saml-sig-cert"
)

func registerYandexOrganizationmanagerIdpApplicationSamlSignatureCertificateResource(add func(string, *resource.Sweeper)) {
	add("yandex_organizationmanager_idp_application_saml_signature_certificate", &resource.Sweeper{
		Name:         "yandex_organizationmanager_idp_application_saml_signature_certificate",
		F:            sweepAllIdpSamlSignatureCertificate,
		Dependencies: []string{},
	})
}

func sweepAllIdpSamlSignatureCertificate(_ string) error {
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
