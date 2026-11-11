package acceptance

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1/saml"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
)

func TestAccDataSourceOrganizationManagerSamlFederationUserAccount_byNameID(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	info := newSamlFederationInfo()
	info.AutoCreateAccountOnLogin = true

	name := info.getResourceName(true)
	config := testAccDataSourceOrganizationManagerSamlFederationByNameId(info, "example@example.org")

	var fed saml.Federation
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testAccCheckSamlFederationDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckSamlFederationExists(name, &fed),
				),
			},
		},
	})
}

const resourceSamlFederationUserAccountConfigTemplateByNameID = `
resource "yandex_organizationmanager_saml_federation_user_account" account {
  federation_id = yandex_organizationmanager_saml_federation.{{.ResourceName}}.id
  name_id       = "{{.NameID}}"
}
`

const dataSamlFederationUserAccountConfigTemplateByNameID = `
data "yandex_organizationmanager_saml_federation_user_account" account {
  federation_id = yandex_organizationmanager_saml_federation.{{.ResourceName}}.id
  name_id       = "{{.NameID}}"

  depends_on = [
	yandex_organizationmanager_saml_federation_user_account.account
  ]
}
`

func testAccDataSourceOrganizationManagerSamlFederationByNameId(info *resourceSamlFederationInfo, nameID string) string {
	m := info.Map()
	config := sdkcommon.TemplateConfig(samlFederationConfigTemplate, m)
	m["NameID"] = nameID
	config += sdkcommon.TemplateConfig(resourceSamlFederationUserAccountConfigTemplateByNameID, m)
	config += sdkcommon.TemplateConfig(dataSamlFederationUserAccountConfigTemplateByNameID, m)
	return config
}
