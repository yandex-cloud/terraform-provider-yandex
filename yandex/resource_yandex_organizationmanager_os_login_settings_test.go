package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	organizationManagerOsLoginSettings = "yandex_organizationmanager_os_login_settings.this"
)

func TestAccOrganizationManagerOsLoginSettings_import(t *testing.T) {
	t.Parallel()
	organizationID := getExampleOrganizationID()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigOrganizationManagerOsLoginSettings(organizationID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(organizationManagerOsLoginSettings, "organization_id", organizationID),
					resource.TestCheckResourceAttr(organizationManagerOsLoginSettings, "user_ssh_key_settings.0.enabled", "true"),
					resource.TestCheckResourceAttr(organizationManagerOsLoginSettings, "user_ssh_key_settings.0.allow_manage_own_keys", "true"),
					resource.TestCheckResourceAttr(organizationManagerOsLoginSettings, "ssh_certificate_settings.0.enabled", "true"),
				),
			},
		},
	})
}

func testAccConfigOrganizationManagerOsLoginSettings(organizationID string) string {
	return fmt.Sprintf(`
resource "yandex_organizationmanager_os_login_settings" "this" {
  organization_id = "%s"
  user_ssh_key_settings {
    enabled               = true
    allow_manage_own_keys = true
  }
  ssh_certificate_settings {
    enabled = true
  }
}
`, organizationID)
}
