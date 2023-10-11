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
	"github.com/yandex-cloud/go-genproto/yandex/cloud/kms/v1/asymmetricsignature"
)

const kmsAsymmetricSignatureKeyResource = "yandex_kms_asymmetric_signature_key.test-key"

func importKMSAsymmetricSignatureKeyIDFunc(asymmetricSignatureKey *kms.AsymmetricSignatureKey, role string) func(*terraform.State) (string, error) {
	return func(s *terraform.State) (string, error) {
		return asymmetricSignatureKey.Id + " " + role, nil
	}
}

func TestAccKMSAsymmetricSignatureKeyIamBinding_basic(t *testing.T) {
	var asymmetricSignatureKey kms.AsymmetricSignatureKey
	asymmetricSignatureKeyName := acctest.RandomWithPrefix("tf-kms-asymmetric-signature-key")

	role := "viewer"
	userID := "system:allUsers"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccKMSAsymmetricSignatureKeyIamBindingBasic(asymmetricSignatureKeyName, role, userID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKMSAsymmetricSignatureKeyExists(kmsAsymmetricSignatureKeyResource, &asymmetricSignatureKey),
					testAccCheckKMSAsymmetricSignatureKeyIam(kmsAsymmetricSignatureKeyResource, role, []string{userID}),
				),
			},
			{
				ResourceName:      "yandex_kms_asymmetric_signature_key_iam_binding.viewer",
				ImportStateIdFunc: importKMSAsymmetricSignatureKeyIDFunc(&asymmetricSignatureKey, role),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccKMSAsymmetricSignatureKeyIamBinding_remove(t *testing.T) {
	var asymmetricSignatureKey kms.AsymmetricSignatureKey
	asymmetricSignatureKeyName := acctest.RandomWithPrefix("tf-kms-asymmetric-signature-key")

	role := "viewer"
	userID := "system:allUsers"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			// Prepare data source
			{
				Config: testAccKMSAsymmetricSignatureKey(asymmetricSignatureKeyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKMSAsymmetricSignatureKeyExists(kmsAsymmetricSignatureKeyResource, &asymmetricSignatureKey),
					testAccCheckKMSAsymmetricSignatureKeyEmptyIam(kmsAsymmetricSignatureKeyResource),
				),
			},
			// Apply IAM bindings
			{
				Config: testAccKMSAsymmetricSignatureKeyIamBindingBasic(asymmetricSignatureKeyName, role, userID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKMSAsymmetricSignatureKeyIam(kmsAsymmetricSignatureKeyResource, role, []string{userID}),
				),
			},
			// Remove the bindings
			{
				Config: testAccKMSAsymmetricSignatureKey(asymmetricSignatureKeyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKMSAsymmetricSignatureKeyEmptyIam(kmsAsymmetricSignatureKeyResource),
				),
			},
		},
	})
}

func testAccKMSAsymmetricSignatureKeyIamBindingBasic(asymmetricSignatureKeyName, role, userID string) string {
	return fmt.Sprintf(`
resource "yandex_kms_asymmetric_signature_key" "test-key" {
  name       = "%s"
}

resource "yandex_kms_asymmetric_signature_key_iam_binding" "viewer" {
  asymmetric_signature_key_id = yandex_kms_asymmetric_signature_key.test-key.id
  role        = "%s"
  members     = ["%s"]
}
`, asymmetricSignatureKeyName, role, userID)
}

func testAccKMSAsymmetricSignatureKey(asymmetricSignatureKeyName string) string {
	return fmt.Sprintf(`
resource "yandex_kms_asymmetric_signature_key" "test-key" {
  name       = "%s"
}
`, asymmetricSignatureKeyName)
}

func testAccCheckKMSAsymmetricSignatureKeyEmptyIam(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		bindings, err := getKMSAsymmetricSignatureKeyResourceAccessBindings(s, resourceName)
		if err != nil {
			return err
		}

		if len(bindings) == 0 {
			return nil
		}

		return fmt.Errorf("Binding found but expected empty for %s", resourceName)
	}
}

func testAccCheckKMSAsymmetricSignatureKeyIam(resourceName, role string, members []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		bindings, err := getKMSAsymmetricSignatureKeyResourceAccessBindings(s, resourceName)
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

func getKMSAsymmetricSignatureKeyResourceAccessBindings(s *terraform.State, resourceName string) ([]*access.AccessBinding, error) {
	config := testAccProvider.Meta().(*Config)

	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("can't find %s in state", resourceName)
	}

	return getKMSAsymmetricSignatureKeyAccessBindings(config.Context(), config, rs.Primary.ID)
}
