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
	samlAppSweepPageSize                                = 1000
	samlAppSweepDeleteTimeout                           = 15 * time.Minute
	testResourceNamePrefixIdpApplicationSamlApplication = "tf-acc-test-saml-app"
)

func registerYandexOrganizationmanagerIdpApplicationSamlApplicationResource(add func(string, *resource.Sweeper)) {
	add("yandex_organizationmanager_idp_application_saml_application", &resource.Sweeper{
		Name:         "yandex_organizationmanager_idp_application_saml_application",
		F:            sweepAllIdpSamlApplication,
		Dependencies: []string{},
	})
}

func sweepAllIdpSamlApplication(_ string) error {
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
		if strings.HasPrefix(app.Name, testResourceNamePrefixIdpApplicationSamlApplication) || strings.Contains(app.Name, "tf-acc-test-") {
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
			if strings.HasPrefix(app.Name, testResourceNamePrefixIdpApplicationSamlApplication) || strings.Contains(app.Name, "tf-acc-test-") {
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
