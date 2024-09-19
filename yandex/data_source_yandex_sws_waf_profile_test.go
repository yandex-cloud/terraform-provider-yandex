package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceSmartwebsecurityWafProfile_byID(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-yc-wafp")
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceWafProfileConfig(name, true),
				Check:  testAccDataSourceWafProfileCheck(name),
			},
		},
	})
}

func TestAccDataSourceSmartwebsecurityWafProfile_byName(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-yc-wafp")
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceWafProfileConfig(name, false),
				Check:  testAccDataSourceWafProfileCheck(name),
			},
		},
	})
}

func testAccDataSourceWafProfileCheck(name string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr("data.yandex_sws_waf_profile.res", "name", name),
		resource.TestCheckResourceAttr("data.yandex_sws_waf_profile.res", "analyze_request_body.0.size_limit_action", "IGNORE"),
		resource.TestCheckResourceAttr("data.yandex_sws_waf_profile.res", "core_rule_set.0.paranoia_level", "4"),
		resource.TestCheckResourceAttr("data.yandex_sws_waf_profile.res", "rule.0.is_enabled", "true"),
		resource.TestCheckResourceAttr("data.yandex_sws_waf_profile.res", "rule.1.is_enabled", "true"),
		resource.TestCheckResourceAttr("data.yandex_sws_waf_profile.res", "rule.2.is_enabled", "true"),
		resource.TestCheckResourceAttr("data.yandex_sws_waf_profile.res", "rule.3.is_enabled", "true"),
	)
}

func testAccWafProfileResourceConfig(name string) string {
	return fmt.Sprintf(`
locals {
  waf_paranoia_level = 4
}
data "yandex_sws_waf_rule_set_descriptor" "owasp4" {
  name = "OWASP Core Ruleset"
  version = "4.0.0"
}
resource "yandex_sws_waf_profile" "this" {
	name = "%[1]v"
    core_rule_set {
        inbound_anomaly_score = 2
        paranoia_level = local.waf_paranoia_level
        rule_set {
            name = "OWASP Core Ruleset"
            version = "4.0.0"
        }
    }
    dynamic "rule" {
        for_each = [
            for rule in data.yandex_sws_waf_rule_set_descriptor.owasp4.rules: rule
            if rule.paranoia_level >= local.waf_paranoia_level
        ]
        content {
            rule_id = rule.value.id
            is_enabled = true
            is_blocking = false
        }
    }
    analyze_request_body {
        is_enabled = true
        size_limit = 8
        size_limit_action = "IGNORE"
    }
}
`, name)
}

const wafProfileDataByIDConfig = `
data "yandex_sws_waf_profile" "res" {
	waf_profile_id = yandex_sws_waf_profile.this.id
}
`

const wafProfileDataByNameConfig = `
data "yandex_sws_waf_profile" "res" {
	name = yandex_sws_waf_profile.this.name
}
`

func testAccDataSourceWafProfileConfig(name string, withDataID bool) string {
	if withDataID {
		return testAccWafProfileResourceConfig(name) + wafProfileDataByIDConfig
	} else {
		return testAccWafProfileResourceConfig(name) + wafProfileDataByNameConfig
	}
}
