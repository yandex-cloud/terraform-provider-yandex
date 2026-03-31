package mdb_mysql_database_v2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/resourceid"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
)

const (
	mysqlDatabaseResourceName  = "yandex_mdb_mysql_database_v2.testdb"
	mysqlDatabaseResourceName1 = "yandex_mdb_mysql_database_v2.testdb1"
	testMySQLDatabasePrefix    = "tf-mysql-database"
	mysqlClusterResourceName   = "yandex_mdb_mysql_cluster_v2.foo"
)

const VPCDependencies = `
resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.1.0.0/24"]
}

resource "yandex_vpc_subnet" "bar" {
  zone           = "ru-central1-b"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.2.0.0/24"]
}
`

const mysqlVPCDependencies = `
resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.1.0.0/24"]
}
`

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func mdbMySQLDatabaseImportStep(name string) resource.TestStep {
	return resource.TestStep{
		ResourceName:      name,
		ImportState:       true,
		ImportStateVerify: true,
	}
}

func testAccCheckMDBMySQLDatabaseExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

		clusterID, dbName, err := resourceid.Deconstruct(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = config.SDK.MDB().MySQL().Database().Get(
			context.Background(),
			&mysql.GetDatabaseRequest{
				ClusterId:    clusterID,
				DatabaseName: dbName,
			},
		)
		if err != nil {
			return fmt.Errorf("MySQL Database not found: %v", err)
		}

		return nil
	}
}

func testAccCheckMDBMySQLDatabaseDestroy(s *terraform.State) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_mdb_mysql_database_v2" {
			continue
		}

		clusterID, dbName, err := resourceid.Deconstruct(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("failed to deconstruct resource ID %s: %w", rs.Primary.ID, err)
		}

		_, err = config.SDK.MDB().MySQL().Database().Get(
			context.Background(),
			&mysql.GetDatabaseRequest{
				ClusterId:    clusterID,
				DatabaseName: dbName,
			},
		)
		if err == nil {
			return fmt.Errorf("MySQL database %q in cluster %q still exists", dbName, clusterID)
		}
	}

	return nil
}

func testAccCheckMDBMySQLDatabaseResourceIDField(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set for %s", resourceName)
		}

		expectedID := resourceid.Construct(
			rs.Primary.Attributes["cluster_id"],
			rs.Primary.Attributes["name"],
		)

		if expectedID != rs.Primary.ID {
			return fmt.Errorf(
				"wrong resource %s id: expected %s, got %s",
				resourceName, expectedID, rs.Primary.ID,
			)
		}

		return nil
	}
}

func testAccLoadMySQLDatabase(s *terraform.State, dbname string) (*mysql.Database, error) {
	rs, ok := s.RootModule().Resources[mysqlClusterResourceName]
	if !ok {
		return nil, fmt.Errorf("resource %q not found", mysqlClusterResourceName)
	}
	if rs.Primary.ID == "" {
		return nil, fmt.Errorf("no ID is set")
	}

	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()
	return config.SDK.MDB().MySQL().Database().Get(
		context.Background(),
		&mysql.GetDatabaseRequest{
			ClusterId:    rs.Primary.ID,
			DatabaseName: dbname,
		},
	)
}

func testAccCheckMDBMySQLClusterHasDatabase(
	t *testing.T,
	dbname string,
	deletionProtectionMode string,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		db, err := testAccLoadMySQLDatabase(s, dbname)
		if err != nil {
			return err
		}

		if db.Name != dbname {
			return fmt.Errorf("expected db name %q, got %q", dbname, db.Name)
		}

		if db.DeletionProtectionMode.String() != deletionProtectionMode {
			return fmt.Errorf(
				"expected deletion_protection_mode %q, got %q",
				deletionProtectionMode,
				db.DeletionProtectionMode.String(),
			)
		}

		return nil
	}
}
