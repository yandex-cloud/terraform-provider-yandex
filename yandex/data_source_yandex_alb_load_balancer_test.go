package yandex

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/apploadbalancer/v1"
)

const albLoadBalancerDataSourceResource = "data.yandex_alb_load_balancer.test-alb-ds"

func TestAccDataSourceALBLoadBalancer_byID(t *testing.T) {
	t.Parallel()

	albName := acctest.RandomWithPrefix("tf-alb")
	albDesc := "Description for test"
	folderID := getExampleFolderID()

	var loadBalancer apploadbalancer.LoadBalancer

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceALBLoadBalancerConfigByID(albName, albDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceALBLoadBalancerExists(albLoadBalancerDataSourceResource, &loadBalancer),
					testAccCheckResourceIDField(albLoadBalancerDataSourceResource, "load_balancer_id"),
					resource.TestCheckResourceAttr(albLoadBalancerDataSourceResource, "name", albName),
					resource.TestCheckResourceAttr(albLoadBalancerDataSourceResource, "description", albDesc),
					resource.TestCheckResourceAttr(albLoadBalancerDataSourceResource, "folder_id", folderID),
					testAccCheckCreatedAtAttr(albLoadBalancerDataSourceResource),
				),
			},
		},
	})
}

func TestAccDataSourceALBLoadBalancer_byName(t *testing.T) {
	t.Parallel()

	albName := acctest.RandomWithPrefix("tf-alb")
	albDesc := "Description for test"
	folderID := getExampleFolderID()

	var loadBalancer apploadbalancer.LoadBalancer

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBLoadBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceALBLoadBalancerConfigByName(albName, albDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceALBLoadBalancerExists(albLoadBalancerDataSourceResource, &loadBalancer),
					testAccCheckResourceIDField(albLoadBalancerDataSourceResource, "load_balancer_id"),
					resource.TestCheckResourceAttr(albLoadBalancerDataSourceResource, "name", albName),
					resource.TestCheckResourceAttr(albLoadBalancerDataSourceResource, "description", albDesc),
					resource.TestCheckResourceAttr(albLoadBalancerDataSourceResource, "folder_id", folderID),
					testAccCheckCreatedAtAttr(albLoadBalancerDataSourceResource),
				),
			},
		},
	})
}

func TestAccDataSourceALBLoadBalancer_streamListener(t *testing.T) {
	t.Parallel()

	albResource := albLoadBalancerInfo()
	albResource.IsStreamListener = true
	albResource.IsStreamHandler = true
	albResource.IsDataSource = true

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
					testAccDataSourceALBLoadBalancerExists(albLoadBalancerDataSourceResource, &alb),
					testExistsFirstElementWithAttr(
						albLoadBalancerDataSourceResource, "listener", "name", &listenerPath,
					),
					testCheckResourceSubAttrFn(
						albLoadBalancerDataSourceResource, &listenerPath, "name", func(value string) error {
							albName := alb.GetListeners()[0].GetName()
							if value != albName {
								return fmt.Errorf("ALB Load Balancer's listener's name doesnt't match. %s != %s", value, albName)
							}
							return nil
						},
					),
					testExistsFirstElementWithAttr(
						albLoadBalancerDataSourceResource, "listener", "endpoint.0.ports.0", &listenerPath,
					),
					testCheckResourceSubAttrFn(
						albLoadBalancerDataSourceResource, &listenerPath, "endpoint.0.ports.0", func(value string) error {
							port := alb.GetListeners()[0].GetEndpoints()[0].GetPorts()[0]
							if realValue, _ := strconv.ParseInt(value, 10, 64); realValue != port {
								return fmt.Errorf("ALB Load Balancer's listener's endpoint's port doesnt't match. %d != %d", realValue, port)
							}
							return nil
						},
					),
					testExistsFirstElementWithAttr(
						albLoadBalancerDataSourceResource, "listener", "stream.0.handler.0.backend_group", &listenerPath,
					),
				),
			},
		},
	})
}

