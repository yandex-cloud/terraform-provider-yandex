package acceptance

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/acceptanceenv"

	"testing"

	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
)

const (
	organizationManagerOsLoginSettingsData = "data.yandex_organizationmanager_os_login_settings.this_data"
)

func TestAccDataSourceOrganizationManagerOsLoginSettings(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()
	organizationID := acceptanceenv.OrganizationID()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccDataOrganizationManagerOsLoginSettings(organizationID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(organizationManagerOsLoginSettingsData, "organization_id", organizationID),
					resource.TestCheckResourceAttr(organizationManagerOsLoginSettingsData, "user_ssh_key_settings.0.enabled", "true"),
					resource.TestCheckResourceAttr(organizationManagerOsLoginSettingsData, "user_ssh_key_settings.0.allow_manage_own_keys", "true"),
					resource.TestCheckResourceAttr(organizationManagerOsLoginSettingsData, "ssh_certificate_settings.0.enabled", "true"),
				),
			},
		},
	})
}

func testAccDataOrganizationManagerOsLoginSettings(organizationID string) string {
	return testAccConfigOrganizationManagerOsLoginSettings(organizationID) + testAccDataConfigOrganizationManagerOsLoginSettings()
}

func testAccDataConfigOrganizationManagerOsLoginSettings() string {
	return `
data "yandex_organizationmanager_os_login_settings" "this_data" {
  organization_id = yandex_organizationmanager_os_login_settings.this.id
}
`
}
