package yandex

import (
	"testing"
)

func TestAccDataSourceOrganizationManagerSamlFederation_byName(t *testing.T) {
	t.Parallel()

	// One federation should be enough to test data-source.
	testAccSamlFederationRunTest(t, testAccDataSourceOrganizationManagerSamlFederationByName, false, 1)
}

func TestAccDataSourceOrganizationManagerSamlFederation_byId(t *testing.T) {
	t.Parallel()

	// One federation should be enough to test data-source.
	testAccSamlFederationRunTest(t, testAccDataSourceOrganizationManagerSamlFederationById, false, 1)
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
	config := templateConfig(samlFederationConfigTemplate, info.Map())
	config += templateConfig(dataSamlFederationConfigTemplateByName, info.Map())
	return config
}

func testAccDataSourceOrganizationManagerSamlFederationById(info *resourceSamlFederationInfo) string {
	config := templateConfig(samlFederationConfigTemplate, info.Map())
	config += templateConfig(dataSamlFederationConfigTemplateById, info.Map())
	return config
}
