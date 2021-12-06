package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/apploadbalancer/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const albVHResource = "yandex_alb_virtual_host.test-vh"

func albVirtualHostImportStep() resource.TestStep {
	return resource.TestStep{
		ResourceName:      albVHResource,
		ImportState:       true,
		ImportStateVerify: true,
	}
}

func TestAccALBVirtualHost_basic(t *testing.T) {
	t.Parallel()

	var virtualHost apploadbalancer.VirtualHost
	virtualHostName := acctest.RandomWithPrefix("tf-virtual-host")
	httpRouterName := acctest.RandomWithPrefix("tf-http-router")
	httpRouterDesc := acctest.RandomWithPrefix("tf-http-router-desc")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBVirtualHostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccALBVirtualHostBasic(httpRouterName, httpRouterDesc, virtualHostName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckALBVirtualHostExists(albVHResource, &virtualHost),
					resource.TestCheckResourceAttr(albVHResource, "name", virtualHostName),
				),
			},
			albVirtualHostImportStep(),
		},
	})
}

func TestAccALBVirtualHost_httpRouteWithHTTPRouteAction(t *testing.T) {
	t.Parallel()

	VHResource := albVirtualHostInfo()
	VHResource.IsHTTPRoute = true
	VHResource.IsHTTPRouteAction = true
	VHResource.IsHTTPRouteActionHostRewrite = true
	VHResource.HTTPRouteActionHostRewrite = "some.host.rewrite"
	var virtualHost apploadbalancer.VirtualHost
	vhPath := ""

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBVirtualHostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testALBVirtualHostConfig_basic(VHResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckALBVirtualHostExists(albVHResource, &virtualHost),
					testAccCheckALBVirtualHostValues(&virtualHost, true, false),
					testExistsFirstElementWithAttr(
						albVHResource, "modify_request_headers", "name", &vhPath,
					),
					testExistsElementWithAttrValue(
						albVHResource, "modify_request_headers", "append", albDefaultModificationAppend, &vhPath,
					),
					testExistsFirstElementWithAttr(
						albVHResource, "route", "name", &vhPath,
					),
					testExistsElementWithAttrValue(
						albVHResource, "route", "http_route.0.http_route_action.0.timeout", albDefaultTimeout, &vhPath,
					),
					testExistsElementWithAttrValue(
						albVHResource, "route", "http_route.0.http_route_action.0.host_rewrite", "some.host.rewrite", &vhPath,
					),
				),
			},
			albVirtualHostImportStep(),
		},
	})
}

func TestAccALBVirtualHost_httpRouteWithRedirectAction(t *testing.T) {
	t.Parallel()

	VHResource := albVirtualHostInfo()
	VHResource.IsHTTPRoute = true
	VHResource.IsRedirectAction = true
	var virtualHost apploadbalancer.VirtualHost
	vhPath := ""

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBVirtualHostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testALBVirtualHostConfig_basic(VHResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckALBVirtualHostExists(albVHResource, &virtualHost),
					testAccCheckALBVirtualHostValues(&virtualHost, true, false),
					testExistsFirstElementWithAttr(
						albVHResource, "modify_request_headers", "name", &vhPath,
					),
					testExistsElementWithAttrValue(
						albVHResource, "modify_request_headers", "append", albDefaultModificationAppend, &vhPath,
					),
					testExistsFirstElementWithAttr(
						albVHResource, "route", "name", &vhPath,
					),
					testExistsElementWithAttrValue(
						albVHResource, "route", "http_route.0.redirect_action.0.replace_prefix", albDefaultRedirectReplacePrefix, &vhPath,
					),
				),
			},
			albVirtualHostImportStep(),
		},
	})
}

func TestAccALBVirtualHost_httpRouteWithDirectResponseAction(t *testing.T) {
	t.Parallel()

	VHResource := albVirtualHostInfo()
	VHResource.IsHTTPRoute = true
	VHResource.IsDirectResponseAction = true
	var virtualHost apploadbalancer.VirtualHost
	vhPath := ""

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBVirtualHostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testALBVirtualHostConfig_basic(VHResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckALBVirtualHostExists(albVHResource, &virtualHost),
					testAccCheckALBVirtualHostValues(&virtualHost, true, false),
					testExistsFirstElementWithAttr(
						albVHResource, "modify_request_headers", "name", &vhPath,
					),
					testExistsElementWithAttrValue(
						albVHResource, "modify_request_headers", "append", albDefaultModificationAppend, &vhPath,
					),
					testExistsFirstElementWithAttr(
						albVHResource, "route", "name", &vhPath,
					),
					testExistsElementWithAttrValue(
						albVHResource, "route", "http_route.0.direct_response_action.0.status", albDefaultDirectResponseStatus, &vhPath,
					),
					testExistsElementWithAttrValue(
						albVHResource, "route", "http_route.0.direct_response_action.0.body", albDefaultDirectResponseBody, &vhPath,
					),
				),
			},
			albVirtualHostImportStep(),
		},
	})
}

