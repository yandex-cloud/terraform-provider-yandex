package yandex

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/apploadbalancer/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
