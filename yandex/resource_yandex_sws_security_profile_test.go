package yandex

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
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
