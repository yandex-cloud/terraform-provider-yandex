package acceptance

import (
	"fmt"
	"testing"

	resourcemanagerfixtures "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/resourcemanager/testfixtures"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
)

func TestAccSmartwebsecurityWafProfile_basic(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	name := acctest.RandomWithPrefix("tf-yc-wafp")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
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

func TestAccSmartwebsecurityWafProfile_defaultCaseSensitivity(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	name := acctest.RandomWithPrefix("tf-yc-waf-default-case")
	config := testAccSmartwebsecurityWafProfileDefaultCaseSensitivity(name)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_sws_waf_profile.this", "name", name),
					resource.TestCheckResourceAttr("yandex_sws_waf_profile.this", "exclusion_rule.0.request_condition.0.header_matcher.0.header_name.0.case_sensitive", "false"),
					resource.TestCheckResourceAttr("yandex_sws_waf_profile.this", "exclusion_rule.0.request_condition.0.param_matcher.0.param_name.0.case_sensitive", "false"),
					resource.TestCheckResourceAttr("yandex_sws_waf_profile.this", "exclusion_rule.0.request_condition.0.cookie_matcher.0.cookie_name.0.case_sensitive", "false"),
				),
			},
			{
				ResourceName:      "yandex_sws_waf_profile.this",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccSmartwebsecurityWAF_UpgradeFromSDKv2(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	name := acctest.RandomWithPrefix("tf-yc-sc")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		CheckDestroy: resourcemanagerfixtures.CheckFolderDestroy(suite.Config),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"yandex": {
						VersionConstraint: "0.200.0",
						Source:            "yandex-cloud/yandex",
					},
				},
				Config: testAccSmartwebsecurityWAFBasicMigration(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_sws_waf_profile.this", "name", name),
				),
			},
			{
				ProtoV6ProviderFactories: suite.Protocol6Factories(),
				Config:                   testAccSmartwebsecurityWAFBasicMigration(name),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func testAccSmartwebsecurityWAFBasicMigration(targetName string) string {
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
			type = "CORE"
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
	rule_set {
		action     = "DENY"
    	is_enabled = true
    	priority   = 1
		core_rule_set {
			inbound_anomaly_score = 2
			paranoia_level = local.waf_paranoia_level
			rule_set {
				name = "OWASP Core Ruleset"
				version = "4.0.0"
				type = "CORE"
				id = "OWASP_CRS_4_0_0"
			}
		}
	}
}
`, targetName)
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
			type = "CORE"
        }
    }
    analyze_request_body {
        is_enabled = true
        size_limit = 8
        size_limit_action = "IGNORE"
    }

	rule_set {
		action     = "DENY"
    	is_enabled = true
    	priority   = 1
		core_rule_set {
        	inbound_anomaly_score = 2
        	paranoia_level = 4
        	rule_set {
            	name = "OWASP Core Ruleset"
            	version = "4.0.0"
				type = "CORE"
				id="OWASP_CRS_4_0_0"
        	}
    	}
	}
}
`, targetName)
}

func testAccSmartwebsecurityWafProfileDefaultCaseSensitivity(targetName string) string {
	return fmt.Sprintf(`
resource "yandex_sws_waf_profile" "this" {
  name = %[1]q

  rule_set {
    action     = "DENY"
    is_enabled = true
    priority   = 1

    core_rule_set {
      inbound_anomaly_score = 2
      paranoia_level        = 4

      rule_set {
        id      = "OWASP_CRS_4_0_0"
        name    = "OWASP Core Ruleset"
        version = "4.0.0"
        type    = "CORE"
      }
    }
  }

  exclusion_rule {
    name = "case-insensitive-header"

    exclude_rules {
      exclude_all = true
    }

    request_condition {
      param_matcher {
        param_name {
          value = "query"
        }
      }

      header_matcher {
        header_name {
          value = "X-Test-Header"
        }
      }

      cookie_matcher {
        cookie_name {
          value = "session"
        }
      }
    }
  }
}
`, targetName)
}
