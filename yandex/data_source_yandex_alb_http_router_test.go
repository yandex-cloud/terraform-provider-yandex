package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/apploadbalancer/v1"
)

const albRouterDataSourceResource = "data.yandex_alb_http_router.test-router-ds"

func TestAccDataSourceALBHTTPRouter_byID(t *testing.T) {
	t.Parallel()

	routerName := acctest.RandomWithPrefix("tf-router")
	routerDesc := "Description for test"
	folderID := getExampleFolderID()

	var httpRouter apploadbalancer.HttpRouter

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBHTTPRouterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceALBHTTPRouterConfigByID(routerName, routerDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceALBHTTPRouterExists(albRouterDataSourceResource, &httpRouter),
					testAccCheckResourceIDField(albRouterDataSourceResource, "http_router_id"),
					resource.TestCheckResourceAttr(albRouterDataSourceResource, "name", routerName),
					resource.TestCheckResourceAttr(albRouterDataSourceResource, "description", routerDesc),
					resource.TestCheckResourceAttr(albRouterDataSourceResource, "folder_id", folderID),
					testAccCheckCreatedAtAttr(albRouterDataSourceResource),
				),
			},
		},
	})
}

func TestAccDataSourceALBHTTPRouter_byName(t *testing.T) {
	t.Parallel()

	routerName := acctest.RandomWithPrefix("tf-router")
	routerDesc := "Description for test"
	folderID := getExampleFolderID()

	var httpRouter apploadbalancer.HttpRouter

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBHTTPRouterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceALBHTTPRouterConfigByName(routerName, routerDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceALBHTTPRouterExists(albRouterDataSourceResource, &httpRouter),
					testAccCheckResourceIDField(albRouterDataSourceResource, "http_router_id"),
					resource.TestCheckResourceAttr(albRouterDataSourceResource, "name", routerName),
					resource.TestCheckResourceAttr(albRouterDataSourceResource, "description", routerDesc),
					resource.TestCheckResourceAttr(albRouterDataSourceResource, "folder_id", folderID),
					testAccCheckCreatedAtAttr(albRouterDataSourceResource),
				),
			},
		},
	})
}

func testAccDataSourceALBHTTPRouterExists(n string, httpRouter *apploadbalancer.HttpRouter) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ds, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if ds.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.ApplicationLoadBalancer().HttpRouter().Get(context.Background(), &apploadbalancer.GetHttpRouterRequest{
			HttpRouterId: ds.Primary.ID,
		})

		if err != nil {
			return err
		}

		if found.Id != ds.Primary.ID {
			return fmt.Errorf("HTTP Router not found")
		}

		*httpRouter = *found

		return nil
	}
}

func testAccDataSourceALBHTTPRouterConfigByID(name, desc string) string {
	return fmt.Sprintf(`
data "yandex_alb_http_router" "test-router-ds" {
  http_router_id = "${yandex_alb_http_router.test-router.id}"
}

resource "yandex_alb_http_router" "test-router" {
  name			= "%s"
  description	= "%s"
}
`, name, desc)
}

func testAccDataSourceALBHTTPRouterConfigByName(name, desc string) string {
	return fmt.Sprintf(`
data "yandex_alb_http_router" "test-router-ds" {
  name = "${yandex_alb_http_router.test-router.name}"
}

resource "yandex_alb_http_router" "test-router" {
  name			= "%s"
  description	= "%s"
}
`, name, desc)
}
