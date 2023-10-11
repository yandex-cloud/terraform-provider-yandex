package yandex

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
)

const yandexLoggingGroupResource = "yandex_logging_group.test-logging-group"

func init() {
	resource.AddTestSweepers("yandex_logging_group", &resource.Sweeper{
		Name: "yandex_logging_group",
		F:    testSweepYandexLoggingGroup,
	})
}

func testSweepYandexLoggingGroup(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := conf.sdk.Logging().LogGroup().List(conf.Context(), &logging.ListLogGroupsRequest{
		FolderId: conf.FolderID,
		PageSize: 1000,
	})
	if err != nil {
		return fmt.Errorf("error getting log group: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.Groups {
		if !sweepYandexLoggingGroup(conf, c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Yandex Cloud Logging group %q", c.Id))
		}
	}

	return result.ErrorOrNil()
}

func sweepYandexLoggingGroup(conf *Config, id string) bool {
	return sweepWithRetry(sweepYandexLoggingGroupOnce, conf, "Yandex Cloud Logging group", id)
}

func sweepYandexLoggingGroupOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexLoggingGroupDefaultTimeout)
	defer cancel()

	op, err := conf.sdk.Logging().LogGroup().Delete(ctx, &logging.DeleteLogGroupRequest{
		LogGroupId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}

func TestAccYandexLoggingGroup_basic(t *testing.T) {
	var group logging.LogGroup
	name := acctest.RandomWithPrefix("tf-yandex-logging-group")
	desc := acctest.RandomWithPrefix("tf-yandex-logging-group-desc")
	labelKey := acctest.RandomWithPrefix("tf-yandex-logging-group-label")
	labelValue := acctest.RandomWithPrefix("tf-yandex-logging-group-label-value")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexLoggingGroupDestroy,
		Steps: []resource.TestStep{
			basicYandexLoggingGroupTestStep(name, desc, labelKey, labelValue, &group),
		},
	})
}

func TestAccYandexLoggingGroup_update(t *testing.T) {
	var group logging.LogGroup
	name := acctest.RandomWithPrefix("tf-yandex-logging-group")
	desc := acctest.RandomWithPrefix("tf-yandex-logging-group-desc")
	labelKey := acctest.RandomWithPrefix("tf-yandex-logging-group-label")
	labelValue := acctest.RandomWithPrefix("tf-yandex-logging-group-label-value")

	nameUpdated := acctest.RandomWithPrefix("tf-yandex-logging-group-updated")
	descUpdated := acctest.RandomWithPrefix("tf-yandex-logging-group-desc-updated")
	labelKeyUpdated := acctest.RandomWithPrefix("tf-yandex-logging-group-label-updated")
	labelValueUpdated := acctest.RandomWithPrefix("tf-yandex-logging-group-label-value-updated")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexLoggingGroupDestroy,
		Steps: []resource.TestStep{
			basicYandexLoggingGroupTestStep(name, desc, labelKey, labelValue, &group),
			basicYandexLoggingGroupTestStep(nameUpdated, descUpdated, labelKeyUpdated, labelValueUpdated, &group),
		},
	})
}

