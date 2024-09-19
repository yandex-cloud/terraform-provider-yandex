package yandex

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	advanced_rate_limiter "github.com/yandex-cloud/go-genproto/yandex/cloud/smartwebsecurity/v1/advanced_rate_limiter"
)

func init() {
	resource.AddTestSweepers("yandex_sws_advanced_rate_limiter_profile", &resource.Sweeper{
		Name:         "yandex_sws_advanced_rate_limiter_profile",
		F:            testSweepArlProfile,
		Dependencies: []string{},
	})
}

func TestAccSmartwebsecurityArlProfile_basic(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-yc-arlp")
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSmartwebsecurityArlProfileBasic(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_sws_advanced_rate_limiter_profile.this", "name", name),
					resource.TestCheckResourceAttr("yandex_sws_advanced_rate_limiter_profile.this", "advanced_rate_limiter_rule.0.priority", "10"),
					resource.TestCheckResourceAttr("yandex_sws_advanced_rate_limiter_profile.this", "advanced_rate_limiter_rule.0.static_quota.0.action", "DENY"),
				),
			},
			{
				ResourceName:      "yandex_sws_advanced_rate_limiter_profile.this",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccSmartwebsecurityArlProfileBasic(targetName string) string {
	return fmt.Sprintf(`
resource "yandex_sws_advanced_rate_limiter_profile" "this" {	
	name = "%[1]v"
	advanced_rate_limiter_rule {
        name = "rule1"
        priority = 10
        description = "First test rule"
        dry_run = true
        static_quota {
            action = "DENY"
            limit = 10000000
            period = 1
            condition {
                request_uri {
                  path {
                      exact_match = "/api"
                  }
                }
            }
        }
    }
}
`, targetName)
}

func testSweepArlProfile(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := conf.sdk.SmartWebSecurityArl().AdvancedRateLimiterProfile().List(conf.Context(), &advanced_rate_limiter.ListAdvancedRateLimiterProfilesRequest{
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

func sweepArlProfile(conf *Config, id string) bool {
	return sweepWithRetry(sweepArlProfileOnce, conf, "ArlProfile", id)
}

func sweepArlProfileOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(time.Minute)
	defer cancel()

	op, err := conf.sdk.SmartWebSecurityArl().AdvancedRateLimiterProfile().Delete(ctx, &advanced_rate_limiter.DeleteAdvancedRateLimiterProfileRequest{
		AdvancedRateLimiterProfileId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}
