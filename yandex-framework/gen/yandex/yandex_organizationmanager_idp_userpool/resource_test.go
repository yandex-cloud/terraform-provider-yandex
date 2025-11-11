package yandex_organizationmanager_idp_userpool_test

import (
	"context"
	"fmt"
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
	userpoolSweepPageSize      = 1000
	userpoolSweepDeleteTimeout = 15 * time.Minute
	testResourceNamePrefix     = "tf-acc-test-userpool"
)

func init() {
	resource.AddTestSweepers("yandex_organizationmanager_idp_userpool", &resource.Sweeper{
		Name:         "yandex_organizationmanager_idp_userpool",
		F:            testSweepIdpUserpool,
		Dependencies: []string{},
	})
}

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestAccOrganizationManagerIdpUserpool_basic(t *testing.T) {
	userpoolName := acctest.RandomWithPrefix("tf-acc-test-userpool")
	organizationID := test.GetExampleOrganizationID()
	testSubdomain := acctest.RandomWithPrefix("tf-acc-test-subdomain")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckIdpUserpoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIdpUserpool_basic(userpoolName, organizationID, testSubdomain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIdpUserpoolExists("yandex_organizationmanager_idp_userpool.foobar"),
					test.AccCheckCreatedAtAttr("yandex_organizationmanager_idp_userpool.foobar"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_userpool.foobar", "name", userpoolName),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_userpool.foobar", "organization_id", organizationID),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_userpool.foobar", "labels.test-label", "example-label-value"),
					resource.TestCheckResourceAttrSet("yandex_organizationmanager_idp_userpool.foobar", "userpool_id"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_userpool.foobar", "default_subdomain", testSubdomain),
				),
			},
		},
	})
}

func TestAccOrganizationManagerIdpUserpool_updateNameAndLabels(t *testing.T) {
	userpoolName := acctest.RandomWithPrefix("tf-acc-test-userpool")
	organizationID := test.GetExampleOrganizationID()
	testSubdomain := acctest.RandomWithPrefix("tf-acc-test-subdomain")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckIdpUserpoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIdpUserpool_basic(userpoolName, organizationID, testSubdomain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIdpUserpoolExists("yandex_organizationmanager_idp_userpool.foobar"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_userpool.foobar", "name", userpoolName),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_userpool.foobar", "labels.test-label", "example-label-value"),
				),
			},
			{
				Config: testAccIdpUserpool_update(userpoolName+"-updated", organizationID, testSubdomain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIdpUserpoolExists("yandex_organizationmanager_idp_userpool.foobar"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_userpool.foobar", "name", userpoolName+"-updated"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_userpool.foobar", "labels.new-label", "only-shows-up-when-updated"),
				),
			},
			{
				ResourceName:            "yandex_organizationmanager_idp_userpool.foobar",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"default_subdomain"},
			},
		},
	})
}

func TestAccOrganizationManagerIdpUserpool_updateDescription(t *testing.T) {
	userpoolName := acctest.RandomWithPrefix("tf-acc-test-userpool")
	organizationID := test.GetExampleOrganizationID()
	testSubdomain := acctest.RandomWithPrefix("tf-acc-test-subdomain")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckIdpUserpoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIdpUserpool_basic(userpoolName, organizationID, testSubdomain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIdpUserpoolExists("yandex_organizationmanager_idp_userpool.foobar"),
				),
			},
			{
				Config: testAccIdpUserpool_updateDescription(userpoolName, organizationID, testSubdomain, "new-description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIdpUserpoolExists("yandex_organizationmanager_idp_userpool.foobar"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_userpool.foobar", "description", "new-description"),
					resource.TestCheckResourceAttr("yandex_organizationmanager_idp_userpool.foobar", "labels.test-label", "example-label-value"),
				),
			},
			{
				ResourceName:            "yandex_organizationmanager_idp_userpool.foobar",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"default_subdomain"},
			},
		},
	})
}

