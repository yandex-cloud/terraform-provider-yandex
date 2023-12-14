package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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

func TestAccALBVirtualHost_httpRouteWithRBAC(t *testing.T) {
	t.Parallel()

	VHResource := albVirtualHostInfo()
	VHResource.IsHTTPRoute = true
	VHResource.IsRouteRBAC = true
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
					testExistsFirstElementWithAttr(
						albVHResource, "route", "route_options.0.rbac", &vhPath,
					),
					testExistsElementWithAttrValue(
						albVHResource, "route", "route_options.0.rbac.0.action", albDefaultRBACAction, &vhPath,
					),
					testExistsFirstElementWithAttr(
						albVHResource, "route", "route_options.0.rbac.0.principals.0.and_principals", &vhPath,
					),
					testExistsElementWithAttrValue(
						albVHResource, "route", "route_options.0.rbac.0.principals.0.and_principals.0.any", albDefaultAnyPrincipal, &vhPath,
					),
				),
			},
			albVirtualHostImportStep(),
		},
	})
}

func TestAccALBVirtualHost_httpVirtualHostWithRBAC(t *testing.T) {
	t.Skip("Wait until CLOUD-103826 released")
	t.Parallel()

	VHResource := albVirtualHostInfo()
	VHResource.IsVirtualHostRBAC = true
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
						albVHResource, "route_options", "rbac", &vhPath,
					),
					testExistsElementWithAttrValue(
						albVHResource, "route_options", "rbac.0.action", albDefaultRBACAction, &vhPath,
					),
					testExistsFirstElementWithAttr(
						albVHResource, "route_options", "rbac.0.principals.0.and_principals", &vhPath,
					),
					testExistsElementWithAttrValue(
						albVHResource, "route_options", "rbac.0.principals.0.and_principals.0.remote_ip", albDefaultRemoteIP, &vhPath,
					),
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
						albVHResource, "route", "http_route.0.redirect_action.0.response_code", albDefaultRedirectResponseCode, &vhPath,
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

