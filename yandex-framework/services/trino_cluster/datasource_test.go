package trino_cluster_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func TestAccDataSourceMDBTrinoCluster_basic(t *testing.T) {
	t.Parallel()

	randSuffix := fmt.Sprintf("%d", acctest.RandInt())
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testhelpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: testhelpers.AccProviderFactories,
		CheckDestroy:             testAccCheckTrinoClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: trinoDatasourceClusterConfig(t, randSuffix, true),
				Check:  datasourceTestCheckComposeFunc(randSuffix),
			},
			{
				Config: trinoDatasourceClusterConfig(t, randSuffix, false),
				Check:  datasourceTestCheckComposeFunc(randSuffix),
			},
		},
	})
}

func trinoDatasourceClusterConfig(t *testing.T, randSuffix string, byID bool) string {
	resource := trinoClusterConfig(t, trinoClusterConfigParams{
		RandSuffix: randSuffix,
		FolderID:   os.Getenv("YC_FOLDER_ID"),
		Coordinator: trinoComponentParams{
			ResourcePresetID: "c8-m32",
		},
		Worker: trinoWorkerParams{
			ResourcePresetID: "c4-m16",
			FixedScale: &FixedScaleParams{
				Count: 2,
			},
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
		RetryPolicy: &RetryPolicyParams{
			Policy: "TASK",
			AdditionalProperties: map[string]string{
				"fault-tolerant-execution-max-task-split-count": "1024",
			},
			ExchangeManager: ExchangeManagerParams{},
		},
	},
	)

	var datasource string
	if byID {
		datasource = `
 data "yandex_trino_cluster" "trino_cluster" {
   id = yandex_trino_cluster.trino_cluster.id
 }`
	} else {
		datasource = `
 data "yandex_trino_cluster" "trino_cluster" {
   name = yandex_trino_cluster.trino_cluster.name
 }`
	}

	return fmt.Sprintf("%s\n%s", resource, datasource)
}

func datasourceTestCheckComposeFunc(randSuffix string) resource.TestCheckFunc {
	folderID := os.Getenv("YC_FOLDER_ID")
	return resource.ComposeTestCheckFunc(
		testAccCheckTrinoExists("yandex_trino_cluster.trino_cluster", nil),
		resource.TestCheckResourceAttrSet("data.yandex_trino_cluster.trino_cluster", "service_account_id"),
		resource.TestCheckResourceAttrSet("data.yandex_trino_cluster.trino_cluster", "subnet_ids.0"),
		resource.TestCheckResourceAttr("data.yandex_trino_cluster.trino_cluster", "name", fmt.Sprintf("trino-%s", randSuffix)),
		resource.TestCheckResourceAttr("data.yandex_trino_cluster.trino_cluster", "folder_id", folderID),
		resource.TestCheckResourceAttr("data.yandex_trino_cluster.trino_cluster", "coordinator.resource_preset_id", "c8-m32"),
		resource.TestCheckResourceAttr("data.yandex_trino_cluster.trino_cluster", "worker.resource_preset_id", "c4-m16"),
		resource.TestCheckResourceAttr("data.yandex_trino_cluster.trino_cluster", "worker.fixed_scale.count", "2"),
		resource.TestCheckResourceAttr("data.yandex_trino_cluster.trino_cluster", "labels.label", "value"),
		resource.TestCheckResourceAttr("data.yandex_trino_cluster.trino_cluster", "deletion_protection", "false"),
		resource.TestCheckResourceAttr("data.yandex_trino_cluster.trino_cluster", "maintenance_window.type", "WEEKLY"),
		resource.TestCheckResourceAttr("data.yandex_trino_cluster.trino_cluster", "maintenance_window.day", "MON"),
		resource.TestCheckResourceAttr("data.yandex_trino_cluster.trino_cluster", "maintenance_window.hour", "2"),

		// Additional parameters
		resource.TestCheckResourceAttrSet("data.yandex_trino_cluster.trino_cluster", "security_group_ids.0"),
		resource.TestCheckResourceAttr("data.yandex_trino_cluster.trino_cluster", "description", "trino-cluster"),
		resource.TestCheckResourceAttr("data.yandex_trino_cluster.trino_cluster", "logging.enabled", "true"),
		resource.TestCheckResourceAttr("data.yandex_trino_cluster.trino_cluster", "logging.folder_id", folderID),
		resource.TestCheckResourceAttr("data.yandex_trino_cluster.trino_cluster", "logging.min_level", "INFO"),
		resource.TestCheckResourceAttr("data.yandex_trino_cluster.trino_cluster", "retry_policy.policy", "TASK"),
	)
}
