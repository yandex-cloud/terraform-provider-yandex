package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccYandexYDBTopic_basic(t *testing.T) {
	ydbResourceName := fmt.Sprintf("ydb-topic-test-%s", acctest.RandString(5))
	topicName := fmt.Sprintf("test-%s", acctest.RandString(5))
	topicResourceName := fmt.Sprintf("ydb-test-topic-%s", acctest.RandString(5))
	ydbLocationId := ydbLocationId

	existingYDBResourceName := fmt.Sprintf("yandex_ydb_database_serverless.%s", ydbResourceName)
	existingTopicResourceName := fmt.Sprintf("yandex_ydb_topic.%s", topicResourceName)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexYDBDatabaseServerlessDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccYDBTopicConfig(
					"",
					ydbResourceName,
					topicResourceName,
					topicName,
					ydbLocationId,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccYDBTopicExist(topicName, existingYDBResourceName, existingTopicResourceName),
				),
			},
		},
	})
}

func testAccYDBTopicConfig(
	subnetsConfig,
	ydbResourceName,
	topicResourceName,
	topicPath,
	ydbLocationId string,
) string {
	return fmt.Sprintf(`
	%s

	resource "yandex_ydb_database_serverless" "%s" {
		name = "%s"
		location_id = "%s"
		sleep_after = 180
	}

	resource "yandex_ydb_topic" "%s" {
		name = "%s"
		database_endpoint = "${yandex_ydb_database_serverless.%s.ydb_full_endpoint}"
		supported_codecs = ["gzip"]
		consumer {
			name = "consumer"
			supported_codecs = ["gzip"]
		}
		retention_period_hours = 12
		partition_write_speed_kbps = 128
		metering_mode = "reserved_capacity"
		partitions_count = 4
	}
	`,
		subnetsConfig,
		ydbResourceName,
		ydbResourceName,
		ydbLocationId,
		topicResourceName,
		topicPath,
		ydbResourceName,
	)
}

func testAccYDBTopicExist(topicPath, ydbResourceName, topicResourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// TODO(shmel1k@): remove copypaste there and in ydb_permissions_test
		prs, ok := s.RootModule().Resources[topicResourceName]
		if !ok {
			return fmt.Errorf("not found: %s", topicResourceName)
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
