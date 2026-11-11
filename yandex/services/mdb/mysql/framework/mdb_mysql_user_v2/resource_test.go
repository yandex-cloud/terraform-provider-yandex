package mdb_mysql_user_v2_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func waitForOperations(seconds time.Duration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		time.Sleep(seconds * time.Second)
		return nil
	}
}

func TestAccMDBMySQLUserV2_basic(t *testing.T) {
	t.Parallel()

	clusterName := acctest.RandomWithPrefix("tf-mysql-user-v2")
	userName := acctest.RandomWithPrefix("testuser")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBMySQLUserV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBMySQLUserV2Basic(clusterName, userName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMySQLUserV2Exists(mysqlUserV2ResourceName),
					testAccCheckMDBMySQLUserV2ResourceIDField(mysqlUserV2ResourceName),
					resource.TestCheckResourceAttr(
						mysqlUserV2ResourceName, "name", userName,
					),
					resource.TestCheckResourceAttrSet(
						mysqlUserV2ResourceName, "cluster_id",
					),
					resource.TestCheckResourceAttr(
						mysqlUserV2ResourceName,
						"deletion_protection_mode",
						"DELETION_PROTECTION_MODE_DISABLED",
					),
					waitForOperations(30),
				),
			},
			mdbMySQLUserV2ImportStep(mysqlUserV2ResourceName),
		},
	})
}

func TestAccMDBMySQLUserV2_update(t *testing.T) {
	t.Parallel()

	clusterName := acctest.RandomWithPrefix("tf-mysql-user-v2")
	userName := acctest.RandomWithPrefix("testuser")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBMySQLUserV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBMySQLUserV2Basic(clusterName, userName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMySQLUserV2Exists(mysqlUserV2ResourceName),
					resource.TestCheckResourceAttr(
						mysqlUserV2ResourceName,
						"deletion_protection_mode",
						"DELETION_PROTECTION_MODE_DISABLED",
					),
					waitForOperations(30),
				),
			},
			{
				Config: testAccMDBMySQLUserV2WithDeletionProtection(clusterName, userName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMySQLUserV2Exists(mysqlUserV2ResourceName),
					resource.TestCheckResourceAttr(
						mysqlUserV2ResourceName,
						"deletion_protection_mode",
						"DELETION_PROTECTION_MODE_ENABLED",
					),
				),
			},
			mdbMySQLUserV2ImportStep(mysqlUserV2ResourceName),
			{
				Config: testAccMDBMySQLUserV2Basic(clusterName, userName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMySQLUserV2Exists(mysqlUserV2ResourceName),
					resource.TestCheckResourceAttr(
						mysqlUserV2ResourceName,
						"deletion_protection_mode",
						"DELETION_PROTECTION_MODE_DISABLED",
					),
					waitForOperations(30),
				),
			},
		},
	})
}

func TestAccMDBMySQLUserV2_full(t *testing.T) {
	t.Parallel()

	clusterName := acctest.RandomWithPrefix("tf-mysql-user-v2")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBMySQLUserV2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBMySQLUserV2ConfigStep1(clusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMySQLUserV2Exists(mysqlUserV2ResourceName),
					testAccCheckMDBMySQLUserV2ResourceIDField(mysqlUserV2ResourceName),
					resource.TestCheckResourceAttr(
						mysqlUserV2ResourceName, "name", "john",
					),
					resource.TestCheckResourceAttr(
						mysqlUserV2ResourceName,
						"deletion_protection_mode",
						"DELETION_PROTECTION_MODE_DISABLED",
					),
					waitForOperations(30),
					resource.TestCheckResourceAttr(
						mysqlUserV2ResourceName,
						"authentication_plugin",
						"MYSQL_NATIVE_PASSWORD",
					),
					resource.TestCheckResourceAttr(
						mysqlUserV2ResourceName,
						"connection_limits.0.max_questions_per_hour",
						"42",
					),
				),
			},
			mdbMySQLUserV2ImportStep(mysqlUserV2ResourceName),
			{
				Config: testAccMDBMySQLUserV2ConfigStep2(clusterName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMDBMySQLClusterHasUserV2(
						t, "john", "DELETION_PROTECTION_MODE_ENABLED",
					),
					resource.TestCheckResourceAttr(
						mysqlUserV2ResourceName,
						"deletion_protection_mode",
						"DELETION_PROTECTION_MODE_ENABLED",
					),
					resource.TestCheckResourceAttr(
						mysqlUserV2ResourceName,
						"connection_limits.0.max_questions_per_hour",
						"10",
					),
					resource.TestCheckResourceAttr(
						mysqlUserV2ResourceName,
						"authentication_plugin",
						"SHA256_PASSWORD",
					),
				),
			},
			mdbMySQLUserV2ImportStep(mysqlUserV2ResourceName),
			{
				Config: testAccMDBMySQLUserV2ConfigStep3(clusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMySQLUserV2Exists(mysqlUserV2ResourceName),
					testAccCheckMDBMySQLUserV2Exists(mysqlUserV2ResourceName1),
					resource.TestCheckResourceAttr(
						mysqlUserV2ResourceName1, "name", "mary",
					),
					resource.TestCheckResourceAttr(
						mysqlUserV2ResourceName1,
						"deletion_protection_mode",
						"DELETION_PROTECTION_MODE_INHERITED",
					),
					waitForOperations(30),
					resource.TestCheckResourceAttr(
						mysqlUserV2ResourceName,
						"deletion_protection_mode",
						"DELETION_PROTECTION_MODE_DISABLED",
					),
					waitForOperations(30),
				),
			},
			mdbMySQLUserV2ImportStep(mysqlUserV2ResourceName1),
			{
				Config: testAccMDBMySQLUserV2ConfigStep4(clusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBMySQLUserV2Exists(mysqlUserV2ResourceName),
					testAccCheckMDBMySQLUserV2Exists(mysqlUserV2ResourceName1),
					resource.TestCheckResourceAttr(
						mysqlUserV2ResourceName,
						"deletion_protection_mode",
						"DELETION_PROTECTION_MODE_DISABLED",
					),
					waitForOperations(30),
					resource.TestCheckResourceAttr(
						mysqlUserV2ResourceName1,
						"deletion_protection_mode",
						"DELETION_PROTECTION_MODE_DISABLED",
					),
					waitForOperations(30),
				),
			},
		},
	})
}

