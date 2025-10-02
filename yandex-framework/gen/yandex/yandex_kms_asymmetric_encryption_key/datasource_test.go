package yandex_kms_asymmetric_encryption_key_test

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/kms/v1/asymmetricencryption"
	asymmetricencryptionsdk "github.com/yandex-cloud/go-sdk/services/kms/v1/asymmetricencryption"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDataSourceKMSAsymmetricEncryptionKey_UpgradeFromSDKv2(t *testing.T) {
	keyName := "a" + acctest.RandString(10)
	keyDesc := "Terraform Test"
	folderID := test.GetExampleFolderID()
	basicData := "data.yandex_kms_asymmetric_encryption_key.basic_key"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { test.AccPreCheck(t) },
		CheckDestroy: testAccCheckYandexKmsAsymmetricEncryptionKeyAllDestroyed,
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"yandex": {
						VersionConstraint: "0.150.0",
						Source:            "yandex-cloud/yandex",
					},
				},
				Config: testAccKMSAsymmetricEncryptionKeyResourceAndData(keyName, keyDesc),
				Check: resource.ComposeTestCheckFunc(
					test.AccCheckResourceIDField(basicData, "asymmetric_encryption_key_id"),
					resource.TestCheckResourceAttr(basicData, "folder_id", folderID),
					resource.TestCheckResourceAttr(basicData, "name", keyName),
					resource.TestCheckResourceAttr(basicData, "description", keyDesc),
					resource.TestCheckResourceAttr(basicData, "deletion_protection", "false"),
					resource.TestCheckResourceAttr(basicData, "labels.%", "2"),
					resource.TestCheckResourceAttr(basicData, "labels.key1", "value1"),
					resource.TestCheckResourceAttr(basicData, "labels.key2", "value2"),
					test.AccCheckCreatedAtAttr(basicData),
				),
			},
			{
				ProtoV6ProviderFactories: test.AccProviderFactories,
				Config:                   testAccKMSAsymmetricEncryptionKeyResourceAndData(keyName, keyDesc),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccDataSourceKMSAsymmetricEncryptionKey_basic(t *testing.T) {
	keyName := "a" + acctest.RandString(10)
	keyDesc := "Terraform Test"
	folderID := test.GetExampleFolderID()
	basicData := "data.yandex_kms_asymmetric_encryption_key.basic_key"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckYandexKmsAsymmetricEncryptionKeyAllDestroyed,
		Steps: []resource.TestStep{
			{
				// Create secret
				Config: testAccKMSAsymmetricEncryptionKeyResourceAndData(keyName, keyDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceKmsAsymmetricEncryptionKeyExists(basicData),
					test.AccCheckResourceIDField(basicData, "asymmetric_encryption_key_id"),
					resource.TestCheckResourceAttr(basicData, "folder_id", folderID),
					resource.TestCheckResourceAttr(basicData, "name", keyName),
					resource.TestCheckResourceAttr(basicData, "description", keyDesc),
					resource.TestCheckResourceAttr(basicData, "deletion_protection", "false"),
					resource.TestCheckResourceAttr(basicData, "labels.%", "2"),
					resource.TestCheckResourceAttr(basicData, "labels.key1", "value1"),
					resource.TestCheckResourceAttr(basicData, "labels.key2", "value2"),
					test.AccCheckCreatedAtAttr(basicData),
				),
			},
		},
	})
}

func testAccKMSAsymmetricEncryptionKeyResourceAndData(name, desc string) string {
	return fmt.Sprintf(`
resource "yandex_kms_asymmetric_encryption_key" "basic_key" {
  name        = "%v"
  description = "%v"
  labels = {
    key1 = "value1"
    key2 = "value2"
  }
}

data "yandex_kms_asymmetric_encryption_key" "basic_key" {
  asymmetric_encryption_key_id = yandex_kms_asymmetric_encryption_key.basic_key.id
}
`, name, desc)
}

func testAccDataSourceKmsAsymmetricEncryptionKeyExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ds, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if ds.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

		found, err := asymmetricencryptionsdk.NewAsymmetricEncryptionKeyClient(config.SDKv2).Get(context.Background(), &kms.GetAsymmetricEncryptionKeyRequest{
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

func testAccCheckYandexKmsAsymmetricEncryptionKeyAllDestroyed(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_kms_asymmetric_encryption_key" {
			continue
		}
		if err := testAccCheckYandexKmsAsymmetricEncryptionKeyDestroyed(rs.Primary.ID); err != nil {
			return err
		}
	}
	return nil
}

func testAccCheckYandexKmsAsymmetricEncryptionKeyDestroyed(id string) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()
	_, err := asymmetricencryptionsdk.NewAsymmetricEncryptionKeyClient(config.SDKv2).Get(context.Background(), &kms.GetAsymmetricEncryptionKeyRequest{
		KeyId: id,
	})
	if err == nil {
		return fmt.Errorf("LockboxSecret %s still exists", id)
	}
	return nil
}
