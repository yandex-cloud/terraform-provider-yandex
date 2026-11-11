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
	kms "github.com/yandex-cloud/go-genproto/yandex/cloud/kms/v1/asymmetricencryption"
	kmsencryptionsdk "github.com/yandex-cloud/go-sdk/services/kms/v1/asymmetricencryption"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
	svc_kms "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/kms/pluginsdk"
)

const kmsAsymmetricEncryptionKeyResource = "yandex_kms_asymmetric_encryption_key.test-key"

func importKMSAsymmetricEncryptionKeyIDFunc(asymmetricEncryptionKey *kms.AsymmetricEncryptionKey, role string) func(*terraform.State) (string, error) {
	return func(s *terraform.State) (string, error) {
		return asymmetricEncryptionKey.Id + "," + role, nil
	}
}

func TestAccKMSAsymmetricEncryptionKeyIamBinding_basic(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	var asymmetricEncryptionKey kms.AsymmetricEncryptionKey
	asymmetricEncryptionKeyName := acctest.RandomWithPrefix("tf-kms-asymmetric-encryption-key")

	role := "viewer"
	userID := "system:allUsers"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		Steps: []resource.TestStep{
			{
				Config: testAccKMSAsymmetricEncryptionKeyIamBindingBasic(asymmetricEncryptionKeyName, role, userID),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckKMSAsymmetricEncryptionKeyExists(kmsAsymmetricEncryptionKeyResource, &asymmetricEncryptionKey),
					suite.testAccCheckKMSAsymmetricEncryptionKeyIam(kmsAsymmetricEncryptionKeyResource, role, []string{userID}),
				),
			},
			{
				ResourceName:                         "yandex_kms_asymmetric_encryption_key_iam_binding.viewer",
				ImportStateIdFunc:                    importKMSAsymmetricEncryptionKeyIDFunc(&asymmetricEncryptionKey, role),
				ImportState:                          true,
				ImportStateVerifyIdentifierAttribute: "asymmetric_encryption_key_id",
			},
		},
	})
}

func TestAccKMSAsymmetricEncryptionKeyIamBinding_remove(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	var asymmetricEncryptionKey kms.AsymmetricEncryptionKey
	asymmetricEncryptionKeyName := acctest.RandomWithPrefix("tf-kms-asymmetric-encryption-key")

	role := "viewer"
	userID := "system:allUsers"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		Steps: []resource.TestStep{
			// Prepare data source
			{
				Config: testAccKMSAsymmetricEncryptionKey(asymmetricEncryptionKeyName),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckKMSAsymmetricEncryptionKeyExists(kmsAsymmetricEncryptionKeyResource, &asymmetricEncryptionKey),
					suite.testAccCheckKMSAsymmetricEncryptionKeyEmptyIam(kmsAsymmetricEncryptionKeyResource),
				),
			},
			// Apply IAM bindings
			{
				Config: testAccKMSAsymmetricEncryptionKeyIamBindingBasic(asymmetricEncryptionKeyName, role, userID),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckKMSAsymmetricEncryptionKeyIam(kmsAsymmetricEncryptionKeyResource, role, []string{userID}),
				),
			},
			// Remove the bindings
			{
				Config: testAccKMSAsymmetricEncryptionKey(asymmetricEncryptionKeyName),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckKMSAsymmetricEncryptionKeyEmptyIam(kmsAsymmetricEncryptionKeyResource),
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

func (suite testSuite) testAccCheckKMSAsymmetricEncryptionKeyEmptyIam(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		bindings, err := suite.getKMSAsymmetricEncryptionKeyResourceAccessBindings(s, resourceName)
		if err != nil {
			return err
		}

		if len(bindings) == 0 {
			return nil
		}

		return fmt.Errorf("Binding found but expected empty for %s", resourceName)
	}
}

func (suite testSuite) testAccCheckKMSAsymmetricEncryptionKeyIam(resourceName, role string, members []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		bindings, err := suite.getKMSAsymmetricEncryptionKeyResourceAccessBindings(s, resourceName)
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

func (suite testSuite) getKMSAsymmetricEncryptionKeyResourceAccessBindings(s *terraform.State, resourceName string) ([]*access.AccessBinding, error) {
	config, configErr := suite.Config()
	if configErr != nil {
		return nil, configErr
	}

	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("can't find %s in state", resourceName)
	}

	return svc_kms.GetKMSAsymmetricEncryptionKeyAccessBindings(config.Context(), config, rs.Primary.ID)
}

func (suite testSuite) testAccCheckKMSAsymmetricEncryptionKeyExists(name string, asymmetricEncryptionKey *kms.AsymmetricEncryptionKey) resource.TestCheckFunc {
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

		client := kmsencryptionsdk.NewAsymmetricEncryptionKeyClient(config.SDK)

		found, err := client.Get(context.Background(), &kms.GetAsymmetricEncryptionKeyRequest{
			KeyId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("KMS AsymmetricEncryptionKey not found")
		}

		*asymmetricEncryptionKey = *found

		return nil
	}
}
