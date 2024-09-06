package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceIAMWorkloadIdentityOidcFederationById(t *testing.T) {
	federationName := "wlif" + acctest.RandString(10)
	folderID := getExampleFolderID()
	resourceName := "data.yandex_iam_workload_identity_oidc_federation.test"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckWorkloadIdentityOidcFederationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkloadIdentityOidcFederationByIdConfig(federationName, folderID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceIDField(resourceName, "federation_id"),
					testAccCheckWorkloadIdentityOidcFederationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", federationName),
					resource.TestCheckResourceAttr(resourceName, "folder_id", folderID),
					resource.TestCheckResourceAttr(resourceName, "description", "test federation description"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "audiences.0", "aud1"),
					resource.TestCheckResourceAttr(resourceName, "audiences.1", "aud2"),
					resource.TestCheckResourceAttr(resourceName, "issuer", "https://test-issuer.example.com"),
					resource.TestCheckResourceAttr(resourceName, "jwks_url", "https://test-issuer.example.com/jwks"),
					resource.TestCheckResourceAttr(resourceName, "labels.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "labels.key2", "value2"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
				),
			},
		},
	})
}

func TestAccDataSourceIAMWorkloadIdentityOidcFederationByName(t *testing.T) {
	federationName := "wlif" + acctest.RandString(10)
	folderID := getExampleFolderID()
	resourceName := "data.yandex_iam_workload_identity_oidc_federation.test"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckWorkloadIdentityOidcFederationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkloadIdentityOidcFederationByNameConfig(federationName, folderID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceIDField(resourceName, "federation_id"),
					testAccCheckWorkloadIdentityOidcFederationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", federationName),
					resource.TestCheckResourceAttr(resourceName, "folder_id", folderID),
					resource.TestCheckResourceAttr(resourceName, "description", "test federation description"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "audiences.0", "aud1"),
					resource.TestCheckResourceAttr(resourceName, "audiences.1", "aud2"),
					resource.TestCheckResourceAttr(resourceName, "issuer", "https://test-issuer.example.com"),
					resource.TestCheckResourceAttr(resourceName, "jwks_url", "https://test-issuer.example.com/jwks"),
					resource.TestCheckResourceAttr(resourceName, "labels.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "labels.key2", "value2"),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
				),
			},
		},
	})
}

func testAccWorkloadIdentityOidcFederationByIdConfig(federationName, folderId string) string {
	return fmt.Sprintf(`
data "yandex_iam_workload_identity_oidc_federation" "test" {
  federation_id = "${yandex_iam_workload_identity_oidc_federation.acceptance.id}"
}

resource "yandex_iam_workload_identity_oidc_federation" "acceptance" {
  name        = "%s"
  folder_id   = "%s"
  description = "test federation description"
  disabled    = false
  audiences   = ["aud1","aud2"]
  issuer      = "https://test-issuer.example.com"
  jwks_url    = "https://test-issuer.example.com/jwks"
  labels      = {
    key1 = "value1"
    key2 = "value2"
  }
}
`, federationName, folderId)
}

func testAccWorkloadIdentityOidcFederationByNameConfig(federationName, folderId string) string {
	return fmt.Sprintf(`
data "yandex_iam_workload_identity_oidc_federation" "test" {
  name = "${yandex_iam_workload_identity_oidc_federation.acceptance.name}"
}

resource "yandex_iam_workload_identity_oidc_federation" "acceptance" {
  name        = "%s"
  folder_id   = "%s"
  description = "test federation description"
  disabled    = false
  audiences   = ["aud1","aud2"]
  issuer      = "https://test-issuer.example.com"
  jwks_url    = "https://test-issuer.example.com/jwks"
  labels      = {
    key1 = "value1"
    key2 = "value2"
  }
}
`, federationName, folderId)
}