func testAccCheckIdpUserpoolDestroy(s *terraform.State) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_organizationmanager_idp_userpool" {
			continue
		}

		_, err := idpsdk.NewUserpoolClient(config.SDKv2).Get(context.Background(), &idp.GetUserpoolRequest{
			UserpoolId: rs.Primary.ID,
		})

		if err != nil {
			if grpcStatus, ok := status.FromError(err); ok && grpcStatus != nil && grpcStatus.Code() == codes.NotFound {
				return nil
			} else if ok {
				return fmt.Errorf("Error while requesting Yandex Cloud: grpc code error : %d, http message error: %s", grpcStatus.Code(), grpcStatus.Message())
			}
			return fmt.Errorf("IdpUserpool still exists")
		}
	}

	return nil
}

func testAccCheckIdpUserpoolExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

		found, err := idpsdk.NewUserpoolClient(config.SDKv2).Get(context.Background(), &idp.GetUserpoolRequest{
			UserpoolId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("IdpUserpool %s not found", n)
		}

		return nil
	}
}

func testAccIdpUserpool_basic(name, organizationID, defaultSubdomain string) string {
	return fmt.Sprintf(`
resource "yandex_organizationmanager_idp_userpool" "foobar" {
  name              = "%s"
  organization_id   = "%s"
  default_subdomain = "%s"

  labels = {
    test-label = "example-label-value"
  }
}
`, name, organizationID, defaultSubdomain)
}

func testAccIdpUserpool_update(name, organizationID, defaultSubdomain string) string {
	return fmt.Sprintf(`
resource "yandex_organizationmanager_idp_userpool" "foobar" {
  name              = "%s"
  organization_id   = "%s"
  default_subdomain = "%s"

  labels = {
    new-label   = "only-shows-up-when-updated"
  }
}
`, name, organizationID, defaultSubdomain)
}

func testAccIdpUserpool_updateDescription(name, organizationID, defaultSubdomain, description string) string {
	return fmt.Sprintf(`
resource "yandex_organizationmanager_idp_userpool" "foobar" {
  name              = "%s"
  organization_id   = "%s"
  default_subdomain = "%s"
  description       = "%s"

  labels = {
    test-label = "example-label-value"
  }
}
`, name, organizationID, defaultSubdomain, description)
}

func testSweepIdpUserpool(_ string) error {
	conf, err := test.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	organizationID := test.GetExampleOrganizationID()
	if organizationID == "" {
		return fmt.Errorf("organization ID is required for sweeping userpools")
	}

	req := &idp.ListUserpoolsRequest{
		OrganizationId: organizationID,
		PageSize:       userpoolSweepPageSize,
	}

	client := idpsdk.NewUserpoolClient(conf.SDKv2)
	resp, err := client.List(context.Background(), req)
	if err != nil {
		return fmt.Errorf("error getting list of userpools: %s", err)
	}

	result := &multierror.Error{}
	for _, userpool := range resp.Userpools {
		if strings.HasPrefix(userpool.Name, testResourceNamePrefix) {
			if !sweepIdpUserpool(conf, userpool.Id) {
				result = multierror.Append(result, fmt.Errorf("failed to sweep Idp Userpool %q", userpool.Id))
			}
		}
	}

	// Handle pagination if needed
	for resp.NextPageToken != "" {
		req.PageToken = resp.NextPageToken
		resp, err = client.List(context.Background(), req)
		if err != nil {
			result = multierror.Append(result, fmt.Errorf("error getting next page of userpools: %s", err))
			break
		}

		for _, userpool := range resp.Userpools {
			if strings.HasPrefix(userpool.Name, testResourceNamePrefix) {
				if !sweepIdpUserpool(conf, userpool.Id) {
					result = multierror.Append(result, fmt.Errorf("failed to sweep Idp Userpool %q", userpool.Id))
				}
			}
		}
	}

	return result.ErrorOrNil()
}

func sweepIdpUserpool(conf *provider_config.Config, id string) bool {
	return test.SweepWithRetry(sweepIdpUserpoolOnce, conf, "Idp Userpool", id)
}

func sweepIdpUserpoolOnce(conf *provider_config.Config, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), userpoolSweepDeleteTimeout)
	defer cancel()

	client := idpsdk.NewUserpoolClient(conf.SDKv2)
	op, err := client.Delete(ctx, &idp.DeleteUserpoolRequest{
		UserpoolId: id,
	})
	if err != nil {
		return err
	}

	_, err = op.Wait(ctx)
	return err
}
