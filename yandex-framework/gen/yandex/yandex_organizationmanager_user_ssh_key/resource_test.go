package yandex_organizationmanager_user_ssh_key_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1"
	organizationmanagerv1sdk "github.com/yandex-cloud/go-sdk/services/organizationmanager/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
)

const (
	organizationManagerUserSshKey = "yandex_organizationmanager_user_ssh_key.this"
	userSSHKeyData                = "restrict,pty ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIAK37Oy9P11LbYqdEVfQlMMtfjkMeGZojICCVWLj0Pmt comment"
	userSSHKeyName                = "user_ssh_key_name"
	userSSHKeyExpiresAt           = "2099-11-11T22:33:44Z"
)

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestAccOrganizationManagerUserSshKey_UpgradeFromSDKv2(t *testing.T) {
	t.Parallel()
	organizationID := test.GetExampleOrganizationID()
	subjectID := test.GetExampleUserID1()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { test.AccPreCheck(t) },
		CheckDestroy: testAccCheckUserSshKeyDestroy,
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"yandex": {
						VersionConstraint: "0.150.0",
						Source:            "yandex-cloud/yandex",
					},
				},
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
			{
				ProtoV6ProviderFactories: test.AccProviderFactories,
				Config:                   testAccConfigOrganizationManagerUserSshKey(organizationID, subjectID, userSSHKeyData, userSSHKeyName, userSSHKeyExpiresAt),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccOrganizationManagerUserSshKey(t *testing.T) {
	t.Parallel()
	organizationID := test.GetExampleOrganizationID()
	subjectID := test.GetExampleUserID1()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckUserSshKeyDestroy,
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
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_organizationmanager_user_ssh_key" {
			continue
		}

		_, err := organizationmanagerv1sdk.NewUserSshKeyClient(config.SDKv2).Get(context.Background(), &organizationmanager.GetUserSshKeyRequest{
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
