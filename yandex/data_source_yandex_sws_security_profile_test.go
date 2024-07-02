package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceSmartwebsecuritySecurityProfile_byID(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-yc-sp")
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSecurityProfileConfig(name, true),
				Check:  testAccDataSourceSecurityProfileCheck(name),
			},
		},
	})
}

func TestAccDataSourceSmartwebsecuritySecurityProfile_byName(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-yc-sp")
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSecurityProfileConfig(name, false),
				Check:  testAccDataSourceSecurityProfileCheck(name),
			},
		},
	})
}

func testAccDataSourceSecurityProfileCheck(name string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr("data.yandex_sws_security_profile.res", "name", name),
		resource.TestCheckResourceAttr("data.yandex_sws_security_profile.res", "default_action", "ALLOW"),
		resource.TestCheckResourceAttr("data.yandex_sws_security_profile.res", "security_rule.0.name", "smart-protection"),
	)
}

func testAccSecurityProfileResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "yandex_sws_security_profile" "this" {	
	name = "%[1]v"
	default_action = "ALLOW"
	security_rule {
		name = "smart-protection"
		priority = 99999
		smart_protection {
			mode = "FULL"
		}
	}
}
`, name)
}

const securityProfileDataByIDConfig = `
data "yandex_sws_security_profile" "res" {
	security_profile_id = yandex_sws_security_profile.this.id
}
`

const securityProfileDataByNameConfig = `
data "yandex_sws_security_profile" "res" {
	name = yandex_sws_security_profile.this.name
}
`

func testAccDataSourceSecurityProfileConfig(name string, withDataID bool) string {
	if withDataID {
		return testAccSecurityProfileResourceConfig(name) + securityProfileDataByIDConfig
	} else {
		return testAccSecurityProfileResourceConfig(name) + securityProfileDataByNameConfig
	}
}
