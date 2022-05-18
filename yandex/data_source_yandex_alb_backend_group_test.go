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

const albBgDataSourceResource = "data.yandex_alb_backend_group.test-bg-ds"

func TestAccDataSourceALBBackendGroup_byID(t *testing.T) {
	t.Parallel()

	bgName := acctest.RandomWithPrefix("tf-bg")
	bgDesc := "tf-bg-description"
	folderID := getExampleFolderID()

	var bg apploadbalancer.BackendGroup

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBBackendGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceALBBackendGroupConfigByID(bgName, bgDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceALBBackendGroupExists(albBgDataSourceResource, &bg),
					testAccCheckResourceIDField(albBgDataSourceResource, "backend_group_id"),
					resource.TestCheckResourceAttr(albBgDataSourceResource, "name", bgName),
					resource.TestCheckResourceAttr(albBgDataSourceResource, "description", bgDesc),
					resource.TestCheckResourceAttr(albBgDataSourceResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(albBgDataSourceResource, "target.#", "0"),
					testAccCheckCreatedAtAttr(albBgDataSourceResource),
					testAccCheckALBBackendGroupValues(&bg, false, false, false),
				),
			},
		},
	})
}

func TestAccDataSourceALBBackendGroup_byName(t *testing.T) {
	t.Parallel()

	bgName := acctest.RandomWithPrefix("tf-bg")
	bgDesc := "tf-bg-description"
	folderID := getExampleFolderID()

	var bg apploadbalancer.BackendGroup

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBBackendGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceALBBackendGroupConfigByName(bgName, bgDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceALBBackendGroupExists(albBgDataSourceResource, &bg),
					testAccCheckResourceIDField(albBgDataSourceResource, "backend_group_id"),
					resource.TestCheckResourceAttr(albBgDataSourceResource, "name", bgName),
					resource.TestCheckResourceAttr(albBgDataSourceResource, "description", bgDesc),
					resource.TestCheckResourceAttr(albBgDataSourceResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(albBgDataSourceResource, "target.#", "0"),
					testAccCheckCreatedAtAttr(albBgDataSourceResource),
					testAccCheckALBBackendGroupValues(&bg, false, false, false),
				),
			},
		},
	})
}

func TestAccDataSourceALBBackendGroup_fullWithHTTPBackend(t *testing.T) {
	t.Parallel()

	BGResource := albBackendGroupInfo()
	BGResource.IsDataSource = true
	BGResource.IsHTTPBackend = true

	backendPath := ""
	var bg apploadbalancer.BackendGroup

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBBackendGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testALBBackendGroupConfig_basic(BGResource),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceALBBackendGroupExists(albBgDataSourceResource, &bg),
					testAccCheckALBBackendGroupValues(&bg, true, false, false),
					testExistsFirstElementWithAttr(
						albBgDataSourceResource, "http_backend", "tls", &backendPath,
					),
					testCheckResourceSubAttrFn(
						albBgDataSourceResource, &backendPath, "tls.0.sni", func(value string) error {
							tlsSni := bg.GetHttp().GetBackends()[0].Tls.Sni
							if value != tlsSni {
								return fmt.Errorf("BackendGroup's http backend's tls sni doesnt't match. %s != %s", value, tlsSni)
							}
							return nil
						},
					),
					testCheckResourceSubAttrFn(
						albBgDataSourceResource, &backendPath, "load_balancing_config.0.locality_aware_routing_percent", func(value string) error {
							lbConfigPercent := bg.GetHttp().GetBackends()[0].LoadBalancingConfig.LocalityAwareRoutingPercent
							if value != strconv.FormatInt(lbConfigPercent, 10) {
								return fmt.Errorf("BackendGroup's http backend's load balancing config locality aware routing percent doesnt't match. %s != %d", value, lbConfigPercent)
							}
							return nil
						},
					),
					testCheckResourceSubAttrFn(
						albBgDataSourceResource, &backendPath, "load_balancing_config.0.panic_threshold", func(value string) error {
							lbConfigPanicThreshold := bg.GetHttp().GetBackends()[0].LoadBalancingConfig.PanicThreshold
							if value != strconv.FormatInt(lbConfigPanicThreshold, 10) {
								return fmt.Errorf("BackendGroup's http backend's load balancing config panic threshold doesnt't match. %s != %d", value, lbConfigPanicThreshold)
							}
							return nil
						},
					),
					testCheckResourceSubAttrFn(
						albBgDataSourceResource, &backendPath, "load_balancing_config.0.strict_locality", func(value string) error {
							lbConfigStrictLocality := bg.GetHttp().GetBackends()[0].LoadBalancingConfig.StrictLocality
							if value != strconv.FormatBool(lbConfigStrictLocality) {
								return fmt.Errorf("BackendGroup's http backend's load balancing config panic threshold doesnt't match. %s != %t", value, lbConfigStrictLocality)
							}
							return nil
						},
					),
				),
			},
		},
	})
}

