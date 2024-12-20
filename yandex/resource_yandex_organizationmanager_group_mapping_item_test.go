package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// All federations and groups in example organization get delete by sweepers
func init() {
	resource.AddTestSweepers("yandex_organizationmanager_group_mapping_item", &resource.Sweeper{
		Name: "yandex_organizationmanager_group_mapping_item",
		F:    func(_ string) error { return nil },
		Dependencies: []string{
			"yandex_organizationmanager_saml_federation",
			"yandex_organizationmanager_group",
		},
	})
}

func TestAccOrganizationManagerGroupMappingItem(t *testing.T) {
	federationInfo := newSamlFederationInfo()
	groupInfo := newGroupInfo()
	federationName := federationInfo.getResourceName(true)
	groupName := groupInfo.getResourceName(true)
	resourceName := "yandex_organizationmanager_group_mapping_item.acceptance"
	externalGroupId := "test_external_group_id"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationManagerSamlFederation(federationInfo) +
					testAccOrganizationManagerGroup(groupInfo) +
					testAccOrganizationManagerGroupMapping(federationName, "true") +
					testAccOrganizationManagerGroupMappingItem(federationName, groupName, externalGroupId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationManagerGroupMappingItemExists(federationName, groupName, externalGroupId),
					resource.TestCheckResourceAttrSet(resourceName, "federation_id"),
					resource.TestCheckResourceAttrSet(resourceName, "internal_group_id"),
					resource.TestCheckResourceAttr(resourceName, "external_group_id", externalGroupId),
				),
			},
			{
				Config: testAccOrganizationManagerSamlFederation(federationInfo) +
					testAccOrganizationManagerGroup(groupInfo) +
					testAccOrganizationManagerGroupMapping(federationName, "true"),
				Check: testAccOrganizationManagerGroupMappingItemCheckDestroy(federationName, groupName, externalGroupId),
			},
		},
	})
}

func testAccOrganizationManagerGroupMappingItem(federationName, groupName, externalGroupId string) string {
	return fmt.Sprintf(`resource yandex_organizationmanager_group_mapping_item "acceptance" {
  federation_id = %s.id
  internal_group_id = %s.id
  external_group_id = "%s"

  depends_on = [yandex_organizationmanager_group_mapping.acceptance]
}`, federationName, groupName, externalGroupId)
}

func testAccCheckOrganizationManagerGroupMappingItemExists(federationName, groupName, externalGroupID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(*Config)
		federationId := s.RootModule().Resources[federationName].Primary.ID
		groupId := s.RootModule().Resources[groupName].Primary.ID

		items, err := getGroupMappingItems(context.Background(), config, federationId)
		if err != nil {
			return err
		}

		for _, item := range items {
			if item.InternalGroupId == groupId &&
				item.ExternalGroupId == externalGroupID {
				return nil
			}
		}

		return fmt.Errorf("organization manager group mapping item not found")
	}
}

func testAccOrganizationManagerGroupMappingItemCheckDestroy(federationName, groupName, externalGroupID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(*Config)
		federationId := s.RootModule().Resources[federationName].Primary.ID
		groupId := s.RootModule().Resources[groupName].Primary.ID

		items, err := getGroupMappingItems(context.Background(), config, federationId)
		if err != nil {
			return err
		}

		for _, item := range items {
			if item.InternalGroupId == groupId &&
				item.ExternalGroupId == externalGroupID {
				return fmt.Errorf("group mapping item still exists")
			}
		}

		return nil
	}
}
