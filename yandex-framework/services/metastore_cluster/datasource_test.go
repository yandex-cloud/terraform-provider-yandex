package metastore_cluster_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func TestAccDataSourceMDBMetastoreCluster_basic(t *testing.T) {
	t.Parallel()

	randSuffix := fmt.Sprintf("%d", acctest.RandInt())
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testhelpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: testhelpers.AccProviderFactories,
		CheckDestroy:             testAccCheckMetastoreClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: metastoreDatasourceClusterConfig(t, randSuffix, true),
				Check:  datasourceTestCheckComposeFunc(randSuffix),
			},
			{
				Config: metastoreDatasourceClusterConfig(t, randSuffix, false),
				Check:  datasourceTestCheckComposeFunc(randSuffix),
			},
		},
	})
}

func metastoreDatasourceClusterConfig(t *testing.T, randSuffix string, byID bool) string {
	resource := metastoreClusterConfig(t, metastoreClusterConfigParams{
		RandSuffix:        randSuffix,
		FolderID:          os.Getenv("YC_FOLDER_ID"),
		FolderIDSpecified: true,
		Labels: map[string]string{
			"label": "value",
		},
		MaintenanceWindow: &MaintenanceWindow{
			Type: "WEEKLY",
			Day:  "MON",
			Hour: 2,
		},
		DeletionProtection: newOptional(false),
		LoggingEnabled:     newOptional(true),
		Description:        newOptional("description"),
		SGIDsSpecified:     newOptional(true),
		SubnetIDVar:        "yandex_vpc_subnet.metastore-a.id",
		ResourcePreset:     "c2-m4",
	})

	var datasource string
	if byID {
		datasource = `
data "yandex_metastore_cluster" "metastore_cluster" {
  id = yandex_metastore_cluster.metastore_cluster.id
}`
	} else {
		datasource = `
data "yandex_metastore_cluster" "metastore_cluster" {
  name = yandex_metastore_cluster.metastore_cluster.name
}`
	}

	return fmt.Sprintf("%s\n%s", resource, datasource)
}

func datasourceTestCheckComposeFunc(randSuffix string) resource.TestCheckFunc {
	folderID := os.Getenv("YC_FOLDER_ID")
	return resource.ComposeTestCheckFunc(
		testAccCheckMetastoreExists("data.yandex_metastore_cluster.metastore_cluster", nil),
		resource.TestCheckResourceAttr("data.yandex_metastore_cluster.metastore_cluster", "name", fmt.Sprintf("metastore-%s", randSuffix)),
		resource.TestCheckResourceAttrSet("data.yandex_metastore_cluster.metastore_cluster", "service_account_id"),
		resource.TestCheckResourceAttrSet("data.yandex_metastore_cluster.metastore_cluster", "subnet_ids.0"),
		resource.TestCheckResourceAttr("data.yandex_metastore_cluster.metastore_cluster", "cluster_config.resource_preset_id", "c2-m4"),
		resource.TestCheckResourceAttr("data.yandex_metastore_cluster.metastore_cluster", "deletion_protection", "false"),
		resource.TestCheckResourceAttr("data.yandex_metastore_cluster.metastore_cluster", "maintenance_window.type", "WEEKLY"),
		resource.TestCheckResourceAttr("data.yandex_metastore_cluster.metastore_cluster", "maintenance_window.day", "MON"),
		resource.TestCheckResourceAttr("data.yandex_metastore_cluster.metastore_cluster", "maintenance_window.hour", "2"),
		resource.TestCheckResourceAttr("data.yandex_metastore_cluster.metastore_cluster", "description", "description"),
		resource.TestCheckResourceAttrSet("data.yandex_metastore_cluster.metastore_cluster", "security_group_ids.0"),
		resource.TestCheckResourceAttr("data.yandex_metastore_cluster.metastore_cluster", "folder_id", folderID),
		resource.TestCheckResourceAttr("data.yandex_metastore_cluster.metastore_cluster", "logging.enabled", "true"),
		resource.TestCheckResourceAttr("data.yandex_metastore_cluster.metastore_cluster", "logging.min_level", "INFO"),
		resource.TestCheckResourceAttr("data.yandex_metastore_cluster.metastore_cluster", "logging.folder_id", folderID),
	)
}
