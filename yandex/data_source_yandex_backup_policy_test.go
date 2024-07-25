package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceBackupPolicy(t *testing.T) {
	policyName := fmt.Sprintf("tf-test-backup-policy-basic-%s", acctest.RandString(10))
	folderID := getExampleFolderID()

	config, dataSourceName := testAccBackupPolicyDataSourceBasicConfig(policyName)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBackupPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceIDField(dataSourceName, "policy_id"),
					resource.TestCheckResourceAttr(dataSourceName, "name", policyName),
					resource.TestCheckResourceAttr(dataSourceName, "folder_id", folderID),
					resource.TestCheckResourceAttr(dataSourceName, "splitting_bytes", "9223372036854775807"),
					resource.TestCheckResourceAttr(dataSourceName, "scheduling.0.enabled", "true"),
					resource.TestCheckResourceAttr(dataSourceName, "scheduling.0.backup_sets.0.execute_by_interval", "86400"),
					resource.TestCheckResourceAttr(dataSourceName, "scheduling.0.backup_sets.0.type", "TYPE_AUTO"),
					resource.TestCheckResourceAttr(dataSourceName, "retention.0.after_backup", "true"),
					resource.TestCheckResourceAttr(dataSourceName, "reattempts.0.enabled", "true"),
					resource.TestCheckResourceAttr(dataSourceName, "reattempts.0.interval", "5m"),
					resource.TestCheckResourceAttr(dataSourceName, "reattempts.0.max_attempts", "5"),
					resource.TestCheckResourceAttr(dataSourceName, "vm_snapshot_reattempts.0.enabled", "true"),
					resource.TestCheckResourceAttr(dataSourceName, "vm_snapshot_reattempts.0.interval", "5m"),
					resource.TestCheckResourceAttr(dataSourceName, "vm_snapshot_reattempts.0.max_attempts", "5"),
					// Default values
					resource.TestCheckResourceAttr(dataSourceName, "archive_name", "[Machine Name]-[Plan ID]-[Unique ID]a"),
					resource.TestCheckResourceAttr(dataSourceName, "cbt", "DO_NOT_USE"),
					resource.TestCheckResourceAttr(dataSourceName, "compression", "NORMAL"),
					resource.TestCheckResourceAttr(dataSourceName, "fast_backup_enabled", "true"),
					resource.TestCheckResourceAttr(dataSourceName, "format", "AUTO"),
					resource.TestCheckResourceAttr(dataSourceName, "multi_volume_snapshotting_enabled", "true"),
					resource.TestCheckResourceAttr(dataSourceName, "performance_window_enabled", "false"),
					resource.TestCheckResourceAttr(dataSourceName, "preserve_file_security_settings", "true"),
				),
			},
		},
	})
}

func testAccBackupPolicyDataSourceBasicConfig(policyName string) (config, outDataSourceName string) {
	const template = `
resource "yandex_backup_policy" "test_policy" {
  name            = %q
  splitting_bytes = "9223372036854775807"
  scheduling {
    enabled             = true
	backup_sets {
      execute_by_interval = 86400
	}
  }

  retention {
    after_backup = true
  }

  reattempts {
    enabled      = true
    interval     = "5m"
    max_attempts = 5
  }

  vm_snapshot_reattempts {
    enabled      = true
    interval     = "5m"
    max_attempts = 5
  }
}

data "yandex_backup_policy" "test_policy_ds" {
  name = yandex_backup_policy.test_policy.name
}
`

	return fmt.Sprintf(template, policyName), "data.yandex_backup_policy.test_policy_ds"
}
