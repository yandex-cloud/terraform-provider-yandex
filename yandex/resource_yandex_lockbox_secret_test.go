package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/lockbox/v1"
)

func init() {
	resource.AddTestSweepers("yandex_lockbox_secret", &resource.Sweeper{
		Name: "yandex_lockbox_secret",
		F:    testSweepLockboxSecret,
	})
}

func TestAccLockboxSecret_basic(t *testing.T) {
	secretName := "a" + acctest.RandString(10)
	secretDesc := "Terraform Test"
	folderID := getExampleFolderID()
	basicResource := "yandex_lockbox_secret.basic_secret"
	basicResourceID := ""
	minimalResource := "yandex_lockbox_secret.minimal_secret"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckYandexLockboxSecretAllDestroyed,
		Steps: []resource.TestStep{
			{
				// Create secret
				Config: testAccLockboxSecretBasic(secretName, secretDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYandexLockboxResourceExists(basicResource, &basicResourceID),
					resource.TestCheckResourceAttr(basicResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(basicResource, "name", secretName),
					resource.TestCheckResourceAttr(basicResource, "description", secretDesc),
					resource.TestCheckResourceAttr(basicResource, "deletion_protection", "true"),
					resource.TestCheckResourceAttr(basicResource, "labels.%", "2"),
					resource.TestCheckResourceAttr(basicResource, "labels.key1", "value1"),
					resource.TestCheckResourceAttr(basicResource, "labels.key2", "value2"),
					resource.TestCheckResourceAttr(basicResource, "status",
						lockbox.Secret_Status_name[int32(lockbox.Secret_ACTIVE)]),
					testAccCheckCreatedAtAttr(basicResource),
				),
			},
			{
				// Update secret
				Config: testAccLockboxSecretBasicModified(secretName+"-modified", secretDesc+" edited"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYandexLockboxResourceExists(basicResource, nil),
					resource.TestCheckResourceAttr(basicResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(basicResource, "name", secretName+"-modified"),
					resource.TestCheckResourceAttr(basicResource, "description", secretDesc+" edited"),
					resource.TestCheckResourceAttr(basicResource, "deletion_protection", "false"),
					resource.TestCheckResourceAttr(basicResource, "labels.%", "3"),
					resource.TestCheckResourceAttr(basicResource, "labels.key1", "value10"),
					resource.TestCheckResourceAttr(basicResource, "labels.key3", "value30"),
					resource.TestCheckResourceAttr(basicResource, "labels.key4", "value40"),
					resource.TestCheckResourceAttr(basicResource, "status",
						lockbox.Secret_Status_name[int32(lockbox.Secret_ACTIVE)]),
					testAccCheckCreatedAtAttr(basicResource),
				),
			},
			{
				// Add new secret (delete previous one)
				Config: testAccLockboxSecretMinimal(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYandexLockboxResourceDoesNotExist(basicResource),
					func(s *terraform.State) error {
						return testAccCheckYandexLockboxSecretDestroyed(basicResourceID)
					},
					testAccCheckYandexLockboxResourceExists(minimalResource, nil),
					resource.TestCheckResourceAttr(minimalResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(minimalResource, "name", ""),
					resource.TestCheckResourceAttr(minimalResource, "description", ""),
					resource.TestCheckResourceAttr(minimalResource, "deletion_protection", "false"),
					resource.TestCheckResourceAttr(minimalResource, "status",
						lockbox.Secret_Status_name[int32(lockbox.Secret_ACTIVE)]),
					testAccCheckCreatedAtAttr(minimalResource),
				),
			},
		},
	})
}

func TestAccLockboxSecret_kms(t *testing.T) {
	secretName := "a" + acctest.RandString(10)
	folderID := getExampleFolderID()
	secretResource := "yandex_lockbox_secret.kms_secret"
	resourceID := ""
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckYandexLockboxSecretAllDestroyed,
		Steps: []resource.TestStep{
			{
				// Create secret
				Config: testAccLockboxSecretWithKmsKey(secretName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYandexLockboxResourceExists(secretResource, &resourceID), // sets resourceID
					resource.TestCheckResourceAttrSet(secretResource, "kms_key_id"),
					resource.TestCheckResourceAttr(secretResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(secretResource, "deletion_protection", "false"),
					resource.TestCheckResourceAttr(secretResource, "status",
						lockbox.Secret_Status_name[int32(lockbox.Secret_ACTIVE)]),
					testAccCheckCreatedAtAttr(secretResource),
				),
			},
			{
				// Update kms_key_id will create new secret
				Config: testAccLockboxSecretWithKmsKeyModified(secretName),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						return testAccCheckYandexLockboxSecretDestroyed(resourceID)
					},
					testAccCheckYandexLockboxResourceExists(secretResource, &resourceID), // checks that now resourceID is different
					resource.TestCheckResourceAttrSet(secretResource, "kms_key_id"),
					resource.TestCheckResourceAttr(secretResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(secretResource, "deletion_protection", "false"),
					resource.TestCheckResourceAttr(secretResource, "status",
						lockbox.Secret_Status_name[int32(lockbox.Secret_ACTIVE)]),
					testAccCheckCreatedAtAttr(secretResource),
				),
			},
		},
	})
}

