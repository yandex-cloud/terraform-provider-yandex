package yandex

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/apploadbalancer/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
)

const albBGResource = "yandex_alb_backend_group.test-bg"

func init() {
	resource.AddTestSweepers("yandex_alb_backend_group", &resource.Sweeper{
		Name: "yandex_alb_backend_group",
		F:    testSweepALBBackendGroups,
		Dependencies: []string{
			"yandex_alb_http_router",
		},
	})
}

func makeCookie(name string) interface{} {
	return map[string]interface{}{
		"cookie": []interface{}{
			map[string]interface{}{
				"name": name,
				"ttl":  formatDuration(durationpb.New(1 * time.Minute)),
			},
		},
	}
}

func makeHeader(name string) interface{} {
	return map[string]interface{}{
		"header": []interface{}{
			map[string]interface{}{
				"header_name": name,
			},
		},
	}
}

func makeConn(ip bool) interface{} {
	return map[string]interface{}{
		"connection": []interface{}{
			map[string]interface{}{
				"source_ip": ip,
			},
		},
	}
}

func TestUnitALBBackendGroupFlatternSessionAffinity(t *testing.T) {
	t.Parallel()

	affinityMap, err := flattenALBHTTPSessionAffinity(&apploadbalancer.HttpBackendGroup{})
	require.NoError(t, err)
	assert.Empty(t, affinityMap)

	const (
		headerName = "x-some-header"
		cookieName = "some-cookie"
	)
	t.Run("http-header-affinity", func(t *testing.T) {
		bg := &apploadbalancer.HttpBackendGroup{
			SessionAffinity: &apploadbalancer.HttpBackendGroup_Header{
				Header: &apploadbalancer.HeaderSessionAffinity{
					HeaderName: headerName,
				},
			},
		}
		affinityMap, err = flattenALBHTTPSessionAffinity(bg)
		require.NoError(t, err)
		assert.EqualValues(t, []interface{}{makeHeader(headerName)}, affinityMap)
	})

	t.Run("http-cookie-affinity", func(t *testing.T) {
		bg := &apploadbalancer.HttpBackendGroup{
			SessionAffinity: &apploadbalancer.HttpBackendGroup_Cookie{
				Cookie: &apploadbalancer.CookieSessionAffinity{
					Name: cookieName,
					Ttl:  durationpb.New(1 * time.Minute),
				},
			},
		}
		affinityMap, err = flattenALBHTTPSessionAffinity(bg)
		require.NoError(t, err)
		assert.EqualValues(t, []interface{}{makeCookie(cookieName)}, affinityMap)
	})

	t.Run("grpc-header-affinity", func(t *testing.T) {
		bg := &apploadbalancer.GrpcBackendGroup{
			SessionAffinity: &apploadbalancer.GrpcBackendGroup_Header{
				Header: &apploadbalancer.HeaderSessionAffinity{
					HeaderName: headerName,
				},
			},
		}
		affinityMap, err = flattenALBGRPCSessionAffinity(bg)
		require.NoError(t, err)
		assert.EqualValues(t, []interface{}{makeHeader(headerName)}, affinityMap)
	})

	t.Run("grpc-cookie-affinity", func(t *testing.T) {
		bg := &apploadbalancer.GrpcBackendGroup{
			SessionAffinity: &apploadbalancer.GrpcBackendGroup_Cookie{
				Cookie: &apploadbalancer.CookieSessionAffinity{
					Name: cookieName,
					Ttl:  durationpb.New(1 * time.Minute),
				},
			},
		}
		affinityMap, err = flattenALBGRPCSessionAffinity(bg)
		require.NoError(t, err)
		assert.EqualValues(t, []interface{}{makeCookie(cookieName)}, affinityMap)
	})

	t.Run("stream-connection-affinity", func(t *testing.T) {
		bg := &apploadbalancer.StreamBackendGroup{
			SessionAffinity: &apploadbalancer.StreamBackendGroup_Connection{
				Connection: &apploadbalancer.ConnectionSessionAffinity{
					SourceIp: true,
				},
			},
		}
		affinityMap, err = flattenALBStreamSessionAffinity(bg)
		require.NoError(t, err)
		assert.EqualValues(t, []interface{}{makeConn(true)}, affinityMap)
	})
}

