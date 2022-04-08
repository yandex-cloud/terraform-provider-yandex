package yandex

import (
	"fmt"
	"testing"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mongodb/v1"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceMDBMongoDBCluster_byName(t *testing.T) {
	t.Parallel()

	clusterName := acctest.RandomWithPrefix("test-acc-ds-mongodb-by-name")
	configData := map[string]interface{}{
		"ClusterName": clusterName,
		"Environment": "PRESTABLE",
		"Lables":      map[string]string{"test_key": "test_value"},
		"BackupWindow": map[string]int64{
			"hours":   3,
			"minutes": 4,
		},
		"Version":   "4.2",
		"Databases": []string{"testdb"},
		"Users": []*mongodb.UserSpec{
			{
				Name:     "john",
				Password: "password",
				Permissions: []*mongodb.Permission{
					{
						DatabaseName: "testdb",
					},
				},
			},
		},
		"Resources": mongodb.Resources{
			ResourcePresetId: "s2.micro",
			DiskSize:         16,
			DiskTypeId:       "network-hdd",
		},
		"Hosts": []map[string]interface{}{
			{"ZoneId": "ru-central1-a", "SubnetId": "${yandex_vpc_subnet.foo.id}"},
			{"ZoneId": "ru-central1-b", "SubnetId": "${yandex_vpc_subnet.bar.id}"},
		},
		"SecurityGroupIds": []string{"${yandex_vpc_security_group.sg-x.id}"},
		"MaintenanceWindow": map[string]interface{}{
			"Type": "WEEKLY",
			"Day":  "FRI",
			"Hour": 20,
		},
		"DeletionProtection": false,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMDBMongoDBClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: makeConfig(t, &configData, nil) + mdbMongoDBClusterByNameConfig,
				Check: testAccDataSourceMDBMongoDBClusterCheck(
					"data.yandex_mdb_mongodb_cluster.bar",
					"yandex_mdb_mongodb_cluster.foo", clusterName),
			},
		},
	})
}

func testAccDataSourceMDBMongoDBClusterAttributesCheck(datasourceName string, resourceName string) resource.TestCheckFunc {
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

		instanceAttrsToTest := []string{
			"name",
			"folder_id",
			"network_id",
			"created_at",
			"description",
			"labels",
			"environment",
			"resources",
			"database",
			"user.0.name",
			"user.0.permission",
			"user.0.database",
			"host",
			"sharded",
			"cluster_config.0.version",
			"security_group_ids",
			"maintenance_window.0.type",
			"maintenance_window.0.day",
			"maintenance_window.0.hour",
			"deletion_protection",
		}

		for _, attrToCheck := range instanceAttrsToTest {
			if datasourceAttributes[attrToCheck] != resourceAttributes[attrToCheck] {
				return fmt.Errorf(
					"%s is %s; want %s",
					attrToCheck,
					datasourceAttributes[attrToCheck],
					resourceAttributes[attrToCheck],
				)
			}
		}

		return nil
	}
}

func testAccDataSourceMDBMongoDBClusterCheck(datasourceName string, resourceName string, mongodbName string) resource.TestCheckFunc {
	folderID := getExampleFolderID()
	env := "PRESTABLE"

	return resource.ComposeTestCheckFunc(
		testAccDataSourceMDBMongoDBClusterAttributesCheck(datasourceName, resourceName),
		testAccCheckResourceIDField(datasourceName, "cluster_id"),
		resource.TestCheckResourceAttr(datasourceName, "name", mongodbName),
		resource.TestCheckResourceAttr(datasourceName, "folder_id", folderID),
		resource.TestCheckResourceAttr(datasourceName, "environment", env),
		resource.TestCheckResourceAttr(datasourceName, "labels.test_key", "test_value"),
		resource.TestCheckResourceAttr(datasourceName, "sharded", "false"),
		resource.TestCheckResourceAttr(datasourceName, "host.#", "2"),
		testAccCheckCreatedAtAttr(datasourceName),
		resource.TestCheckResourceAttr(datasourceName, "security_group_ids.#", "1"),
		resource.TestCheckResourceAttr(datasourceName, "maintenance_window.0.type", "WEEKLY"),
		resource.TestCheckResourceAttr(datasourceName, "maintenance_window.0.day", "FRI"),
		resource.TestCheckResourceAttr(datasourceName, "maintenance_window.0.hour", "20"),
		resource.TestCheckResourceAttr(datasourceName, "deletion_protection", "false"),
	)
}

const mdbMongoDBClusterByNameConfig = `
data "yandex_mdb_mongodb_cluster" "bar" {
  name = "${yandex_mdb_mongodb_cluster.foo.name}"
}
`
