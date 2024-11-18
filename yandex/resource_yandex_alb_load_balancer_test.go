package yandex

import (
	"context"
	"fmt"
	"log"
	"testing"

	terraform2 "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/apploadbalancer/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const albLoadBalancerResource = "yandex_alb_load_balancer.test-balancer"

func init() {
	resource.AddTestSweepers("yandex_alb_load_balancer", &resource.Sweeper{
		Name:         "yandex_alb_load_balancer",
		F:            testSweepALBLoadBalancers,
		Dependencies: []string{},
	})
}

func testSweepALBLoadBalancers(_ string) error {
	log.Printf("[DEBUG] Sweeping ALB Load Balancer")
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	result := &multierror.Error{}

	req := &apploadbalancer.ListLoadBalancersRequest{FolderId: conf.FolderID}
	it := conf.sdk.ApplicationLoadBalancer().LoadBalancer().LoadBalancerIterator(conf.Context(), req)
	for it.Next() {
		id := it.Value().GetId()

		if !sweepALBLoadBalancer(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep ALB Load Balancer %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepALBLoadBalancer(conf *Config, id string) bool {
	return sweepWithRetry(sweepALBLoadBalancerOnce, conf, "ALB Load Balancer", id)
}

func sweepALBLoadBalancerOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexALBLoadBalancerDefaultTimeout)
	defer cancel()

	op, err := conf.sdk.ApplicationLoadBalancer().LoadBalancer().Delete(ctx, &apploadbalancer.DeleteLoadBalancerRequest{
		LoadBalancerId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}

func albLoadBalancerImportStep() resource.TestStep {
	return resource.TestStep{
		ResourceName:            albLoadBalancerResource,
		ImportState:             true,
		ImportStateVerifyIgnore: []string{"status"},
		ImportStateVerify:       true,
	}
}

func TestAccALBLoadBalancer_basic(t *testing.T) {
	t.Parallel()

	var balancer apploadbalancer.LoadBalancer
	balancerName := acctest.RandomWithPrefix("tf-load-balancer")
	balancerDescription := acctest.RandomWithPrefix("tf-load-balancer-description")
	folderID := getExampleFolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccALBLoadBalancerBasic(balancerName, balancerDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckALBLoadBalancerExists(albLoadBalancerResource, &balancer),
					resource.TestCheckResourceAttr(albLoadBalancerResource, "name", balancerName),
					resource.TestCheckResourceAttrSet(albLoadBalancerResource, "folder_id"),
					resource.TestCheckResourceAttr(albLoadBalancerResource, "folder_id", folderID),
					testAccCheckALBLoadBalancerContainsLabel(&balancer, "tf-label", "tf-label-value"),
					testAccCheckALBLoadBalancerContainsLabel(&balancer, "empty-label", ""),
					testAccCheckCreatedAtAttr(albLoadBalancerResource),
				),
			},
			albLoadBalancerImportStep(),
		},
	})
}

func TestAccALBLoadBalancer_streamListener(t *testing.T) {
	t.Parallel()

	albResource := albLoadBalancerInfo()
	albResource.IsStreamListener = true
	albResource.IsStreamHandler = true

	var alb apploadbalancer.LoadBalancer
	listenerPath := ""

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testALBLoadBalancerConfig_basic(albResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckALBLoadBalancerExists(albLoadBalancerResource, &alb),
					testExistsFirstElementWithAttr(
						albLoadBalancerResource, "listener", "name", &listenerPath,
					),
					testExistsElementWithAttrValue(
						albLoadBalancerResource, "listener", "endpoint.0.ports.0", albDefaultPort, &listenerPath,
					),
					testExistsElementWithAttrValue(
						albLoadBalancerResource, "listener", "name", albResource.ListenerName, &listenerPath,
					),
					testExistsFirstElementWithAttr(
						albLoadBalancerResource, "listener", "stream.0.handler.0.backend_group_id", &listenerPath,
					),
					testExistsFirstElementWithAttr(
						albLoadBalancerResource, "listener", "stream.0.handler.0.idle_timeout", &listenerPath,
					),
				),
			},
			albLoadBalancerImportStep(),
		},
	})
}

func TestAccALBLoadBalancer_httpListenerWithHTTP2Options(t *testing.T) {
	t.Parallel()

	albResource := albLoadBalancerInfo()
	albResource.IsHTTPListener = true
	albResource.IsHTTPHandler = true
	albResource.IsHTTP2Options = true

	var alb apploadbalancer.LoadBalancer
	listenerPath := ""

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testALBLoadBalancerConfig_basic(albResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckALBLoadBalancerExists(albLoadBalancerResource, &alb),
					testExistsFirstElementWithAttr(
						albLoadBalancerResource, "listener", "name", &listenerPath,
					),
					testExistsElementWithAttrValue(
						albLoadBalancerResource, "listener", "endpoint.0.ports.0", albDefaultPort, &listenerPath,
					),
					testExistsElementWithAttrValue(
						albLoadBalancerResource, "listener", "name", albResource.ListenerName, &listenerPath,
					),
					testExistsElementWithAttrValue(
						albLoadBalancerResource, "listener", "http.0.handler.0.http2_options.0.max_concurrent_streams", albDefaultMaxConcurrentStreams, &listenerPath,
					),
				),
			},
			albLoadBalancerImportStep(),
		},
	})
}

func TestAccALBLoadBalancer_httpListenerWithAllowHTTP10(t *testing.T) {
	t.Parallel()

	albResource := albLoadBalancerInfo()
	albResource.IsHTTPListener = true
	albResource.IsHTTPHandler = true
	albResource.IsAllowHTTP10 = true

	var alb apploadbalancer.LoadBalancer
	listenerPath := ""

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testALBLoadBalancerConfig_basic(albResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckALBLoadBalancerExists(albLoadBalancerResource, &alb),
					testExistsFirstElementWithAttr(
						albLoadBalancerResource, "listener", "name", &listenerPath,
					),
					testExistsElementWithAttrValue(
						albLoadBalancerResource, "listener", "endpoint.0.ports.0", albDefaultPort, &listenerPath,
					),
					testExistsElementWithAttrValue(
						albLoadBalancerResource, "listener", "name", albResource.ListenerName, &listenerPath,
					),
					testExistsElementWithAttrValue(
						albLoadBalancerResource, "listener", "http.0.handler.0.allow_http10", albDefaultAllowHTTP10, &listenerPath,
					),
				),
			},
			albLoadBalancerImportStep(),
		},
	})
}

