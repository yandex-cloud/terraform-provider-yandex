package yandex

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/kms/v1"
)

const kmsSymmetricKeyResource = "yandex_kms_symmetric_key.test-key"

func importKMSSymmetricKeyIDFunc(symmetricKey *kms.SymmetricKey, role string) func(*terraform.State) (string, error) {
	return func(s *terraform.State) (string, error) {
		return symmetricKey.Id + "," + role, nil
	}
}

func TestAccKMSSymmetricKeyIamBinding_basic(t *testing.T) {
	var symmetricKey kms.SymmetricKey
	symmetricKeyName := acctest.RandomWithPrefix("tf-kms-symmetric-key")

	role := "viewer"
	userID := "system:allUsers"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactoriesV6,
		Steps: []resource.TestStep{
			{
				Config: testAccKMSSymmetricKeyIamBindingBasic(symmetricKeyName, role, userID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKMSSymmetricKeyExists(kmsSymmetricKeyResource, &symmetricKey),
					testAccCheckKMSSymmetricKeyIam(kmsSymmetricKeyResource, role, []string{userID}),
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
	var symmetricKey kms.SymmetricKey
	symmetricKeyName := acctest.RandomWithPrefix("tf-kms-symmetric-key")

	role := "viewer"
	userID := "system:allUsers"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactoriesV6,
		Steps: []resource.TestStep{
			// Prepare data source
			{
				Config: testAccKMSSymmetricKey(symmetricKeyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKMSSymmetricKeyExists(kmsSymmetricKeyResource, &symmetricKey),
					testAccCheckKMSSymmetricKeyEmptyIam(kmsSymmetricKeyResource),
				),
			},
			// Apply IAM bindings
			{
				Config: testAccKMSSymmetricKeyIamBindingBasic(symmetricKeyName, role, userID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKMSSymmetricKeyIam(kmsSymmetricKeyResource, role, []string{userID}),
				),
			},
			// Remove the bindings
			{
				Config: testAccKMSSymmetricKey(symmetricKeyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKMSSymmetricKeyEmptyIam(kmsSymmetricKeyResource),
				),
			},
		},
	})
}

func testAccCheckKMSSymmetricKeyExists(name string, symmetricKey *kms.SymmetricKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.KMS().SymmetricKey().Get(context.Background(), &kms.GetSymmetricKeyRequest{
			KeyId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("KMS Symmetric Key not found")
		}

		*symmetricKey = *found

		return nil
	}
}

func testAccCheckKMSSymmetricKeyDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_kms_symmetric_key" {
			continue
		}

		_, err := config.sdk.KMS().SymmetricKey().Get(context.Background(), &kms.GetSymmetricKeyRequest{
			KeyId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("KMS Symmetric Key still exists")
		}
	}

	return nil
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

func testAccCheckKMSSymmetricKeyEmptyIam(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		bindings, err := getKMSSymmetricKeyResourceAccessBindings(s, resourceName)
		if err != nil {
			return err
		}

		if len(bindings) == 0 {
			return nil
		}

		return fmt.Errorf("Binding found but expected empty for %s", resourceName)
	}
}

func testAccCheckKMSSymmetricKeyIam(resourceName, role string, members []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		bindings, err := getKMSSymmetricKeyResourceAccessBindings(s, resourceName)
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

func getKMSSymmetricKeyResourceAccessBindings(s *terraform.State, resourceName string) ([]*access.AccessBinding, error) {
	config := testAccProvider.Meta().(*Config)

	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("can't find %s in state", resourceName)
	}

	return getKMSSymmetricKeyAccessBindings(config.Context(), config, rs.Primary.ID)
}
