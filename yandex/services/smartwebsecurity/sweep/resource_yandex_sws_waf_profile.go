package sweep

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	waf "github.com/yandex-cloud/go-genproto/yandex/cloud/smartwebsecurity/v1/waf"
	wafsdk "github.com/yandex-cloud/go-sdk/services/smartwebsecurity/v1/waf"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
)

func registerResourceYandexSwsWafProfile(add func(string, *resource.Sweeper)) {
	add("yandex_sws_waf_profile", &resource.Sweeper{
		Name:         "yandex_sws_waf_profile",
		F:            sweepAllWafProfile,
		Dependencies: []string{},
	})
}

func sweepAllWafProfile(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	client := wafsdk.NewWafProfileClient(conf.SDK)
	resp, err := client.List(conf.Context(), &waf.ListWafProfilesRequest{
		FolderId: conf.FolderID,
	})
	if err != nil {
		return fmt.Errorf("error getting SmartWebSecurity WAF profiles: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.WafProfiles {
		if !sweepWafProfile(conf, c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep SmartWebSecurity WAF profile %q", c.Id))
		}
	}

	return result.ErrorOrNil()
}

func sweepWafProfile(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepWafProfileOnce, conf, "WafProfile", id)
}

func sweepWafProfileOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(time.Minute)
	defer cancel()

	client := wafsdk.NewWafProfileClient(conf.SDK)
	op, err := client.Delete(ctx, &waf.DeleteWafProfileRequest{
		WafProfileId: id,
	})
	return sweep.HandleSweepOperationV2(ctx, op, err)
}
