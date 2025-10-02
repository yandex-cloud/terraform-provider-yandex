package yandex

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
)

const (
	yandexLoggingGroupDataSource = "data.yandex_logging_group.test-logging-group"
	dataStreamName               = "logging-yds"
	ydbResource                  = "logging-ydb"
	topicResource                = "logging-topic"
)

func TestAccDataSourceYandexLoggingGroup_byID(t *testing.T) {
	var group logging.LogGroup
	name := acctest.RandomWithPrefix("tf-yandex-logging-group")
	desc := acctest.RandomWithPrefix("tf-yandex-logging-group-desc")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexLoggingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexLoggingGroupByID(name, desc),
				Check: resource.ComposeTestCheckFunc(
					testYandexLoggingGroupExists(yandexLoggingGroupDataSource, &group),
					resource.TestCheckResourceAttrSet(yandexLoggingGroupDataSource, "group_id"),
					resource.TestCheckResourceAttr(yandexLoggingGroupDataSource, "name", name),
					resource.TestCheckResourceAttr(yandexLoggingGroupDataSource, "description", desc),
					resource.TestCheckResourceAttrSet(yandexLoggingGroupDataSource, "retention_period"),
					resource.TestCheckResourceAttrSet(yandexLoggingGroupDataSource, "folder_id"),
					resource.TestCheckResourceAttrSet(yandexLoggingGroupDataSource, "cloud_id"),
					resource.TestCheckResourceAttrSet(yandexLoggingGroupDataSource, "created_at"),
					testAccCheckCreatedAtAttr(yandexLoggingGroupDataSource),
				),
			},
		},
	})
}

func TestAccDataSourceYandexLoggingGroup_byName(t *testing.T) {
	var group logging.LogGroup
	name := acctest.RandomWithPrefix("tf-yandex-logging-group")
	desc := acctest.RandomWithPrefix("tf-yandex-logging-group-desc")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexLoggingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexLoggingGroupByName(name, desc),
				Check: resource.ComposeTestCheckFunc(
					testYandexLoggingGroupExists(yandexLoggingGroupDataSource, &group),
					resource.TestCheckResourceAttrSet(yandexLoggingGroupDataSource, "group_id"),
					resource.TestCheckResourceAttr(yandexLoggingGroupDataSource, "name", name),
					resource.TestCheckResourceAttr(yandexLoggingGroupDataSource, "description", desc),
					resource.TestCheckResourceAttrSet(yandexLoggingGroupDataSource, "retention_period"),
					resource.TestCheckResourceAttrSet(yandexLoggingGroupDataSource, "folder_id"),
					resource.TestCheckResourceAttrSet(yandexLoggingGroupDataSource, "cloud_id"),
					resource.TestCheckResourceAttrSet(yandexLoggingGroupDataSource, "created_at"),
					testAccCheckCreatedAtAttr(yandexLoggingGroupDataSource),
				),
			},
		},
	})
}

func TestAccDataSourceYandexLoggingGroup_full(t *testing.T) {
	var group logging.LogGroup
	params := testYandexLoggingGroupParameters{
		name:            acctest.RandomWithPrefix("tf-yandex-logging-group"),
		desc:            acctest.RandomWithPrefix("tf-yandex-logging-group-desc"),
		dataStream:      dataStreamName,
		labelKey:        acctest.RandomWithPrefix("tf-yandex-logging-group-label"),
		labelValue:      acctest.RandomWithPrefix("tf-yandex-logging-group-label-value"),
		retentionPeriod: time.Hour + time.Duration(rand.Uint32())*time.Nanosecond,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexLoggingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexLoggingGroupDataSource(params),
				Check: func(s *terraform.State) error {
					databasePath := s.RootModule().Resources["yandex_ydb_database_serverless."+ydbResource].Primary.Attributes["database_path"]
					dataStreamFullName := databasePath + "/" + params.dataStream
					return resource.ComposeTestCheckFunc(
						testYandexLoggingGroupExists(yandexLoggingGroupDataSource, &group),
						resource.TestCheckResourceAttr(yandexLoggingGroupDataSource, "name", params.name),
						resource.TestCheckResourceAttr(yandexLoggingGroupDataSource, "description", params.desc),
						resource.TestCheckResourceAttr(yandexLoggingGroupDataSource, "data_stream", dataStreamFullName),
						resource.TestCheckResourceAttr(yandexLoggingGroupDataSource, "retention_period", params.retentionPeriod.String()),
						resource.TestCheckResourceAttrSet(yandexLoggingGroupDataSource, "folder_id"),
						resource.TestCheckResourceAttrSet(yandexLoggingGroupDataSource, "cloud_id"),
						resource.TestCheckResourceAttrSet(yandexLoggingGroupDataSource, "created_at"),
						testYandexLoggingGroupContainsLabel(&group, params.labelKey, params.labelValue),
						testAccCheckCreatedAtAttr(yandexLoggingGroupDataSource),
					)(s)
				},
			},
		},
	})
}

func testYandexLoggingGroupDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_logging_group" {
			continue
		}

		_, err := testGetYandexLoggingGroupByID(config, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Yandex Cloud Logging group still exists")
		}
	}

	return nil
}

func testYandexLoggingGroupExists(name string, group *logging.LogGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := testGetYandexLoggingGroupByID(config, rs.Primary.ID)
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Yandex Cloud Logging group not found")
		}

		*group = *found
		return nil
	}
}

func testGetYandexLoggingGroupByID(config *Config, ID string) (*logging.LogGroup, error) {
	req := logging.GetLogGroupRequest{
		LogGroupId: ID,
	}

	return config.sdk.Logging().LogGroup().Get(context.Background(), &req)
}

func testYandexLoggingGroupContainsLabel(group *logging.LogGroup, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		v, ok := group.Labels[key]
		if !ok {
			return fmt.Errorf("expected label with key '%s' not found", key)
		}
		if v != value {
			return fmt.Errorf("incorrect label value for key '%s': expected '%s' but found '%s'", key, value, v)
		}
		return nil
	}
}

type testYandexLoggingGroupParameters struct {
	name            string
	desc            string
	dataStream      string
	labelKey        string
	labelValue      string
	retentionPeriod time.Duration
}

func testYandexLoggingGroupByID(name string, desc string) string {
	return fmt.Sprintf(`
data "yandex_logging_group" "test-logging-group" {
  group_id = "${yandex_logging_group.test-logging-group.id}"
}

resource "yandex_logging_group" "test-logging-group" {
  name        = "%s"
  description = "%s"
}`, name, desc)
}

func testYandexLoggingGroupByName(name string, desc string) string {
	return fmt.Sprintf(`
data "yandex_logging_group" "test-logging-group" {
  name = "${yandex_logging_group.test-logging-group.name}"
}

resource "yandex_logging_group" "test-logging-group" {
  name        = "%s"
  description = "%s"
}
`, name, desc)
}

func testYandexLoggingGroupDataSource(params testYandexLoggingGroupParameters) string {
	return fmt.Sprintf(`
data "yandex_logging_group" "test-logging-group" {
  group_id = "${yandex_logging_group.test-logging-group.id}"
}

resource "yandex_ydb_database_serverless" "%s" {
	name = "%s"
	location_id = "ru-central1"
}

resource "yandex_ydb_topic" "%s" {
	name = "%s"
	database_endpoint = "${yandex_ydb_database_serverless.%s.ydb_full_endpoint}"
}

resource "yandex_logging_group" "test-logging-group" {
  name        = "%s"
  description = "%s"
  labels = {
    %s          = "%s"
    empty-label = ""
  }
  retention_period = "%s"
  data_stream = "${yandex_ydb_database_serverless.%s.database_path}/%s"
}
`,
		ydbResource,
		ydbResource,
		topicResource,
		dataStreamName,
		ydbResource,
		params.name,
		params.desc,
		params.labelKey,
		params.labelValue,
		params.retentionPeriod,
		ydbResource,
		dataStreamName,
	)
}
