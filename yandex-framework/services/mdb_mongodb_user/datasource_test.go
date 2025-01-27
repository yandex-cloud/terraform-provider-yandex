package mdb_mongodb_user_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mongodb/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/resourceid"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	//mongodbtpl "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/tests/mdb/mongodb"
)

func TestAccDataSourceMDBMongoDBUser_basic(t *testing.T) {
	t.Parallel()

	clusterName := acctest.RandomWithPrefix("ds-mongodb-user")
	description := "MongoDB User Terraform Datasource Test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckMDBMongoDBUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMDBMongoDBUserConfig(clusterName, description),
				Check: testAccDataSourceMDBMGUserCheck(
					"data.yandex_mdb_mongodb_user.bar", "yandex_mdb_mongodb_user.foo",
				),
			},
		},
	})
}

func testAccDataSourceMDBMGUserAttributesCheck(datasourceName string, resourceName string) resource.TestCheckFunc {
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
				"permission.#",
				"permission.#",
			},
			{
				"permission.0.roles.#",
				"permission.0.roles.#",
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

func testAccDataSourceMDBMGUserCheck(datasourceName string, resourceName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccDataSourceMDBMGUserAttributesCheck(datasourceName, resourceName),
		testAccDataSourceMDBMgUserCheckResourceIDField(resourceName),
		resource.TestCheckResourceAttr(datasourceName, "name", "bob"),
	)
}

func testAccDataSourceMDBMgUserCheckResourceIDField(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		expectedResourceId := resourceid.Construct(rs.Primary.Attributes["cluster_id"], rs.Primary.Attributes["name"])

		if expectedResourceId != rs.Primary.ID {
			return fmt.Errorf("Wrong resource %s id. Expected %s, got %s", resourceName, expectedResourceId, rs.Primary.ID)
		}

		return nil
	}
}

func testAccDataSourceMDBMongoDBUserConfig(name string, description string) string {
	return fmt.Sprintf(VPCDependencies+`
resource "yandex_mdb_mongodb_cluster" "foo" {
	name        = "%s"
	description = "%s"
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

resource "yandex_mdb_mongodb_database" "foo" {
	cluster_id = yandex_mdb_mongodb_cluster.foo.id
	name       = "foo"
}

resource "yandex_mdb_mongodb_user" "foo" {
	cluster_id = yandex_mdb_mongodb_cluster.foo.id
	name       = "bob"
	password   = "mysecureP@ssw0rd"
	permission {
    	database_name=yandex_mdb_mongodb_database.foo.name
    	roles = ["readWrite", "read"]
  	}
}

data "yandex_mdb_mongodb_user" "bar" {
	cluster_id = yandex_mdb_mongodb_cluster.foo.id
	name       = yandex_mdb_mongodb_user.foo.name
}
`, name, description)
}

func testAccCheckMDBMongoDBUserDestroy(s *terraform.State) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_mdb_mongodb_user" {
			continue
		}

		clusterId, userName, err := resourceid.Deconstruct(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = config.SDK.MDB().MongoDB().User().Get(context.Background(), &mongodb.GetUserRequest{
			ClusterId: clusterId,
			UserName:  userName,
		})

		if err == nil {
			return fmt.Errorf("MongoDB Database still exists")
		}
	}

	return nil
}
