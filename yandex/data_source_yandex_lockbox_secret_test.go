package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/lockbox/v1"
)

func TestAccDataSourceLockboxSecret_byID(t *testing.T) {
	secretName := "a" + acctest.RandString(10)
	secretDesc := "Terraform Test"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckYandexLockboxSecretAllDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSecretConfig(secretName, secretDesc, true),
				Check:  testAccDataSourceSecretCheck(secretName, secretDesc),
			},
		},
	})
}

func TestAccDataSourceLockboxSecret__byName(t *testing.T) {
	secretName := "a" + acctest.RandString(10)
	secretDesc := "Terraform Test"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckYandexLockboxSecretAllDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceSecretConfig(secretName, secretDesc, false),
				Check:  testAccDataSourceSecretCheck(secretName, secretDesc),
			},
		},
	})
}

func testAccDataSourceSecretCheck(secretName string, secretDesc string) resource.TestCheckFunc {
	basicData := "data.yandex_lockbox_secret.basic_secret"
	folderID := getExampleFolderID()
	return resource.ComposeTestCheckFunc(
		testAccDataSourceLockboxSecretExists(basicData),
		testAccCheckResourceIDField(basicData, "secret_id"),
		resource.TestCheckResourceAttr(basicData, "folder_id", folderID),
		resource.TestCheckResourceAttr(basicData, "name", secretName),
		resource.TestCheckResourceAttr(basicData, "description", secretDesc),
		resource.TestCheckResourceAttr(basicData, "deletion_protection", "false"),
		resource.TestCheckResourceAttr(basicData, "labels.%", "2"),
		resource.TestCheckResourceAttr(basicData, "labels.key1", "value1"),
		resource.TestCheckResourceAttr(basicData, "labels.key2", "value2"),
		resource.TestCheckResourceAttr(basicData, "status",
			lockbox.Secret_Status_name[int32(lockbox.Secret_ACTIVE)]),
		testAccCheckCreatedAtAttr(basicData),
		//showResourceAttributes(basicData),
	)
}

func testAccLockboxSecretResourceConfig(name, desc string) string {
	return fmt.Sprintf(`
resource "yandex_lockbox_secret" "basic_secret" {
  name        = "%v"
  description = "%v"
  labels      = {
    key1 = "value1"
    key2 = "value2"
  }
}
`, name, desc)
}

const secretLockboxDataByByIDConfig = `
data "yandex_lockbox_secret" "basic_secret" {
  secret_id = yandex_lockbox_secret.basic_secret.id
}
`

const secretLockboxDataByNameConfig = `
data "yandex_lockbox_secret" "basic_secret" {
  name = yandex_lockbox_secret.basic_secret.name
}
`

func testAccDataSourceSecretConfig(secretName string, secretDescription string, withDataID bool) string {
	if withDataID {
		return testAccLockboxSecretResourceConfig(secretName, secretDescription) + secretLockboxDataByByIDConfig
	} else {
		return testAccLockboxSecretResourceConfig(secretName, secretDescription) + secretLockboxDataByNameConfig
	}
}

func testAccDataSourceLockboxSecretExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ds, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if ds.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.LockboxSecret().Secret().Get(context.Background(), &lockbox.GetSecretRequest{
			SecretId: ds.Primary.ID,
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
