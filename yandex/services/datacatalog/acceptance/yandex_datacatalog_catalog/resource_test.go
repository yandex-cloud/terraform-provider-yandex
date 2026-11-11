package yandex_datacatalog_catalog_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

var testCatalogResourceName = "yandex_datacatalog_catalog.test-catalog"

func TestAccDatacatalogCatalogResource_basic(t *testing.T) {
	var (
		folderId           = testhelpers.GetExampleFolderID()
		catalogName        = testhelpers.ResourceName(50)
		catalogDescription = acctest.RandStringFromCharSet(256, acctest.CharSetAlpha)
		labelKey           = acctest.RandStringFromCharSet(63, acctest.CharSetAlpha)
		labelValue         = acctest.RandStringFromCharSet(63, acctest.CharSetAlphaNum)
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testhelpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: testhelpers.AccProviderFactories,
		CheckDestroy:             AccCheckCatalogDestroy,
		Steps: []resource.TestStep{
			basicCatalogTestStep(folderId, catalogName, catalogDescription, labelKey, labelValue),
		},
	})
}

func TestAccDatacatalogCatalogResource_update(t *testing.T) {
	var (
		folderId           = testhelpers.GetExampleFolderID()
		catalogName        = testhelpers.ResourceName(50)
		catalogDescription = acctest.RandStringFromCharSet(256, acctest.CharSetAlpha)
		labelKey           = acctest.RandStringFromCharSet(63, acctest.CharSetAlpha)
		labelValue         = acctest.RandStringFromCharSet(63, acctest.CharSetAlphaNum)
		updatedDescription = acctest.RandStringFromCharSet(250, acctest.CharSetAlpha)
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testhelpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: testhelpers.AccProviderFactories,
		CheckDestroy:             AccCheckCatalogDestroy,
		Steps: []resource.TestStep{
			basicCatalogTestStep(folderId, catalogName, catalogDescription, labelKey, labelValue),
			updateCatalogTestStep(folderId, catalogName, updatedDescription, labelKey, labelKey),
		},
	})
}

func testCatalogBasic(folderID, name, description, labelKey, labelValue string) string {
	return fmt.Sprintf(`
resource "yandex_datacatalog_catalog" "test-catalog" {
  folder_id = "%s"
  name = "%s"
  description = "%s"
  labels = {
    "%s" = "%s"
  }
}
`, folderID, name, description, labelKey, labelValue)
}

func basicCatalogTestStep(folderId, name, description, labelKey, labelValue string) resource.TestStep {
	return resource.TestStep{
		Config: testCatalogBasic(folderId, name, description, labelKey, labelValue),
		Check: resource.ComposeTestCheckFunc(
			CatalogExists(testCatalogResourceName),
			resource.TestCheckResourceAttr(testCatalogResourceName, "folder_id", folderId),
			resource.TestCheckResourceAttr(testCatalogResourceName, "name", name),
			resource.TestCheckResourceAttr(testCatalogResourceName, "description", description),
			resource.TestCheckResourceAttr(testCatalogResourceName, fmt.Sprintf("labels.%s", labelKey), labelValue),
			resource.TestCheckResourceAttrSet(testCatalogResourceName, "created_at"),
			resource.TestCheckResourceAttrSet(testCatalogResourceName, "updated_at"),
			resource.TestCheckResourceAttrSet(testCatalogResourceName, "created_by"),
		),
	}
}

func updateCatalogTestStep(folderId, name, updatedDescription, labelKey, labelValue string) resource.TestStep {
	return resource.TestStep{
		Config: testCatalogBasic(folderId, name, updatedDescription, labelKey, labelValue),
		Check: resource.ComposeTestCheckFunc(
			CatalogExists(testCatalogResourceName),
			resource.TestCheckResourceAttr(testCatalogResourceName, "folder_id", folderId),
			resource.TestCheckResourceAttr(testCatalogResourceName, "name", name),
			resource.TestCheckResourceAttr(testCatalogResourceName, "description", updatedDescription),
			resource.TestCheckResourceAttr(testCatalogResourceName, fmt.Sprintf("labels.%s", labelKey), labelValue),
			resource.TestCheckResourceAttrSet(testCatalogResourceName, "updated_at"),
		),
	}
}
