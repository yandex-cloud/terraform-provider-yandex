package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/fatih/structs"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1"
)

func init() {
	resource.AddTestSweepers("yandex_organizationmanager_group", &resource.Sweeper{
		Name:         "yandex_organizationmanager_group",
		F:            testSweepGroups,
		Dependencies: []string{},
	})
}

func testSweepGroupOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexOrganizationManagerGroupDefaultTimeout)
	defer cancel()

	op, err := conf.sdk.OrganizationManager().Group().Delete(ctx, &organizationmanager.DeleteGroupRequest{
		GroupId: id,
	})

	return handleSweepOperation(ctx, conf, op, err)
}

func testSweepGroups(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &organizationmanager.ListGroupsRequest{
		OrganizationId: getExampleOrganizationID(),
	}
	it := conf.sdk.OrganizationManager().Group().GroupIterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepWithRetry(testSweepGroupOnce, conf, "Group", id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Group %q", id))
		}
	}

	return result.ErrorOrNil()
}

func TestAccOrganizationManagerGroup_createAndUpdate(t *testing.T) {
	t.Parallel()

	// Doing 2 runs effectively means one create and one subsequent update operation.
	testAccGroupRunTest(t, testAccOrganizationManagerGroup, true, 1)
}

func TestAccOrganizationManagerGroup_import(t *testing.T) {
	t.Parallel()

	info := newGroupInfo()
	name := info.getResourceName(true)

	var group organizationmanager.Group
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationManagerGroup(info),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(name, &group),
				),
			},
			organizationGroupImportStep(name, group.Id),
		},
	})
}

// The config here should match as closely as possible to the one presented to the user in the docs.
// Serves as a proof that the example config is viable.
func TestAccOrganizationManagerGroup_example(t *testing.T) {
	t.Parallel()

	config := fmt.Sprintf(`
resource "yandex_organizationmanager_group" group {
  name            = "my-group"
  description     = "My new Group"
  organization_id = "%s"
}
`, getExampleOrganizationID())

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
			},
		},
	})
}

type GroupConfigGenerateFunc func(info *resourceGroupInfo) string

func testAccGroupRunTest(t *testing.T, fun GroupConfigGenerateFunc, rs bool, n int) {
	// Generate n groups, apply them to Terraform using fun and test according to resource type.
	for i := 0; i < n; i++ {
		info := newGroupInfo()
		var group organizationmanager.Group
		resource.Test(t, resource.TestCase{
			PreCheck:          func() { testAccPreCheck(t) },
			ProviderFactories: testAccProviderFactories,
			CheckDestroy:      testAccCheckGroupDestroy,
			Steps: []resource.TestStep{
				{
					Config: fun(info),
					Check: resource.ComposeTestCheckFunc(
						testAccCheckGroupExists(info.getResourceName(rs), &group),
						GroupResourceTestCheckFunc(&group, info, rs),
					),
				},
			},
		})
	}
}

func testAccCheckGroupExists(n string, group *organizationmanager.Group) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.OrganizationManager().Group().Get(context.Background(), &organizationmanager.GetGroupRequest{
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

func testAccCheckGroupDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_organizationmanager_group" {
			continue
		}

		_, err := config.sdk.OrganizationManager().Group().Get(context.Background(), &organizationmanager.GetGroupRequest{
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
	return newGroupInfoByOrganizationID(getExampleOrganizationID())
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
	return templateConfig(groupConfigTemplate, info.Map())
}

func organizationGroupImportStep(resourceName, groupID string) resource.TestStep {
	return resource.TestStep{
		ResourceName:      resourceName,
		ImportStateId:     groupID,
		ImportState:       true,
		ImportStateVerify: true,
	}
}
