package acceptance

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/acceptanceenv"

	statecheck "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance/statecheck"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
)

//revive:disable:var-naming
func TestAccDataSourceYandexResourceManagerFolder_byID(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	folderID := acceptanceenv.FolderID()
	folderName := acceptanceenv.FolderName()

	CustomProvidersTest(t, suite.defaultAndEmptyFolderProviders(t), resource.TestCase{
		PreCheck:  func() { acceptance.PreCheck(t) },
		Providers: suite.Providers(),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckYandexResourceManagerFolder_byID(folderID),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceYandexResourceManagerFolderCheck("data.yandex_resourcemanager_folder.folder", folderID, folderName),
					statecheck.CheckResourceIDField("data.yandex_resourcemanager_folder.folder", "folder_id"),
					statecheck.CheckCreatedAtAttr("data.yandex_resourcemanager_folder.folder"),
				),
			},
		},
	})
}

func TestAccDataSourceYandexResourceManagerFolder_byName(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	folderID := acceptanceenv.FolderID()
	folderName := acceptanceenv.FolderName()

	CustomProvidersTest(t, suite.defaultAndEmptyFolderProviders(t), resource.TestCase{
		PreCheck:  func() { acceptance.PreCheck(t) },
		Providers: suite.Providers(),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckYandexResourceManagerFolder_byName(folderName),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceYandexResourceManagerFolderCheck("data.yandex_resourcemanager_folder.folder", folderID, folderName),
					statecheck.CheckResourceIDField("data.yandex_resourcemanager_folder.folder", "folder_id"),
					statecheck.CheckCreatedAtAttr("data.yandex_resourcemanager_folder.folder"),
				),
			},
		},
	})
}

func TestAccDataSourceYandexResourceManagerFolder_byNameAndCloudID(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	folderName := acceptanceenv.FolderName()
	cloudID := acceptanceenv.CloudID()
	folderID := acceptanceenv.FolderID()

	CustomProvidersTest(t, suite.defaultAndEmptyFolderProviders(t), resource.TestCase{
		PreCheck:  func() { acceptance.PreCheck(t) },
		Providers: suite.Providers(),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckYandexResourceManagerFolder_byNameAndCloudID(folderName, cloudID),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceYandexResourceManagerFolderCheck("data.yandex_resourcemanager_folder.folder", folderID, folderName),
					statecheck.CheckResourceIDField("data.yandex_resourcemanager_folder.folder", "folder_id"),
					resource.TestCheckResourceAttr("data.yandex_resourcemanager_folder.folder", "cloud_id", cloudID),
					statecheck.CheckCreatedAtAttr("data.yandex_resourcemanager_folder.folder"),
				),
			},
		},
	})
}

func TestAccDataSourceYandexResourceManagerFolder_wrongCloudID(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	folderName := acceptanceenv.FolderName()
	wrongCloudID := acctest.RandString(12)

	CustomProvidersTest(t, suite.defaultAndEmptyFolderProviders(t), resource.TestCase{
		PreCheck:  func() { acceptance.PreCheck(t) },
		Providers: suite.Providers(),
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckYandexResourceManagerFolder_wrongCloudID(folderName, wrongCloudID),
				ExpectError: regexp.MustCompile("NotFound"),
			},
		},
	})
}

func TestAccDataSourceYandexResourceManagerFolder_byIDNotFound(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	name := "terraform-test-" + acctest.RandString(12)

	CustomProvidersTest(t, suite.defaultAndEmptyFolderProviders(t), resource.TestCase{
		PreCheck:  func() { acceptance.PreCheck(t) },
		Providers: suite.Providers(),
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckYandexResourceManagerFolder_byIDNotFound(name),
				ExpectError: regexp.MustCompile("NotFound"),
			},
		},
	})
}

func testAccDataSourceYandexResourceManagerFolderCheck(dataSourceName string, folderID, folderName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ds, ok := s.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", dataSourceName)
		}

		attributes := ds.Primary.Attributes

		if attributes["folder_id"] != folderID {
			return fmt.Errorf("folder_id attr is %s; want %s", attributes["folder_id"], folderID)
		}

		if attributes["name"] != folderName {
			return fmt.Errorf("Name attr is %s; want %s", attributes["name"], folderName)
		}

		return nil
	}
}

func testAccCheckYandexResourceManagerFolder_byID(folderID string) string {
	return fmt.Sprintf(`
data "yandex_resourcemanager_folder" "folder" {
  folder_id = "%s"
}
`, folderID)
}

func testAccCheckYandexResourceManagerFolder_byName(folderName string) string {
	return fmt.Sprintf(`
data "yandex_resourcemanager_folder" "folder" {
  name = "%s"
}
`, folderName)
}

func testAccCheckYandexResourceManagerFolder_byNameAndCloudID(folderName, cloudID string) string {
	return fmt.Sprintf(`
data "yandex_resourcemanager_folder" "folder" {
  name     = "%s"
  cloud_id = "%s"
}
`, folderName, cloudID)
}

func testAccCheckYandexResourceManagerFolder_byIDNotFound(folderID string) string {
	return fmt.Sprintf(`
data "yandex_resourcemanager_folder" "folder" {
  folder_id = "%s"
}
`, folderID)
}

func testAccCheckYandexResourceManagerFolder_wrongCloudID(folderName, cloudID string) string {
	return fmt.Sprintf(`
data "yandex_resourcemanager_folder" "folder" {
  name     = "%s"
  cloud_id = "%s"
}
`, folderName, cloudID)
}
