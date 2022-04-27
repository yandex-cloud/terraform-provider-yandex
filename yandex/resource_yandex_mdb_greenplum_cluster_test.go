package yandex

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/greenplum/v1"
	"google.golang.org/genproto/protobuf/field_mask"
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

	GreenplumName := acctest.RandomWithPrefix("tf-greenplum")
	greenplumNameMod := GreenplumName + "_mod"
	GreenplumDesc := "Greenplum Cluster Terraform Test"
	greenplumDescMod := GreenplumDesc + "_mod"
	folderID := getExampleFolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMDBGreenplumClusterDestroy,
		Steps: []resource.TestStep{
			//Create Greenplum Cluster
			{
				Config: testAccMDBGreenplumClusterConfigMain(GreenplumName, GreenplumDesc, "PRESTABLE", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBGreenplumClusterExists(greenplumResource, 2, 5),
					resource.TestCheckResourceAttr(greenplumResource, "name", GreenplumName),
					resource.TestCheckResourceAttr(greenplumResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(greenplumResource, "description", GreenplumDesc),
					testAccCheckCreatedAtAttr(greenplumResource),
					resource.TestCheckResourceAttr(greenplumResource, "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(greenplumResource, "deletion_protection", "false"),
				),
			},
			mdbGreenplumClusterImportStep(greenplumResource),
			// Change some options
			{
				Config: testAccMDBGreenplumClusterConfigUpdate(greenplumNameMod, greenplumDescMod, "PRESTABLE", false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(greenplumResource, "name", greenplumNameMod),
					resource.TestCheckResourceAttr(greenplumResource, "description", greenplumDescMod),
					resource.TestCheckResourceAttr(greenplumResource, "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(greenplumResource, "access.0.data_lens", "true"),
					resource.TestCheckResourceAttr(greenplumResource, "backup_window_start.0.minutes", "15"),
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

resource "yandex_vpc_security_group" "mdb-greenplum-test-sg-y" {
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

func testAccMDBGreenplumClusterConfigMain(name, desc, environment string, deletionProtection bool) string {
	return fmt.Sprintf(greenplumVPCDependencies+`
resource "yandex_mdb_greenplum_cluster" "foo" {
  name        = "%s"
  description = "%s"
  environment = "%s"
  network_id  = yandex_vpc_network.mdb-greenplum-test-net.id
  zone = "ru-central1-b"
  subnet_id = yandex_vpc_subnet.mdb-greenplum-test-subnet-b.id
  assign_public_ip = false
  version = "6.19"

  labels = { test_key_create : "test_value_create" }

  master_host_count  = 2
  segment_host_count = 5
  segment_in_host    = 1

  master_subcluster {
    resources {
      resource_preset_id = "s2.micro"
      disk_size          = 24
      disk_type_id       = "network-ssd"
    }
  }
  segment_subcluster {
    resources {
      resource_preset_id = "s2.micro"
      disk_size          = 24
      disk_type_id       = "network-ssd"
    }
  }

  user_name     = "user1"
  user_password = "mysecurepassword"

  security_group_ids = [yandex_vpc_security_group.mdb-greenplum-test-sg-x.id]

  deletion_protection = %t
}
`, name, desc, environment, deletionProtection)
}

func testAccMDBGreenplumClusterConfigUpdate(name, desc, environment string, deletionProtection bool) string {
	return fmt.Sprintf(greenplumVPCDependencies+`
resource "yandex_mdb_greenplum_cluster" "foo" {
  name        = "%s"
  description = "%s"
  environment = "%s"
  network_id  = yandex_vpc_network.mdb-greenplum-test-net.id
  zone = "ru-central1-b"
  subnet_id = yandex_vpc_subnet.mdb-greenplum-test-subnet-b.id
  assign_public_ip = false
  version = "6.19"

  labels = { test_key_create2 : "test_value_create2" }

  master_host_count  = 2
  segment_host_count = 5
  segment_in_host    = 1

  master_subcluster {
    resources {
      resource_preset_id = "s2.micro"
      disk_size          = 24
      disk_type_id       = "network-ssd"
    }
  }
  segment_subcluster {
    resources {
      resource_preset_id = "s2.micro"
      disk_size          = 24
      disk_type_id       = "network-ssd"
    }
  }

  access {
    data_lens = true
  }

  backup_window_start {
    hours = 22
    minutes = 15
  }

  user_name     = "user1"
  user_password = "mysecurepassword"

  security_group_ids = [yandex_vpc_security_group.mdb-greenplum-test-sg-x.id]

  deletion_protection = %t
}
`, name, desc, environment, deletionProtection)
}
