package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceSmartwebsecurityArlProfile_byID(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-yc-arlp")
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceArlProfileConfig(name, true),
				Check:  testAccDataSourceArlProfileCheck(name),
			},
		},
	})
}

func TestAccDataSourceSmartwebsecurityArlProfile_byName(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-yc-arlp")
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceArlProfileConfig(name, false),
				Check:  testAccDataSourceArlProfileCheck(name),
			},
		},
	})
}

func testAccDataSourceArlProfileCheck(name string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr("data.yandex_sws_advanced_rate_limiter_profile.res", "name", name),
		resource.TestCheckResourceAttr("data.yandex_sws_advanced_rate_limiter_profile.res", "advanced_rate_limiter_rule.0.priority", "10"),
		resource.TestCheckResourceAttr("data.yandex_sws_advanced_rate_limiter_profile.res", "advanced_rate_limiter_rule.0.static_quota.0.action", "DENY"),
	)
}

func testAccArlProfileResourceConfig(name string) string {
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
`, name)
}

const arlProfileDataByIDConfig = `
data "yandex_sws_advanced_rate_limiter_profile" "res" {
	advanced_rate_limiter_profile_id = yandex_sws_advanced_rate_limiter_profile.this.id
}
`

const arlProfileDataByNameConfig = `
data "yandex_sws_advanced_rate_limiter_profile" "res" {
	name = yandex_sws_advanced_rate_limiter_profile.this.name
}
`

func testAccDataSourceArlProfileConfig(name string, withDataID bool) string {
	if withDataID {
		return testAccArlProfileResourceConfig(name) + arlProfileDataByIDConfig
	} else {
		return testAccArlProfileResourceConfig(name) + arlProfileDataByNameConfig
	}
}
