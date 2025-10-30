package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-multierror"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

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

func TestAccCDNResource_CName_ForceNew(t *testing.T) {
	t.Parallel()

	groupName := fmt.Sprintf("tf-test-cdn-resource-%s", acctest.RandString(10))
	originalCName := fmt.Sprintf("cdn-tf-test-%s.yandex.net", acctest.RandString(4))
	newCName := fmt.Sprintf("cdn-tf-test-%s.yandex.net", acctest.RandString(4))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCDNResourceDestroy,
		Steps: []resource.TestStep{
			// Create the CDN resource with the initial CNAME.
			{
				Config: testAccCDNResource_basicByName(groupName, originalCName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "cname", originalCName),
				),
			},
			// Attempt to change the CNAME, which should result in an error.
			{
				Config: testAccCDNResource_basicByName(groupName, newCName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("yandex_cdn_resource.foobar_resource", plancheck.ResourceActionReplace),
					},
				},
			},
		},
	})
}

// TODO: secondary_hostnames

// TODO: ssl_certificate

func TestAccCDNResource_Labels(t *testing.T) {
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
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "labels.%", "0"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "origin_protocol", "http"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "active", "true"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "secondary_hostnames.#", "0"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "ssl_certificate.0.type", "not_used"),
					testAccCheckCreatedAtAttr("yandex_cdn_resource.foobar_resource"),
				),
			},
			{
				Config: testAccCDNResource_addLabel(groupName, resourceCName),
				Check: resource.ComposeTestCheckFunc(
					testCDNResourceExists("yandex_cdn_resource.foobar_resource", &cdnResourceUpdated),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "cname", resourceCName),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "origin_group_name", groupName),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "folder_id", folderID),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "labels.%", "1"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "labels.environment", "testing"),
					testAccCheckCreatedAtAttr("yandex_cdn_resource.foobar_resource"),
				),
			},
			{
				Config: testAccCDNResource_updateLabel(groupName, resourceCName),
				Check: resource.ComposeTestCheckFunc(
					testCDNResourceExists("yandex_cdn_resource.foobar_resource", &cdnResourceUpdated),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "cname", resourceCName),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "origin_group_name", groupName),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "folder_id", folderID),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "labels.%", "1"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "labels.environment", "production"),
					testAccCheckCreatedAtAttr("yandex_cdn_resource.foobar_resource"),
				),
			},
			{
				Config: testAccCDNResource_basicByName(groupName, resourceCName),
				Check: resource.ComposeTestCheckFunc(
					testCDNResourceExists("yandex_cdn_resource.foobar_resource", &cdnResource),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "cname", resourceCName),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "origin_group_name", groupName),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "folder_id", folderID),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "labels.%", "0"),
					testAccCheckCreatedAtAttr("yandex_cdn_resource.foobar_resource"),
				),
			},
		},
	})
}

// TODO: provider

func TestAccCDNResource_Shielding(t *testing.T) {
	t.Parallel()
	groupName := fmt.Sprintf("tf-og-%s", acctest.RandString(10))
	cname := fmt.Sprintf("cdn-%s.yandex.net", acctest.RandString(4))
	locationId := int64(1)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCDNResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCDNResource_shielding(cname, groupName, nil),
				Check:  resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "shielding", ""),
			},
			// enabling shielding
			{
				Config: testAccCDNResource_shielding(cname, groupName, &locationId),
				Check:  resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "shielding", fmt.Sprint(locationId)),
			},
			// disabling shielding
			{
				Config: testAccCDNResource_shielding(cname, groupName, nil),
				Check:  resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "shielding", ""),
			},
		},
	})
}

// Options

