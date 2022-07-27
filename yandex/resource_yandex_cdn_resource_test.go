package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-multierror"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/cdn/v1"
)

func init() {
	resource.AddTestSweepers("yandex_cdn_resource", &resource.Sweeper{
		Name: "yandex_cdn_resource",
		F:    testSweepCDNResource,
	})
}

func TestAccCDNResource_basicByGroupID(t *testing.T) {
	t.Parallel()

	folderID := getExampleFolderID()

	groupName := fmt.Sprintf("tf-test-cdn-resource-%s", acctest.RandString(10))
	resourceCName := fmt.Sprintf("cdn-tf-test-%s.yandex.net", acctest.RandString(4))

	var cdnResource cdn.Resource

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCDNResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCDNResource_basicByID(groupName, resourceCName),
				Check: resource.ComposeTestCheckFunc(
					testCDNResourceExists("yandex_cdn_resource.foobar_resource", &cdnResource),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "cname", resourceCName),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "folder_id", folderID),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "origin_protocol", "http"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "active", "true"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "secondary_hostnames.#", "0"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "ssl_certificate.0.type", "not_used"),
					testAccCheckCreatedAtAttr("yandex_cdn_resource.foobar_resource"),
				),
			},
			{
				ResourceName: "yandex_cdn_resource.foobar_resource",
				ImportStateIdFunc: func(*terraform.State) (string, error) {
					return cdnResource.Id, nil
				},
				ImportState:             true,
				ImportStateVerifyIgnore: []string{"origin_group_id", "origin_group_name"},
				ImportStateVerify:       true,
			},
		},
	})
}

func TestAccCDNResource_basicByName(t *testing.T) {
	t.Parallel()

	folderID := getExampleFolderID()

	groupName := fmt.Sprintf("tf-test-cdn-resource-%s", acctest.RandString(10))
	resourceCName := fmt.Sprintf("cdn-tf-test-%s.yandex.net", acctest.RandString(4))

	var cdnResource cdn.Resource

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCDNResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCDNResource_basicByName(groupName, resourceCName),
				Check: resource.ComposeTestCheckFunc(
					testCDNResourceExists("yandex_cdn_resource.foobar_resource", &cdnResource),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "cname", resourceCName),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "folder_id", folderID),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "origin_protocol", "http"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "active", "true"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "secondary_hostnames.#", "0"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "ssl_certificate.0.type", "not_used"),
					testAccCheckCreatedAtAttr("yandex_cdn_resource.foobar_resource"),
				),
			},
		},
	})
}

func TestAccCDNResource_basicByNameWithFolderID(t *testing.T) {
	t.Parallel()

	folderID := getExampleFolderID()

	groupName := fmt.Sprintf("tf-test-cdn-resource-%s", acctest.RandString(10))
	resourceCName := fmt.Sprintf("cdn-tf-test-%s.yandex.net", acctest.RandString(4))

	var cdnResource cdn.Resource

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCDNResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCDNResource_basicByNameWithFolderID(groupName, resourceCName, folderID),
				Check: resource.ComposeTestCheckFunc(
					testCDNResourceExists("yandex_cdn_resource.foobar_resource", &cdnResource),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "cname", resourceCName),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "folder_id", folderID),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "origin_protocol", "http"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "active", "true"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "secondary_hostnames.#", "0"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "ssl_certificate.0.type", "not_used"),
					testAccCheckCreatedAtAttr("yandex_cdn_resource.foobar_resource"),
				),
			},
		},
	})
}

