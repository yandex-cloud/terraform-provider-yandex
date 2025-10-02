package yandex_kms_asymmetric_signature_key_test

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/kms/v1/asymmetricsignature"
	asymmetricsignaturesdk "github.com/yandex-cloud/go-sdk/services/kms/v1/asymmetricsignature"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDataSourceKMSAsymmetricSignatureKey_UpgradeFromSDKv2(t *testing.T) {
	keyName := "a" + acctest.RandString(10)
	keyDesc := "Terraform Test"
	folderID := test.GetExampleFolderID()
	basicData := "data.yandex_kms_asymmetric_signature_key.basic_key"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { test.AccPreCheck(t) },
		CheckDestroy: testAccCheckYandexKmsAsymmetricSignatureKeyAllDestroyed,
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"yandex": {
						VersionConstraint: "0.150.0",
						Source:            "yandex-cloud/yandex",
					},
				},
				Config: testAccKMSAsymmetricSignatureKeyResourceAndData(keyName, keyDesc),
				Check: resource.ComposeTestCheckFunc(
					test.AccCheckResourceIDField(basicData, "asymmetric_signature_key_id"),
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
				Config:                   testAccKMSAsymmetricSignatureKeyResourceAndData(keyName, keyDesc),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccDataSourceKMSAsymmetricSignatureKey_basic(t *testing.T) {
	keyName := "a" + acctest.RandString(10)
	keyDesc := "Terraform Test"
	folderID := test.GetExampleFolderID()
	basicData := "data.yandex_kms_asymmetric_signature_key.basic_key"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckYandexKmsAsymmetricSignatureKeyAllDestroyed,
		Steps: []resource.TestStep{
			{
				// Create secret
				Config: testAccKMSAsymmetricSignatureKeyResourceAndData(keyName, keyDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceKmsAsymmetricSignatureKeyExists(basicData),
					test.AccCheckResourceIDField(basicData, "asymmetric_signature_key_id"),
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

		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

		found, err := asymmetricsignaturesdk.NewAsymmetricSignatureKeyClient(config.SDKv2).Get(context.Background(), &kms.GetAsymmetricSignatureKeyRequest{
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
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()
	_, err := asymmetricsignaturesdk.NewAsymmetricSignatureKeyClient(config.SDKv2).Get(context.Background(), &kms.GetAsymmetricSignatureKeyRequest{
		KeyId: id,
	})
	if err == nil {
		return fmt.Errorf("LockboxSecret %s still exists", id)
	}
	return nil
}
