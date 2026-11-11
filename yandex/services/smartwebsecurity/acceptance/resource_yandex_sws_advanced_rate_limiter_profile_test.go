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

func TestAccSmartwebsecurityArlProfile_basic(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	name := acctest.RandomWithPrefix("tf-yc-arlp")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		Steps: []resource.TestStep{
			{
				Config: testAccSmartwebsecurityArlProfileBasic(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_sws_advanced_rate_limiter_profile.this", "name", name),
					resource.TestCheckResourceAttr("yandex_sws_advanced_rate_limiter_profile.this", "advanced_rate_limiter_rule.0.priority", "10"),
					resource.TestCheckResourceAttr("yandex_sws_advanced_rate_limiter_profile.this", "advanced_rate_limiter_rule.0.static_quota.0.action", "DENY"),
					resource.TestCheckResourceAttr("yandex_sws_advanced_rate_limiter_profile.this", "advanced_rate_limiter_rule.1.dynamic_quota.0.ban_period", "60"),
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

func TestAccSmartwebsecurityArlProfile_UpgradeFromSDKv2(t *testing.T) {
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
				Config: testAccSmartwebsecurityArlProfileLegacy(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_sws_advanced_rate_limiter_profile.this", "name", name),
				),
			},
			{
				ProtoV6ProviderFactories: suite.Protocol6Factories(),
				Config:                   testAccSmartwebsecurityArlProfileLegacy(name),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
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

    advanced_rate_limiter_rule {
        name        = "rule2"
        priority    = 20
        description = "Dynamic quota with temporary ban"
        dry_run     = false

        dynamic_quota {
            action     = "DENY"
            limit      = 100
            period     = 60
            ban_period = 60

            characteristic {
                simple_characteristic {
                    type = "REQUEST_PATH"
                }
            }
        }
    }
}
`, targetName)
}

func testAccSmartwebsecurityArlProfileLegacy(targetName string) string {
	return fmt.Sprintf(`
resource "yandex_sws_advanced_rate_limiter_profile" "this" {
    name = "%[1]v"
    advanced_rate_limiter_rule {
        name        = "rule1"
        priority    = 10
        description = "First test rule"
        dry_run     = true
        static_quota {
            action = "DENY"
            limit  = 10000000
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
