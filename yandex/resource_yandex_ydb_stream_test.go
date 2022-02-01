package yandex

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_PersQueue_V1"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const (
	ydbDatabaseStreamResource = "yandex_ydb_stream.test-ydb-stream"
)

func init() {
	resource.AddTestSweepers("yandex_ydb_stream", &resource.Sweeper{
		Name: "yandex_ydb_stream",
		F:    testSweepYDBDatabaseServerless, // NOTE(shmel1k@): all streams are stored in ydb databases.
	})
}

func testGetYDBStreamByID(config *Config, databaseEndpoint, streamName string) (*Ydb_PersQueue_V1.TopicSettings, error) {
	client, err := createYDBStreamClient(context.Background(), databaseEndpoint, config)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = client.Close()
	}()

	result, err := client.DescribeTopic(context.Background(), streamName)
	if err != nil {
		return nil, err
	}

	return result.GetSettings(), nil
}

func testYandexYDBStreamExists(resourceName, streamName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found resource %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("ID is not set for resource %+v", rs)
		}

		if !strings.HasSuffix(rs.Primary.ID, streamName) {
			return fmt.Errorf("got primary id %q without streamName %q", rs.Primary.ID, streamName)
		}

		config := testAccProvider.Meta().(*Config)

		ctx := context.Background()

		client, err := createYDBStreamClient(ctx, rs.Primary.Attributes["database_endpoint"], config)
		if err != nil {
			return err
		}
		defer func() {
			_ = client.Close()
		}()

		_, err = client.DescribeTopic(ctx, streamName)
		return err
	}
}

func testYandexYDBStreamBasic(databaseName, streamName string) string {
	return fmt.Sprintf(`
resource "yandex_ydb_database_serverless" "test-ydb-database" {
  name = "%s"
}

resource "yandex_ydb_stream" "test-ydb-stream" {
  stream_name = "%s"
  database_endpoint = "${yandex_ydb_serverless.test-ydb-database.id}"
}
`, databaseName, streamName)
}

type testYandexYDBStreamParams struct {
	streamName        string
	partitionsCount   int
	retentionPeriodMS int
	supportedCodecs   []string
}

func (t *testYandexYDBStreamParams) formatCodecs() string {
	return strings.Join(t.supportedCodecs, "\n")
}

func testYandexYDBStreamFull(ydbDatabaseName string, params testYandexYDBStreamParams) string {
	return fmt.Sprintf(`
resource "yandex_ydb_database_serverless" "test-ydb-database-serverless" {
  name = "%s",
}

resource "yandex_ydb_stream" "test-ydb-stream" {
  database_endpoint = "${yandex_ydb_database_serverless.test-ydb-database-serverless.ydb_full_endpoint}"
  name = "%s"
  partitions_count = %d
  supported_codecs = [
    %s
  ]
  retention_period_ms = %d
}
`, ydbDatabaseName, params.streamName, params.partitionsCount, params.formatCodecs(), params.retentionPeriodMS)
}

func basicYandexYDBStreamTestStep(ydbDatabaseName, streamName string) resource.TestStep {
	return resource.TestStep{
		Config: testYandexYDBStreamBasic(ydbDatabaseName, streamName),
		Check: resource.ComposeTestCheckFunc(
			testYandexYDBStreamExists(ydbDatabaseStreamResource, streamName),
			resource.TestCheckResourceAttr(ydbDatabaseStreamResource, "name", streamName),
			testAccCheckCreatedAtAttr(ydbDatabaseStreamResource),
		),
	}
}

func testYandexYDBStreamDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_ydb_stream" {
			continue
		}

		_, err := testGetYDBStreamByID(
			config,
			rs.Primary.Attributes["database_endpoint"],
			rs.Primary.Attributes["stream_name"],
		)
		if err == nil {
			return fmt.Errorf("YDB Stream still exists")
		}
	}

	return nil
}

func TestAccYandexYDBStream_basic(t *testing.T) {
	streamName := acctest.RandomWithPrefix("tf-ydb-stream-stream")
	databaseName := acctest.RandomWithPrefix("tf-ydb-stream-database")
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexYDBStreamDestroy,
		Steps: []resource.TestStep{
			basicYandexYDBStreamTestStep(databaseName, streamName),
		},
	})
}

