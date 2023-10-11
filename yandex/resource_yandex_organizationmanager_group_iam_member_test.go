package yandex

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1"
)

// Test that an IAM member can be applied to a Group.
func TestAccOrganizationManagerGroupIamMember_basic(t *testing.T) {
	t.Parallel()
	role := "viewer"
	userID := getExampleUserID1()

	groupInfo := newGroupInfo()
	groupResourceName := groupInfo.getResourceName(true)
	var group organizationmanager.Group

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			// Apply an IAM member
			{
				Config: testAccOrganizationManagerGroup(groupInfo) +
					testAccGroupAssociateMemberBasic(groupResourceName, role, userID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(groupResourceName, &group),
					testAccCheckGroupIam(
						&group,
						role,
						[]string{"userAccount:" + userID})),
			},
			groupIamMemberImportStep(
				"yandex_organizationmanager_group_iam_member.acceptance",
				&group,
				role,
				userID),
		},
	})
}

// Test that an IAM member can work with existing bindings.
func TestAccOrganizationManagerGroupIamMember_existingBinding(t *testing.T) {
	t.Parallel()
	role := "viewer"
	userID := getExampleUserID1()
	userID2 := getExampleUserID2()

	groupInfo := newGroupInfo()
	groupResourceName := groupInfo.getResourceName(true)
	var group organizationmanager.Group

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			// Add an access binding for the first user.
			{
				Config: testAccOrganizationManagerGroup(groupInfo),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(groupResourceName, &group),
					func(state *terraform.State) error {
						return testAccGroupAddAccessBinding(group.Id, role, userID)
					},
					testAccCheckGroupIam(
						&group,
						role,
						[]string{
							"userAccount:" + userID,
						}),
				),
			},
			// Apply an IAM binding, ensure previously added binding is respected by yandex_organizationmanager_group_iam_member.
			{
				Config: testAccOrganizationManagerGroup(groupInfo) +
					testAccGroupAssociateMemberBasic(groupResourceName, role, userID2),
				Check: testAccCheckGroupIam(
					&group,
					role,
					[]string{
						"userAccount:" + userID,
						"userAccount:" + userID2,
					}),
			},
			// Remove an access binding for the first user as it will not be removed by the
			// yandex_organizationmanager_group_iam_member in a way it is done by
			// yandex_organizationmanager_group_iam_binding resource.
			{
				Config: testAccOrganizationManagerGroup(groupInfo),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupIam(
						&group,
						role,
						[]string{
							"userAccount:" + userID,
						}),
				),
			},
		},
	})
}

// Test that an IAM member can change role.
func TestAccOrganizationManagerGroupIamMember_changeRole(t *testing.T) {
	t.Parallel()
	role1 := "viewer"
	role2 := "editor"
	userID := getExampleUserID1()

	groupInfo := newGroupInfo()
	groupResourceName := groupInfo.getResourceName(true)
	var group organizationmanager.Group

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			// Apply an IAM member
			{
				Config: testAccOrganizationManagerGroup(groupInfo) +
					testAccGroupAssociateMemberBasic(groupResourceName, role1, userID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(groupResourceName, &group),
					testAccCheckGroupIam(
						&group,
						role1,
						[]string{"userAccount:" + userID})),
			},
			// Apply an IAM member
			{
				Config: testAccOrganizationManagerGroup(groupInfo) +
					testAccGroupAssociateMemberBasic(groupResourceName, role2, userID),
				Check: testAccCheckGroupIam(
					&group,
					role2,
					[]string{"userAccount:" + userID}),
			},
			groupIamMemberImportStep(
				"yandex_organizationmanager_group_iam_member.acceptance",
				&group,
				role2,
				userID),
		},
	})
}

// Test that an IAM member can change users.
func TestAccOrganizationManagerGroupIamMember_changeUser(t *testing.T) {
	t.Parallel()
	role := "viewer"
	userID1 := getExampleUserID1()
	userID2 := getExampleUserID2()

	groupInfo := newGroupInfo()
	groupResourceName := groupInfo.getResourceName(true)
	var group organizationmanager.Group

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			// Apply an IAM member
			{
				Config: testAccOrganizationManagerGroup(groupInfo) +
					testAccGroupAssociateMemberBasic(groupResourceName, role, userID1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(groupResourceName, &group),
					testAccCheckGroupIam(
						&group,
						role,
						[]string{
							"userAccount:" + userID1,
						})),
			},
			// Apply an IAM member
			{
				Config: testAccOrganizationManagerGroup(groupInfo) +
					testAccGroupAssociateMemberBasic(groupInfo.getResourceName(true), role, userID2),
				Check: testAccCheckGroupIam(
					&group,
					role,
					[]string{"userAccount:" + userID2}),
			},
			groupIamMemberImportStep(
				"yandex_organizationmanager_group_iam_member.acceptance",
				&group,
				role,
				userID2),
		},
	})
}

