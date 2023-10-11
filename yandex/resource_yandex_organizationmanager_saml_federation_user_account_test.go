package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// All federations in example organization get delete by federation sweeper
func init() {
	resource.AddTestSweepers("yandex_organizationmanager_saml_federation_user_account", &resource.Sweeper{
		Name:         "yandex_organizationmanager_saml_federation_user_account",
		F:            func(_ string) error { return nil },
		Dependencies: []string{"yandex_organizationmanager_saml_federation"},
	})
}

func TestAccOrganizationManagerSamlFederationUser_createDestroy(t *testing.T) {
	nameID := acctest.RandomWithPrefix("tf-acc")
	info := newSamlFederationInfo()
	federationName := info.getResourceName(true)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccOrganizationManagerSamlFederationUserCheckDestroy(nameID, federationName),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationManagerSamlFederation(info) +
					testAccOrganizationManagerSamlFederationUser(nameID, federationName),
				Check: testAccOrganizationManagerSamlFederationUserCheckCreate(nameID, federationName),
			},
			{
				ImportState:  true,
				ResourceName: "yandex_organizationmanager_saml_federation_user_account.basic",
				Config: testAccOrganizationManagerSamlFederation(info) +
					testAccOrganizationManagerSamlFederationUser(nameID, federationName),
			},
		},
	})
}

func testAccOrganizationManagerSamlFederationUser(nameID, federationName string) string {
	return fmt.Sprintf(`resource yandex_organizationmanager_saml_federation_user_account "basic" {
  name_id       = "%s"
  federation_id = %s.id
}`, nameID, federationName)
}

func testAccOrganizationManagerSamlFederationUserCheckCreate(nameID, federationName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(*Config)
		federationID := getFederationID(s, federationName)

		_, err := getSamlUserAccount(context.Background(), config, federationID, nameID)
		if err != nil {
			return err
		}
		return nil
	}
}

func testAccOrganizationManagerSamlFederationUserCheckDestroy(nameID, federationName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(*Config)
		federationID := getFederationID(s, federationName)

		_, err := getSamlUserAccount(context.Background(), config, federationID, nameID)
		if err == nil {
			return fmt.Errorf("saml user account %s was not supposed to be found", nameID)
		}

		return nil
	}
}

func getFederationID(s *terraform.State, federationName string) string {
	return s.RootModule().Resources[federationName].Primary.ID
}
