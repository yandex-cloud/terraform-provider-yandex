package yandex

import (
	"testing"
)

func TestAccDataSourceOrganizationManagerGroup_byName(t *testing.T) {
	t.Parallel()

	// One group should be enough to test data-source.
	testAccGroupRunTest(t, testAccDataSourceOrganizationManagerGroupByName, false, 1)
}

func TestAccDataSourceOrganizationManagerGroup_byId(t *testing.T) {
	t.Parallel()

	// One group should be enough to test data-source.
	testAccGroupRunTest(t, testAccDataSourceOrganizationManagerGroupById, false, 1)
}

const dataGroupConfigTemplateByName = `
data "yandex_organizationmanager_group" {{.ResourceName}} {
  name            = yandex_organizationmanager_group.{{.ResourceName}}.name
  organization_id = "{{.OrganizationId}}"
}
`

const dataGroupConfigTemplateById = `
data "yandex_organizationmanager_group" {{.ResourceName}} {
  group_id = yandex_organizationmanager_group.{{.ResourceName}}.id
}
`

func testAccDataSourceOrganizationManagerGroupByName(info *resourceGroupInfo) string {
	config := templateConfig(groupConfigTemplate, info.Map())
	config += templateConfig(dataGroupConfigTemplateByName, info.Map())
	return config
}

func testAccDataSourceOrganizationManagerGroupById(info *resourceGroupInfo) string {
	config := templateConfig(groupConfigTemplate, info.Map())
	config += templateConfig(dataGroupConfigTemplateById, info.Map())
	return config
}
