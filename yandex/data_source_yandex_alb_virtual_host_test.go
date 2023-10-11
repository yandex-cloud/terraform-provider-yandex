package yandex

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/apploadbalancer/v1"
)

const albVirtualHostDataSourceResource = "data.yandex_alb_virtual_host.test-virtual-host-ds"

func TestAccDataSourceALBVirtualHost_byID(t *testing.T) {
	t.Parallel()

	vhName := acctest.RandomWithPrefix("tf-virtual-host")
	routerName := acctest.RandomWithPrefix("tf-http-router")
	routerDesc := acctest.RandomWithPrefix("tf-http-router-desc")

	var virtualHost apploadbalancer.VirtualHost

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBVirtualHostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceALBVirtualHostConfigByID(routerName, routerDesc, vhName),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceALBVirtualHostExists(albVirtualHostDataSourceResource, &virtualHost),
					testAccCheckResourceIDField(albVirtualHostDataSourceResource, "virtual_host_id"),
					resource.TestCheckResourceAttr(albVirtualHostDataSourceResource, "name", vhName),
				),
			},
		},
	})
}

func TestAccDataSourceALBVirtualHost_byName(t *testing.T) {
	t.Parallel()

	vhName := acctest.RandomWithPrefix("tf-virtual-host")
	routerName := acctest.RandomWithPrefix("tf-http-router")
	routerDesc := acctest.RandomWithPrefix("tf-http-router-desc")

	var virtualHost apploadbalancer.VirtualHost

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBVirtualHostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceALBVirtualHostConfigByName(routerName, routerDesc, vhName),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceALBVirtualHostExists(albVirtualHostDataSourceResource, &virtualHost),
					testAccCheckResourceIDField(albVirtualHostDataSourceResource, "virtual_host_id"),
					resource.TestCheckResourceAttr(albVirtualHostDataSourceResource, "name", vhName),
				),
			},
		},
	})
}

func TestAccDataSourceALBVirtualHost_httpRouteWithHTTPRouteAction(t *testing.T) {
	t.Parallel()

	VHResource := albVirtualHostInfo()
	VHResource.IsDataSource = true
	VHResource.IsHTTPRoute = true
	VHResource.IsHTTPRouteAction = true
	var virtualHost apploadbalancer.VirtualHost
	routePath := ""

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBVirtualHostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testALBVirtualHostConfig_basic(VHResource),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceALBVirtualHostExists(albVirtualHostDataSourceResource, &virtualHost),
					testAccCheckALBVirtualHostValues(&virtualHost, true, false),
					testExistsFirstElementWithAttr(
						albVirtualHostDataSourceResource, "route", "name", &routePath,
					),
					testCheckResourceSubAttrFn(
						albVirtualHostDataSourceResource, &routePath, "name", func(value string) error {
							routeName := virtualHost.GetRoutes()[0].Name
							if value != routeName {
								return fmt.Errorf("Virtual Host's route's name doesnt't match. %s != %s", value, routeName)
							}
							return nil
						},
					),
					testExistsFirstElementWithAttr(
						albVirtualHostDataSourceResource, "route", "http_route.0.http_route_action.0.backend_group_id", &routePath,
					),
					testCheckResourceSubAttrFn(
						albVirtualHostDataSourceResource, &routePath, "http_route.0.http_route_action.0.backend_group_id", func(value string) error {
							bgID := virtualHost.GetRoutes()[0].GetHttp().GetRoute().GetBackendGroupId()
							if value != bgID {
								return fmt.Errorf("Virtual Host's http route's http route action's backend group id doesnt't match. %s != %s", value, bgID)
							}
							return nil
						},
					),
					testExistsElementWithAttrValue(
						albVirtualHostDataSourceResource, "route", "http_route.0.http_route_action.0.timeout", albDefaultTimeout, &routePath,
					),
					testCheckResourceSubAttrFn(
						albVirtualHostDataSourceResource, &routePath, "http_route.0.http_route_action.0.timeout", func(value string) error {
							timeout := formatDuration(virtualHost.GetRoutes()[0].GetHttp().GetRoute().GetTimeout())
							if value != timeout {
								return fmt.Errorf("Virtual Host's http route's http route action's timeout doesnt't match. %s != %s", value, timeout)
							}
							return nil
						},
					),
				),
			},
		},
	})
}