func TestAccDataSourceALBLoadBalancer_httpListenerWithHTTP2Options(t *testing.T) {
	t.Parallel()

	albResource := albLoadBalancerInfo()
	albResource.IsHTTPListener = true
	albResource.IsHTTPHandler = true
	albResource.IsHTTP2Options = true
	albResource.IsDataSource = true

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
					testAccDataSourceALBLoadBalancerExists(albLoadBalancerDataSourceResource, &alb),
					testExistsFirstElementWithAttr(
						albLoadBalancerDataSourceResource, "listener", "name", &listenerPath,
					),
					testCheckResourceSubAttrFn(
						albLoadBalancerDataSourceResource, &listenerPath, "name", func(value string) error {
							albName := alb.GetListeners()[0].GetName()
							if value != albName {
								return fmt.Errorf("ALB Load Balancer's listener's name doesnt't match. %s != %s", value, albName)
							}
							return nil
						},
					),
					testExistsFirstElementWithAttr(
						albLoadBalancerDataSourceResource, "listener", "endpoint.0.ports.0", &listenerPath,
					),
					testCheckResourceSubAttrFn(
						albLoadBalancerDataSourceResource, &listenerPath, "endpoint.0.ports.0", func(value string) error {
							port := alb.GetListeners()[0].GetEndpoints()[0].GetPorts()[0]
							if realValue, _ := strconv.ParseInt(value, 10, 64); realValue != port {
								return fmt.Errorf("ALB Load Balancer's listener's endpoint's port doesnt't match. %d != %d", realValue, port)
							}
							return nil
						},
					),
					testExistsFirstElementWithAttr(
						albLoadBalancerDataSourceResource, "listener", "http.0.handler.0.http2_options.0.max_concurrent_streams", &listenerPath,
					),
					testCheckResourceSubAttrFn(
						albLoadBalancerDataSourceResource, &listenerPath, "http.0.handler.0.http2_options.0.max_concurrent_streams", func(value string) error {
							streams := alb.GetListeners()[0].GetHttp().GetHandler().GetHttp2Options().GetMaxConcurrentStreams()
							if realValue, _ := strconv.ParseInt(value, 10, 64); realValue != streams {
								return fmt.Errorf("ALB Load Balancer's HTTP listener's max concurrent streams doesnt't match. %d != %d", realValue, streams)
							}
							return nil
						},
					),
				),
			},
		},
	})
}

func TestAccDataSourceALBLoadBalancer_httpListenerWithAllowHTTP10(t *testing.T) {
	t.Parallel()

	albResource := albLoadBalancerInfo()
	albResource.IsHTTPListener = true
	albResource.IsHTTPHandler = true
	albResource.IsAllowHTTP10 = true
	albResource.IsDataSource = true

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
					testAccCheckALBLoadBalancerExists(albLoadBalancerDataSourceResource, &alb),
					testExistsFirstElementWithAttr(
						albLoadBalancerDataSourceResource, "listener", "name", &listenerPath,
					),
					testCheckResourceSubAttrFn(
						albLoadBalancerDataSourceResource, &listenerPath, "name", func(value string) error {
							albName := alb.GetListeners()[0].GetName()
							if value != albName {
								return fmt.Errorf("ALB Load Balancer's listener's name doesnt't match. %s != %s", value, albName)
							}
							return nil
						},
					),
					testExistsFirstElementWithAttr(
						albLoadBalancerDataSourceResource, "listener", "endpoint.0.ports.0", &listenerPath,
					),
					testCheckResourceSubAttrFn(
						albLoadBalancerDataSourceResource, &listenerPath, "endpoint.0.ports.0", func(value string) error {
							port := alb.GetListeners()[0].GetEndpoints()[0].GetPorts()[0]
							if realValue, _ := strconv.ParseInt(value, 10, 64); realValue != port {
								return fmt.Errorf("ALB Load Balancer's listener's endpoint's port doesnt't match. %d != %d", realValue, port)
							}
							return nil
						},
					),
					testExistsFirstElementWithAttr(
						albLoadBalancerDataSourceResource, "listener", "http.0.handler.0.allow_http10", &listenerPath,
					),
					testCheckResourceSubAttrFn(
						albLoadBalancerDataSourceResource, &listenerPath, "http.0.handler.0.allow_http10", func(value string) error {
							allowHTTP1 := alb.GetListeners()[0].GetHttp().GetHandler().GetAllowHttp10()
							if realValue, _ := strconv.ParseBool(value); realValue != allowHTTP1 {
								return fmt.Errorf("ALB Load Balancer's HTTP listener's allow HTTP 1.0 doesnt't match. %t != %t", realValue, allowHTTP1)
							}
							return nil
						},
					),
				),
			},
		},
	})
}

