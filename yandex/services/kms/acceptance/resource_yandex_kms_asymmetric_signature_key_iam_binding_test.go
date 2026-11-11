package acceptance

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
	kms "github.com/yandex-cloud/go-genproto/yandex/cloud/kms/v1/asymmetricsignature"
	kmssignaturesdk "github.com/yandex-cloud/go-sdk/services/kms/v1/asymmetricsignature"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
	svc_kms "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/kms/pluginsdk"
)

const kmsAsymmetricSignatureKeyResource = "yandex_kms_asymmetric_signature_key.test-key"

func importKMSAsymmetricSignatureKeyIDFunc(asymmetricSignatureKey *kms.AsymmetricSignatureKey, role string) func(*terraform.State) (string, error) {
	return func(s *terraform.State) (string, error) {
		return asymmetricSignatureKey.Id + "," + role, nil
	}
}

func TestAccKMSAsymmetricSignatureKeyIamBinding_basic(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	var asymmetricSignatureKey kms.AsymmetricSignatureKey
	asymmetricSignatureKeyName := acctest.RandomWithPrefix("tf-kms-asymmetric-signature-key")

	role := "viewer"
	userID := "system:allUsers"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		Steps: []resource.TestStep{
			{
				Config: testAccKMSAsymmetricSignatureKeyIamBindingBasic(asymmetricSignatureKeyName, role, userID),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckKMSAsymmetricSignatureKeyExists(kmsAsymmetricSignatureKeyResource, &asymmetricSignatureKey),
					suite.testAccCheckKMSAsymmetricSignatureKeyIam(kmsAsymmetricSignatureKeyResource, role, []string{userID}),
				),
			},
			{
				ResourceName:                         "yandex_kms_asymmetric_signature_key_iam_binding.viewer",
				ImportStateIdFunc:                    importKMSAsymmetricSignatureKeyIDFunc(&asymmetricSignatureKey, role),
				ImportState:                          true,
				ImportStateVerifyIdentifierAttribute: "asymmetric_signature_key_id",
			},
		},
	})
}

func TestAccKMSAsymmetricSignatureKeyIamBinding_remove(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	var asymmetricSignatureKey kms.AsymmetricSignatureKey
	asymmetricSignatureKeyName := acctest.RandomWithPrefix("tf-kms-asymmetric-signature-key")

	role := "viewer"
	userID := "system:allUsers"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		Steps: []resource.TestStep{
			// Prepare data source
			{
				Config: testAccKMSAsymmetricSignatureKey(asymmetricSignatureKeyName),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckKMSAsymmetricSignatureKeyExists(kmsAsymmetricSignatureKeyResource, &asymmetricSignatureKey),
					suite.testAccCheckKMSAsymmetricSignatureKeyEmptyIam(kmsAsymmetricSignatureKeyResource),
				),
			},
			// Apply IAM bindings
			{
				Config: testAccKMSAsymmetricSignatureKeyIamBindingBasic(asymmetricSignatureKeyName, role, userID),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckKMSAsymmetricSignatureKeyIam(kmsAsymmetricSignatureKeyResource, role, []string{userID}),
				),
			},
			// Remove the bindings
			{
				Config: testAccKMSAsymmetricSignatureKey(asymmetricSignatureKeyName),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckKMSAsymmetricSignatureKeyEmptyIam(kmsAsymmetricSignatureKeyResource),
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

func (suite testSuite) testAccCheckKMSAsymmetricSignatureKeyEmptyIam(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		bindings, err := suite.getKMSAsymmetricSignatureKeyResourceAccessBindings(s, resourceName)
		if err != nil {
			return err
		}

		if len(bindings) == 0 {
			return nil
		}

		return fmt.Errorf("Binding found but expected empty for %s", resourceName)
	}
}

func (suite testSuite) testAccCheckKMSAsymmetricSignatureKeyIam(resourceName, role string, members []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		bindings, err := suite.getKMSAsymmetricSignatureKeyResourceAccessBindings(s, resourceName)
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

func (suite testSuite) getKMSAsymmetricSignatureKeyResourceAccessBindings(s *terraform.State, resourceName string) ([]*access.AccessBinding, error) {
	config, configErr := suite.Config()
	if configErr != nil {
		return nil, configErr
	}

	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("can't find %s in state", resourceName)
	}

	return svc_kms.GetKMSAsymmetricSignatureKeyAccessBindings(config.Context(), config, rs.Primary.ID)
}

func (suite testSuite) testAccCheckKMSAsymmetricSignatureKeyExists(name string, asymmetricSignatureKey *kms.AsymmetricSignatureKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}

		client := kmssignaturesdk.NewAsymmetricSignatureKeyClient(config.SDK)

		found, err := client.Get(context.Background(), &kms.GetAsymmetricSignatureKeyRequest{
			KeyId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("KMS AsymmetricSignatureKey not found")
		}

		*asymmetricSignatureKey = *found

		return nil
	}
}