func TestUnitALBBackendGroupCreateFromResource(t *testing.T) {
	t.Parallel()

	bgResource := resourceYandexALBBackendGroup()

	makeBackend := func() interface{} {
		return []interface{}{
			map[string]interface{}{
				"name":             "backend1",
				"port":             8080,
				"target_group_ids": []interface{}{"tg1"},
			},
		}
	}

	t.Run("http-backend-group-cookie", func(t *testing.T) {
		rawValues := map[string]interface{}{
			"id":   "bgid",
			"name": "bg-name",
			"session_affinity": []interface{}{
				makeCookie("cook-name"),
			},
			"http_backend": makeBackend(),
		}
		resourceData := schema.TestResourceDataRaw(t, bgResource.Schema, rawValues)

		resourceData.SetId("bgid")

		req, err := buildALBBackendGroupCreateRequest(resourceData, "test-folder")
		require.NoError(t, err, "failed to build create request")

		assert.Equal(t, req.GetFolderId(), "test-folder")
		assert.Equal(t, req.GetName(), "bg-name")
		assert.NotNil(t, req.GetHttp())
		assert.NotNil(t, req.GetHttp().GetCookie())
		assert.Equal(t, 1*time.Minute, req.GetHttp().GetCookie().GetTtl().AsDuration())
	})

	t.Run("http-backend-group-header-affinity", func(t *testing.T) {
		rawValues := map[string]interface{}{
			"id":   "bgid",
			"name": "bg-name",
			"session_affinity": []interface{}{
				makeHeader("hdr-name"),
			},
			"http_backend": makeBackend(),
		}
		resourceData := schema.TestResourceDataRaw(t, bgResource.Schema, rawValues)

		resourceData.SetId("bgid")

		req, err := buildALBBackendGroupCreateRequest(resourceData, "test-folder")
		require.NoError(t, err, "failed to build create request")

		assert.Equal(t, req.GetFolderId(), "test-folder")
		assert.Equal(t, req.GetName(), "bg-name")
		assert.NotNil(t, req.GetHttp())
		assert.NotNil(t, req.GetHttp().GetHeader())
		assert.Equal(t, "hdr-name", req.GetHttp().GetHeader().GetHeaderName())
	})

	t.Run("http-backend-group-connection-affinity", func(t *testing.T) {
		rawValues := map[string]interface{}{
			"id":   "bgid",
			"name": "bg-name",
			"session_affinity": []interface{}{
				makeConn(true),
			},
			"http_backend": makeBackend(),
		}
		resourceData := schema.TestResourceDataRaw(t, bgResource.Schema, rawValues)

		resourceData.SetId("bgid")

		req, err := buildALBBackendGroupCreateRequest(resourceData, "test-folder")
		require.NoError(t, err, "failed to build create request")

		assert.Equal(t, req.GetFolderId(), "test-folder")
		assert.Equal(t, req.GetName(), "bg-name")
		assert.NotNil(t, req.GetHttp())
		assert.NotNil(t, req.GetHttp().GetConnection())
		assert.True(t, req.GetHttp().GetConnection().GetSourceIp())
	})

	t.Run("grpc-backend-group-cookie", func(t *testing.T) {
		rawValues := map[string]interface{}{
			"id":   "bgid",
			"name": "bg-name",
			"session_affinity": []interface{}{
				makeCookie("cook-name"),
			},
			"grpc_backend": makeBackend(),
		}
		resourceData := schema.TestResourceDataRaw(t, bgResource.Schema, rawValues)

		resourceData.SetId("bgid")

		req, err := buildALBBackendGroupCreateRequest(resourceData, "test-folder")
		require.NoError(t, err, "failed to build create request")

		assert.Equal(t, req.GetFolderId(), "test-folder")
		assert.Equal(t, req.GetName(), "bg-name")
		assert.NotNil(t, req.GetGrpc())
		assert.NotNil(t, req.GetGrpc().GetCookie())
	})

	t.Run("grpc-backend-group-header-affinity", func(t *testing.T) {
		rawValues := map[string]interface{}{
			"id":   "bgid",
			"name": "bg-name",
			"session_affinity": []interface{}{
				makeHeader("hdr-name"),
			},
			"grpc_backend": makeBackend(),
		}
		resourceData := schema.TestResourceDataRaw(t, bgResource.Schema, rawValues)

		resourceData.SetId("bgid")

		req, err := buildALBBackendGroupCreateRequest(resourceData, "test-folder")
		require.NoError(t, err, "failed to build create request")

		assert.Equal(t, req.GetFolderId(), "test-folder")
		assert.Equal(t, req.GetName(), "bg-name")
		assert.NotNil(t, req.GetGrpc())
		assert.NotNil(t, req.GetGrpc().GetHeader())
		assert.Equal(t, "hdr-name", req.GetGrpc().GetHeader().GetHeaderName())
	})

	t.Run("grpc-backend-group-connection-affinity", func(t *testing.T) {
		rawValues := map[string]interface{}{
			"id":   "bgid",
			"name": "bg-name",
			"session_affinity": []interface{}{
				makeConn(true),
			},
			"grpc_backend": makeBackend(),
		}
		resourceData := schema.TestResourceDataRaw(t, bgResource.Schema, rawValues)

		resourceData.SetId("bgid")

		req, err := buildALBBackendGroupCreateRequest(resourceData, "test-folder")
		require.NoError(t, err, "failed to build create request")

		assert.Equal(t, req.GetFolderId(), "test-folder")
		assert.Equal(t, req.GetName(), "bg-name")
		assert.NotNil(t, req.GetGrpc())
		assert.NotNil(t, req.GetGrpc().GetConnection())
		assert.True(t, req.GetGrpc().GetConnection().GetSourceIp())
	})

	t.Run("stream-backend-group-connection-affinity", func(t *testing.T) {
		rawValues := map[string]interface{}{
			"id":   "bgid",
			"name": "bg-name",
			"session_affinity": []interface{}{
				makeConn(true),
			},
			"stream_backend": makeBackend(),
		}
		resourceData := schema.TestResourceDataRaw(t, bgResource.Schema, rawValues)

		resourceData.SetId("bgid")

		req, err := buildALBBackendGroupCreateRequest(resourceData, "test-folder")
		require.NoError(t, err, "failed to build create request")

		assert.Equal(t, req.GetFolderId(), "test-folder")
		assert.Equal(t, req.GetName(), "bg-name")
		assert.NotNil(t, req.GetStream())
		assert.NotNil(t, req.GetStream().GetConnection())
		assert.True(t, req.GetStream().GetConnection().GetSourceIp())
	})

	t.Run("stream-backend-group-header-affinity-err", func(t *testing.T) {
		rawValues := map[string]interface{}{
			"id":   "bgid",
			"name": "bg-name",
			"session_affinity": []interface{}{
				makeHeader("hdr-name"),
			},
			"stream_backend": makeBackend(),
		}
		resourceData := schema.TestResourceDataRaw(t, bgResource.Schema, rawValues)

		resourceData.SetId("bgid")

		_, err := buildALBBackendGroupCreateRequest(resourceData, "test-folder")
		require.Error(t, err)
	})
}

