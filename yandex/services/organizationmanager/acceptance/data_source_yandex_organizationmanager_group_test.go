package acceptance

import (
	"testing"

	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
)

func TestAccDataSourceOrganizationManagerGroup_byName(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	// One group should be enough to test data-source.
	suite.testAccGroupRunTest(t, testAccDataSourceOrganizationManagerGroupByName, false, 1)
}

func TestAccDataSourceOrganizationManagerGroup_byId(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	// One group should be enough to test data-source.
	suite.testAccGroupRunTest(t, testAccDataSourceOrganizationManagerGroupById, false, 1)
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
	config := sdkcommon.TemplateConfig(groupConfigTemplate, info.Map())
	config += sdkcommon.TemplateConfig(dataGroupConfigTemplateByName, info.Map())
	return config
}

func testAccDataSourceOrganizationManagerGroupById(info *resourceGroupInfo) string {
	config := sdkcommon.TemplateConfig(groupConfigTemplate, info.Map())
	config += sdkcommon.TemplateConfig(dataGroupConfigTemplateById, info.Map())
	return config
}
