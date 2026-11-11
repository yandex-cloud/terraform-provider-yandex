package acceptance

import (
	"fmt"
	"reflect"
	"sort"
	"testing"

	kmsfixtures "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/kms/testfixtures"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/kms/v1"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
	svc_kms "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/kms/pluginsdk"
)

const kmsSymmetricKeyResource = "yandex_kms_symmetric_key.test-key"

func importKMSSymmetricKeyIDFunc(symmetricKey *kms.SymmetricKey, role string) func(*terraform.State) (string, error) {
	return func(s *terraform.State) (string, error) {
		return symmetricKey.Id + "," + role, nil
	}
}

func TestAccKMSSymmetricKeyIamBinding_basic(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	var symmetricKey kms.SymmetricKey
	symmetricKeyName := acctest.RandomWithPrefix("tf-kms-symmetric-key")

	role := "viewer"
	userID := "system:allUsers"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		Steps: []resource.TestStep{
			{
				Config: testAccKMSSymmetricKeyIamBindingBasic(symmetricKeyName, role, userID),
				Check: resource.ComposeTestCheckFunc(
					kmsfixtures.CheckSymmetricKeyExists(suite.Config, kmsSymmetricKeyResource, &symmetricKey),
					suite.testAccCheckKMSSymmetricKeyIam(kmsSymmetricKeyResource, role, []string{userID}),
				),
			},
			{
				ResourceName:                         "yandex_kms_symmetric_key_iam_binding.viewer",
				ImportStateIdFunc:                    importKMSSymmetricKeyIDFunc(&symmetricKey, role),
				ImportState:                          true,
				ImportStateVerifyIdentifierAttribute: "symmetric_key_id",
			},
		},
	})
}

func TestAccKMSSymmetricKeyIamBinding_remove(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	var symmetricKey kms.SymmetricKey
	symmetricKeyName := acctest.RandomWithPrefix("tf-kms-symmetric-key")

	role := "viewer"
	userID := "system:allUsers"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		Steps: []resource.TestStep{
			// Prepare data source
			{
				Config: testAccKMSSymmetricKey(symmetricKeyName),
				Check: resource.ComposeTestCheckFunc(
					kmsfixtures.CheckSymmetricKeyExists(suite.Config, kmsSymmetricKeyResource, &symmetricKey),
					suite.testAccCheckKMSSymmetricKeyEmptyIam(kmsSymmetricKeyResource),
				),
			},
			// Apply IAM bindings
			{
				Config: testAccKMSSymmetricKeyIamBindingBasic(symmetricKeyName, role, userID),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckKMSSymmetricKeyIam(kmsSymmetricKeyResource, role, []string{userID}),
				),
			},
			// Remove the bindings
			{
				Config: testAccKMSSymmetricKey(symmetricKeyName),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckKMSSymmetricKeyEmptyIam(kmsSymmetricKeyResource),
				),
			},
		},
	})
}

func testAccKMSSymmetricKeyIamBindingBasic(symmetricKeyName, role, userID string) string {
	return fmt.Sprintf(`
resource "yandex_kms_symmetric_key" "test-key" {
  name       = "%s"
}

resource "yandex_kms_symmetric_key_iam_binding" "viewer" {
  symmetric_key_id = yandex_kms_symmetric_key.test-key.id
  role        = "%s"
  members     = ["%s"]
}
`, symmetricKeyName, role, userID)
}

func testAccKMSSymmetricKey(symmetricKeyName string) string {
	return fmt.Sprintf(`
resource "yandex_kms_symmetric_key" "test-key" {
  name       = "%s"
}
`, symmetricKeyName)
}

func (suite testSuite) testAccCheckKMSSymmetricKeyEmptyIam(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		bindings, err := suite.getKMSSymmetricKeyResourceAccessBindings(s, resourceName)
		if err != nil {
			return err
		}

		if len(bindings) == 0 {
			return nil
		}

		return fmt.Errorf("Binding found but expected empty for %s", resourceName)
	}
}

func (suite testSuite) testAccCheckKMSSymmetricKeyIam(resourceName, role string, members []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		bindings, err := suite.getKMSSymmetricKeyResourceAccessBindings(s, resourceName)
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

func (suite testSuite) getKMSSymmetricKeyResourceAccessBindings(s *terraform.State, resourceName string) ([]*access.AccessBinding, error) {
	config, configErr := suite.Config()
	if configErr != nil {
		return nil, configErr
	}

	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("can't find %s in state", resourceName)
	}

	return svc_kms.GetKMSSymmetricKeyAccessBindings(config.Context(), config, rs.Primary.ID)
}
