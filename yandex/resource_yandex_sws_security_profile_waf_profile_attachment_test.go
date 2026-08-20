package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	smartwebsecurity "github.com/yandex-cloud/go-genproto/yandex/cloud/smartwebsecurity/v1"
	smartwebsecuritysdk "github.com/yandex-cloud/go-sdk/services/smartwebsecurity/v1"
)

func TestAccSmartwebsecuritySecurityProfileWAFAttachment_deleteBeforeProfiles(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-yc-sws-waf-attachment")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactoriesV6,
		Steps: []resource.TestStep{
			{
				Config: testAccSmartwebsecuritySecurityProfileWAFAttachment(name, true),
				Check:  testAccCheckSmartwebsecuritySecurityProfileWAFAttachment(true),
			},
			{
				Config: testAccSmartwebsecuritySecurityProfileWAFAttachment(name, false),
				Check:  testAccCheckSmartwebsecuritySecurityProfileWAFAttachment(false),
			},
		},
	})
}

func testAccCheckSmartwebsecuritySecurityProfileWAFAttachment(attached bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		const securityProfileResource = "yandex_sws_security_profile.this"
		const securityRuleName = "waf"

		rs, ok := s.RootModule().Resources[securityProfileResource]
		if !ok {
			return fmt.Errorf("resource %q not found in state", securityProfileResource)
		}

		conf, err := configForSweepers()
		if err != nil {
			return fmt.Errorf("creating SWS client: %w", err)
		}
		profile, err := smartwebsecuritysdk.NewSecurityProfileClient(conf.SDK).Get(conf.Context(), &smartwebsecurity.GetSecurityProfileRequest{
			SecurityProfileId: rs.Primary.ID,
		})
		if err != nil {
			return fmt.Errorf("reading SWS security profile %q: %w", rs.Primary.ID, err)
		}

		found := false
		persistentRuleFound := false
		for _, rule := range profile.SecurityRules {
			if rule.Name == securityRuleName && rule.GetWaf() != nil {
				found = true
			}
			if rule.Name == "persistent" {
				persistentRuleFound = true
			}
		}
		if found != attached {
			return fmt.Errorf("WAF rule %q attachment state: got %t, want %t", securityRuleName, found, attached)
		}
		if !persistentRuleFound {
			return fmt.Errorf("persistent Security Profile rule was removed while detaching WAF rule %q", securityRuleName)
		}
		return nil
	}
}

func testAccSmartwebsecuritySecurityProfileWAFAttachment(targetName string, enabled bool) string {
	return fmt.Sprintf(`
resource "yandex_sws_waf_profile" "this" {
  count = %[2]t ? 1 : 0
  name  = "%[1]s-waf"

  core_rule_set {
    inbound_anomaly_score = 2
    paranoia_level        = 4
    rule_set {
      name    = "OWASP Core Ruleset"
      version = "4.0.0"
      type    = "CORE"
    }
  }

  analyze_request_body {
    is_enabled        = true
    size_limit        = 8
    size_limit_action = "IGNORE"
  }

  rule_set {
    action     = "DENY"
    is_enabled = true
    priority   = 1
    core_rule_set {
      inbound_anomaly_score = 2
      paranoia_level        = 4
      rule_set {
        name    = "OWASP Core Ruleset"
        version = "4.0.0"
        type    = "CORE"
        id      = "OWASP_CRS_4_0_0"
      }
    }
  }
}

resource "yandex_sws_security_profile" "this" {
  name                     = "%[1]s"
  default_action           = "ALLOW"
  disallow_data_processing = false

  security_rule {
    name     = "persistent"
    priority = 2

    rule_condition {
      action = "ALLOW"
    }
  }

  dynamic "security_rule" {
    for_each = %[2]t ? [1] : []
    content {
      name     = "waf"
      priority = 1

      waf {
        mode           = "FULL"
        waf_profile_id = yandex_sws_waf_profile.this[0].id
      }
    }
  }
}

resource "yandex_sws_security_profile_waf_profile_attachment" "this" {
  count               = %[2]t ? 1 : 0
  security_profile_id = yandex_sws_security_profile.this.id
  waf_profile_id      = yandex_sws_waf_profile.this[0].id
  security_rule_name  = "waf"
}
`, targetName, enabled)
}