func TestUnitALBBackendGroupUpdateFromResource(t *testing.T) {
	t.Parallel()

	bgResource := resourceYandexALBBackendGroup()

	makeCookie := func(name string) interface{} {
		return map[string]interface{}{
			"cookie": []interface{}{
				map[string]interface{}{
					"name": name,
				},
			},
		}
	}

	makeHeader := func(name string) interface{} {
		return map[string]interface{}{
			"header": []interface{}{
				map[string]interface{}{
					"header_name": name,
				},
			},
		}
	}

	makeConn := func(ip bool) interface{} {
		return map[string]interface{}{
			"connection": []interface{}{
				map[string]interface{}{
					"source_ip": ip,
				},
			},
		}
	}

	makeBackend := func() interface{} {
		return []interface{}{
			map[string]interface{}{
				"name":             "backend1",
				"port":             8080,
				"target_group_ids": []interface{}{"tg1"},
			},
		}
	}

	t.Run("http-backend-group-cookie", func(t *testing.T) {
		rawValues := map[string]interface{}{
			"id":   "bgid",
			"name": "bg-name",
			"session_affinity": []interface{}{
				makeCookie("cook-name"),
			},
			"http_backend": makeBackend(),
		}
		resourceData := schema.TestResourceDataRaw(t, bgResource.Schema, rawValues)

		resourceData.SetId("bgid")

		req, err := buildALBBackendGroupUpdateRequest(resourceData)
		require.NoError(t, err, "failed to build create request")

		assert.Equal(t, req.GetBackendGroupId(), "bgid")
		assert.Equal(t, req.GetName(), "bg-name")
		assert.NotNil(t, req.GetHttp())
		assert.NotNil(t, req.GetHttp().GetCookie())
	})

	t.Run("http-backend-group-header-affinity", func(t *testing.T) {
		rawValues := map[string]interface{}{
			"id":   "bgid",
			"name": "bg-name",
			"session_affinity": []interface{}{
				makeHeader("hdr-name"),
			},
			"http_backend": makeBackend(),
		}
		resourceData := schema.TestResourceDataRaw(t, bgResource.Schema, rawValues)

		resourceData.SetId("bgid")

		req, err := buildALBBackendGroupUpdateRequest(resourceData)
		require.NoError(t, err, "failed to build create request")

		assert.Equal(t, req.GetBackendGroupId(), "bgid")
		assert.Equal(t, req.GetName(), "bg-name")
		assert.NotNil(t, req.GetHttp())
		assert.NotNil(t, req.GetHttp().GetHeader())
		assert.Equal(t, "hdr-name", req.GetHttp().GetHeader().GetHeaderName())
	})

	t.Run("http-backend-group-connection-affinity", func(t *testing.T) {
		rawValues := map[string]interface{}{
			"id":   "bgid",
			"name": "bg-name",
			"session_affinity": []interface{}{
				makeConn(true),
			},
			"http_backend": makeBackend(),
		}
		resourceData := schema.TestResourceDataRaw(t, bgResource.Schema, rawValues)

		resourceData.SetId("bgid")

		req, err := buildALBBackendGroupUpdateRequest(resourceData)
		require.NoError(t, err, "failed to build create request")

		assert.Equal(t, req.GetBackendGroupId(), "bgid")
		assert.Equal(t, req.GetName(), "bg-name")
		assert.NotNil(t, req.GetHttp())
		assert.NotNil(t, req.GetHttp().GetConnection())
		assert.True(t, req.GetHttp().GetConnection().GetSourceIp())
	})

	t.Run("grpc-backend-group-cookie", func(t *testing.T) {
		rawValues := map[string]interface{}{
			"id":   "bgid",
			"name": "bg-name",
			"session_affinity": []interface{}{
				makeCookie("cook-name"),
			},
			"grpc_backend": makeBackend(),
		}
		resourceData := schema.TestResourceDataRaw(t, bgResource.Schema, rawValues)

		resourceData.SetId("bgid")

		req, err := buildALBBackendGroupUpdateRequest(resourceData)
		require.NoError(t, err, "failed to build create request")

		assert.Equal(t, req.GetBackendGroupId(), "bgid")
		assert.Equal(t, req.GetName(), "bg-name")
		assert.NotNil(t, req.GetGrpc())
		assert.NotNil(t, req.GetGrpc().GetCookie())
	})

	t.Run("grpc-backend-group-header-affinity", func(t *testing.T) {
		rawValues := map[string]interface{}{
			"id":   "bgid",
			"name": "bg-name",
			"session_affinity": []interface{}{
				makeHeader("hdr-name"),
			},
			"grpc_backend": makeBackend(),
		}
		resourceData := schema.TestResourceDataRaw(t, bgResource.Schema, rawValues)

		resourceData.SetId("bgid")

		req, err := buildALBBackendGroupUpdateRequest(resourceData)
		require.NoError(t, err, "failed to build create request")

		assert.Equal(t, req.GetBackendGroupId(), "bgid")
		assert.Equal(t, req.GetName(), "bg-name")
		assert.NotNil(t, req.GetGrpc())
		assert.NotNil(t, req.GetGrpc().GetHeader())
		assert.Equal(t, "hdr-name", req.GetGrpc().GetHeader().GetHeaderName())
	})

	t.Run("grpc-backend-group-connection-affinity", func(t *testing.T) {
		rawValues := map[string]interface{}{
			"id":   "bgid",
			"name": "bg-name",
			"session_affinity": []interface{}{
				makeConn(true),
			},
			"grpc_backend": makeBackend(),
		}
		resourceData := schema.TestResourceDataRaw(t, bgResource.Schema, rawValues)

		resourceData.SetId("bgid")

		req, err := buildALBBackendGroupUpdateRequest(resourceData)
		require.NoError(t, err, "failed to build create request")

		assert.Equal(t, req.GetBackendGroupId(), "bgid")
		assert.Equal(t, req.GetName(), "bg-name")
		assert.NotNil(t, req.GetGrpc())
		assert.NotNil(t, req.GetGrpc().GetConnection())
		assert.True(t, req.GetGrpc().GetConnection().GetSourceIp())
	})

	t.Run("stream-backend-group-connection-affinity", func(t *testing.T) {
		rawValues := map[string]interface{}{
			"id":   "bgid",
			"name": "bg-name",
			"session_affinity": []interface{}{
				makeConn(true),
			},
			"stream_backend": makeBackend(),
		}
		resourceData := schema.TestResourceDataRaw(t, bgResource.Schema, rawValues)

		resourceData.SetId("bgid")

		req, err := buildALBBackendGroupUpdateRequest(resourceData)
		require.NoError(t, err, "failed to build create request")

		assert.Equal(t, req.GetBackendGroupId(), "bgid")
		assert.Equal(t, req.GetName(), "bg-name")
		assert.NotNil(t, req.GetStream())
		assert.NotNil(t, req.GetStream().GetConnection())
		assert.True(t, req.GetStream().GetConnection().GetSourceIp())
	})

	t.Run("stream-backend-group-header-affinity-err", func(t *testing.T) {
		rawValues := map[string]interface{}{
			"id":   "bgid",
			"name": "bg-name",
			"session_affinity": []interface{}{
				makeHeader("hdr-name"),
			},
			"stream_backend": makeBackend(),
		}
		resourceData := schema.TestResourceDataRaw(t, bgResource.Schema, rawValues)

		resourceData.SetId("bgid")

		_, err := buildALBBackendGroupUpdateRequest(resourceData)
		require.Error(t, err)
	})
}