func TestAccCDNResource_basicUpdate(t *testing.T) {
	folderID := getExampleFolderID()

	groupName := fmt.Sprintf("tf-test-cdn-resource-%s", acctest.RandString(10))
	resourceCName := fmt.Sprintf("cdn-tf-test-%s.yandex.net", acctest.RandString(4))

	var cdnResource, cdnResourceUpdated cdn.Resource

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCDNResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCDNResource_basicByName(groupName, resourceCName),
				Check: resource.ComposeTestCheckFunc(
					testCDNResourceExists("yandex_cdn_resource.foobar_resource", &cdnResource),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "cname", resourceCName),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "origin_group_name", groupName),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "folder_id", folderID),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "origin_protocol", "http"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "active", "true"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "secondary_hostnames.#", "0"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "ssl_certificate.0.type", "not_used"),
					testAccCheckCreatedAtAttr("yandex_cdn_resource.foobar_resource"),
				),
			},
			{
				Config: testAccCDNResource_basicUpdate(groupName, resourceCName),
				Check: resource.ComposeTestCheckFunc(
					testCDNResourceExists("yandex_cdn_resource.foobar_resource", &cdnResourceUpdated),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "cname", resourceCName),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "folder_id", folderID),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "origin_group_name", groupName),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "origin_protocol", "https"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "active", "false"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "secondary_hostnames.#", "2"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "secondary_hostnames.0", "cdn-test-3.yandex.ru"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "secondary_hostnames.1", "cdn-test-4.yandex.ru"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "ssl_certificate.0.type", "not_used"),
					testAccCheckCreatedAtAttr("yandex_cdn_resource.foobar_resource"),
				),
			},
		},
	})
}

func TestAccCDNResource_updateGroupById(t *testing.T) {
	t.Parallel()

	folderID := getExampleFolderID()

	groupName := fmt.Sprintf("tf-test-cdn-group-%s", acctest.RandString(10))
	updatedGroupName := fmt.Sprintf("tf-test-cdn-group-updated-%s", acctest.RandString(10))

	resourceCName := fmt.Sprintf("cdn-tf-test-%s.yandex.net", acctest.RandString(4))

	var cdnResource, cdnResourceUpdated cdn.Resource

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCDNResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCDNResource_basicByName(groupName, resourceCName),
				Check: resource.ComposeTestCheckFunc(
					testCDNResourceExists("yandex_cdn_resource.foobar_resource", &cdnResource),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "cname", resourceCName),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "folder_id", folderID),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "origin_group_name", groupName),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "origin_protocol", "http"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "active", "true"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "secondary_hostnames.#", "0"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "ssl_certificate.0.type", "not_used"),
					testAccCheckCreatedAtAttr("yandex_cdn_resource.foobar_resource"),
				),
			},
			{
				Config: testAccCDNResource_updateOriginGroupByID(groupName, updatedGroupName, resourceCName),
				Check: resource.ComposeTestCheckFunc(
					testCDNResourceExists("yandex_cdn_resource.foobar_resource", &cdnResourceUpdated),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "cname", resourceCName),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "folder_id", folderID),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "origin_protocol", "https"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "active", "false"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "secondary_hostnames.#", "2"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "secondary_hostnames.0", "cdn-test-1.yandex.ru"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "secondary_hostnames.1", "cdn-test-2.yandex.ru"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "ssl_certificate.0.type", "not_used"),
					testAccCheckCreatedAtAttr("yandex_cdn_resource.foobar_resource"),
				),
			},
		},
	})
}

func TestAccCDNResource_updateGroupByName(t *testing.T) {
	t.Parallel()

	folderID := getExampleFolderID()

	groupName := fmt.Sprintf("tf-test-cdn-group-%s", acctest.RandString(10))
	updatedGroupName := fmt.Sprintf("tf-test-cdn-group-updated-%s", acctest.RandString(10))

	resourceCName := fmt.Sprintf("cdn-tf-test-%s.yandex.net", acctest.RandString(4))

	var cdnResource, cdnResourceUpdated cdn.Resource

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCDNResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCDNResource_basicByName(groupName, resourceCName),
				Check: resource.ComposeTestCheckFunc(
					testCDNResourceExists("yandex_cdn_resource.foobar_resource", &cdnResource),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "cname", resourceCName),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "folder_id", folderID),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "origin_group_name", groupName),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "origin_protocol", "http"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "active", "true"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "secondary_hostnames.#", "0"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "ssl_certificate.0.type", "not_used"),
					testAccCheckCreatedAtAttr("yandex_cdn_resource.foobar_resource"),
				),
			},
			{
				Config: testAccCDNResource_updateOriginGroupByName(groupName, updatedGroupName, resourceCName),
				Check: resource.ComposeTestCheckFunc(
					testCDNResourceExists("yandex_cdn_resource.foobar_resource", &cdnResourceUpdated),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "cname", resourceCName),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "folder_id", folderID),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "origin_group_name", updatedGroupName),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "origin_protocol", "https"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "active", "false"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "secondary_hostnames.#", "2"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "secondary_hostnames.0", "cdn-test-5.yandex.ru"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "secondary_hostnames.1", "cdn-test-6.yandex.ru"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "ssl_certificate.0.type", "not_used"),
					testAccCheckCreatedAtAttr("yandex_cdn_resource.foobar_resource"),
				),
			},
		},
	})
}

