package sweep

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	oauth "github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1/idp/application/oauth"
	oauthsdk "github.com/yandex-cloud/go-sdk/services/organizationmanager/v1/idp/application/oauth"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/internal/provider/framework/config"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

const (
	applicationSweepPageSize                             = 1000
	applicationSweepDeleteTimeout                        = 15 * time.Minute
	testResourceNamePrefixIdpApplicationOauthApplication = "tf-acc-test-oauth-app"
	oauthClientFolderID                                  = "b1g6bop87aoiekbkko82"
)

func registerYandexOrganizationmanagerIdpApplicationOauthApplicationResource(add func(string, *resource.Sweeper)) {
	add("yandex_organizationmanager_idp_application_oauth_application", &resource.Sweeper{
		Name:         "yandex_organizationmanager_idp_application_oauth_application",
		F:            sweepAllIdpApplicationOauthApplication,
		Dependencies: []string{},
	})
}

func sweepAllIdpApplicationOauthApplication(_ string) error {
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
		if strings.HasPrefix(app.Name, testResourceNamePrefixIdpApplicationOauthApplication) {
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
			if strings.HasPrefix(app.Name, testResourceNamePrefixIdpApplicationOauthApplication) {
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