func testAccMDBMySQLUserV2Basic(clusterName, userName string) string {
	return fmt.Sprintf(mysqlUserV2VPCDependencies+`
resource "yandex_mdb_mysql_cluster_v2" "foo" {
  name        = "%s"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id
  version     = "8.0"

  hosts = {
    "host1" = {
      zone      = "ru-central1-a"
      subnet_id = yandex_vpc_subnet.foo.id
    }
  }

  resources {
    resource_preset_id = "s2.micro"
    disk_size          = 10
    disk_type_id       = "network-ssd"
  }
}

resource "yandex_mdb_mysql_database_v2" "testdb" {
  cluster_id = yandex_mdb_mysql_cluster_v2.foo.id
  name       = "testdb"
}

resource "yandex_mdb_mysql_user_v2" "testuser" {
  cluster_id = yandex_mdb_mysql_cluster_v2.foo.id
  name       = "%s"
  password   = "Password123!"

  permission {
    database_name = yandex_mdb_mysql_database_v2.testdb.name
    roles         = ["ALL"]
  }
}
`, clusterName, userName)
}

func testAccMDBMySQLUserV2WithDeletionProtection(clusterName, userName string) string {
	return fmt.Sprintf(mysqlUserV2VPCDependencies+`
resource "yandex_mdb_mysql_cluster_v2" "foo" {
  name        = "%s"
  environment = "PRESTABLE"
  network_id  = yandex_vpc_network.foo.id
  version     = "8.0"

  hosts = {
    "host1" = {
      zone      = "ru-central1-a"
      subnet_id = yandex_vpc_subnet.foo.id
    }
  }

  resources {
    resource_preset_id = "s2.micro"
    disk_size          = 10
    disk_type_id       = "network-ssd"
  }
}

resource "yandex_mdb_mysql_database_v2" "testdb" {
  cluster_id = yandex_mdb_mysql_cluster_v2.foo.id
  name       = "testdb"
}

resource "yandex_mdb_mysql_user_v2" "testuser" {
  cluster_id               = yandex_mdb_mysql_cluster_v2.foo.id
  name                     = "%s"
  password                 = "Password123!"
  deletion_protection_mode = "DELETION_PROTECTION_MODE_ENABLED"

  permission {
    database_name = yandex_mdb_mysql_database_v2.testdb.name
    roles         = ["ALL"]
  }
}
`, clusterName, userName)
}

func testAccMDBMySQLUserV2ConfigStep1(name string) string {
	return clusterConfigForUserTests(name) + `
resource "yandex_mdb_mysql_user_v2" "testuser" {
  cluster_id = yandex_mdb_mysql_cluster_v2.foo.id
  name       = "john"
  password   = "Password123!"

  permission {
    database_name = yandex_mdb_mysql_database_v2.testdb.name
    roles         = ["ALL", "INSERT"]
  }

  connection_limits {
    max_questions_per_hour = 42
  }

  global_permissions    = ["REPLICATION_SLAVE", "PROCESS", "FLUSH_OPTIMIZER_COSTS", "MDB_ADMIN", "SHOW_ROUTINE"]
  authentication_plugin = "MYSQL_NATIVE_PASSWORD"

  depends_on = [
    yandex_mdb_mysql_database_v2.testdb,
    yandex_mdb_mysql_database_v2.new_testdb,
  ]
}
`
}

