package yandex

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
)

// Test that an IAM binding can be applied to an organization
func TestAccOrganizationIamBinding_basic(t *testing.T) {
	organizationID := getExampleOrganizationID()
	role := "viewer"
	userID := getExampleUserID2()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			// Apply an IAM binding
			{
				Config: testAccOrganizationAssociateBindingBasic(organizationID, role, userID),
				Check: testAccCheckOrganizationIam(
					"yandex_organizationmanager_organization_iam_binding.acceptance",
					role,
					[]string{"userAccount:" + userID}),
			},
			organizationIamBindingImportStep("yandex_organizationmanager_organization_iam_binding.acceptance", organizationID, role),
		},
	})
}

// Test that an IAM binding can be applied to an organization with an existing binding.
func TestAccOrganizationIamBinding_existingBinding(t *testing.T) {
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
			// Apply an IAM binding, ensure previously added binding is respected by yandex_organizationmanager_organization_iam_binding.
			{
				Config: testAccOrganizationAssociateBindingBasic(organizationID, role, userID2),
				Check: testAccCheckOrganizationIam(
					"yandex_organizationmanager_organization_iam_binding.acceptance",
					role,
					[]string{"userAccount:" + userID, "userAccount:" + userID2}),
			},
		},
	})
}

// Test that multiple IAM bindings can be applied to an organization, one at a time
func TestAccOrganizationIamBinding_multiple(t *testing.T) {
	organizationID := getExampleOrganizationID()
	role1 := "editor"
	role2 := "viewer"
	userID1 := getExampleUserID1()
	userID2 := getExampleUserID2()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			// Apply an IAM binding
			{
				Config: testAccOrganizationAssociateBindingBasic(organizationID, role1, userID1),
				Check: testAccCheckOrganizationIam(
					"yandex_organizationmanager_organization_iam_binding.acceptance",
					role1,
					[]string{"userAccount:" + userID1}),
			},
			// Apply another IAM binding
			{
				Config: testAccOrganizationAssociateBindingMultiple(organizationID, role1, role2, userID1, userID2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationIam(
						"yandex_organizationmanager_organization_iam_binding.acceptance",
						role1,
						[]string{"userAccount:" + userID1, "userAccount:" + userID2}),
					testAccCheckOrganizationIam(
						"yandex_organizationmanager_organization_iam_binding.multiple",
						role2,
						[]string{"userAccount:" + userID1, "userAccount:" + userID2}),
				),
			},
			organizationIamBindingImportStep("yandex_organizationmanager_organization_iam_binding.acceptance", organizationID, role1),
			organizationIamBindingImportStep("yandex_organizationmanager_organization_iam_binding.multiple", organizationID, role2),
		},
	})
}

// Test that multiple IAM bindings can be applied to an organization all at once.
func TestAccOrganizationIamBinding_multipleAtOnce(t *testing.T) {
	organizationID := getExampleOrganizationID()
	role := "editor"
	role2 := "viewer"
	userID1 := getExampleUserID1()
	userID2 := getExampleUserID2()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			// Apply an IAM binding
			{
				Config: testAccOrganizationAssociateBindingMultiple(organizationID, role, role2, userID1, userID2),
			},
			organizationIamBindingImportStep(
				"yandex_organizationmanager_organization_iam_binding.acceptance",
				organizationID,
				role),
			organizationIamBindingImportStep("yandex_organizationmanager_organization_iam_binding.multiple",
				organizationID,
				role2),
		},
	})
}

// Test that an IAM binding can be updated once applied to an organization.
func TestAccOrganizationIamBinding_update(t *testing.T) {
	organizationID := getExampleOrganizationID()
	role := "editor"
	userID1 := getExampleUserID1()
	userID2 := getExampleUserID2()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			// Apply an IAM binding
			{
				Config: testAccOrganizationAssociateBindingBasic(organizationID, role, userID1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationIam(
						"yandex_organizationmanager_organization_iam_binding.acceptance",
						role,
						[]string{"userAccount:" + userID1})),
			},
			organizationIamBindingImportStep("yandex_organizationmanager_organization_iam_binding.acceptance", organizationID, role),
			// Apply an updated IAM binding
			{
				Config: testAccOrganizationAssociateBindingUpdated(organizationID, role, userID1, userID2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationIam(
						"yandex_organizationmanager_organization_iam_binding.acceptance",
						role,
						[]string{"userAccount:" + userID1, "userAccount:" + userID2})),
			},
			organizationIamBindingImportStep("yandex_organizationmanager_organization_iam_binding.acceptance", organizationID, role),
			// Drop the original member
			{
				Config: testAccOrganizationAssociateBindingDropMemberFromBasic(organizationID, role, userID1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationIam(
						"yandex_organizationmanager_organization_iam_binding.acceptance",
						role,
						[]string{"userAccount:" + userID1})),
			},
			organizationIamBindingImportStep("yandex_organizationmanager_organization_iam_binding.acceptance", organizationID, role),
		},
	})
}