func TestAccCDNResource_Option_EdgeCacheSettings(t *testing.T) {
	t.Parallel()

	groupName := fmt.Sprintf("tfog%s", acctest.RandString(10))
	cname := fmt.Sprintf("tf%s.yandex.net", acctest.RandString(4))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCDNResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: makeCDNResourceWithOptions(groupName, cname, ``),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.edge_cache_settings", "86400"),
					resource.TestCheckNoResourceAttr("yandex_cdn_resource.foo", "options.0.edge_cache_settings_codes.0"),
				),
			},
			{
				Config: makeCDNResourceWithOptions(groupName, cname, `edge_cache_settings = 40`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.edge_cache_settings", "40"),
					resource.TestCheckNoResourceAttr("yandex_cdn_resource.foo", "options.0.edge_cache_settings_codes.0"),
				),
			},
			{
				Config: makeCDNResourceWithOptions(groupName, cname, ``),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.edge_cache_settings", "86400"),
					resource.TestCheckNoResourceAttr("yandex_cdn_resource.foo", "options.0.edge_cache_settings_codes.0"),
				),
			},
			{
				Config: makeCDNResourceWithOptions(
					groupName, cname, `edge_cache_settings_codes { value = 80 }`,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.edge_cache_settings", "0"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.edge_cache_settings_codes.0.value", "80"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.edge_cache_settings_codes.0.custom_values.%", "0"),
				),
			},
			{
				Config: makeCDNResourceWithOptions(
					groupName, cname, `edge_cache_settings_codes { value = 40 }`,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.edge_cache_settings", "0"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.edge_cache_settings_codes.0.value", "40"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.edge_cache_settings_codes.0.custom_values.%", "0"),
				),
			},
			{
				Config: makeCDNResourceWithOptions(
					groupName, cname,
					`edge_cache_settings_codes { 
						value = 40
						custom_values = { 
							"200" = 1200
							"400" = 0
						} 
					}`,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.edge_cache_settings", "0"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.edge_cache_settings_codes.0.value", "40"),

					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.edge_cache_settings_codes.0.custom_values.%", "2"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.edge_cache_settings_codes.0.custom_values.200", "1200"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.edge_cache_settings_codes.0.custom_values.400", "0"),
				),
			},
			{
				Config: makeCDNResourceWithOptions(groupName, cname, ``),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.edge_cache_settings", "86400"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.edge_cache_settings_codes.#", "0"),
				),
			},
		},
	})
}

func TestAccCDNResource_Option_BrowserCacheSettings(t *testing.T) {
	t.Parallel()

	groupName := fmt.Sprintf("tf%s", acctest.RandString(10))
	resourceCName := fmt.Sprintf("tf%s.yandex.net", acctest.RandString(4))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCDNResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: makeCDNResourceWithOptions(groupName, resourceCName, ``),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.browser_cache_settings", "0"),
				),
			},
			{
				Config: makeCDNResourceWithOptions(groupName, resourceCName, `browser_cache_settings = 3600`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.browser_cache_settings", "3600"),
				),
			},
			{
				Config: makeCDNResourceWithOptions(groupName, resourceCName, `browser_cache_settings = 2400`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.browser_cache_settings", "2400"),
				),
			},
			{
				Config: makeCDNResourceWithOptions(groupName, resourceCName, ``),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.browser_cache_settings", "0"),
				),
			},
		},
	})
}

func TestAccCDNResource_Option_QueryParams(t *testing.T) {
	t.Parallel()

	groupName := fmt.Sprintf("tf%s", acctest.RandString(10))
	resourceCName := fmt.Sprintf("tf%s.yandex.net", acctest.RandString(4))

	t.Run("ignore_query_params", func(t *testing.T) {
		t.Skip("current provider implementation assumes bug")
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckCDNResourceDestroy,
			Steps: []resource.TestStep{
				{
					Config: makeCDNResourceWithOptions(groupName, resourceCName, ``),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.ignore_query_params", "true"),
						resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.query_params_whitelist.#", "0"),
						resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.query_params_blacklist.#", "0"),
					),
				},
				{
					Config: makeCDNResourceWithOptions(groupName, resourceCName, `ignore_query_params = false`),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.ignore_query_params", "false"),
						resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.query_params_whitelist.#", "0"),
						resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.query_params_blacklist.#", "0"),
					),
				},
				{
					Config: makeCDNResourceWithOptions(groupName, resourceCName, ``),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.ignore_query_params", "true"),
						resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.query_params_whitelist.#", "0"),
						resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.query_params_blacklist.#", "0"),
					),
				},
			},
		})
	})
	// TODO: query_params_whitelist, query_params_blacklist
}

func TestAccCDNResource_Option_Slice(t *testing.T) {
	t.Parallel()

	groupName := fmt.Sprintf("tf%s", acctest.RandString(10))
	resourceCName := fmt.Sprintf("tf%s.yandex.net", acctest.RandString(4))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCDNResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: makeCDNResourceWithOptions(groupName, resourceCName, ``),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.slice", "false"),
				),
			},
			{
				Config: makeCDNResourceWithOptions(groupName, resourceCName, `slice = true`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.slice", "true"),
				),
			},
			{
				Config: makeCDNResourceWithOptions(groupName, resourceCName, `slice = false`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.slice", "false"),
				),
			},
		},
	})
}