func testAccMDBMySQLUserV2ConfigStep2(name string) string {
	return clusterConfigForUserTests(name) + `
resource "yandex_mdb_mysql_user_v2" "testuser" {
  cluster_id               = yandex_mdb_mysql_cluster_v2.foo.id
  name                     = "john"
  password                 = "Password123!"
  deletion_protection_mode = "DELETION_PROTECTION_MODE_ENABLED"

  permission {
    database_name = yandex_mdb_mysql_database_v2.testdb.name
    roles         = ["ALL", "DROP", "DELETE"]
  }

  permission {
    database_name = yandex_mdb_mysql_database_v2.new_testdb.name
    roles         = ["ALL", "INSERT"]
  }

  connection_limits {
    max_questions_per_hour   = 10
    max_updates_per_hour     = 20
    max_connections_per_hour = 30
    max_user_connections     = 40
  }

  global_permissions    = ["PROCESS"]
  authentication_plugin = "SHA256_PASSWORD"

  depends_on = [
    yandex_mdb_mysql_database_v2.testdb,
    yandex_mdb_mysql_database_v2.new_testdb,
  ]
}
`
}

func testAccMDBMySQLUserV2ConfigStep3(name string) string {
	return clusterConfigForUserTests(name) + `
resource "yandex_mdb_mysql_user_v2" "testuser" {
  cluster_id = yandex_mdb_mysql_cluster_v2.foo.id
  name       = "john"
  password   = "Password123!"

  permission {
    database_name = yandex_mdb_mysql_database_v2.testdb.name
    roles         = ["ALL", "DROP", "DELETE"]
  }

  permission {
    database_name = yandex_mdb_mysql_database_v2.new_testdb.name
    roles         = ["ALL", "INSERT"]
  }

  connection_limits {
    max_questions_per_hour   = 10
    max_updates_per_hour     = 20
    max_connections_per_hour = 30
    max_user_connections     = 40
  }

  global_permissions    = ["PROCESS"]
  authentication_plugin = "SHA256_PASSWORD"

  depends_on = [
    yandex_mdb_mysql_database_v2.testdb,
    yandex_mdb_mysql_database_v2.new_testdb,
  ]
}

resource "yandex_mdb_mysql_user_v2" "testuser1" {
  cluster_id               = yandex_mdb_mysql_cluster_v2.foo.id
  name                     = "mary"
  generate_password        = true
  deletion_protection_mode = "DELETION_PROTECTION_MODE_INHERITED"

  permission {
    database_name = yandex_mdb_mysql_database_v2.new_testdb.name
    roles         = ["ALTER", "CREATE", "INSERT", "DROP", "DELETE"]
  }

  depends_on = [
    yandex_mdb_mysql_user_v2.testuser,
    yandex_mdb_mysql_database_v2.testdb,
    yandex_mdb_mysql_database_v2.new_testdb,
  ]
}
`
}

func testAccMDBMySQLUserV2ConfigStep4(name string) string {
	return clusterConfigForUserTests(name) + `
resource "yandex_mdb_mysql_user_v2" "testuser" {
  cluster_id = yandex_mdb_mysql_cluster_v2.foo.id
  name       = "john"
  password   = "Password123!"

  permission {
    database_name = yandex_mdb_mysql_database_v2.testdb.name
    roles         = ["ALL", "DROP", "DELETE"]
  }

  permission {
    database_name = yandex_mdb_mysql_database_v2.new_testdb.name
    roles         = ["ALL", "INSERT"]
  }

  connection_limits {
    max_questions_per_hour   = 10
    max_updates_per_hour     = 20
    max_connections_per_hour = 30
    max_user_connections     = 40
  }

  global_permissions    = ["PROCESS"]
  authentication_plugin = "SHA256_PASSWORD"

  depends_on = [
    yandex_mdb_mysql_database_v2.testdb,
    yandex_mdb_mysql_database_v2.new_testdb,
  ]
}

resource "yandex_mdb_mysql_user_v2" "testuser1" {
  cluster_id               = yandex_mdb_mysql_cluster_v2.foo.id
  name                     = "mary"
  generate_password        = true
  deletion_protection_mode = "DELETION_PROTECTION_MODE_DISABLED"

  permission {
    database_name = yandex_mdb_mysql_database_v2.new_testdb.name
    roles         = ["ALTER", "CREATE", "INSERT", "DROP", "DELETE"]
  }

  depends_on = [
    yandex_mdb_mysql_user_v2.testuser,
    yandex_mdb_mysql_database_v2.testdb,
    yandex_mdb_mysql_database_v2.new_testdb,
  ]
}
`
}
