package acceptance

import (
	"context"
	"fmt"
	"testing"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/acceptanceenv"

	statecheck "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance/statecheck"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1/workload/oidc"
	oidcsdk "github.com/yandex-cloud/go-sdk/v2/services/iam/v1/workload/oidc"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
)

func TestAccDataSourceIAMWorkloadIdentityOidcFederationById(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	federationName := "wlif" + acctest.RandString(10)
	folderID := acceptanceenv.FolderID()
	resourceName := "data.yandex_iam_workload_identity_oidc_federation.test"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckWorkloadIdentityOidcFederationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkloadIdentityOidcFederationByIdConfig(federationName, folderID),
				Check: resource.ComposeTestCheckFunc(
					statecheck.CheckResourceIDField(resourceName, "federation_id"),
					suite.testAccCheckWorkloadIdentityOidcFederationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", federationName),
					resource.TestCheckResourceAttr(resourceName, "folder_id", folderID),
					resource.TestCheckResourceAttr(resourceName, "description", "test federation description"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "audiences.#"),
					resource.TestCheckResourceAttr(resourceName, "audiences.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "audiences.*", "aud1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "audiences.*", "aud2"),
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
	suite := testSuite{acceptance.New(t)}
	federationName := "wlif" + acctest.RandString(10)
	folderID := acceptanceenv.FolderID()
	resourceName := "data.yandex_iam_workload_identity_oidc_federation.test"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             suite.testAccCheckWorkloadIdentityOidcFederationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkloadIdentityOidcFederationByNameConfig(federationName, folderID),
				Check: resource.ComposeTestCheckFunc(
					statecheck.CheckResourceIDField(resourceName, "federation_id"),
					suite.testAccCheckWorkloadIdentityOidcFederationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", federationName),
					resource.TestCheckResourceAttr(resourceName, "folder_id", folderID),
					resource.TestCheckResourceAttr(resourceName, "description", "test federation description"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "audiences.#"),
					resource.TestCheckResourceAttr(resourceName, "audiences.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "audiences.*", "aud1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "audiences.*", "aud2"),
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

func (suite testSuite) testAccCheckWorkloadIdentityOidcFederationExists(r string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("not found: %s", r)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}
		config, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}

		client := oidcsdk.NewFederationClient(config.SDK)

		_, err := client.Get(context.Background(), &oidc.GetFederationRequest{
			FederationId: rs.Primary.ID,
		})

		return err
	}
}

func (suite testSuite) testAccCheckWorkloadIdentityOidcFederationDestroy(s *terraform.State) error {
	config, configErr := suite.Config()
	if configErr != nil {
		return configErr
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_iam_workload_identity_oidc_federation" {
			continue
		}

		client := oidcsdk.NewFederationClient(config.SDK)

		_, err := client.Get(context.Background(), &oidc.GetFederationRequest{
			FederationId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("WLI OIDC federation still exists")
		}
	}

	return nil
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
