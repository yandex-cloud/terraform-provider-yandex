package yandex_cloud_desktops_image_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

// tests that resource can be accessed both ways: by ID and by folderID and name
func TestAccDataSourceCloudDesktopsImage_basic(t *testing.T) {
	t.Parallel()

	name := acctest.RandomWithPrefix("normal-name")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceDesktopsImageConfig(name+"_1", name+"_2", name+"_3"),
				Check:  testAccDataSourceDesktopsImagesCheck(name+"_1", name+"_2", name+"_3"),
			},
		},
	})
}

const (
	imageID   = "fdvvheamqk751hr09co9"
	imageName = "Ubuntu 20.04 LTS (2024-12-03)"
)

var (
	imageLabels = map[string]string{"x-hopper-operation-id": "d9pvdgi26mio0o1ci5l1", "x-hopper-source-image-id": "fd8pi4srg296eb52abuk"}
)

func testAccDataSourceDesktopsImagesCheck(name1, name2, name3 string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccDataSourceDesktopsImageCheck(name1, imageID, imageName, "", imageLabels),
		testAccDataSourceDesktopsImageCheck(name2, imageID, imageName, "", imageLabels),
		testAccDataSourceDesktopsImageCheck(name3, imageID, imageName, "", imageLabels),
	)
}

func testAccDataSourceDesktopsImageCheck(rsName, id, name, folder_id string, labels map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources["data.yandex_cloud_desktops_image."+rsName]
		if !ok {
			return fmt.Errorf("Datasource not found: %s", rsName)
		}

		if err := checkAttributesEquality(rs.Primary.Attributes, "id", id); err != nil {
			return fmt.Errorf("DataSource %s attributes aren't the expected ones: %w", rsName, err)
		}
		if err := checkAttributesEquality(rs.Primary.Attributes, "name", name); err != nil {
			return fmt.Errorf("DataSource %s attributes aren't the expected ones: %w", rsName, err)
		}
		if err := checkAttributesEquality(rs.Primary.Attributes, "folder_id", folder_id); err != nil {
			return fmt.Errorf("DataSource %s attributes aren't the expected ones: %w", rsName, err)
		}
		for labelKey, labelVal := range labels {
			fullKey := "labels." + labelKey
			if err := checkAttributesEquality(rs.Primary.Attributes, fullKey, labelVal); err != nil {
				return fmt.Errorf("DataSource %s label attributes aren't the expected ones: %w", rsName, err)
			}
		}
		return nil
	}
}

func checkAttributesEquality(stateAttributes map[string]string, field, expected string) error {
	actual, ok := stateAttributes[field]
	if !ok {
		return fmt.Errorf("state field %s not found", field)
	}
	if actual != expected {
		return fmt.Errorf("state field %s is not the expected: expected = %s, actual = %s", field, expected, actual)
	}
	return nil
}

func testAccDataSourceDesktopsImageConfig(name1, name2, name3 string) string {
	return fmt.Sprintf(`
data "yandex_cloud_desktops_image" "%s" {
	id = "fdvvheamqk751hr09co9"
}
	
data "yandex_cloud_desktops_image" "%s" {
	folder_id = "%s"
	name 	  = "Ubuntu 20.04 LTS (2024-12-03)"
}

data "yandex_cloud_desktops_image" "%s" {
	name = "Ubuntu 20.04 LTS (2024-12-03)"
}
`, name1, name2, test.GetExampleFolderID(), name3)
}
