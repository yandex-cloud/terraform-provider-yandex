package acceptance

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/acceptanceenv"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1"
	organizationmanagersdk "github.com/yandex-cloud/go-sdk/services/organizationmanager/v1"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
	svc_organization_manager "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/organizationmanager/pluginsdk"
)

// Test that a Group Membership can be applied to an organization.
func TestAccOrganizationManagerGroupMembership_basic(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()
	userID := acceptanceenv.UserID1()

	groupInfo := newGroupInfo()
	groupResourceName := groupInfo.getResourceName(true)
	var group organizationmanager.Group

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		Steps: []resource.TestStep{
			// Apply a group membership
			{
				Config: testAccOrganizationManagerGroup(groupInfo) +
					testAccGroupMembership(groupResourceName, "membership", userID),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckGroupExists(groupResourceName, &group),
					suite.testAccCheckGroupMembership(
						&group,
						[]string{
							"userAccount:" + userID,
						}),
				),
			},
		},
	})
}

// Test that a Group Membership can work with existing members.
func TestAccOrganizationManagerGroupMembership_existingMember(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()
	userID := acceptanceenv.UserID1()
	userID2 := acceptanceenv.UserID2()

	groupInfo := newGroupInfo()
	groupResourceName := groupInfo.getResourceName(true)
	var group organizationmanager.Group

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		Steps: []resource.TestStep{
			// Add a member for the first user.
			{
				Config: testAccOrganizationManagerGroup(groupInfo),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckGroupExists(groupResourceName, &group),
					suite.testAccOrganizationManagerGroupMembershipAddMember(&group, userID)),
			},
			// Apply a Group Membership, ensure previously added membership is respected by yandex_organizationmanager_group_membership.
			{
				Config: testAccOrganizationManagerGroup(groupInfo) +
					testAccGroupMembership(groupResourceName, "membership", userID2),
				Check: suite.testAccCheckGroupMembership(
					&group,
					[]string{
						"userAccount:" + userID,
						"userAccount:" + userID2,
					}),
			},
			// Destroy memberships by passing in a config with a Group only, ensure that the first membership is still there.
			{
				Config: testAccOrganizationManagerGroup(groupInfo),
				Check: suite.testAccCheckGroupMembership(
					&group,
					[]string{
						"userAccount:" + userID,
					}),
			},
			// Do not delete membership explicitly as it will be deleted with the group.
		},
	})
}

// Test that a Group Membership can remove users from membership
func TestAccOrganizationManagerGroupMembership_deleteUser(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()
	userID1 := acceptanceenv.UserID1()
	userID2 := acceptanceenv.UserID2()

	groupInfo := newGroupInfo()
	groupResourceName := groupInfo.getResourceName(true)
	var group organizationmanager.Group

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		Steps: []resource.TestStep{
			// Apply a group membership with 2 users.
			{
				Config: testAccOrganizationManagerGroup(groupInfo) +
					testAccGroupMembership(groupResourceName, "membership", userID1, userID2),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckGroupExists(groupResourceName, &group),
					suite.testAccCheckGroupMembership(
						&group,
						[]string{
							"userAccount:" + userID1,
							"userAccount:" + userID2,
						}),
				),
			},
			// Apply a group membership with second user deleted. Ensure the user gets deleted from membership in the cloud.
			{
				Config: testAccOrganizationManagerGroup(groupInfo) +
					testAccGroupMembership(groupResourceName, "membership", userID1),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckGroupMembership(
						&group,
						[]string{
							"userAccount:" + userID1,
						}),
				),
			},
		},
	})
}

// Test that a Group Membership can add users to an existing membership
func TestAccOrganizationManagerGroupMembership_addUser(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()
	userID1 := acceptanceenv.UserID1()
	userID2 := acceptanceenv.UserID2()

	groupInfo := newGroupInfo()
	groupResourceName := groupInfo.getResourceName(true)
	var group organizationmanager.Group

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		Steps: []resource.TestStep{
			// Apply a group membership with one user.
			{
				Config: testAccOrganizationManagerGroup(groupInfo) +
					testAccGroupMembership(groupResourceName, "membership", userID1),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckGroupExists(groupResourceName, &group),
					suite.testAccCheckGroupMembership(
						&group,
						[]string{
							"userAccount:" + userID1,
						}),
				),
			},
			// Apply a group membership with 2 users.
			{
				Config: testAccOrganizationManagerGroup(groupInfo) +
					testAccGroupMembership(groupResourceName, "membership", userID1, userID2),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckGroupMembership(
						&group,
						[]string{
							"userAccount:" + userID1,
							"userAccount:" + userID2,
						}),
				),
			},
		},
	})
}