func TestAccALBLoadBalancer_httpListenerWithRewriteRequestID(t *testing.T) {
	t.Parallel()

	albResource := albLoadBalancerInfo()
	albResource.IsHTTPListener = true
	albResource.IsHTTPHandler = true
	albResource.IsRewriteRequestID = true

	var alb apploadbalancer.LoadBalancer
	listenerPath := ""

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testALBLoadBalancerConfig_basic(albResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckALBLoadBalancerExists(albLoadBalancerResource, &alb),
					testExistsFirstElementWithAttr(
						albLoadBalancerResource, "listener", "name", &listenerPath,
					),
					testExistsElementWithAttrValue(
						albLoadBalancerResource, "listener", "endpoint.0.ports.0", albDefaultPort, &listenerPath,
					),
					testExistsElementWithAttrValue(
						albLoadBalancerResource, "listener", "name", albResource.ListenerName, &listenerPath,
					),
					testExistsElementWithAttrValue(
						albLoadBalancerResource, "listener", "http.0.handler.0.rewrite_request_id", albDefaultRewriteRequestID, &listenerPath,
					),
				),
			},
			albLoadBalancerImportStep(),
		},
	})
}

func TestAccALBLoadBalancer_httpListenerWithRedirects(t *testing.T) {
	t.Parallel()

	albResource := albLoadBalancerInfo()
	albResource.IsHTTPListener = true
	albResource.IsRedirects = true

	var alb apploadbalancer.LoadBalancer
	listenerPath := ""

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testALBLoadBalancerConfig_basic(albResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckALBLoadBalancerExists(albLoadBalancerResource, &alb),
					testExistsFirstElementWithAttr(
						albLoadBalancerResource, "listener", "name", &listenerPath,
					),
					testExistsElementWithAttrValue(
						albLoadBalancerResource, "listener", "endpoint.0.ports.0", albDefaultPort, &listenerPath,
					),
					testExistsElementWithAttrValue(
						albLoadBalancerResource, "listener", "name", albResource.ListenerName, &listenerPath,
					),
					testExistsElementWithAttrValue(
						albLoadBalancerResource, "listener", "http.0.redirects.0.http_to_https", albDefaultHTTPToHTTPS, &listenerPath,
					),
				),
			},
			albLoadBalancerImportStep(),
		},
	})
}

func TestAccALBLoadBalancer_tlsListenerWithStreamHandler(t *testing.T) {
	t.Parallel()

	albResource := albLoadBalancerInfo()
	albResource.IsTLSListener = true
	albResource.IsStreamHandler = true

	var alb apploadbalancer.LoadBalancer
	listenerPath := ""

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testALBLoadBalancerConfig_basic(albResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckALBLoadBalancerExists(albLoadBalancerResource, &alb),
					testExistsFirstElementWithAttr(
						albLoadBalancerResource, "listener", "name", &listenerPath,
					),
					testExistsElementWithAttrValue(
						albLoadBalancerResource, "listener", "endpoint.0.ports.0", albDefaultPort, &listenerPath,
					),
					testExistsElementWithAttrValue(
						albLoadBalancerResource, "listener", "name", albResource.ListenerName, &listenerPath,
					),
					testExistsFirstElementWithAttr(
						albLoadBalancerResource, "listener", "tls.0.default_handler.0.stream_handler.0.backend_group_id", &listenerPath,
					),
					testExistsFirstElementWithAttr(
						albLoadBalancerResource, "listener", "tls.0.default_handler.0.stream_handler.0.idle_timeout", &listenerPath,
					),
					testExistsElementWithAttrValue(
						albLoadBalancerResource, "listener", "tls.0.default_handler.0.certificate_ids.*", albResource.CertificateID, &listenerPath,
					),
					testExistsElementWithAttrValue(
						albLoadBalancerResource, "listener", "tls.0.sni_handler.0.handler.0.certificate_ids.*", albResource.CertificateID, &listenerPath,
					),
				),
			},
			albLoadBalancerImportStep(),
		},
	})
}

func TestAccALBLoadBalancer_tlsListenerWithHTTP2Options(t *testing.T) {
	t.Parallel()

	albResource := albLoadBalancerInfo()
	albResource.IsTLSListener = true
	albResource.IsHTTPHandler = true
	albResource.IsHTTP2Options = true

	var alb apploadbalancer.LoadBalancer
	listenerPath := ""

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testALBLoadBalancerConfig_basic(albResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckALBLoadBalancerExists(albLoadBalancerResource, &alb),
					testExistsFirstElementWithAttr(
						albLoadBalancerResource, "listener", "name", &listenerPath,
					),
					testExistsElementWithAttrValue(
						albLoadBalancerResource, "listener", "endpoint.0.ports.0", albDefaultPort, &listenerPath,
					),
					testExistsElementWithAttrValue(
						albLoadBalancerResource, "listener", "name", albResource.ListenerName, &listenerPath,
					),
					testExistsElementWithAttrValue(
						albLoadBalancerResource, "listener", "tls.0.default_handler.0.http_handler.0.http2_options.0.max_concurrent_streams", albDefaultMaxConcurrentStreams, &listenerPath,
					),
					testExistsElementWithAttrValue(
						albLoadBalancerResource, "listener", "tls.0.default_handler.0.certificate_ids.*", albResource.CertificateID, &listenerPath,
					),
					testExistsElementWithAttrValue(
						albLoadBalancerResource, "listener", "tls.0.sni_handler.0.handler.0.certificate_ids.*", albResource.CertificateID, &listenerPath,
					),
				),
			},
			albLoadBalancerImportStep(),
		},
	})
}

