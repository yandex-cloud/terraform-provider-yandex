package yandex

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/greenplum/v1"
)

const greenplumResource = "yandex_mdb_greenplum_cluster.foo"

func init() {
	resource.AddTestSweepers("yandex_mdb_greenplum_cluster", &resource.Sweeper{
		Name: "yandex_mdb_greenplum_cluster",
		F:    testSweepMDBGreenplumCluster,
	})
}

func testSweepMDBGreenplumCluster(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := conf.sdk.MDB().Greenplum().Cluster().List(conf.Context(), &greenplum.ListClustersRequest{
		FolderId: conf.FolderID,
		PageSize: defaultMDBPageSize,
	})
	if err != nil {
		return fmt.Errorf("error getting Greenplum clusters: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.Clusters {
		if !sweepMDBGreenplumCluster(conf, c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Greenplum cluster %q", c.Id))
		}
	}

	return result.ErrorOrNil()
}

func sweepMDBGreenplumCluster(conf *Config, id string) bool {
	return sweepWithRetry(sweepMDBGreenplumClusterOnce, conf, "Greenplum cluster", id)
}

func sweepMDBGreenplumClusterOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexMDBGreenplumClusterDefaultTimeout)
	defer cancel()

	mask := field_mask.FieldMask{Paths: []string{"deletion_protection"}}
	op, err := conf.sdk.MDB().Greenplum().Cluster().Update(ctx, &greenplum.UpdateClusterRequest{
		ClusterId:          id,
		DeletionProtection: false,
		UpdateMask:         &mask,
	})
	err = handleSweepOperation(ctx, conf, op, err)
	if err != nil && !strings.EqualFold(errorMessage(err), "no changes detected") {
		return err
	}

	op, err = conf.sdk.MDB().Greenplum().Cluster().Delete(ctx, &greenplum.DeleteClusterRequest{
		ClusterId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}

func mdbGreenplumClusterImportStep(name string) resource.TestStep {
	return resource.TestStep{
		ResourceName:      name,
		ImportState:       true,
		ImportStateVerify: true,
		ImportStateVerifyIgnore: []string{
			"user_password", // passwords are not returned
			"health",        // volatile value
		},
	}
}

// Test that a Greenplum Cluster can be created, updated and destroyed
func TestAccMDBGreenplumCluster_full(t *testing.T) {
	t.Parallel()

	clusterName := acctest.RandomWithPrefix("tf-greenplum")
	clusterNameUpdated := clusterName + "_updated"
	clusterDescription := "Greenplum Cluster Terraform Test"
	clusterDescriptionUpdated := clusterDescription + " Updated"
	folderID := getExampleFolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMDBGreenplumClusterDestroy,
		Steps: []resource.TestStep{
			// Create Greenplum Cluster
			{
				Config: testAccMDBGreenplumClusterConfigStep1(clusterName, clusterDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBGreenplumClusterExists(greenplumResource, 2, 5),
					resource.TestCheckResourceAttr(greenplumResource, "name", clusterName),
					resource.TestCheckResourceAttr(greenplumResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(greenplumResource, "description", clusterDescription),
					testAccCheckCreatedAtAttr(greenplumResource),
					resource.TestCheckResourceAttr(greenplumResource, "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(greenplumResource, "deletion_protection", "false"),

					resource.TestCheckResourceAttr(greenplumResource, "pooler_config.0.pooling_mode", "TRANSACTION"),
					resource.TestCheckResourceAttr(greenplumResource, "pooler_config.0.pool_size", "10"),
					resource.TestCheckResourceAttr(greenplumResource, "pooler_config.0.pool_client_idle_timeout", "0"),

					resource.TestCheckResourceAttr(greenplumResource, "pxf_config.0.connection_timeout", "600"),
					resource.TestCheckResourceAttr(greenplumResource, "pxf_config.0.upload_timeout", "40"),
					resource.TestCheckResourceAttr(greenplumResource, "pxf_config.0.max_threads", "100"),
					resource.TestCheckResourceAttr(greenplumResource, "pxf_config.0.pool_allow_core_thread_timeout", "false"),
					resource.TestCheckResourceAttr(greenplumResource, "pxf_config.0.pool_core_size", "8"),
					resource.TestCheckResourceAttr(greenplumResource, "pxf_config.0.pool_queue_capacity", "600"),
					resource.TestCheckResourceAttr(greenplumResource, "pxf_config.0.pool_max_size", "128"),
					resource.TestCheckResourceAttr(greenplumResource, "pxf_config.0.xmx", "2048"),
					resource.TestCheckResourceAttr(greenplumResource, "pxf_config.0.xms", "1024"),

					resource.TestCheckResourceAttr(greenplumResource, "greenplum_config.max_connections", "395"),
					resource.TestCheckResourceAttr(greenplumResource, "greenplum_config.max_slot_wal_keep_size", "1048576"),
					resource.TestCheckResourceAttr(greenplumResource, "greenplum_config.gp_workfile_limit_per_segment", "0"),
					resource.TestCheckResourceAttr(greenplumResource, "greenplum_config.gp_workfile_limit_per_query", "0"),
					resource.TestCheckResourceAttr(greenplumResource, "greenplum_config.gp_workfile_limit_files_per_query", "100000"),
					resource.TestCheckResourceAttr(greenplumResource, "greenplum_config.max_prepared_transactions", "500"),
					resource.TestCheckResourceAttr(greenplumResource, "greenplum_config.gp_workfile_compression", "false"),

					resource.TestCheckResourceAttr(greenplumResource, "master_subcluster.0.resources.0.resource_preset_id", "s2.micro"),
					resource.TestCheckResourceAttr(greenplumResource, "master_subcluster.0.resources.0.disk_size", "24"),
					resource.TestCheckResourceAttr(greenplumResource, "master_subcluster.0.resources.0.disk_type_id", "network-ssd"),
					resource.TestCheckResourceAttr(greenplumResource, "segment_subcluster.0.resources.0.resource_preset_id", "s2.small"),
					resource.TestCheckResourceAttr(greenplumResource, "segment_subcluster.0.resources.0.disk_size", "24"),
					resource.TestCheckResourceAttr(greenplumResource, "segment_subcluster.0.resources.0.disk_type_id", "network-ssd"),
				),
			},

			// Changing resource_preset_id
			{
				Config: testAccMDBGreenplumClusterConfigStep2(clusterName, clusterDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBGreenplumClusterExists(greenplumResource, 2, 5),
					resource.TestCheckResourceAttr(greenplumResource, "master_subcluster.0.resources.0.resource_preset_id", "s2.small"),
					resource.TestCheckResourceAttr(greenplumResource, "segment_subcluster.0.resources.0.resource_preset_id", "s2.small"),
				),
			},
			mdbGreenplumClusterImportStep(greenplumResource),
			// Update name and description of the cluster
			{
				Config: testAccMDBGreenplumClusterConfigStep3(clusterNameUpdated, clusterDescriptionUpdated),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(greenplumResource, "name", clusterNameUpdated),
					resource.TestCheckResourceAttr(greenplumResource, "description", clusterDescriptionUpdated),
				),
			},
			mdbGreenplumClusterImportStep(greenplumResource),
			// Update pooler_config, pxf_config and greenplum_config
			{
				Config: testAccMDBGreenplumClusterConfigStep4(clusterNameUpdated, clusterDescriptionUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBGreenplumClusterExists(greenplumResource, 2, 5),
					resource.TestCheckResourceAttr(greenplumResource, "greenplum_config.max_connections", "400"),
					resource.TestCheckResourceAttr(greenplumResource, "greenplum_config.gp_workfile_compression", "true"),
					resource.TestCheckResourceAttr(greenplumResource, "pooler_config.0.pooling_mode", "SESSION"),
					resource.TestCheckResourceAttr(greenplumResource, "pooler_config.0.pool_size", "10"),
					resource.TestCheckResourceAttr(greenplumResource, "pooler_config.0.pool_client_idle_timeout", "0"),
					resource.TestCheckResourceAttr(greenplumResource, "pxf_config.0.connection_timeout", "300"),
					resource.TestCheckResourceAttr(greenplumResource, "pxf_config.0.upload_timeout", "440"),
					resource.TestCheckResourceAttr(greenplumResource, "pxf_config.0.max_threads", "200"),
					resource.TestCheckResourceAttr(greenplumResource, "pxf_config.0.pool_allow_core_thread_timeout", "true"),
					resource.TestCheckResourceAttr(greenplumResource, "pxf_config.0.pool_core_size", "4"),
					resource.TestCheckResourceAttr(greenplumResource, "pxf_config.0.pool_max_size", "48"),
					resource.TestCheckResourceAttr(greenplumResource, "pxf_config.0.xmx", "1536"),
					resource.TestCheckResourceAttr(greenplumResource, "pxf_config.0.xms", "1024"),
				),
			},
			mdbGreenplumClusterImportStep(greenplumResource),
			// Update deletion_protection
			{
				Config: testAccMDBGreenplumClusterConfigStep5(clusterNameUpdated, clusterDescriptionUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBGreenplumClusterExists(greenplumResource, 2, 5),
					testAccCheckCreatedAtAttr(greenplumResource),
					resource.TestCheckResourceAttr(greenplumResource, "deletion_protection", "true"),
				),
			},
			mdbGreenplumClusterImportStep(greenplumResource),
			// Add access, background_activities and backup_window_start fields
			{
				Config: testAccMDBGreenplumClusterConfigStep6(clusterNameUpdated, clusterDescriptionUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBGreenplumClusterExists(greenplumResource, 2, 5),

					resource.TestCheckResourceAttr(greenplumResource, "access.0.web_sql", "true"),
					resource.TestCheckResourceAttr(greenplumResource, "access.0.data_lens", "true"),
					resource.TestCheckResourceAttr(greenplumResource, "access.0.data_transfer", "true"),
					resource.TestCheckResourceAttr(greenplumResource, "access.0.yandex_query", "true"),
					resource.TestCheckResourceAttr(greenplumResource, "backup_window_start.0.minutes", "15"),
					resource.TestCheckResourceAttr(greenplumResource, "maintenance_window.0.day", "SAT"),
					resource.TestCheckResourceAttr(greenplumResource, "maintenance_window.0.hour", "12"),
					resource.TestCheckResourceAttr(greenplumResource, "deletion_protection", "false"),
					resource.TestCheckResourceAttr(greenplumResource, "background_activities.0.analyze_and_vacuum.0.start_time", "12:34"),
					resource.TestCheckResourceAttr(greenplumResource, "background_activities.0.analyze_and_vacuum.0.analyze_timeout", "7200"),
					resource.TestCheckResourceAttr(greenplumResource, "background_activities.0.analyze_and_vacuum.0.vacuum_timeout", "7300"),
					resource.TestCheckResourceAttr(greenplumResource, "background_activities.0.query_killer_idle.0.enable", "true"),
					resource.TestCheckResourceAttr(greenplumResource, "background_activities.0.query_killer_idle.0.max_age", "6000"),
					resource.TestCheckResourceAttr(greenplumResource, "background_activities.0.query_killer_idle.0.ignore_users.0", "manager"),
					resource.TestCheckResourceAttr(greenplumResource, "background_activities.0.query_killer_idle.0.ignore_users.1", "developer"),
					resource.TestCheckResourceAttr(greenplumResource, "background_activities.0.query_killer_idle_in_transaction.0.enable", "true"),
					resource.TestCheckResourceAttr(greenplumResource, "background_activities.0.query_killer_idle_in_transaction.0.ignore_users.#", "0"),
					resource.TestCheckResourceAttr(greenplumResource, "background_activities.0.query_killer_long_running.0.enable", "true"),
					resource.TestCheckResourceAttr(greenplumResource, "background_activities.0.query_killer_long_running.0.max_age", "8000"),
				),
			},

			mdbGreenplumClusterImportStep(greenplumResource),
			{
				Config: testAccMDBGreenplumClusterConfigStep7(clusterNameUpdated, clusterDescriptionUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBGreenplumClusterExists(greenplumResource, 2, 7),
					testAccCheckCreatedAtAttr(greenplumResource),
					resource.TestCheckResourceAttr(greenplumResource, "segment_in_host", "2"),
					resource.TestCheckResourceAttr(greenplumResource, "access.0.web_sql", "true"),
					resource.TestCheckResourceAttr(greenplumResource, "access.0.data_lens", "false"),
					resource.TestCheckResourceAttr(greenplumResource, "access.0.data_transfer", "true"),
					resource.TestCheckResourceAttr(greenplumResource, "access.0.yandex_query", "false"),
				),
			},
			mdbGreenplumClusterImportStep(greenplumResource),
		},
	})
}

func testAccCheckMDBGreenplumClusterDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_mdb_greenplum_cluster" {
			continue
		}

		_, err := config.sdk.MDB().Greenplum().Cluster().Get(context.Background(), &greenplum.GetClusterRequest{
			ClusterId: rs.Primary.ID,
		})

		if err == nil {
			return fmt.Errorf("Greenplum Cluster still exists")
		}
	}

	return nil
}

func testAccCheckMDBGreenplumClusterExists(n string, masterHosts int, segmentHosts int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.MDB().Greenplum().Cluster().Get(context.Background(), &greenplum.GetClusterRequest{
			ClusterId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Greenplum Cluster not found")
		}

		resp, err := config.sdk.MDB().Greenplum().Cluster().ListMasterHosts(context.Background(), &greenplum.ListClusterHostsRequest{
			ClusterId: rs.Primary.ID,
			PageSize:  defaultMDBPageSize,
		})
		if err != nil {
			return err
		}

		if len(resp.Hosts) != masterHosts {
			return fmt.Errorf("Expected %d hosts, got %d", masterHosts, len(resp.Hosts))
		}

		resp, err = config.sdk.MDB().Greenplum().Cluster().ListSegmentHosts(context.Background(), &greenplum.ListClusterHostsRequest{
			ClusterId: rs.Primary.ID,
			PageSize:  defaultMDBPageSize,
		})
		if err != nil {
			return err
		}

		if len(resp.Hosts) != segmentHosts {
			return fmt.Errorf("Expected %d hosts, got %d", segmentHosts, len(resp.Hosts))
		}

		return nil
	}
}

