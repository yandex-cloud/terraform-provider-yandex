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
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/internal/provider/framework"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAccOrganizationManagerIdpUser_basic(t *testing.T) {
	userpoolName := acctest.RandomWithPrefix("tf-acc-test-userpool")
	userNameBase := acctest.RandomWithPrefix("tf-acc-test-user")
	organizationID := test.GetExampleOrganizationID()
	testSubdomain := acctest.RandomWithPrefix("tf-acc-test-subdomain")
	username := fmt.Sprintf("%s@%s.idp.yandexcloud.net", userNameBase, testSubdomain)
	password := acctest.RandomWithPrefix("Random195!-")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckIdpUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIdpUser_basic(userpoolName, username, organizationID, testSubdomain, password),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIdpUserExists("yandex_organizationmanager_idp_user.foobar"),
					test.AccCheckCreatedAtAttr("yandex_organizationmanager_idp_user.foobar"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_user.foobar", "username", username),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_user.foobar", "full_name", "Test User"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_user.foobar", "email", fmt.Sprintf("%s@example.com", userNameBase)),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_user.foobar", "is_active", "true"),
					resource.TestCheckResourceAttrSet("yandex_organizationmanager_idp_user.foobar", "user_id"),
					resource.TestCheckResourceAttrSet("yandex_organizationmanager_idp_user.foobar", "userpool_id"),
				),
			},
		},
	})
}

func TestAccOrganizationManagerIdpUser_update(t *testing.T) {
	userpoolName := acctest.RandomWithPrefix("tf-userpool")
	userNameBase := acctest.RandomWithPrefix("tf-user")
	organizationID := test.GetExampleOrganizationID()
	testSubdomain := acctest.RandomWithPrefix("tf-acc-test-subdomain")
	username := fmt.Sprintf("%s@%s.idp.yandexcloud.net", userNameBase, testSubdomain)
	password := acctest.RandomWithPrefix("Random195!-")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckIdpUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIdpUser_basic(userpoolName, username, organizationID, testSubdomain, password),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIdpUserExists("yandex_organizationmanager_idp_user.foobar"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_user.foobar", "full_name", "Test User"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_user.foobar", "email", fmt.Sprintf("%s@example.com", userNameBase)),
				),
			},
			{
				Config: testAccIdpUser_update(userpoolName, username, organizationID, testSubdomain, password),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIdpUserExists("yandex_organizationmanager_idp_user.foobar"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_user.foobar", "full_name", "Updated Test User"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_user.foobar", "email", fmt.Sprintf("updated-%s@example.com", userNameBase)),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_user.foobar", "family_name", "Updated"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_user.foobar", "given_name", "Test"),
				),
			},
			{
				ResourceName:            "yandex_organizationmanager_idp_user.foobar",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password_spec", "is_active"},
			},
		},
	})
}

func testAccCheckIdpUserDestroy(s *terraform.State) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_organizationmanager_idp_user" {
			continue
		}

		_, err := idpsdk.NewUserClient(config.SDKv2).Get(context.Background(), &idp.GetUserRequest{
			UserId: rs.Primary.ID,
		})

		if err != nil {
			if grpcStatus, ok := status.FromError(err); ok && grpcStatus != nil && grpcStatus.Code() == codes.NotFound {
				return nil
			} else if ok {
				return fmt.Errorf("Error while requesting Yandex Cloud: grpc code error : %d, http message error: %s", grpcStatus.Code(), grpcStatus.Message())
			}
			return fmt.Errorf("IdpUser still exists")
		}
	}

	return nil
}

func testAccCheckIdpUserExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

		found, err := idpsdk.NewUserClient(config.SDKv2).Get(context.Background(), &idp.GetUserRequest{
			UserId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("IdpUser %s not found", n)
		}

		return nil
	}
}

func testAccIdpUser_basic(userpoolName, username, organizationID, defaultSubdomain, password string) string {
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

func testAccIdpUser_update(userpoolName, username, organizationID, defaultSubdomain, password string) string {
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
  full_name   = "Updated Test User"
  given_name  = "Test"
  family_name = "Updated"
  email       = "updated-%s@example.com"
  is_active   = true
  password_spec = {
    password = "%s"
  }
}
`, userpoolName, organizationID, defaultSubdomain, username, userNameBase, password)
}