func TestAccALBLoadBalancer_tlsListenerWithAllowHTTP10(t *testing.T) {
	t.Parallel()

	albResource := albLoadBalancerInfo()
	albResource.IsTLSListener = true
	albResource.IsHTTPHandler = true
	albResource.IsAllowHTTP10 = true

	var alb apploadbalancer.LoadBalancer
	listenerPath := ""

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testALBLoadBalancerConfig_basic(albResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckALBLoadBalancerExists(albLoadBalancerResource, &alb),
					testExistsFirstElementWithAttr(
						albLoadBalancerResource, "listener", "name", &listenerPath,
					),
					testExistsElementWithAttrValue(
						albLoadBalancerResource, "listener", "endpoint.0.ports.0", albDefaultPort, &listenerPath,
					),
					testExistsElementWithAttrValue(
						albLoadBalancerResource, "listener", "name", albResource.ListenerName, &listenerPath,
					),
					testExistsElementWithAttrValue(
						albLoadBalancerResource, "listener", "tls.0.default_handler.0.http_handler.0.allow_http10", albDefaultAllowHTTP10, &listenerPath,
					),
				),
			},
			albLoadBalancerImportStep(),
		},
	})
}

func TestAccALBLoadBalancer_tlsListenerWithRewriteRequestID(t *testing.T) {
	t.Parallel()

	albResource := albLoadBalancerInfo()
	albResource.IsTLSListener = true
	albResource.IsHTTPHandler = true
	albResource.IsRewriteRequestID = true

	var alb apploadbalancer.LoadBalancer
	listenerPath := ""

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testALBLoadBalancerConfig_basic(albResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckALBLoadBalancerExists(albLoadBalancerResource, &alb),
					testExistsFirstElementWithAttr(
						albLoadBalancerResource, "listener", "name", &listenerPath,
					),
					testExistsElementWithAttrValue(
						albLoadBalancerResource, "listener", "endpoint.0.ports.0", albDefaultPort, &listenerPath,
					),
					testExistsElementWithAttrValue(
						albLoadBalancerResource, "listener", "name", albResource.ListenerName, &listenerPath,
					),
					testExistsElementWithAttrValue(
						albLoadBalancerResource, "listener", "tls.0.default_handler.0.http_handler.0.rewrite_request_id", albDefaultRewriteRequestID, &listenerPath,
					),
				),
			},
			albLoadBalancerImportStep(),
		},
	})
}

func TestAccALBLoadBalancer_update(t *testing.T) {
	var alb apploadbalancer.LoadBalancer

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccALBLoadBalancerBasic(
					"tf-alb", "tf-descr",
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckALBLoadBalancerExists(albLoadBalancerResource, &alb),
				),
			},
			{
				Config: testAccALBLoadBalancerBasic(
					"tf-alb-updated", "tf-descr-updated",
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckALBLoadBalancerExists(albLoadBalancerResource, &alb),
				),
			},
			{
				Config: testAccALBLoadBalancerBasic(
					"tf-alb-updated", "tf-descr-updated",
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckALBLoadBalancerExists(albLoadBalancerResource, &alb),
				),
			},
			albLoadBalancerImportStep(),
		},
	})
}

func TestAccALBLoadBalancer_logOptions(t *testing.T) {
	t.Parallel()
	albResource := albLoadBalancerInfo()
	albResource.IsLogOptions = true

	var alb apploadbalancer.LoadBalancer
	var rulesPath string

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testALBLoadBalancerConfig_basic(albResource),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckALBLoadBalancerExists(albLoadBalancerResource, &alb),
					testExistsFirstElementWithAttr(
						albLoadBalancerResource, "log_options", "discard_rule.0.http_codes", &rulesPath,
					),
					testExistsFirstElementWithAttr(
						albLoadBalancerResource, "log_options", "discard_rule.0.http_code_intervals", &rulesPath,
					),
					testExistsFirstElementWithAttr(
						albLoadBalancerResource, "log_options", "discard_rule.0.grpc_codes", &rulesPath,
					),
					testExistsFirstElementWithAttr(
						albLoadBalancerResource, "log_options", "discard_rule.1.http_code_intervals", &rulesPath,
					),
					testExistsFirstElementWithAttr(
						albLoadBalancerResource, "log_options", "disable", &rulesPath,
					),
				),
			},
			albLoadBalancerImportStep(),
		},
	})
}

func testAccCheckALBLoadBalancerDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_alb_load_balancer" {
			continue
		}

		_, err := config.sdk.ApplicationLoadBalancer().LoadBalancer().Get(context.Background(), &apploadbalancer.GetLoadBalancerRequest{
			LoadBalancerId: rs.Primary.ID,
		})
		if status.Code(err) != codes.NotFound {
			return fmt.Errorf("ALB Load Balancer still exists")
		}
	}

	return nil
}

