package yandex

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/smartwebsecurity/v1"
)

func init() {
	resource.AddTestSweepers("yandex_sws_security_profile", &resource.Sweeper{
		Name: "yandex_sws_security_profile",
		F:    testSweepSecurityProfile,
		Dependencies: []string{
			"yandex_smartcaptcha_captcha",
		},
	})
}

func TestAccSmartwebsecuritySecurityProfile_basic(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-yc-sc")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactoriesV6,
		Steps: []resource.TestStep{
			{
				Config: testAccSmartwebsecuritySecurityProfileBasic(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_sws_security_profile.this", "name", name),
					resource.TestCheckResourceAttr("yandex_sws_security_profile.this", "default_action", "ALLOW"),
					resource.TestCheckResourceAttr("yandex_sws_security_profile.this", "security_rule.0.name", "smart-protection"),
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

func TestAccSmartwebsecuritySecurityProfile_UpgradeFromSDKv2(t *testing.T) {
	t.Parallel()

	name := acctest.RandomWithPrefix("tf-yc-sc")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckFolderDestroy,
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
				ProtoV6ProviderFactories: testAccProviderFactoriesV6,
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

	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("failed to create SWS client: %w", err)
	}

	profile, err := conf.sdk.SmartWebSecurity().SecurityProfile().Get(
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
  name                             = "%[1]v"
  default_action                   = "ALLOW"
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
  name = "%[1]v"
  default_action = "ALLOW"
  captcha_id = yandex_smartcaptcha_captcha.this.id
  security_rule {
    name = "smart-protection"
    priority = 99999
    smart_protection {
      mode = "FULL"
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

func testSweepSecurityProfile(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := conf.sdk.SmartWebSecurity().SecurityProfile().List(conf.Context(), &smartwebsecurity.ListSecurityProfilesRequest{
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

func sweepSecurityProfile(conf *Config, id string) bool {
	return sweepWithRetry(sweepSecurityProfileOnce, conf, "SecurityProfile", id)
}

func sweepSecurityProfileOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(time.Minute)
	defer cancel()

	op, err := conf.sdk.SmartWebSecurity().SecurityProfile().Delete(ctx, &smartwebsecurity.DeleteSecurityProfileRequest{
		SecurityProfileId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}
