package acceptance

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/acceptanceenv"
	rmfixtures "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/resourcemanager/testfixtures"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
	organizationmanagersdk "github.com/yandex-cloud/go-sdk/services/organizationmanager/v1"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
	svc_organization_manager "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/organizationmanager/pluginsdk"
)

// Test that an IAM binding can be applied to an organization
func TestAccOrganizationIamBinding_basic(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	organizationID := acceptanceenv.OrganizationID()
	role := "viewer"
	userID := acceptanceenv.UserID2()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		Steps: []resource.TestStep{
			// Apply an IAM binding
			{
				Config: testAccOrganizationAssociateBindingBasic(organizationID, role, userID),
				Check: suite.testAccCheckOrganizationIam(
					organizationID,
					role,
					[]string{"userAccount:" + userID}),
			},
			organizationIamBindingImportStep("yandex_organizationmanager_organization_iam_binding.acceptance", organizationID, role),
		},
	})
}

// Test that an IAM binding can be applied to an organization with an existing binding.
func TestAccOrganizationIamBinding_existingBinding(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	organizationID := acceptanceenv.OrganizationID()
	cloudID := acceptanceenv.CloudID()
	role := "viewer"
	userID := acceptanceenv.UserID1()
	userID2 := acceptanceenv.UserID2()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		Steps: []resource.TestStep{
			// Add an access binding for the first user.
			{
				Config: rmfixtures.CloudDataSourceByID(cloudID),
				Check: func(state *terraform.State) error {
					return suite.testAccOrganizationAddAccessBinding(organizationID, role, userID)
				},
			},
			// Apply an IAM binding, ensure previously added binding is respected by yandex_organizationmanager_organization_iam_binding.
			{
				Config: testAccOrganizationAssociateBindingBasic(organizationID, role, userID2),
				Check: suite.testAccCheckOrganizationIam(
					organizationID,
					role,
					[]string{"userAccount:" + userID, "userAccount:" + userID2}),
			},
		},
	})
}

// Test that multiple IAM bindings can be applied to an organization, one at a time
func TestAccOrganizationIamBinding_multiple(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	organizationID := acceptanceenv.OrganizationID()
	role1 := "editor"
	role2 := "viewer"
	userID1 := acceptanceenv.UserID1()
	userID2 := acceptanceenv.UserID2()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		Steps: []resource.TestStep{
			// Apply an IAM binding
			{
				Config: testAccOrganizationAssociateBindingBasic(organizationID, role1, userID1),
				Check: suite.testAccCheckOrganizationIam(
					organizationID,
					role1,
					[]string{"userAccount:" + userID1}),
			},
			// Apply another IAM binding
			{
				Config: testAccOrganizationAssociateBindingMultiple(organizationID, role1, role2, userID1, userID2),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckOrganizationIam(
						organizationID,
						role1,
						[]string{"userAccount:" + userID1, "userAccount:" + userID2}),
					suite.testAccCheckOrganizationIam(
						organizationID,
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
	suite := testSuite{acceptance.New(t)}
	organizationID := acceptanceenv.OrganizationID()
	role := "editor"
	role2 := "viewer"
	userID1 := acceptanceenv.UserID1()
	userID2 := acceptanceenv.UserID2()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
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
	suite := testSuite{acceptance.New(t)}
	organizationID := acceptanceenv.OrganizationID()
	role := "editor"
	userID1 := acceptanceenv.UserID1()
	userID2 := acceptanceenv.UserID2()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		Steps: []resource.TestStep{
			// Apply an IAM binding
			{
				Config: testAccOrganizationAssociateBindingBasic(organizationID, role, userID1),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckOrganizationIam(
						organizationID,
						role,
						[]string{"userAccount:" + userID1})),
			},
			organizationIamBindingImportStep("yandex_organizationmanager_organization_iam_binding.acceptance", organizationID, role),
			// Apply an updated IAM binding
			{
				Config: testAccOrganizationAssociateBindingUpdated(organizationID, role, userID1, userID2),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckOrganizationIam(
						organizationID,
						role,
						[]string{"userAccount:" + userID1, "userAccount:" + userID2})),
			},
			organizationIamBindingImportStep("yandex_organizationmanager_organization_iam_binding.acceptance", organizationID, role),
			// Drop the original member
			{
				Config: testAccOrganizationAssociateBindingDropMemberFromBasic(organizationID, role, userID1),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckOrganizationIam(
						organizationID,
						role,
						[]string{"userAccount:" + userID1})),
			},
			organizationIamBindingImportStep("yandex_organizationmanager_organization_iam_binding.acceptance", organizationID, role),
		},
	})
}

// Test that an IAM binding can be removed from an organization.
func TestAccOrganizationIamBinding_remove(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	cloudID := acceptanceenv.CloudID()
	organizationID := acceptanceenv.OrganizationID()
	role1 := "editor"
	role2 := "viewer"
	userID1 := acceptanceenv.UserID1()
	userID2 := acceptanceenv.UserID2()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		Steps: []resource.TestStep{
			// Prepare data source about cloud ID
			{
				Config: rmfixtures.CloudDataSourceByID(cloudID),
			},
			// Apply multiple IAM bindings
			{
				Config: testAccOrganizationAssociateBindingMultiple(organizationID, role1, role2, userID1, userID2),
			},
			organizationIamBindingImportStep("yandex_organizationmanager_organization_iam_binding.acceptance", organizationID, role1),
			organizationIamBindingImportStep("yandex_organizationmanager_organization_iam_binding.multiple", organizationID, role2),
			// Remove the bindings
			{
				Config: rmfixtures.CloudDataSourceByID(cloudID),
			},
		},
	})
}

func (suite testSuite) testAccCheckOrganizationIam(organizationID, role string, members []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}

		bindings, err := svc_organization_manager.GetOrganizationAccessBindings(config.Context(), config, organizationID)
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
		ResourceName:                         resourceName,
		ImportStateId:                        fmt.Sprintf("%s,%s", organizationID, role),
		ImportState:                          true,
		ImportStateVerifyIdentifierAttribute: "organization_id",
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

func (suite testSuite) testAccOrganizationAddAccessBinding(organizationID, role, userID string) error {
	config, configErr := suite.Config()
	if configErr != nil {
		return configErr
	}

	client := organizationmanagersdk.NewOrganizationClient(config.SDK)

	ctx, cancel := context.WithTimeout(config.Context(), svc_organization_manager.YandexOrganizationManagerOrganizationDefaultTimeout)
	defer cancel()
	op, err := client.UpdateAccessBindings(config.Context(), &access.UpdateAccessBindingsRequest{
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
	})
	if err != nil {
		return err
	}

	_, err = op.Wait(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (suite testSuite) testAccOrganizationRemoveAccessBinding(organizationID, role, userID string) error {
	config, configErr := suite.Config()
	if configErr != nil {
		return configErr
	}

	client := organizationmanagersdk.NewOrganizationClient(config.SDK)

	ctx, cancel := context.WithTimeout(config.Context(), svc_organization_manager.YandexOrganizationManagerOrganizationDefaultTimeout)
	defer cancel()
	op, err := client.UpdateAccessBindings(config.Context(), &access.UpdateAccessBindingsRequest{
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
	})
	if err != nil {
		return err
	}

	_, err = op.Wait(ctx)
	if err != nil {
		return err
	}
	return nil
}