const greenplumVPCDependencies = `
resource "yandex_vpc_network" "mdb-greenplum-test-net" {}

resource "yandex_vpc_subnet" "mdb-greenplum-test-subnet-b" {
  zone           = "ru-central1-b"
  network_id     = yandex_vpc_network.mdb-greenplum-test-net.id
  v4_cidr_blocks = ["10.2.0.0/24"]
}

resource "yandex_vpc_security_group" "mdb-greenplum-test-sg-x" {
  network_id = yandex_vpc_network.mdb-greenplum-test-net.id
  ingress {
    protocol       = "ANY"
    description    = "Allow incoming traffic from members of the same security group"
    from_port      = 0
    to_port        = 65535
    v4_cidr_blocks = ["0.0.0.0/0"]
  }
  egress {
    protocol       = "ANY"
    description    = "Allow outgoing traffic to members of the same security group"
    from_port      = 0
    to_port        = 65535
    v4_cidr_blocks = ["0.0.0.0/0"]
  }
}
`

func testAccMDBGreenplumClusterConfigTemplate(name, description, resourcePresetId string) string {
	return fmt.Sprintf(greenplumVPCDependencies+`
resource "yandex_mdb_greenplum_cluster" "foo" {
  name        = "%s"
  description = "%s"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.mdb-greenplum-test-net.id
  zone = "ru-central1-b"
  subnet_id = yandex_vpc_subnet.mdb-greenplum-test-subnet-b.id
  assign_public_ip = false
  version = "6.25"

  labels = { test_key_create : "test_value_create" }

  master_host_count  = 2

  master_subcluster {
    resources {
      resource_preset_id = "%s"
      disk_size          = 24
      disk_type_id       = "network-ssd"
    }
  }
  segment_subcluster {
    resources {
      resource_preset_id = "s2.small"
      disk_size          = 24
      disk_type_id       = "network-ssd"
    }
  }

  user_name     = "user1"
  user_password = "mysecurepassword"
  security_group_ids = [yandex_vpc_security_group.mdb-greenplum-test-sg-x.id]

`, name, description, resourcePresetId)
}

