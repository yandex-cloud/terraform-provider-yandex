package acceptance

import (
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	resourcemanagerfixtures "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/resourcemanager/testfixtures"

	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/smartwebsecurity/v1"
	smartwebsecuritysdk "github.com/yandex-cloud/go-sdk/services/smartwebsecurity/v1"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
)

func TestAccSmartwebsecuritySecurityProfile_basic(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	name := acctest.RandomWithPrefix("tf-yc-sc")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		Steps: []resource.TestStep{
			{
				Config: testAccSmartwebsecuritySecurityProfileBasic(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_sws_security_profile.this", "name", name),
					resource.TestCheckResourceAttr("yandex_sws_security_profile.this", "default_action", "ALLOW"),
					resource.TestCheckResourceAttr("yandex_sws_security_profile.this", "security_rule.0.name", "smart-protection"),
					resource.TestCheckResourceAttr("yandex_sws_security_profile.this", "security_rule.1.name", "captcha-rule"),
					resource.TestCheckResourceAttr("yandex_sws_security_profile.this", "security_rule.1.rule_condition.0.action", "CAPTCHA"),
				),
			},
			{
				ResourceName:      "yandex_sws_security_profile.this",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSmartwebsecuritySecurityProfile_defaultsRoundTrip(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	name := acctest.RandomWithPrefix("tf-yc-sws-default-captcha")
	explicitConfig := testAccSmartwebsecuritySecurityProfileExplicitDefaults(name)
	clearedConfig := testAccSmartwebsecuritySecurityProfileClearedDefaults(name)
	defaultConfig := testAccSmartwebsecuritySecurityProfileDefaultCaptcha(name)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		Steps: []resource.TestStep{
			{
				Config: explicitConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_sws_security_profile.this", "name", name),
					resource.TestCheckResourceAttr("yandex_sws_security_profile.this", "security_rule.0.smart_protection.0.mode", "FULL"),
					resource.TestCheckResourceAttrPair("yandex_sws_security_profile.this", "captcha_id", "yandex_smartcaptcha_captcha.this", "id"),
					resource.TestCheckResourceAttr("yandex_sws_security_profile.this", "disallow_data_processing", "false"),
					resource.TestCheckResourceAttr("yandex_sws_security_profile.this", "analyze_request_body.#", "1"),
					resource.TestCheckResourceAttr("yandex_sws_security_profile.this", "analyze_request_body.0.size_limit", "8"),
					resource.TestCheckResourceAttr("yandex_sws_security_profile.this", "analyze_request_body.0.size_limit_action", "DENY"),
				),
			},
			{
				Config: clearedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_sws_security_profile.this", "disallow_data_processing", "true"),
					resource.TestCheckResourceAttr("yandex_sws_security_profile.this", "analyze_request_body.#", "0"),
				),
			},
			{
				Config: defaultConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_sws_security_profile.this", "disallow_data_processing", "false"),
					resource.TestCheckResourceAttr("yandex_sws_security_profile.this", "analyze_request_body.#", "0"),
				),
			},
			{
				ResourceName:      "yandex_sws_security_profile.this",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: defaultConfig,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccSmartwebsecuritySecurityProfile_bodySizeOmission(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	name := acctest.RandomWithPrefix("tf-yc-sws-default-body-size")
	config := testAccSmartwebsecuritySecurityProfileBodySizeOmission(name)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_sws_security_profile.this", "analyze_request_body.#", "1"),
					resource.TestCheckResourceAttr("yandex_sws_security_profile.this", "analyze_request_body.0.size_limit", "0"),
					resource.TestCheckResourceAttr("yandex_sws_security_profile.this", "analyze_request_body.0.size_limit_action", "DENY"),
				),
			},
			{
				ResourceName:      "yandex_sws_security_profile.this",
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

func TestAccSmartwebsecuritySecurityProfile_arlDetach(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	name := acctest.RandomWithPrefix("tf-yc-sws-default-arl")
	explicitConfig := testAccSmartwebsecuritySecurityProfileWithARL(name, true, false)
	clearedConfig := testAccSmartwebsecuritySecurityProfileWithARL(name, false, true)
	omittedConfig := testAccSmartwebsecuritySecurityProfileWithARL(name, false, false)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		Steps: []resource.TestStep{
			{
				Config: explicitConfig,
				Check: resource.TestCheckResourceAttrPair(
					"yandex_sws_security_profile.this", "advanced_rate_limiter_profile_id",
					"yandex_sws_advanced_rate_limiter_profile.this", "id",
				),
			},
			{
				Config: clearedConfig,
				Check: resource.TestCheckResourceAttr(
					"yandex_sws_security_profile.this", "advanced_rate_limiter_profile_id", "",
				),
			},
			{
				Config: omittedConfig,
				Check: resource.TestCheckResourceAttr(
					"yandex_sws_security_profile.this", "advanced_rate_limiter_profile_id", "",
				),
			},
			{
				ResourceName:      "yandex_sws_security_profile.this",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: omittedConfig,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func testAccSmartwebsecuritySecurityProfileExplicitDefaults(targetName string) string {
	return fmt.Sprintf(`
resource "yandex_sws_security_profile" "this" {
  name                     = %[1]q
  default_action           = "ALLOW"
  disallow_data_processing = false
  captcha_id               = yandex_smartcaptcha_captcha.this.id

  analyze_request_body {
    size_limit        = 8
    size_limit_action = "DENY"
  }

  security_rule {
    name     = "smart-protection"
    priority = 99999

    smart_protection {
      mode = "FULL"
    }
  }
}

resource "yandex_smartcaptcha_captcha" "this" {
  name           = "%[1]s-captcha"
  complexity     = "MEDIUM"
  pre_check_type = "CHECKBOX"
  challenge_type = "IMAGE_TEXT"
  allowed_sites  = ["*"]
}
`, targetName)
}

func testAccSmartwebsecuritySecurityProfileClearedDefaults(targetName string) string {
	return fmt.Sprintf(`
resource "yandex_sws_security_profile" "this" {
  name                     = %[1]q
  default_action           = "ALLOW"
  disallow_data_processing = true
  captcha_id               = ""

  security_rule {
    name     = "smart-protection"
    priority = 99999

    smart_protection {
      mode = "FULL"
    }
  }
}
`, targetName)
}

func testAccSmartwebsecuritySecurityProfileBodySizeOmission(targetName string) string {
	return fmt.Sprintf(`
resource "yandex_sws_security_profile" "this" {
  name                     = %[1]q
  default_action           = "ALLOW"
  disallow_data_processing = false

  analyze_request_body {
    size_limit_action = "DENY"
  }
}
`, targetName)
}

func testAccSmartwebsecuritySecurityProfileWithARL(targetName string, attach, clear bool) string {
	arlID := ""
	if attach {
		arlID = "  advanced_rate_limiter_profile_id = yandex_sws_advanced_rate_limiter_profile.this.id\n"
	} else if clear {
		arlID = "  advanced_rate_limiter_profile_id = \"\"\n"
	}

	return fmt.Sprintf(`
resource "yandex_sws_security_profile" "this" {
  name                     = %[1]q
  default_action           = "ALLOW"
  disallow_data_processing = false
%[2]s}

resource "yandex_sws_advanced_rate_limiter_profile" "this" {
  name = "%[1]s-arl"

  advanced_rate_limiter_rule {
    name     = "rule1"
    priority = 10

    static_quota {
      action = "DENY"
      limit  = 100
      period = 1
    }
  }
}
`, targetName, arlID)
}

func TestAccSmartwebsecuritySecurityProfile_UpgradeFromSDKv2(t *testing.T) {
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
						Source: "yandex-cloud/yandex",
					},
				},
				Config: testAccSmartwebsecuritySecurityProfileBasicMigration(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_sws_security_profile.this", "name", name),
					resource.TestCheckResourceAttr("yandex_sws_security_profile.this", "default_action", "ALLOW"),
					testAccCheckSmartwebsecuritySecurityProfileASNMatchers,
				),
			},
			{
				ProtoV6ProviderFactories: suite.Protocol6Factories(),
				Config:                   testAccSmartwebsecuritySecurityProfileBasicMigration(name),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func testAccCheckSmartwebsecuritySecurityProfileASNMatchers(s *terraform.State) error {
	const resourceName = "yandex_sws_security_profile.this"

	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return fmt.Errorf("resource %q not found in state", resourceName)
	}

	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("failed to create SWS client: %w", err)
	}

	client := smartwebsecuritysdk.NewSecurityProfileClient(conf.SDK)
	profile, err := client.Get(
		conf.Context(),
		&smartwebsecurity.GetSecurityProfileRequest{SecurityProfileId: rs.Primary.ID},
	)
	if err != nil {
		return fmt.Errorf("failed to get SWS security profile %q: %w", rs.Primary.ID, err)
	}

	for _, rule := range profile.SecurityRules {
		if rule.Name != "rule-condition-3" {
			continue
		}

		sourceIP := rule.GetRuleCondition().GetCondition().GetSourceIp()
		if sourceIP == nil {
			return fmt.Errorf("source_ip is missing in rule %q", rule.Name)
		}

		if got, want := sourceIP.GetAsnRangesMatch().GetAsnRanges(), []int64{64496, 4294967295}; !reflect.DeepEqual(got, want) {
			return fmt.Errorf("unexpected ASN ranges to match: got %v, want %v", got, want)
		}
		if got, want := sourceIP.GetAsnRangesNotMatch().GetAsnRanges(), []int64{0}; !reflect.DeepEqual(got, want) {
			return fmt.Errorf("unexpected ASN ranges to not match: got %v, want %v", got, want)
		}

		return nil
	}

	return fmt.Errorf("rule %q not found in SWS security profile", "rule-condition-3")
}

func testAccSmartwebsecuritySecurityProfileBasicMigration(targetName string) string {
	return fmt.Sprintf(`
resource "yandex_sws_security_profile" "this" {
  name                     = "%[1]v"
  default_action           = "ALLOW"
  disallow_data_processing = false
  captcha_id = yandex_smartcaptcha_captcha.this.id

  security_rule {
    name     = "smart-protection"
    priority = 99999

    smart_protection {
      mode = "API"
    }
  }

  security_rule {
    name     = "rule-condition-1"
    priority = 1

    rule_condition {
      action = "ALLOW"

      condition {
        authority {
          authority_matcher {
            exact_match = "example.com"
          }
        }
      }
    }
  }

  security_rule {
    name     = "rule-condition-2"
    priority = 2

    rule_condition {
      action = "DENY"

      condition {
        http_method {
          http_method_matcher {
            exact_match = "DELETE"
          }
        }
      }
    }
  }

  security_rule {
    name     = "rule-condition-3"
    priority = 3

    rule_condition {
      action = "DENY"

      condition {
        request_uri {
          path {
            prefix_match = "/form"
          }
          queries {
            key = "firstname"
            value {
              pire_regex_match = ".*ivan.*"
            }
          }
          queries {
            key = "lastname"
            value {
              pire_regex_match = ".*petr.*"
            }
          }
        }

        headers {
          name = "User-Agent"
          value {
            pire_regex_match = ".*curl.*"
          }
        }
        headers {
          name = "Referer"
          value {
            pire_regex_not_match = ".*bot.*"
          }
        }

        source_ip {
          ip_ranges_match {
            ip_ranges = ["1.2.33.44", "2.3.4.56"]
          }
          ip_ranges_not_match {
            ip_ranges = ["8.8.0.0/16", "10::1234:1abc:1/64"]
          }
          geo_ip_match {
            locations = ["ru", "es"]
          }
          geo_ip_not_match {
            locations = ["us", "fm", "gb"]
          }
          asn_ranges_match {
            asn_ranges = [64496, 4294967295]
          }
          asn_ranges_not_match {
            asn_ranges = [0]
          }
        }
      }
    }
  }
}

resource "yandex_smartcaptcha_captcha" "this" {
  name = "%[1]v-captcha"
  complexity = "MEDIUM"
  pre_check_type = "CHECKBOX"
  challenge_type = "IMAGE_TEXT"
  allowed_sites = ["*"]
}
`, targetName)
}

func testAccSmartwebsecuritySecurityProfileBasic(targetName string) string {
	return fmt.Sprintf(`
resource "yandex_sws_security_profile" "this" {	
  name                     = "%[1]v"
  default_action           = "ALLOW"
  disallow_data_processing = false
  captcha_id               = yandex_smartcaptcha_captcha.this.id
  security_rule {
    name = "smart-protection"
    priority = 99999
    smart_protection {
      mode = "FULL"
    }
  }

  security_rule {
    name     = "captcha-rule"
    priority = 1

    rule_condition {
      action = "CAPTCHA"

      condition {
        request_uri {
          path {
            prefix_match = "/captcha"
          }
        }
      }
    }
  }
}
resource "yandex_smartcaptcha_captcha" "this" {
  name = "%[1]v-captcha"
  complexity = "MEDIUM"
  pre_check_type = "CHECKBOX"
  challenge_type = "IMAGE_TEXT"
  allowed_sites = ["*"]
}
`, targetName)
}

func testAccSmartwebsecuritySecurityProfileDefaultCaptcha(targetName string) string {
	return fmt.Sprintf(`
resource "yandex_sws_security_profile" "this" {
  name                     = %[1]q
  default_action           = "ALLOW"
  disallow_data_processing = false

  security_rule {
    name     = "smart-protection"
    priority = 99999

    smart_protection {
      mode = "FULL"
    }
  }
}
`, targetName)
}
