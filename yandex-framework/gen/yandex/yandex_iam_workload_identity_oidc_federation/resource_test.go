package yandex_iam_workload_identity_oidc_federation_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1/workload/oidc"
	oidcsdk "github.com/yandex-cloud/go-sdk/services/iam/v1/workload/oidc"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
)

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestAccIAMWorkloadIdentityOidcFederation_UpgradeFromSDKv2(t *testing.T) {
	federationName := "wlif" + acctest.RandString(10)
	folderID := test.GetExampleFolderID()
	resourceName := "yandex_iam_workload_identity_oidc_federation.acceptance"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { test.AccPreCheck(t) },
		CheckDestroy: testAccCheckWorkloadIdentityOidcFederationDestroy,
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"yandex": {
						VersionConstraint: "0.150.0",
						Source:            "yandex-cloud/yandex",
					},
				},
				Config: testAccWorkloadIdentityOidcFederationConfig_basic(federationName, folderID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "federation_id"),
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
			{
				ProtoV6ProviderFactories: test.AccProviderFactories,
				Config:                   testAccWorkloadIdentityOidcFederationConfig_basic(federationName, folderID),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccIAMWorkloadIdentityOidcFederation(t *testing.T) {
	federationName := "wlif" + acctest.RandString(10)
	folderID := test.GetExampleFolderID()
	resourceName := "yandex_iam_workload_identity_oidc_federation.acceptance"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckWorkloadIdentityOidcFederationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkloadIdentityOidcFederationConfig_basic(federationName, folderID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkloadIdentityOidcFederationExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "federation_id"),
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
			{
				Config: testAccWorkloadIdentityOidcFederationConfig_update("new-federation-name", folderID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkloadIdentityOidcFederationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", "new-federation-name"),
					resource.TestCheckResourceAttr(resourceName, "folder_id", folderID),
					resource.TestCheckResourceAttr(resourceName, "description", "updated federation description"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "audiences.0", "aud1"),
					resource.TestCheckResourceAttr(resourceName, "audiences.1", "aud3"),
					resource.TestCheckResourceAttr(resourceName, "issuer", "https://test-issuer.example.com"),
					resource.TestCheckResourceAttr(resourceName, "jwks_url", "https://test-issuer.example.com/updated_jwks"),
					resource.TestCheckResourceAttr(resourceName, "labels.key1", "new_value1"),
					resource.TestCheckResourceAttr(resourceName, "labels.key3", "value3"),
				),
			},
		},
	})
}

func testAccCheckWorkloadIdentityOidcFederationExists(r string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("not found: %s", r)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}
		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

		_, err := oidcsdk.NewFederationClient(config.SDKv2).Get(context.Background(), &oidc.GetFederationRequest{
			FederationId: rs.Primary.ID,
		})

		return err
	}
}

func testAccCheckWorkloadIdentityOidcFederationDestroy(s *terraform.State) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_iam_workload_identity_oidc_federation" {
			continue
		}

		_, err := oidcsdk.NewFederationClient(config.SDKv2).Get(context.Background(), &oidc.GetFederationRequest{
			FederationId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("WLI OIDC federation still exists")
		}
	}

	return nil
}

func testAccWorkloadIdentityOidcFederationConfig_basic(federationName, folderId string) string {
	return fmt.Sprintf(`
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

func testAccWorkloadIdentityOidcFederationConfig_update(federationName, folderId string) string {
	return fmt.Sprintf(`
resource "yandex_iam_workload_identity_oidc_federation" "acceptance" {
  name        = "%s"
  folder_id   = "%s"
  description = "updated federation description"
  disabled    = true
  audiences   = ["aud1","aud3"]
  issuer      = "https://test-issuer.example.com"
  jwks_url    = "https://test-issuer.example.com/updated_jwks"
  labels      = {
    key1 = "new_value1"
    key3 = "value3"
  }
}
`, federationName, folderId)
}
