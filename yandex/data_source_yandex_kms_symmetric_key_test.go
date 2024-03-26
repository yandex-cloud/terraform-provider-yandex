package yandex

import (
	"context"
	"fmt"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/kms/v1"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDataSourceKMSSymmetricKey_basic(t *testing.T) {
	keyName := "a" + acctest.RandString(10)
	keyDesc := "Terraform Test"
	folderID := getExampleFolderID()
	basicData := "data.yandex_kms_symmetric_key.basic_key"
	basicDataByName := "data.yandex_kms_symmetric_key.basic_key_by_name"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckYandexKmsSymmetricKeyAllDestroyed,
		Steps: []resource.TestStep{
			{
				// Create secret
				Config: testAccKMSSymmetricKeyResourceAndData(keyName, keyDesc),
				Check: resource.ComposeTestCheckFunc(
					// checks for the key obtained by ID
					testAccDataSourceKmsSymmetricKeyExists(basicData),
					testAccCheckResourceIDField(basicData, "symmetric_key_id"),
					resource.TestCheckResourceAttr(basicData, "folder_id", folderID),
					resource.TestCheckResourceAttr(basicData, "name", keyName),
					resource.TestCheckResourceAttr(basicData, "description", keyDesc),
					resource.TestCheckResourceAttr(basicData, "deletion_protection", "false"),
					resource.TestCheckResourceAttr(basicData, "labels.%", "2"),
					resource.TestCheckResourceAttr(basicData, "labels.key1", "value1"),
					resource.TestCheckResourceAttr(basicData, "labels.key2", "value2"),
					testAccCheckCreatedAtAttr(basicData),
					// same checks, now for the key obtained by name
					testAccDataSourceKmsSymmetricKeyExists(basicDataByName),
					testAccCheckResourceIDField(basicDataByName, "symmetric_key_id"),
					resource.TestCheckResourceAttr(basicDataByName, "folder_id", folderID),
					resource.TestCheckResourceAttr(basicDataByName, "name", keyName),
					resource.TestCheckResourceAttr(basicDataByName, "description", keyDesc),
					resource.TestCheckResourceAttr(basicDataByName, "deletion_protection", "false"),
					resource.TestCheckResourceAttr(basicDataByName, "labels.%", "2"),
					resource.TestCheckResourceAttr(basicDataByName, "labels.key1", "value1"),
					resource.TestCheckResourceAttr(basicDataByName, "labels.key2", "value2"),
					testAccCheckCreatedAtAttr(basicDataByName),
				),
			},
		},
	})
}

func testAccKMSSymmetricKeyResourceAndData(name, desc string) string {
	return fmt.Sprintf(`
resource "yandex_kms_symmetric_key" "basic_key" {
  name        = "%v"
  description = "%v"
  labels = {
    key1 = "value1"
    key2 = "value2"
  }
}

data "yandex_kms_symmetric_key" "basic_key" {
  symmetric_key_id = yandex_kms_symmetric_key.basic_key.id
}

data "yandex_kms_symmetric_key" "basic_key_by_name" {
  name = yandex_kms_symmetric_key.basic_key.name
}
`, name, desc)
}

func testAccDataSourceKmsSymmetricKeyExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ds, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if ds.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.KMS().SymmetricKey().Get(context.Background(), &kms.GetSymmetricKeyRequest{
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

func testAccCheckYandexKmsSymmetricKeyAllDestroyed(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_kms_symmetric_key" {
			continue
		}
		if err := testAccCheckYandexKmsSymmetricKeyDestroyed(rs.Primary.ID); err != nil {
			return err
		}
	}
	return nil
}

func testAccCheckYandexKmsSymmetricKeyDestroyed(id string) error {
	config := testAccProvider.Meta().(*Config)
	_, err := config.sdk.KMS().SymmetricKey().Get(context.Background(), &kms.GetSymmetricKeyRequest{
		KeyId: id,
	})
	if err == nil {
		return fmt.Errorf("LockboxSecret %s still exists", id)
	}
	return nil
}
