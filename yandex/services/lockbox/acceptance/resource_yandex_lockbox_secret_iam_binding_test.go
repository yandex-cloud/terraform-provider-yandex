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
	"github.com/yandex-cloud/go-genproto/yandex/cloud/lockbox/v1"
	lockboxsdk "github.com/yandex-cloud/go-sdk/services/lockbox/v1"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
	svc_lockbox "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/lockbox/pluginsdk"
)

const lockboxSecretResource = "yandex_lockbox_secret.test-secret"

func importLockboxSecretIDFunc(lockboxSecret *lockbox.Secret, role string) func(*terraform.State) (string, error) {
	return func(s *terraform.State) (string, error) {
		return lockboxSecret.Id + "," + role, nil
	}
}

func TestAccLockboxSecretIamBinding_basic(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	var lockboxSecret lockbox.Secret
	lockboxSecretName := acctest.RandomWithPrefix("tf-lockbox-secret")

	role := "viewer"
	userID := "system:allUsers"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		Steps: []resource.TestStep{
			{
				Config: testAccLockboxSecretIamBindingBasic(lockboxSecretName, role, userID),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckLockboxSecretExists(lockboxSecretResource, &lockboxSecret),
					suite.testAccCheckLockboxSecretIam(lockboxSecretResource, role, []string{userID}),
				),
			},
			{
				ResourceName:                         "yandex_lockbox_secret_iam_binding.viewer",
				ImportStateIdFunc:                    importLockboxSecretIDFunc(&lockboxSecret, role),
				ImportState:                          true,
				ImportStateVerifyIdentifierAttribute: "secret_id",
			},
		},
	})
}

func TestAccLockboxSecretIamBinding_remove(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	var lockboxSecret lockbox.Secret
	secretName := acctest.RandomWithPrefix("tf-lockbox-secret")

	role := "viewer"
	userID := "system:allUsers"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		Steps: []resource.TestStep{
			// Prepare data source
			{
				Config: testAccLockboxSecret(secretName),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckLockboxSecretExists(lockboxSecretResource, &lockboxSecret),
					suite.testAccCheckLockboxSecretEmptyIam(lockboxSecretResource),
				),
			},
			// Apply IAM bindings
			{
				Config: testAccLockboxSecretIamBindingBasic(secretName, role, userID),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckLockboxSecretIam(lockboxSecretResource, role, []string{userID}),
				),
			},
			// Remove the bindings
			{
				Config: testAccLockboxSecret(secretName),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckLockboxSecretEmptyIam(lockboxSecretResource),
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

func (suite testSuite) testAccCheckLockboxSecretEmptyIam(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		bindings, err := suite.getLockboxSecretResourceAccessBindings(s, resourceName)
		if err != nil {
			return err
		}

		if len(bindings) == 0 {
			return nil
		}

		return fmt.Errorf("binding found but expected empty for %s", resourceName)
	}
}

func (suite testSuite) testAccCheckLockboxSecretIam(resourceName, role string, members []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		bindings, err := suite.getLockboxSecretResourceAccessBindings(s, resourceName)
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

func (suite testSuite) getLockboxSecretResourceAccessBindings(s *terraform.State, resourceName string) ([]*access.AccessBinding, error) {
	config, configErr := suite.Config()
	if configErr != nil {
		return nil, configErr
	}

	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("can't find %s in state", resourceName)
	}

	return svc_lockbox.GetLockboxSecretAccessBindings(config.Context(), config, rs.Primary.ID)
}

func (suite testSuite) testAccCheckLockboxSecretExists(name string, lockboxSecret *lockbox.Secret) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		config, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}

		client := lockboxsdk.NewSecretClient(config.SDK)

		found, err := client.Get(context.Background(), &lockbox.GetSecretRequest{
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
