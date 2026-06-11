package datalens_dataset_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func TestAccDatalensDataset_dataSource_basic(t *testing.T) {
	t.Parallel()

	saName := test.ResourceName(63)
	dbName := test.ResourceName(63)
	connName := test.ResourceName(63)
	datasetName := test.ResourceName(63)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckDatalensDatasetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatalensDatasetConfig_basic(saName, dbName, connName, datasetName) + fmt.Sprintf(`
data "yandex_datalens_dataset" "test" {
  id              = yandex_datalens_dataset.test.id
  organization_id = "%s"
}`, test.GetExampleOrganizationID()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dlDatasetDataSource, "id", dlDatasetResource, "id"),
					resource.TestCheckResourceAttrPair(dlDatasetDataSource, "name", dlDatasetResource, "name"),
				),
			},
		},
	})
}
