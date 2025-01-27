package mdb_mongodb_database_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mongodb/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
)

const (
	mgDatabaseResourceName    = "yandex_mdb_mongodb_database.testdb"
	mgDatabaseResourceName1   = "yandex_mdb_mongodb_database.testdb1"
	testMongoDBDatabasePrefix = "tf-mongodb-database"
	mgClusterResourceName     = "yandex_mdb_mongodb_cluster.foo"
)

const VPCDependencies = `
resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.1.0.0/24"]
}
`

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

// Test that a MongoDB Database can be created, updated and destroyed
func TestAccMDBMongoDBDatabase_full(t *testing.T) {
	t.Parallel()
	clusterName := acctest.RandomWithPrefix(testMongoDBDatabasePrefix)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBMongoDBDatabaseConfigStep1(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(mgDatabaseResourceName, "name", "testdb"),
				),
			},
			mdbMongoDBDatabaseImportStep(mgDatabaseResourceName),
			{
				Config: testAccMDBMongoDBDatabaseConfigStep2(clusterName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMDBMongoDBClusterHasDatabase(t, "renamed_testdb"),
				),
			},
			mdbMongoDBDatabaseImportStep(mgDatabaseResourceName),
			{
				Config: testAccMDBMongoDBDatabaseConfigStep4(clusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(mgDatabaseResourceName1, "name", "testdb1"),
					resource.TestCheckResourceAttr(mgDatabaseResourceName, "name", "testdb"),
				),
			},
			mdbMongoDBDatabaseImportStep(mgDatabaseResourceName1),
		},
	})
}

func mdbMongoDBDatabaseImportStep(name string) resource.TestStep {
	return resource.TestStep{
		ResourceName:      name,
		ImportState:       true,
		ImportStateVerify: true,
	}
}

func testAccLoadMongoDBDatabase(s *terraform.State, dbname string) (*mongodb.Database, error) {
	rs, ok := s.RootModule().Resources[mgClusterResourceName]

	if !ok {
		return nil, fmt.Errorf("resource %q not found", mgDatabaseResourceName)
	}
	if rs.Primary.ID == "" {
		return nil, fmt.Errorf("no ID is set")
	}

	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()
	return config.SDK.MDB().MongoDB().Database().Get(context.Background(), &mongodb.GetDatabaseRequest{
		ClusterId:    rs.Primary.ID,
		DatabaseName: dbname,
	})
}

func testAccCheckMDBMongoDBClusterHasDatabase(t *testing.T, dbname string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db, err := testAccLoadMongoDBDatabase(s, dbname)
		if err != nil {
			return err
		}
		assert.Equal(t, db.Name, dbname)
		return nil
	}
}

func testAccMDBMongoDBDatabaseConfigStep0(name string) string {
	return fmt.Sprintf(VPCDependencies+`
resource "yandex_mdb_mongodb_cluster" "foo" {
	name        = "%s"
	description = "MongoDB User Terraform Test"
	environment = "PRESTABLE"
	network_id  = yandex_vpc_network.foo.id

	cluster_config {
	    version = "6.0"
	}

	host {
		zone_id      = "ru-central1-a"
		subnet_id  = yandex_vpc_subnet.foo.id
	}
	resources_mongod {
		  resource_preset_id = "s2.micro"
		  disk_size          = 10
		  disk_type_id       = "network-ssd"
	    }
}
`, name)
}

// Create database
func testAccMDBMongoDBDatabaseConfigStep1(name string) string {
	return testAccMDBMongoDBDatabaseConfigStep0(name) + `
resource "yandex_mdb_mongodb_database" "testdb" {
	cluster_id = yandex_mdb_mongodb_cluster.foo.id
	name       = "testdb"
}
`
}

// Database rename is not supported yet
func testAccMDBMongoDBDatabaseConfigStep2(name string) string {
	return testAccMDBMongoDBDatabaseConfigStep0(name) + `
resource "yandex_mdb_mongodb_database" "testdb" {
	cluster_id 			= yandex_mdb_mongodb_cluster.foo.id
	name       			= "renamed_testdb"
}
`
}

// Create database with template_db
func testAccMDBMongoDBDatabaseConfigStep4(name string) string {
	return testAccMDBMongoDBDatabaseConfigStep0(name) + `
resource "yandex_mdb_mongodb_database" "testdb1" {
	cluster_id  = yandex_mdb_mongodb_cluster.foo.id
	name        = "testdb1"
}
` + `
resource "yandex_mdb_mongodb_database" "testdb" {
	cluster_id = yandex_mdb_mongodb_cluster.foo.id
	name       = "testdb"
}
`
}