func TestAccDataSourceALBLoadBalancer_httpListenerWithRedirects(t *testing.T) {
	t.Parallel()

	albResource := albLoadBalancerInfo()
	albResource.IsHTTPListener = true
	albResource.IsRedirects = true
	albResource.IsDataSource = true

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
					testAccCheckALBLoadBalancerExists(albLoadBalancerDataSourceResource, &alb),
					testExistsFirstElementWithAttr(
						albLoadBalancerDataSourceResource, "listener", "name", &listenerPath,
					),
					testCheckResourceSubAttrFn(
						albLoadBalancerDataSourceResource, &listenerPath, "name", func(value string) error {
							albName := alb.GetListeners()[0].GetName()
							if value != albName {
								return fmt.Errorf("ALB Load Balancer's listener's name doesnt't match. %s != %s", value, albName)
							}
							return nil
						},
					),
					testExistsFirstElementWithAttr(
						albLoadBalancerDataSourceResource, "listener", "endpoint.0.ports.0", &listenerPath,
					),
					testCheckResourceSubAttrFn(
						albLoadBalancerDataSourceResource, &listenerPath, "endpoint.0.ports.0", func(value string) error {
							port := alb.GetListeners()[0].GetEndpoints()[0].GetPorts()[0]
							if realValue, _ := strconv.ParseInt(value, 10, 64); realValue != port {
								return fmt.Errorf("ALB Load Balancer's listener's endpoint's port doesnt't match. %d != %d", realValue, port)
							}
							return nil
						},
					),
					testExistsElementWithAttrValue(
						albLoadBalancerDataSourceResource, "listener", "http.0.redirects.0.http_to_https", albDefaultHTTPToHTTPS, &listenerPath,
					),
				),
			},
		},
	})
}

func TestAccDataSourceALBLoadBalancer_tlsListenerWithStreamHandler(t *testing.T) {
	t.Parallel()

	albResource := albLoadBalancerInfo()
	albResource.IsTLSListener = true
	albResource.IsStreamHandler = true
	albResource.IsDataSource = true

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
					testAccDataSourceALBLoadBalancerExists(albLoadBalancerDataSourceResource, &alb),
					testExistsFirstElementWithAttr(
						albLoadBalancerDataSourceResource, "listener", "name", &listenerPath,
					),
					testCheckResourceSubAttrFn(
						albLoadBalancerDataSourceResource, &listenerPath, "name", func(value string) error {
							albName := alb.GetListeners()[0].GetName()
							if value != albName {
								return fmt.Errorf("ALB Load Balancer's listener's name doesnt't match. %s != %s", value, albName)
							}
							return nil
						},
					),
					testExistsFirstElementWithAttr(
						albLoadBalancerDataSourceResource, "listener", "endpoint.0.ports.0", &listenerPath,
					),
					testCheckResourceSubAttrFn(
						albLoadBalancerDataSourceResource, &listenerPath, "endpoint.0.ports.0", func(value string) error {
							port := alb.GetListeners()[0].GetEndpoints()[0].GetPorts()[0]
							if realValue, _ := strconv.ParseInt(value, 10, 64); realValue != port {
								return fmt.Errorf("ALB Load Balancer's listener's endpoint's port doesnt't match. %d != %d", realValue, port)
							}
							return nil
						},
					),
					testExistsFirstElementWithAttr(
						albLoadBalancerDataSourceResource, "listener", "tls.0.default_handler.0.stream_handler.0.backend_group", &listenerPath,
					),
				),
			},
		},
	})
}

func TestAccDataSourceALBLoadBalancer_tlsListenerWithHTTP2Options(t *testing.T) {
	t.Parallel()

	albResource := albLoadBalancerInfo()
	albResource.IsTLSListener = true
	albResource.IsHTTPHandler = true
	albResource.IsHTTP2Options = true
	albResource.IsDataSource = true

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
					testAccCheckALBLoadBalancerExists(albLoadBalancerDataSourceResource, &alb),
					testExistsFirstElementWithAttr(
						albLoadBalancerDataSourceResource, "listener", "name", &listenerPath,
					),
					testCheckResourceSubAttrFn(
						albLoadBalancerDataSourceResource, &listenerPath, "name", func(value string) error {
							albName := alb.GetListeners()[0].GetName()
							if value != albName {
								return fmt.Errorf("ALB Load Balancer's listener's name doesnt't match. %s != %s", value, albName)
							}
							return nil
						},
					),
					testExistsFirstElementWithAttr(
						albLoadBalancerDataSourceResource, "listener", "endpoint.0.ports.0", &listenerPath,
					),
					testCheckResourceSubAttrFn(
						albLoadBalancerDataSourceResource, &listenerPath, "endpoint.0.ports.0", func(value string) error {
							port := alb.GetListeners()[0].GetEndpoints()[0].GetPorts()[0]
							if realValue, _ := strconv.ParseInt(value, 10, 64); realValue != port {
								return fmt.Errorf("ALB Load Balancer's listener's endpoint's port doesnt't match. %d != %d", realValue, port)
							}
							return nil
						},
					),
					testExistsFirstElementWithAttr(
						albLoadBalancerDataSourceResource, "listener", "tls.0.default_handler.0.http_handler.0.http2_options.0.max_concurrent_streams", &listenerPath,
					),
					testCheckResourceSubAttrFn(
						albLoadBalancerDataSourceResource, &listenerPath, "tls.0.default_handler.0.http_handler.0.http2_options.0.max_concurrent_streams", func(value string) error {
							streams := alb.GetListeners()[0].GetTls().GetDefaultHandler().GetHttpHandler().GetHttp2Options().GetMaxConcurrentStreams()
							if realValue, _ := strconv.ParseInt(value, 10, 64); realValue != streams {
								return fmt.Errorf("ALB Load Balancer's TLS listener's max concurrent streams doesnt't match. %d != %d", realValue, streams)
							}
							return nil
						},
					),
				),
			},
		},
	})
}