func TestAccLockboxSecret_passwordPayloadSpec(t *testing.T) {
	secretName := "a" + acctest.RandString(10)
	secretDesc := "Terraform Test With Password Payload Spec"
	folderID := getExampleFolderID()
	basicResource := "yandex_lockbox_secret.basic_secret"
	basicResourceID := ""
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckYandexLockboxSecretAllDestroyed,
		Steps: []resource.TestStep{
			{
				// Create secret
				Config: testAccLockboxSecretWithPasswordPayloadSpec(secretName, secretDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYandexLockboxResourceExists(basicResource, &basicResourceID),
					resource.TestCheckResourceAttr(basicResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(basicResource, "name", secretName),
					resource.TestCheckResourceAttr(basicResource, "description", secretDesc),
					resource.TestCheckResourceAttr(basicResource, "deletion_protection", "false"),
					resource.TestCheckNoResourceAttr(basicResource, "labels.%"),
					resource.TestCheckResourceAttr(basicResource, "status",
						lockbox.Secret_Status_name[int32(lockbox.Secret_ACTIVE)]),
					resource.TestCheckResourceAttr(basicResource, "password_payload_specification.#", "1"),
					resource.TestCheckResourceAttr(basicResource, "password_payload_specification.0.password_key", "password_key"),
					resource.TestCheckResourceAttr(basicResource, "password_payload_specification.0.length", "10"),
					testAccCheckCreatedAtAttr(basicResource),
				),
			},
			{
				// Update to basic secret
				Config: testAccLockboxSecretBasicModified(secretName+"-modified", secretDesc+" edited"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYandexLockboxResourceExists(basicResource, nil),
					resource.TestCheckResourceAttr(basicResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(basicResource, "name", secretName+"-modified"),
					resource.TestCheckResourceAttr(basicResource, "description", secretDesc+" edited"),
					resource.TestCheckResourceAttr(basicResource, "deletion_protection", "false"),
					resource.TestCheckResourceAttr(basicResource, "labels.%", "3"),
					resource.TestCheckResourceAttr(basicResource, "labels.key1", "value10"),
					resource.TestCheckResourceAttr(basicResource, "labels.key3", "value30"),
					resource.TestCheckResourceAttr(basicResource, "labels.key4", "value40"),
					resource.TestCheckResourceAttr(basicResource, "status",
						lockbox.Secret_Status_name[int32(lockbox.Secret_ACTIVE)]),
					resource.TestCheckResourceAttr(basicResource, "password_payload_specification.#", "0"),
					testAccCheckCreatedAtAttr(basicResource),
				),
			},
		},
	})
}

