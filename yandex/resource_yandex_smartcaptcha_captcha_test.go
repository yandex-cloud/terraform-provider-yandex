package yandex

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/smartcaptcha/v1"
)

func init() {
	resource.AddTestSweepers("yandex_smartcaptcha_captcha", &resource.Sweeper{
		Name: "yandex_smartcaptcha_captcha",
		F:    testSweepCaptcha,
	})
}

func TestAccSmartcaptchaCaptcha_basic(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-yc-sc")
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSmartcaptchaCaptchaBasic(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_smartcaptcha_captcha.this", "name", name),
					resource.TestCheckResourceAttr("yandex_smartcaptcha_captcha.this", "complexity", "HARD"),
					resource.TestCheckResourceAttr("yandex_smartcaptcha_captcha.this", "pre_check_type", "SLIDER"),
					resource.TestCheckResourceAttr("yandex_smartcaptcha_captcha.this", "challenge_type", "IMAGE_TEXT"),
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

func testAccSmartcaptchaCaptchaBasic(targetName string) string {
	return fmt.Sprintf(`
resource "yandex_smartcaptcha_captcha" "this" {
  name = "%s"
  deletion_protection = false
  complexity = "HARD"
  pre_check_type = "SLIDER"
  challenge_type = "IMAGE_TEXT"
  allowed_sites = ["example.com", "example.ru"]
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

func testSweepCaptcha(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := conf.sdk.SmartCaptcha().Captcha().List(conf.Context(), &smartcaptcha.ListCaptchasRequest{
		FolderId: conf.FolderID,
	})
	if err != nil {
		return fmt.Errorf("error getting SmartCaptcha captchas: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.Resources {
		if !sweepCaptcha(conf, c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep SmartCaptcha captcha %q", c.Id))
		}
	}

	return result.ErrorOrNil()
}

func sweepCaptcha(conf *Config, id string) bool {
	return sweepWithRetry(sweepCaptchaOnce, conf, "Captcha", id)
}

func sweepCaptchaOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(time.Minute)
	defer cancel()

	op, err := conf.sdk.SmartCaptcha().Captcha().Delete(ctx, &smartcaptcha.DeleteCaptchaRequest{
		CaptchaId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}
