package mdb_greenplum_cluster_v2_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/go-multierror"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/greenplum/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

const gpVPCDependencies = `
resource "yandex_vpc_network" "mdb-gp-test-net" {}

resource "yandex_vpc_subnet" "mdb-gp-test-subnet-a" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.mdb-gp-test-net.id
  v4_cidr_blocks = ["10.1.0.0/24"]
}
`

const (
	defaultMDBPageSize            = 1000
	greenplumClusterDeleteTimeout = 15 * time.Minute
)

func init() {
	resource.AddTestSweepers("yandex_mdb_greenplum_cluster_v2", &resource.Sweeper{
		Name: "yandex_mdb_greenplum_cluster_v2",
		F:    testSweepMDBPostgreSQLCluster,
	})
}

func testSweepMDBPostgreSQLCluster(_ string) error {
	conf, err := testhelpers.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := conf.SDK.MDB().Greenplum().Cluster().List(context.Background(), &greenplum.ListClustersRequest{
		FolderId:  conf.ProviderState.FolderID.ValueString(),
		PageSize:  defaultMDBPageSize,
		PageToken: "",
		Filter:    "",
	})
	if err != nil {
		return fmt.Errorf("error getting PostgreSQL clusters: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.Clusters {
		if !testhelpers.SweepWithRetry(sweepMDBGreenplumCluster, conf, "Greenplum cluster", c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep PostgreSQL cluster %q", c.Id))
		}
	}

	return result.ErrorOrNil()
}

func sweepMDBGreenplumCluster(conf *config.Config, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), greenplumClusterDeleteTimeout)
	defer cancel()

	mask := field_mask.FieldMask{Paths: []string{"deletion_protection"}}

	op, err := conf.SDK.MDB().Greenplum().Cluster().Update(ctx, &greenplum.UpdateClusterRequest{
		ClusterId:          id,
		DeletionProtection: false,
		UpdateMask:         &mask,
	})
	err = testhelpers.HandleSweepOperation(ctx, conf, op, err)
	if err != nil && !strings.EqualFold(testhelpers.ErrorMessage(err), "no changes detected") {
		return err
	}

	op, err = conf.SDK.MDB().Greenplum().Cluster().Delete(ctx, &greenplum.DeleteClusterRequest{
		ClusterId: id,
	})
	return testhelpers.HandleSweepOperation(ctx, conf, op, err)
}

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestAccResourceYandexMdbGreenplumClusterV2_full(t *testing.T) {
	t.Parallel()

	clusterName := acctest.RandomWithPrefix("tf-greenplum-cluster")
	clusterDesc := "Test Greenplum Cluster"
	updatedDesc := "Updated Test Greenplum Cluster"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testhelpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: testhelpers.AccProviderFactories,
		CheckDestroy:             testAccCheckYandexMdbGreenplumClusterV2Destroy,
		Steps: []resource.TestStep{
			// Test basic creation
			{
				Config: testAccResourceYandexMdbGreenplumClusterV2_basic(clusterName, clusterDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYandexMdbGreenplumClusterV2Exists("yandex_mdb_greenplum_cluster_v2.test", clusterName),
					resource.TestCheckResourceAttr("yandex_mdb_greenplum_cluster_v2.test", "name", clusterName),
					resource.TestCheckResourceAttr("yandex_mdb_greenplum_cluster_v2.test", "description", clusterDesc),
					resource.TestCheckResourceAttrSet("yandex_mdb_greenplum_cluster_v2.test", "folder_id"),
					resource.TestCheckResourceAttrSet("yandex_mdb_greenplum_cluster_v2.test", "created_at"),
					testAccCheckYandexMdbGreenplumClusterV2HasCloudStorage("yandex_mdb_greenplum_cluster_v2.test", true),
					resource.TestCheckResourceAttr("yandex_mdb_greenplum_cluster_v2.test", "segment_host_count", "2"),
					resource.TestCheckResourceAttr("yandex_mdb_greenplum_cluster_v2.test", "segment_in_host", "1"),
				),
			},
			// Test update with background activities and segment configuration (Update + Expand)
			{
				Config: testAccResourceYandexMdbGreenplumClusterV2_updated(clusterName, updatedDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckYandexMdbGreenplumClusterV2Exists("yandex_mdb_greenplum_cluster_v2.test", clusterName),
					resource.TestCheckResourceAttr("yandex_mdb_greenplum_cluster_v2.test", "description", updatedDesc),
					resource.TestCheckResourceAttr("yandex_mdb_greenplum_cluster_v2.test", "cluster_config.background_activities.analyze_and_vacuum.analyze_timeout", "10800"),
					resource.TestCheckResourceAttr("yandex_mdb_greenplum_cluster_v2.test", "cluster_config.background_activities.analyze_and_vacuum.vacuum_timeout", "10800"),
					resource.TestCheckResourceAttr("yandex_mdb_greenplum_cluster_v2.test", "segment_host_count", "3"),
					resource.TestCheckResourceAttr("yandex_mdb_greenplum_cluster_v2.test", "segment_in_host", "2"),
				),
			},
			// Test import
			{
				ResourceName:      "yandex_mdb_greenplum_cluster_v2.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"restore",       // restore is write-only
					"user_password", // user_password is write-only
				},
			},
		},
	})
}