func TestAccDataSourceTestAccDataSourceALBVirtualHost_httpRouteWithRedirectAction(t *testing.T) {
	t.Parallel()

	VHResource := albVirtualHostInfo()
	VHResource.IsDataSource = true
	VHResource.IsHTTPRoute = true
	VHResource.IsRedirectAction = true
	var virtualHost apploadbalancer.VirtualHost
	routePath := ""

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBVirtualHostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testALBVirtualHostConfig_basic(VHResource),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceALBVirtualHostExists(albVirtualHostDataSourceResource, &virtualHost),
					testAccCheckALBVirtualHostValues(&virtualHost, true, false),
					testExistsFirstElementWithAttr(
						albVirtualHostDataSourceResource, "route", "name", &routePath,
					),
					testCheckResourceSubAttrFn(
						albVirtualHostDataSourceResource, &routePath, "name", func(value string) error {
							routeName := virtualHost.GetRoutes()[0].Name
							if value != routeName {
								return fmt.Errorf("Virtual Host's route's name doesnt't match. %s != %s", value, routeName)
							}
							return nil
						},
					),
					testExistsElementWithAttrValue(
						albVirtualHostDataSourceResource, "route", "http_route.0.redirect_action.0.response_code", albDefaultRedirectResponseCode, &routePath,
					),
				),
			},
		},
	})
}

func TestAccDataSourceALBVirtualHost_httpRouteWithDirectResponseAction(t *testing.T) {
	t.Parallel()

	VHResource := albVirtualHostInfo()
	VHResource.IsDataSource = true
	VHResource.IsHTTPRoute = true
	VHResource.IsDirectResponseAction = true
	var virtualHost apploadbalancer.VirtualHost
	routePath := ""

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBVirtualHostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testALBVirtualHostConfig_basic(VHResource),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceALBVirtualHostExists(albVirtualHostDataSourceResource, &virtualHost),
					testAccCheckALBVirtualHostValues(&virtualHost, true, false),
					testExistsFirstElementWithAttr(
						albVirtualHostDataSourceResource, "route", "name", &routePath,
					),
					testCheckResourceSubAttrFn(
						albVirtualHostDataSourceResource, &routePath, "name", func(value string) error {
							routeName := virtualHost.GetRoutes()[0].Name
							if value != routeName {
								return fmt.Errorf("Virtual Host's route's name doesnt't match. %s != %s", value, routeName)
							}
							return nil
						},
					),
					testExistsElementWithAttrValue(
						albVirtualHostDataSourceResource, "route", "http_route.0.direct_response_action.0.status", albDefaultDirectResponseStatus, &routePath,
					),
					testCheckResourceSubAttrFn(
						albVirtualHostDataSourceResource, &routePath, "http_route.0.direct_response_action.0.status", func(value string) error {
							status := virtualHost.GetRoutes()[0].GetHttp().GetDirectResponse().GetStatus()
							if realValue, _ := strconv.ParseInt(value, 10, 64); realValue != status {
								return fmt.Errorf("Virtual Host's http route's direct response action's status doesnt't match. %d != %d", realValue, status)
							}
							return nil
						},
					),
					testExistsElementWithAttrValue(
						albVirtualHostDataSourceResource, "route", "http_route.0.direct_response_action.0.body", albDefaultDirectResponseBody, &routePath,
					),
					testCheckResourceSubAttrFn(
						albVirtualHostDataSourceResource, &routePath, "http_route.0.direct_response_action.0.body", func(value string) error {
							body := virtualHost.GetRoutes()[0].GetHttp().GetDirectResponse().GetBody().GetText()
							if value != body {
								return fmt.Errorf("Virtual Host's http route's direct response action's status doesnt't match. %s != %s", value, body)
							}
							return nil
						},
					),
				),
			},
		},
	})
}

func TestAccDataSourceALBVirtualHost_grpcRouteWithGRPCRouteAction(t *testing.T) {
	t.Parallel()

	VHResource := albVirtualHostInfo()
	VHResource.IsDataSource = true
	VHResource.IsGRPCRoute = true
	VHResource.IsGRPCRouteAction = true
	var virtualHost apploadbalancer.VirtualHost
	routePath := ""

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBVirtualHostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testALBVirtualHostConfig_basic(VHResource),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceALBVirtualHostExists(albVirtualHostDataSourceResource, &virtualHost),
					testAccCheckALBVirtualHostValues(&virtualHost, false, true),
					testExistsFirstElementWithAttr(
						albVirtualHostDataSourceResource, "route", "name", &routePath,
					),
					testCheckResourceSubAttrFn(
						albVirtualHostDataSourceResource, &routePath, "name", func(value string) error {
							routeName := virtualHost.GetRoutes()[0].Name
							if value != routeName {
								return fmt.Errorf("Virtual Host's route's name doesnt't match. %s != %s", value, routeName)
							}
							return nil
						},
					),
					testExistsElementWithAttrValue(
						albVirtualHostDataSourceResource, "route", "grpc_route.0.grpc_route_action.0.max_timeout", albDefaultTimeout, &routePath,
					),
					testCheckResourceSubAttrFn(
						albVirtualHostDataSourceResource, &routePath, "grpc_route.0.grpc_route_action.0.max_timeout", func(value string) error {
							timeout := formatDuration(virtualHost.GetRoutes()[0].GetGrpc().GetRoute().GetMaxTimeout())
							if value != timeout {
								return fmt.Errorf("Virtual Host's grpc route's route action's max timeout doesnt't match. %s != %s", value, timeout)
							}
							return nil
						},
					),
					testExistsElementWithAttrValue(
						albVirtualHostDataSourceResource, "route", "grpc_route.0.grpc_route_action.0.auto_host_rewrite", albDefaultAutoHostRewrite, &routePath,
					),
					testCheckResourceSubAttrFn(
						albVirtualHostDataSourceResource, &routePath, "grpc_route.0.grpc_route_action.0.auto_host_rewrite", func(value string) error {
							autoHostRewrite := virtualHost.GetRoutes()[0].GetGrpc().GetRoute().GetAutoHostRewrite()
							if realValue, _ := strconv.ParseBool(value); realValue != autoHostRewrite {
								return fmt.Errorf("Virtual Host's grpc route's route action's auto host rewrite doesnt't match. %s != %t", value, autoHostRewrite)
							}
							return nil
						},
					),
				),
			},
		},
	})
}

