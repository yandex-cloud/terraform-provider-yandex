package yandex

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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
	clusterName := acctest.RandomWithPrefix("tf-postgresql-user")
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBPostgreSQLUserConfigStep1(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(pgUserResourceNameAlice, "name", "alice"),
					resource.TestCheckResourceAttr(pgUserResourceNameAlice, "login", "true"),
					resource.TestCheckResourceAttr(pgUserResourceNameAlice, "deletion_protection", "unspecified"),
					resource.TestCheckResourceAttr(pgUserResourceNameAlice, "generate_password", "false"),
					resource.TestCheckResourceAttr(pgUserResourceNameAlice, "connection_manager.%", "1"),
					testAccCheckMDBPostgreSQLUserHasGrants(t, "alice", []string{"mdb_admin", "mdb_replication"}),
					resource.TestCheckResourceAttr(pgUserResourceNameAlice, "conn_limit", "50"),
					testAccCheckMDBPostgreSQLUserHasSettings(t, "alice", map[string]interface{}{"default_transaction_isolation": postgresql.UserSettings_TRANSACTION_ISOLATION_READ_COMMITTED, "log_min_duration_statement": int64(5000), "pool_mode": postgresql.UserSettings_TRANSACTION, "catchup_timeout": int64(350)}),
				),
			},
			mdbPostgreSQLUserImportStep(pgUserResourceNameAlice),
			{
				Config: testAccMDBPostgreSQLUserConfigStep2(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(pgUserResourceNameBob, "name", "bob"),
					resource.TestCheckResourceAttr(pgUserResourceNameBob, "deletion_protection", "false"),
					resource.TestCheckResourceAttr(pgUserResourceNameBob, "generate_password", "false"),
					resource.TestCheckResourceAttr(pgUserResourceNameBob, "connection_manager.%", "1"),
					testAccCheckMDBPostgreSQLUserHasPermission(t, "bob", []string{"testdb"}),
				),
			},
			mdbPostgreSQLUserImportStep(pgUserResourceNameAlice),
			mdbPostgreSQLUserImportStep(pgUserResourceNameBob),
			{
				Config: testAccMDBPostgreSQLUserConfigStep3(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(pgUserResourceNameBob, "name", "bob"),
					resource.TestCheckResourceAttr(pgUserResourceNameBob, "deletion_protection", "false"),
					resource.TestCheckResourceAttr(pgUserResourceNameBob, "generate_password", "false"),
					resource.TestCheckResourceAttr(pgUserResourceNameBob, "connection_manager.%", "1"),
					testAccCheckMDBPostgreSQLUserHasPermission(t, "bob", []string{}),
				),
			},
			mdbPostgreSQLUserImportStep(pgUserResourceNameAlice),
			{
				Config: testAccMDBPostgreSQLUserConfigStep4(clusterName, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(pgUserResourceNameAlice, "name", "alice"),
					resource.TestCheckResourceAttr(pgUserResourceNameAlice, "conn_limit", "42"),
					resource.TestCheckResourceAttr(pgUserResourceNameAlice, "deletion_protection", "true"),
					resource.TestCheckResourceAttr(pgUserResourceNameAlice, "generate_password", "false"),
					resource.TestCheckResourceAttr(pgUserResourceNameAlice, "connection_manager.%", "1"),
					testAccCheckMDBPostgreSQLUserHasPermission(t, "alice", []string{"testdb"}),
					testAccCheckMDBPostgreSQLUserHasSettings(t, "alice", map[string]interface{}{"default_transaction_isolation": postgresql.UserSettings_TRANSACTION_ISOLATION_READ_UNCOMMITTED, "log_min_duration_statement": int64(1234), "pool_mode": postgresql.UserSettings_SESSION, "statement_timeout": int64(0)}),
				),
			},
			mdbPostgreSQLUserImportStep(pgUserResourceNameAlice),
			{
				Config: testAccMDBPostgreSQLUserConfigStep4(clusterName, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(pgUserResourceNameAlice, "name", "alice"),
					resource.TestCheckResourceAttr(pgUserResourceNameAlice, "deletion_protection", "false"),
					resource.TestCheckResourceAttr(pgUserResourceNameAlice, "generate_password", "false"),
					resource.TestCheckResourceAttr(pgUserResourceNameAlice, "connection_manager.%", "1"),
				),
			},
			mdbPostgreSQLUserImportStep(pgUserResourceNameAlice),
			{
				Config: testAccMDBPostgreSQLUserConfigStep5(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(pgUserResourceNameAlice, "name", "alice"),
					resource.TestCheckResourceAttr(pgUserResourceNameAlice, "deletion_protection", "false"),
					resource.TestCheckResourceAttr(pgUserResourceNameAlice, "generate_password", "false"),
					resource.TestCheckResourceAttr(pgUserResourceNameAlice, "connection_manager.%", "1"),
					testAccCheckMDBPostgreSQLUserHasSettings(t, "alice", map[string]interface{}{"default_transaction_isolation": postgresql.UserSettings_TRANSACTION_ISOLATION_READ_UNCOMMITTED, "log_min_duration_statement": int64(1234), "pool_mode": postgresql.UserSettings_SESSION}),
					testAccCheckMDBPostgreSQLUserHasNoSettings(t, "alice", []string{"statement_timeout"}),
				),
			},
			mdbPostgreSQLUserImportStep(pgUserResourceNameAlice),
			{
				Config: testAccMDBPostgreSQLUserConfigStep6(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(pgUserResourceNameCharlie, "name", "charlie"),
					resource.TestCheckResourceAttr(pgUserResourceNameCharlie, "login", "false"),
					resource.TestCheckResourceAttr(pgUserResourceNameCharlie, "conn_limit", "0"),
					resource.TestCheckResourceAttr(pgUserResourceNameCharlie, "generate_password", "true"),
					resource.TestCheckResourceAttr(pgUserResourceNameCharlie, "connection_manager.%", "1"),
				),
			},
			mdbPostgreSQLUserImportStep(pgUserResourceNameCharlie),
		},
	})
}