func testSweepALBBackendGroups(_ string) error {
	log.Printf("[DEBUG] Sweeping Backend Group")
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	result := &multierror.Error{}

	req := &apploadbalancer.ListBackendGroupsRequest{FolderId: conf.FolderID}
	it := conf.sdk.ApplicationLoadBalancer().BackendGroup().BackendGroupIterator(conf.Context(), req)
	for it.Next() {
		id := it.Value().GetId()

		if !sweepALBBackendGroup(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep ALB Backend Group %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepALBBackendGroup(conf *Config, id string) bool {
	return sweepWithRetry(sweepALBBackendGroupOnce, conf, "ALB Backend Group", id)
}

func sweepALBBackendGroupOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexIAMServiceAccountDefaultTimeout)
	defer cancel()

	op, err := conf.sdk.ApplicationLoadBalancer().BackendGroup().Delete(ctx, &apploadbalancer.DeleteBackendGroupRequest{
		BackendGroupId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}

func albBackendGroupImportStep() resource.TestStep {
	return resource.TestStep{
		ResourceName:      albBGResource,
		ImportState:       true,
		ImportStateVerify: true,
	}
}

func TestAccALBBackendGroup_basic(t *testing.T) {
	t.Parallel()

	var bg apploadbalancer.BackendGroup
	bgName := acctest.RandomWithPrefix("tf-backend-group")
	folderID := getExampleFolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBBackendGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccALBBackendGroupBasic(bgName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckALBBackendGroupExists(albBGResource, &bg),
					resource.TestCheckResourceAttr(albBGResource, "name", bgName),
					resource.TestCheckResourceAttrSet(albBGResource, "folder_id"),
					resource.TestCheckResourceAttr(albBGResource, "folder_id", folderID),
					testAccCheckALBBackendGroupContainsLabel(&bg, "tf-label", "tf-label-value"),
					testAccCheckALBBackendGroupContainsLabel(&bg, "empty-label", ""),
					testAccCheckCreatedAtAttr(albBGResource),
					testAccCheckALBBackendGroupValues(&bg, false, false, false),
				),
			},
			albBackendGroupImportStep(),
		},
	})
}

