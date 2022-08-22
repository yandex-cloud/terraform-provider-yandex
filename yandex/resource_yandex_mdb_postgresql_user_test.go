package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
)

const (
	pgClusterResourceName     = "yandex_mdb_postgresql_cluster.foo"
	pgUserResourceNameAlice   = "yandex_mdb_postgresql_user.alice"
	pgUserResourceNameBob     = "yandex_mdb_postgresql_user.bob"
	pgUserResourceNameCharlie = "yandex_mdb_postgresql_user.charlie"
)

// Test that a PostgreSQL User can be created, updated and destroyed
func TestAccMDBPostgreSQLUser_full(t *testing.T) {
	t.Parallel()
	clusterName := acctest.RandomWithPrefix("tf-postgresql")
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBPostgreSQLUserConfigStep1(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(pgUserResourceNameAlice, "name", "alice"),
					resource.TestCheckResourceAttr(pgUserResourceNameAlice, "login", "true"),
					testAccCheckMDBPostgreSQLUserHasGrants(t, "alice", []string{"mdb_admin", "mdb_replication"}),
					resource.TestCheckResourceAttr(pgUserResourceNameAlice, "conn_limit", "50"),
					testAccCheckMDBPostgreSQLUserHasSettings(t, "alice", map[string]interface{}{"default_transaction_isolation": postgresql.UserSettings_TRANSACTION_ISOLATION_READ_COMMITTED, "log_min_duration_statement": int64(5000)}),
				),
			},
			mdbPostgreSQLUserImportStep(pgUserResourceNameAlice),
			{
				Config: testAccMDBPostgreSQLUserConfigStep2(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(pgUserResourceNameBob, "name", "bob"),
					testAccCheckMDBPostgreSQLUserHasPermission(t, "bob", []string{"testdb"}),
				),
			},
			mdbPostgreSQLUserImportStep(pgUserResourceNameAlice),
			mdbPostgreSQLUserImportStep(pgUserResourceNameBob),
			{
				Config: testAccMDBPostgreSQLUserConfigStep3(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(pgUserResourceNameAlice, "name", "alice"),
					resource.TestCheckResourceAttr(pgUserResourceNameAlice, "conn_limit", "42"),
					testAccCheckMDBPostgreSQLUserHasPermission(t, "alice", []string{"testdb"}),
					testAccCheckMDBPostgreSQLUserHasSettings(t, "alice", map[string]interface{}{"default_transaction_isolation": postgresql.UserSettings_TRANSACTION_ISOLATION_READ_UNCOMMITTED, "log_min_duration_statement": int64(1234)}),
				),
			},
			mdbPostgreSQLUserImportStep(pgUserResourceNameAlice),
			{
				Config: testAccMDBPostgreSQLUserConfigStep4(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(pgUserResourceNameCharlie, "name", "charlie"),
					resource.TestCheckResourceAttr(pgUserResourceNameCharlie, "login", "false"),
					resource.TestCheckResourceAttr(pgUserResourceNameCharlie, "conn_limit", "0"),
				),
			},
			mdbPostgreSQLUserImportStep(pgUserResourceNameCharlie),
		},
	})
}

func mdbPostgreSQLUserImportStep(name string) resource.TestStep {
	return resource.TestStep{
		ResourceName:      name,
		ImportState:       true,
		ImportStateVerify: true,
		ImportStateVerifyIgnore: []string{
			"password", // password is not returned
		},
	}
}

func testAccLoadPostgreSQLUser(s *terraform.State, username string) (*postgresql.User, error) {
	rs, ok := s.RootModule().Resources[pgResource]

	if !ok {
		return nil, fmt.Errorf("resource %q not found", pgUserResourceNameAlice)
	}
	if rs.Primary.ID == "" {
		return nil, fmt.Errorf("no ID is set")
	}

	config := testAccProvider.Meta().(*Config)
	return config.sdk.MDB().PostgreSQL().User().Get(context.Background(), &postgresql.GetUserRequest{
		ClusterId: rs.Primary.ID,
		UserName:  username,
	})
}

