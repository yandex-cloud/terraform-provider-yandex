package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	backuppb "github.com/yandex-cloud/go-genproto/yandex/cloud/backup/v1"
)

func init() {
	resource.AddTestSweepers("yandex_backup_policy", &resource.Sweeper{
		Name:         "yandex_backup_policy",
		F:            testSweepBackupPolicy,
		Dependencies: []string{},
	})
}

func TestAccResourceBackupPolicyBasic(t *testing.T) {
	policyName := fmt.Sprintf("tf-test-backup-policy-basic-%s", acctest.RandString(10))
	folderID := getExampleFolderID()

	config, resourceName := testAccBackupPolicyBasicConfig(policyName)

	var policy backuppb.Policy
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBackupPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testBackupPolicyExists(resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "name", policyName),
					resource.TestCheckResourceAttr(resourceName, "folder_id", folderID),
					resource.TestCheckResourceAttr(resourceName, "name", policyName),
					resource.TestCheckResourceAttr(resourceName, "splitting_bytes", "9223372036854775807"),
					resource.TestCheckResourceAttr(resourceName, "scheduling.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "scheduling.0.backup_sets.0.execute_by_interval", "1024"),
					resource.TestCheckResourceAttr(resourceName, "retention.0.after_backup", "false"),
					resource.TestCheckResourceAttr(resourceName, "retention.0.rules.0.max_count", "10"),
					resource.TestCheckResourceAttr(resourceName, "reattempts.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "reattempts.0.interval", "5m"),
					resource.TestCheckResourceAttr(resourceName, "reattempts.0.max_attempts", "5"),
					resource.TestCheckResourceAttr(resourceName, "vm_snapshot_reattempts.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "vm_snapshot_reattempts.0.interval", "5m"),
					resource.TestCheckResourceAttr(resourceName, "vm_snapshot_reattempts.0.max_attempts", "5"),
					// Default values
					resource.TestCheckResourceAttr(resourceName, "archive_name", "[Machine Name]-[Plan ID]-[Unique ID]a"),
					resource.TestCheckResourceAttr(resourceName, "cbt", "DO_NOT_USE"),
					resource.TestCheckResourceAttr(resourceName, "compression", "NORMAL"),
					resource.TestCheckResourceAttr(resourceName, "fast_backup_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "format", "AUTO"),
					resource.TestCheckResourceAttr(resourceName, "multi_volume_snapshotting_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "performance_window_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "preserve_file_security_settings", "true"),
				),
			},
			{
				ResourceName: resourceName,
				ImportStateIdFunc: func(*terraform.State) (string, error) {
					return policy.GetId(), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"created_at",
					"updated_at",
					"enabled",
				},
			},
		},
	})
}

func TestAccResourceBackupPolicyMultipleBackupSets(t *testing.T) {
	policyName := fmt.Sprintf("tf-test-backup-policy-multiple-backup-sets-%s", acctest.RandString(10))
	folderID := getExampleFolderID()

	config, resourceName := testAccBackupPolicyMultipleBackupSetsConfig(policyName)

	var policy backuppb.Policy
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckBackupPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testBackupPolicyExists(resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "name", policyName),
					resource.TestCheckResourceAttr(resourceName, "folder_id", folderID),
					resource.TestCheckResourceAttr(resourceName, "name", policyName),
					resource.TestCheckResourceAttr(resourceName, "splitting_bytes", "9223372036854775807"),
					resource.TestCheckResourceAttr(resourceName, "scheduling.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "scheduling.0.backup_sets.0.execute_by_interval", "1000"),
					resource.TestCheckResourceAttr(resourceName, "scheduling.0.backup_sets.0.type", "TYPE_INCREMENTAL"),
					resource.TestCheckResourceAttr(resourceName, "scheduling.0.backup_sets.1.execute_by_interval", "2000"),
					resource.TestCheckResourceAttr(resourceName, "scheduling.0.backup_sets.1.type", "TYPE_FULL"),
					resource.TestCheckResourceAttr(resourceName, "scheduling.0.scheme", "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "retention.0.after_backup", "false"),
					resource.TestCheckResourceAttr(resourceName, "retention.0.rules.0.max_count", "10"),
					resource.TestCheckResourceAttr(resourceName, "reattempts.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "reattempts.0.interval", "5m"),
					resource.TestCheckResourceAttr(resourceName, "reattempts.0.max_attempts", "5"),
					resource.TestCheckResourceAttr(resourceName, "vm_snapshot_reattempts.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "vm_snapshot_reattempts.0.interval", "5m"),
					resource.TestCheckResourceAttr(resourceName, "vm_snapshot_reattempts.0.max_attempts", "5"),
					// Default values
					resource.TestCheckResourceAttr(resourceName, "archive_name", "[Machine Name]-[Plan ID]-[Unique ID]a"),
					resource.TestCheckResourceAttr(resourceName, "cbt", "DO_NOT_USE"),
					resource.TestCheckResourceAttr(resourceName, "compression", "NORMAL"),
					resource.TestCheckResourceAttr(resourceName, "fast_backup_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "format", "AUTO"),
					resource.TestCheckResourceAttr(resourceName, "multi_volume_snapshotting_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "performance_window_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "preserve_file_security_settings", "true"),
				),
			},
			{
				ResourceName: resourceName,
				ImportStateIdFunc: func(*terraform.State) (string, error) {
					return policy.GetId(), nil
				},
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"created_at",
					"updated_at",
					"enabled",
				},
			},
		},
	})
}

