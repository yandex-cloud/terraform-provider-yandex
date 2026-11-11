package trino_access_control_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func TestAccDataSourceMDBTrinoAccessControl_basic(t *testing.T) {
	t.Parallel()

	randSuffix := fmt.Sprintf("%d", acctest.RandInt())
	folderID := os.Getenv("YC_FOLDER_ID")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testhelpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: testhelpers.AccProviderFactories,
		CheckDestroy:             testAccCheckTrinoAccessControlDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceTrinoAccessControlConfig(t, trinoAccessControlConfigParams{
					RandSuffix: randSuffix,
					FolderID:   folderID,
					CatalogRules: []CatalogRule{
						{
							Description: "updated rule",
							Permission:  "READ_ONLY",
						},
					},
				}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.yandex_trino_access_control.trino_access_control", "catalogs.#", "1"),
					resource.TestCheckResourceAttr("data.yandex_trino_access_control.trino_access_control", "catalogs.0.description", "updated rule"),
					resource.TestCheckResourceAttr("data.yandex_trino_access_control.trino_access_control", "catalogs.0.permission", "READ_ONLY"),

					resource.TestCheckNoResourceAttr("data.yandex_trino_access_control.trino_access_control", "schemas"),
					resource.TestCheckNoResourceAttr("data.yandex_trino_access_control.trino_access_control", "tables"),
					resource.TestCheckNoResourceAttr("data.yandex_trino_access_control.trino_access_control", "functions"),
					resource.TestCheckNoResourceAttr("data.yandex_trino_access_control.trino_access_control", "procedures"),
					resource.TestCheckNoResourceAttr("data.yandex_trino_access_control.trino_access_control", "queries"),
					resource.TestCheckNoResourceAttr("data.yandex_trino_access_control.trino_access_control", "system_session_properties"),
					resource.TestCheckNoResourceAttr("data.yandex_trino_access_control.trino_access_control", "catalog_session_properties"),
				),
			},
		},
	})
}

func testAccDataSourceTrinoAccessControlConfig(t *testing.T, params trinoAccessControlConfigParams) string {
	accessControlConfig := trinoAccessControlConfig(t, params)
	dataSourceConfig := `
data "yandex_trino_access_control" "trino_access_control" {
  cluster_id = yandex_trino_access_control.trino_access_control.cluster_id
}`
	return fmt.Sprintf("%s\n%s", accessControlConfig, dataSourceConfig)
}
