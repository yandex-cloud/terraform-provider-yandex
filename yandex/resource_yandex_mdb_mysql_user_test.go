package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	mysqlUserResourceJohn = "yandex_mdb_mysql_user.john"
	mysqlUserResourceMary = "yandex_mdb_mysql_user.mary"
	mysqlUserResourceJane = "yandex_mdb_mysql_user.jane"
)

// Test that a MySQL User can be created, updated and destroyed
func TestAccMDBMySQLUser_full(t *testing.T) {
	t.Parallel()
	clusterName := acctest.RandomWithPrefix("tf-mysql")
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviderFactoriesV6,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBMySQLUserConfigStep1(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(mysqlUserResourceJohn, "name", "john"),
					testAccCheckMDBMysqlClusterHasUsers(mysqlResource, map[string][]MockPermission{"john": {MockPermission{"testdb", []string{"ALL", "INSERT"}}}}),
					resource.TestCheckResourceAttr(mysqlUserResourceJohn, "connection_limits.0.max_questions_per_hour", "42"),
					resource.TestCheckResourceAttr(mysqlUserResourceJohn, "global_permissions.#", "5"),
					resource.TestCheckResourceAttr(mysqlUserResourceJohn, "authentication_plugin", "MYSQL_NATIVE_PASSWORD"),
					resource.TestCheckResourceAttr(mysqlUserResourceJohn, "generate_password", "false"),
					resource.TestCheckResourceAttr(mysqlUserResourceJohn, "connection_manager.%", "1"),
				),
			},
			mdbMySQLUserImportStep(mysqlUserResourceJohn),
			{
				Config: testAccMDBMySQLUserConfigStep2(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(mysqlUserResourceJohn, "name", "john"),
					testAccCheckMDBMysqlClusterHasUsers(mysqlResource, map[string][]MockPermission{"john": {MockPermission{"testdb", []string{"ALL", "DROP", "DELETE"}}, MockPermission{"new_testdb", []string{"ALL", "INSERT"}}}}),
					resource.TestCheckResourceAttr(mysqlUserResourceJohn, "connection_limits.0.max_questions_per_hour", "10"),
					resource.TestCheckResourceAttr(mysqlUserResourceJohn, "connection_limits.0.max_updates_per_hour", "20"),
					resource.TestCheckResourceAttr(mysqlUserResourceJohn, "connection_limits.0.max_connections_per_hour", "30"),
					resource.TestCheckResourceAttr(mysqlUserResourceJohn, "connection_limits.0.max_user_connections", "40"),
					resource.TestCheckResourceAttr(mysqlUserResourceJohn, "global_permissions.#", "1"),
					resource.TestCheckResourceAttr(mysqlUserResourceJohn, "authentication_plugin", "SHA256_PASSWORD"),
					resource.TestCheckResourceAttr(mysqlUserResourceJohn, "generate_password", "false"),
					resource.TestCheckResourceAttr(mysqlUserResourceJohn, "connection_manager.%", "1"),
				),
			},
			mdbMySQLUserImportStep(mysqlUserResourceJohn),
			{
				Config: testAccMDBMySQLUserConfigStep3(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(mysqlUserResourceMary, "name", "mary"),
					testAccCheckMDBMysqlClusterHasUsers(mysqlResource, map[string][]MockPermission{
						"john": {MockPermission{"testdb", []string{"ALL", "DROP", "DELETE"}}, MockPermission{"new_testdb", []string{"ALL", "INSERT"}}},
						"mary": {MockPermission{"new_testdb", []string{"ALTER", "CREATE", "INSERT", "DROP", "DELETE"}}},
					}),
					resource.TestCheckResourceAttr(mysqlUserResourceMary, "connection_limits.#", "0"),
					resource.TestCheckResourceAttr(mysqlUserResourceMary, "global_permissions.#", "0"),
					resource.TestCheckResourceAttr(mysqlUserResourceMary, "generate_password", "true"),
					resource.TestCheckResourceAttr(mysqlUserResourceMary, "connection_manager.%", "1"),
				),
			},
			mdbMySQLUserImportStep(mysqlUserResourceMary),
			{
				Config: testAccMDBMySQLUserConfigStep4(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(mysqlUserResourceJane, "name", "jane"),
					testAccCheckMDBMysqlClusterHasUsers(mysqlResource, map[string][]MockPermission{
						"john": {MockPermission{"testdb", []string{"ALL", "DROP", "DELETE"}}, MockPermission{"new_testdb", []string{"ALL", "INSERT"}}},
						"jane": {MockPermission{"new_testdb", []string{"ALL"}}},
					}),
					resource.TestCheckResourceAttr(mysqlUserResourceJane, "connection_limits.#", "0"),
					resource.TestCheckResourceAttr(mysqlUserResourceJane, "global_permissions.#", "0"),
					resource.TestCheckResourceAttr(mysqlUserResourceJane, "connection_manager.%", "0"),
					resource.TestCheckResourceAttr(mysqlUserResourceJane, "authentication_plugin", "MDB_IAMPROXY_AUTH"),
				),
			},
			mdbMySQLUserImportStep(mysqlUserResourceJane),
		},
	})
}