// Test that a PostgreSQL User can't be created with grants = [""]
func TestAccMDBPostgreSQLUserIncorrectGrants(t *testing.T) {
	t.Parallel()
	clusterName := acctest.RandomWithPrefix("tf-postgresql-user")
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBPostgreSQLUserConfigStep0(clusterName) + `
resource "yandex_mdb_postgresql_user" "alice" {
	cluster_id = yandex_mdb_postgresql_cluster.foo.id
	name       = "alice"
	password   = "mysecureP@ssw0rd"
	login      = true
	grants     = [""]
	conn_limit = 50
	settings = {
		default_transaction_isolation = "read committed"
		log_min_duration_statement    = 5000
	}
}`,
				ExpectError: regexp.MustCompile(".*expected .*? to not be an empty string.*"),
			},
		},
	})
}

func mdbPostgreSQLUserImportStep(name string) resource.TestStep {
	return resource.TestStep{
		ResourceName:      name,
		ImportState:       true,
		ImportStateVerify: true,
		ImportStateVerifyIgnore: []string{
			"password", "generate_password", // password and generate_password is not returned
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

		val, ok := expected["default_transaction_isolation"]
		if ok {
			assert.Equal(t, val, user.Settings.DefaultTransactionIsolation, "setting name: default_transaction_isolation")
		}
		val, ok = expected["pool_mode"]
		if ok {
			assert.Equal(t, val, user.Settings.PoolMode, "setting name: pool_mode")
		}
		val, ok = expected["log_min_duration_statement"]
		if ok {
			assert.Equal(t, val, user.Settings.LogMinDurationStatement.GetValue(), "setting name: log_min_duration_statement")
		}
		val, ok = expected["catchup_timeout"]
		if ok {
			assert.Equal(t, val, user.Settings.CatchupTimeout.GetValue(), "setting name: catchup_timeout")
		}
		val, ok = expected["statement_timeout"]
		if ok {
			assert.Equal(t, val, user.Settings.StatementTimeout.GetValue(), "setting name: statement_timeout")
		}
		val, ok = expected["lock_timeout"]
		if ok {
			assert.Equal(t, val, user.Settings.LockTimeout.GetValue(), "setting name: lock_timeout")
		}
		val, ok = expected["synchronous_commit"]
		if ok {
			assert.Equal(t, val, user.Settings.SynchronousCommit, "setting name: synchronous_commit")
		}
		val, ok = expected["temp_file_limit"]
		if ok {
			assert.Equal(t, val, user.Settings.TempFileLimit.GetValue(), "setting name: temp_file_limit")
		}
		val, ok = expected["log_statement"]
		if ok {
			assert.Equal(t, val, user.Settings.LogStatement, "setting name: log_statement")
		}
		val, ok = expected["prepared_statements_pooling"]
		if ok {
			assert.Equal(t, val, user.Settings.PreparedStatementsPooling.GetValue(), "setting name: prepared_statements_pooling")
		}
		val, ok = expected["wal_sender_timeout"]
		if ok {
			assert.Equal(t, val, user.Settings.WalSenderTimeout.GetValue(), "setting name: wal_sender_timeout")
		}
		val, ok = expected["idle_in_transaction_session_timeout"]
		if ok {
			assert.Equal(t, val, user.Settings.IdleInTransactionSessionTimeout.GetValue(), "setting name: idle_in_transaction_session_timeout")
		}
		val, ok = expected["pgaudit"]
		if ok {
			assert.Equal(t, val, user.Settings.Pgaudit, "setting name: pgaudit")
		}

		return nil
	}
}

func testAccCheckMDBPostgreSQLUserHasNoSettings(t *testing.T, username string, notExpected []string) resource.TestCheckFunc {
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
		isContains := slices.Contains(notExpected, "default_transaction_isolation")
		if isContains {
			assert.Nil(t, user.Settings.DefaultTransactionIsolation)
		}
		isContains = slices.Contains(notExpected, "pool_mode")
		if isContains {
			assert.Nil(t, user.Settings.PoolMode)
		}
		isContains = slices.Contains(notExpected, "log_min_duration_statement")
		if isContains {
			assert.Nil(t, user.Settings.LogMinDurationStatement)
		}
		isContains = slices.Contains(notExpected, "catchup_timeout")
		if isContains {
			assert.Nil(t, user.Settings.CatchupTimeout)
		}
		isContains = slices.Contains(notExpected, "statement_timeout")
		if isContains {
			assert.Nil(t, user.Settings.StatementTimeout)
		}
		isContains = slices.Contains(notExpected, "lock_timeout")
		if isContains {
			assert.Nil(t, user.Settings.LockTimeout)
		}
		isContains = slices.Contains(notExpected, "synchronous_commit")
		if isContains {
			assert.Nil(t, user.Settings.SynchronousCommit)
		}
		isContains = slices.Contains(notExpected, "temp_file_limit")
		if isContains {
			assert.Nil(t, user.Settings.TempFileLimit)
		}
		isContains = slices.Contains(notExpected, "log_statement")
		if isContains {
			assert.Nil(t, user.Settings.LogStatement)
		}
		isContains = slices.Contains(notExpected, "prepared_statements_pooling")
		if isContains {
			assert.Nil(t, user.Settings.PreparedStatementsPooling)
		}
		isContains = slices.Contains(notExpected, "wal_sender_timeout")
		if isContains {
			assert.Nil(t, user.Settings.WalSenderTimeout)
		}
		isContains = slices.Contains(notExpected, "idle_in_transaction_session_timeout")
		if isContains {
			assert.Nil(t, user.Settings.IdleInTransactionSessionTimeout)
		}
		isContains = slices.Contains(notExpected, "pgaudit")
		if isContains {
			assert.Nil(t, user.Settings.Pgaudit)
		}

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
	    version = 16
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
	password   = "mysecureP@ssw0rd"
	login      = true
	grants     = ["mdb_admin", "mdb_replication"]
	conn_limit = 50
	settings = {
		default_transaction_isolation = "read committed"
		log_min_duration_statement    = 5000
		pool_mode                     = "transaction"
		catchup_timeout				  = 350	
	}
}`
}

// Create another user and give permission to database
func testAccMDBPostgreSQLUserConfigStep2(name string) string {
	return testAccMDBPostgreSQLUserConfigStep1(name) + `
resource "yandex_mdb_postgresql_user" "bob" {
	cluster_id = yandex_mdb_postgresql_cluster.foo.id
	name       = "bob"
	password   = "mysecureP@ssw0rd"
    permission {
		database_name = "testdb"
	}
	deletion_protection = false
}`
}

// Drop permissions field. Bug report: changing permissions works but dropping permissions field do nothing
func testAccMDBPostgreSQLUserConfigStep3(name string) string {
	return testAccMDBPostgreSQLUserConfigStep1(name) + `
resource "yandex_mdb_postgresql_user" "bob" {
	cluster_id = yandex_mdb_postgresql_cluster.foo.id
	name       = "bob"
	password   = "mysecureP@ssw0rd"
	deletion_protection = false
}`
}

// Change Alice's settings and conn_limit
func testAccMDBPostgreSQLUserConfigStep4(name string, deletionProtection bool) string {
	return testAccMDBPostgreSQLUserConfigStep0(name) + fmt.Sprintf(`
resource "yandex_mdb_postgresql_user" "alice" {
	cluster_id = yandex_mdb_postgresql_cluster.foo.id
	name       = "alice"
	password   = "mysecureP@ssw0rd"
    
	conn_limit = 42
	settings = {
		default_transaction_isolation = "read uncommitted"
		log_min_duration_statement    = 1234
		pool_mode                     = "session"
		statement_timeout 			  = 0
	}
	deletion_protection = "%v"
}`, deletionProtection)
}

// Change Alice's settings and conn_limit
func testAccMDBPostgreSQLUserConfigStep5(name string) string {
	return testAccMDBPostgreSQLUserConfigStep0(name) + `
resource "yandex_mdb_postgresql_user" "alice" {
	cluster_id = yandex_mdb_postgresql_cluster.foo.id
	name       = "alice"
	password   = "mysecureP@ssw0rd"
    
	conn_limit = 42
	settings = {
		default_transaction_isolation = "read uncommitted"
		log_min_duration_statement    = 1234
		pool_mode                     = "session"
	}
	deletion_protection = false
}`
}

// Check login and conn_limit. Bug report: https://github.com/KazanExpress/yc-tf-bugreports/tree/master/bugs/postgres-3
func testAccMDBPostgreSQLUserConfigStep6(name string) string {
	return testAccMDBPostgreSQLUserConfigStep5(name) + `
resource "yandex_mdb_postgresql_user" "charlie" {
	cluster_id = yandex_mdb_postgresql_cluster.foo.id
	name       = "charlie"
	generate_password = true
    
	login      = false
	conn_limit = 0
}`
}
