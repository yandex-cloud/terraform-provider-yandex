package sweep

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	organizationmanager "github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1"
	organizationmanagersdk "github.com/yandex-cloud/go-sdk/services/organizationmanager/v1"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/internal/provider/framework/config"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

const (
	sweepPageSize                        = 1000
	sweepDeleteTimeout                   = 15 * time.Minute
	testResourceNamePrefixMfaEnforcement = "tf-acc-test-mfa-enforcement"
)

func registerYandexOrganizationmanagerMfaEnforcementResource(add func(string, *resource.Sweeper)) {
	add("yandex_organizationmanager_mfa_enforcement", &resource.Sweeper{
		Name:         "yandex_organizationmanager_mfa_enforcement",
		F:            sweepAllMfaEnforcement,
		Dependencies: []string{},
	})
}

func sweepAllMfaEnforcement(_ string) error {
	conf, err := test.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	organizationId := test.GetExampleOrganizationID()
	if organizationId == "" {
		log.Printf("[WARN] organization id is not set, skipping mfa enforcement sweep")
		return nil
	}

	req := &organizationmanager.ListMfaEnforcementsRequest{
		OrganizationId: organizationId,
		PageSize:       sweepPageSize,
	}

	client := organizationmanagersdk.NewMfaEnforcementClient(conf.SDKv2)
	resp, err := client.List(context.Background(), req)
	if err != nil {
		return fmt.Errorf("error getting list of mfa enforcements: %s", err)
	}

	result := &multierror.Error{}
	for _, mfaEnforcement := range resp.MfaEnforcements {
		if strings.HasPrefix(mfaEnforcement.Name, testResourceNamePrefixMfaEnforcement) {
			if !sweepMfaEnforcement(conf, mfaEnforcement.Id) {
				result = multierror.Append(result, fmt.Errorf("failed to sweep mfa enforcement %q", mfaEnforcement.Id))
			}
		}
	}

	// Handle pagination if needed
	for resp.NextPageToken != "" {
		req.PageToken = resp.NextPageToken
		resp, err = client.List(context.Background(), req)
		if err != nil {
			result = multierror.Append(result, fmt.Errorf("error getting next page of mfa enforcements: %s", err))
			break
		}

		for _, mfaEnforcement := range resp.MfaEnforcements {
			if strings.HasPrefix(mfaEnforcement.Name, testResourceNamePrefixMfaEnforcement) {
				if !sweepMfaEnforcement(conf, mfaEnforcement.Id) {
					result = multierror.Append(result, fmt.Errorf("failed to sweep mfa enforcement %q", mfaEnforcement.Id))
				}
			}
		}
	}

	return result.ErrorOrNil()
}

func sweepMfaEnforcement(conf *provider_config.Config, id string) bool {
	return test.SweepWithRetry(sweepMfaEnforcementOnce, conf, "Mfa Enforcement", id)
}

func sweepMfaEnforcementOnce(conf *provider_config.Config, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), sweepDeleteTimeout)
	defer cancel()

	client := organizationmanagersdk.NewMfaEnforcementClient(conf.SDKv2)
	op, err := client.Delete(ctx, &organizationmanager.DeleteMfaEnforcementRequest{
		MfaEnforcementId: id,
	})
	if err != nil {
		return err
	}

	_, err = op.Wait(ctx)
	return err
}
