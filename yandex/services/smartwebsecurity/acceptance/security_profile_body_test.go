package acceptance

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
)

func TestAccSmartwebsecuritySecurityProfile_bodyRoundTrip(t *testing.T) {
	for _, tc := range []struct{ name, field, value string }{
		{"regex", "pire_regex_match", ".*php2shell.*"},
		{"exact", "exact_match", "php2shell"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			suite := testSuite{acceptance.New(t)}
			name := acctest.RandomWithPrefix("tf-yc-sws-body")
			matcherPath := "security_rule.0.rule_condition.0.condition.0.body.0.body_matcher.0." + tc.field
			check := resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("yandex_sws_security_profile.body", matcherPath, tc.value),
				resource.TestCheckResourceAttrPair("data.yandex_sws_security_profile.body", matcherPath, "yandex_sws_security_profile.body", matcherPath),
			)
			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { acceptance.PreCheck(t) },
				ProtoV6ProviderFactories: suite.Protocol6Factories(),
				Steps: []resource.TestStep{
					{
						Config: testAccSecurityProfileBody(name, 352, tc.field, tc.value),
						Check:  check,
					},
					{
						Config: testAccSecurityProfileBody(name, 353, tc.field, tc.value),
						Check: resource.ComposeTestCheckFunc(check,
							resource.TestCheckResourceAttr("yandex_sws_security_profile.body", "security_rule.0.priority", "353")),
					},
					{
						ResourceName:      "yandex_sws_security_profile.body",
						ImportState:       true,
						ImportStateVerify: true,
					},
					{
						Config: testAccSecurityProfileBody(name, 353, tc.field, tc.value),
						ConfigPlanChecks: resource.ConfigPlanChecks{
							PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
						},
						Check: check,
					},
				},
			})
		})
	}
}

func testAccSecurityProfileBody(name string, priority int, field, value string) string {
	return fmt.Sprintf(`
resource "yandex_sws_security_profile" "body" {
  name           = "%[1]s"
  default_action = "ALLOW"
  analyze_request_body {
    size_limit        = 8
    size_limit_action = "IGNORE"
  }
  security_rule {
    name     = "block-php2shell"
    priority = %[2]d
    rule_condition {
      action = "DENY"
      condition {
        http_method {
          http_method_matcher {
            exact_match = "POST"
          }
        }
        body {
          body_matcher {
            %[3]s = %[4]q
          }
        }
      }
    }
  }
}

data "yandex_sws_security_profile" "body" {
  security_profile_id = yandex_sws_security_profile.body.id
}
`, name, priority, field, value)
}