// Test that an IAM binding can be removed from an organization.
func TestAccOrganizationIamBinding_remove(t *testing.T) {
	cloudID := getExampleCloudID()
	organizationID := getExampleOrganizationID()
	role1 := "editor"
	role2 := "viewer"
	userID1 := getExampleUserID1()
	userID2 := getExampleUserID2()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			// Prepare data source about cloud ID
			{
				Config: testAccCheckResourceManagerCloud_byID(cloudID),
			},
			// Apply multiple IAM bindings
			{
				Config: testAccOrganizationAssociateBindingMultiple(organizationID, role1, role2, userID1, userID2),
			},
			organizationIamBindingImportStep("yandex_organizationmanager_organization_iam_binding.acceptance", organizationID, role1),
			organizationIamBindingImportStep("yandex_organizationmanager_organization_iam_binding.multiple", organizationID, role2),
			// Remove the bindings
			{
				Config: testAccCheckResourceManagerCloud_byID(cloudID),
			},
		},
	})
}

func testAccCheckOrganizationIam(resourceName, role string, members []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(*Config)

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("can't find %s in state", resourceName)
		}

		organizationID := strings.SplitN(rs.Primary.ID, "/", 2)[0]

		bindings, err := getOrganizationAccessBindings(config.Context(), config, organizationID)
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

		if reflect.DeepEqual(members, roleMembers) {
			return nil
		}

		return fmt.Errorf("Binding found but expected members is %v, got %v", members, roleMembers)
	}
}

func testAccOrganizationAssociateBindingBasic(organizationID, role, userID string) string {
	return fmt.Sprintf(`
resource "yandex_organizationmanager_organization_iam_binding" "acceptance" {
  organization_id = "%s"
  role            = "%s"
  members         = ["userAccount:%s"]
}
`, organizationID, role, userID)
}

func testAccOrganizationAssociateBindingUpdated(organizationID, role, userID1, userID2 string) string {
	return fmt.Sprintf(`
resource "yandex_organizationmanager_organization_iam_binding" "acceptance" {
  organization_id = "%s"
  role            = "%s"
  members         = ["userAccount:%s", "userAccount:%s"]
}
`, organizationID, role, userID1, userID2)
}

func testAccOrganizationAssociateBindingMultiple(organizationID, role1, role2, userID1, userID2 string) string {
	multiple1 := fmt.Sprintf(`
resource "yandex_organizationmanager_organization_iam_binding" "acceptance" {
  organization_id = "%s"
  role            = "%s"
  members         = ["userAccount:%s", "userAccount:%s"]
}
`, organizationID, role1, userID1, userID2)

	multiple2 := fmt.Sprintf(`
resource "yandex_organizationmanager_organization_iam_binding" "multiple" {
  organization_id = "%s"
  role            = "%s"
  members         = ["userAccount:%s", "userAccount:%s"]
}
`, organizationID, role2, userID1, userID2)

	return multiple1 + multiple2
}

func organizationIamBindingImportStep(resourceName, organizationID, role string) resource.TestStep {
	return resource.TestStep{
		ResourceName:      resourceName,
		ImportStateId:     fmt.Sprintf("%s %s", organizationID, role),
		ImportState:       true,
		ImportStateVerify: true,
	}
}

func testAccOrganizationAssociateBindingDropMemberFromBasic(cloudID, role, userID string) string {
	return fmt.Sprintf(`
resource "yandex_organizationmanager_organization_iam_binding" "acceptance" {
  organization_id = "%s"
  role            = "%s"
  members         = ["userAccount:%s"]
}
`, cloudID, role, userID)
}

func testAccOrganizationAddAccessBinding(organizationID, role, userID string) error {
	config := testAccProvider.Meta().(*Config)
	ctx, cancel := context.WithTimeout(config.Context(), yandexOrganizationManagerOrganizationDefaultTimeout)
	defer cancel()
	op, err := config.sdk.WrapOperation(config.sdk.OrganizationManager().Organization().UpdateAccessBindings(config.Context(), &access.UpdateAccessBindingsRequest{
		ResourceId: organizationID,
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

func testAccOrganizationRemoveAccessBinding(organizationID, role, userID string) error {
	config := testAccProvider.Meta().(*Config)
	ctx, cancel := context.WithTimeout(config.Context(), yandexOrganizationManagerOrganizationDefaultTimeout)
	defer cancel()
	op, err := config.sdk.WrapOperation(config.sdk.OrganizationManager().Organization().UpdateAccessBindings(config.Context(), &access.UpdateAccessBindingsRequest{
		ResourceId: organizationID,
		AccessBindingDeltas: []*access.AccessBindingDelta{
			{
				Action: access.AccessBindingAction_REMOVE,
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
