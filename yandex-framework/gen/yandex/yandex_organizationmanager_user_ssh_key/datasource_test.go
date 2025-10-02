package yandex_organizationmanager_user_ssh_key_test

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	"testing"
)

const (
	organizationManagerUserSshKeyData = "data.yandex_organizationmanager_user_ssh_key.this_data"
)

func TestAccDataSourceOrganizationManagerUserSshKey_UpgradeFromSDKv2(t *testing.T) {
	t.Parallel()
	organizationID := test.GetExampleOrganizationID()
	subjectID := test.GetExampleUserID1()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { test.AccPreCheck(t) },
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"yandex": {
						VersionConstraint: "0.150.0",
						Source:            "yandex-cloud/yandex",
					},
				},
				Config: testAccDataOrganizationManagerUserSshKey(organizationID, subjectID, userSSHKeyData, userSSHKeyName, userSSHKeyExpiresAt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(organizationManagerUserSshKeyData, "organization_id", organizationID),
					resource.TestCheckResourceAttr(organizationManagerUserSshKeyData, "subject_id", subjectID),
					resource.TestCheckResourceAttr(organizationManagerUserSshKeyData, "data", userSSHKeyData),
					resource.TestCheckResourceAttr(organizationManagerUserSshKeyData, "name", userSSHKeyName),
					resource.TestCheckResourceAttr(organizationManagerUserSshKeyData, "expires_at", userSSHKeyExpiresAt),
					resource.TestCheckResourceAttrSet(organizationManagerUserSshKeyData, "id"),
					resource.TestCheckResourceAttrSet(organizationManagerUserSshKeyData, "fingerprint"),
					resource.TestCheckResourceAttrSet(organizationManagerUserSshKeyData, "created_at"),
				),
			},
			{
				ProtoV6ProviderFactories: test.AccProviderFactories,
				Config:                   testAccDataOrganizationManagerUserSshKey(organizationID, subjectID, userSSHKeyData, userSSHKeyName, userSSHKeyExpiresAt),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccDataSourceOrganizationManagerUserSshKey(t *testing.T) {
	t.Parallel()
	organizationID := test.GetExampleOrganizationID()
	subjectID := test.GetExampleUserID1()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccDataOrganizationManagerUserSshKey(organizationID, subjectID, userSSHKeyData, userSSHKeyName, userSSHKeyExpiresAt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(organizationManagerUserSshKeyData, "organization_id", organizationID),
					resource.TestCheckResourceAttr(organizationManagerUserSshKeyData, "subject_id", subjectID),
					resource.TestCheckResourceAttr(organizationManagerUserSshKeyData, "data", userSSHKeyData),
					resource.TestCheckResourceAttr(organizationManagerUserSshKeyData, "name", userSSHKeyName),
					resource.TestCheckResourceAttr(organizationManagerUserSshKeyData, "expires_at", userSSHKeyExpiresAt),
					resource.TestCheckResourceAttrSet(organizationManagerUserSshKeyData, "id"),
					resource.TestCheckResourceAttrSet(organizationManagerUserSshKeyData, "fingerprint"),
					resource.TestCheckResourceAttrSet(organizationManagerUserSshKeyData, "created_at"),
				),
			},
		},
	})
}

func testAccDataOrganizationManagerUserSshKey(organizationID string, subjectID string, userSSHKeyData string, userSSHKeyName string, userSSHKeyExpiresAt string) string {
	return testAccConfigOrganizationManagerUserSshKey(organizationID, subjectID, userSSHKeyData, userSSHKeyName, userSSHKeyExpiresAt) + testAccDataConfigOrganizationManagerUserSshKey()
}

func testAccDataConfigOrganizationManagerUserSshKey() string {
	return `
data "yandex_organizationmanager_user_ssh_key" "this_data" {
  user_ssh_key_id = yandex_organizationmanager_user_ssh_key.this.id
}
`
}