func testAccCheckALBLoadBalancerExists(balancerName string, balancer *apploadbalancer.LoadBalancer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[balancerName]
		if !ok {
			return fmt.Errorf("Not found: %s", balancerName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.ApplicationLoadBalancer().LoadBalancer().Get(context.Background(), &apploadbalancer.GetLoadBalancerRequest{
			LoadBalancerId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("ALB Load Balancer not found")
		}

		*balancer = *found

		return nil
	}
}

func testAccCheckALBLoadBalancerContainsLabel(balancer *apploadbalancer.LoadBalancer, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		v, ok := balancer.Labels[key]
		if !ok {
			return fmt.Errorf("Expected label with key '%s' not found", key)
		}
		if v != value {
			return fmt.Errorf("Incorrect label value for key '%s': expected '%s' but found '%s'", key, value, v)
		}
		return nil
	}
}

func testAccALBLoadBalancerBasic(name, desc string) string {
	return fmt.Sprintf(`
resource "yandex_alb_load_balancer" "test-balancer" {
  name        = "%s"
  description = "%s"

  network_id = yandex_vpc_network.test-network.id

  security_group_ids = [yandex_vpc_security_group.test-security-group.id]

  labels = {
    tf-label    = "tf-label-value"
    empty-label = ""
  }
  
  allocation_policy {
    location {
      zone_id   = "ru-central1-a"
      subnet_id = yandex_vpc_subnet.test-subnet.id 
    }
  }
}

resource "yandex_vpc_network" "test-network" {}

resource "yandex_vpc_subnet" "test-subnet" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.test-network.id
  v4_cidr_blocks = ["192.168.0.0/24"]
}

resource "yandex_vpc_security_group" "test-security-group" {
  network_id = yandex_vpc_network.test-network.id

  ingress {
    protocol       = "TCP"
    description    = "healthchecks"
    port           = 30080
    v4_cidr_blocks = ["198.18.235.0/24", "198.18.248.0/24"]
  }
}
`, name, desc)
}

func testMakeAllocations(zones ...string) interface{} {
	var locs []interface{}
	for _, z := range zones {
		locs = append(locs, map[string]interface{}{
			"zone_id":         z,
			"subnet_id":       "subnet" + z,
			"disable_traffic": false,
		})
	}
	return []interface{}{
		map[string]interface{}{
			"location": locs,
		},
	}
}

func TestUnitALBLoadBalancerCreateFromResource(t *testing.T) {
	t.Parallel()

	lbResource := resourceYandexALBLoadBalancer()

	t.Run("missing-alloc-policy", func(t *testing.T) {
		rawValues := map[string]interface{}{
			"id":   "lbid",
			"name": "lb-name",
		}
		resourceData := schema.TestResourceDataRaw(t, lbResource.Schema, rawValues)

		resourceData.SetId("lbid")

		config := Config{
			FolderID: "folder1",
		}
		_, err := buildALBLoadBalancerCreateRequest(resourceData, &config)
		assert.Error(t, err)
	})

	t.Run("alloc-policy", func(t *testing.T) {
		rawValues := map[string]interface{}{
			"id":                "lbid",
			"name":              "lb-name",
			"allocation_policy": testMakeAllocations("1", "2"),
		}
		resourceData := schema.TestResourceDataRaw(t, lbResource.Schema, rawValues)

		resourceData.SetId("lbid")

		t.Log(rawValues, resourceData, resourceData.Get("allocation_policy.0.location.2"))

		config := Config{
			FolderID: "folder1",
		}
		req, err := buildALBLoadBalancerCreateRequest(resourceData, &config)
		require.NoError(t, err, "failed to build create request")

		assert.Equal(t, req.GetFolderId(), "folder1")
		assert.Equal(t, req.GetName(), "lb-name")
		assert.NotNil(t, req.GetAllocationPolicy())
		assert.Len(t, req.GetAllocationPolicy().GetLocations(), 2)
	})

	t.Run("empty-log-options", func(t *testing.T) {
		rawValues := map[string]interface{}{
			"id":                "lbid",
			"name":              "lb-name",
			"log_options":       []interface{}{nil},
			"allocation_policy": testMakeAllocations("1"),
		}
		resourceData := schema.TestResourceDataRaw(t, lbResource.Schema, rawValues)

		resourceData.SetId("lbid")

		config := Config{
			FolderID: "folder1",
		}
		req, err := buildALBLoadBalancerCreateRequest(resourceData, &config)
		require.NoError(t, err, "failed to build create request")

		assert.Equal(t, req.GetFolderId(), "folder1")
		assert.Equal(t, req.GetName(), "lb-name")
		assert.NotNil(t, req.GetLogOptions())
	})

	t.Run("full-log-options", func(t *testing.T) {
		rawValues := map[string]interface{}{
			"id":   "lbid",
			"name": "lb-name",
			"log_options": []interface{}{
				map[string]interface{}{
					"disable": true,
					"discard_rule": []interface{}{
						map[string]interface{}{
							"discard_percent": 99,
						},
					},
					"log_group_id": "lg1",
				},
			},
			"allocation_policy": testMakeAllocations("1"),
		}
		resourceData := schema.TestResourceDataRaw(t, lbResource.Schema, rawValues)

		resourceData.SetId("lbid")

		config := Config{
			FolderID: "folder1",
		}
		req, err := buildALBLoadBalancerCreateRequest(resourceData, &config)
		require.NoError(t, err, "failed to build create request")

		assert.Equal(t, req.GetFolderId(), "folder1")
		assert.Equal(t, req.GetName(), "lb-name")
		assert.NotNil(t, req.GetLogOptions())
		assert.True(t, req.GetLogOptions().GetDisable())
		assert.Equal(t, "lg1", req.GetLogOptions().GetLogGroupId())
		require.Len(t, req.GetLogOptions().GetDiscardRules(), 1)
		assert.EqualValues(t, 99, req.GetLogOptions().GetDiscardRules()[0].GetDiscardPercent().GetValue())
	})
}

// these tests were write special for CLOUD-169947 and common cases
func TestUnitALBLoadBalancerValidateListenerTypeAttributes(t *testing.T) {
	t.Parallel()

	type modifiedAttributes struct {
		http   map[string]*terraform2.ResourceAttrDiff
		tls    map[string]*terraform2.ResourceAttrDiff
		stream map[string]*terraform2.ResourceAttrDiff
	}

	// common diff initial
	arrayInit := &terraform2.ResourceAttrDiff{Old: "0", New: "1"}
	arrayDelete := &terraform2.ResourceAttrDiff{Old: "1", New: "0"}
	falseInit := &terraform2.ResourceAttrDiff{Old: "", New: "false"}
	zeroInit := &terraform2.ResourceAttrDiff{Old: "1", New: "0"}
	emptyStringInit := &terraform2.ResourceAttrDiff{Old: "empty", New: ""}

	filledStringInit := &terraform2.ResourceAttrDiff{Old: "", New: "123"}

	// http
	emptyHttpDiff := map[string]*terraform2.ResourceAttrDiff{
		"listener.0.http.#": arrayInit,
		// handler
		"listener.0.http.0.handler.#":                    arrayInit,
		"listener.0.http.0.handler.0.http_router_id":     emptyStringInit,
		"listener.0.http.0.handler.0.rewrite_request_id": falseInit,
		"listener.0.http.0.handler.0.allow_http10":       falseInit,
		// handler {http2}
		"listener.0.http.0.handler.0.http2_options.#":                        arrayInit,
		"listener.0.http.0.handler.0.http2_options.0.max_concurrent_streams": zeroInit,
		// redirects
		"listener.0.http.0.redirects.#":               arrayInit,
		"listener.0.http.0.redirects.0.http_to_https": falseInit,
	}
	filledHttpDiff := map[string]*terraform2.ResourceAttrDiff{
		"listener.0.http.#": arrayInit,
		// handler
		"listener.0.http.0.handler.#":                arrayInit,
		"listener.0.http.0.handler.0.http_router_id": filledStringInit,
	}

	// tls
	emptyTlsDefaultHttpHandler := map[string]*terraform2.ResourceAttrDiff{
		"listener.0.tls.#": arrayInit,
		// default_handler
		"listener.0.tls.0.default_handler.#": arrayInit,
		//default_handler {http_handler}
		"listener.0.tls.0.default_handler.0.http_handler.#":                    arrayInit,
		"listener.0.tls.0.default_handler.0.http_handler.0.http_router_id":     emptyStringInit,
		"listener.0.tls.0.default_handler.0.http_handler.0.rewrite_request_id": falseInit,
		"listener.0.tls.0.default_handler.0.http_handler.0.allow_http10":       falseInit,
		// default_handler {http_handler {http2_options}}
		"listener.0.tls.0.default_handler.0.http_handler.0.http2_options.#":                        arrayInit,
		"listener.0.tls.0.default_handler.0.http_handler.0.http2_options.0.max_concurrent_streams": zeroInit,
		//default_handler {certificate_ids}
		"listener.0.tls.0.default_handler.0.certificate_ids.#": arrayDelete,
	}

	emptyTlsDefaultStreamHandler := map[string]*terraform2.ResourceAttrDiff{
		"listener.0.tls.#": arrayInit,
		// default_handler
		"listener.0.tls.0.default_handler.#": arrayInit,
		//default_handler {stream_handler}
		"listener.0.tls.0.default_handler.0.stream_handler.#":                  arrayInit,
		"listener.0.tls.0.default_handler.0.stream_handler.0.backend_group_id": emptyStringInit,
		//default_handler {certificate_ids}
		"listener.0.tls.0.default_handler.0.certificate_ids.#": arrayDelete,
	}

	emptyTlsSniHttpHandler := map[string]*terraform2.ResourceAttrDiff{
		"listener.0.tls.#": arrayInit,
		// sni_handler
		"listener.0.tls.0.sni_handler.#":              arrayInit,
		"listener.0.tls.0.sni_handler.0.name":         emptyStringInit,
		"listener.0.tls.0.sni_handler.0.server_names": arrayDelete,
		// sni_handler {handler}
		"listener.0.tls.0.sni_handler.0.handler.#": arrayInit,
		// sni_handler {handler {http_handler}}
		"listener.0.tls.0.sni_handler.0.handler.0.http_handler.#":                    arrayInit,
		"listener.0.tls.0.sni_handler.0.handler.0.http_handler.0.http_router_id":     emptyStringInit,
		"listener.0.tls.0.sni_handler.0.handler.0.http_handler.0.rewrite_request_id": falseInit,
		"listener.0.tls.0.sni_handler.0.handler.0.http_handler.0.allow_http10":       falseInit,
		// sni_handler {handler {http_handler{http2_options}}}
		"listener.0.tls.0.sni_handler.0.handler.0.http_handler.0.http2_options.#":                        arrayInit,
		"listener.0.tls.0.sni_handler.0.handler.0.http_handler.0.http2_options.0.max_concurrent_streams": zeroInit,
		// sni_handler {handler {certificates}}
		"listener.0.tls.0.sni_handler.0.handler.0.certificate_ids.#": arrayDelete,
	}

	emptyTlsSniStreamHandler := map[string]*terraform2.ResourceAttrDiff{
		"listener.0.tls.#": arrayInit,
		// sni_handler
		"listener.0.tls.0.sni_handler.#":              arrayInit,
		"listener.0.tls.0.sni_handler.0.name":         emptyStringInit,
		"listener.0.tls.0.sni_handler.0.server_names": arrayDelete,
		// sni_handler {handler}
		"listener.0.tls.0.sni_handler.0.handler.#": arrayInit,
		// sni_handler {handler {stream_handler}}
		"listener.0.tls.0.sni_handler.0.handler.0.stream_handler.#":                  arrayInit,
		"listener.0.tls.0.sni_handler.0.handler.0.stream_handler.0.backend_group_id": emptyStringInit,
		// sni_handler {handler {certificates}}
		"listener.0.tls.0.sni_handler.0.handler.0.certificate_ids.#": arrayDelete,
	}
	updateTlsDefaultHttpToStreamHandler := map[string]*terraform2.ResourceAttrDiff{
		"listener.0.tls.#": arrayInit,
		// default_handler
		"listener.0.tls.0.default_handler.#": arrayInit,
		// default_handler {http_handler}
		"listener.0.tls.0.default_handler.0.http_handler.#":                    arrayDelete,
		"listener.0.tls.0.default_handler.0.http_handler.0.http_router_id":     emptyStringInit,
		"listener.0.tls.0.default_handler.0.http_handler.0.rewrite_request_id": falseInit,
		"listener.0.tls.0.default_handler.0.http_handler.0.allow_http10":       falseInit,
		// default_handler {http_handler {http2_options}}
		"listener.0.tls.0.default_handler.0.http_handler.0.http2_options.#":                        arrayDelete,
		"listener.0.tls.0.default_handler.0.http_handler.0.http2_options.0.max_concurrent_streams": zeroInit,
		// default_handler {stream_handler}
		"listener.0.tls.0.default_handler.0.stream_handler.#":                  arrayInit,
		"listener.0.tls.0.default_handler.0.stream_handler.0.backend_group_id": filledStringInit,
		// default_handler {certificate_ids}
		"listener.0.tls.0.default_handler.0.certificate_ids.#": arrayDelete,
	}
	updateTlsDefaultStreamToHttpHandler := map[string]*terraform2.ResourceAttrDiff{
		"listener.0.tls.#": arrayInit,
		// default_handler
		"listener.0.tls.0.default_handler.#": arrayInit,
		// default_handler {http_handler}
		"listener.0.tls.0.default_handler.0.http_handler.#":                arrayInit,
		"listener.0.tls.0.default_handler.0.http_handler.0.http_router_id": filledStringInit,
		"listener.0.tls.0.default_handler.0.http_handler.0.allow_http10":   {Old: "", New: "true"},
		// default_handler {stream_handler}
		"listener.0.tls.0.default_handler.0.stream_handler.#":                  arrayDelete,
		"listener.0.tls.0.default_handler.0.stream_handler.0.backend_group_id": emptyStringInit,
		// default_handler {certificate_ids}
		"listener.0.tls.0.default_handler.0.certificate_ids.#": arrayDelete,
	}
	filledTlsDefaultStreamHandler := map[string]*terraform2.ResourceAttrDiff{
		"listener.0.tls.#": arrayInit,
		// default_handler
		"listener.0.tls.0.default_handler.#": arrayInit,
		// default_handler {stream_handler}
		"listener.0.tls.0.default_handler.0.stream_handler.#":                  arrayInit,
		"listener.0.tls.0.default_handler.0.stream_handler.0.backend_group_id": filledStringInit,
	}

	// stream
	emptyStreamHandler := map[string]*terraform2.ResourceAttrDiff{
		"listener.0.stream.#": arrayInit,
		// handler
		"listener.0.stream.0.handler.#":                  arrayInit,
		"listener.0.stream.0.handler.0.backend_group_id": emptyStringInit,
	}

	filledStreamHandler := map[string]*terraform2.ResourceAttrDiff{
		"listener.0.stream.#": arrayInit,
		// handler
		"listener.0.stream.0.handler.#":                  arrayInit,
		"listener.0.stream.0.handler.0.backend_group_id": filledStringInit,
	}

	tests := []struct {
		name               string
		modifiedAttributes modifiedAttributes
		isError            bool
	}{
		{
			name:    "no listeners",
			isError: true,
		},
		{
			name: "empty diffs for listeners",
			modifiedAttributes: modifiedAttributes{
				http:   emptyHttpDiff,
				tls:    emptyTlsDefaultHttpHandler,
				stream: emptyStreamHandler,
			},
			isError: true,
		},
		{
			name: "http listener is filled other listeners are empty",
			modifiedAttributes: modifiedAttributes{
				http:   filledHttpDiff,
				tls:    emptyTlsDefaultHttpHandler,
				stream: emptyStreamHandler,
			},
			isError: false,
		},
		{
			name: "tls listener is filled other listeners are empty",
			modifiedAttributes: modifiedAttributes{
				http:   emptyHttpDiff,
				tls:    filledTlsDefaultStreamHandler,
				stream: emptyStreamHandler,
			},
			isError: false,
		},
		{
			name: "stream listener is filled other listeners are empty",
			modifiedAttributes: modifiedAttributes{
				http:   emptyHttpDiff,
				tls:    emptyTlsDefaultHttpHandler,
				stream: filledStreamHandler,
			},
			isError: false,
		},
		{
			name: "http listener is empty other listeners are filled",
			modifiedAttributes: modifiedAttributes{
				http:   emptyHttpDiff,
				tls:    filledTlsDefaultStreamHandler,
				stream: filledStreamHandler,
			},
			isError: true,
		},
		{
			name: "tls listener is empty other listeners are filled",
			modifiedAttributes: modifiedAttributes{
				http:   filledHttpDiff,
				tls:    emptyTlsDefaultHttpHandler,
				stream: filledStreamHandler,
			},
			isError: true,
		},
		{
			name: "stream listener is empty other listeners are filled",
			modifiedAttributes: modifiedAttributes{
				http:   filledHttpDiff,
				tls:    filledTlsDefaultStreamHandler,
				stream: emptyStreamHandler,
			},
			isError: true,
		},
		{
			name: "all listeners are filled",
			modifiedAttributes: modifiedAttributes{
				http:   filledHttpDiff,
				tls:    filledTlsDefaultStreamHandler,
				stream: filledStreamHandler,
			},
			isError: true,
		},
		{
			name: "tcp listener empty with default stream handler",
			modifiedAttributes: modifiedAttributes{
				http:   filledHttpDiff,
				tls:    emptyTlsDefaultStreamHandler,
				stream: emptyStreamHandler,
			},
			isError: false,
		},
		{
			name: "tcp listener empty with sni http handler",
			modifiedAttributes: modifiedAttributes{
				http:   filledHttpDiff,
				tls:    emptyTlsSniHttpHandler,
				stream: emptyStreamHandler,
			},
			isError: false,
		},
		{
			name: "tcp listener empty with sni stream handler",
			modifiedAttributes: modifiedAttributes{
				http:   filledHttpDiff,
				tls:    emptyTlsSniStreamHandler,
				stream: emptyStreamHandler,
			},
			isError: false,
		},
		{
			name: "update in tls default_header http header to stream",
			modifiedAttributes: modifiedAttributes{
				tls: updateTlsDefaultHttpToStreamHandler,
			},
			isError: false,
		},
		{
			name: "update in tls default_header stream header to http",
			modifiedAttributes: modifiedAttributes{
				tls: updateTlsDefaultStreamToHttpHandler,
			},
			isError: false,
		},
	}

	rawValues := map[string]interface{}{
		"id":   "lbid",
		"name": "lb-name",
		"listener": []interface{}{map[string]interface{}{
			"name": "test-listener",
		}},
		"allocation_policy": testMakeAllocations("1"),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resModifiedAttributes :=
				mergeMaps(tt.modifiedAttributes.http, tt.modifiedAttributes.tls, tt.modifiedAttributes.stream)

			resourceData := createResourceDataWithModifiedAttributes(t,
				resourceYandexALBLoadBalancer().Schema,
				rawValues,
				resModifiedAttributes)

			resourceData.SetId("lbid")

			config := Config{
				FolderID: "folder1",
			}
			_, err := buildALBLoadBalancerCreateRequest(resourceData, &config)
			if tt.isError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func createResourceDataWithModifiedAttributes(t *testing.T, schemaObject map[string]*schema.Schema, rawInitialState map[string]interface{},
	modifiedAttributes map[string]*terraform2.ResourceAttrDiff,
) *schema.ResourceData {
	t.Helper()
	ctx := context.Background()
	internalMap := schema.InternalMap(schemaObject)

	initialDiff, err := internalMap.Diff(ctx, nil,
		terraform2.NewResourceConfigRaw(rawInitialState),
		nil, nil, true)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	for key, diffState := range modifiedAttributes {
		initialDiff.Attributes[key] = diffState
	}

	resourceData, err := internalMap.Data(nil, initialDiff)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	return resourceData
}

func mergeMaps[K comparable, V any](maps ...map[K]V) map[K]V {
	merged := make(map[K]V)

	for _, m := range maps {
		for key, value := range m {
			merged[key] = value
		}
	}

	return merged
}

func Test_redirectsDiffSuppress(t *testing.T) {
	t.Parallel()

	testsTable := []struct {
		name           string
		key            string
		oldValue       string
		newValue       string
		oldState       map[string]string
		newState       map[string]interface{}
		expectPanic    bool
		expectedResult bool
	}{
		{
			name:        "unexpected resource key",
			key:         fmt.Sprintf("listener.0.http.0.%v.0.%v", resourceNameRedirects, "unexpected_key"),
			oldValue:    "false",
			newValue:    "false",
			expectPanic: true,
		},
		{
			name:           "compare inner fields: no changes: http_to_https is false",
			key:            fmt.Sprintf("listener.0.http.0.%v.0.%v", resourceNameRedirects, resourceNameHTTPToHTTPS),
			oldValue:       "false",
			newValue:       "false",
			expectedResult: true,
		},
		{
			name:           "compare inner fields: no changes: http_to_https is true",
			key:            fmt.Sprintf("listener.0.http.0.%v.0.%v", resourceNameRedirects, resourceNameHTTPToHTTPS),
			oldValue:       "true",
			newValue:       "true",
			expectedResult: true,
		},
		{
			name:     "compare inner fields: has changes: old http_to_https is 'false', new one is 'true'",
			key:      fmt.Sprintf("listener.0.http.0.%v.0.%v", resourceNameRedirects, resourceNameHTTPToHTTPS),
			oldValue: "false",
			newValue: "true",
		},
		{
			name:     "compare inner fields: has changes: old http_to_https is 'true', new one is 'false'",
			key:      fmt.Sprintf("listener.0.http.0.%v.0.%v", resourceNameRedirects, resourceNameHTTPToHTTPS),
			oldValue: "true",
			newValue: "false",
		},
		{
			name:     "compare inner fields: has changes: old http_to_https is empty, new one is 'true'",
			key:      fmt.Sprintf("listener.0.http.0.%v.0.%v", resourceNameRedirects, resourceNameHTTPToHTTPS),
			newValue: "true",
		},
		{
			name:     "compare inner fields: has changes: old http_to_https is empty, new one is 'false'",
			key:      fmt.Sprintf("listener.0.http.0.%v.0.%v", resourceNameRedirects, resourceNameHTTPToHTTPS),
			newValue: "false",
		},
		{
			name:     "compare inner fields: has changes: old http_to_https is 'false', new one is empty",
			key:      fmt.Sprintf("listener.0.http.0.%v.0.%v", resourceNameRedirects, resourceNameHTTPToHTTPS),
			oldValue: "false",
		},
		{
			name:     "compare inner fields: has changes: old http_to_https is 'true', new one is empty",
			key:      fmt.Sprintf("listener.0.http.0.%v.0.%v", resourceNameRedirects, resourceNameHTTPToHTTPS),
			oldValue: "true",
		},
		{
			name: "compare redirects: too many elements for old state",
			key:  fmt.Sprintf("listener.0.http.0.%v.#", resourceNameRedirects),
			oldState: map[string]string{
				"listener.0.http.0.redirects.#": "2",
				fmt.Sprintf("listener.0.http.0.%v.0.%v", resourceNameRedirects, resourceNameHTTPToHTTPS): "false",
				fmt.Sprintf("listener.0.http.0.%v.1.%v", resourceNameRedirects, resourceNameHTTPToHTTPS): "true",
			},
			newState: map[string]interface{}{
				"listener": []interface{}{
					map[string]interface{}{
						"http": []interface{}{
							map[string]interface{}{
								"redirects": []interface{}{
									map[string]interface{}{
										resourceNameHTTPToHTTPS: "false",
									},
								},
							},
						},
					},
				},
			},
			expectPanic: true,
		},
		{
			name: "compare redirects: too many elements for new state",
			key:  fmt.Sprintf("listener.0.http.0.%v.#", resourceNameRedirects),
			oldState: map[string]string{
				fmt.Sprintf("listener.0.http.0.%v.#", resourceNameRedirects):                             "1",
				fmt.Sprintf("listener.0.http.0.%v.0.%v", resourceNameRedirects, resourceNameHTTPToHTTPS): "false",
			},
			newState: map[string]interface{}{
				"listener": []interface{}{
					map[string]interface{}{
						"http": []interface{}{
							map[string]interface{}{
								resourceNameRedirects: []interface{}{
									map[string]interface{}{
										resourceNameHTTPToHTTPS: "false",
									},
									map[string]interface{}{
										resourceNameHTTPToHTTPS: "true",
									},
								},
							},
						},
					},
				},
			},
			expectPanic: true,
		},
		{
			name: "compare redirects: no changes: no redirects",
			key:  fmt.Sprintf("listener.0.http.0.%v.#", resourceNameRedirects),
			oldState: map[string]string{
				fmt.Sprintf("listener.0.http.0.%v.#", resourceNameRedirects): "0",
			},
			newState: map[string]interface{}{
				"listener": []interface{}{
					map[string]interface{}{
						"http": []interface{}{
							map[string]interface{}{
								resourceNameRedirects: []interface{}{},
							},
						},
					},
				},
			},
			expectedResult: true,
		},
		{
			name: "compare redirects: no changes: add new redirect as empty object",
			key:  fmt.Sprintf("listener.0.http.0.%v.#", resourceNameRedirects),
			oldState: map[string]string{
				fmt.Sprintf("listener.0.http.0.%v.#", resourceNameRedirects): "0",
			},
			newState: map[string]interface{}{
				"listener": []interface{}{
					map[string]interface{}{
						"http": []interface{}{
							map[string]interface{}{
								resourceNameRedirects: []interface{}{
									map[string]interface{}{},
								},
							},
						},
					},
				},
			},
			expectedResult: true,
		},
		{
			name: "compare redirects: no changes: add new redirect object with zero values",
			key:  fmt.Sprintf("listener.0.http.0.%v.#", resourceNameRedirects),
			oldState: map[string]string{
				fmt.Sprintf("listener.0.http.0.%v.#", resourceNameRedirects): "0",
			},
			newState: map[string]interface{}{
				"listener": []interface{}{
					map[string]interface{}{
						"http": []interface{}{
							map[string]interface{}{
								resourceNameRedirects: []interface{}{
									map[string]interface{}{
										resourceNameHTTPToHTTPS: "false",
									},
								},
							},
						},
					},
				},
			},
			expectedResult: true,
		},
		{
			name: "compare redirects: no changes: add new redirect object",
			key:  fmt.Sprintf("listener.0.http.0.%v.#", resourceNameRedirects),
			oldState: map[string]string{
				fmt.Sprintf("listener.0.http.0.%v.#", resourceNameRedirects): "0",
			},
			newState: map[string]interface{}{
				"listener": []interface{}{
					map[string]interface{}{
						"http": []interface{}{
							map[string]interface{}{
								"redirects": []interface{}{
									map[string]interface{}{
										resourceNameHTTPToHTTPS: "true",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "compare redirects: no changes: remove redirect object with zero values",
			key:  fmt.Sprintf("listener.0.http.0.%v.#", resourceNameRedirects),
			oldState: map[string]string{
				fmt.Sprintf("listener.0.http.0.%v.#", resourceNameRedirects):                             "1",
				fmt.Sprintf("listener.0.http.0.%v.0.%v", resourceNameRedirects, resourceNameHTTPToHTTPS): "false",
			},
			newState: map[string]interface{}{
				"listener": []interface{}{
					map[string]interface{}{
						"http": []interface{}{
							map[string]interface{}{},
						},
					},
				},
			},
			expectedResult: true,
		},
		{
			name: "compare redirects: has changes: remove redirect object",
			key:  fmt.Sprintf("listener.0.http.0.%v.#", resourceNameRedirects),
			oldState: map[string]string{
				fmt.Sprintf("listener.0.http.0.%v.#", resourceNameRedirects):                             "1",
				fmt.Sprintf("listener.0.http.0.%v.0.%v", resourceNameRedirects, resourceNameHTTPToHTTPS): "true",
			},
			newState: map[string]interface{}{
				"listener": []interface{}{
					map[string]interface{}{
						"http": []interface{}{
							map[string]interface{}{
								resourceNameRedirects: []interface{}{nil},
							},
						},
					},
				},
			},
		},
		{
			name: "compare redirects: has changes: change redirect object: http_to_https = false -> true",
			key:  fmt.Sprintf("listener.0.http.0.%v.#", resourceNameRedirects),
			oldState: map[string]string{
				fmt.Sprintf("listener.0.http.0.%v.#", resourceNameRedirects):                             "1",
				fmt.Sprintf("listener.0.http.0.%v.0.%v", resourceNameRedirects, resourceNameHTTPToHTTPS): "false",
			},
			newState: map[string]interface{}{
				"listener": []interface{}{
					map[string]interface{}{
						"http": []interface{}{
							map[string]interface{}{
								resourceNameRedirects: []interface{}{
									map[string]interface{}{
										resourceNameHTTPToHTTPS: "true",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "compare redirects: has changes: change redirect object: http_to_https = true -> false",
			key:  fmt.Sprintf("listener.0.http.0.%v.#", resourceNameRedirects),
			oldState: map[string]string{
				fmt.Sprintf("listener.0.http.0.%v.#", resourceNameRedirects):                             "1",
				fmt.Sprintf("listener.0.http.0.%v.0.%v", resourceNameRedirects, resourceNameHTTPToHTTPS): "true",
			},
			newState: map[string]interface{}{
				"listener": []interface{}{
					map[string]interface{}{
						"http": []interface{}{
							map[string]interface{}{
								resourceNameRedirects: []interface{}{
									map[string]interface{}{
										resourceNameHTTPToHTTPS: "false",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, testCase := range testsTable {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// initialize terraform resource data.

			if testCase.expectPanic {
				assert.Panics(t, func() {
					data := terraformResourceData(t, testCase.oldState, testCase.newState)

					redirectsDiffSuppress(
						testCase.key,
						testCase.oldValue,
						testCase.newValue,
						data,
					)
				})
				return
			}

			data := terraformResourceData(t, testCase.oldState, testCase.newState)

			actualResult := redirectsDiffSuppress(
				testCase.key,
				testCase.oldValue,
				testCase.newValue,
				data,
			)

			assert.Equal(t, testCase.expectedResult, actualResult)
		})
	}
}

func terraformResourceData(t *testing.T, oldState map[string]string, newState map[string]interface{}) *schema.ResourceData {
	config := terraform2.NewResourceConfigRaw(newState)

	sm := schema.InternalMap(resourceYandexALBLoadBalancer().Schema)
	diff, err := sm.Diff(context.Background(), nil, config, nil, nil, false)
	require.NoError(t, err)

	data, err := sm.Data(&terraform2.InstanceState{Attributes: oldState}, diff)
	require.NoError(t, err)

	return data
}