func TestAccCDNResource_optionDisableCacheOn(t *testing.T) {
	folderID := getExampleFolderID()

	groupName := fmt.Sprintf("tf-test-cdn-resource-%s", acctest.RandString(10))
	resourceCName := fmt.Sprintf("cdn-tf-test-%s.yandex.net", acctest.RandString(4))

	var cdnResource cdn.Resource

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCDNResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCDNResource_optionDisableCache(groupName, resourceCName, true),
				Check: resource.ComposeTestCheckFunc(
					testCDNResourceExists("yandex_cdn_resource.foobar_resource", &cdnResource),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "cname", resourceCName),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "folder_id", folderID),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "options.#", "1"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "options.0.disable_cache", "true"),
					testAccCheckCreatedAtAttr("yandex_cdn_resource.foobar_resource"),
				),
			},
		},
	})
}

func TestAccCDNResource_optionDisableCacheOff(t *testing.T) {
	folderID := getExampleFolderID()

	groupName := fmt.Sprintf("tf-test-cdn-resource-%s", acctest.RandString(10))
	resourceCName := fmt.Sprintf("cdn-tf-test-%s.yandex.net", acctest.RandString(4))

	var cdnResource cdn.Resource

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCDNResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCDNResource_optionDisableCache(groupName, resourceCName, false),
				Check: resource.ComposeTestCheckFunc(
					testCDNResourceExists("yandex_cdn_resource.foobar_resource", &cdnResource),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "cname", resourceCName),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "folder_id", folderID),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "options.#", "1"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "options.0.disable_cache", "false"),
					testAccCheckCreatedAtAttr("yandex_cdn_resource.foobar_resource"),
				),
			},
		},
	})
}

func TestAccCDNResource_optionEdgeCacheSettings(t *testing.T) {
	folderID := getExampleFolderID()

	groupName := fmt.Sprintf("tf-test-cdn-resource-%s", acctest.RandString(10))
	resourceCName := fmt.Sprintf("cdn-tf-test-%s.yandex.net", acctest.RandString(4))
	ttl := acctest.RandIntRange(10, 100500)

	var cdnResource cdn.Resource

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCDNResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCDNResource_optionEdgeCacheSetting(groupName, resourceCName, ttl),
				Check: resource.ComposeTestCheckFunc(
					testCDNResourceExists("yandex_cdn_resource.foobar_resource", &cdnResource),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "cname", resourceCName),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "folder_id", folderID),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "options.#", "1"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "options.0.edge_cache_settings", fmt.Sprintf("%d", ttl)),
					testAccCheckCreatedAtAttr("yandex_cdn_resource.foobar_resource"),
				),
			},
		},
	})
}

func TestAccCDNResource_optionIgnoreQueryParams(t *testing.T) {
	folderID := getExampleFolderID()

	groupName := fmt.Sprintf("tf-test-cdn-resource-%s", acctest.RandString(10))
	resourceCName := fmt.Sprintf("cdn-tf-test-%s.yandex.net", acctest.RandString(4))

	var cdnResource cdn.Resource

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCDNResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCDNResource_optionIgnoreQueryParams(groupName, resourceCName),
				Check: resource.ComposeTestCheckFunc(
					testCDNResourceExists("yandex_cdn_resource.foobar_resource", &cdnResource),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "cname", resourceCName),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "folder_id", folderID),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "options.#", "1"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "options.0.ignore_query_params", "true"),
					testAccCheckCreatedAtAttr("yandex_cdn_resource.foobar_resource"),
				),
			},
		},
	})
}

