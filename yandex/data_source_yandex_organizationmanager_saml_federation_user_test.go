package yandex

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1/saml"
)

func TestAccDataSourceOrganizationManagerSamlFederationUser_byUserAccountID(t *testing.T) {
	t.Parallel()

	info := newSamlFederationInfo()
	info.AutoCreateAccountOnLogin = true

	name := info.getResourceName(true)
	config := testAccDataSourceOrganizationManagerSamlFederationUserByUserAccountID(info, "example@example.org")

	var fed saml.Federation
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSamlFederationDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSamlFederationExists(name, &fed),
					resource.TestCheckResourceAttr("data.yandex_organizationmanager_saml_federation_user.account", "name_id", "example@example.org"),
					resource.TestCheckResourceAttrSet("data.yandex_organizationmanager_saml_federation_user.account", "id"),
				),
			},
		},
	})
}

const resourceSamlFederationUserConfigTemplateByUserAccountID = `
resource "yandex_organizationmanager_saml_federation_user_account" account {
  federation_id = yandex_organizationmanager_saml_federation.{{.ResourceName}}.id
  name_id       = "{{.NameID}}"
}
`

const dataSamlFederationUserConfigTemplateByUserAccountID = `
data "yandex_organizationmanager_saml_federation_user" account {
  federation_id   = yandex_organizationmanager_saml_federation.{{.ResourceName}}.id
  user_account_id = yandex_organizationmanager_saml_federation_user_account.account.id
}
`

func testAccDataSourceOrganizationManagerSamlFederationUserByUserAccountID(info *resourceSamlFederationInfo, nameID string) string {
	m := info.Map()
	config := templateConfig(samlFederationConfigTemplate, m)
	m["NameID"] = nameID
	config += templateConfig(resourceSamlFederationUserConfigTemplateByUserAccountID, m)
	config += templateConfig(dataSamlFederationUserConfigTemplateByUserAccountID, m)
	return config
}
