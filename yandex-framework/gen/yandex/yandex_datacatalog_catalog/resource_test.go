package yandex_datacatalog_catalog_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/datacatalog/v1"
	datacatalogv1sdk "github.com/yandex-cloud/go-sdk/services/datacatalog/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

var testCatalogResourceName = "yandex_datacatalog_catalog.test-catalog"

func init() {
	resource.AddTestSweepers("yandex_datacatalog_catalog", &resource.Sweeper{
		Name:         "yandex_datacatalog_catalog",
		F:            testSweepCatalog,
		Dependencies: []string{},
	})
}

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func testSweepCatalog(_ string) error {
	conf, err := testhelpers.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	resp, err := datacatalogv1sdk.NewCatalogClient(conf.SDKv2).ListCatalogs(context.Background(), &datacatalog.ListCatalogsRequest{
		FolderId: conf.ProviderState.FolderID.ValueString(),
	})
	if err != nil {
		return fmt.Errorf("error getting Datacatalog catalog: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.Catalogs {
		if !sweepCatalog(conf, c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Datacatalog catalog %q", c.Id))
		}
	}

	return result.ErrorOrNil()
}

func sweepCatalog(conf *provider_config.Config, id string) bool {
	return testhelpers.SweepWithRetry(sweepCatalogOnce, conf, "yandex_datacatalog_catalog", id)
}

func sweepCatalogOnce(conf *provider_config.Config, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	op, err := datacatalogv1sdk.NewCatalogClient(conf.SDKv2).DeleteCatalog(ctx, &datacatalog.DeleteCatalogRequest{
		CatalogId: id,
	})
	_, err = op.Wait(context.Background())
	return err
}

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