func testAccCheckMDBPostgreSQLUserHasGrants(t *testing.T, username string, expected []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		user, err := testAccLoadPostgreSQLUser(s, username)
		if err != nil {
			return err
		}

		assert.Equal(t, user.Grants, expected)
		return nil
	}
}
func testAccCheckMDBPostgreSQLUserHasSettings(t *testing.T, username string, expected map[string]interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[pgClusterResourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", pgClusterResourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		user, err := config.sdk.MDB().PostgreSQL().User().Get(context.Background(), &postgresql.GetUserRequest{
			ClusterId: rs.Primary.ID,
			UserName:  username,
		})

		if err != nil {
			return err
		}

		assert.Equal(t, user.Settings.DefaultTransactionIsolation, expected["default_transaction_isolation"])
		assert.Equal(t, user.Settings.LogMinDurationStatement.GetValue(), expected["log_min_duration_statement"])

		return nil
	}
}

func testAccCheckMDBPostgreSQLUserHasPermission(t *testing.T, username string, expected []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		user, err := testAccLoadPostgreSQLUser(s, username)
		if err != nil {
			return err
		}
		permissions := []string{}
		for _, permission := range user.Permissions {
			permissions = append(permissions, permission.DatabaseName)
		}

		assert.Equal(t, permissions, expected)

		return nil
	}
}

func testAccMDBPostgreSQLUserConfigStep0(name string) string {
	return fmt.Sprintf(pgVPCDependencies+`
resource "yandex_mdb_postgresql_cluster" "foo" {
	name        = "%s"
	description = "PostgreSQL User Terraform Test"
	environment = "PRODUCTION"
	network_id  = "${yandex_vpc_network.mdb-pg-test-net.id}"

	config {
	    version = 14
	    resources {
		  resource_preset_id = "s2.micro"
		  disk_size          = 10
		  disk_type_id       = "network-ssd"
	    }
	}

	host {
		name      = "a"
		zone      = "ru-central1-a"
		subnet_id  = yandex_vpc_subnet.mdb-pg-test-subnet-a.id
	  }
}

resource "yandex_mdb_postgresql_database" "testdb" {
	cluster_id = yandex_mdb_postgresql_cluster.foo.id
	name       = "testdb"
	owner      = yandex_mdb_postgresql_user.alice.name
	lc_collate = "en_US.UTF-8"
	lc_type    = "en_US.UTF-8"
}
`, name)
}

// Create cluster, user and database
func testAccMDBPostgreSQLUserConfigStep1(name string) string {
	return testAccMDBPostgreSQLUserConfigStep0(name) + `
resource "yandex_mdb_postgresql_user" "alice" {
	cluster_id = yandex_mdb_postgresql_cluster.foo.id
	name       = "alice"
	password   = "mysecurepassword"
	login      = true
	grants     = ["mdb_admin", "mdb_replication"]
	conn_limit = 50
	settings = {
		default_transaction_isolation = "read committed"
		log_min_duration_statement    = 5000
	}
}`
}

// Create another user and give permission to database
func testAccMDBPostgreSQLUserConfigStep2(name string) string {
	return testAccMDBPostgreSQLUserConfigStep1(name) + `
resource "yandex_mdb_postgresql_user" "bob" {
	cluster_id = yandex_mdb_postgresql_cluster.foo.id
	name       = "bob"
	password   = "mysecurepassword"
    permission {
		database_name = "testdb"
	}
}`
}

// Change Alice's settings and conn_limit
func testAccMDBPostgreSQLUserConfigStep3(name string) string {
	return testAccMDBPostgreSQLUserConfigStep0(name) + `
resource "yandex_mdb_postgresql_user" "alice" {
	cluster_id = yandex_mdb_postgresql_cluster.foo.id
	name       = "alice"
	password   = "mysecurepassword"
	permission {
		database_name = "testdb"
	}
    
	conn_limit = 42
	settings = {
		default_transaction_isolation = "read uncommitted"
		log_min_duration_statement    = 1234
	}
}`
}

// Check login and conn_limit. Bug report: https://github.com/KazanExpress/yc-tf-bugreports/tree/master/bugs/postgres-3
func testAccMDBPostgreSQLUserConfigStep4(name string) string {
	return testAccMDBPostgreSQLUserConfigStep3(name) + `
resource "yandex_mdb_postgresql_user" "charlie" {
	cluster_id = yandex_mdb_postgresql_cluster.foo.id
	name       = "charlie"
	password   = "password"
    
	login      = false
	conn_limit = 0
}`
}