func TestAccALBVirtualHost_grpcRouteWithGRPCRouteAction(t *testing.T) {
	t.Parallel()

	VHResource := albVirtualHostInfo()
	VHResource.IsGRPCRoute = true
	VHResource.IsGRPCRouteAction = true
	var virtualHost apploadbalancer.VirtualHost
	vhPath := ""

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBVirtualHostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testALBVirtualHostConfig_basic(VHResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckALBVirtualHostExists(albVHResource, &virtualHost),
					testAccCheckALBVirtualHostValues(&virtualHost, false, true),
					testExistsFirstElementWithAttr(
						albVHResource, "modify_request_headers", "name", &vhPath,
					),
					testExistsElementWithAttrValue(
						albVHResource, "modify_request_headers", "append", albDefaultModificationAppend, &vhPath,
					),
					testExistsFirstElementWithAttr(
						albVHResource, "route", "name", &vhPath,
					),
					testExistsElementWithAttrValue(
						albVHResource, "route", "grpc_route.0.grpc_route_action.0.max_timeout", albDefaultTimeout, &vhPath,
					),
					testExistsElementWithAttrValue(
						albVHResource, "route", "grpc_route.0.grpc_route_action.0.auto_host_rewrite", albDefaultAutoHostRewrite, &vhPath,
					),
				),
			},
			albVirtualHostImportStep(),
		},
	})
}

func TestAccALBVirtualHost_grpcRouteWithGRPCStatusResponseAction(t *testing.T) {
	t.Parallel()

	VHResource := albVirtualHostInfo()
	VHResource.IsGRPCRoute = true
	VHResource.IsGRPCStatusResponseAction = true
	var virtualHost apploadbalancer.VirtualHost
	vhPath := ""

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBVirtualHostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testALBVirtualHostConfig_basic(VHResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckALBVirtualHostExists(albVHResource, &virtualHost),
					testAccCheckALBVirtualHostValues(&virtualHost, false, true),
					testExistsFirstElementWithAttr(
						albVHResource, "modify_request_headers", "name", &vhPath,
					),
					testExistsElementWithAttrValue(
						albVHResource, "modify_request_headers", "append", albDefaultModificationAppend, &vhPath,
					),
					testExistsFirstElementWithAttr(
						albVHResource, "route", "name", &vhPath,
					),
					testExistsElementWithAttrValue(
						albVHResource, "route", "grpc_route.0.grpc_status_response_action.0.status", albDefaultStatusResponse, &vhPath,
					),
				),
			},
			albVirtualHostImportStep(),
		},
	})
}

func testAccCheckALBVirtualHostDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_alb_virtual_host" {
			continue
		}
		httpRouterID := rs.Primary.Attributes["http_router_id"]
		virtualHostName := rs.Primary.Attributes["name"]
		if httpRouterID == "" || virtualHostName == "" {
			httpRouterID, virtualHostName = retrieveDataFromVirtualHostID(rs.Primary.ID)
		}

		_, err := config.sdk.ApplicationLoadBalancer().VirtualHost().Get(context.Background(), &apploadbalancer.GetVirtualHostRequest{
			HttpRouterId:    httpRouterID,
			VirtualHostName: virtualHostName,
		})
		if status.Code(err) != codes.NotFound {
			return fmt.Errorf("Virtual Host still exists")
		}
	}

	return nil
}

func testAccCheckALBVirtualHostExists(virtualHostName string, virtualHost *apploadbalancer.VirtualHost) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[virtualHostName]
		if !ok {
			return fmt.Errorf("Not found: %s", virtualHostName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.ApplicationLoadBalancer().VirtualHost().Get(context.Background(), &apploadbalancer.GetVirtualHostRequest{
			HttpRouterId:    rs.Primary.Attributes["http_router_id"],
			VirtualHostName: rs.Primary.Attributes["name"],
		})
		if err != nil {
			return err
		}

		if found.Name != rs.Primary.Attributes["name"] {
			return fmt.Errorf("Virtual Host not found")
		}

		*virtualHost = *found

		return nil
	}
}

func testAccALBVirtualHostBasic(httpRouterName, httpRouterDesc, virtualHostName string) string {
	return testAccALBGeneralHTTPRouterTemplate(httpRouterName, httpRouterDesc) + fmt.Sprintf(`
resource "yandex_alb_virtual_host" "test-vh" {
  http_router_id = yandex_alb_http_router.test-router.id
  name		= "%s"
}
`, virtualHostName)
}