func TestAccDataSourceALBBackendGroup_fullWithHTTPBackendForStorageBucket(t *testing.T) {
	t.Parallel()

	BGResource := albBackendGroupInfo()
	BGResource.IsDataSource = true
	BGResource.IsHTTPBackend = true
	BGResource.IsStorageBackend = true
	BGResource.StorageBackendBucket = "test-tf-bucket"

	backendPath := ""
	var bg apploadbalancer.BackendGroup

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBBackendGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testALBBackendGroupConfig_basic(BGResource),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceALBBackendGroupExists(albBgDataSourceResource, &bg),
					testAccCheckALBBackendGroupValues(&bg, true, false, false),
					testExistsFirstElementWithAttr(
						albBgDataSourceResource, "http_backend", "tls", &backendPath,
					),
					testCheckResourceSubAttrFn(
						albBgDataSourceResource, &backendPath, "tls.0.sni", func(value string) error {
							tlsSni := bg.GetHttp().GetBackends()[0].Tls.Sni
							if value != tlsSni {
								return fmt.Errorf("BackendGroup's http backend's tls sni doesnt't match. %s != %s", value, tlsSni)
							}
							return nil
						},
					),
					testCheckResourceSubAttrFn(
						albBgDataSourceResource, &backendPath, "storage_bucket", func(value string) error {
							bucket := bg.GetHttp().GetBackends()[0].GetStorageBucket().GetBucket()
							if value != bucket {
								return fmt.Errorf("BackendGroup's http backend's storage bucket doesnt't match. %s != %s", value, bucket)
							}
							return nil
						},
					),
				),
			},
		},
	})
}

func TestAccDataSourceALBBackendGroup_withSessionAffinityHeader(t *testing.T) {
	t.Parallel()

	BGResource := albBackendGroupInfo()
	BGResource.IsDataSource = true
	BGResource.IsHTTPBackend = true
	BGResource.UseHeaderAffinity = true

	affinityPath := "session_affinity"
	var bg apploadbalancer.BackendGroup

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBBackendGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testALBBackendGroupConfig_basic(BGResource),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceALBBackendGroupExists(albBgDataSourceResource, &bg),
					testAccCheckALBBackendGroupValues(&bg, true, false, false),
					testCheckResourceSubAttrFn(
						albBgDataSourceResource, &affinityPath, "0.header.0.header_name", func(value string) error {
							hdrName := bg.GetHttp().GetHeader().GetHeaderName()
							if value != hdrName {
								return fmt.Errorf("BackendGroup's http backend's header affinity doesnt't match. %s != %s", value, hdrName)
							}
							return nil
						},
					),
				),
			},
		},
	})
}

func TestAccDataSourceALBBackendGroup_fullWithStreamBackend(t *testing.T) {
	t.Parallel()

	BGResource := albBackendGroupInfo()
	BGResource.IsDataSource = true
	BGResource.IsStreamBackend = true

	backendPath := ""
	var bg apploadbalancer.BackendGroup

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBBackendGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testALBBackendGroupConfig_basic(BGResource),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceALBBackendGroupExists(albBgDataSourceResource, &bg),
					testAccCheckALBBackendGroupValues(&bg, false, false, true),
					testExistsFirstElementWithAttr(
						albBgDataSourceResource, "stream_backend", "tls", &backendPath,
					),
					testCheckResourceSubAttrFn(
						albBgDataSourceResource, &backendPath, "tls.0.sni", func(value string) error {
							tlsSni := bg.GetStream().GetBackends()[0].Tls.Sni
							if value != tlsSni {
								return fmt.Errorf("BackendGroup's Stream backend's tls sni doesnt't match. %s != %s", value, tlsSni)
							}
							return nil
						},
					),
					testCheckResourceSubAttrFn(
						albBgDataSourceResource, &backendPath, "load_balancing_config.0.locality_aware_routing_percent", func(value string) error {
							lbConfigPercent := bg.GetStream().GetBackends()[0].LoadBalancingConfig.LocalityAwareRoutingPercent
							if value != strconv.FormatInt(lbConfigPercent, 10) {
								return fmt.Errorf("BackendGroup's Stream backend's load balancing config locality aware routing percent doesnt't match. %s != %d", value, lbConfigPercent)
							}
							return nil
						},
					),
					testCheckResourceSubAttrFn(
						albBgDataSourceResource, &backendPath, "load_balancing_config.0.panic_threshold", func(value string) error {
							lbConfigPanicThreshold := bg.GetStream().GetBackends()[0].LoadBalancingConfig.PanicThreshold
							if value != strconv.FormatInt(lbConfigPanicThreshold, 10) {
								return fmt.Errorf("BackendGroup's Stream backend's load balancing config panic threshold doesnt't match. %s != %d", value, lbConfigPanicThreshold)
							}
							return nil
						},
					),
					testCheckResourceSubAttrFn(
						albBgDataSourceResource, &backendPath, "load_balancing_config.0.strict_locality", func(value string) error {
							lbConfigStrictLocality := bg.GetStream().GetBackends()[0].LoadBalancingConfig.StrictLocality
							if value != strconv.FormatBool(lbConfigStrictLocality) {
								return fmt.Errorf("BackendGroup's Stream backend's load balancing config panic threshold doesnt't match. %s != %t", value, lbConfigStrictLocality)
							}
							return nil
						},
					),
				),
			},
		},
	})
}