func TestAccALBBackendGroup_fullWithEmptyTLS(t *testing.T) {
	t.Parallel()

	BGResource := albBackendGroupInfo()
	BGResource.IsHTTPBackend = true
	BGResource.IsEmptyTLS = true

	var bg apploadbalancer.BackendGroup
	backendPath := ""

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBBackendGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testALBBackendGroupConfig_basic(BGResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckALBBackendGroupExists(albBGResource, &bg),
					testAccCheckALBBackendGroupValues(&bg, true, false, false),
					testExistsFirstElementWithAttr(
						albBGResource, "http_backend", "name", &backendPath,
					),
					testExistsElementWithAttrValue(
						albBGResource, "http_backend", "weight", albDefaultBackendWeight, &backendPath,
					),
					testExistsElementWithAttrValue(
						albBGResource, "http_backend", "port", albDefaultPort, &backendPath,
					),
					testExistsElementWithAttrValue(
						albBGResource, "http_backend", "load_balancing_config.0.strict_locality", albDefaultStrictLocality, &backendPath,
					),
					testExistsElementWithAttrValue(
						albBGResource, "http_backend", "load_balancing_config.0.locality_aware_routing_percent", albDefaultLocalityPercent, &backendPath,
					),
					testExistsElementWithAttrValue(
						albBGResource, "http_backend", "load_balancing_config.0.panic_threshold", albDefaultPanicThreshold, &backendPath,
					),
				),
			},
			albBackendGroupImportStep(),
		},
	})
}

func TestAccALBBackendGroup_sessionAffinityHeader(t *testing.T) {
	t.Parallel()

	BGResource := albBackendGroupInfo()
	BGResource.IsHTTPBackend = true
	BGResource.UseHeaderAffinity = true

	var bg apploadbalancer.BackendGroup
	backendPath := ""

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBBackendGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testALBBackendGroupConfig_basic(BGResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckALBBackendGroupExists(albBGResource, &bg),
					testAccCheckALBBackendGroupValues(&bg, true, false, false),
					testAccCheckALBBackendGroupHTTPBackend(&bg, albDefaultValidationContext),
					testExistsFirstElementWithAttr(
						albBGResource, "http_backend", "name", &backendPath,
					),
					testExistsElementWithAttrValue(
						albBGResource, "session_affinity", "header.0.header_name", albDefaultHeaderAffinity, &backendPath,
					),
				),
			},
			albBackendGroupImportStep(),
		},
	})
}

func TestAccALBBackendGroup_fullWithHTTPBackend(t *testing.T) {
	t.Parallel()

	BGResource := albBackendGroupInfo()
	BGResource.IsHTTPBackend = true

	var bg apploadbalancer.BackendGroup
	backendPath := ""

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBBackendGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testALBBackendGroupConfig_basic(BGResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckALBBackendGroupExists(albBGResource, &bg),
					testAccCheckALBBackendGroupValues(&bg, true, false, false),
					testAccCheckALBBackendGroupHTTPBackend(&bg, albDefaultValidationContext),
					testExistsFirstElementWithAttr(
						albBGResource, "http_backend", "name", &backendPath,
					),
					testExistsElementWithAttrValue(
						albBGResource, "http_backend", "tls.0.sni", albDefaultSni, &backendPath,
					),
					testExistsElementWithAttrTrimmedValue(
						albBGResource, "http_backend", "tls.0.validation_context.0.trusted_ca_bytes", albDefaultValidationContext, &backendPath,
					),
					testExistsElementWithAttrValue(
						albBGResource, "http_backend", "weight", albDefaultBackendWeight, &backendPath,
					),
					testExistsElementWithAttrValue(
						albBGResource, "http_backend", "port", albDefaultPort, &backendPath,
					),
					testExistsElementWithAttrValue(
						albBGResource, "http_backend", "load_balancing_config.0.strict_locality", albDefaultStrictLocality, &backendPath,
					),
					testExistsElementWithAttrValue(
						albBGResource, "http_backend", "load_balancing_config.0.locality_aware_routing_percent", albDefaultLocalityPercent, &backendPath,
					),
					testExistsElementWithAttrValue(
						albBGResource, "http_backend", "load_balancing_config.0.panic_threshold", albDefaultPanicThreshold, &backendPath,
					),
				),
			},
			albBackendGroupImportStep(),
		},
	})
}

