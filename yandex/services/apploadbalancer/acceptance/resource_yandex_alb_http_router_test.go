package acceptance

import (
	"context"
	"fmt"
	"testing"

	statecheck "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance/statecheck"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/apploadbalancer/v1"
	albsdk "github.com/yandex-cloud/go-sdk/services/apploadbalancer/v1"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const albRouterResource = "yandex_alb_http_router.test-router"

func albHTTPRouterImportStep() resource.TestStep {
	return resource.TestStep{
		ResourceName:      albRouterResource,
		ImportState:       true,
		ImportStateVerify: true,
	}
}

func TestAccALBHTTPRouter_basic(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	var router apploadbalancer.HttpRouter
	routerName := acctest.RandomWithPrefix("tf-http-router")
	routerDesc := acctest.RandomWithPrefix("tf-desc")
	folderID := acceptance.FolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testAccCheckALBHTTPRouterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccALBGeneralHTTPRouterTemplate(routerName, routerDesc),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckALBHTTPRouterExists(albRouterResource, &router),
					resource.TestCheckResourceAttr(albRouterResource, "name", routerName),
					resource.TestCheckResourceAttrSet(albRouterResource, "folder_id"),
					resource.TestCheckResourceAttr(albRouterResource, "folder_id", folderID),
					testAccCheckALBHTTPRouterContainsLabel(&router, "tf-label", "tf-label-value"),
					testAccCheckALBHTTPRouterContainsLabel(&router, "empty-label", ""),
					statecheck.CheckCreatedAtAttr(albRouterResource),
				),
			},
			albHTTPRouterImportStep(),
		},
	})
}

func TestAccALBHTTPRouter_full(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	routerResource := albHTTPRouterInfo()
	routerResource.IsRBAC = true

	var router apploadbalancer.HttpRouter
	routerPath := ""

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testAccCheckALBHTTPRouterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testALBHTTPRouterConfig_basic(routerResource),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckALBHTTPRouterExists(albRouterResource, &router),
					statecheck.ExistsFirstElementWithAttr(
						albRouterResource, "route_options", "rbac", &routerPath,
					),
					statecheck.TestExistsElementWithAttrValue(
						albRouterResource, "route_options", "rbac.0.action", albDefaultRBACAction, &routerPath,
					),
					statecheck.ExistsFirstElementWithAttr(
						albRouterResource, "route_options", "rbac.0.principals.0.and_principals.0.header", &routerPath,
					),
					statecheck.TestExistsElementWithAttrValue(
						albRouterResource, "route_options", "rbac.0.principals.0.and_principals.0.header.0.name", albDefaultHeaderName, &routerPath,
					),
					statecheck.ExistsFirstElementWithAttr(
						albRouterResource, "route_options", "rbac.0.principals.0.and_principals.0.header.0.value", &routerPath,
					),
					statecheck.TestExistsElementWithAttrValue(
						albRouterResource, "route_options", "rbac.0.principals.0.and_principals.0.header.0.value.0.exact", albDefaultHeaderValue, &routerPath,
					),
				),
			},
			albHTTPRouterImportStep(),
		},
	})
}

func TestAccALBHTTPRouter_update(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	var router apploadbalancer.HttpRouter

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testAccCheckALBHTTPRouterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccALBGeneralHTTPRouterTemplate(
					"tf-http-router", "tf-descr",
				),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckALBHTTPRouterExists(albRouterResource, &router),
				),
			},
			{
				Config: testAccALBGeneralHTTPRouterTemplate(
					"tf-http-router-updated", "tf-descr-updated",
				),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckALBHTTPRouterExists(albRouterResource, &router),
				),
			},
			{
				Config: testAccALBGeneralHTTPRouterTemplate(
					"tf-http-router-updated", "tf-descr-updated",
				),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckALBHTTPRouterExists(albRouterResource, &router),
				),
			},
			albHTTPRouterImportStep(),
		},
	})
}

func (suite testSuite) testAccCheckALBHTTPRouterDestroy(s *terraform.State) error {
	config, configErr := suite.Config()
	if configErr != nil {
		return configErr
	}

	client := albsdk.NewHttpRouterClient(config.SDK)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_alb_http_router" {
			continue
		}

		_, err := client.Get(context.Background(), &apploadbalancer.GetHttpRouterRequest{
			HttpRouterId: rs.Primary.ID,
		})
		if status.Code(err) != codes.NotFound {
			return fmt.Errorf("Http Router still exists")
		}
	}

	return nil
}

func (suite testSuite) testAccCheckALBHTTPRouterExists(routerName string, router *apploadbalancer.HttpRouter) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[routerName]
		if !ok {
			return fmt.Errorf("Not found: %s", routerName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}

		client := albsdk.NewHttpRouterClient(config.SDK)

		found, err := client.Get(context.Background(), &apploadbalancer.GetHttpRouterRequest{
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
