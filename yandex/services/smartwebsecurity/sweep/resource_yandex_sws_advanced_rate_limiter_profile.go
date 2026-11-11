package sweep

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	advanced_rate_limiter "github.com/yandex-cloud/go-genproto/yandex/cloud/smartwebsecurity/v1/advanced_rate_limiter"
	arlsdk "github.com/yandex-cloud/go-sdk/services/smartwebsecurity/v1/advanced_rate_limiter"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
)

func registerResourceYandexSwsAdvancedRateLimiterProfile(add func(string, *resource.Sweeper)) {
	add("yandex_sws_advanced_rate_limiter_profile", &resource.Sweeper{
		Name:         "yandex_sws_advanced_rate_limiter_profile",
		F:            sweepAllArlProfile,
		Dependencies: []string{},
	})
}

func sweepAllArlProfile(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	client := arlsdk.NewAdvancedRateLimiterProfileClient(conf.SDK)
	resp, err := client.List(conf.Context(), &advanced_rate_limiter.ListAdvancedRateLimiterProfilesRequest{
		FolderId: conf.FolderID,
	})
	if err != nil {
		return fmt.Errorf("error getting SmartWebSecurity ARL profiles: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.AdvancedRateLimiterProfiles {
		if !sweepArlProfile(conf, c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep SmartWebSecurity ARL profile %q", c.Id))
		}
	}

	return result.ErrorOrNil()
}

func sweepArlProfile(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepArlProfileOnce, conf, "ArlProfile", id)
}

func sweepArlProfileOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(time.Minute)
	defer cancel()

	client := arlsdk.NewAdvancedRateLimiterProfileClient(conf.SDK)
	op, err := client.Delete(ctx, &advanced_rate_limiter.DeleteAdvancedRateLimiterProfileRequest{
		AdvancedRateLimiterProfileId: id,
	})
	return sweep.HandleSweepOperationV2(ctx, op, err)
}
