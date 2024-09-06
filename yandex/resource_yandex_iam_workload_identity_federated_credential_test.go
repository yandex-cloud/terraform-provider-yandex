package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1/workload"
)

func TestAccIAMWorkloadIdentityFederatedCredential(t *testing.T) {
	federationName := "wlif" + acctest.RandString(10)
	saName := "sa" + acctest.RandString(10)
	folderID := getExampleFolderID()
	resourceName := "yandex_iam_workload_identity_federated_credential.acceptance"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckWorkloadIdentityFederatedCredentialDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkloadIdentityFederatedCredentialConfig(federationName, folderID, saName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkloadIdentityFederatedCredentialExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "service_account_id"),
					resource.TestCheckResourceAttrSet(resourceName, "federation_id"),
					resource.TestCheckResourceAttr(resourceName, "external_subject_id", "test_external_subject_id"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
				),
			},
		},
	})
}

func testAccCheckWorkloadIdentityFederatedCredentialExists(r string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("not found: %s", r)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}
		config := testAccProvider.Meta().(*Config)

		_, err := config.sdk.Workload().FederatedCredential().Get(context.Background(), &workload.GetFederatedCredentialRequest{
			FederatedCredentialId: rs.Primary.ID,
		})

		return err
	}
}

func testAccCheckWorkloadIdentityFederatedCredentialDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_iam_workload_identity_federated_credential" {
			continue
		}

		_, err := config.sdk.Workload().FederatedCredential().Get(context.Background(), &workload.GetFederatedCredentialRequest{
			FederatedCredentialId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("WLI federated credential still exists")
		}
	}

	return nil
}

func testAccWorkloadIdentityFederatedCredentialConfig(federationName, folderId, saName string) string {
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

resource "yandex_iam_workload_identity_federated_credential" "acceptance" {
  service_account_id  = "${yandex_iam_service_account.acceptance.id}"
  federation_id       = "${yandex_iam_workload_identity_oidc_federation.acceptance.id}"
  external_subject_id = "test_external_subject_id"
}
`, federationName, folderId, saName)
}