func TestAccYandexYDBStream_update(t *testing.T) {
	streamName := acctest.RandomWithPrefix("tf-ydb-stream-stream")
	databaseName := acctest.RandomWithPrefix("tf-ydb-stream-database")

	testConfigFn := func(params testYandexYDBStreamParams) resource.TestStep {
		return resource.TestStep{
			Config: testYandexYDBStreamFull(databaseName, params),
			Check: resource.ComposeTestCheckFunc(
				testYandexYDBStreamExists(ydbDatabaseStreamResource, params.streamName),
				resource.TestCheckResourceAttr(ydbDatabaseStreamResource, "name", params.streamName),
				resource.TestCheckResourceAttr(ydbDatabaseStreamResource, "partitions_count", strconv.Itoa(params.partitionsCount)),
				resource.TestCheckResourceAttrSet(ydbDatabaseStreamResource, "supported_codecs"),
				resource.TestCheckResourceAttr(ydbDatabaseStreamResource, "retention_period_ms", strconv.Itoa(params.retentionPeriodMS)),
			),
		}
	}

	beforeParams := testYandexYDBStreamParams{
		streamName:        streamName,
		partitionsCount:   2,
		retentionPeriodMS: 1000 * 60 * 60, // NOTE(shmel1k@): 1 hour
		supportedCodecs:   ydbStreamAllowedCodecs,
	}
	afterParams := testYandexYDBStreamParams{
		streamName:        streamName,
		partitionsCount:   4,
		retentionPeriodMS: 1000 * 60 * 60 * 4, // NOTE(shmel1k@): 4 hours.
		supportedCodecs:   ydbStreamAllowedCodecs,
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexYDBStreamDestroy,
		Steps: []resource.TestStep{
			testConfigFn(beforeParams),
			testConfigFn(afterParams),
		},
	})
}

