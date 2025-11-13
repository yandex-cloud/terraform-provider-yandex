package yandex_organizationmanager_idp_user_test

import (
	"context"
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	idp "github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1/idp"
	idpsdk "github.com/yandex-cloud/go-sdk/services/organizationmanager/v1/idp"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	userSweepPageSize      = 1000
	userpoolSweepPageSize  = 1000
	userSweepDeleteTimeout = 15 * time.Minute
	testResourceNamePrefix = "tf-acc-test-user"
)

func init() {
	resource.AddTestSweepers("yandex_organizationmanager_idp_user", &resource.Sweeper{
		Name:         "yandex_organizationmanager_idp_user",
		F:            testSweepIdpUser,
		Dependencies: []string{},
	})
}

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

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

func testSweepIdpUser(_ string) error {
	conf, err := test.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	organizationID := test.GetExampleOrganizationID()
	if organizationID == "" {
		log.Printf("[WARN] organization ID is not set, skipping user sweep")
		return nil
	}

	// First, we need to get all userpools to iterate through their users
	userpoolClient := idpsdk.NewUserpoolClient(conf.SDKv2)
	userpoolReq := &idp.ListUserpoolsRequest{
		OrganizationId: organizationID,
		PageSize:       userpoolSweepPageSize,
	}

	userpoolResp, err := userpoolClient.List(context.Background(), userpoolReq)
	if err != nil {
		return fmt.Errorf("error getting list of userpools: %s", err)
	}

	result := &multierror.Error{}
	userClient := idpsdk.NewUserClient(conf.SDKv2)

	// Collect all userpools (handle pagination)
	userpools := userpoolResp.Userpools
	for userpoolResp.NextPageToken != "" {
		userpoolReq.PageToken = userpoolResp.NextPageToken
		userpoolResp, err = userpoolClient.List(context.Background(), userpoolReq)
		if err != nil {
			result = multierror.Append(result, fmt.Errorf("error getting next page of userpools: %s", err))
			break
		}
		userpools = append(userpools, userpoolResp.Userpools...)
	}

	// For each userpool, get and sweep users
	for _, pool := range userpools {
		userReq := &idp.ListUsersRequest{
			UserpoolId: pool.Id,
			PageSize:   userSweepPageSize,
		}

		userResp, err := userClient.List(context.Background(), userReq)
		if err != nil {
			result = multierror.Append(result, fmt.Errorf("error getting list of users for userpool %q: %s", pool.Id, err))
			continue
		}

		// Sweep users with test prefixes
		for _, user := range userResp.Users {
			username := user.Username
			if strings.HasPrefix(username, testResourceNamePrefix) || strings.Contains(username, "@tf-acc-test-") {
				if !sweepIdpUser(conf, user.Id) {
					result = multierror.Append(result, fmt.Errorf("failed to sweep Idp User %q", user.Id))
				}
			}
		}

		// Handle pagination for users
		for userResp.NextPageToken != "" {
			userReq.PageToken = userResp.NextPageToken
			userResp, err = userClient.List(context.Background(), userReq)
			if err != nil {
				result = multierror.Append(result, fmt.Errorf("error getting next page of users for userpool %q: %s", pool.Id, err))
				break
			}

			for _, user := range userResp.Users {
				username := user.Username
				if strings.HasPrefix(username, testResourceNamePrefix) || strings.Contains(username, "@tf-acc-test-") {
					if !sweepIdpUser(conf, user.Id) {
						result = multierror.Append(result, fmt.Errorf("failed to sweep Idp User %q", user.Id))
					}
				}
			}
		}
	}

	return result.ErrorOrNil()
}

func sweepIdpUser(conf *provider_config.Config, id string) bool {
	return test.SweepWithRetry(sweepIdpUserOnce, conf, "Idp User", id)
}

func sweepIdpUserOnce(conf *provider_config.Config, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), userSweepDeleteTimeout)
	defer cancel()

	client := idpsdk.NewUserClient(conf.SDKv2)
	op, err := client.Delete(ctx, &idp.DeleteUserRequest{
		UserId: id,
	})
	if err != nil {
		return err
	}

	_, err = op.Wait(ctx)
	return err
}