func testAccMDBGreenplumClusterConfigStep0(name, description, resourcePresetId string) string {
	return testAccMDBGreenplumClusterConfigTemplate(name, description, resourcePresetId) + `
  segment_host_count = 5
  segment_in_host    = 1
`
}

func testAccMDBGreenplumClusterConfigStep1(name string, description string) string {
	return testAccMDBGreenplumClusterConfigStep0(name, description, "s2.micro") + `
  pooler_config {
    pooling_mode             = "TRANSACTION"
    pool_size                = 10
    pool_client_idle_timeout = 0
  }

  pxf_config {
    connection_timeout             = 600
	upload_timeout                 = 40
	max_threads                    = 100
	pool_allow_core_thread_timeout = "false"
	pool_core_size                 = 8
	pool_queue_capacity            = 600
	pool_max_size                  = 128
	xmx                            = 2048
	xms                            = 1024
  }

  greenplum_config = {
    max_connections                   = 395
    max_slot_wal_keep_size            = 1048576
    gp_workfile_limit_per_segment     = 0
    gp_workfile_limit_per_query       = 0
    gp_workfile_limit_files_per_query = 100000
    max_prepared_transactions         = 500
    gp_workfile_compression           = "false"
  }
}`

}

func testAccMDBGreenplumClusterConfigStep2(name string, description string) string {
	return testAccMDBGreenplumClusterConfigStep0(name, description, "s2.small") + `
  pooler_config {
    pooling_mode             = "TRANSACTION"
    pool_size                = 10
    pool_client_idle_timeout = 0
  }

  pxf_config {
    connection_timeout             = 600
	upload_timeout                 = 40
	max_threads                    = 100
	pool_allow_core_thread_timeout = "false"
	pool_core_size                 = 8
	pool_queue_capacity            = 600
	pool_max_size                  = 128
	xmx                            = 2048
	xms                            = 1024
  }

  greenplum_config = {
    max_connections                   = 395
    max_slot_wal_keep_size            = 1048576
    gp_workfile_limit_per_segment     = 0
    gp_workfile_limit_per_query       = 0
    gp_workfile_limit_files_per_query = 100000
    max_prepared_transactions         = 500
    gp_workfile_compression           = "false"
  }
}`
}