func TestAccYandexLoggingGroup_full(t *testing.T) {
	var group logging.LogGroup
	base := testYandexLoggingGroupParameters{
		name:            acctest.RandomWithPrefix("tf-yandex-logging-group"),
		desc:            acctest.RandomWithPrefix("tf-yandex-logging-group-desc"),
		dataStream:      dataStreamName,
		labelKey:        acctest.RandomWithPrefix("tf-yandex-logging-group-label"),
		labelValue:      acctest.RandomWithPrefix("tf-yandex-logging-group-label-value"),
		retentionPeriod: time.Hour + time.Duration(rand.Uint32())*time.Nanosecond,
	}

	updated := testYandexLoggingGroupParameters{
		name:            acctest.RandomWithPrefix("tf-yandex-logging-group-updated"),
		desc:            acctest.RandomWithPrefix("tf-yandex-logging-group-desc-updated"),
		dataStream:      dataStreamName,
		labelKey:        acctest.RandomWithPrefix("tf-yandex-logging-group-label-updated"),
		labelValue:      acctest.RandomWithPrefix("tf-yandex-logging-group-label-value-updated"),
		retentionPeriod: time.Hour + time.Duration(rand.Uint32())*time.Nanosecond,
	}

	testConfigFunc := func(params testYandexLoggingGroupParameters) resource.TestStep {
		return resource.TestStep{
			Config: testYandexLoggingGroupFull(params),
			Check: func(s *terraform.State) error {
				databasePath := s.RootModule().Resources["yandex_ydb_database_serverless."+ydbResource].Primary.Attributes["database_path"]
				dataStreamFullName := databasePath + "/" + params.dataStream
				return resource.ComposeTestCheckFunc(
					testYandexLoggingGroupExists(yandexLoggingGroupResource, &group),
					resource.TestCheckResourceAttr(yandexLoggingGroupResource, "name", params.name),
					resource.TestCheckResourceAttr(yandexLoggingGroupResource, "description", params.desc),
					resource.TestCheckResourceAttr(yandexLoggingGroupResource, "data_stream", dataStreamFullName),
					resource.TestCheckResourceAttr(yandexLoggingGroupResource, "retention_period", params.retentionPeriod.String()),
					resource.TestCheckResourceAttrSet(yandexLoggingGroupResource, "folder_id"),
					resource.TestCheckResourceAttrSet(yandexLoggingGroupResource, "cloud_id"),
					resource.TestCheckResourceAttrSet(yandexLoggingGroupResource, "created_at"),
					testYandexLoggingGroupContainsLabel(&group, params.labelKey, params.labelValue),
					testAccCheckCreatedAtAttr(yandexLoggingGroupResource),
				)(s)
			},
		}
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexLoggingGroupDestroy,
		Steps: []resource.TestStep{
			testConfigFunc(base),
			testConfigFunc(updated),
		},
	})
}

func basicYandexLoggingGroupTestStep(name, desc, labelKey, labelValue string, group *logging.LogGroup) resource.TestStep {
	return resource.TestStep{
		Config: testYandexLoggingGroupBasic(name, desc, labelKey, labelValue),
		Check: resource.ComposeTestCheckFunc(
			testYandexLoggingGroupExists(yandexLoggingGroupResource, group),
			resource.TestCheckResourceAttr(yandexLoggingGroupResource, "name", name),
			resource.TestCheckResourceAttr(yandexLoggingGroupResource, "description", desc),
			resource.TestCheckResourceAttrSet(yandexLoggingGroupResource, "retention_period"),
			resource.TestCheckResourceAttrSet(yandexLoggingGroupResource, "folder_id"),
			resource.TestCheckResourceAttrSet(yandexLoggingGroupResource, "cloud_id"),
			resource.TestCheckResourceAttrSet(yandexLoggingGroupResource, "created_at"),
			testYandexLoggingGroupContainsLabel(group, labelKey, labelValue),
			testAccCheckCreatedAtAttr(yandexLoggingGroupResource),
		),
	}
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

func testYandexLoggingGroupBasic(name string, desc string, labelKey string, labelValue string) string {
	return fmt.Sprintf(`
resource "yandex_logging_group" "test-logging-group" {
  name        = "%s"
  description = "%s"
  labels = {
    %s          = "%s"
    empty-label = ""
  }
}
`, name, desc, labelKey, labelValue)
}

type testYandexLoggingGroupParameters struct {
	name            string
	desc            string
	dataStream      string
	labelKey        string
	labelValue      string
	retentionPeriod time.Duration
}

func testYandexLoggingGroupFull(params testYandexLoggingGroupParameters) string {
	return fmt.Sprintf(`

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
  data_stream = "${yandex_ydb_database_serverless.%s.database_path}/%s"
  labels = {
    %s          = "%s"
    empty-label = ""
  }
  retention_period = "%s"
}
`,
		ydbResource,
		ydbResource,
		topicResource,
		dataStreamName,
		ydbResource,
		params.name,
		params.desc,
		ydbResource,
		params.dataStream,
		params.labelKey,
		params.labelValue,
		params.retentionPeriod.String(),
	)
}
