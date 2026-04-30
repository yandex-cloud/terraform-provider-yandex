package yandex_organizationmanager_group_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1"
	organizationmanagerv1sdk "github.com/yandex-cloud/go-sdk/services/organizationmanager/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
)

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

// The config here should match as closely as possible to the one presented to the user in the docs.
// Serves as a proof that the example config is viable.
func TestAccOrganizationManagerGroup_UpgradeFromSDKv2(t *testing.T) {
	t.Parallel()

	config := fmt.Sprintf(`
resource "yandex_organizationmanager_group" group {
  name            = "my-group"
  description     = "My new Group"
  organization_id = "%s"
}
`, test.GetExampleOrganizationID())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { test.AccPreCheck(t) },
		CheckDestroy: testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"yandex": {
						VersionConstraint: "0.150.0",
						Source:            "yandex-cloud/yandex",
					},
				},
				Config: config},
			{
				ProtoV6ProviderFactories: test.AccProviderFactories,
				Config:                   config, ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccOrganizationManagerGroup_basic(t *testing.T) {
	t.Parallel()

	groupName := acctest.RandomWithPrefix("tf-acc")
	groupDesc := acctest.RandString(20)
	resourceAddr := "yandex_organizationmanager_group.group"
	var group organizationmanager.Group

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroup_basic(groupName, groupDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceAddr, &group),
					resource.TestCheckResourceAttr(resourceAddr, "name", groupName),
					resource.TestCheckResourceAttr(resourceAddr, "description", groupDesc),
					resource.TestCheckResourceAttr(resourceAddr, "labels.%", "2"),
					resource.TestCheckResourceAttr(resourceAddr, "labels.test-label", "example-label-value"),
					resource.TestCheckResourceAttr(resourceAddr, "labels.removed-label", "will-be-deleted"),
					resource.TestCheckResourceAttrSet(resourceAddr, "created_at"),
					resource.TestCheckResourceAttrSet(resourceAddr, "organization_id"),
				),
			},
			{
				ResourceName:      resourceAddr,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccOrganizationManagerGroup_update(t *testing.T) {
	t.Parallel()

	groupName := acctest.RandomWithPrefix("tf-acc")
	groupDesc := acctest.RandString(20)
	resourceAddr := "yandex_organizationmanager_group.group"
	var group organizationmanager.Group

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroup_basic(groupName, groupDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceAddr, &group),
					resource.TestCheckResourceAttr(resourceAddr, "name", groupName),
					resource.TestCheckResourceAttr(resourceAddr, "description", groupDesc),
					resource.TestCheckResourceAttr(resourceAddr, "labels.test-label", "example-label-value"),
					resource.TestCheckResourceAttr(resourceAddr, "labels.removed-label", "will-be-deleted"),
				),
			},
			{
				Config: testAccGroup_update(groupName+"-updated", groupDesc+" updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceAddr, &group),
					resource.TestCheckResourceAttrPtr(resourceAddr, "id", &group.Id),
					resource.TestCheckResourceAttr(resourceAddr, "name", groupName+"-updated"),
					resource.TestCheckResourceAttr(resourceAddr, "description", groupDesc+" updated"),
					resource.TestCheckResourceAttr(resourceAddr, "labels.%", "2"),
					// existing label, value modified
					resource.TestCheckResourceAttr(resourceAddr, "labels.test-label", "modified-value"),
					// new label, added
					resource.TestCheckResourceAttr(resourceAddr, "labels.added-label", "brand-new"),
					// previously existing label, removed
					resource.TestCheckNoResourceAttr(resourceAddr, "labels.removed-label"),
				),
			},
			{
				ResourceName:      resourceAddr,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccGroup_basic(name, description string) string {
	return fmt.Sprintf(`
resource "yandex_organizationmanager_group" "group" {
  name            = "%s"
  description     = "%s"
  organization_id = "%s"

  labels = {
    test-label    = "example-label-value"
    removed-label = "will-be-deleted"
  }
}
`, name, description, test.GetExampleOrganizationID())
}

func testAccGroup_update(name, description string) string {
	return fmt.Sprintf(`
resource "yandex_organizationmanager_group" "group" {
  name            = "%s"
  description     = "%s"
  organization_id = "%s"

  labels = {
    test-label  = "modified-value"
    added-label = "brand-new"
  }
}
`, name, description, test.GetExampleOrganizationID())
}

func testAccCheckGroupExists(n string, group *organizationmanager.Group) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID set for %s", n)
		}

		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()
		found, err := organizationmanagerv1sdk.NewGroupClient(config.SDKv2).Get(context.Background(), &organizationmanager.GetGroupRequest{
			GroupId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}
		if found.Id != rs.Primary.ID {
			return fmt.Errorf("group not found")
		}
		*group = *found
		return nil
	}
}

func testAccCheckGroupDestroy(s *terraform.State) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_organizationmanager_group" {
			continue
		}

		_, err := organizationmanagerv1sdk.NewGroupClient(config.SDKv2).Get(context.Background(), &organizationmanager.GetGroupRequest{
			GroupId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("Group still exists")
		}
	}

	return nil
}
