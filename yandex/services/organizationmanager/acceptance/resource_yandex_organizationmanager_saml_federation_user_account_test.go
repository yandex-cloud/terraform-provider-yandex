package acceptance

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
	svc_organization_manager "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/organizationmanager/pluginsdk"
)

func TestAccOrganizationManagerSamlFederationUser_createDestroy(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	nameID := acctest.RandomWithPrefix("tf-acc")
	info := newSamlFederationInfo()
	federationName := info.getResourceName(true)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testAccOrganizationManagerSamlFederationUserCheckDestroy(nameID, federationName),
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationManagerSamlFederation(info) +
					testAccOrganizationManagerSamlFederationUser(nameID, federationName),
				Check: suite.testAccOrganizationManagerSamlFederationUserCheckCreate(nameID, federationName),
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

func (suite testSuite) testAccOrganizationManagerSamlFederationUserCheckCreate(nameID, federationName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}

		federationID := getFederationID(s, federationName)

		_, err := svc_organization_manager.GetSamlUserAccount(context.Background(), config, federationID, nameID)
		if err != nil {
			return err
		}
		return nil
	}
}

func (suite testSuite) testAccOrganizationManagerSamlFederationUserCheckDestroy(nameID, federationName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}

		federationID := getFederationID(s, federationName)

		_, err := svc_organization_manager.GetSamlUserAccount(context.Background(), config, federationID, nameID)
		if err == nil {
			return fmt.Errorf("saml user account %s was not supposed to be found", nameID)
		}

		return nil
	}
}

func getFederationID(s *terraform.State, federationName string) string {
	return s.RootModule().Resources[federationName].Primary.ID
}
