package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1"
)

const (
	organizationManagerUserSshKey = "yandex_organizationmanager_user_ssh_key.this"
	userSSHKeyData                = "restrict,pty ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIAK37Oy9P11LbYqdEVfQlMMtfjkMeGZojICCVWLj0Pmt comment"
	userSSHKeyName                = "user_ssh_key_name"
	userSSHKeyExpiresAt           = "2099-11-11T22:33:44Z"
)

func init() {
	resource.AddTestSweepers("yandex_organizationmanager_user_ssh_key", &resource.Sweeper{
		Name:         "yandex_organizationmanager_user_ssh_key",
		F:            testSweepUserSshKeys,
		Dependencies: []string{},
	})
}

func testSweepUserSshKeyOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexOrganizationManagerUserSshKeyDefaultTimeout)
	defer cancel()

	op, err := conf.sdk.OrganizationManager().UserSshKey().Delete(ctx, &organizationmanager.DeleteUserSshKeyRequest{
		UserSshKeyId: id,
	})

	return handleSweepOperation(ctx, conf, op, err)
}

func testSweepUserSshKeys(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &organizationmanager.ListUserSshKeysRequest{
		OrganizationId: getExampleOrganizationID(),
		SubjectId:      getExampleUserID1(),
	}
	it := conf.sdk.OrganizationManager().UserSshKey().UserSshKeyIterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepWithRetry(testSweepUserSshKeyOnce, conf, "UserSshKey", id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep UserSshKey %q", id))
		}
	}

	return result.ErrorOrNil()
}

func TestAccOrganizationManagerUserSshKey(t *testing.T) {
	t.Parallel()
	organizationID := getExampleOrganizationID()
	subjectID := getExampleUserID1()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckUserSshKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigOrganizationManagerUserSshKey(organizationID, subjectID, userSSHKeyData, userSSHKeyName, userSSHKeyExpiresAt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(organizationManagerUserSshKey, "organization_id", organizationID),
					resource.TestCheckResourceAttr(organizationManagerUserSshKey, "subject_id", subjectID),
					resource.TestCheckResourceAttr(organizationManagerUserSshKey, "data", userSSHKeyData),
					resource.TestCheckResourceAttr(organizationManagerUserSshKey, "name", userSSHKeyName),
					resource.TestCheckResourceAttr(organizationManagerUserSshKey, "expires_at", userSSHKeyExpiresAt),
					resource.TestCheckResourceAttrSet(organizationManagerUserSshKey, "id"),
					resource.TestCheckResourceAttrSet(organizationManagerUserSshKey, "fingerprint"),
					resource.TestCheckResourceAttrSet(organizationManagerUserSshKey, "created_at"),
				),
			},
		},
	})
}

func testAccCheckUserSshKeyDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_organizationmanager_user_ssh_key" {
			continue
		}

		_, err := config.sdk.OrganizationManager().UserSshKey().Get(context.Background(), &organizationmanager.GetUserSshKeyRequest{
			UserSshKeyId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("UserSshKey still exists")
		}
	}

	return nil
}

func testAccConfigOrganizationManagerUserSshKey(organizationID string, subjectID string, userSSHKeyData string, userSSHKeyName string, userSSHKeyExpiresAt string) string {
	return fmt.Sprintf(`
resource "yandex_organizationmanager_user_ssh_key" "this" {
  organization_id = "%s"
  subject_id      = "%s"
  data            = "%s"
  name            = "%s"
  expires_at      = "%s"
}
`, organizationID, subjectID, userSSHKeyData, userSSHKeyName, userSSHKeyExpiresAt)
}