func TestMergeYDBStreamConsumerSettings(t *testing.T) {
	var testData = []struct {
		testName                     string
		consumers                    []interface{}
		readRules                    []*Ydb_PersQueue_V1.TopicSettings_ReadRule
		expectedReadRules            []*Ydb_PersQueue_V1.TopicSettings_ReadRule
		expectedConsumersForDeletion []string
	}{
		{
			testName:                     "empty config-consumers and stream consumers",
			consumers:                    []interface{}{},
			readRules:                    nil,
			expectedReadRules:            nil,
			expectedConsumersForDeletion: nil,
		},
		{
			testName:  "empty config-consumers with non-empty stream consumers",
			consumers: []interface{}{},
			readRules: []*Ydb_PersQueue_V1.TopicSettings_ReadRule{
				{
					ConsumerName:    "a",
					SupportedFormat: Ydb_PersQueue_V1.TopicSettings_FORMAT_BASE,
					SupportedCodecs: ydbStreamDefaultCodecs,
				},
			},
			expectedReadRules: []*Ydb_PersQueue_V1.TopicSettings_ReadRule{
				{
					ConsumerName:    "a",
					SupportedFormat: Ydb_PersQueue_V1.TopicSettings_FORMAT_BASE,
					SupportedCodecs: ydbStreamDefaultCodecs,
				},
			},
		},
		{
			testName: "non-empty config consumers with empty config consumers",
			consumers: []interface{}{
				map[string]interface{}{
					"name": "stream_name",
					"supported_codecs": []interface{}{
						"gzip",
					},
				},
			},
			readRules: nil,
			expectedReadRules: []*Ydb_PersQueue_V1.TopicSettings_ReadRule{
				{
					ConsumerName: "stream_name",
					SupportedCodecs: []Ydb_PersQueue_V1.Codec{
						Ydb_PersQueue_V1.Codec_CODEC_GZIP,
					},
					SupportedFormat: Ydb_PersQueue_V1.TopicSettings_FORMAT_BASE,
				},
			},
		},
		{
			testName: "non-empty config consumers with missing fields",
			consumers: []interface{}{
				map[string]interface{}{
					"name": "consumer_name",
					"supported_codecs": []interface{}{
						"gzip",
					},
				},
				map[string]interface{}{
					"name":         "consumer_name_2",
					"service_type": "data-transfer",
				},
				map[string]interface{}{
					"name":                          "consumer_name_3",
					"starting_message_timestamp_ms": 42,
				},
			},
			readRules: nil,
			expectedReadRules: []*Ydb_PersQueue_V1.TopicSettings_ReadRule{
				{
					ConsumerName: "consumer_name",
					SupportedCodecs: []Ydb_PersQueue_V1.Codec{
						Ydb_PersQueue_V1.Codec_CODEC_GZIP,
					},
					SupportedFormat: Ydb_PersQueue_V1.TopicSettings_FORMAT_BASE,
				},
				{
					ConsumerName:    "consumer_name_2",
					SupportedCodecs: ydbStreamDefaultCodecs,
					SupportedFormat: Ydb_PersQueue_V1.TopicSettings_FORMAT_BASE,
					ServiceType:     "data-transfer",
				},
				{
					ConsumerName:               "consumer_name_3",
					SupportedCodecs:            ydbStreamDefaultCodecs,
					SupportedFormat:            Ydb_PersQueue_V1.TopicSettings_FORMAT_BASE,
					StartingMessageTimestampMs: 42,
				},
			},
		},
		{
			testName: "non-empty config consumers without settings",
			consumers: []interface{}{
				map[string]interface{}{
					"name": "consumer_name",
				},
				map[string]interface{}{
					"name": "consumer_name_2",
				},
			},
			readRules: nil,
			expectedReadRules: []*Ydb_PersQueue_V1.TopicSettings_ReadRule{
				{
					ConsumerName:    "consumer_name",
					SupportedCodecs: ydbStreamDefaultCodecs,
					SupportedFormat: Ydb_PersQueue_V1.TopicSettings_FORMAT_BASE,
				},
				{
					ConsumerName:    "consumer_name_2",
					SupportedCodecs: ydbStreamDefaultCodecs,
					SupportedFormat: Ydb_PersQueue_V1.TopicSettings_FORMAT_BASE,
				},
			},
		},
		{
			testName: "consumers with different settings",
			consumers: []interface{}{
				map[string]interface{}{
					"name": "consumer_name",
					"supported_codecs": []interface{}{
						ydbStreamCodecRAW,
					},
					"service_type": "some_service_type",
				},
				map[string]interface{}{
					"name": "consumer_name_2",
					"supported_codecs": []interface{}{
						ydbStreamCodecGZIP,
					},
					"service_type": "another_service_type",
				},
			},
			readRules: []*Ydb_PersQueue_V1.TopicSettings_ReadRule{
				{
					ConsumerName: "consumer_name",
					SupportedCodecs: []Ydb_PersQueue_V1.Codec{
						Ydb_PersQueue_V1.Codec_CODEC_GZIP,
					},
					SupportedFormat: Ydb_PersQueue_V1.TopicSettings_FORMAT_UNSPECIFIED,
					ServiceType:     "some_service_type_1",
				},
				{
					ConsumerName: "consumer_name_2",
					SupportedCodecs: []Ydb_PersQueue_V1.Codec{
						Ydb_PersQueue_V1.Codec_CODEC_RAW,
					},
					SupportedFormat: Ydb_PersQueue_V1.TopicSettings_FORMAT_BASE,
					ServiceType:     "another_service_type_2",
				},
			},
			expectedReadRules: []*Ydb_PersQueue_V1.TopicSettings_ReadRule{
				{
					ConsumerName: "consumer_name",
					SupportedCodecs: []Ydb_PersQueue_V1.Codec{
						Ydb_PersQueue_V1.Codec_CODEC_RAW,
					},
					SupportedFormat: Ydb_PersQueue_V1.TopicSettings_FORMAT_UNSPECIFIED,
					ServiceType:     "some_service_type",
				},
				{
					ConsumerName: "consumer_name_2",
					SupportedCodecs: []Ydb_PersQueue_V1.Codec{
						Ydb_PersQueue_V1.Codec_CODEC_GZIP,
					},
					SupportedFormat: Ydb_PersQueue_V1.TopicSettings_FORMAT_BASE,
					ServiceType:     "another_service_type",
				},
			},
		},
	}

	for _, v := range testData {
		v := v
		t.Run(v.testName, func(t *testing.T) {
			newReadRules := mergeYDBStreamConsumerSettings(v.consumers, v.readRules)
			if !reflect.DeepEqual(newReadRules, v.expectedReadRules) {
				t.Errorf("got readrules %+v\nexpected %+v", newReadRules, v.expectedReadRules)
			}
		})
	}
}
