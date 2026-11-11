package acceptance

import (
	"testing"

	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
)

func TestAccDataSourceOrganizationManagerSamlFederation_byName(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	// One federation should be enough to test data-source.
	suite.testAccSamlFederationRunTest(t, testAccDataSourceOrganizationManagerSamlFederationByName, false, 1)
}

func TestAccDataSourceOrganizationManagerSamlFederation_byId(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	// One federation should be enough to test data-source.
	suite.testAccSamlFederationRunTest(t, testAccDataSourceOrganizationManagerSamlFederationById, false, 1)
}

const dataSamlFederationConfigTemplateByName = `
data "yandex_organizationmanager_saml_federation" {{.ResourceName}} {
  name            = yandex_organizationmanager_saml_federation.{{.ResourceName}}.name
  organization_id = "{{.OrganizationId}}"
}
`

const dataSamlFederationConfigTemplateById = `
data "yandex_organizationmanager_saml_federation" {{.ResourceName}} {
  federation_id = yandex_organizationmanager_saml_federation.{{.ResourceName}}.id
}
`

func testAccDataSourceOrganizationManagerSamlFederationByName(info *resourceSamlFederationInfo) string {
	config := sdkcommon.TemplateConfig(samlFederationConfigTemplate, info.Map())
	config += sdkcommon.TemplateConfig(dataSamlFederationConfigTemplateByName, info.Map())
	return config
}

func testAccDataSourceOrganizationManagerSamlFederationById(info *resourceSamlFederationInfo) string {
	config := sdkcommon.TemplateConfig(samlFederationConfigTemplate, info.Map())
	config += sdkcommon.TemplateConfig(dataSamlFederationConfigTemplateById, info.Map())
	return config
}