func testAccMDBGreenplumClusterConfigStep3(name string, description string) string {
	return testAccMDBGreenplumClusterConfigStep2(name, description)
}

func testAccMDBGreenplumClusterConfigStep4(name string, description string) string {
	return testAccMDBGreenplumClusterConfigStep0(name, description, "s2.small") + `
  pooler_config {
    pooling_mode             = "SESSION"
    pool_size                = 10
    pool_client_idle_timeout = 0
  }

  pxf_config {
    connection_timeout             = 300
	upload_timeout                 = 440
	max_threads                    = 200
	pool_allow_core_thread_timeout = "true"
	pool_core_size                 = 4
	pool_queue_capacity            = 600
	pool_max_size                  = 48
	xmx                            = 1536
	xms                            = 1024
  }

  greenplum_config = {
    max_connections                   = 400
    max_slot_wal_keep_size            = 1048576
    gp_workfile_limit_per_segment     = 0
    gp_workfile_limit_per_query       = 0
    gp_workfile_limit_files_per_query = 100000
    max_prepared_transactions         = 500
    gp_workfile_compression           = "true"
  }
}`
}

func testAccMDBGreenplumClusterConfigStep5(name string, description string) string {
	return testAccMDBGreenplumClusterConfigStep0(name, description, "s2.small") + `
  pooler_config {
    pooling_mode             = "SESSION"
    pool_size                = 10
    pool_client_idle_timeout = 0
  }

  pxf_config {
    connection_timeout             = 300
	upload_timeout                 = 440
	max_threads                    = 200
	pool_allow_core_thread_timeout = "true"
	pool_core_size                 = 4
	pool_queue_capacity            = 600
	pool_max_size                  = 48
	xmx                            = 1536
	xms                            = 1024
  }

  greenplum_config = {
    max_connections                   = 400
    max_slot_wal_keep_size            = 1048576
    gp_workfile_limit_per_segment     = 0
    gp_workfile_limit_per_query       = 0
    gp_workfile_limit_files_per_query = 100000
    max_prepared_transactions         = 500
    gp_workfile_compression           = "true"
  }
  
  deletion_protection = true
}`
}