func TestUnitALBVirtualHostParseStringMatch(t *testing.T) {
	t.Parallel()

	bgResource := resourceYandexALBVirtualHost()
	makeRouteOptions := func(stringMatch interface{}) interface{} {
		return []interface{}{
			map[string]interface{}{
				"rbac": []interface{}{
					map[string]interface{}{
						"principals": []interface{}{
							map[string]interface{}{
								"and_principals": []interface{}{
									map[string]interface{}{
										"header": []interface{}{
											map[string]interface{}{
												"value": stringMatch,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}
	}
	stringMatchPath := "route_options.0.rbac.0.principals.0.and_principals.0.header.0.value.0."

	t.Run("string-match-path-regex", func(t *testing.T) {
		stringMatchValue := []interface{}{
			map[string]interface{}{
				"regex": "my_cool_regex",
			},
		}
		rawValues := map[string]interface{}{
			"http_router_id": "id_0",
			"name":           "name_0",
			"route_options":  makeRouteOptions(stringMatchValue),
		}
		resourceData := schema.TestResourceDataRaw(t, bgResource.Schema, rawValues)
		stringMatch, _ := expandALBStringMatch(resourceData, stringMatchPath)

		assert.Equal(t, stringMatch.GetRegexMatch(), "my_cool_regex")
	})

	t.Run("string-match-path-prefix", func(t *testing.T) {
		stringMatchValue := []interface{}{
			map[string]interface{}{
				"prefix": "my_cool_prefix",
			},
		}
		rawValues := map[string]interface{}{
			"http_router_id": "id_0",
			"name":           "name_0",
			"route_options":  makeRouteOptions(stringMatchValue),
		}
		resourceData := schema.TestResourceDataRaw(t, bgResource.Schema, rawValues)
		stringMatch, _ := expandALBStringMatch(resourceData, stringMatchPath)

		assert.Equal(t, stringMatch.GetPrefixMatch(), "my_cool_prefix")
	})

	t.Run("string-match-path-exact", func(t *testing.T) {
		stringMatchValue := []interface{}{
			map[string]interface{}{
				"exact": "my_cool_exact",
			},
		}
		rawValues := map[string]interface{}{
			"http_router_id": "id_0",
			"name":           "name_0",
			"route_options":  makeRouteOptions(stringMatchValue),
		}
		resourceData := schema.TestResourceDataRaw(t, bgResource.Schema, rawValues)
		stringMatch, _ := expandALBStringMatch(resourceData, stringMatchPath)

		assert.Equal(t, stringMatch.GetExactMatch(), "my_cool_exact")
	})
}

func TestUnitALBVirtualHostCreateFromResource(t *testing.T) {
	t.Parallel()

	vhResource := resourceYandexALBVirtualHost()

	makePayload := func(body string) *apploadbalancer.Payload {
		return &apploadbalancer.Payload{
			Payload: &apploadbalancer.Payload_Text{
				Text: body,
			},
		}
	}

	type M = map[string]interface{}
	type S = []interface{}

	t.Run("vh-basic", func(t *testing.T) {
		authority := "example.com"
		rawValues := M{
			"http_router_id": "my-router-id",
			"name":           "vh-name",
			"authority":      S{authority},
		}
		resourceData := schema.TestResourceDataRaw(t, vhResource.Schema, rawValues)
		req, err := buildALBVirtualHostCreateRequest(resourceData)
		require.NoError(t, err, "failed to build create request")

		assert.Equal(t, req.GetHttpRouterId(), "my-router-id")
		assert.Equal(t, req.GetName(), "vh-name")
		assert.Equal(t, req.GetAuthority(), []string{authority})
		assert.Nil(t, req.GetRouteOptions())
	})

	t.Run("vh-route", func(t *testing.T) {
		rawValues := M{
			"http_router_id": "my-router-id",
			"name":           "vh-name",
			"route": S{
				M{
					"name": "teapot-route-1",
					"http_route": S{
						M{
							"direct_response_action": S{
								M{
									"status": 418,
									"body":   "I'm a teapot",
								},
							},
						},
					},
				},
			},
		}
		resourceData := schema.TestResourceDataRaw(t, vhResource.Schema, rawValues)
		req, err := buildALBVirtualHostCreateRequest(resourceData)
		require.NoError(t, err, "failed to build create request")

		assert.Equal(t, req.GetHttpRouterId(), "my-router-id")
		assert.Equal(t, req.GetName(), "vh-name")
		assert.Len(t, req.GetRoutes(), 1)
		route := req.GetRoutes()[0]
		assert.Equal(t, route.GetName(), "teapot-route-1")
		httpRoute := route.GetHttp()
		assert.NotNil(t, httpRoute)
		assert.Nil(t, httpRoute.GetRedirect())
		assert.Nil(t, httpRoute.GetRoute())
		assert.Equal(t, httpRoute.GetDirectResponse().GetStatus(), int64(418))
		assert.Equal(t, httpRoute.GetDirectResponse().GetBody(), makePayload("I'm a teapot"))
	})

	t.Run("vh-route-options", func(t *testing.T) {
		rawValues := M{
			"http_router_id": "my-router-id",
			"name":           "vh-name",
			"route_options": S{
				M{
					"security_profile_id": "sec-profile-id",
				},
			},
		}
		resourceData := schema.TestResourceDataRaw(t, vhResource.Schema, rawValues)
		req, err := buildALBVirtualHostCreateRequest(resourceData)
		require.NoError(t, err, "failed to build create request")

		assert.Equal(t, req.GetHttpRouterId(), "my-router-id")
		assert.Equal(t, req.GetName(), "vh-name")
		assert.NotNil(t, req.GetRouteOptions())
		assert.Equal(t, req.GetRouteOptions().GetSecurityProfileId(), "sec-profile-id")
	})
}

func TestUnitALBVirtualHostUpdateFromResource(t *testing.T) {
	t.Parallel()

	vhResource := resourceYandexALBVirtualHost()

	type M = map[string]interface{}
	type S = []interface{}

	t.Run("vh-basic", func(t *testing.T) {
		authority := "example.com"
		rawValues := M{
			"http_router_id": "my-router-id",
			"name":           "vh-name",
			"authority":      S{authority},
		}
		resourceData := schema.TestResourceDataRaw(t, vhResource.Schema, rawValues)
		req, err := buildALBVirtualHostUpdateRequest(resourceData)
		require.NoError(t, err, "failed to build update request")

		assert.Equal(t, req.GetHttpRouterId(), "my-router-id")
		assert.Equal(t, req.GetVirtualHostName(), "vh-name")
		assert.Equal(t, req.GetAuthority(), []string{authority})
		assert.Nil(t, req.GetRouteOptions())
	})

	t.Run("vh-route-options", func(t *testing.T) {
		rawValues := M{
			"http_router_id": "my-router-id",
			"name":           "vh-name",
			"route_options": S{
				M{
					"security_profile_id": "sec-profile-id",
				},
			},
		}
		resourceData := schema.TestResourceDataRaw(t, vhResource.Schema, rawValues)
		req, err := buildALBVirtualHostUpdateRequest(resourceData)
		require.NoError(t, err, "failed to build update request")

		assert.Equal(t, req.GetHttpRouterId(), "my-router-id")
		assert.Equal(t, req.GetVirtualHostName(), "vh-name")
		assert.NotNil(t, req.GetRouteOptions())
		assert.Equal(t, req.GetRouteOptions().GetSecurityProfileId(), "sec-profile-id")
	})
}