func TestAccCDNResource_Option_CompressionOptions(t *testing.T) {
	t.Parallel()

	t.Run("fetched_compressed", func(t *testing.T) {
		t.Skip("current provider implementation assumes bug")
		groupName := fmt.Sprintf("tf%s", acctest.RandString(10))
		resourceCName := fmt.Sprintf("tf%s.yandex.net", acctest.RandString(4))
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckCDNResourceDestroy,
			Steps: []resource.TestStep{
				{
					Config: makeCDNResourceWithOptions(groupName, resourceCName, ``),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.fetched_compressed", "false"),
						resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.gzip_on", "false"),
					),
				},
				{
					Config: makeCDNResourceWithOptions(groupName, resourceCName, `fetched_compressed = true`),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.fetched_compressed", "true"),
						resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.gzip_on", "false"),
					),
				},
				{
					Config: makeCDNResourceWithOptions(groupName, resourceCName, ``),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.fetched_compressed", "false"),
						resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.gzip_on", "false"),
					),
				},
			},
		})
	})
	t.Run("gzip_on", func(t *testing.T) {
		t.Skip("current provider implementation assumes bug")
		groupName := fmt.Sprintf("tf%s", acctest.RandString(10))
		resourceCName := fmt.Sprintf("tf%s.yandex.net", acctest.RandString(4))
		resource.ParallelTest(t, resource.TestCase{
			PreCheck:     func() { testAccPreCheck(t) },
			Providers:    testAccProviders,
			CheckDestroy: testAccCheckCDNResourceDestroy,
			Steps: []resource.TestStep{
				{
					Config: makeCDNResourceWithOptions(groupName, resourceCName, ``),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.fetched_compressed", "false"),
						resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.gzip_on", "false"),
					),
				},
				{
					Config: makeCDNResourceWithOptions(groupName, resourceCName, `gzip_on = true`),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.fetched_compressed", "false"),
						resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.gzip_on", "true"),
					),
				},
				{
					Config: makeCDNResourceWithOptions(groupName, resourceCName, ``),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.fetched_compressed", "false"),
						resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.gzip_on", "false"),
					),
				},
			},
		})
	})
}

// TODO: RedirectOptions

func TestAccCDNResource_Option_HostOption(t *testing.T) {
	t.Parallel()

	t.Skip("current provider implementation assumes bug")
	groupName := fmt.Sprintf("tf%s", acctest.RandString(10))
	resourceCName := fmt.Sprintf("tf%s.yandex.net", acctest.RandString(4))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCDNResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: makeCDNResourceWithOptions(groupName, resourceCName, ``),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.forward_host_header", "false"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.custom_host_header", ""),
				),
			},
			{
				Config: makeCDNResourceWithOptions(groupName, resourceCName, `forward_host_header = true`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.forward_host_header", "true"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.custom_host_header", ""),
				),
			},
			{
				Config: makeCDNResourceWithOptions(groupName, resourceCName, ``),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.forward_host_header", "false"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.custom_host_header", ""),
				),
			},
			{
				Config: makeCDNResourceWithOptions(groupName, resourceCName, `custom_host_header = "google.com"`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.forward_host_header", "false"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.custom_host_header", "google.com"),
				),
			},
			{
				Config: makeCDNResourceWithOptions(groupName, resourceCName, `custom_host_header = "ya.ru"`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.forward_host_header", "false"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.custom_host_header", "ya.ru"),
				),
			},
			{
				Config: makeCDNResourceWithOptions(groupName, resourceCName, ``),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.forward_host_header", "false"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.custom_host_header", ""),
				),
			},
		},
	})
}

func TestAccCDNResource_Option_StaticHeaders(t *testing.T) {
	t.Parallel()

	groupName := fmt.Sprintf("tf%s", acctest.RandString(10))
	resourceCName := fmt.Sprintf("tf%s.yandex.net", acctest.RandString(4))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCDNResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: makeCDNResourceWithOptions(groupName, resourceCName, ``),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.static_response_headers.%", "0"),
				),
			},
			{
				Config: makeCDNResourceWithOptions(groupName, resourceCName,
					`static_response_headers = {
						"key1" = "value1",
					}`,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.static_response_headers.%", "1"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.static_response_headers.key1", "value1"),
				),
			},
			{
				Config: makeCDNResourceWithOptions(groupName, resourceCName,
					`static_response_headers = {
						"key1" = "value1",
						"key2" = "value2",
					}`,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.static_response_headers.%", "2"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.static_response_headers.key1", "value1"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.static_response_headers.key2", "value2"),
				),
			},
			{
				Config: makeCDNResourceWithOptions(groupName, resourceCName, ``),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.static_response_headers.%", "0"),
				),
			},
		},
	})
}

