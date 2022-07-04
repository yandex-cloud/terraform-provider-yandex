package yandex

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/sqlserver/v1"
)

const sqlserverResource = "yandex_mdb_sqlserver_cluster.foo"

func init() {
	resource.AddTestSweepers("yandex_mdb_sqlserver_cluster", &resource.Sweeper{
		Name: "yandex_mdb_sqlserver_cluster",
		F:    testSweepMDBSQLServerCluster,
	})
}

func testSweepMDBSQLServerCluster(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := conf.sdk.MDB().SQLServer().Cluster().List(conf.Context(), &sqlserver.ListClustersRequest{
		FolderId: conf.FolderID,
		PageSize: defaultMDBPageSize,
	})
	if err != nil {
		return fmt.Errorf("error getting SQLServer clusters: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.Clusters {
		if !sweepMDBSQLServerCluster(conf, c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep SQLServer cluster %q", c.Id))
		}
	}

	return result.ErrorOrNil()
}

func sweepMDBSQLServerCluster(conf *Config, id string) bool {
	return sweepWithRetry(sweepMDBSQLServerClusterOnce, conf, "SQLServer cluster", id)
}

func sweepMDBSQLServerClusterOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexMDBSQLServerClusterDefaultTimeout)
	defer cancel()

	mask := field_mask.FieldMask{Paths: []string{"deletion_protection"}}
	op, err := conf.sdk.MDB().SQLServer().Cluster().Update(ctx, &sqlserver.UpdateClusterRequest{
		ClusterId:          id,
		DeletionProtection: false,
		UpdateMask:         &mask,
	})
	err = handleSweepOperation(ctx, conf, op, err)
	if err != nil && !strings.EqualFold(errorMessage(err), "no changes detected") {
		return err
	}

	op, err = conf.sdk.MDB().SQLServer().Cluster().Delete(ctx, &sqlserver.DeleteClusterRequest{
		ClusterId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}

func mdbSQLServerClusterImportStep(name string) resource.TestStep {
	return resource.TestStep{
		ResourceName:      name,
		ImportState:       true,
		ImportStateVerify: true,
		ImportStateVerifyIgnore: []string{
			"user",   // passwords are not returned
			"health", // volatile value
		},
	}
}

// Test that a SQLServer Cluster can be created, updated and destroyed
func TestAccMDBSQLServerCluster_full(t *testing.T) {
	if os.Getenv("TF_SQL_LICENSE_ACCEPTED") != "1" {
		t.Skip()
	}
	t.Parallel()

	SQLServerName := acctest.RandomWithPrefix("tf-sqlserver")
	SQLServerDesc := "SQLServer Cluster Terraform Test"
	SQLServerDesc2 := "SQLServer Cluster Terraform Test Updated"
	folderID := getExampleFolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMDBSQLServerClusterDestroy,
		Steps: []resource.TestStep{
			//Create SQLServer Cluster
			{
				Config: testAccMDBSQLServerClusterConfigMain(SQLServerName, SQLServerDesc, "PRESTABLE", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBSQLServerClusterExists(sqlserverResource, 1),
					resource.TestCheckResourceAttr(sqlserverResource, "name", SQLServerName),
					resource.TestCheckResourceAttr(sqlserverResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(sqlserverResource, "description", SQLServerDesc),
					resource.TestCheckResourceAttrSet(sqlserverResource, "host.0.fqdn"),
					resource.TestCheckResourceAttr(sqlserverResource, "resources.0.disk_size", "10"),
					resource.TestCheckResourceAttr(sqlserverResource, "resources.0.disk_type_id", "network-ssd"),
					resource.TestCheckResourceAttr(sqlserverResource, "resources.0.resource_preset_id", "s2.small"),
					testAccCheckCreatedAtAttr(sqlserverResource),
					resource.TestCheckResourceAttr(sqlserverResource, "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(sqlserverResource, "deletion_protection", "true"),
					resource.TestCheckResourceAttr(sqlserverResource, "sqlcollation", "Cyrillic_General_CI_AI"),
				),
			},
			mdbSQLServerClusterImportStep(sqlserverResource),
			// uncheck 'deletion_protection
			{
				Config: testAccMDBSQLServerClusterConfigMain(SQLServerName, SQLServerDesc, "PRESTABLE", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBSQLServerClusterExists(sqlserverResource, 1),
					resource.TestCheckResourceAttr(sqlserverResource, "deletion_protection", "false"),
				),
			},
			mdbSQLServerClusterImportStep(sqlserverResource),
			// check 'deletion_protection
			{
				Config: testAccMDBSQLServerClusterConfigMain(SQLServerName, SQLServerDesc, "PRESTABLE", true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBSQLServerClusterExists(sqlserverResource, 1),
					resource.TestCheckResourceAttr(sqlserverResource, "deletion_protection", "true"),
				),
			},
			// trigger deletion by changing environment
			{
				Config:      testAccMDBSQLServerClusterConfigMain(SQLServerName, SQLServerDesc, "PRODUCTION", true),
				ExpectError: regexp.MustCompile(".*The operation was rejected because cluster has 'deletion_protection' = ON.*"),
			},
			mdbSQLServerClusterImportStep(sqlserverResource),
			// uncheck 'deletion_protection
			{
				Config: testAccMDBSQLServerClusterConfigMain(SQLServerName, SQLServerDesc, "PRESTABLE", false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBSQLServerClusterExists(sqlserverResource, 1),
					resource.TestCheckResourceAttr(sqlserverResource, "deletion_protection", "false"),
				),
			},
			mdbSQLServerClusterImportStep(sqlserverResource),
			// Change some options
			{
				Config: testAccMDBSQLServerClusterConfigUpdated(SQLServerName, SQLServerDesc2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBSQLServerClusterExists(sqlserverResource, 1),
					resource.TestCheckResourceAttr(sqlserverResource, "name", SQLServerName),
					resource.TestCheckResourceAttr(sqlserverResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(sqlserverResource, "description", SQLServerDesc2),
					resource.TestCheckResourceAttrSet(sqlserverResource, "host.0.fqdn"),
					resource.TestCheckResourceAttr(sqlserverResource, "labels.test_key", "test_value"),
					resource.TestCheckResourceAttr(sqlserverResource, "resources.0.disk_size", "20"),
					resource.TestCheckResourceAttr(sqlserverResource, "backup_window_start.0.hours", "20"),
					resource.TestCheckResourceAttr(sqlserverResource, "backup_window_start.0.minutes", "30"),
					resource.TestCheckResourceAttr(sqlserverResource, "sqlserver_config.fill_factor_percent", "49"),
					resource.TestCheckResourceAttr(sqlserverResource, "sqlserver_config.optimize_for_ad_hoc_workloads", "true"),
					resource.TestCheckResourceAttr(sqlserverResource, "user.0.name", "bob"),
					resource.TestCheckResourceAttr(sqlserverResource, "user.1.name", "alice"),
					resource.TestCheckResourceAttr(sqlserverResource, "user.2.name", "chuck"),
					resource.TestCheckResourceAttr(sqlserverResource, "user.0.permission.#", "0"),
					resource.TestCheckResourceAttr(sqlserverResource, "user.1.permission.#", "1"),
					resource.TestCheckResourceAttr(sqlserverResource, "user.2.permission.#", "3"),
					resource.TestCheckResourceAttr(sqlserverResource, "database.0.name", "testdb"),
					resource.TestCheckResourceAttr(sqlserverResource, "database.1.name", "testdb-a"),
					resource.TestCheckResourceAttr(sqlserverResource, "database.2.name", "testdb-b"),
					testAccCheckCreatedAtAttr(sqlserverResource),
					resource.TestCheckResourceAttr(sqlserverResource, "security_group_ids.#", "2"),
					resource.TestCheckResourceAttr(sqlserverResource, "sqlcollation", "Cyrillic_General_CI_AI"),
				),
			},
			mdbSQLServerClusterImportStep(sqlserverResource),
		},
	})
}

// Test that a SQLServer Cluster can't be created with 2 hosts
func TestAccMDBSQLServerCluster_ha2(t *testing.T) {
	if os.Getenv("TF_SQL_LICENSE_ACCEPTED") != "1" {
		t.Skip()
	}
	t.Parallel()

	SQLServerName := acctest.RandomWithPrefix("tf-sqlserver-ha2")
	SQLServerDesc := "SQLServer Cluster Terraform Test 2 hosts"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMDBSQLServerClusterDestroy,
		Steps: []resource.TestStep{
			//Create SQLServer Cluster
			{
				Config:      testAccMDBSQLServerClusterConfigHA2(SQLServerName, SQLServerDesc),
				ExpectError: regexp.MustCompile(".*code = InvalidArgument desc = only 1-node and 3-node clusters are supported now"),
			},
		},
	})
}

