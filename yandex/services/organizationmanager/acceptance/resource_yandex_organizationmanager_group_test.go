package acceptance

import (
	"context"
	"fmt"
	"testing"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/acceptanceenv"

	"github.com/fatih/structs"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1"
	organizationmanagersdk "github.com/yandex-cloud/go-sdk/services/organizationmanager/v1"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
)

type GroupConfigGenerateFunc func(info *resourceGroupInfo) string

func (suite testSuite) testAccGroupRunTest(t *testing.T, fun GroupConfigGenerateFunc, rs bool, n int) {
	// Generate n groups, apply them to Terraform using fun and test according to resource type.
	for i := 0; i < n; i++ {
		info := newGroupInfo()
		var group organizationmanager.Group
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { acceptance.PreCheck(t) },
			ProtoV6ProviderFactories: suite.Protocol6Factories(),
			CheckDestroy:             suite.testAccCheckGroupDestroy,
			Steps: []resource.TestStep{
				{
					Config: fun(info),
					Check: resource.ComposeTestCheckFunc(
						suite.testAccCheckGroupExists(info.getResourceName(rs), &group),
						GroupResourceTestCheckFunc(&group, info, rs),
					),
				},
			},
		})
	}
}

func (suite testSuite) testAccCheckGroupExists(n string, group *organizationmanager.Group) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		config, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}

		client := organizationmanagersdk.NewGroupClient(config.SDK)

		found, err := client.Get(context.Background(), &organizationmanager.GetGroupRequest{
			GroupId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Group not found")
		}

		*group = *found
		return nil
	}
}

func GroupResourceTestCheckFunc(group *organizationmanager.Group, groupInfo *resourceGroupInfo, rs bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		name := groupInfo.getResourceName(rs)
		checkFuncsAr := []resource.TestCheckFunc{
			resource.TestCheckResourceAttr(name, "name", groupInfo.Name),
			resource.TestCheckResourceAttr(name, "name", group.Name),

			resource.TestCheckResourceAttr(name, "description", groupInfo.Description),
			resource.TestCheckResourceAttr(name, "description", group.Description),
			resource.TestCheckResourceAttrSet(name, "created_at"),
			resource.TestCheckResourceAttrSet(name, "organization_id"),
		}
		if !rs {
			checkFuncsAr = append(checkFuncsAr, resource.TestCheckResourceAttrSet(name, "members.#"))
		}
		return resource.ComposeTestCheckFunc(checkFuncsAr...)(s)
	}
}

func (suite testSuite) testAccCheckGroupDestroy(s *terraform.State) error {
	config, configErr := suite.Config()
	if configErr != nil {
		return configErr
	}

	client := organizationmanagersdk.NewGroupClient(config.SDK)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_organizationmanager_group" {
			continue
		}

		_, err := client.Get(context.Background(), &organizationmanager.GetGroupRequest{
			GroupId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("Group still exists")
		}
	}

	return nil
}

type resourceGroupInfo struct {
	OrganizationId string
	Name           string
	Description    string
	ResourceName   string
}

func newGroupInfo() *resourceGroupInfo {
	return newGroupInfoByOrganizationID(acceptanceenv.OrganizationID())
}

func newGroupInfoByOrganizationID(organizationID string) *resourceGroupInfo {
	return &resourceGroupInfo{
		OrganizationId: organizationID,
		Name:           acctest.RandomWithPrefix("tf-acc"),
		Description:    acctest.RandString(20),
		ResourceName:   "foobar",
	}
}

func (i *resourceGroupInfo) Map() map[string]interface{} {
	return structs.Map(i)
}

func (i *resourceGroupInfo) getResourceName(rs bool) string {
	if rs {
		return "yandex_organizationmanager_group." + i.ResourceName
	}
	return "data.yandex_organizationmanager_group." + i.ResourceName
}

const groupConfigTemplate = `
resource "yandex_organizationmanager_group" {{.ResourceName}} {
  name                         = "{{.Name}}"
  description                  = "{{.Description}}"
  organization_id              = "{{.OrganizationId}}"
}
`

func testAccOrganizationManagerGroup(info *resourceGroupInfo) string {
	return sdkcommon.TemplateConfig(groupConfigTemplate, info.Map())
}
