package acceptance

import (
	"fmt"
	"testing"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/acceptanceenv"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
)

const (
	organizationManagerOsLoginSettings = "yandex_organizationmanager_os_login_settings.this"
)

func TestAccOrganizationManagerOsLoginSettings_import(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()
	organizationID := acceptanceenv.OrganizationID()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
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