func TestAccALBBackendGroup_fullWithGRPCBackend(t *testing.T) {
	t.Parallel()

	BGResource := albBackendGroupInfo()
	BGResource.IsGRPCBackend = true

	var bg apploadbalancer.BackendGroup
	backendPath := ""

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBBackendGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testALBBackendGroupConfig_basic(BGResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckALBBackendGroupExists(albBGResource, &bg),
					testAccCheckALBBackendGroupValues(&bg, false, true, false),
					testAccCheckALBBackendGroupGRPCBackend(&bg, albDefaultValidationContext),
					testExistsFirstElementWithAttr(
						albBGResource, "grpc_backend", "name", &backendPath,
					),
					testExistsElementWithAttrValue(
						albBGResource, "grpc_backend", "tls.0.sni", albDefaultSni, &backendPath,
					),
					testExistsElementWithAttrTrimmedValue(
						albBGResource, "grpc_backend", "tls.0.validation_context.0.trusted_ca_bytes", albDefaultValidationContext, &backendPath,
					),
					testExistsElementWithAttrValue(
						albBGResource, "grpc_backend", "weight", albDefaultBackendWeight, &backendPath,
					),
					testExistsElementWithAttrValue(
						albBGResource, "grpc_backend", "port", albDefaultPort, &backendPath,
					),
					testExistsElementWithAttrValue(
						albBGResource, "grpc_backend", "load_balancing_config.0.strict_locality", albDefaultStrictLocality, &backendPath,
					),
					testExistsElementWithAttrValue(
						albBGResource, "grpc_backend", "load_balancing_config.0.locality_aware_routing_percent", albDefaultLocalityPercent, &backendPath,
					),
					testExistsElementWithAttrValue(
						albBGResource, "grpc_backend", "load_balancing_config.0.panic_threshold", albDefaultPanicThreshold, &backendPath,
					),
				),
			},
			albBackendGroupImportStep(),
		},
	})
}

func TestAccALBBackendGroup_httpBackendWithHttpHealthCheck(t *testing.T) {
	t.Parallel()

	BGResource := albBackendGroupInfo()
	BGResource.IsHTTPBackend = true
	BGResource.IsHTTPCheck = true

	var bg apploadbalancer.BackendGroup
	backendPath := ""

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBBackendGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testALBBackendGroupConfig_basic(BGResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckALBBackendGroupExists(albBGResource, &bg),
					testAccCheckALBBackendGroupValues(&bg, true, false, false),
					testExistsFirstElementWithAttr(
						albBGResource, "http_backend", "name", &backendPath,
					),
					testExistsElementWithAttrValue(
						albBGResource, "http_backend", "healthcheck.*.timeout", albDefaultTimeout, &backendPath,
					),
					testExistsElementWithAttrValue(
						albBGResource, "http_backend", "healthcheck.*.interval", albDefaultInterval, &backendPath,
					),
					testExistsElementWithAttrValue(
						albBGResource, "http_backend", "healthcheck.*.http_healthcheck.0.host", albDefaultHost, &backendPath,
					),
					testExistsElementWithAttrValue(
						albBGResource, "http_backend", "healthcheck.*.http_healthcheck.0.path", albDefaultPath, &backendPath,
					),
					testExistsElementWithAttrValue(
						albBGResource, "http_backend", "healthcheck.*.http_healthcheck.0.http2", albDefaultHTTP2, &backendPath,
					),
				),
			},
			albBackendGroupImportStep(),
		},
	})
}

func TestAccALBBackendGroup_httpBackendWithGRPCHealthCheck(t *testing.T) {
	t.Parallel()

	BGResource := albBackendGroupInfo()
	BGResource.IsHTTPBackend = true
	BGResource.IsGRPCCheck = true

	var bg apploadbalancer.BackendGroup
	backendPath := ""

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBBackendGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testALBBackendGroupConfig_basic(BGResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckALBBackendGroupExists(albBGResource, &bg),
					testAccCheckALBBackendGroupValues(&bg, true, false, false),
					testExistsFirstElementWithAttr(
						albBGResource, "http_backend", "name", &backendPath,
					),
					testExistsElementWithAttrValue(
						albBGResource, "http_backend", "healthcheck.*.timeout", albDefaultTimeout, &backendPath,
					),
					testExistsElementWithAttrValue(
						albBGResource, "http_backend", "healthcheck.*.interval", albDefaultInterval, &backendPath,
					),
					testExistsElementWithAttrValue(
						albBGResource, "http_backend", "healthcheck.*.grpc_healthcheck.0.service_name", albDefaultServiceName, &backendPath,
					),
				),
			},
			albBackendGroupImportStep(),
		},
	})
}