// Test that a Group Membership restores the config when a member is deleted from outside of Terraform.
func TestAccOrganizationManagerGroupMembership_restoresDeleted(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()
	userID1 := acceptanceenv.UserID1()
	userID2 := acceptanceenv.UserID2()

	groupInfo := newGroupInfo()
	groupResourceName := groupInfo.getResourceName(true)
	var group organizationmanager.Group

	config := testAccOrganizationManagerGroup(groupInfo) +
		testAccGroupMembership(groupResourceName, "membership", userID1, userID2)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		Steps: []resource.TestStep{
			// Apply a Group Membership with 2 users, then delete the first user from the group and ensure there is
			// only the second user in the group.
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckGroupExists(groupResourceName, &group),
					suite.testAccCheckGroupMembership(
						&group,
						[]string{
							"userAccount:" + userID1,
							"userAccount:" + userID2,
						}),
					suite.testAccOrganizationManagerGroupMembershipRemoveMember(&group, userID1),
					suite.testAccCheckGroupMembership(
						&group,
						[]string{
							"userAccount:" + userID2,
						}),
				),
				ExpectNonEmptyPlan: true,
			},
			// Ensure the same config restores the deleted user.
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckGroupMembership(
						&group,
						[]string{
							"userAccount:" + userID1,
							"userAccount:" + userID2,
						}),
				),
			},
		},
	})
}

// Test that a Group Membership can change users.
func TestAccOrganizationManagerGroupMembership_changeUser(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()
	userID1 := acceptanceenv.UserID1()
	userID2 := acceptanceenv.UserID2()

	groupInfo := newGroupInfo()
	groupResourceName := groupInfo.getResourceName(true)
	var group organizationmanager.Group

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		Steps: []resource.TestStep{
			// Apply a Group Membership
			{
				Config: testAccOrganizationManagerGroup(groupInfo) +
					testAccGroupMembership(groupResourceName, "membership", userID1),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckGroupExists(groupResourceName, &group),
					suite.testAccCheckGroupMembership(
						&group,
						[]string{
							"userAccount:" + userID1,
						}),
				),
			},
			// Apply another Group Membership
			{
				Config: testAccOrganizationManagerGroup(groupInfo) +
					testAccGroupMembership(groupResourceName, "membership", userID2),
				Check: suite.testAccCheckGroupMembership(
					&group,
					[]string{
						"userAccount:" + userID2,
					}),
			},
		},
	})
}

func testAccGroupMembership(groupResourceName, name string, users ...string) string {
	members := strings.Join(users, "\", \"")
	return fmt.Sprintf(`
resource "yandex_organizationmanager_group_membership" %s {
  group_id = %s.id
  members  = ["%s"]
}
`, name, groupResourceName, members)
}

func (suite testSuite) testAccCheckGroupMembership(group *organizationmanager.Group, members []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ctx, cancel := context.WithTimeout(context.Background(), svc_organization_manager.YandexOrganizationManagerGroupMembershipDefaultTimeout)
		defer cancel()

		config, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}

		cloudMembers, err := svc_organization_manager.GetGroupMembers(ctx, config, group.Id)
		if err != nil {
			return err
		}

		var groupMembers []string
		for _, m := range cloudMembers {
			member := m.SubjectType + ":" + m.SubjectId
			groupMembers = append(groupMembers, member)
		}
		sort.Strings(members)
		sort.Strings(groupMembers)

		if reflect.DeepEqual(members, groupMembers) {
			return nil
		}

		return fmt.Errorf("Members found but expected members is %v, got %v", members, groupMembers)
	}
}

func (suite testSuite) testAccOrganizationManagerGroupMembershipAddMember(group *organizationmanager.Group, userID string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		req := organizationmanager.UpdateGroupMembersRequest{
			GroupId: group.Id,
			MemberDeltas: []*organizationmanager.MemberDelta{
				{
					Action:    organizationmanager.MemberDelta_ADD,
					SubjectId: userID,
				},
			},
		}

		config, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}

		client := organizationmanagersdk.NewGroupClient(config.SDK)

		ctx, cancel := context.WithTimeout(config.Context(), svc_organization_manager.YandexOrganizationManagerGroupMembershipDefaultTimeout)
		defer cancel()

		op, err := client.UpdateMembers(ctx, &req)
		if err != nil {
			return fmt.Errorf("Error while requesting API to create GroupMembership: %s", err)
		}

		_ = op.Metadata()
		_, err = op.Wait(ctx)
		if err != nil {
			return fmt.Errorf("Error while waiting operation to create GroupMembership: %s", err)
		}

		return nil
	}
}

func (suite testSuite) testAccOrganizationManagerGroupMembershipRemoveMember(group *organizationmanager.Group, userID string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		req := organizationmanager.UpdateGroupMembersRequest{
			GroupId: group.Id,
			MemberDeltas: []*organizationmanager.MemberDelta{
				{
					Action:    organizationmanager.MemberDelta_REMOVE,
					SubjectId: userID,
				},
			},
		}

		config, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}

		client := organizationmanagersdk.NewGroupClient(config.SDK)

		ctx, cancel := context.WithTimeout(config.Context(), svc_organization_manager.YandexOrganizationManagerGroupMembershipDefaultTimeout)
		defer cancel()

		op, err := client.UpdateMembers(ctx, &req)
		if err != nil {
			return fmt.Errorf("Error while requesting API to create GroupMembership: %s", err)
		}

		_ = op.Metadata()
		_, err = op.Wait(ctx)
		if err != nil {
			return fmt.Errorf("Error while waiting operation to create GroupMembership: %s", err)
		}

		return nil
	}
}
