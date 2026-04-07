package yandex_datacatalog_catalog_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/datacatalog/v1"
	datacatalogv1sdk "github.com/yandex-cloud/go-sdk/services/datacatalog/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	"google.golang.org/grpc"

	"google.golang.org/grpc/metadata"
)

const testCatalogDataSourceName = "data.yandex_datacatalog_catalog.test-catalog-data"

func TestAccDataSourceDatacatalogCatalog(t *testing.T) {
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
			{
				Config: testCatalogDataConfig(folderId, catalogName, catalogDescription, labelKey, labelValue),
				Check: resource.ComposeTestCheckFunc(
					CatalogExists(testCatalogDataSourceName),
					resource.TestCheckResourceAttr(testCatalogDataSourceName, "folder_id", folderId),
					resource.TestCheckResourceAttr(testCatalogDataSourceName, "name", catalogName),
					resource.TestCheckResourceAttr(testCatalogDataSourceName, "description", catalogDescription),
					resource.TestCheckResourceAttr(testCatalogDataSourceName, fmt.Sprintf("labels.%s", labelKey), labelValue),
					resource.TestCheckResourceAttrSet(testCatalogDataSourceName, "created_at"),
					resource.TestCheckResourceAttrSet(testCatalogDataSourceName, "updated_at"),
					resource.TestCheckResourceAttrSet(testCatalogDataSourceName, "created_by"),
				),
			},
		},
	})
}

func testCatalogDataConfig(folderID, name, description, labelKey, labelValue string) string {
	return fmt.Sprintf(`
data "yandex_datacatalog_catalog" "test-catalog-data" {
	catalog_id = yandex_datacatalog_catalog.test-catalog.id
}

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

func CatalogExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testhelpers.AccProvider.(*yandex_framework.Provider).GetConfig()
		a := s.RootModule().Resources
		fmt.Printf("%s", a)
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		id := rs.Primary.ID

		reqApi := &datacatalog.GetCatalogRequest{
			CatalogId: id,
		}
		md := new(metadata.MD)

		found, err := datacatalogv1sdk.NewCatalogClient(config.SDKv2).GetCatalog(context.Background(), reqApi, grpc.Header(md))
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("catalog not found")
		}

		return nil
	}
}

func AccCheckCatalogDestroy(s *terraform.State) error {
	config := testhelpers.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_datacatalog_catalog" {
			continue
		}
		id := rs.Primary.ID

		reqApi := &datacatalog.GetCatalogRequest{
			CatalogId: id,
		}
		md := new(metadata.MD)

		_, err := datacatalogv1sdk.NewCatalogClient(config.SDKv2).GetCatalog(context.Background(), reqApi, grpc.Header(md))

		if err == nil {
			return fmt.Errorf("catalog still exists")
		}
	}

	return nil
}
