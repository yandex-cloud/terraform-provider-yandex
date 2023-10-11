package yandex

import (
	"context"
	"fmt"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/kms/v1/asymmetricsignature"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDataSourceKMSAsymmetricSignatureKey_basic(t *testing.T) {
	keyName := "a" + acctest.RandString(10)
	keyDesc := "Terraform Test"
	folderID := getExampleFolderID()
	basicData := "data.yandex_kms_asymmetric_signature_key.basic_key"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckYandexKmsAsymmetricSignatureKeyAllDestroyed,
		Steps: []resource.TestStep{
			{
				// Create secret
				Config: testAccKMSAsymmetricSignatureKeyResourceAndData(keyName, keyDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceKmsAsymmetricSignatureKeyExists(basicData),
					testAccCheckResourceIDField(basicData, "asymmetric_signature_key_id"),
					resource.TestCheckResourceAttr(basicData, "folder_id", folderID),
					resource.TestCheckResourceAttr(basicData, "name", keyName),
					resource.TestCheckResourceAttr(basicData, "description", keyDesc),
					resource.TestCheckResourceAttr(basicData, "deletion_protection", "false"),
					resource.TestCheckResourceAttr(basicData, "labels.%", "2"),
					resource.TestCheckResourceAttr(basicData, "labels.key1", "value1"),
					resource.TestCheckResourceAttr(basicData, "labels.key2", "value2"),
					testAccCheckCreatedAtAttr(basicData),
				),
			},
		},
	})
}

func testAccKMSAsymmetricSignatureKeyResourceAndData(name, desc string) string {
	return fmt.Sprintf(`
resource "yandex_kms_asymmetric_signature_key" "basic_key" {
  name        = "%v"
  description = "%v"
  labels = {
    key1 = "value1"
    key2 = "value2"
  }
}

data "yandex_kms_asymmetric_signature_key" "basic_key" {
  asymmetric_signature_key_id = yandex_kms_asymmetric_signature_key.basic_key.id
}
`, name, desc)
}

func testAccDataSourceKmsAsymmetricSignatureKeyExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ds, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if ds.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.KMSAsymmetricSignature().AsymmetricSignatureKey().Get(context.Background(), &kms.GetAsymmetricSignatureKeyRequest{
			KeyId: ds.Primary.ID,
		})

		if err != nil {
			return err
		}

		if found.Id != ds.Primary.ID {
			return fmt.Errorf("secret not found: %v", ds.Primary.ID)
		}

		return nil
	}
}

func testAccCheckYandexKmsAsymmetricSignatureKeyAllDestroyed(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_kms_asymmetric_signature_key" {
			continue
		}
		if err := testAccCheckYandexKmsAsymmetricSignatureKeyDestroyed(rs.Primary.ID); err != nil {
			return err
		}
	}
	return nil
}

func testAccCheckYandexKmsAsymmetricSignatureKeyDestroyed(id string) error {
	config := testAccProvider.Meta().(*Config)
	_, err := config.sdk.KMSAsymmetricSignature().AsymmetricSignatureKey().Get(context.Background(), &kms.GetAsymmetricSignatureKeyRequest{
		KeyId: id,
	})
	if err == nil {
		return fmt.Errorf("LockboxSecret %s still exists", id)
	}
	return nil
}
