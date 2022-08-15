package yandex

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
)

func TestAccDataSourceMDBPostgreSQLDatabase_basic(t *testing.T) {
	t.Parallel()

	clusterName := acctest.RandomWithPrefix("ds-pg-by-id")
	description := "PostgreSQL Database Terraform Datasource Test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMDBPostgreSQLDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMDBPostgreSQLDatabaseConfig(clusterName, description, true),
				Check: testAccDataSourceMDBPGDatabaseCheck(
					"data.yandex_mdb_postgresql_database.bar",
					"yandex_mdb_postgresql_database.foo", clusterName, description),
			},
		},
	})
}

func testAccDataSourceMDBPGDatabaseAttributesCheck(datasourceName string, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ds, ok := s.RootModule().Resources[datasourceName]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", datasourceName)
		}

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("can't find %s in state", resourceName)
		}

		if ds.Primary.ID != rs.Primary.ID {
			return fmt.Errorf("instance `data source` ID does not match `resource` ID: %s and %s", ds.Primary.ID, rs.Primary.ID)
		}

		datasourceAttributes := ds.Primary.Attributes
		resourceAttributes := rs.Primary.Attributes

		instanceAttrsToTest := []struct {
			dataSourcePath string
			resourcePath   string
		}{
			{
				"cluster_id",
				"cluster_id",
			},
			{
				"name",
				"name",
			},
			{
				"owner",
				"owner",
			},
			{
				"extension.#",
				"extension.#",
			},
			{
				"lc_collate",
				"lc_collate",
			},
			{
				"lc_type",
				"lc_type",
			},
			{
				"template_db",
				"template_db",
			},
		}

		for _, attrToCheck := range instanceAttrsToTest {
			if _, ok := datasourceAttributes[attrToCheck.dataSourcePath]; !ok {
				return fmt.Errorf("%s is not present in data source attributes", attrToCheck.dataSourcePath)
			}
			if _, ok := resourceAttributes[attrToCheck.resourcePath]; !ok {
				return fmt.Errorf("%s is not present in resource attributes", attrToCheck.resourcePath)
			}
			if datasourceAttributes[attrToCheck.dataSourcePath] != resourceAttributes[attrToCheck.resourcePath] {
				return fmt.Errorf(
					"%s is %s; want %s",
					attrToCheck.dataSourcePath,
					datasourceAttributes[attrToCheck.dataSourcePath],
					resourceAttributes[attrToCheck.resourcePath],
				)
			}
		}

		return nil
	}
}

func testAccDataSourceMDBPGDatabaseCheck(datasourceName string, resourceName string, clusterName string, description string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccDataSourceMDBPGDatabaseAttributesCheck(datasourceName, resourceName),
		testAccDataSourceMDBPGDatabaseCheckResourceIDField(datasourceName),
		resource.TestCheckResourceAttr(datasourceName, "name", "foo"),
		resource.TestCheckResourceAttr(datasourceName, "owner", "alice"),
		resource.TestCheckResourceAttr(datasourceName, "lc_collate", "en_US.UTF-8"),
		resource.TestCheckResourceAttr(datasourceName, "lc_type", "en_US.UTF-8"),
	)
}

func testAccDataSourceMDBPGDatabaseCheckResourceIDField(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		expectedResourceId := constructResourceId(rs.Primary.Attributes["cluster_id"], rs.Primary.Attributes["name"])

		if expectedResourceId != rs.Primary.ID {
			return fmt.Errorf("Wrong resource %s id. Expected %s, got %s", resourceName, expectedResourceId, rs.Primary.ID)
		}

		return nil
	}
}

func testAccDataSourceMDBPostgreSQLDatabaseConfig(name string, description string, useDataID bool) string {
	return fmt.Sprintf(pgVPCDependencies+`
resource "yandex_mdb_postgresql_cluster" "foo" {
	name        = "%s"
	description = "%s"
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

resource "yandex_mdb_postgresql_database" "foo" {
	cluster_id = yandex_mdb_postgresql_cluster.foo.id
	name       = "foo"
	owner      = yandex_mdb_postgresql_user.alice.name
	lc_collate = "en_US.UTF-8"
	lc_type    = "en_US.UTF-8"
}

data "yandex_mdb_postgresql_database" "bar" {
	cluster_id = yandex_mdb_postgresql_cluster.foo.id
	name       = yandex_mdb_postgresql_database.foo.name
}
`, name, description)
}

func testAccCheckMDBPostgreSQLDatabaseDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_mdb_postgresql_database" {
			continue
		}

		clusterId, dbname, err := deconstructResourceId(rs.Primary.ID)
		if err != nil {
			return err
		}
		_, err = config.sdk.MDB().PostgreSQL().Database().Get(context.Background(), &postgresql.GetDatabaseRequest{
			ClusterId:    clusterId,
			DatabaseName: dbname,
		})

		if err == nil {
			return fmt.Errorf("PostgreSQL Database still exists")
		}
	}

	return nil
}
