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
	pgDatabaseResourceName  = "yandex_mdb_postgresql_database.testdb"
	pgDatabaseResourceName1 = "yandex_mdb_postgresql_database.testdb1"
)

// Test that a PostgreSQL Database can be created, updated and destroyed
func TestAccMDBPostgreSQLDatabase_full(t *testing.T) {
	t.Parallel()
	clusterName := acctest.RandomWithPrefix("tf-postgresql")
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBPostgreSQLDatabaseConfigStep1(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(pgDatabaseResourceName, "name", "testdb"),
					resource.TestCheckResourceAttr(pgDatabaseResourceName, "owner", "alice"),
					resource.TestCheckResourceAttr(pgDatabaseResourceName, "lc_collate", "en_US.UTF-8"),
					resource.TestCheckResourceAttr(pgDatabaseResourceName, "lc_type", "en_US.UTF-8"),
					testAccCheckMDBPostgreSQLClusterHasDatabase(t, "testdb", make([]string, 0)),
				),
			},
			mdbPostgreSQLDatabaseImportStep(pgDatabaseResourceName),
			{
				Config: testAccMDBPostgreSQLDatabaseConfigStep2(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(pgDatabaseResourceName, "name", "testdb"),
					testAccCheckMDBPostgreSQLClusterHasDatabase(t, "testdb", []string{"uuid-ossp", "xml2"}),
				),
			},
			mdbPostgreSQLDatabaseImportStep(pgDatabaseResourceName),
			{
				Config: testAccMDBPostgreSQLDatabaseConfigStep3(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(pgDatabaseResourceName1, "name", "testdb1"),
					resource.TestCheckResourceAttr(pgDatabaseResourceName1, "owner", "alice"),
					resource.TestCheckResourceAttr(pgDatabaseResourceName1, "template_db", "testdb"),
					resource.TestCheckResourceAttr(pgDatabaseResourceName1, "lc_collate", "en_US.UTF-8"),
					resource.TestCheckResourceAttr(pgDatabaseResourceName1, "lc_type", "en_US.UTF-8"),
					testAccCheckMDBPostgreSQLClusterHasDatabase(t, "testdb1", make([]string, 0)),
				),
			},
			mdbPostgreSQLDatabaseImportStep(pgDatabaseResourceName1),
		},
	})
}

func mdbPostgreSQLDatabaseImportStep(name string) resource.TestStep {
	return resource.TestStep{
		ResourceName:      name,
		ImportState:       true,
		ImportStateVerify: true,
	}
}

func testAccLoadPostgreSQLDatabase(s *terraform.State, dbname string) (*postgresql.Database, error) {
	rs, ok := s.RootModule().Resources[pgResource]

	if !ok {
		return nil, fmt.Errorf("resource %q not found", pgDatabaseResourceName)
	}
	if rs.Primary.ID == "" {
		return nil, fmt.Errorf("no ID is set")
	}

	config := testAccProvider.Meta().(*Config)
	return config.sdk.MDB().PostgreSQL().Database().Get(context.Background(), &postgresql.GetDatabaseRequest{
		ClusterId:    rs.Primary.ID,
		DatabaseName: dbname,
	})
}

func testAccCheckMDBPostgreSQLClusterHasDatabase(t *testing.T, dbname string, extensions []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db, err := testAccLoadPostgreSQLDatabase(s, dbname)
		if err != nil {
			return err
		}
		actual := []string{}
		for _, ext := range db.Extensions {
			actual = append(actual, ext.Name)
		}
		assert.Equal(t, actual, extensions)
		return nil
	}
}

func testAccMDBPostgreSQLDatabaseConfigStep0(name string) string {
	return fmt.Sprintf(pgVPCDependencies+`
resource "yandex_mdb_postgresql_cluster" "foo" {
	name        = "%s"
	description = "PostgreSQL Database Terraform Test"
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

resource "yandex_mdb_postgresql_user" "alice" {
	cluster_id = yandex_mdb_postgresql_cluster.foo.id
	name       = "alice"
	password   = "mysecurepassword"
}
`, name)
}

// Create database
func testAccMDBPostgreSQLDatabaseConfigStep1(name string) string {
	return testAccMDBPostgreSQLDatabaseConfigStep0(name) + `
resource "yandex_mdb_postgresql_database" "testdb" {
	cluster_id = yandex_mdb_postgresql_cluster.foo.id
	name       = "testdb"
	owner      = yandex_mdb_postgresql_user.alice.name
	lc_collate = "en_US.UTF-8"
	lc_type    = "en_US.UTF-8"
}
`
}

// Extensions change works
func testAccMDBPostgreSQLDatabaseConfigStep2(name string) string {
	return testAccMDBPostgreSQLDatabaseConfigStep0(name) + `
resource "yandex_mdb_postgresql_database" "testdb" {
	cluster_id = yandex_mdb_postgresql_cluster.foo.id
	name       = "testdb"
	owner      = yandex_mdb_postgresql_user.alice.name
	lc_collate = "en_US.UTF-8"
	lc_type    = "en_US.UTF-8"

	extension {
		name    = "uuid-ossp"
	}
	extension {
		name    = "xml2"
	}
}
`
}

// Create database with template_db
func testAccMDBPostgreSQLDatabaseConfigStep3(name string) string {
	return testAccMDBPostgreSQLDatabaseConfigStep0(name) + `
resource "yandex_mdb_postgresql_database" "testdb1" {
	cluster_id  = yandex_mdb_postgresql_cluster.foo.id
	name        = "testdb1"
	template_db = "testdb"
	owner       = yandex_mdb_postgresql_user.alice.name
	lc_collate  = "en_US.UTF-8"
	lc_type     = "en_US.UTF-8"
}
` + `
resource "yandex_mdb_postgresql_database" "testdb" {
	cluster_id = yandex_mdb_postgresql_cluster.foo.id
	name       = "testdb"
	owner      = yandex_mdb_postgresql_user.alice.name
	lc_collate = "en_US.UTF-8"
	lc_type    = "en_US.UTF-8"
}
`
}
