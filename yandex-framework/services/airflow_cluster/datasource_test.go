package airflow_cluster_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func TestAccDataSourceMDBAirflowCluster_basic(t *testing.T) {
	t.Parallel()

	randSuffix := fmt.Sprintf("%d", acctest.RandInt())
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testhelpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: testhelpers.AccProviderFactories,
		CheckDestroy:             testAccCheckAirflowClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: airflowDatasourceClusterConfig(t, randSuffix, true),
				Check:  datasourceTestCheckComposeFunc(randSuffix),
			},
			{
				Config: airflowDatasourceClusterConfig(t, randSuffix, false),
				Check:  datasourceTestCheckComposeFunc(randSuffix),
			},
		},
	})
}

func airflowDatasourceClusterConfig(t *testing.T, randSuffix string, byID bool) string {
	resource := airflowClusterConfig(t, airflowClusterConfigParams{
		RandSuffix: randSuffix,
		FolderID:   os.Getenv("YC_FOLDER_ID"),
		Webserver: airflowComponentParams{
			Count:            1,
			ResourcePresetID: "c1-m2",
		},
		Scheduler: airflowComponentParams{
			Count:            1,
			ResourcePresetID: "c1-m2",
		},
		Worker: airflowWorkerParams{
			MinCount:         1,
			MaxCount:         1,
			ResourcePresetID: "c1-m2",
		},
		Triggerer: &airflowComponentParams{
			Count:            1,
			ResourcePresetID: "c1-m2",
		},
		Labels: map[string]string{
			"label": "value",
		},
		MaintenanceWindow: &MaintenanceWindow{
			Type: "WEEKLY",
			Day:  "MON",
			Hour: 2,
		},
		AdditionalParams: true,
	})

	var datasource string
	if byID {
		datasource = `
data "yandex_airflow_cluster" "airflow_cluster" {
  id = yandex_airflow_cluster.airflow_cluster.id
}`
	} else {
		datasource = `
data "yandex_airflow_cluster" "airflow_cluster" {
  name = yandex_airflow_cluster.airflow_cluster.name
}`
	}

	return fmt.Sprintf("%s\n%s", resource, datasource)
}

func datasourceTestCheckComposeFunc(randSuffix string) resource.TestCheckFunc {
	folderID := os.Getenv("YC_FOLDER_ID")
	return resource.ComposeTestCheckFunc(
		testAccCheckAirflowExists("yandex_airflow_cluster.airflow_cluster", nil),
		resource.TestCheckResourceAttrSet("data.yandex_airflow_cluster.airflow_cluster", "service_account_id"),
		resource.TestCheckResourceAttrSet("data.yandex_airflow_cluster.airflow_cluster", "subnet_ids.0"),
		resource.TestCheckResourceAttrSet("data.yandex_airflow_cluster.airflow_cluster", "subnet_ids.1"),
		resource.TestCheckResourceAttrSet("data.yandex_airflow_cluster.airflow_cluster", "subnet_ids.2"),
		resource.TestCheckResourceAttrSet("yandex_airflow_cluster.airflow_cluster", "airflow_version"),
		resource.TestCheckResourceAttrSet("yandex_airflow_cluster.airflow_cluster", "python_version"),
		resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "code_sync.s3.bucket", fmt.Sprintf("airflow-tf-%s", randSuffix)),
		resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "name", fmt.Sprintf("airflow-%s", randSuffix)),
		resource.TestCheckResourceAttr("data.yandex_airflow_cluster.airflow_cluster", "folder_id", folderID),
		resource.TestCheckResourceAttr("data.yandex_airflow_cluster.airflow_cluster", "webserver.count", "1"),
		resource.TestCheckResourceAttr("data.yandex_airflow_cluster.airflow_cluster", "webserver.resource_preset_id", "c1-m2"),
		resource.TestCheckResourceAttr("data.yandex_airflow_cluster.airflow_cluster", "scheduler.count", "1"),
		resource.TestCheckResourceAttr("data.yandex_airflow_cluster.airflow_cluster", "scheduler.resource_preset_id", "c1-m2"),
		resource.TestCheckResourceAttr("data.yandex_airflow_cluster.airflow_cluster", "worker.min_count", "1"),
		resource.TestCheckResourceAttr("data.yandex_airflow_cluster.airflow_cluster", "worker.max_count", "1"),
		resource.TestCheckResourceAttr("data.yandex_airflow_cluster.airflow_cluster", "worker.resource_preset_id", "c1-m2"),
		resource.TestCheckResourceAttr("data.yandex_airflow_cluster.airflow_cluster", "triggerer.count", "1"),
		resource.TestCheckResourceAttr("data.yandex_airflow_cluster.airflow_cluster", "triggerer.resource_preset_id", "c1-m2"),
		resource.TestCheckResourceAttr("data.yandex_airflow_cluster.airflow_cluster", "labels.label", "value"),
		resource.TestCheckResourceAttr("data.yandex_airflow_cluster.airflow_cluster", "deletion_protection", "false"),

		// Additional
		resource.TestCheckResourceAttrSet("data.yandex_airflow_cluster.airflow_cluster", "airflow_config.api.auth_backends"),
		resource.TestCheckResourceAttrSet("data.yandex_airflow_cluster.airflow_cluster", "security_group_ids.0"),
		resource.TestCheckResourceAttr("data.yandex_airflow_cluster.airflow_cluster", "pip_packages.0", "dbt"),
		resource.TestCheckResourceAttr("data.yandex_airflow_cluster.airflow_cluster", "deb_packages.0", "tree"),
		resource.TestCheckResourceAttr("data.yandex_airflow_cluster.airflow_cluster", "description", "airflow-cluster"),
		resource.TestCheckResourceAttr("data.yandex_airflow_cluster.airflow_cluster", "lockbox_secrets_backend.enabled", "true"),
		resource.TestCheckResourceAttr("data.yandex_airflow_cluster.airflow_cluster", "logging.enabled", "true"),
		resource.TestCheckResourceAttr("data.yandex_airflow_cluster.airflow_cluster", "logging.folder_id", folderID),
		resource.TestCheckResourceAttr("data.yandex_airflow_cluster.airflow_cluster", "logging.min_level", "INFO"),
		resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "maintenance_window.type", "WEEKLY"),
		resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "maintenance_window.day", "MON"),
		resource.TestCheckResourceAttr("yandex_airflow_cluster.airflow_cluster", "maintenance_window.hour", "2"),
	)
}
