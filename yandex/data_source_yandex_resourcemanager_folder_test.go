package yandex

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

//revive:disable:var-naming
func TestAccDataSourceYandexResourceManagerFolder_byID(t *testing.T) {
	folderID := getExampleFolderID()
	folderName := getExampleFolderName()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckYandexResourceManagerFolder_byID(folderID),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceYandexResourceManagerFolderCheck("data.yandex_resourcemanager_folder.folder", folderName),
					resource.TestCheckResourceAttrSet("data.yandex_resourcemanager_folder.folder", "created_at"),
				),
			},
		},
	})
}

func TestAccDataSourceYandexResourceManagerFolder_byIDNotFound(t *testing.T) {
	name := "terraform-test-" + acctest.RandString(12)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckYandexResourceManagerFolder_byIDNotFound(name),
				ExpectError: regexp.MustCompile("rpc error: code = PermissionDenied desc = You are not authorized for this operation."),
			},
		},
	})
}

func testAccDataSourceYandexResourceManagerFolderCheck(data_source_name string, folderName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ds, ok := s.RootModule().Resources[data_source_name]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", data_source_name)
		}

		ds_attr := ds.Primary.Attributes

		if ds_attr["name"] != folderName {
			return fmt.Errorf("Name attr is %s; want %s", ds_attr["name"], folderName)

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

func testAccCheckYandexResourceManagerFolder_byIDNotFound(folderID string) string {
	return fmt.Sprintf(`
data "yandex_resourcemanager_folder" "folder" {
  folder_id = "%s"
}
`, folderID)
}