func testAccBackupPolicyBasicConfig(policyName string) (config, outResourceName string) {
	const (
		resourceType = "yandex_backup_policy"
		resourceName = "test_policy"
		template     = `resource %q %q {
    name = %q
    splitting_bytes = "9223372036854775807"
    scheduling {
      enabled               = true
	  backup_sets {
        execute_by_interval = 1024
      }
    }

    retention {
      after_backup = false
	  rules {
		max_count  = 10
	  }
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
  }`
	)

	outResourceName = resourceType + "." + resourceName

	return fmt.Sprintf(template, resourceType, resourceName, policyName), outResourceName
}

func testAccBackupPolicyMultipleBackupSetsConfig(policyName string) (config, outResourceName string) {
	const (
		resourceType = "yandex_backup_policy"
		resourceName = "test_policy"
		template     = `resource %q %q {
    name = %q
    splitting_bytes = "9223372036854775807"
    scheduling {
      enabled               = true
      scheme 			    = "CUSTOM"
	  backup_sets {
		execute_by_interval = 1000
		type = "TYPE_INCREMENTAL"
	  }

	  backup_sets {
		execute_by_interval = 2000
		type = "TYPE_FULL"
	  }
    }

    retention {
      after_backup = false
	  rules {
		max_count  = 10
	  }
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
  }`
	)

	outResourceName = resourceType + "." + resourceName

	return fmt.Sprintf(template, resourceType, resourceName, policyName), outResourceName
}

func sweepBackupPolicy(conf *Config, id string) bool {
	return sweepWithRetryByFunc(conf, "Backup Policy", func(conf *Config) error {
		ctx, cancel := conf.ContextWithTimeout(yandexBackupDefaultTimeout)
		defer cancel()

		op, err := conf.sdk.Backup().Policy().Delete(ctx, &backuppb.DeletePolicyRequest{
			PolicyId: id,
		})

		return handleSweepOperation(ctx, conf, op, err)
	})
}

func testSweepBackupPolicy(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &backuppb.ListPoliciesRequest{FolderId: conf.FolderID}
	it := conf.sdk.Backup().Policy().PolicyIterator(conf.Context(), req)
	result := &multierror.Error{}

	for it.Next() {
		id := it.Value().GetId()
		if !sweepBackupPolicy(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Backup Policy %q", id))
		}
	}

	return result.ErrorOrNil()
}

func testAccCheckBackupPolicyDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_backup_policy" {
			continue
		}

		id := rs.Primary.ID
		_, err := config.sdk.Backup().Policy().Get(config.Context(), &backuppb.GetPolicyRequest{
			PolicyId: id,
		})

		if err == nil {
			return fmt.Errorf("backup policy %s still exists", id)
		}
	}

	return nil
}

func testBackupPolicyExists(resourceName string, policy *backuppb.Policy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("backup policy not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		config := testAccProvider.Meta().(*Config)
		id := rs.Primary.ID

		found, err := config.sdk.Backup().Policy().Get(context.Background(), &backuppb.GetPolicyRequest{
			PolicyId: id,
		})

		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("backup policy is not found")
		}

		//goland:noinspection GoVetCopyLock
		*policy = *found

		return nil
	}
}
