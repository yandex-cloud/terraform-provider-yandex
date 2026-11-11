package sweep

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/smartwebsecurity/v1"
	smartwebsecuritysdk "github.com/yandex-cloud/go-sdk/services/smartwebsecurity/v1"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
)

func registerResourceYandexSwsSecurityProfile(add func(string, *resource.Sweeper)) {
	add("yandex_sws_security_profile", &resource.Sweeper{
		Name: "yandex_sws_security_profile",
		F:    sweepAllSecurityProfile,
		Dependencies: []string{
			"yandex_smartcaptcha_captcha",
		},
	})
}

func sweepAllSecurityProfile(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	client := smartwebsecuritysdk.NewSecurityProfileClient(conf.SDK)
	resp, err := client.List(conf.Context(), &smartwebsecurity.ListSecurityProfilesRequest{
		FolderId: conf.FolderID,
	})
	if err != nil {
		return fmt.Errorf("error getting SmartWebSecurity security profiles: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.SecurityProfiles {
		if !sweepSecurityProfile(conf, c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep SmartWebSecurity security profile %q", c.Id))
		}
	}

	return result.ErrorOrNil()
}

func sweepSecurityProfile(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepSecurityProfileOnce, conf, "SecurityProfile", id)
}

func sweepSecurityProfileOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(time.Minute)
	defer cancel()

	client := smartwebsecuritysdk.NewSecurityProfileClient(conf.SDK)
	op, err := client.Delete(ctx, &smartwebsecurity.DeleteSecurityProfileRequest{
		SecurityProfileId: id,
	})
	return sweep.HandleSweepOperationV2(ctx, op, err)
}
