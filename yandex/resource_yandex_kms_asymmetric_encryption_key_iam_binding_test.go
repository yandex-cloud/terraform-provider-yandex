package yandex

import (
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/kms/v1/asymmetricencryption"
)

const kmsAsymmetricEncryptionKeyResource = "yandex_kms_asymmetric_encryption_key.test-key"

func importKMSAsymmetricEncryptionKeyIDFunc(asymmetricEncryptionKey *kms.AsymmetricEncryptionKey, role string) func(*terraform.State) (string, error) {
	return func(s *terraform.State) (string, error) {
		return asymmetricEncryptionKey.Id + " " + role, nil
	}
}

func TestAccKMSAsymmetricEncryptionKeyIamBinding_basic(t *testing.T) {
	var asymmetricEncryptionKey kms.AsymmetricEncryptionKey
	asymmetricEncryptionKeyName := acctest.RandomWithPrefix("tf-kms-asymmetric-encryption-key")

	role := "viewer"
	userID := "system:allUsers"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccKMSAsymmetricEncryptionKeyIamBindingBasic(asymmetricEncryptionKeyName, role, userID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKMSAsymmetricEncryptionKeyExists(kmsAsymmetricEncryptionKeyResource, &asymmetricEncryptionKey),
					testAccCheckKMSAsymmetricEncryptionKeyIam(kmsAsymmetricEncryptionKeyResource, role, []string{userID}),
				),
			},
			{
				ResourceName:      "yandex_kms_asymmetric_encryption_key_iam_binding.viewer",
				ImportStateIdFunc: importKMSAsymmetricEncryptionKeyIDFunc(&asymmetricEncryptionKey, role),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccKMSAsymmetricEncryptionKeyIamBinding_remove(t *testing.T) {
	var asymmetricEncryptionKey kms.AsymmetricEncryptionKey
	asymmetricEncryptionKeyName := acctest.RandomWithPrefix("tf-kms-asymmetric-encryption-key")

	role := "viewer"
	userID := "system:allUsers"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			// Prepare data source
			{
				Config: testAccKMSAsymmetricEncryptionKey(asymmetricEncryptionKeyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKMSAsymmetricEncryptionKeyExists(kmsAsymmetricEncryptionKeyResource, &asymmetricEncryptionKey),
					testAccCheckKMSAsymmetricEncryptionKeyEmptyIam(kmsAsymmetricEncryptionKeyResource),
				),
			},
			// Apply IAM bindings
			{
				Config: testAccKMSAsymmetricEncryptionKeyIamBindingBasic(asymmetricEncryptionKeyName, role, userID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKMSAsymmetricEncryptionKeyIam(kmsAsymmetricEncryptionKeyResource, role, []string{userID}),
				),
			},
			// Remove the bindings
			{
				Config: testAccKMSAsymmetricEncryptionKey(asymmetricEncryptionKeyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKMSAsymmetricEncryptionKeyEmptyIam(kmsAsymmetricEncryptionKeyResource),
				),
			},
		},
	})
}

func testAccKMSAsymmetricEncryptionKeyIamBindingBasic(asymmetricEncryptionKeyName, role, userID string) string {
	return fmt.Sprintf(`
resource "yandex_kms_asymmetric_encryption_key" "test-key" {
  name       = "%s"
}

resource "yandex_kms_asymmetric_encryption_key_iam_binding" "viewer" {
  asymmetric_encryption_key_id = yandex_kms_asymmetric_encryption_key.test-key.id
  role        = "%s"
  members     = ["%s"]
}
`, asymmetricEncryptionKeyName, role, userID)
}

func testAccKMSAsymmetricEncryptionKey(asymmetricEncryptionKeyName string) string {
	return fmt.Sprintf(`
resource "yandex_kms_asymmetric_encryption_key" "test-key" {
  name       = "%s"
}
`, asymmetricEncryptionKeyName)
}

func testAccCheckKMSAsymmetricEncryptionKeyEmptyIam(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		bindings, err := getKMSAsymmetricEncryptionKeyResourceAccessBindings(s, resourceName)
		if err != nil {
			return err
		}

		if len(bindings) == 0 {
			return nil
		}

		return fmt.Errorf("Binding found but expected empty for %s", resourceName)
	}
}

func testAccCheckKMSAsymmetricEncryptionKeyIam(resourceName, role string, members []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		bindings, err := getKMSAsymmetricEncryptionKeyResourceAccessBindings(s, resourceName)
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

func getKMSAsymmetricEncryptionKeyResourceAccessBindings(s *terraform.State, resourceName string) ([]*access.AccessBinding, error) {
	config := testAccProvider.Meta().(*Config)

	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("can't find %s in state", resourceName)
	}

	return getKMSAsymmetricEncryptionKeyAccessBindings(config.Context(), config, rs.Primary.ID)
}