func TestAccDataSourceALBBackendGroup_fullWithGrpcBackend(t *testing.T) {
	t.Parallel()

	BGResource := albBackendGroupInfo()
	BGResource.IsDataSource = true
	BGResource.IsGRPCBackend = true

	backendPath := ""
	var bg apploadbalancer.BackendGroup

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckALBBackendGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testALBBackendGroupConfig_basic(BGResource),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceALBBackendGroupExists(albBgDataSourceResource, &bg),
					testAccCheckALBBackendGroupValues(&bg, false, true, false),
					testExistsFirstElementWithAttr(
						albBgDataSourceResource, "grpc_backend", "tls", &backendPath,
					),
					testCheckResourceSubAttrFn(
						albBgDataSourceResource, &backendPath, "tls.0.sni", func(value string) error {
							tlsSni := bg.GetGrpc().GetBackends()[0].Tls.Sni
							if value != tlsSni {
								return fmt.Errorf("BackendGroup's grpc backend's tls sni doesnt't match. %s != %s", value, tlsSni)
							}
							return nil
						},
					),
					testCheckResourceSubAttrFn(
						albBgDataSourceResource, &backendPath, "load_balancing_config.0.locality_aware_routing_percent", func(value string) error {
							lbConfigPercent := bg.GetGrpc().GetBackends()[0].LoadBalancingConfig.LocalityAwareRoutingPercent
							if value != strconv.FormatInt(lbConfigPercent, 10) {
								return fmt.Errorf("BackendGroup's grpc backend's load balancing config locality aware routing percent doesnt't match. %s != %d", value, lbConfigPercent)
							}
							return nil
						},
					),
					testCheckResourceSubAttrFn(
						albBgDataSourceResource, &backendPath, "load_balancing_config.0.panic_threshold", func(value string) error {
							lbConfigPanicThreshold := bg.GetGrpc().GetBackends()[0].LoadBalancingConfig.PanicThreshold
							if value != strconv.FormatInt(lbConfigPanicThreshold, 10) {
								return fmt.Errorf("BackendGroup's Grpc backend's load balancing config panic threshold doesnt't match. %s != %d", value, lbConfigPanicThreshold)
							}
							return nil
						},
					),
					testCheckResourceSubAttrFn(
						albBgDataSourceResource, &backendPath, "load_balancing_config.0.strict_locality", func(value string) error {
							lbConfigStrictLocality := bg.GetGrpc().GetBackends()[0].LoadBalancingConfig.StrictLocality
							if value != strconv.FormatBool(lbConfigStrictLocality) {
								return fmt.Errorf("BackendGroup's grpc backend's load balancing config panic threshold doesnt't match. %s != %t", value, lbConfigStrictLocality)
							}
							return nil
						},
					),
				),
			},
		},
	})
}

func testAccDataSourceALBBackendGroupExists(bgName string, bg *apploadbalancer.BackendGroup) resource.TestCheckFunc {
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

func testAccDataSourceALBBackendGroupConfigByID(name, desc string) string {
	return fmt.Sprintf(`
data "yandex_alb_backend_group" "test-bg-ds" {
  backend_group_id = "${yandex_alb_backend_group.test-bg.id}"
}

resource "yandex_alb_backend_group" "test-bg" {
  name			= "%s"
  description	= "%s"
}
`, name, desc)
}

func testAccDataSourceALBBackendGroupConfigByName(name, desc string) string {
	return fmt.Sprintf(`
data "yandex_alb_backend_group" "test-bg-ds" {
  name = "${yandex_alb_backend_group.test-bg.name}"
}

resource "yandex_alb_backend_group" "test-bg" {
  name			= "%s"
  description	= "%s"
}
`, name, desc)
}