func mdbMySQLUserImportStep(name string) resource.TestStep {
	return resource.TestStep{
		ResourceName:      name,
		ImportState:       true,
		ImportStateVerify: true,
		ImportStateVerifyIgnore: []string{
			"password", "generate_password", // not returned
		},
	}
}

func testAccMDBMySQLUserConfigStep0(name string) string {
	return fmt.Sprintf(mysqlVPCDependencies+`
resource "yandex_mdb_mysql_cluster" "foo" {
	name        = "%s"
	description = "MySQL User Terraform Test"
	environment = "PRESTABLE"
	network_id  = yandex_vpc_network.foo.id
	version     = "8.0"
	
	resources {
	  resource_preset_id = "s2.micro"
	  disk_type_id       = "network-ssd"
	  disk_size          = 24
	}

	host {
	  zone      = "ru-central1-d"
	  subnet_id = yandex_vpc_subnet.foo_c.id
	}
}

resource "yandex_mdb_mysql_database" "testdb" {
	cluster_id = yandex_mdb_mysql_cluster.foo.id
	name       = "testdb"
}
`, name)
}

// Create user
func testAccMDBMySQLUserConfigStep1(clusterName string) string {
	return testAccMDBMySQLUserConfigStep0(clusterName) + `
resource "yandex_mdb_mysql_user" "john" {
	cluster_id = yandex_mdb_mysql_cluster.foo.id
    name       = "john"
    password   = "password"

    permission {
      database_name = yandex_mdb_mysql_database.testdb.name
      roles         = ["ALL", "INSERT"]
    }

    connection_limits {
      max_questions_per_hour = 42
    }

    global_permissions = ["REPLICATION_SLAVE", "PROCESS", "FLUSH_OPTIMIZER_COSTS", "MDB_ADMIN", "SHOW_ROUTINE"]

    authentication_plugin = "MYSQL_NATIVE_PASSWORD"
}
`
}

// Update the old user
func testAccMDBMySQLUserConfigStep2(clusterName string) string {
	return testAccMDBMySQLUserConfigStep0(clusterName) + `
resource "yandex_mdb_mysql_database" "new_testdb" {
	cluster_id = yandex_mdb_mysql_cluster.foo.id
	name       = "new_testdb"
}

resource "yandex_mdb_mysql_user" "john" {
	cluster_id = yandex_mdb_mysql_cluster.foo.id
    name       = "john"
    password   = "password"

    permission {
      database_name = yandex_mdb_mysql_database.testdb.name
      roles         = ["ALL", "DROP", "DELETE"]
    }

    permission {
      database_name = yandex_mdb_mysql_database.new_testdb.name
      roles         = ["ALL", "INSERT"]
    }

	connection_limits {
	  max_questions_per_hour   = 10
	  max_updates_per_hour     = 20
	  max_connections_per_hour = 30
	  max_user_connections     = 40
	}

	global_permissions = ["PROCESS"]

	authentication_plugin = "SHA256_PASSWORD"
}
`
}

// Create a new user
func testAccMDBMySQLUserConfigStep3(clusterName string) string {
	return testAccMDBMySQLUserConfigStep2(clusterName) + `
resource "yandex_mdb_mysql_user" "mary" {
	cluster_id = yandex_mdb_mysql_cluster.foo.id
    name       = "mary"
	generate_password = "true"

    permission {
      database_name = yandex_mdb_mysql_database.new_testdb.name
	  roles         = ["ALTER", "CREATE", "INSERT", "DROP", "DELETE"]
    }
}
`
}

// Create a new user
func testAccMDBMySQLUserConfigStep4(clusterName string) string {
	return testAccMDBMySQLUserConfigStep2(clusterName) + `
resource "yandex_mdb_mysql_user" "jane" {
	cluster_id = yandex_mdb_mysql_cluster.foo.id
    name       = "jane"

    permission {
      database_name = yandex_mdb_mysql_database.new_testdb.name
      roles         = ["ALL"]
    }

    authentication_plugin = "MDB_IAMPROXY_AUTH"
}
`
}
