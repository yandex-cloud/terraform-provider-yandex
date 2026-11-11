package acceptance

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/acceptanceenv"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/cdn/v1"
	cdnsdk "github.com/yandex-cloud/go-sdk/services/cdn/v1"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
)

func testAccCDNOriginsContainsSources(originGroup *cdn.OriginGroup, sources ...string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
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
	suite := testSuite{acceptance.New(t)}
	groupName := fmt.Sprintf("tf-test-cdn-origin-group-basic-%s", acctest.RandString(10))
	var originGroup cdn.OriginGroup

	folderID := acceptanceenv.FolderID()

	extractStringGroupID := func() string {
		return strconv.FormatInt(originGroup.Id, 10)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testAccCheckCDNOriginGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCDNOriginGroup_basic(groupName),
				Check: resource.ComposeTestCheckFunc(
					suite.testOriginGroupExists("yandex_cdn_origin_group.test_cdn_group", &originGroup),
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

func TestAccCDNOriginGroup_basicWithFolderID(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	groupName := fmt.Sprintf("tf-test-cdn-origin-group-basic-%s", acctest.RandString(10))
	var originGroup cdn.OriginGroup

	folderID := acceptanceenv.FolderID()

	extractStringGroupID := func() string {
		return strconv.FormatInt(originGroup.Id, 10)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testAccCheckCDNOriginGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCDNOriginGroup_basicWithFolderID(groupName, folderID),
				Check: resource.ComposeTestCheckFunc(
					suite.testOriginGroupExists("yandex_cdn_origin_group.test_cdn_group", &originGroup),
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
	suite := testSuite{acceptance.New(t)}
	groupName := fmt.Sprintf("tf-test-cdn-origin-group-source-%s", acctest.RandString(10))
	groupNameUpdated := fmt.Sprintf("tf-test-cdn-origin-group-target-%s", acctest.RandString(10))

	var originGroup, originGroupUpdated cdn.OriginGroup

	folderID := acceptanceenv.FolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testAccCheckCDNOriginGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCDNOriginGroup_basic(groupName),
				Check: resource.ComposeTestCheckFunc(
					suite.testOriginGroupExists("yandex_cdn_origin_group.test_cdn_group", &originGroup),
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
					suite.testOriginGroupExists("yandex_cdn_origin_group.test_cdn_group", &originGroupUpdated),
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

func (suite testSuite) testAccCheckCDNOriginGroupDestroy(s *terraform.State) error {
	config, configErr := suite.Config()
	if configErr != nil {
		return configErr
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_cdn_origin_group" {
			continue
		}

		id, err := strconv.ParseInt(rs.Primary.ID, 10, 64)
		if err != nil {
			return err
		}

		client := cdnsdk.NewOriginGroupClient(config.SDK)

		_, err = client.Get(context.Background(), &cdn.GetOriginGroupRequest{
			OriginGroupId: id,
		})

		if err == nil {
			return fmt.Errorf(" still exists")
		}
	}

	return nil
}

func (suite testSuite) testOriginGroupExists(resourceName string, originGroup *cdn.OriginGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("origin group not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		config, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}

		groupID, err := strconv.ParseInt(rs.Primary.ID, 10, 64)
		if err != nil {
			return err
		}

		folderID := acceptanceenv.FolderID()

		client := cdnsdk.NewOriginGroupClient(config.SDK)

		found, err := client.Get(context.Background(), &cdn.GetOriginGroupRequest{
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

func testAccCDNOriginGroup_basicWithFolderID(groupName, folderID string) string {
	return fmt.Sprintf(`
resource "yandex_cdn_origin_group" "test_cdn_group" {
  name      = "%s"
  folder_id = "%s"

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

`, groupName, folderID)
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
