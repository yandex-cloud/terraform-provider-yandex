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
)

const albRouterDataSourceResource = "data.yandex_alb_http_router.test-router-ds"

func TestAccDataSourceALBHTTPRouter_byID(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	routerName := acctest.RandomWithPrefix("tf-router")
	routerDesc := "Description for test"
	folderID := acceptance.FolderID()

	var httpRouter apploadbalancer.HttpRouter

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testAccCheckALBHTTPRouterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceALBHTTPRouterConfigByID(routerName, routerDesc),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccDataSourceALBHTTPRouterExists(albRouterDataSourceResource, &httpRouter),
					statecheck.CheckResourceIDField(albRouterDataSourceResource, "http_router_id"),
					resource.TestCheckResourceAttr(albRouterDataSourceResource, "name", routerName),
					resource.TestCheckResourceAttr(albRouterDataSourceResource, "description", routerDesc),
					resource.TestCheckResourceAttr(albRouterDataSourceResource, "folder_id", folderID),
					statecheck.CheckCreatedAtAttr(albRouterDataSourceResource),
				),
			},
		},
	})
}

func TestAccDataSourceALBHTTPRouter_byName(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	routerName := acctest.RandomWithPrefix("tf-router")
	routerDesc := "Description for test"
	folderID := acceptance.FolderID()

	var httpRouter apploadbalancer.HttpRouter

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testAccCheckALBHTTPRouterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceALBHTTPRouterConfigByName(routerName, routerDesc),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccDataSourceALBHTTPRouterExists(albRouterDataSourceResource, &httpRouter),
					statecheck.CheckResourceIDField(albRouterDataSourceResource, "http_router_id"),
					resource.TestCheckResourceAttr(albRouterDataSourceResource, "name", routerName),
					resource.TestCheckResourceAttr(albRouterDataSourceResource, "description", routerDesc),
					resource.TestCheckResourceAttr(albRouterDataSourceResource, "folder_id", folderID),
					statecheck.CheckCreatedAtAttr(albRouterDataSourceResource),
				),
			},
		},
	})
}

func (suite testSuite) testAccDataSourceALBHTTPRouterExists(n string, httpRouter *apploadbalancer.HttpRouter) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ds, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if ds.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}

		client := albsdk.NewHttpRouterClient(config.SDK)

		found, err := client.Get(context.Background(), &apploadbalancer.GetHttpRouterRequest{
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
