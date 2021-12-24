package yandex

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/go-multierror"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/cdn/v1"
)

func init() {
	resource.AddTestSweepers("yandex_cdn_origin_group", &resource.Sweeper{
		Name: "yandex_cdn_origin_group",
		F:    testSweepCDNOriginGroups,
		Dependencies: []string{
			"yandex_cdn_resource",
		},
	})
}

func testAccCDNOriginsContainsSources(originGroup *cdn.OriginGroup, sources ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if originGroup == nil {
			return fmt.Errorf("CDN Origin Group must exist")
		}

		if len(originGroup.Origins) != len(sources) {
			return fmt.Errorf("Unexpected origins count, should be %d, but found %d",
				len(sources),
				len(originGroup.Origins),
			)
		}

		sourceSet := make(map[string]struct{})
		for _, origin := range originGroup.Origins {
			sourceSet[origin.Source] = struct{}{}
		}

		for _, source := range sources {
			if _, ok := sourceSet[source]; !ok {
				return fmt.Errorf("Source %s is missing in resulting CDN Origin group", source)
			}
		}

		return nil
	}
}

func TestAccCDNOriginGroup_basic(t *testing.T) {
	groupName := fmt.Sprintf("tf-test-cdn-origin-group-basic-%s", acctest.RandString(10))
	var originGroup cdn.OriginGroup

	folderID := getExampleFolderID()

	extractStringGroupID := func() string {
		return strconv.FormatInt(originGroup.Id, 10)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCDNOriginGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCDNOriginGroup_basic(groupName),
				Check: resource.ComposeTestCheckFunc(
					testOriginGroupExists("yandex_cdn_origin_group.test_cdn_group", &originGroup),
					resource.TestCheckResourceAttr("yandex_cdn_origin_group.test_cdn_group", "name", groupName),
					resource.TestCheckResourceAttr("yandex_cdn_origin_group.test_cdn_group", "folder_id", folderID),
					resource.TestCheckResourceAttr("yandex_cdn_origin_group.test_cdn_group", "use_next", "true"),
					testAccCDNOriginsContainsSources(&originGroup,
						"ya.ru",
						"yandex.ru",
						"goo.gl",
						"amazon.com",
					),
				),
			},
			{
				ResourceName: "yandex_cdn_origin_group.test_cdn_group",
				ImportStateIdFunc: func(*terraform.State) (string, error) {
					return extractStringGroupID(), nil
				},
				ImportState:             true,
				ImportStateVerifyIgnore: []string{"origin"},
				ImportStateVerify:       true,
			},
		},
	})
}

func TestAccCDNOriginGroup_update(t *testing.T) {
	groupName := fmt.Sprintf("tf-test-cdn-origin-group-source-%s", acctest.RandString(10))
	groupNameUpdated := fmt.Sprintf("tf-test-cdn-origin-group-target-%s", acctest.RandString(10))

	var originGroup, originGroupUpdated cdn.OriginGroup

	folderID := getExampleFolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCDNOriginGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCDNOriginGroup_basic(groupName),
				Check: resource.ComposeTestCheckFunc(
					testOriginGroupExists("yandex_cdn_origin_group.test_cdn_group", &originGroup),
					resource.TestCheckResourceAttr("yandex_cdn_origin_group.test_cdn_group", "name", groupName),
					resource.TestCheckResourceAttr("yandex_cdn_origin_group.test_cdn_group", "folder_id", folderID),
					resource.TestCheckResourceAttr("yandex_cdn_origin_group.test_cdn_group", "use_next", "true"),
					testAccCDNOriginsContainsSources(&originGroup,
						"ya.ru",
						"yandex.ru",
						"goo.gl",
						"amazon.com",
					),
				),
			},
			{
				Config: testAccCDNOriginGroup_update(groupNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testOriginGroupExists("yandex_cdn_origin_group.test_cdn_group", &originGroupUpdated),
					resource.TestCheckResourceAttr("yandex_cdn_origin_group.test_cdn_group", "name", groupNameUpdated),
					resource.TestCheckResourceAttr("yandex_cdn_origin_group.test_cdn_group", "folder_id", folderID),
					resource.TestCheckResourceAttr("yandex_cdn_origin_group.test_cdn_group", "use_next", "false"),
					testAccCDNOriginsContainsSources(&originGroupUpdated,
						"ya.ru",
					),
				),
			},
		},
	})
}

func testSweepCDNOriginGroups(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &cdn.ListOriginGroupsRequest{FolderId: conf.FolderID}
	it := conf.sdk.CDN().OriginGroup().OriginGroupIterator(conf.Context(), req)
	result := &multierror.Error{}

	for it.Next() {
		id := it.Value().GetId()
		if !sweepCDNOriginGroup(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep CDN Origin group %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepCDNOriginGroup(conf *Config, id int64) bool {
	return sweepWithRetryByFunc(conf, "Origin Group", func(conf *Config) error {
		ctx, cancel := conf.ContextWithTimeout(yandexCDNOriginGroupDefaultTimeout)
		defer cancel()

		op, err := conf.sdk.CDN().OriginGroup().Delete(ctx, &cdn.DeleteOriginGroupRequest{
			OriginGroupId: id,
		})

		return handleSweepOperation(ctx, conf, op, err)
	})
}

func testAccCheckCDNOriginGroupDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_cdn_origin_group" {
			continue
		}

		id, err := strconv.ParseInt(rs.Primary.ID, 10, 64)
		if err != nil {
			return err
		}

		_, err = config.sdk.CDN().OriginGroup().Get(context.Background(), &cdn.GetOriginGroupRequest{
			OriginGroupId: id,
		})

		if err == nil {
			return fmt.Errorf(" still exists")
		}
	}

	return nil
}

func testOriginGroupExists(resourceName string, originGroup *cdn.OriginGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("origin group not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		config := testAccProvider.Meta().(*Config)
		groupID, err := strconv.ParseInt(rs.Primary.ID, 10, 64)
		if err != nil {
			return err
		}

		folderID := getExampleFolderID()

		found, err := config.sdk.CDN().OriginGroup().Get(context.Background(), &cdn.GetOriginGroupRequest{
			FolderId:      folderID,
			OriginGroupId: groupID,
		})

		if err != nil {
			return err
		}

		if strconv.FormatInt(found.Id, 10) != rs.Primary.ID {
			return fmt.Errorf("origin group is not found")
		}

		//goland:noinspection GoVetCopyLock
		*originGroup = *found

		return nil
	}
}

func testAccCDNOriginGroup_basic(groupName string) string {
	return fmt.Sprintf(`
resource "yandex_cdn_origin_group" "test_cdn_group" {
  name     = "%s"

  origin {
	source = "ya.ru"
  }

  origin {
	source = "yandex.ru"
  }

  origin {
	source = "goo.gl"
  }

  origin {
	source = "amazon.com"
  }
}
`, groupName)
}

func testAccCDNOriginGroup_update(groupName string) string {
	return fmt.Sprintf(`
resource "yandex_cdn_origin_group" "test_cdn_group" {
  name     = "%s"

  use_next = false

  origin {
	source = "ya.ru"
  }
}
`, groupName)
}
