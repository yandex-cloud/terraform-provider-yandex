package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceIAMWorkloadIdentityFederatedCredential(t *testing.T) {
	federationName := "wlif" + acctest.RandString(10)
	saName := "sa" + acctest.RandString(10)
	folderID := getExampleFolderID()
	resourceName := "data.yandex_iam_workload_identity_federated_credential.test"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckWorkloadIdentityFederatedCredentialDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataWorkloadIdentityFederatedCredentialConfig(federationName, folderID, saName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceIDField(resourceName, "federated_credential_id"),
					resource.TestCheckResourceAttrSet(resourceName, "service_account_id"),
					resource.TestCheckResourceAttrSet(resourceName, "federation_id"),
					resource.TestCheckResourceAttr(resourceName, "external_subject_id", "test_external_subject_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
				),
			},
		},
	})
}

func testAccDataWorkloadIdentityFederatedCredentialConfig(federationName, folderId, saName string) string {
	return fmt.Sprintf(`
resource "yandex_iam_workload_identity_oidc_federation" "acceptance" {
  name        = "%s"
  folder_id   = "%s"
  description = "test federation description"
  disabled    = false
  audiences   = ["aud"]
  issuer      = "https://test-issuer.example.com"
  jwks_url    = "https://test-issuer.example.com/jwks"
}

resource "yandex_iam_service_account" "acceptance" {
  name        = "%s"
  description = "test sa description"
}

data "yandex_iam_workload_identity_federated_credential" "test" {
  federated_credential_id = "${yandex_iam_workload_identity_federated_credential.acceptance.id}"
}

resource "yandex_iam_workload_identity_federated_credential" "acceptance" {
  service_account_id  = "${yandex_iam_service_account.acceptance.id}"
  federation_id       = "${yandex_iam_workload_identity_oidc_federation.acceptance.id}"
  external_subject_id = "test_external_subject_id"
}
`, federationName, folderId, saName)
}