func TestAccCDNResource_optionIgnoreCookie(t *testing.T) {
	folderID := getExampleFolderID()

	groupName := fmt.Sprintf("tf-test-cdn-resource-%s", acctest.RandString(10))
	resourceCName := fmt.Sprintf("cdn-tf-test-%s.yandex.net", acctest.RandString(4))

	var cdnResource cdn.Resource

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCDNResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCDNResource_optionIgnoreCookie(groupName, resourceCName),
				Check: resource.ComposeTestCheckFunc(
					testCDNResourceExists("yandex_cdn_resource.foobar_resource", &cdnResource),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "cname", resourceCName),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "folder_id", folderID),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "options.#", "1"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "options.0.ignore_cookie", "true"),
					testAccCheckCreatedAtAttr("yandex_cdn_resource.foobar_resource"),
				),
			},
		},
	})
}

func TestAccCDNResource_optionCustomHostHeader(t *testing.T) {
	folderID := getExampleFolderID()

	groupName := fmt.Sprintf("tf-test-cdn-resource-%s", acctest.RandString(10))
	resourceCName := fmt.Sprintf("cdn-tf-test-%s.yandex.net", acctest.RandString(4))
	customHostHeader := fmt.Sprintf("cdn%02d.yandex.net", acctest.RandIntRange(1, 64))

	var cdnResource cdn.Resource

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCDNResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCDNResource_optionCustomHostHeader(groupName, resourceCName, customHostHeader),
				Check: resource.ComposeTestCheckFunc(
					testCDNResourceExists("yandex_cdn_resource.foobar_resource", &cdnResource),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "cname", resourceCName),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "folder_id", folderID),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "options.#", "1"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "options.0.custom_host_header", customHostHeader),
					testAccCheckCreatedAtAttr("yandex_cdn_resource.foobar_resource"),
				),
			},
		},
	})
}

func TestAccCDNResource_optionForwardHostHeader(t *testing.T) {
	folderID := getExampleFolderID()

	groupName := fmt.Sprintf("tf-test-cdn-resource-%s", acctest.RandString(10))
	resourceCName := fmt.Sprintf("cdn-tf-test-%s.yandex.net", acctest.RandString(4))

	var cdnResource cdn.Resource

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCDNResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCDNResource_optionForwardHostHeader(groupName, resourceCName, true),
				Check: resource.ComposeTestCheckFunc(
					testCDNResourceExists("yandex_cdn_resource.foobar_resource", &cdnResource),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "cname", resourceCName),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "folder_id", folderID),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "options.#", "1"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "options.0.forward_host_header", "true"),
					testAccCheckCreatedAtAttr("yandex_cdn_resource.foobar_resource"),
				),
			},
		},
	})
}

func TestAccCDNResource_optionStaticHeaders(t *testing.T) {
	folderID := getExampleFolderID()

	groupName := fmt.Sprintf("tf-test-cdn-resource-%s", acctest.RandString(10))
	resourceCName := fmt.Sprintf("cdn-tf-test-%s.yandex.net", acctest.RandString(4))

	var cdnResource cdn.Resource

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCDNResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCDNResource_optionStaticHeaders(groupName, resourceCName),
				Check: resource.ComposeTestCheckFunc(
					testCDNResourceExists("yandex_cdn_resource.foobar_resource", &cdnResource),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "cname", resourceCName),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "folder_id", folderID),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "options.#", "1"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "options.0.static_response_headers.X-Tf-Check-1", "some test value #1"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "options.0.static_response_headers.X-Tf-Check-2", "some test value #2"),
					testAccCheckCreatedAtAttr("yandex_cdn_resource.foobar_resource"),
				),
			},
		},
	})
}

func testSweepCDNResource(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &cdn.ListResourcesRequest{
		FolderId: conf.FolderID,
	}

	it := conf.sdk.CDN().Resource().ResourceIterator(conf.Context(), req)
	result := &multierror.Error{}

	for it.Next() {
		id := it.Value().GetId()
		if !sweepCDNResource(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep CDN resource %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepCDNResource(conf *Config, id string) bool {
	return sweepWithRetryByFunc(conf, "CDN Resource", func(conf *Config) error {
		ctx, cancel := conf.ContextWithTimeout(yandexCDNOriginGroupDefaultTimeout)
		defer cancel()

		op, err := conf.sdk.CDN().Resource().Delete(ctx, &cdn.DeleteResourceRequest{
			ResourceId: id,
		})

		return handleSweepOperation(ctx, conf, op, err)
	})
}

func testAccCheckCDNResourceDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_cdn_resource" {
			continue
		}

		_, err := config.sdk.CDN().Resource().Get(context.Background(), &cdn.GetResourceRequest{
			ResourceId: rs.Primary.ID,
		})

		if err == nil {
			return fmt.Errorf("CDN Resource still exists")
		}
	}

	return nil
}

