package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// Test that an IAM member can be applied to an organization.
func TestAccOrganizationIamMember_basic(t *testing.T) {
	organizationID := getExampleOrganizationID()
	role := "viewer"
	userID := getExampleUserID1()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			// Apply an IAM member
			{
				Config: testAccOrganizationAssociateMemberBasic(organizationID, role, userID),
				Check: testAccCheckOrganizationIam(
					"yandex_organizationmanager_organization_iam_member.acceptance",
					role,
					[]string{"userAccount:" + userID}),
			},
			organizationIamMemberImportStep(
				"yandex_organizationmanager_organization_iam_member.acceptance",
				organizationID,
				role,
				userID),
		},
	})
}

// Test that an IAM member can work with existing bindings.
func TestAccOrganizationIamMember_existingBinding(t *testing.T) {
	organizationID := getExampleOrganizationID()
	cloudID := getExampleCloudID()
	role := "viewer"
	userID := getExampleUserID1()
	userID2 := getExampleUserID2()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			// Add an access binding for the first user.
			{
				Config: testAccCheckResourceManagerCloud_byID(cloudID),
				Check: func(state *terraform.State) error {
					return testAccOrganizationAddAccessBinding(organizationID, role, userID)
				},
			},
			// Apply an IAM binding, ensure previously added binding is respected by yandex_organizationmanager_organization_iam_member.
			{
				Config: testAccOrganizationAssociateMemberBasic(organizationID, role, userID2),
				Check: testAccCheckOrganizationIam(
					"yandex_organizationmanager_organization_iam_member.acceptance",
					role,
					[]string{"userAccount:" + userID, "userAccount:" + userID2}),
			},
			// Remove an access binding for the first user as it will not be removed by the
			// yandex_organizationmanager_organization_iam_member in a way it is done by
			// yandex_organizationmanager_organization_iam_binding resource.
			{
				Config: testAccOrganizationAssociateMemberBasic(organizationID, role, userID2),
				Check: func(state *terraform.State) error {
					return testAccOrganizationRemoveAccessBinding(organizationID, role, userID)
				},
			},
			organizationIamMemberImportStep(
				"yandex_organizationmanager_organization_iam_member.acceptance",
				organizationID,
				role,
				userID2),
		},
	})
}

// Test that an IAM member can change role.
func TestAccOrganizationIamMember_changeRole(t *testing.T) {
	organizationID := getExampleOrganizationID()
	role1 := "viewer"
	role2 := "editor"
	userID := getExampleUserID1()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			// Apply an IAM member
			{
				Config: testAccOrganizationAssociateMemberBasic(organizationID, role1, userID),
				Check: testAccCheckOrganizationIam(
					"yandex_organizationmanager_organization_iam_member.acceptance",
					role1,
					[]string{"userAccount:" + userID}),
			},
			// Apply an IAM member
			{
				Config: testAccOrganizationAssociateMemberBasic(organizationID, role2, userID),
				Check: testAccCheckOrganizationIam(
					"yandex_organizationmanager_organization_iam_member.acceptance",
					role2,
					[]string{"userAccount:" + userID}),
			},
			organizationIamMemberImportStep(
				"yandex_organizationmanager_organization_iam_member.acceptance",
				organizationID,
				role2,
				userID),
		},
	})
}

// Test that an IAM member can change users.
func TestAccOrganizationIamMember_changeUser(t *testing.T) {
	organizationID := getExampleOrganizationID()
	role := "viewer"
	userID1 := getExampleUserID1()
	userID2 := getExampleUserID2()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			// Apply an IAM member
			{
				Config: testAccOrganizationAssociateMemberBasic(organizationID, role, userID1),
				Check: testAccCheckOrganizationIam(
					"yandex_organizationmanager_organization_iam_member.acceptance",
					role,
					[]string{"userAccount:" + userID1}),
			},
			// Apply an IAM member
			{
				Config: testAccOrganizationAssociateMemberBasic(organizationID, role, userID2),
				Check: testAccCheckOrganizationIam(
					"yandex_organizationmanager_organization_iam_member.acceptance",
					role,
					[]string{"userAccount:" + userID2}),
			},
			organizationIamMemberImportStep(
				"yandex_organizationmanager_organization_iam_member.acceptance",
				organizationID,
				role,
				userID2),
		},
	})
}

// Test that an IAM member can be applied to two users in one organization.
func TestAccOrganizationIamMember_separateMembers(t *testing.T) {
	organizationID := getExampleOrganizationID()
	role := "viewer"
	userID1 := getExampleUserID1()
	userID2 := getExampleUserID2()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			// Apply two IAM members. Check that both records are visible.
			{
				Config: testAccOrganizationTwoIamMembers(organizationID, role, userID1, userID2, "acceptance1", "acceptance2"),
				Check: testAccCheckOrganizationIam(
					"yandex_organizationmanager_organization_iam_member.acceptance1",
					role,
					[]string{"userAccount:" + userID1, "userAccount:" + userID2}),
			},
			organizationIamMemberImportStep(
				"yandex_organizationmanager_organization_iam_member.acceptance1",
				organizationID,
				role,
				userID1),
			organizationIamMemberImportStep(
				"yandex_organizationmanager_organization_iam_member.acceptance2",
				organizationID,
				role,
				userID2),
		},
	})
}

func testAccOrganizationAssociateMemberBasic(organizationID, role, userID string) string {
	return fmt.Sprintf(`
resource "yandex_organizationmanager_organization_iam_member" "acceptance" {
  organization_id = "%s"
  role            = "%s"
  member          = "userAccount:%s"
}
`, organizationID, role, userID)
}

func testAccOrganizationTwoIamMembers(organizationID, role, userID1, userID2, member1, member2 string) string {
	return fmt.Sprintf(`
resource "yandex_organizationmanager_organization_iam_member" "%[1]s" {
  organization_id = "%[3]s"
  role            = "%[4]s"
  member          = "userAccount:%[5]s"
}
resource "yandex_organizationmanager_organization_iam_member" "%[2]s" {
  organization_id = "%[3]s"
  role            = "%[4]s"
  member          = "userAccount:%[6]s"
}
`, member1, member2, organizationID, role, userID1, userID2)
}

func organizationIamMemberImportStep(resourceName, organizationID, role, userID string) resource.TestStep {
	return resource.TestStep{
		ResourceName:      resourceName,
		ImportStateId:     fmt.Sprintf("%s %s %s", organizationID, role, "userAccount:"+userID),
		ImportState:       true,
		ImportStateVerify: true,
	}
}