func TestAccCDNResource_Option_Cors(t *testing.T) {
	t.Parallel()

	groupName := fmt.Sprintf("tf%s", acctest.RandString(10))
	resourceCName := fmt.Sprintf("tf%s.yandex.net", acctest.RandString(4))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCDNResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: makeCDNResourceWithOptions(groupName, resourceCName, ``),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.cors.#", "0"),
				),
			},
			{
				Config: makeCDNResourceWithOptions(groupName, resourceCName, `cors = ["google.com"]`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.cors.#", "1"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.cors.0", "google.com"),
				),
			},
			{
				Config: makeCDNResourceWithOptions(groupName, resourceCName, `cors = ["google.com", "ya.ru"]`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.cors.#", "2"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.cors.0", "google.com"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.cors.1", "ya.ru"),
				),
			},
			{
				Config: makeCDNResourceWithOptions(groupName, resourceCName, ``),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.cors.#", "0"),
				),
			},
		},
	})
}

// TODO: stale

func TestAccCDNResource_Option_AllowedHttpMethods(t *testing.T) {
	t.Parallel()

	t.Skip("current provider implementation assumes bug")
	groupName := fmt.Sprintf("tf%s", acctest.RandString(10))
	resourceCName := fmt.Sprintf("tf%s.yandex.net", acctest.RandString(4))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCDNResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: makeCDNResourceWithOptions(groupName, resourceCName, ``),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.allowed_http_methods.#", "3"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.allowed_http_methods.0", "GET"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.allowed_http_methods.1", "HEAD"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.allowed_http_methods.2", "OPTIONS"),
				),
			},
			{
				Config: makeCDNResourceWithOptions(groupName, resourceCName, `allowed_http_methods = ["GET", "POST"]`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.allowed_http_methods.#", "2"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.allowed_http_methods.0", "GET"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.allowed_http_methods.1", "POST"),
				),
			},
			{
				Config: makeCDNResourceWithOptions(groupName, resourceCName, `allowed_http_methods = ["GET"]`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.allowed_http_methods.#", "1"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.allowed_http_methods.0", "GET"),
				),
			},
			{
				Config: makeCDNResourceWithOptions(groupName, resourceCName, ``),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.allowed_http_methods.#", "3"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.allowed_http_methods.0", "GET"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.allowed_http_methods.1", "HEAD"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.allowed_http_methods.2", "OPTIONS"),
				),
			},
		},
	})
}

// TODO: proxy_cache_methods_set
// TODO: disable_proxy_force_ranges

func TestAccCDNResource_Option_StaticRequestHeaders(t *testing.T) {
	t.Parallel()

	groupName := fmt.Sprintf("tf%s", acctest.RandString(10))
	resourceCName := fmt.Sprintf("tf%s.yandex.net", acctest.RandString(4))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCDNResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: makeCDNResourceWithOptions(groupName, resourceCName, ``),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.static_request_headers.%", "0"),
				),
			},
			{
				Config: makeCDNResourceWithOptions(groupName, resourceCName,
					`static_request_headers = {
						"key1" = "value1",
					}`,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.static_request_headers.%", "1"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.static_request_headers.key1", "value1"),
				),
			},
			{
				Config: makeCDNResourceWithOptions(groupName, resourceCName,
					`static_request_headers = {
						"key1" = "value1",
						"key2" = "value2",
					}`,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.static_request_headers.%", "2"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.static_request_headers.key1", "value1"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.static_request_headers.key2", "value2"),
				),
			},
			{
				Config: makeCDNResourceWithOptions(groupName, resourceCName, ``),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.static_request_headers.%", "0"),
				),
			},
		},
	})
}

func TestAccCDNResource_Option_IgnoreCookie(t *testing.T) {
	t.Parallel()

	groupName := fmt.Sprintf("tf%s", acctest.RandString(10))
	resourceCName := fmt.Sprintf("tf%s.yandex.net", acctest.RandString(4))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCDNResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: makeCDNResourceWithOptions(groupName, resourceCName, ``),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.ignore_cookie", "true"),
				),
			},
			{
				Config: makeCDNResourceWithOptions(groupName, resourceCName, `ignore_cookie = false`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.ignore_cookie", "false"),
				),
			},
			{
				Config: makeCDNResourceWithOptions(groupName, resourceCName, ``),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_cdn_resource.foo", "options.0.ignore_cookie", "true"),
				),
			},
		},
	})
}

