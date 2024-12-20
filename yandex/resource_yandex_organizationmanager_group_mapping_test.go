package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1"
)

// All federations in example organization get delete by federation sweeper
func init() {
	resource.AddTestSweepers("yandex_organizationmanager_group_mapping", &resource.Sweeper{
		Name:         "yandex_organizationmanager_group_mapping",
		F:            func(_ string) error { return nil },
		Dependencies: []string{"yandex_organizationmanager_saml_federation"},
	})
}

func TestAccOrganizationManagerGroupMapping(t *testing.T) {
	info := newSamlFederationInfo()
	federationName := info.getResourceName(true)
	resourceName := "yandex_organizationmanager_group_mapping.acceptance"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccOrganizationManagerGroupMappingCheckDestroy(federationName),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationManagerSamlFederation(info) +
					testAccOrganizationManagerGroupMapping(federationName, "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOrganizationManagerGroupMappingExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "federation_id"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
				),
			},
			{
				Config: testAccOrganizationManagerSamlFederation(info) +
					testAccOrganizationManagerGroupMapping(federationName, "false"),
				Check: testAccCheckOrganizationManagerGroupMappingExists(federationName),
			},
		},
	})
}

func testAccOrganizationManagerGroupMapping(federationName, enabled string) string {
	return fmt.Sprintf(`resource yandex_organizationmanager_group_mapping "acceptance" {
  federation_id = %s.id
  enabled = %s
}
`, federationName, enabled)
}

func testAccCheckOrganizationManagerGroupMappingExists(federationName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(*Config)
		federationId := s.RootModule().Resources[federationName].Primary.ID

		_, err := config.sdk.OrganizationManager().GroupMapping().Get(context.Background(), &organizationmanager.GetGroupMappingRequest{
			FederationId: federationId,
		})

		return err
	}
}

func testAccOrganizationManagerGroupMappingCheckDestroy(federationName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(*Config)
		federationId := s.RootModule().Resources[federationName].Primary.ID

		_, err := config.sdk.OrganizationManager().GroupMapping().Get(context.Background(), &organizationmanager.GetGroupMappingRequest{
			FederationId: federationId,
		})

		if err == nil {
			return fmt.Errorf("group mapping still exists")
		}

		return nil
	}
}