func testAccLockboxSecretBasic(name, desc string) string {
	return fmt.Sprintf(`
resource "yandex_lockbox_secret" "basic_secret" {
  name        = "%v"
  description = "%v"
  labels      = {
    key1 = "value1"
    key2 = "value2"
  }
  deletion_protection = true
}
`, name, desc)
}

func testAccLockboxSecretBasicModified(name, desc string) string {
	return fmt.Sprintf(`
resource "yandex_lockbox_secret" "basic_secret" {
  name                = "%v"
  description         = "%v"
  labels              = {
    key1 = "value10"
    key3 = "value30"
    key4 = "value40"
  }
  deletion_protection = false
}
`, name, desc)
}

func testAccLockboxSecretMinimal() string {
	return `
resource "yandex_lockbox_secret" "minimal_secret" {
}
`
}

func testAccLockboxSecretWithKmsKey(name string) string {
	return fmt.Sprintf(`
resource "yandex_lockbox_secret" "kms_secret" {
  name       = "%v"
  kms_key_id = yandex_kms_symmetric_key.some_key1.id
}
%s
`, name, testAccLockboxSecretKmsKeys())
}

func testAccLockboxSecretWithKmsKeyModified(name string) string {
	return fmt.Sprintf(`
resource "yandex_lockbox_secret" "kms_secret" {
  name       = "%v"
  kms_key_id = yandex_kms_symmetric_key.some_key2.id
}
%s
`, name, testAccLockboxSecretKmsKeys())
}

func testAccLockboxSecretKmsKeys() string {
	return `
resource "yandex_kms_symmetric_key" "some_key1" {
}

resource "yandex_kms_symmetric_key" "some_key2" {
}
`
}

func testAccLockboxSecretWithPasswordPayloadSpec(name, desc string) string {
	return fmt.Sprintf(`
resource "yandex_lockbox_secret" "basic_secret" {
  name        = "%v"
  description = "%v"
  password_payload_specification {
	password_key = "password_key"
	length       = 10 
  }
}
`, name, desc)
}

// If idPtr is provided:
// - idPtr must be different from the resource ID (you can use it to check that the resource is different)
// - resource ID will be set to idPtr
func testAccCheckYandexLockboxResourceExists(r string, idPtr *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("not found resource: %s", r)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set for the resource: %s", r)
		}
		if idPtr != nil {
			if rs.Primary.ID == *idPtr {
				return fmt.Errorf("ID %s of resource %s is the same", rs.Primary.ID, r)
			}
			*idPtr = rs.Primary.ID
		}
		return nil
	}
}

func testAccCheckYandexLockboxResourceDoesNotExist(r string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[r]
		if ok {
			return fmt.Errorf("resource exists: %s", r)
		}
		return nil
	}
}

func testAccCheckYandexLockboxSecretAllDestroyed(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_lockbox_secret" {
			continue
		}
		if err := testAccCheckYandexLockboxSecretDestroyed(rs.Primary.ID); err != nil {
			return err
		}
	}
	return nil
}

func testAccCheckYandexLockboxSecretDestroyed(id string) error {
	config := testAccProvider.Meta().(*Config)
	_, err := config.sdk.LockboxSecret().Secret().Get(context.Background(), &lockbox.GetSecretRequest{
		SecretId: id,
	})
	if err == nil {
		return fmt.Errorf("LockboxSecret %s still exists", id)
	}
	return nil
}

func testSweepLockboxSecret(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &lockbox.ListSecretsRequest{FolderId: conf.FolderID}
	it := conf.sdk.LockboxSecret().Secret().SecretIterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepLockboxSecret(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep lockbox secret %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepLockboxSecret(conf *Config, id string) bool {
	return sweepWithRetry(sweepLockboxSecretOnce, conf, "Lockbox secret", id)
}

func sweepLockboxSecretOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexLockboxSecretDefaultTimeout)
	defer cancel()

	op, err := conf.sdk.LockboxSecret().Secret().Delete(ctx, &lockbox.DeleteSecretRequest{
		SecretId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}
