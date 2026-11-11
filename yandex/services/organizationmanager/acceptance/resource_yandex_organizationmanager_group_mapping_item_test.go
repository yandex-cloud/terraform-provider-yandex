package acceptance

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
	svc_organization_manager "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/organizationmanager/pluginsdk"
)

func TestAccOrganizationManagerGroupMappingItem(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	federationInfo := newSamlFederationInfo()
	groupInfo := newGroupInfo()
	federationName := federationInfo.getResourceName(true)
	groupName := groupInfo.getResourceName(true)
	resourceName := "yandex_organizationmanager_group_mapping_item.acceptance"
	externalGroupId := "test_external_group_id"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationManagerSamlFederation(federationInfo) +
					testAccOrganizationManagerGroup(groupInfo) +
					testAccOrganizationManagerGroupMapping(federationName, "true") +
					testAccOrganizationManagerGroupMappingItem(federationName, groupName, externalGroupId),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckOrganizationManagerGroupMappingItemExists(federationName, groupName, externalGroupId),
					resource.TestCheckResourceAttrSet(resourceName, "federation_id"),
					resource.TestCheckResourceAttrSet(resourceName, "internal_group_id"),
					resource.TestCheckResourceAttr(resourceName, "external_group_id", externalGroupId),
				),
			},
			{
				Config: testAccOrganizationManagerSamlFederation(federationInfo) +
					testAccOrganizationManagerGroup(groupInfo) +
					testAccOrganizationManagerGroupMapping(federationName, "true"),
				Check: suite.testAccOrganizationManagerGroupMappingItemCheckDestroy(federationName, groupName, externalGroupId),
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

func (suite testSuite) testAccCheckOrganizationManagerGroupMappingItemExists(federationName, groupName, externalGroupID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}

		federationId := s.RootModule().Resources[federationName].Primary.ID
		groupId := s.RootModule().Resources[groupName].Primary.ID

		items, err := svc_organization_manager.GetGroupMappingItems(context.Background(), config, federationId)
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

func (suite testSuite) testAccOrganizationManagerGroupMappingItemCheckDestroy(federationName, groupName, externalGroupID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}

		federationId := s.RootModule().Resources[federationName].Primary.ID
		groupId := s.RootModule().Resources[groupName].Primary.ID

		items, err := svc_organization_manager.GetGroupMappingItems(context.Background(), config, federationId)
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
