package acceptance

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/acceptanceenv"
	rmfixtures "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/resourcemanager/testfixtures"

	statecheck "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance/statecheck"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
)

//revive:disable:var-naming
func TestAccDataSourceYandexResourceManagerCloud_byIDNotFound(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	notExistCloudID := acctest.RandString(18)

	CustomProvidersTest(t, suite.defaultAndEmptyFolderProviders(t), resource.TestCase{
		PreCheck: func() { acceptance.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: rmfixtures.CloudDataSourceByID(notExistCloudID),
				// "PermissionDenied" returned for non existed cloud id
				ExpectError: regexp.MustCompile("NotFound"),
			},
		},
	})
}

func TestAccDataSourceYandexResourceManagerCloud_byDefaultID(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	defaultCloudID := acceptanceenv.CloudID()
	defaultCloudName := acceptanceenv.CloudName()

	CustomProvidersTest(t, suite.defaultAndEmptyFolderProviders(t), resource.TestCase{
		PreCheck: func() { acceptance.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: rmfixtures.CloudDataSourceByID(defaultCloudID),
				Check: resource.ComposeTestCheckFunc(
					statecheck.CheckResourceIDField("data.yandex_resourcemanager_cloud.acceptance", "cloud_id"),
					resource.TestCheckResourceAttr("data.yandex_resourcemanager_cloud.acceptance", "id", defaultCloudID),
					resource.TestCheckResourceAttr("data.yandex_resourcemanager_cloud.acceptance", "name", defaultCloudName),
					statecheck.CheckCreatedAtAttr("data.yandex_resourcemanager_cloud.acceptance"),
				),
			},
		},
	})
}

func TestAccDataSourceYandexResourceManagerCloud_byDefaultCloudName(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	defaultCloudName := acceptanceenv.CloudName()

	CustomProvidersTest(t, suite.defaultAndEmptyFolderProviders(t), resource.TestCase{
		PreCheck: func() { acceptance.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccCheckResourceManagerCloud_byName(defaultCloudName),
				Check: resource.ComposeTestCheckFunc(
					statecheck.CheckResourceIDField("data.yandex_resourcemanager_cloud.acceptance", "cloud_id"),
					resource.TestCheckResourceAttrSet("data.yandex_resourcemanager_cloud.acceptance", "id"),
					resource.TestCheckResourceAttrSet("data.yandex_resourcemanager_cloud.acceptance", "created_at"),
					resource.TestCheckResourceAttr("data.yandex_resourcemanager_cloud.acceptance", "name", defaultCloudName),
					statecheck.CheckCreatedAtAttr("data.yandex_resourcemanager_cloud.acceptance"),
				),
			},
		},
	})
}

func TestAccDataSourceYandexResourceManagerCloud_byName(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	cloudName := acctest.RandString(10)

	CustomProvidersTest(t, suite.defaultAndEmptyFolderProviders(t), resource.TestCase{
		PreCheck: func() { acceptance.PreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckResourceManagerCloud_byName(cloudName),
				ExpectError: regexp.MustCompile(`failed to resolve data source cloud by name: cloud with name "` + cloudName + `" not found`),
			},
		},
	})
}

func testAccCheckResourceManagerCloud_byName(name string) string {
	return fmt.Sprintf(`
data "yandex_resourcemanager_cloud" "acceptance" {
  name = "%s"
}
`, name)
}
