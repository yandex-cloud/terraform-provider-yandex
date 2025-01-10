package yandex

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
	"github.com/yandex-cloud/go-genproto/yandex/cloud/lockbox/v1"
)

const lockboxSecretResource = "yandex_lockbox_secret.test-secret"

func importLockboxSecretIDFunc(lockboxSecret *lockbox.Secret, role string) func(*terraform.State) (string, error) {
	return func(s *terraform.State) (string, error) {
		return lockboxSecret.Id + " " + role, nil
	}
}

func TestAccLockboxSecretIamBinding_basic(t *testing.T) {
	var lockboxSecret lockbox.Secret
	lockboxSecretName := acctest.RandomWithPrefix("tf-lockbox-secret")

	role := "viewer"
	userID := "system:allUsers"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccLockboxSecretIamBindingBasic(lockboxSecretName, role, userID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLockboxSecretExists(lockboxSecretResource, &lockboxSecret),
					testAccCheckLockboxSecretIam(lockboxSecretResource, role, []string{userID}),
				),
			},
			{
				ResourceName:      "yandex_lockbox_secret_iam_binding.viewer",
				ImportStateIdFunc: importLockboxSecretIDFunc(&lockboxSecret, role),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLockboxSecretIamBinding_remove(t *testing.T) {
	var lockboxSecret lockbox.Secret
	secretName := acctest.RandomWithPrefix("tf-lockbox-secret")

	role := "viewer"
	userID := "system:allUsers"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			// Prepare data source
			{
				Config: testAccLockboxSecret(secretName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLockboxSecretExists(lockboxSecretResource, &lockboxSecret),
					testAccCheckLockboxSecretEmptyIam(lockboxSecretResource),
				),
			},
			// Apply IAM bindings
			{
				Config: testAccLockboxSecretIamBindingBasic(secretName, role, userID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLockboxSecretIam(lockboxSecretResource, role, []string{userID}),
				),
			},
			// Remove the bindings
			{
				Config: testAccLockboxSecret(secretName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLockboxSecretEmptyIam(lockboxSecretResource),
				),
			},
		},
	})
}

func testAccLockboxSecretIamBindingBasic(secretName, role, userID string) string {
	return fmt.Sprintf(`
resource "yandex_lockbox_secret" "test-secret" {
  name = "%s"
}

resource "yandex_lockbox_secret_iam_binding" "viewer" {
  secret_id = yandex_lockbox_secret.test-secret.id
  role      = "%s"
  members   = ["%s"]
}
`, secretName, role, userID)
}

func testAccLockboxSecret(secretName string) string {
	return fmt.Sprintf(`
resource "yandex_lockbox_secret" "test-secret" {
  name = "%s"
}
`, secretName)
}

func testAccCheckLockboxSecretEmptyIam(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		bindings, err := getLockboxSecretResourceAccessBindings(s, resourceName)
		if err != nil {
			return err
		}

		if len(bindings) == 0 {
			return nil
		}

		return fmt.Errorf("binding found but expected empty for %s", resourceName)
	}
}

func testAccCheckLockboxSecretIam(resourceName, role string, members []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		bindings, err := getLockboxSecretResourceAccessBindings(s, resourceName)
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

		return fmt.Errorf("binding found but expected members is %v, got %v", members, roleMembers)
	}
}

func getLockboxSecretResourceAccessBindings(s *terraform.State, resourceName string) ([]*access.AccessBinding, error) {
	config := testAccProvider.Meta().(*Config)

	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("can't find %s in state", resourceName)
	}

	return getLockboxSecretAccessBindings(config.Context(), config, rs.Primary.ID)
}

func testAccCheckLockboxSecretExists(name string, lockboxSecret *lockbox.Secret) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.LockboxSecret().Secret().Get(context.Background(), &lockbox.GetSecretRequest{
			SecretId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("lockbox secret not found")
		}

		*lockboxSecret = *found

		return nil
	}
}