func TestAccDataSourceALBLoadBalancer_tlsListenerWithAllowHTTP10(t *testing.T) {
	t.Parallel()

	albResource := albLoadBalancerInfo()
	albResource.IsTLSListener = true
	albResource.IsHTTPHandler = true
	albResource.IsAllowHTTP10 = true
	albResource.IsDataSource = true

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
					testAccCheckALBLoadBalancerExists(albLoadBalancerDataSourceResource, &alb),
					testExistsFirstElementWithAttr(
						albLoadBalancerDataSourceResource, "listener", "name", &listenerPath,
					),
					testCheckResourceSubAttrFn(
						albLoadBalancerDataSourceResource, &listenerPath, "name", func(value string) error {
							albName := alb.GetListeners()[0].GetName()
							if value != albName {
								return fmt.Errorf("ALB Load Balancer's listener's name doesnt't match. %s != %s", value, albName)
							}
							return nil
						},
					),
					testExistsFirstElementWithAttr(
						albLoadBalancerDataSourceResource, "listener", "endpoint.0.ports.0", &listenerPath,
					),
					testCheckResourceSubAttrFn(
						albLoadBalancerDataSourceResource, &listenerPath, "endpoint.0.ports.0", func(value string) error {
							port := alb.GetListeners()[0].GetEndpoints()[0].GetPorts()[0]
							if realValue, _ := strconv.ParseInt(value, 10, 64); realValue != port {
								return fmt.Errorf("ALB Load Balancer's listener's endpoint's port doesnt't match. %d != %d", realValue, port)
							}
							return nil
						},
					),
					testExistsFirstElementWithAttr(
						albLoadBalancerDataSourceResource, "listener", "tls.0.default_handler.0.http_handler.0.allow_http10", &listenerPath,
					),
					testCheckResourceSubAttrFn(
						albLoadBalancerDataSourceResource, &listenerPath, "tls.0.default_handler.0.http_handler.0.allow_http10", func(value string) error {
							allowHTTP1 := alb.GetListeners()[0].GetTls().GetDefaultHandler().GetHttpHandler().GetAllowHttp10()
							if realValue, _ := strconv.ParseBool(value); realValue != allowHTTP1 {
								return fmt.Errorf("ALB Load Balancer's TLS listener's allow HTTP 1.0 doesnt't match. %t != %t", realValue, allowHTTP1)
							}
							return nil
						},
					),
				),
			},
		},
	})
}

func testAccDataSourceALBLoadBalancerExists(n string, loadBalancer *apploadbalancer.LoadBalancer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ds, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if ds.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.ApplicationLoadBalancer().LoadBalancer().Get(context.Background(), &apploadbalancer.GetLoadBalancerRequest{
			LoadBalancerId: ds.Primary.ID,
		})

		if err != nil {
			return err
		}

		if found.Id != ds.Primary.ID {
			return fmt.Errorf("ALB Load Balancer not found")
		}

		*loadBalancer = *found

		return nil
	}
}

func testAccDataSourceALBLoadBalancerConfigByID(name, desc string) string {
	return fmt.Sprintf(`
data "yandex_alb_load_balancer" "test-alb-ds" {
  load_balancer_id = "${yandex_alb_load_balancer.test-alb.id}"
}

resource "yandex_alb_load_balancer" "test-alb" {
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

func testAccDataSourceALBLoadBalancerConfigByName(name, desc string) string {
	return fmt.Sprintf(`
data "yandex_alb_load_balancer" "test-alb-ds" {
  name = "${yandex_alb_load_balancer.test-alb.name}"
}

resource "yandex_alb_load_balancer" "test-alb" {
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
