package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceSmartcaptchaCaptcha_byID(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-yc-sc")
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceCaptchaConfig(name, true),
				Check:  testAccDataSourceCaptchaCheck(name),
			},
		},
	})
}

func TestAccDataSourceSmartcaptchaCaptcha_byName(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-yc-sc")
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceCaptchaConfig(name, false),
				Check:  testAccDataSourceCaptchaCheck(name),
			},
		},
	})
}

func testAccDataSourceCaptchaCheck(name string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr("yandex_smartcaptcha_captcha.this", "name", name),
		resource.TestCheckResourceAttr("yandex_smartcaptcha_captcha.this", "complexity", "HARD"),
		resource.TestCheckResourceAttr("yandex_smartcaptcha_captcha.this", "pre_check_type", "SLIDER"),
		resource.TestCheckResourceAttr("yandex_smartcaptcha_captcha.this", "challenge_type", "IMAGE_TEXT"),
	)
}

func testAccCaptchaResourceConfig(name string) string {
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
`, name)
}

const CaptchaDataByIDConfig = `
data "yandex_smartcaptcha_captcha" "res" {
	captcha_id = yandex_smartcaptcha_captcha.this.id
}
`

const CaptchaDataByNameConfig = `
data "yandex_smartcaptcha_captcha" "res" {
	name = yandex_smartcaptcha_captcha.this.name
}
`

func testAccDataSourceCaptchaConfig(name string, withDataID bool) string {
	if withDataID {
		return testAccCaptchaResourceConfig(name) + CaptchaDataByIDConfig
	} else {
		return testAccCaptchaResourceConfig(name) + CaptchaDataByNameConfig
	}
}