func testCDNResourceExists(resourceName string, cdnResource *cdn.Resource) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.CDN().Resource().Get(context.Background(), &cdn.GetResourceRequest{
			ResourceId: rs.Primary.ID,
		})

		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("cdn resource is not found")
		}

		//goland:noinspection GoVetCopyLock
		*cdnResource = *found

		return nil
	}
}

func testAccCDNResource_optionDisableCache(groupName, resourceCNAME string, on bool) string {
	return makeGroupResource(groupName) + fmt.Sprintf(`
resource "yandex_cdn_resource" "foobar_resource" {
	cname = "%s"

	origin_group_id = "${yandex_cdn_origin_group.foo_cdn_group.id}"

	options {
		disable_cache = %t
	}
}
`, resourceCNAME, on)
}

func testAccCDNResource_optionEdgeCacheSetting(groupName, resourceCNAME string, edgeCacheSettings int) string {
	return makeGroupResource(groupName) + fmt.Sprintf(`
resource "yandex_cdn_resource" "foobar_resource" {
	cname = "%s"

	origin_group_id = "${yandex_cdn_origin_group.foo_cdn_group.id}"

	options {
		edge_cache_settings = %d
	}
}
`, resourceCNAME, edgeCacheSettings)
}

func testAccCDNResource_optionIgnoreQueryParams(groupName, resourceCNAME string) string {
	return makeGroupResource(groupName) + fmt.Sprintf(`
resource "yandex_cdn_resource" "foobar_resource" {
	cname = "%s"

	origin_group_id = "${yandex_cdn_origin_group.foo_cdn_group.id}"

	options {
		ignore_query_params = true
	}
}
`, resourceCNAME)
}

func testAccCDNResource_optionStaticHeaders(groupName, resourceCNAME string) string {
	return makeGroupResource(groupName) + fmt.Sprintf(`
resource "yandex_cdn_resource" "foobar_resource" {
	cname = "%s"

	origin_group_id = "${yandex_cdn_origin_group.foo_cdn_group.id}"

	options {
		static_response_headers = {
			X-Tf-Check-1 = "some test value #1"
			X-Tf-Check-2 = "some test value #2"
		}
	}
}
`, resourceCNAME)
}

func testAccCDNResource_optionIgnoreCookie(groupName, resourceCNAME string) string {
	return makeGroupResource(groupName) + fmt.Sprintf(`
resource "yandex_cdn_resource" "foobar_resource" {
	cname = "%s"

	origin_group_id = "${yandex_cdn_origin_group.foo_cdn_group.id}"

	options {
		ignore_cookie = true
	}
}
`, resourceCNAME)
}

func testAccCDNResource_optionCustomHostHeader(groupName, resourceCNAME, customHostHeader string) string {
	return makeGroupResource(groupName) + fmt.Sprintf(`
resource "yandex_cdn_resource" "foobar_resource" {
	cname = "%s"

	origin_group_id = "${yandex_cdn_origin_group.foo_cdn_group.id}"

	options {
		custom_host_header = "%s"
	}
}
`, resourceCNAME, customHostHeader)
}

func testAccCDNResource_optionForwardHostHeader(groupName, resourceCNAME string, forwardHostHeader bool) string {
	return makeGroupResource(groupName) + fmt.Sprintf(`
resource "yandex_cdn_resource" "foobar_resource" {
	cname = "%s"

	origin_group_id = "${yandex_cdn_origin_group.foo_cdn_group.id}"

	options {
		forward_host_header = %t
	}
}
`, resourceCNAME, forwardHostHeader)
}