func TestAccALBBackendGroup_streamBackend(t *testing.T) {
	t.Parallel()

	BGResource := albBackendGroupInfo()
	BGResource.IsStreamBackend = true
	BGResource.IsHTTPCheck = true

	var bg apploadbalancer.BackendGroup
	backendPath := ""

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBBackendGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testALBBackendGroupConfig_basic(BGResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckALBBackendGroupExists(albBGResource, &bg),
					testAccCheckALBBackendGroupValues(&bg, false, false, true),
					testExistsFirstElementWithAttr(
						albBGResource, "stream_backend", "name", &backendPath,
					),
					testExistsElementWithAttrValue(
						albBGResource, "stream_backend", "healthcheck.*.timeout", albDefaultTimeout, &backendPath,
					),
					testExistsElementWithAttrValue(
						albBGResource, "stream_backend", "healthcheck.*.interval", albDefaultInterval, &backendPath,
					),
					testExistsElementWithAttrValue(
						albBGResource, "stream_backend", "healthcheck.*.http_healthcheck.0.host", albDefaultHost, &backendPath,
					),
					testExistsElementWithAttrValue(
						albBGResource, "stream_backend", "enable_proxy_protocol", albDefaultProxyProtocol, &backendPath,
					),
				),
			},
			albBackendGroupImportStep(),
		},
	})
}

func TestAccALBBackendGroup_streamBackendWithProxyProtocol(t *testing.T) {
	t.Parallel()

	proxyProtocol := "true"
	BGResource := albBackendGroupInfo()
	BGResource.IsStreamBackend = true
	BGResource.IsHTTPCheck = true
	BGResource.ProxyProtocol = proxyProtocol

	var bg apploadbalancer.BackendGroup
	backendPath := ""

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBBackendGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testALBBackendGroupConfig_basic(BGResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckALBBackendGroupExists(albBGResource, &bg),
					testAccCheckALBBackendGroupValues(&bg, false, false, true),
					testExistsFirstElementWithAttr(
						albBGResource, "stream_backend", "name", &backendPath,
					),
					testExistsElementWithAttrValue(
						albBGResource, "stream_backend", "healthcheck.*.timeout", albDefaultTimeout, &backendPath,
					),
					testExistsElementWithAttrValue(
						albBGResource, "stream_backend", "healthcheck.*.interval", albDefaultInterval, &backendPath,
					),
					testExistsElementWithAttrValue(
						albBGResource, "stream_backend", "healthcheck.*.http_healthcheck.0.host", albDefaultHost, &backendPath,
					),
					testExistsElementWithAttrValue(
						albBGResource, "stream_backend", "enable_proxy_protocol", proxyProtocol, &backendPath,
					),
				),
			},
			albBackendGroupImportStep(),
		},
	})
}

func TestAccALBBackendGroup_httpBackendWithStreamHealthCheck(t *testing.T) {
	t.Parallel()

	BGResource := albBackendGroupInfo()
	BGResource.IsHTTPBackend = true
	BGResource.IsStreamCheck = true

	var bg apploadbalancer.BackendGroup
	backendPath := ""

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBBackendGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testALBBackendGroupConfig_basic(BGResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckALBBackendGroupExists(albBGResource, &bg),
					testAccCheckALBBackendGroupValues(&bg, true, false, false),
					testExistsFirstElementWithAttr(
						albBGResource, "http_backend", "name", &backendPath,
					),
					testExistsElementWithAttrValue(
						albBGResource, "http_backend", "healthcheck.*.timeout", albDefaultTimeout, &backendPath,
					),
					testExistsElementWithAttrValue(
						albBGResource, "http_backend", "healthcheck.*.interval", albDefaultInterval, &backendPath,
					),
					testExistsElementWithAttrValue(
						albBGResource, "http_backend", "healthcheck.*.stream_healthcheck.0.send", albDefaultSend, &backendPath,
					),
					testExistsElementWithAttrValue(
						albBGResource, "http_backend", "healthcheck.*.stream_healthcheck.0.receive", albDefaultReceive, &backendPath,
					),
				),
			},
			albBackendGroupImportStep(),
		},
	})
}