func testAccMDBGreenplumClusterConfigStep6(name string, description string) string {
	return testAccMDBGreenplumClusterConfigStep0(name, description, "s2.small") + `
  pooler_config {
    pooling_mode             = "SESSION"
    pool_size                = 10
    pool_client_idle_timeout = 0
  }

  pxf_config {
    connection_timeout             = 300
	upload_timeout                 = 440
	max_threads                    = 200
	pool_allow_core_thread_timeout = "true"
	pool_core_size                 = 4
	pool_queue_capacity            = 600
	pool_max_size                  = 48
	xmx                            = 1536
	xms                            = 1024
  }

  greenplum_config = {
    max_connections                   = 400
    max_slot_wal_keep_size            = 1048576
    gp_workfile_limit_per_segment     = 0
    gp_workfile_limit_per_query       = 0
    gp_workfile_limit_files_per_query = 100000
    max_prepared_transactions         = 500
    gp_workfile_compression           = "true"
  }
  
  deletion_protection = false

  backup_window_start {
    hours = 22
    minutes = 15
  }

  maintenance_window {
    type = "WEEKLY"
    day  = "SAT"
    hour = 12
  }
  access {
	web_sql       = true
	data_lens     = true
	data_transfer = true
	yandex_query  = true
  }
  cloud_storage {
	enable = true
  }

  background_activities {
	analyze_and_vacuum {
		start_time      = "12:34"
		analyze_timeout = 7200
		vacuum_timeout  = 7300
	}
	query_killer_idle {
		enable       = true
		max_age      = 6000
		ignore_users = [ "manager", "developer" ]
	}
	query_killer_idle_in_transaction {
		enable       = true
		ignore_users = []
	}
	query_killer_long_running {
		enable       = true
		max_age      = 8000
	}
  }
}`
}

func testAccMDBGreenplumClusterConfigStep7(name string, description string) string {
	return testAccMDBGreenplumClusterConfigTemplate(name, description, "s2.small") + `
  segment_host_count = 7
  segment_in_host    = 2

  pooler_config {
    pooling_mode             = "TRANSACTION"
    pool_size                = 10
    pool_client_idle_timeout = 0
  }

  pxf_config {
    connection_timeout             = 600
	upload_timeout                 = 40
	max_threads                    = 100
	pool_allow_core_thread_timeout = "true"
	pool_core_size                 = 8
	pool_queue_capacity            = 600
	pool_max_size                  = 128
	xmx                            = 2048
	xms                            = 1024
  }

  access {
    web_sql       = true
    data_transfer = true
  }
  cloud_storage {
    enable = true
  }
  background_activities {
	analyze_and_vacuum {
		start_time      = "12:34"
		analyze_timeout = 7200
		vacuum_timeout  = 8000
	}
	query_killer_idle {
		enable       = true
		max_age      = 6000
		ignore_users = [ "manager", "developer" ]
	}
	query_killer_idle_in_transaction {
		enable       = true
		ignore_users = []
	}
	query_killer_long_running {
		enable       = true
		max_age      = 8000
	}
  }
}`
}
