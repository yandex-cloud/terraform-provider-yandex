package yandex

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/apploadbalancer/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const albRouterResource = "yandex_alb_http_router.test-router"

func init() {
	resource.AddTestSweepers("yandex_alb_http_router", &resource.Sweeper{
		Name: "yandex_alb_http_router",
		F:    testSweepALBHTTPRouters,
		Dependencies: []string{
			"yandex_alb_load_balancer",
		},
	})
}

func testSweepALBHTTPRouters(_ string) error {
	log.Printf("[DEBUG] Sweeping Http Router")
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	result := &multierror.Error{}

	req := &apploadbalancer.ListHttpRoutersRequest{FolderId: conf.FolderID}
	it := conf.sdk.ApplicationLoadBalancer().HttpRouter().HttpRouterIterator(conf.Context(), req)
	for it.Next() {
		id := it.Value().GetId()

		if !sweepALBHTTPRouter(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep ALB Http Router %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepALBHTTPRouter(conf *Config, id string) bool {
	return sweepWithRetry(sweepALBHTTPRouterOnce, conf, "ALB Http Router", id)
}

func sweepALBHTTPRouterOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexIAMServiceAccountDefaultTimeout)
	defer cancel()

	op, err := conf.sdk.ApplicationLoadBalancer().HttpRouter().Delete(ctx, &apploadbalancer.DeleteHttpRouterRequest{
		HttpRouterId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}

func albHTTPRouterImportStep() resource.TestStep {
	return resource.TestStep{
		ResourceName:      albRouterResource,
		ImportState:       true,
		ImportStateVerify: true,
	}
}

func TestAccALBHTTPRouter_basic(t *testing.T) {
	t.Parallel()

	var router apploadbalancer.HttpRouter
	routerName := acctest.RandomWithPrefix("tf-http-router")
	routerDesc := acctest.RandomWithPrefix("tf-desc")
	folderID := getExampleFolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBHTTPRouterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccALBGeneralHTTPRouterTemplate(routerName, routerDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckALBHTTPRouterExists(albRouterResource, &router),
					resource.TestCheckResourceAttr(albRouterResource, "name", routerName),
					resource.TestCheckResourceAttrSet(albRouterResource, "folder_id"),
					resource.TestCheckResourceAttr(albRouterResource, "folder_id", folderID),
					testAccCheckALBHTTPRouterContainsLabel(&router, "tf-label", "tf-label-value"),
					testAccCheckALBHTTPRouterContainsLabel(&router, "empty-label", ""),
					testAccCheckCreatedAtAttr(albRouterResource),
				),
			},
			albHTTPRouterImportStep(),
		},
	})
}

func TestAccALBHTTPRouter_full(t *testing.T) {
	t.Parallel()

	routerResource := albHTTPRouterInfo()
	routerResource.IsRBAC = true

	var router apploadbalancer.HttpRouter
	routerPath := ""

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBHTTPRouterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testALBHTTPRouterConfig_basic(routerResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckALBHTTPRouterExists(albRouterResource, &router),
					testExistsFirstElementWithAttr(
						albRouterResource, "route_options", "rbac", &routerPath,
					),
					testExistsElementWithAttrValue(
						albRouterResource, "route_options", "rbac.0.action", albDefaultRBACAction, &routerPath,
					),
					testExistsFirstElementWithAttr(
						albRouterResource, "route_options", "rbac.0.principals.0.and_principals.0.header", &routerPath,
					),
					testExistsElementWithAttrValue(
						albRouterResource, "route_options", "rbac.0.principals.0.and_principals.0.header.0.name", albDefaultHeaderName, &routerPath,
					),
					testExistsFirstElementWithAttr(
						albRouterResource, "route_options", "rbac.0.principals.0.and_principals.0.header.0.value", &routerPath,
					),
					testExistsElementWithAttrValue(
						albRouterResource, "route_options", "rbac.0.principals.0.and_principals.0.header.0.value.0.exact", albDefaultHeaderValue, &routerPath,
					),
				),
			},
			albHTTPRouterImportStep(),
		},
	})
}

func TestAccALBHTTPRouter_update(t *testing.T) {
	var router apploadbalancer.HttpRouter

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBHTTPRouterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccALBGeneralHTTPRouterTemplate(
					"tf-http-router", "tf-descr",
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckALBHTTPRouterExists(albRouterResource, &router),
				),
			},
			{
				Config: testAccALBGeneralHTTPRouterTemplate(
					"tf-http-router-updated", "tf-descr-updated",
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckALBHTTPRouterExists(albRouterResource, &router),
				),
			},
			{
				Config: testAccALBGeneralHTTPRouterTemplate(
					"tf-http-router-updated", "tf-descr-updated",
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckALBHTTPRouterExists(albRouterResource, &router),
				),
			},
			albHTTPRouterImportStep(),
		},
	})
}

func testAccCheckALBHTTPRouterDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_alb_http_router" {
			continue
		}

		_, err := config.sdk.ApplicationLoadBalancer().HttpRouter().Get(context.Background(), &apploadbalancer.GetHttpRouterRequest{
			HttpRouterId: rs.Primary.ID,
		})
		if status.Code(err) != codes.NotFound {
			return fmt.Errorf("Http Router still exists")
		}
	}

	return nil
}

func testAccCheckALBHTTPRouterExists(routerName string, router *apploadbalancer.HttpRouter) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[routerName]
		if !ok {
			return fmt.Errorf("Not found: %s", routerName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.ApplicationLoadBalancer().HttpRouter().Get(context.Background(), &apploadbalancer.GetHttpRouterRequest{
			HttpRouterId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Http Router not found")
		}

		*router = *found

		return nil
	}
}

func testAccCheckALBHTTPRouterContainsLabel(router *apploadbalancer.HttpRouter, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		v, ok := router.Labels[key]
		if !ok {
			return fmt.Errorf("Expected label with key '%s' not found", key)
		}
		if v != value {
			return fmt.Errorf("Incorrect label value for key '%s': expected '%s' but found '%s'", key, value, v)
		}
		return nil
	}
}

func testAccALBGeneralHTTPRouterTemplate(name, desc string) string {
	return fmt.Sprintf(`
resource "yandex_alb_http_router" "test-router" {
  name		  = "%s"
  description = "%s"

  labels = {
    tf-label    = "tf-label-value"
    empty-label = ""
  }
}
`, name, desc)
}
