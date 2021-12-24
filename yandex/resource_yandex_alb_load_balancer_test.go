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
