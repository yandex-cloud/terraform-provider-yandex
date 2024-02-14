package yandex

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

const (
	organizationManagerUserSshKeyData = "data.yandex_organizationmanager_user_ssh_key.this_data"
)

func TestAccDataSourceOrganizationManagerUserSshKey(t *testing.T) {
	t.Parallel()
	organizationID := getExampleOrganizationID()
	subjectID := getExampleUserID1()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
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