// TODO: rewrite

func TestAccCDNResource_Option_SecureKey(t *testing.T) {
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
				Config: testAccCDNResource_optionSecureKey(groupName, resourceCName),
				Check: resource.ComposeTestCheckFunc(
					testCDNResourceExists("yandex_cdn_resource.foobar_resource", &cdnResource),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "cname", resourceCName),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "folder_id", folderID),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "options.#", "1"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "options.0.secure_key", "testsecurekey"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "options.0.enable_ip_url_signing", "true"),
					testAccCheckCreatedAtAttr("yandex_cdn_resource.foobar_resource"),
				),
			},
		},
	})
}

func TestAccCDNResource_Option_IPAddressACL(t *testing.T) {
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
				Config: testAccCDNResource_optionIPAddressACL(groupName, resourceCName),
				Check: resource.ComposeTestCheckFunc(
					testCDNResourceExists("yandex_cdn_resource.foobar_resource", &cdnResource),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "cname", resourceCName),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "folder_id", folderID),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "options.#", "1"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "options.0.ip_address_acl.#", "1"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "options.0.ip_address_acl.0.policy_type", "allow"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "options.0.ip_address_acl.0.excepted_values.#", "1"),
					resource.TestCheckResourceAttr("yandex_cdn_resource.foobar_resource", "options.0.ip_address_acl.0.excepted_values.0", "192.168.3.2/32"),
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

func testAccCDNResource_optionSecureKey(groupName, resourceCNAME string) string {
	return makeGroupResource(groupName) + fmt.Sprintf(`
resource "yandex_cdn_resource" "foobar_resource" {
	cname = "%s"

	origin_group_id = "${yandex_cdn_origin_group.foo_cdn_group.id}"

	options {
		secure_key = "testsecurekey"
		enable_ip_url_signing = true
	}
}
`, resourceCNAME)
}

func testAccCDNResource_optionIPAddressACL(groupName, resourceCNAME string) string {
	return makeGroupResource(groupName) + fmt.Sprintf(`
resource "yandex_cdn_resource" "foobar_resource" {
	cname = "%s"

	origin_group_id = "${yandex_cdn_origin_group.foo_cdn_group.id}"

	options {
		ip_address_acl {
			policy_type = "allow"
			excepted_values = ["192.168.3.2/32"]
		}
	}
}
`, resourceCNAME)
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

func testAccCDNResource_addLabel(groupName, resourceCNAME string) string {
	return makeGroupResourceByName(groupName) + fmt.Sprintf(`
resource "yandex_cdn_resource" "foobar_resource" {
	cname = "%s"

	origin_group_name = "${yandex_cdn_origin_group.foo_cdn_group_by_name.name}"

	labels = {
		environment = "testing"
	}
}
`, resourceCNAME)
}

func testAccCDNResource_updateLabel(groupName, resourceCNAME string) string {
	return makeGroupResourceByName(groupName) + fmt.Sprintf(`
resource "yandex_cdn_resource" "foobar_resource" {
	cname = "%s"

	origin_group_name = "${yandex_cdn_origin_group.foo_cdn_group_by_name.name}"

	labels = {
		environment = "production"
	}
}
`, resourceCNAME)
}

func testAccCDNResource_shielding(cname string, groupName string, location *int64) string {
	tmp := makeGroupResource(groupName) + `
resource "yandex_cdn_resource" "foo" {
	cname = "%s"
	origin_group_name = yandex_cdn_origin_group.foo_cdn_group.name
	%s
}`
	shielding := ""
	if location != nil {
		shielding = fmt.Sprintf(`shielding = "%v"`, *location)
	}
	return fmt.Sprintf(tmp, cname, shielding)
}

func makeCDNResourceWithOptions(groupName string, cname string, options string) string {
	return makeGroupResource(groupName) + fmt.Sprintf(`
resource "yandex_cdn_resource" "foo" {
	cname = "%s"
	origin_group_name = yandex_cdn_origin_group.foo_cdn_group.name
	options {
		%s
	}
}`, cname, options)
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

func makeGroupResourceByName(groupName string) string {
	return fmt.Sprintf(`
	resource "yandex_cdn_origin_group" "foo_cdn_group_by_name" {
		name     = "%s"

		origin {
			source = "ya.ru"
		}
	}
	`, groupName)
}
