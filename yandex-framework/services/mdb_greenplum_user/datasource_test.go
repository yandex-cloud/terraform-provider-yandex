package mdb_greenplum_user_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/resourceid"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func TestAccDataSourceMDBGreenplumUser_basic(t *testing.T) {
	t.Parallel()

	clusterName := acctest.RandomWithPrefix("ds-greenplum-user")
	description := "Greenplum User Terraform Datasource Test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMDBGreenplumUserConfig(clusterName, description),
				Check: testAccDataSourceMDBMGUserCheck(
					"data.yandex_mdb_greenplum_user.bar", "yandex_mdb_greenplum_user.foo",
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
				"resource_group",
				"resource_group",
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

func testAccDataSourceMDBGreenplumUserConfig(name string, description string) string {
	return fmt.Sprintf(VPCDependencies+`
resource "yandex_mdb_greenplum_cluster" "foo" {
	name        = "%s"
	description = "%s"
	environment = "PRESTABLE"
	network_id  = yandex_vpc_network.foo.id

	zone = "ru-central1-b"
	subnet_id = yandex_vpc_subnet.foo.id
	assign_public_ip = false
	version = "6.25"
	
	labels = { test_key_create : "test_value_create" }
	
	master_host_count  = 2
	
	master_subcluster {
		resources {
			resource_preset_id = "s2.micro"
			disk_size          = 24
			disk_type_id       = "network-ssd"
		}
	}
	segment_subcluster {
		resources {
			resource_preset_id = "s2.small"
			disk_size          = 24
			disk_type_id       = "network-ssd"
		}
	}

	segment_host_count = 2
	segment_in_host = 2
	
	user_name     = "user1"
	user_password = "mysecurepassword"
}

resource "yandex_mdb_greenplum_resource_group" "some_group1" {
	cluster_id     = yandex_mdb_greenplum_cluster.foo.id
	name           = "some_group1"
	cpu_rate_limit = 25
}

resource "yandex_mdb_greenplum_user" "foo" {
	cluster_id     = yandex_mdb_greenplum_cluster.foo.id
	name           = "bob"
	password       = "mysecureP@ssw0rd"
	resource_group = yandex_mdb_greenplum_resource_group.some_group1.name
}

data "yandex_mdb_greenplum_user" "bar" {
	cluster_id = yandex_mdb_greenplum_cluster.foo.id
	name       = yandex_mdb_greenplum_user.foo.name
}
`, name, description)
}
