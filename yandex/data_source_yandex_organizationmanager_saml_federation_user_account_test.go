package yandex

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1/saml"
)

func TestAccDataSourceOrganizationManagerSamlFederationUserAccount_byNameID(t *testing.T) {
	t.Parallel()

	info := newSamlFederationInfo()
	info.AutoCreateAccountOnLogin = true

	name := info.getResourceName(true)
	config := testAccDataSourceOrganizationManagerSamlFederationByNameId(info, "example@example.org")

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
				),
			},
		},
	})
}

const dataSamlFederationUserAccountConfigTemplateByNameID = `
data "yandex_organizationmanager_saml_federation_user_account" account {
  federation_id = yandex_organizationmanager_saml_federation.{{.ResourceName}}.id
  name_id       = "{{.NameID}}"
}
`

func testAccDataSourceOrganizationManagerSamlFederationByNameId(info *resourceSamlFederationInfo, nameID string) string {
	m := info.Map()
	config := templateConfig(samlFederationConfigTemplate, m)
	m["NameID"] = nameID
	config += templateConfig(dataSamlFederationUserAccountConfigTemplateByNameID, m)
	return config
}
