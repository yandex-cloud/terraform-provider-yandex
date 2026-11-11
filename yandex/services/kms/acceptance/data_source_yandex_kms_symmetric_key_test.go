package acceptance

import (
	"context"
	"fmt"
	"testing"

	kmsfixtures "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/kms/testfixtures"

	statecheck "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance/statecheck"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/kms/v1"
	kmssdk "github.com/yandex-cloud/go-sdk/services/kms/v1"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
)

func TestAccDataSourceKMSSymmetricKey_basic(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	keyName := "a" + acctest.RandString(10)
	keyDesc := "Terraform Test"
	folderID := acceptance.FolderID()
	basicData := "data.yandex_kms_symmetric_key.basic_key"
	basicDataByName := "data.yandex_kms_symmetric_key.basic_key_by_name"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy: kmsfixtures.CheckSymmetricKeysDestroy(func() (kmsfixtures.SymmetricKeyGetter, error) {
			config, err := suite.Config()
			if err != nil {
				return nil, err
			}
			return kmssdk.NewSymmetricKeyClient(config.SDK), nil
		}),
		Steps: []resource.TestStep{
			{
				// Create secret
				Config: testAccKMSSymmetricKeyResourceAndData(keyName, keyDesc),
				Check: resource.ComposeTestCheckFunc(
					// checks for the key obtained by ID
					suite.testAccDataSourceKmsSymmetricKeyExists(basicData),
					statecheck.CheckResourceIDField(basicData, "symmetric_key_id"),
					resource.TestCheckResourceAttr(basicData, "folder_id", folderID),
					resource.TestCheckResourceAttr(basicData, "name", keyName),
					resource.TestCheckResourceAttr(basicData, "description", keyDesc),
					resource.TestCheckResourceAttr(basicData, "deletion_protection", "false"),
					resource.TestCheckResourceAttr(basicData, "labels.%", "2"),
					resource.TestCheckResourceAttr(basicData, "labels.key1", "value1"),
					resource.TestCheckResourceAttr(basicData, "labels.key2", "value2"),
					statecheck.CheckCreatedAtAttr(basicData),
					// same checks, now for the key obtained by name
					suite.testAccDataSourceKmsSymmetricKeyExists(basicDataByName),
					statecheck.CheckResourceIDField(basicDataByName, "symmetric_key_id"),
					resource.TestCheckResourceAttr(basicDataByName, "folder_id", folderID),
					resource.TestCheckResourceAttr(basicDataByName, "name", keyName),
					resource.TestCheckResourceAttr(basicDataByName, "description", keyDesc),
					resource.TestCheckResourceAttr(basicDataByName, "deletion_protection", "false"),
					resource.TestCheckResourceAttr(basicDataByName, "labels.%", "2"),
					resource.TestCheckResourceAttr(basicDataByName, "labels.key1", "value1"),
					resource.TestCheckResourceAttr(basicDataByName, "labels.key2", "value2"),
					statecheck.CheckCreatedAtAttr(basicDataByName),
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

func (suite testSuite) testAccDataSourceKmsSymmetricKeyExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ds, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if ds.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		config, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}

		client := kmssdk.NewSymmetricKeyClient(config.SDK)

		found, err := client.Get(context.Background(), &kms.GetSymmetricKeyRequest{
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