func testAccResourceYandexMdbGreenplumClusterV2_basic(name, desc string) string {
	return fmt.Sprintf(gpVPCDependencies+`
resource "yandex_mdb_greenplum_cluster_v2" "test" {
  depends_on = [yandex_vpc_subnet.mdb-gp-test-subnet-a]
  
  name        = "%s"
  description = "%s"
  folder_id   = "%s"
  environment = "PRESTABLE"

  segment_host_count = 2
  segment_in_host   = 1

  user_name = "test-user"
  user_password = "test-user-password"
  network_id = yandex_vpc_network.mdb-gp-test-net.id

  cluster_config = {
    assign_public_ip = true
    backup_window_start = {
      hours   = 1
      minutes = 30
    }
  }

  config = {
	  zone_id = "ru-central1-a"
  }

  master_config = {
    resources = {
      resource_preset_id = "s2.small"
      disk_type_id       = "network-ssd"
      disk_size          = 10
    }
  }

  segment_config = {
    resources = {
      resource_preset_id = "s2.small"
      disk_type_id       = "network-ssd"
      disk_size          = 10
    }
  }

  cloud_storage = {
    enable = true
  }
}
`, name, desc, testhelpers.GetExampleFolderID())
}

func testAccResourceYandexMdbGreenplumClusterV2_updated(name, desc string) string {
	return fmt.Sprintf(gpVPCDependencies+`
resource "yandex_mdb_greenplum_cluster_v2" "test" {
  depends_on = [yandex_vpc_subnet.mdb-gp-test-subnet-a]

  name        = "%s"
  description = "%s"
  folder_id   = "%s"
  environment = "PRESTABLE"

  segment_host_count = 3
  segment_in_host   = 2

  user_name = "test-user"
  user_password = "test-user-password"
  network_id = yandex_vpc_network.mdb-gp-test-net.id

  cluster_config = {
    assign_public_ip = true
    backup_window_start = {
      hours   = 1
      minutes = 30
    }
    
    background_activities = {
      analyze_and_vacuum = {
        analyze_timeout = 10800
        vacuum_timeout  = 10800
        start = {
          hours   = 2
          minutes = 0
        }
      }
      
      query_killer_scripts = {
        idle = {
          enable       = true
          max_age      = 3600
          ignore_users = ["monitoring"]
        }
        idle_in_transaction = {
          enable       = true
          max_age      = 1800
          ignore_users = ["admin"]
        }
        long_running = {
          enable       = true
          max_age      = 7200
          ignore_users = ["dba"]
        }
      }
    }
  }

  config = {
	  zone_id = "ru-central1-a"
  }

  master_config = {
    resources = {
      resource_preset_id = "s2.small"
      disk_type_id       = "network-ssd"
      disk_size          = 10
    }
  }

  segment_config = {
    resources = {
      resource_preset_id = "s2.small"
      disk_type_id       = "network-ssd"
      disk_size          = 10
    }
  }

  cloud_storage = {
    enable = true
  }
}
`, name, desc, testhelpers.GetExampleFolderID())
}

func testAccCheckYandexMdbGreenplumClusterV2Destroy(s *terraform.State) error {
	config := testhelpers.AccProvider.(*provider.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_mdb_greenplum_cluster_v2" {
			continue
		}

		_, err := config.SDK.MDB().Greenplum().Cluster().Get(context.Background(), &greenplum.GetClusterRequest{
			ClusterId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("Greenplum Cluster still exists: %s", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckYandexMdbGreenplumClusterV2Exists(n string, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Note: Since testhelpers doesn't have Greenplum-specific functions,
		// we'll need to implement the exists check using generic provider methods
		// This is a placeholder - actual implementation would depend on available helpers
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testhelpers.AccProvider.(*provider.Provider).GetConfig()
		cluster, err := config.SDK.MDB().Greenplum().Cluster().Get(context.Background(), &greenplum.GetClusterRequest{
			ClusterId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if cluster.Name != name {
			return fmt.Errorf("Greenplum Cluster name is wrong: expected %s, got %s", name, cluster.Name)
		}

		return nil
	}
}

func testAccCheckYandexMdbGreenplumClusterV2HasCloudStorage(n string, enabled bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Note: Since testhelpers doesn't have Greenplum-specific functions,
		// we'll need to implement the cloud storage check using generic provider methods
		// This is a placeholder - actual implementation would depend on available helpers
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testhelpers.AccProvider.(*provider.Provider).GetConfig()
		cluster, err := config.SDK.MDB().Greenplum().Cluster().Get(context.Background(), &greenplum.GetClusterRequest{
			ClusterId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if cluster.CloudStorage == nil {
			return fmt.Errorf("Cloud storage config is missing")
		}

		if cluster.CloudStorage.Enable != enabled {
			return fmt.Errorf("Cloud storage enable flag is wrong: expected %v, got %v", enabled, cluster.CloudStorage.Enable)
		}

		return nil
	}
}