func TestAccDataSourceALBVirtualHost_grpcRouteWithGRPCStatusResponseAction(t *testing.T) {
	t.Parallel()

	VHResource := albVirtualHostInfo()
	VHResource.IsDataSource = true
	VHResource.IsGRPCRoute = true
	VHResource.IsGRPCStatusResponseAction = true
	var virtualHost apploadbalancer.VirtualHost
	routePath := ""

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBVirtualHostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testALBVirtualHostConfig_basic(VHResource),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceALBVirtualHostExists(albVirtualHostDataSourceResource, &virtualHost),
					testAccCheckALBVirtualHostValues(&virtualHost, false, true),
					testExistsFirstElementWithAttr(
						albVirtualHostDataSourceResource, "route", "name", &routePath,
					),
					testCheckResourceSubAttrFn(
						albVirtualHostDataSourceResource, &routePath, "name", func(value string) error {
							routeName := virtualHost.GetRoutes()[0].Name
							if value != routeName {
								return fmt.Errorf("Virtual Host's route's name doesnt't match. %s != %s", value, routeName)
							}
							return nil
						},
					),
					testExistsElementWithAttrValue(
						albVirtualHostDataSourceResource, "route", "grpc_route.0.grpc_status_response_action.0.status", albDefaultStatusResponse, &routePath,
					),
					testCheckResourceSubAttrFn(
						albVirtualHostDataSourceResource, &routePath, "grpc_route.0.grpc_status_response_action.0.status", func(value string) error {
							status := strings.ToLower(virtualHost.GetRoutes()[0].GetGrpc().GetStatusResponse().GetStatus().String())
							if value != status {
								return fmt.Errorf("Virtual Host's grpc route's status response action's status doesnt't match. %s != %s", value, status)
							}
							return nil
						},
					),
				),
			},
		},
	})
}

func testAccDataSourceALBVirtualHostExists(n string, virtualHost *apploadbalancer.VirtualHost) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ds, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		var httpRouterID string
		var virtualHostName string
		if ds.Primary.ID == "" {
			if ds.Primary.Attributes["http_router_id"] == "" || ds.Primary.Attributes["name"] == "" {
				return fmt.Errorf("No ID and no http_router_id with name are set")
			}
			httpRouterID = ds.Primary.Attributes["http_router_id"]
			virtualHostName = ds.Primary.Attributes["name"]
		} else {
			httpRouterID, virtualHostName = retrieveDataFromVirtualHostID(ds.Primary.ID)
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.ApplicationLoadBalancer().VirtualHost().Get(context.Background(), &apploadbalancer.GetVirtualHostRequest{
			HttpRouterId:    httpRouterID,
			VirtualHostName: virtualHostName,
		})
		if err != nil {
			return err
		}

		if found.Name != ds.Primary.Attributes["name"] {
			return fmt.Errorf("Virtual Host not found")
		}

		*virtualHost = *found

		return nil
	}
}

func testAccDataSourceALBVirtualHostConfigByID(routerName, routerDesc, vhName string) string {
	return testAccALBGeneralHTTPRouterTemplate(routerName, routerDesc) + fmt.Sprintf(`
data "yandex_alb_virtual_host" "test-virtual-host-ds" {
  virtual_host_id = "${yandex_alb_virtual_host.test-virtual-host.id}"
}

resource "yandex_alb_virtual_host" "test-virtual-host" {
  http_router_id = yandex_alb_http_router.test-router.id
  name			= "%s"
}
`, vhName)
}

func testAccDataSourceALBVirtualHostConfigByName(routerName, routerDesc, vhName string) string {
	return testAccALBGeneralHTTPRouterTemplate(routerName, routerDesc) + fmt.Sprintf(`
data "yandex_alb_virtual_host" "test-virtual-host-ds" {
  name = yandex_alb_virtual_host.test-virtual-host.name
   http_router_id = yandex_alb_virtual_host.test-virtual-host.http_router_id
}

resource "yandex_alb_virtual_host" "test-virtual-host" {
  name			= "%s"
  http_router_id = yandex_alb_http_router.test-router.id
}
`, vhName)
}
