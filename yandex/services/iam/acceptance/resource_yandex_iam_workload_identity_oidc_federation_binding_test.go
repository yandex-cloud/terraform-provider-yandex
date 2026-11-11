package acceptance

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/acceptanceenv"
	rmfixtures "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/resourcemanager/testfixtures"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1/workload/oidc"
	oidcsdk "github.com/yandex-cloud/go-sdk/v2/services/iam/v1/workload/oidc"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
	svc_iam "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/iam/pluginsdk"
)

func importFederationIDFunc(federation *oidc.Federation, role string) func(*terraform.State) (string, error) {
	return func(s *terraform.State) (string, error) {
		return federation.Id + "," + role, nil
	}
}

func TestAccIAMWorkloadIdentityOidcFederationIamBinding(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	var federation oidc.Federation
	federationName := acctest.RandomWithPrefix("tf-test")
	cloudID := acceptanceenv.CloudID()
	userID := acceptanceenv.UserID1()
	role := "viewer"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkloadIdentityOidcFederationIamBinding_basic(cloudID, federationName, role, userID),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckWorkloadIdentityOidcFederationExistsWithID("yandex_iam_workload_identity_oidc_federation.test_federation", &federation),
					suite.testAccCheckWorkloadIdentityOidcFederationIam("yandex_iam_workload_identity_oidc_federation.test_federation", role, []string{"userAccount:" + userID}),
				),
			},
			{
				ResourceName:                         "yandex_iam_workload_identity_oidc_federation_iam_binding.foo",
				ImportStateIdFunc:                    importFederationIDFunc(&federation, role),
				ImportState:                          true,
				ImportStateVerifyIdentifierAttribute: "federation_id",
			},
		},
	})
}

//revive:disable:var-naming
func testAccWorkloadIdentityOidcFederationIamBinding_basic(cloudID, federationName, role, userID string) string {
	prerequisiteMembership, deps := rmfixtures.CloudMemberRole(cloudID, userID)
	return prerequisiteMembership + fmt.Sprintf(`
resource "yandex_iam_workload_identity_oidc_federation" "test_federation" {
  name        = "%s"
  disabled    = false
  audiences   = ["aud"]
  issuer      = "https://test-issuer.example.com"
  jwks_url    = "https://test-issuer.example.com/updated_jwks"
}

resource "yandex_iam_workload_identity_oidc_federation_iam_binding" "foo" {
  federation_id = "${yandex_iam_workload_identity_oidc_federation.test_federation.id}"
  role               = "%s"
  members            = ["userAccount:%s"]

  depends_on = [%s]
}
`, federationName, role, userID, deps)
}

func (suite testSuite) testAccCheckWorkloadIdentityOidcFederationExistsWithID(n string, f *oidc.Federation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}

		client := oidcsdk.NewFederationClient(config.SDK)

		found, err := client.Get(context.Background(), &oidc.GetFederationRequest{
			FederationId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("workload identity OIDC federation not found")
		}

		*f = *found

		return nil
	}
}

func (suite testSuite) testAccCheckWorkloadIdentityOidcFederationIam(resourceName, role string, members []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("can't find %s in state", resourceName)
		}

		bindings, err := svc_iam.GetWorkloadIdentityOidcFederationAccessBindings(config.Context(), config, rs.Primary.ID)
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

		return fmt.Errorf("Binding found but expected members is %v, got %v", members, roleMembers)
	}
}
