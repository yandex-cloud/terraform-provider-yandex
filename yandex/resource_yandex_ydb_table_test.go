package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccYandexYDBTable_basic(t *testing.T) {
	ydbResourceName := fmt.Sprintf("ydb-table-test-%s", acctest.RandString(5))
	tableName := fmt.Sprintf("test-%s", acctest.RandString(5))
	tableResourceName := fmt.Sprintf("ydb-test-table-%s", acctest.RandString(5))
	changefeedName := fmt.Sprintf("test-changefeed-%s", acctest.RandString(5))
	changefeedResourceName := fmt.Sprintf("ydb-test-table-changefeed-%s", acctest.RandString(5))
	indexName := fmt.Sprintf("test-index-%s", acctest.RandString(5))
	indexResourceName := fmt.Sprintf("ydb-test-table-index-%s", acctest.RandString(5))
	ydbLocationId := ydbLocationId

	existingYDBResourceName := fmt.Sprintf("yandex_ydb_database_serverless.%s", ydbResourceName)
	existingTableResourceName := fmt.Sprintf("yandex_ydb_table.%s", tableResourceName)
	existingChangefeedResourceName := fmt.Sprintf("yandex_ydb_table_changefeed.%s", changefeedResourceName)
	existingIndexResourceName := fmt.Sprintf("yandex_ydb_table_index.%s", indexResourceName)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexYDBDatabaseServerlessDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccYDBTableConfig(
					"",
					ydbResourceName,
					tableResourceName,
					tableName,
					indexResourceName,
					indexName,
					changefeedResourceName,
					changefeedName,
					ydbLocationId,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccYDBTableExist(tableName, existingYDBResourceName, existingTableResourceName),
					testAccYDBChangefeedExist(tableName+"/"+changefeedName, existingYDBResourceName, existingChangefeedResourceName),
					testAccYDBIndexExist(indexName, tableName, existingYDBResourceName, existingIndexResourceName),
				),
			},
			{
				ResourceName:      existingTableResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccYDBTableConfig(
	subnetsConfig,
	ydbResourceName,
	tableResourceName,
	tablePath,
	indexResourceName,
	indexName,
	changefeedResourceName,
	changefeedName,
	ydbLocationId string,
) string {
	return fmt.Sprintf(`
	%s

	resource "yandex_ydb_database_serverless" "%s" {
		name = "%s"
		location_id = "%s"
		sleep_after = 180
	}
	
	resource "yandex_ydb_table" "%s" {
		path = "%s"
		connection_string = "${yandex_ydb_database_serverless.%s.ydb_full_endpoint}"
        column {
          name = "a" 
          type = "Uint64"
          not_null = true 
        }
        column {
          name     = "b"
          type     = "Uint8"
          not_null = true
        }
        column {
          name = "c"
          type = "Utf8"
        }
        column {
          name = "f"
          type = "Utf8"
        }
        column {
          name = "e"
          type = "String"
        }
        column {
          name = "d"
          type = "Timestamp"
        }

        primary_key = [
          "a", "b"
        ]

        ttl {                     
          column_name     = "d"   
          expire_interval = "PT5S" 
        }

        partitioning_settings {
          auto_partitioning_by_load = false
          auto_partitioning_partition_size_mb    = 256
          auto_partitioning_min_partitions_count = 6
          auto_partitioning_max_partitions_count = 8
        }
        
        read_replicas_settings = "PER_AZ:1"
        
        key_bloom_filter = true

        lifecycle {
          ignore_changes = [
          ]
        }
	}

    resource "yandex_ydb_table_index" "%s" {
        name     = "%s"
        table_id = "${yandex_ydb_table.%s.id}"
        columns  = ["a", "c"]
        type     = "global_async"
        cover    = ["d", "e"]
    }

    resource "yandex_ydb_table_changefeed" "%s" {
        name = "%s"
        table_id = "${yandex_ydb_table.%s.id}"
        mode              = "NEW_IMAGE"
        retention_period  = "PT1H"
        format            = "JSON"
        consumer {
          name = "abacaba"
        }
        consumer {
          name                          = "abacabadabacaba"
          starting_message_timestamp_ms = 1675865515000
        }
      }
	`,
		subnetsConfig,
		ydbResourceName,
		ydbResourceName,
		ydbLocationId,
		tableResourceName,
		tablePath,
		ydbResourceName,
		indexResourceName,
		indexName,
		tableResourceName,
		changefeedResourceName,
		changefeedName,
		tableResourceName,
	)
}

func testAccYDBChangefeedExist(changefeedPath, ydbResourceName, changefeedResourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// TODO(shmel1k@): remove copypaste there and in ydb_permissions_test
		prs, ok := s.RootModule().Resources[changefeedResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", changefeedResourceName)
		}
		if prs.Primary.ID == "" {
			return fmt.Errorf("%s", "no ID for topic is set")
		}

		rs, ok := s.RootModule().Resources[ydbResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", ydbResourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		_, _, _, err := parseYandexYDBDatabaseEndpoint(rs.Primary.Attributes["ydb_full_endpoint"])
		return err
	}
}

func testAccYDBTableExist(tablePath, ydbResourceName, tableResourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// TODO(shmel1k@): remove copypaste there and in ydb_permissions_test
		prs, ok := s.RootModule().Resources[tableResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", tableResourceName)
		}
		if prs.Primary.ID == "" {
			return fmt.Errorf("%s", "no ID for table is set")
		}

		rs, ok := s.RootModule().Resources[ydbResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", ydbResourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		_, _, _, err := parseYandexYDBDatabaseEndpoint(rs.Primary.Attributes["ydb_full_endpoint"])
		return err
	}
}

func testAccYDBIndexExist(indexName, tablePath, ydbResourceName, indexResourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		prs, ok := s.RootModule().Resources[indexResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", indexResourceName)
		}
		if prs.Primary.ID == "" {
			return fmt.Errorf("%s", "no ID for permission is set")
		}

		rs, ok := s.RootModule().Resources[ydbResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", ydbResourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		_, _, _, err := parseYandexYDBDatabaseEndpoint(rs.Primary.Attributes["ydb_full_endpoint"])
		return err
	}
}
