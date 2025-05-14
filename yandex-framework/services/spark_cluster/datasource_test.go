package spark_cluster_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func TestAccDataSourceSparkCluster_basic(t *testing.T) {
	t.Parallel()

	randSuffix := fmt.Sprintf("%d", acctest.RandInt())
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testhelpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: testhelpers.AccProviderFactories,
		CheckDestroy:             testAccCheckSparkClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: sparkDatasourceClusterConfig(t, randSuffix, true),
				Check:  datasourceTestCheckComposeFunc(randSuffix),
			},
			{
				Config: sparkDatasourceClusterConfig(t, randSuffix, false),
				Check:  datasourceTestCheckComposeFunc(randSuffix),
			},
		},
	})
}

func sparkDatasourceClusterConfig(t *testing.T, randSuffix string, byID bool) string {
	resource := sparkClusterConfig(t, sparkClusterConfigParams{
		RandSuffix:         randSuffix,
		Description:        "created with terraform",
		IncludeBlockLabels: true,
		Labels: map[string]string{
			"my_label": "my_value",
		},
		DriverResourcePresetID:    "c2-m8",
		DriverSize:                1,
		ExecutorResourcePresetID:  "c4-m16",
		ExecutorMinSize:           1,
		ExecutorMaxSize:           2,
		IncludeBlockDependencies:  true,
		PipPackage:                "numpy==2.2.2",
		DebPackage:                "git",
		IncludeBlockHistoryServer: true,
		HistoryServerEnabled:      true,
		IncludeBlockMetastore:     true,
		MetastoreClusterID:        os.Getenv("YC_METASTORE_ID"),
		SecurityGroup:             true,
		DeletionProtection:        false,
		LoggingFolderID:           os.Getenv("YC_FOLDER_ID"),
		IncludeBlockMaintenance:   true,
		MaintenanceWindowWeekly:   true,
		MaintenanceWindowDay:      "TUE",
		MaintenanceWindowHour:     10,
	})

	var datasource string
	if byID {
		datasource = `
data "yandex_spark_cluster" "spark_cluster" {
  id = yandex_spark_cluster.spark_cluster.id
}`
	} else {
		datasource = `
data "yandex_spark_cluster" "spark_cluster" {
  name = yandex_spark_cluster.spark_cluster.name
}`
	}

	return fmt.Sprintf("%s\n%s", resource, datasource)
}

func datasourceTestCheckComposeFunc(randSuffix string) resource.TestCheckFunc {
	folderID := os.Getenv("YC_FOLDER_ID")
	metastoreClusterID := os.Getenv("YC_METASTORE_ID")

	return resource.ComposeTestCheckFunc(
		testAccCheckSparkExists("yandex_spark_cluster.spark_cluster", nil),
		resource.TestCheckResourceAttrSet("data.yandex_spark_cluster.spark_cluster", "service_account_id"),
		resource.TestCheckResourceAttrSet("data.yandex_spark_cluster.spark_cluster", "network.subnet_ids.0"),
		resource.TestCheckResourceAttrSet("data.yandex_spark_cluster.spark_cluster", "network.security_group_ids.0"),
		resource.TestCheckResourceAttr("yandex_spark_cluster.spark_cluster", "name", fmt.Sprintf("spark-%s", randSuffix)),
		resource.TestCheckResourceAttr("data.yandex_spark_cluster.spark_cluster", "description", "created with terraform"),
		resource.TestCheckResourceAttr("data.yandex_spark_cluster.spark_cluster", "labels.my_label", "my_value"),
		resource.TestCheckResourceAttr("data.yandex_spark_cluster.spark_cluster", "deletion_protection", "false"),
		resource.TestCheckResourceAttr("data.yandex_spark_cluster.spark_cluster", "config.resource_pools.driver.resource_preset_id", "c2-m8"),
		resource.TestCheckResourceAttr("data.yandex_spark_cluster.spark_cluster", "config.resource_pools.driver.size", "1"),
		resource.TestCheckResourceAttr("data.yandex_spark_cluster.spark_cluster", "config.resource_pools.executor.resource_preset_id", "c4-m16"),
		resource.TestCheckResourceAttr("data.yandex_spark_cluster.spark_cluster", "config.resource_pools.executor.min_size", "1"),
		resource.TestCheckResourceAttr("data.yandex_spark_cluster.spark_cluster", "config.resource_pools.executor.max_size", "2"),
		resource.TestCheckResourceAttr("data.yandex_spark_cluster.spark_cluster", "config.dependencies.pip_packages.0", "numpy==2.2.2"),
		resource.TestCheckResourceAttr("data.yandex_spark_cluster.spark_cluster", "config.dependencies.deb_packages.0", "git"),
		resource.TestCheckResourceAttr("data.yandex_spark_cluster.spark_cluster", "config.history_server.enabled", "true"),
		resource.TestCheckResourceAttr("data.yandex_spark_cluster.spark_cluster", "config.metastore.cluster_id", metastoreClusterID),
		resource.TestCheckResourceAttr("data.yandex_spark_cluster.spark_cluster", "logging.enabled", "true"),
		resource.TestCheckResourceAttr("data.yandex_spark_cluster.spark_cluster", "logging.folder_id", folderID),
		resource.TestCheckResourceAttr("data.yandex_spark_cluster.spark_cluster", "maintenance_window.type", "WEEKLY"),
		resource.TestCheckResourceAttr("data.yandex_spark_cluster.spark_cluster", "maintenance_window.day", "TUE"),
		resource.TestCheckResourceAttr("data.yandex_spark_cluster.spark_cluster", "maintenance_window.hour", "10"),
	)
}