func TestAccALBBackendGroup_grpcBackendWithHttpHealthCheck(t *testing.T) {
	t.Parallel()

	BGResource := albBackendGroupInfo()
	BGResource.IsGRPCBackend = true
	BGResource.IsHTTPCheck = true

	var bg apploadbalancer.BackendGroup
	backendPath := ""

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBBackendGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testALBBackendGroupConfig_basic(BGResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckALBBackendGroupExists(albBGResource, &bg),
					testAccCheckALBBackendGroupValues(&bg, false, true, false),
					testExistsFirstElementWithAttr(
						albBGResource, "grpc_backend", "name", &backendPath,
					),
					testExistsElementWithAttrValue(
						albBGResource, "grpc_backend", "healthcheck.*.timeout", albDefaultTimeout, &backendPath,
					),
					testExistsElementWithAttrValue(
						albBGResource, "grpc_backend", "healthcheck.*.interval", albDefaultInterval, &backendPath,
					),
					testExistsElementWithAttrValue(
						albBGResource, "grpc_backend", "healthcheck.*.http_healthcheck.0.host", albDefaultHost, &backendPath,
					),
					testExistsElementWithAttrValue(
						albBGResource, "grpc_backend", "healthcheck.*.http_healthcheck.0.path", albDefaultPath, &backendPath,
					),
					testExistsElementWithAttrValue(
						albBGResource, "grpc_backend", "healthcheck.*.http_healthcheck.0.http2", albDefaultHTTP2, &backendPath,
					),
				),
			},
			albBackendGroupImportStep(),
		},
	})
}

func TestAccALBBackendGroup_grpcBackendWithGRPCHealthCheck(t *testing.T) {
	t.Parallel()

	BGResource := albBackendGroupInfo()
	BGResource.IsGRPCBackend = true
	BGResource.IsGRPCCheck = true

	var bg apploadbalancer.BackendGroup
	backendPath := ""

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBBackendGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testALBBackendGroupConfig_basic(BGResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckALBBackendGroupExists(albBGResource, &bg),
					testAccCheckALBBackendGroupValues(&bg, false, true, false),
					testExistsFirstElementWithAttr(
						albBGResource, "grpc_backend", "name", &backendPath,
					),
					testExistsElementWithAttrValue(
						albBGResource, "grpc_backend", "healthcheck.*.timeout", albDefaultTimeout, &backendPath,
					),
					testExistsElementWithAttrValue(
						albBGResource, "grpc_backend", "healthcheck.*.interval", albDefaultInterval, &backendPath,
					),
					testExistsElementWithAttrValue(
						albBGResource, "grpc_backend", "healthcheck.*.grpc_healthcheck.0.service_name", albDefaultServiceName, &backendPath,
					),
				),
			},
			albBackendGroupImportStep(),
		},
	})
}

func TestAccALBBackendGroup_grpcBackendWithStreamHealthCheck(t *testing.T) {
	t.Parallel()

	BGResource := albBackendGroupInfo()
	BGResource.IsGRPCBackend = true
	BGResource.IsStreamCheck = true

	var bg apploadbalancer.BackendGroup
	backendPath := ""

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBBackendGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testALBBackendGroupConfig_basic(BGResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckALBBackendGroupExists(albBGResource, &bg),
					testAccCheckALBBackendGroupValues(&bg, false, true, false),
					testExistsFirstElementWithAttr(
						albBGResource, "grpc_backend", "name", &backendPath,
					),
					testExistsElementWithAttrValue(
						albBGResource, "grpc_backend", "healthcheck.*.timeout", albDefaultTimeout, &backendPath,
					),
					testExistsElementWithAttrValue(
						albBGResource, "grpc_backend", "healthcheck.*.interval", albDefaultInterval, &backendPath,
					),
					testExistsElementWithAttrValue(
						albBGResource, "grpc_backend", "healthcheck.*.stream_healthcheck.0.send", albDefaultSend, &backendPath,
					),
					testExistsElementWithAttrValue(
						albBGResource, "grpc_backend", "healthcheck.*.stream_healthcheck.0.receive", albDefaultReceive, &backendPath,
					),
				),
			},
			albBackendGroupImportStep(),
		},
	})
}

func testAccCheckALBBackendGroupDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_alb_backend_group" {
			continue
		}

		_, err := config.sdk.ApplicationLoadBalancer().BackendGroup().Get(context.Background(), &apploadbalancer.GetBackendGroupRequest{
			BackendGroupId: rs.Primary.ID,
		})
		if status.Code(err) != codes.NotFound {
			return fmt.Errorf("Backend Group still exists")
		}
	}

	return nil
}

func testAccCheckALBBackendGroupExists(bgName string, bg *apploadbalancer.BackendGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[bgName]
		if !ok {
			return fmt.Errorf("Not found: %s", bgName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.ApplicationLoadBalancer().BackendGroup().Get(context.Background(), &apploadbalancer.GetBackendGroupRequest{
			BackendGroupId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Backend Group not found")
		}

		*bg = *found

		return nil
	}
}

func testAccCheckALBBackendGroupContainsLabel(bg *apploadbalancer.BackendGroup, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		v, ok := bg.Labels[key]
		if !ok {
			return fmt.Errorf("Expected label with key '%s' not found", key)
		}
		if v != value {
			return fmt.Errorf("Incorrect label value for key '%s': expected '%s' but found '%s'", key, value, v)
		}
		return nil
	}
}

func testAccALBBackendGroupBasic(name string) string {
	return fmt.Sprintf(`
resource "yandex_alb_backend_group" "test-bg" {
  name		= "%s"

  labels = {
    tf-label    = "tf-label-value"
    empty-label = ""
  }
}
`, name)
}