func testAccCDNResource_basicByName(groupName, resourceCNAME string) string {
	return fmt.Sprintf(`
resource "yandex_cdn_origin_group" "foo_cdn_group_by_name" {
	name     = "%s"

	origin {
		source = "ya.ru"
	}
}

resource "yandex_cdn_resource" "foobar_resource" {
	cname = "%s"

	origin_group_name = "${yandex_cdn_origin_group.foo_cdn_group_by_name.name}"
}
`, groupName, resourceCNAME)
}

func testAccCDNResource_basicByNameWithFolderID(groupName, resourceCNAME, folderID string) string {
	return fmt.Sprintf(`
resource "yandex_cdn_origin_group" "foo_cdn_group_by_name" {
  name      = "%s"
  folder_id = "%s"

  origin {
    source = "ya.ru"
  }
}

resource "yandex_cdn_resource" "foobar_resource" {
  cname = "%s"

  origin_group_name = yandex_cdn_origin_group.foo_cdn_group_by_name.name
}

`, groupName, folderID, resourceCNAME)
}

func testAccCDNResource_basicByID(groupName, resourceCNAME string) string {
	return fmt.Sprintf(`
resource "yandex_cdn_origin_group" "foo_cdn_group_by_id" {
	name     = "%s"

	origin {
		source = "ya.ru"
	}
}

resource "yandex_cdn_resource" "foobar_resource" {
	cname = "%s"

	origin_group_id = "${yandex_cdn_origin_group.foo_cdn_group_by_id.id}"
}
`, groupName, resourceCNAME)
}

func testAccCDNResource_basicUpdate(groupName, resourceCNAME string) string {
	return fmt.Sprintf(`
resource "yandex_cdn_origin_group" "foo_cdn_group_by_name" {
	name     = "%s"

	origin {
		source = "ya.ru"
	}
}

resource "yandex_cdn_resource" "foobar_resource" {
	cname = "%s"

	active = false

	origin_protocol = "https"

	secondary_hostnames = ["cdn-test-3.yandex.ru", "cdn-test-4.yandex.ru"]

	origin_group_name = "${yandex_cdn_origin_group.foo_cdn_group_by_name.name}"
}
`, groupName, resourceCNAME)
}

func testAccCDNResource_updateOriginGroupByID(originalGroupName, groupName, resourceCNAME string) string {
	return fmt.Sprintf(`
resource "yandex_cdn_origin_group" "foo_cdn_group_by_name" {
	name     = "%s"

	origin {
		source = "ya.ru"
	}
}

resource "yandex_cdn_origin_group" "update_foo_cdn_group_by_id" {
	name     = "%s"

	origin {
		source = "yandex.ru"
	}
}

resource "yandex_cdn_resource" "foobar_resource" {
	cname = "%s"

	active = false

	origin_protocol = "https"

	secondary_hostnames = ["cdn-test-1.yandex.ru", "cdn-test-2.yandex.ru"]

	origin_group_id = yandex_cdn_origin_group.update_foo_cdn_group_by_id.id
}

`, originalGroupName, groupName, resourceCNAME)
}

func testAccCDNResource_updateOriginGroupByName(originalGroupName, groupName, resourceCNAME string) string {
	return fmt.Sprintf(`
resource "yandex_cdn_origin_group" "foo_cdn_group_by_name" {
	name     = "%s"

	origin {
		source = "ya.ru"
	}
}

resource "yandex_cdn_origin_group" "update_foo_cdn_group_by_name" {
	name     = "%s"

	origin {
		source = "yandex.ru"
	}
}

resource "yandex_cdn_resource" "foobar_resource" {
	cname = "%s"

	active = false

	origin_protocol = "https"

	secondary_hostnames = ["cdn-test-5.yandex.ru", "cdn-test-6.yandex.ru"]

	origin_group_name = yandex_cdn_origin_group.update_foo_cdn_group_by_name.name
}

`, originalGroupName, groupName, resourceCNAME)
}

func makeGroupResource(groupName string) string {
	return fmt.Sprintf(`
	resource "yandex_cdn_origin_group" "foo_cdn_group" {
		name     = "%s"

		origin {
		  source = "ya.ru"
		}
	}
	`, groupName)
}