// Test that a SQLServer Cluster can be created with 3 hosts, updated and destroyed
func TestAccMDBSQLServerCluster_ha3(t *testing.T) {
	if os.Getenv("TF_SQL_LICENSE_ACCEPTED") != "1" {
		t.Skip()
	}
	t.Parallel()

	SQLServerName := acctest.RandomWithPrefix("tf-sqlserver-ha3")
	SQLServerDesc := "SQLServer Cluster Terraform Test 3 hosts"
	folderID := getExampleFolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMDBSQLServerClusterDestroy,
		Steps: []resource.TestStep{
			//Create SQLServer Cluster
			{
				Config: testAccMDBSQLServerClusterConfigHA3(SQLServerName, SQLServerDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBSQLServerClusterExists(sqlserverResource, 3),
					resource.TestCheckResourceAttr(sqlserverResource, "name", SQLServerName),
					resource.TestCheckResourceAttr(sqlserverResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(sqlserverResource, "description", SQLServerDesc),
					resource.TestCheckResourceAttrSet(sqlserverResource, "host.0.fqdn"),
					resource.TestCheckResourceAttrSet(sqlserverResource, "host.1.fqdn"),
					resource.TestCheckResourceAttrSet(sqlserverResource, "host.2.fqdn"),
					resource.TestCheckResourceAttr(sqlserverResource, "resources.0.disk_size", "10"),
					resource.TestCheckResourceAttr(sqlserverResource, "resources.0.disk_type_id", "network-ssd"),
					resource.TestCheckResourceAttr(sqlserverResource, "resources.0.resource_preset_id", "s2.small"),
					testAccCheckCreatedAtAttr(sqlserverResource),
					resource.TestCheckResourceAttr(sqlserverResource, "security_group_ids.#", "1"),
				),
			},
			mdbSQLServerClusterImportStep(sqlserverResource),
		},
	})
}

func testAccCheckMDBSQLServerClusterDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_mdb_SQLServer_cluster" {
			continue
		}

		_, err := config.sdk.MDB().SQLServer().Cluster().Get(context.Background(), &sqlserver.GetClusterRequest{
			ClusterId: rs.Primary.ID,
		})

		if err == nil {
			return fmt.Errorf("SQLServer Cluster still exists")
		}
	}

	return nil
}

func testAccCheckMDBSQLServerClusterExists(n string, hosts int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.MDB().SQLServer().Cluster().Get(context.Background(), &sqlserver.GetClusterRequest{
			ClusterId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("SQLServer Cluster not found")
		}

		resp, err := config.sdk.MDB().SQLServer().Cluster().ListHosts(context.Background(), &sqlserver.ListClusterHostsRequest{
			ClusterId: rs.Primary.ID,
			PageSize:  defaultMDBPageSize,
		})
		if err != nil {
			return err
		}

		if len(resp.Hosts) != hosts {
			return fmt.Errorf("Expected %d hosts, got %d", hosts, len(resp.Hosts))
		}

		return nil
	}
}

const sqlserverVPCDependencies = `
resource "yandex_vpc_network" "mdb-sqlserver-test-net" {}

resource "yandex_vpc_subnet" "mdb-sqlserver-test-subnet-a" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.mdb-sqlserver-test-net.id
  v4_cidr_blocks = ["10.1.0.0/24"]
}

resource "yandex_vpc_subnet" "mdb-sqlserver-test-subnet-b" {
  zone           = "ru-central1-b"
  network_id     = yandex_vpc_network.mdb-sqlserver-test-net.id
  v4_cidr_blocks = ["10.2.0.0/24"]
}

resource "yandex_vpc_subnet" "mdb-sqlserver-test-subnet-c" {
  zone           = "ru-central1-c"
  network_id     = yandex_vpc_network.mdb-sqlserver-test-net.id
  v4_cidr_blocks = ["10.3.0.0/24"]
}

resource "yandex_vpc_security_group" "mdb-sqlserver-test-sg-x" {
  network_id = yandex_vpc_network.mdb-sqlserver-test-net.id
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

resource "yandex_vpc_security_group" "mdb-sqlserver-test-sg-y" {
  network_id = yandex_vpc_network.mdb-sqlserver-test-net.id

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

func testAccMDBSQLServerClusterConfigMain(name, desc, environment string, deletionProtection bool) string {
	return fmt.Sprintf(sqlserverVPCDependencies+`
resource "yandex_mdb_sqlserver_cluster" "foo" {
  name        = "%s"
  description = "%s"
  environment = "%s"
  network_id  = yandex_vpc_network.mdb-sqlserver-test-net.id


  version = "2016sp2ent"

  labels = { test_key_create : "test_value_create" }

  resources {
    resource_preset_id = "s2.small"
    disk_size          = 10
    disk_type_id       = "network-ssd"
  }

  user {
    name     = "alice"
    password = "mysecurepassword"

    permission {
      database_name = "testdb"
      roles         = ["OWNER", "DDLADMIN"]
    }
  }

  host {
    zone      = "ru-central1-a"
    subnet_id = yandex_vpc_subnet.mdb-sqlserver-test-subnet-a.id
  }

  database {
    name = "testdb"
  }

  security_group_ids = [yandex_vpc_security_group.mdb-sqlserver-test-sg-x.id]

  sqlcollation = "Cyrillic_General_CI_AI"

  deletion_protection = %t
}
`, name, desc, environment, deletionProtection)
}

func testAccMDBSQLServerClusterConfigUpdated(name, desc string) string {
	return fmt.Sprintf(sqlserverVPCDependencies+`
