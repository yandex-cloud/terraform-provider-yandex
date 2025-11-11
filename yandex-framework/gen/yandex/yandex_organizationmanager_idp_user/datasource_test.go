package yandex_organizationmanager_idp_user_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	idp "github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1/idp"
	idpsdk "github.com/yandex-cloud/go-sdk/services/organizationmanager/v1/idp"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAccDataSourceOrganizationManagerIdpUser_byID(t *testing.T) {
	userpoolName := acctest.RandomWithPrefix("tf-userpool")
	userNameBase := acctest.RandomWithPrefix("tf-user")
	organizationID := test.GetExampleOrganizationID()
	testSubdomain := acctest.RandomWithPrefix("tf-acc-test-subdomain")
	username := fmt.Sprintf("%s@%s.idp.yandexcloud.net", userNameBase, testSubdomain)
	password := acctest.RandomWithPrefix("Random195!-")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckIdpUserDataSourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIdpUserConfig(userpoolName, username, organizationID, testSubdomain, password, true),
				Check: resource.ComposeTestCheckFunc(
					test.AccCheckResourceIDField("data.yandex_organizationmanager_idp_user.source", "user_id"),
					resource.TestCheckResourceAttrSet("data.yandex_organizationmanager_idp_user.source", "id"),
					resource.TestCheckResourceAttr("data.yandex_organizationmanager_idp_user.source", "username", username),
					resource.TestCheckResourceAttrSet("data.yandex_organizationmanager_idp_user.source", "userpool_id"),
					test.AccCheckCreatedAtAttr("data.yandex_organizationmanager_idp_user.source"),
				),
			},
		},
	})
}

func testAccDataSourceIdpUserConfig(userpoolName, username, organizationID, defaultSubdomain, password string, useID bool) string {
	if useID {
		return testAccDataSourceIdpUserResourceConfig(userpoolName, username, organizationID, defaultSubdomain, password) + idpUserDataByIDConfig
	}

	return testAccDataSourceIdpUserResourceConfig(userpoolName, username, organizationID, defaultSubdomain, password)
}

func testAccDataSourceIdpUserResourceConfig(userpoolName, username, organizationID, defaultSubdomain, password string) string {
	userNameBase := strings.Split(username, "@")[0]
	return fmt.Sprintf(`
resource "yandex_organizationmanager_idp_userpool" "foobar" {
  name              = "%s"
  organization_id   = "%s"
  default_subdomain = "%s"
}

resource "yandex_organizationmanager_idp_user" "foobar" {
  userpool_id = yandex_organizationmanager_idp_userpool.foobar.userpool_id
  username    = "%s"
  full_name   = "Test User"
  email       = "%s@example.com"
  is_active   = true
  password_spec = {
    password = "%s"
  }
}
`, userpoolName, organizationID, defaultSubdomain, username, userNameBase, password)
}

const idpUserDataByIDConfig = `
data "yandex_organizationmanager_idp_user" "source" {
  user_id = yandex_organizationmanager_idp_user.foobar.user_id
}
`

func testAccCheckIdpUserDataSourceDestroy(s *terraform.State) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type == "yandex_organizationmanager_idp_user" {
			_, err := idpsdk.NewUserClient(config.SDKv2).Get(context.Background(), &idp.GetUserRequest{
				UserId: rs.Primary.ID,
			})

			if err != nil {
				if grpcStatus, ok := status.FromError(err); ok && grpcStatus != nil && grpcStatus.Code() == codes.NotFound {
					continue
				} else if ok {
					return fmt.Errorf("Error while requesting Yandex Cloud: grpc code error : %d, http message error: %s", grpcStatus.Code(), grpcStatus.Message())
				}
				return fmt.Errorf("IdpUser still exists")
			}
			return fmt.Errorf("IdpUser still exists")
		} else if rs.Type == "yandex_organizationmanager_idp_userpool" {
			_, err := idpsdk.NewUserpoolClient(config.SDKv2).Get(context.Background(), &idp.GetUserpoolRequest{
				UserpoolId: rs.Primary.ID,
			})

			if err != nil {
				if grpcStatus, ok := status.FromError(err); ok && grpcStatus != nil && grpcStatus.Code() == codes.NotFound {
					continue
				} else if ok {
					return fmt.Errorf("Error while requesting Yandex Cloud: grpc code error : %d, http message error: %s", grpcStatus.Code(), grpcStatus.Message())
				}
				return fmt.Errorf("IdpUserpool still exists")
			}
			return fmt.Errorf("IdpUserpool still exists")
		}
	}

	return nil
}
