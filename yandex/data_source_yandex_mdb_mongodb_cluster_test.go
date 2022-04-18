package yandex

import (
	"fmt"
	"testing"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mongodb/v1"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
)

const mdbMongoDBClusterByNameConfig = `
data "yandex_mdb_mongodb_cluster" "bar" {
  name = "${yandex_mdb_mongodb_cluster.foo.name}"
}
`

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

	datasourceName := "data.yandex_mdb_mongodb_cluster.bar"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckMDBMongoDBClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: makeConfig(t, &configData, nil) + mdbMongoDBClusterByNameConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceMDBMongoDBClusterAttributesCheck(datasourceName, "yandex_mdb_mongodb_cluster.foo"),
					testAccCheckResourceIDField(datasourceName, "cluster_id"),
					resource.TestCheckResourceAttr(datasourceName, "name", clusterName),
					resource.TestCheckResourceAttr(datasourceName, "folder_id", getExampleFolderID()),
					resource.TestCheckResourceAttr(datasourceName, "environment", "PRESTABLE"),
					resource.TestCheckResourceAttr(datasourceName, "labels.test_key", "test_value"),
					resource.TestCheckResourceAttr(datasourceName, "sharded", "false"),
					resource.TestCheckResourceAttr(datasourceName, "host.#", "2"),
					testAccCheckCreatedAtAttr(datasourceName),
					resource.TestCheckResourceAttr(datasourceName, "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(datasourceName, "maintenance_window.0.type", "WEEKLY"),
					resource.TestCheckResourceAttr(datasourceName, "maintenance_window.0.day", "FRI"),
					resource.TestCheckResourceAttr(datasourceName, "maintenance_window.0.hour", "20"),
					resource.TestCheckResourceAttr(datasourceName, "deletion_protection", "false"),
				),
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

func TestDataSourceMDBMongoDBClusterSchema(t *testing.T) {
	resourceSchema := resourceYandexMDBMongodbCluster().Schema
	dsSchema := dataSourceYandexMDBMongodbCluster().Schema

	checkRequiredDiff(t, resourceSchema, dsSchema, map[string]interface{}{
		"name":           nil,
		"network_id":     nil,
		"environment":    nil,
		"user":           nil,
		"database":       nil,
		"host":           nil,
		"resources":      nil,
		"cluster_config": nil,
	})

	// check nested list items, for example "host"
	rHost := resourceSchema["host"].Elem.(*schema.Resource)
	dsHost := dsSchema["host"].Elem.(*schema.Resource)
	checkRequiredDiff(t, rHost.Schema, dsHost.Schema, map[string]interface{}{
		"zone_id":   nil,
		"subnet_id": nil,
	})

	// check nested set items, for example "user"
	rUser := resourceSchema["user"].Elem.(*schema.Resource)
	dsUser := dsSchema["user"].Elem.(*schema.Resource)
	checkRequiredDiff(t, rUser.Schema, dsUser.Schema, map[string]interface{}{
		"name":     nil,
		"password": nil,
	})
}

func checkRequiredDiff(t *testing.T, rSchema map[string]*schema.Schema, dsSchema map[string]*schema.Schema,
	requiredOptions map[string]interface{}) {

	for requiredOption := range requiredOptions {
		value, requiredOptionExists := rSchema[requiredOption]
		assert.True(t, requiredOptionExists, "Key %v should be in resource schema", requiredOption)
		assert.True(t, value.Required, "Key %v in resource should by required", requiredOption)
	}

	for key, value := range rSchema {
		_, expectedRequired := requiredOptions[key]
		assert.Equal(t, expectedRequired, value.Required, "Key %v in resource should be required", key)

		dsValue := dsSchema[key]
		assert.False(t, dsValue.Required, "Key %v in ds should be non required", key)
		assert.True(t, dsValue.Optional, "Key %v in ds should be optional", key)
	}
}
