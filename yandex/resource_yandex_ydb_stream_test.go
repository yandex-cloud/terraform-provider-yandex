package yandex

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_PersQueue_V1"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/ydb/v1"
)

const (
	ydbDatabaseStreamResource = "yandex_ydb_stream.test-ydb-stream"
)

func init() {
	resource.AddTestSweepers("yandex_ydb_stream", &resource.Sweeper{
		Name: "yandex_ydb_stream",
		F:    testSweepYandexYDBStream,
	})
}

func testSweepYandexYDBStream(_ string) error {
	// NOTE(shmel1k@): destroy databases for stream.
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	_ = conf

	return nil
}

func testGetYDBStreamByID(config *Config, ydbEndpoint, streamID string) error {
	return nil
}

func testYandexYDBStreamExists(streamName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[streamName]
		if !ok {
			return fmt.Errorf("not found resource %s", streamName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("ID is not set for resource %+v", rs)
		}

		// config := testAccProvider.Meta().(*Config)

		//found, err := testGetYDBStreamByID(config, "", rs.Primary.ID)
		//if err != nil {
		//	return err
		//}

		return nil
	}
}

func basicYandexYDBStreamTestStep(streamName string) resource.TestStep {
	return resource.TestStep{
		Config: testYandexYDBStreamBasic(streamName),
		Check: resource.ComposeTestCheckFunc(
			testYandexYDBStreamExists(streamName),
		),
	}
}

func testYandexYDBStreamBasic(name string) string {
	return fmt.Sprintf(`
resource "yandex_ydb_stream" "test-ydb-stream" {
  name = "%s"
}
`, name)
}

func TestAccYandexYDBStream_basic(t *testing.T) {
	var database ydb.Database
	databaseName := acctest.RandomWithPrefix("tf-ydb-database-stream")
	databaseDesc := acctest.RandomWithPrefix("tf-ydb-database-stream-desc")
	labelKey := acctest.RandomWithPrefix("tf-ydb-database-stream-label")
	labelValue := acctest.RandomWithPrefix("tf-ydb-database-stream-label-value")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexYDBDatabaseServerlessDestroy,
		Steps: []resource.TestStep{
			basicYandexYDBDatabaseServerlessTestStep(databaseName, databaseDesc, labelKey, labelValue, &database),
		},
	})
}

func TestAccYandexYDBStream_update(t *testing.T) {
}

func TestAccYandexYDBStream_full(t *testing.T) {
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
			expectedConsumersForDeletion: nil,
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
	}

	for _, v := range testData {
		v := v
		t.Run(v.testName, func(t *testing.T) {
			newReadRules, got := mergeYDBStreamConsumerSettings(v.consumers, v.readRules)
			if !reflect.DeepEqual(newReadRules, v.expectedReadRules) {
				t.Errorf("got readrules %+v\nexpected %+v", newReadRules, v.expectedReadRules)
			}
			if !reflect.DeepEqual(got, v.expectedConsumersForDeletion) {
				t.Errorf("got consumers %v\nexpected %v", got, v.expectedConsumersForDeletion)
			}
		})
	}
}
