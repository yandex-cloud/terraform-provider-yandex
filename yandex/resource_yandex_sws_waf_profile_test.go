package yandex

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	waf "github.com/yandex-cloud/go-genproto/yandex/cloud/smartwebsecurity/v1/waf"
)

func init() {
	resource.AddTestSweepers("yandex_sws_waf_profile", &resource.Sweeper{
		Name:         "yandex_sws_waf_profile",
		F:            testSweepWafProfile,
		Dependencies: []string{},
	})
}

func TestAccSmartwebsecurityWafProfile_basic(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-yc-wafp")
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSmartwebsecurityWafProfileBasic(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_sws_waf_profile.this", "name", name),
					resource.TestCheckResourceAttr("yandex_sws_waf_profile.this", "analyze_request_body.0.size_limit_action", "IGNORE"),
					resource.TestCheckResourceAttr("yandex_sws_waf_profile.this", "core_rule_set.0.paranoia_level", "4"),
				),
			},
			{
				ResourceName:      "yandex_sws_waf_profile.this",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccSmartwebsecurityWafProfileBasic(targetName string) string {
	return fmt.Sprintf(`
resource "yandex_sws_waf_profile" "this" {	
	name = "%[1]v"
    core_rule_set {
        inbound_anomaly_score = 2
        paranoia_level = 4
        rule_set {
            name = "OWASP Core Ruleset"
            version = "4.0.0"
        }
    }
    analyze_request_body {
        is_enabled = true
        size_limit = 8
        size_limit_action = "IGNORE"
    }
}
`, targetName)
}

func testSweepWafProfile(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := conf.sdk.SmartWebSecurityWaf().WafProfile().List(conf.Context(), &waf.ListWafProfilesRequest{
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

func sweepWafProfile(conf *Config, id string) bool {
	return sweepWithRetry(sweepWafProfileOnce, conf, "WafProfile", id)
}

func sweepWafProfileOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(time.Minute)
	defer cancel()

	op, err := conf.sdk.SmartWebSecurityWaf().WafProfile().Delete(ctx, &waf.DeleteWafProfileRequest{
		WafProfileId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}