// Test that an IAM member can be applied to two users in one Group.
func TestAccOrganizationManagerGroupIamMember_separateMembers(t *testing.T) {
	t.Parallel()
	role := "viewer"
	userID1 := getExampleUserID1()
	userID2 := getExampleUserID2()

	groupInfo := newGroupInfo()
	groupResourceName := groupInfo.getResourceName(true)
	var group organizationmanager.Group

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			// Apply two IAM members. Check that both records are visible.
			{
				Config: testAccOrganizationManagerGroup(groupInfo) +
					testAccGroupTwoIamMembers(
						groupResourceName,
						role,
						userID1,
						userID2,
						"acceptance1",
						"acceptance2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(groupResourceName, &group),
					testAccCheckGroupIam(
						&group,
						role,
						[]string{
							"userAccount:" + userID1,
							"userAccount:" + userID2,
						})),
			},
			groupIamMemberImportStep(
				"yandex_organizationmanager_group_iam_member.acceptance1",
				&group,
				role,
				userID1),
			groupIamMemberImportStep(
				"yandex_organizationmanager_group_iam_member.acceptance2",
				&group,
				role,
				userID2),
		},
	})
}

func testAccGroupAssociateMemberBasic(groupResourceName, role, userID string) string {
	return fmt.Sprintf(`
resource "yandex_organizationmanager_group_iam_member" "acceptance" {
  group_id = %s.id
  role     = "%s"
  member   = "userAccount:%s"
}
`, groupResourceName, role, userID)
}

func testAccGroupTwoIamMembers(groupResourceName, role, userID1, userID2, member1, member2 string) string {
	return fmt.Sprintf(`
resource "yandex_organizationmanager_group_iam_member" "%[1]s" {
  group_id = %[3]s.id
  role     = "%[4]s"
  member   = "userAccount:%[5]s"
}
resource "yandex_organizationmanager_group_iam_member" "%[2]s" {
  group_id = %[3]s.id
  role     = "%[4]s"
  member   = "userAccount:%[6]s"
}
`, member1, member2, groupResourceName, role, userID1, userID2)
}

func groupIamMemberImportStep(resourceName string, group *organizationmanager.Group, role, userID string) resource.TestStep {
	return resource.TestStep{
		ResourceName: resourceName,
		ImportStateIdFunc: func(s *terraform.State) (string, error) {
			return group.Id + " " + role + " " + "userAccount:" + userID, nil
		},
		ImportState:       true,
		ImportStateVerify: true,
	}
}

func testAccCheckGroupIam(group *organizationmanager.Group, role string, members []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(*Config)
		bindings, err := getGroupAccessBindings(config.Context(), config, group.Id)
		if err != nil {
			return err
		}

		var roleMembers []string
		for _, binding := range bindings {
			if binding.RoleId == role {
				member := binding.Subject.Type + ":" + binding.Subject.Id
				roleMembers = append(roleMembers, member)
			}
		}
		sort.Strings(members)
		sort.Strings(roleMembers)

		if len(members) == 0 && len(roleMembers) == 0 {
			return nil
		}

		if reflect.DeepEqual(members, roleMembers) {
			return nil
		}

		return fmt.Errorf("Binding found but expected members is %v, got %v", members, roleMembers)
	}
}

func testAccGroupAddAccessBinding(groupID, role, userID string) error {
	config := testAccProvider.Meta().(*Config)
	ctx, cancel := context.WithTimeout(config.Context(), yandexOrganizationManagerIAMGroupDefaultTimeout)
	defer cancel()
	op, err := config.sdk.WrapOperation(config.sdk.OrganizationManager().Group().UpdateAccessBindings(config.Context(), &access.UpdateAccessBindingsRequest{
		ResourceId: groupID,
		AccessBindingDeltas: []*access.AccessBindingDelta{
			{
				Action: access.AccessBindingAction_ADD,
				AccessBinding: &access.AccessBinding{
					RoleId: role,
					Subject: &access.Subject{
						Id:   userID,
						Type: "userAccount",
					},
				},
			},
		},
	}))
	if err != nil {
		return err
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}
	return nil
}
