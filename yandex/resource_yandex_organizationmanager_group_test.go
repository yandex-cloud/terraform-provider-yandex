package yandex

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/fatih/structs"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1"
	organizationmanagersdk "github.com/yandex-cloud/go-sdk/services/organizationmanager/v1"
)

func init() {
	resource.AddTestSweepers("yandex_organizationmanager_group", &resource.Sweeper{
		Name:         "yandex_organizationmanager_group",
		F:            testSweepGroups,
		Dependencies: []string{},
	})
}

func testSweepGroupOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(1 * time.Minute)
	defer cancel()
	client := organizationmanagersdk.NewGroupClient(conf.SDK)

	op, err := client.Delete(ctx, &organizationmanager.DeleteGroupRequest{
		GroupId: id,
	})

	return handleSweepOperationV2(ctx, op, err)
}

func testSweepGroups(_ string) error {
	return testSweepGroupsForOrganization(getExampleOrganizationID())
}

func testSweepGroupsForOrganization(organizationID string) error {
	if organizationID == "" {
		return nil
	}

	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &organizationmanager.ListGroupsRequest{
		OrganizationId: organizationID,
	}
	client := organizationmanagersdk.NewGroupClient(conf.SDK)

	it := client.Iterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepWithRetry(testSweepGroupOnce, conf, "Group", id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Group %q", id))
		}
	}

	return result.ErrorOrNil()
}

func TestSweepGroupsWithoutOrganizationIDIsNoop(t *testing.T) {
	t.Setenv("YC_TOKEN", "")
	t.Setenv("YC_SERVICE_ACCOUNT_KEY_FILE", "")

	if err := testSweepGroupsForOrganization(""); err != nil {
		t.Fatalf("sweeper without organization ID must be a no-op: %v", err)
	}
}

// Resource-level acceptance tests (TestAccOrganizationManagerGroup_basic, _Labels,
// _UpgradeFromSDKv2) live next to the framework resource at
// yandex-framework/gen/yandex/yandex_organizationmanager_group/resource_test.go.
// Helpers below remain because data source / iam_member / membership tests in
// this package still depend on them.

type GroupConfigGenerateFunc func(info *resourceGroupInfo) string

func testAccGroupRunTest(t *testing.T, fun GroupConfigGenerateFunc, rs bool, n int) {
	// Generate n groups, apply them to Terraform using fun and test according to resource type.
	for i := 0; i < n; i++ {
		info := newGroupInfo()
		var group organizationmanager.Group
		resource.Test(t, resource.TestCase{
			PreCheck:                 func() { testAccPreCheck(t) },
			ProtoV6ProviderFactories: testAccProviderFactoriesV6,
			CheckDestroy:             testAccCheckGroupDestroy,
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

func testAccCheckGroupDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)
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