resource "yandex_mdb_sqlserver_cluster" "foo" {
  name        = "%s"
  description = "%s"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.mdb-sqlserver-test-net.id


  version = "2016sp2ent"

  resources {
    resource_preset_id = "s2.small"
    disk_size          = 20
    disk_type_id       = "network-ssd"
  }


  labels = { test_key : "test_value" }

  backup_window_start {
    hours   = 20
    minutes = 30
  }


  sqlserver_config = {
    fill_factor_percent           = 49
    optimize_for_ad_hoc_workloads = true
  }


  user {
    name     = "bob"
    password = "mysecurepassword"

  }

  user {
    name     = "alice"
    password = "mysecurepassword"

    permission {
      database_name = "testdb"
      roles         = ["DDLADMIN"]
    }
  }


  user {
    name     = "chuck"
    password = "mysecurepassword"

    permission {
      database_name = "testdb-b"
      roles         = ["OWNER"]
    }
    permission {
      database_name = "testdb"
      roles         = ["OWNER", "DDLADMIN"]
    }
    permission {
      database_name = "testdb-a"
      roles         = ["OWNER", "DDLADMIN"]
    }
  }

  host {
    zone      = "ru-central1-a"
    subnet_id = yandex_vpc_subnet.mdb-sqlserver-test-subnet-a.id
  }

  database {
    name = "testdb"
  }
  database {
    name = "testdb-a"
  }
  database {
    name = "testdb-b"
  }

  security_group_ids = [yandex_vpc_security_group.mdb-sqlserver-test-sg-x.id, yandex_vpc_security_group.mdb-sqlserver-test-sg-y.id]

  sqlcollation = "Cyrillic_General_CS_AS"
}
`, name, desc)
}

func testAccMDBSQLServerClusterConfigHA2(name, desc string) string {
	return fmt.Sprintf(sqlserverVPCDependencies+`
resource "yandex_mdb_sqlserver_cluster" "foo" {
  name        = "%s"
  description = "%s"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.mdb-sqlserver-test-net.id


  version = "2016sp2ent"

  labels = { test_key_create : "test_value_create" }

  resources {
    resource_preset_id = "s2.small"
    disk_size          = 10
    disk_type_id       = "network-ssd"
  }

  user {
    name     = "alice"
    password = "mysecurepassword"

    permission {
      database_name = "testdb"
      roles         = ["OWNER", "DDLADMIN"]
    }
  }

  host {
    zone      = "ru-central1-a"
    subnet_id = yandex_vpc_subnet.mdb-sqlserver-test-subnet-a.id
  }
  host {
    zone      = "ru-central1-a"
    subnet_id = yandex_vpc_subnet.mdb-sqlserver-test-subnet-a.id
  }

  database {
    name = "testdb"
  }

  security_group_ids = [yandex_vpc_security_group.mdb-sqlserver-test-sg-x.id]
}
`, name, desc)
}

func testAccMDBSQLServerClusterConfigHA3(name, desc string) string {
	return fmt.Sprintf(sqlserverVPCDependencies+`
resource "yandex_mdb_sqlserver_cluster" "foo" {
  name        = "%s"
  description = "%s"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.mdb-sqlserver-test-net.id


  version = "2016sp2ent"

  labels = { test_key_create : "test_value_create" }

  resources {
    resource_preset_id = "s2.small"
    disk_size          = 10
    disk_type_id       = "network-ssd"
  }

  user {
    name     = "alice"
    password = "mysecurepassword"

    permission {
      database_name = "testdb"
      roles         = ["OWNER", "DDLADMIN"]
    }
  }

  host {
    zone      = "ru-central1-a"
    subnet_id = yandex_vpc_subnet.mdb-sqlserver-test-subnet-a.id
  }
  host {
    zone      = "ru-central1-a"
    subnet_id = yandex_vpc_subnet.mdb-sqlserver-test-subnet-a.id
  }
  host {
    zone      = "ru-central1-b"
    subnet_id = yandex_vpc_subnet.mdb-sqlserver-test-subnet-b.id
  }

  database {
    name = "testdb"
  }

  security_group_ids = [yandex_vpc_security_group.mdb-sqlserver-test-sg-x.id]
}
`, name, desc)
}
