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

func TestAccSmartcaptchaCaptcha_basic(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	name := acctest.RandomWithPrefix("tf-yc-sc")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		Steps: []resource.TestStep{
			{
				Config: testAccSmartcaptchaCaptchaBasic(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_smartcaptcha_captcha.this", "name", name),
					resource.TestCheckResourceAttr("yandex_smartcaptcha_captcha.this", "complexity", "HARD"),
					resource.TestCheckResourceAttr("yandex_smartcaptcha_captcha.this", "pre_check_type", "SLIDER"),
					resource.TestCheckResourceAttr("yandex_smartcaptcha_captcha.this", "challenge_type", "IMAGE_TEXT"),
					resource.TestCheckResourceAttr("yandex_smartcaptcha_captcha.this", "disallow_data_processing", "false"),
					resource.TestCheckResourceAttr("yandex_smartcaptcha_captcha.this", "description", "description"),
					resource.TestCheckResourceAttr("yandex_smartcaptcha_captcha.this", "labels.key", "value"),
				),
			},
			{
				ResourceName:      "yandex_smartcaptcha_captcha.this",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSmartcaptchaCaptcha_UpgradeFromSDKv2(t *testing.T) {
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
				Config: testAccSmartcaptchaCaptchaBasicMigration(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_smartcaptcha_captcha.this", "name", name),
					resource.TestCheckResourceAttr("yandex_smartcaptcha_captcha.this", "complexity", "HARD"),
					resource.TestCheckResourceAttr("yandex_smartcaptcha_captcha.this", "pre_check_type", "SLIDER"),
					resource.TestCheckResourceAttr("yandex_smartcaptcha_captcha.this", "challenge_type", "IMAGE_TEXT"),
				),
			},
			{
				ProtoV6ProviderFactories: suite.Protocol6Factories(),
				Config:                   testAccSmartcaptchaCaptchaBasicMigration(name),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func testAccSmartcaptchaCaptchaBasicMigration(targetName string) string {
	return fmt.Sprintf(`
resource "yandex_smartcaptcha_captcha" "this" {
  name = "%s"
  deletion_protection = false
  complexity = "HARD"
  pre_check_type = "SLIDER"
  challenge_type = "IMAGE_TEXT"
  allowed_sites = ["example.com", "example.ru"]
  timeouts {
    create = "30m"
    update = "72h"
    delete = "20m"
  }
  override_variant {
    uuid = "yyy"
    description = "override variant 2"

    complexity = "HARD"
    pre_check_type = "CHECKBOX"
    challenge_type = "KALEIDOSCOPE"
  }
  security_rule {
    name = "rule3"
    priority = 99999
    description = "Empty condition rule"
    override_variant_uuid = "yyy"
	condition {
      host {
        host_matcher {
          exact_match = "example.com"
        }
      }

      uri {
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
      }
    }
  }
}
`, targetName)
}

func testAccSmartcaptchaCaptchaBasic(targetName string) string {
	return fmt.Sprintf(`
resource "yandex_smartcaptcha_captcha" "this" {
  name = "%s"
  deletion_protection = false
  complexity = "HARD"
  pre_check_type = "SLIDER"
  challenge_type = "IMAGE_TEXT"
  allowed_sites = ["example.com", "example.ru"]
  labels = {
	key = "value"
  }
  disallow_data_processing = false
  description = "description"
  override_variant {
    uuid = "yyy"
    description = "override variant 2"

    complexity = "HARD"
    pre_check_type = "CHECKBOX"
    challenge_type = "KALEIDOSCOPE"
  }
  security_rule {
    name = "rule3"
    priority = 99999
    description = "Empty condition rule"
    override_variant_uuid = "yyy"
  }
}
`, targetName)
}
