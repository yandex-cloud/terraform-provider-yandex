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

func TestAccDataSourceMDBPostgreSQLUser_basic(t *testing.T) {
	t.Parallel()

	clusterName := acctest.RandomWithPrefix("ds-pg-by-id")
	description := "PostgreSQL User Terraform Datasource Test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMDBPostgreSQLUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMDBPostgreSQLUserConfig(clusterName, description, true),
				Check: testAccDataSourceMDBPGUserCheck(
					"data.yandex_mdb_postgresql_user.alice",
					"yandex_mdb_postgresql_user.alice", clusterName, description),
			},
		},
	})
}

func testAccDataSourceMDBPGUserAttributesCheck(datasourceName string, resourceName string) resource.TestCheckFunc {
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
				"name",
				"name",
			},
			{
				"grants.#",
				"grants.#",
			},
			{
				"login",
				"login",
			},
			{
				"permission.#",
				"permission.#",
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

func testAccDataSourceMDBPGUserCheck(datasourceName string, resourceName string, clusterName string, description string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccDataSourceMDBPGUserAttributesCheck(datasourceName, resourceName),
		testAccDataSourceMDBPGUserCheckResourceIDField(datasourceName),
		resource.TestCheckResourceAttr(datasourceName, "name", "alice"),
		resource.TestCheckResourceAttr(datasourceName, "login", "true"),
	)
}

func testAccDataSourceMDBPGUserCheckResourceIDField(resourceName string) resource.TestCheckFunc {
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

func testAccDataSourceMDBPostgreSQLUserConfig(name string, description string, useDataID bool) string {
	return fmt.Sprintf(pgVPCDependencies+`
resource "yandex_mdb_postgresql_cluster" "foo" {
	name        = "%s"
	description = "%s"
	environment = "PRODUCTION"
	network_id  = "${yandex_vpc_network.mdb-pg-test-net.id}"

	config {
		version = 11
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
	login      = "true"
	grants     = ["mdb_admin", "mdb_replication"]
}

data "yandex_mdb_postgresql_user" "alice" {
	cluster_id = yandex_mdb_postgresql_cluster.foo.id
	name       = yandex_mdb_postgresql_user.alice.name
}
`, name, description)
}

func testAccCheckMDBPostgreSQLUserDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_mdb_postgresql_user" {
			continue
		}

		clusterId, username, err := deconstructResourceId(rs.Primary.ID)
		if err != nil {
			return err
		}
		_, err = config.sdk.MDB().PostgreSQL().User().Get(context.Background(), &postgresql.GetUserRequest{
			ClusterId: clusterId,
			UserName:  username,
		})

		if err == nil {
			return fmt.Errorf("PostgreSQL User still exists")
		}
	}

	return nil
}
