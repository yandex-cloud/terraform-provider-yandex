package yandex

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/kafka/v1"
	"google.golang.org/grpc/codes"
)

const (
	kafkaClusterResourceName = "yandex_mdb_kafka_cluster.foo"
)

func TestAccMDBKafkaTopic(t *testing.T) {
	t.Parallel()
	clusterName := acctest.RandomWithPrefix("tf-kafka")
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccMDBKafkaTopicConfigStep1(clusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBKafkaTopicHasPartitions("events", 6),
					testAccCheckMDBKafkaTopicHasReplicationFactor("events", 1),
					testAccCheckMDBKafkaTopicHasConfig("events", &kafka.TopicConfig2_8{
						DeleteRetentionMs: &wrappers.Int64Value{Value: 86400000},
						FlushMs:           &wrappers.Int64Value{Value: 2000},
					}),
					testAccCheckMDBKafkaClusterHasTopic("transactions"),
				),
			},
			mdbKafkaTopicImportStep("yandex_mdb_kafka_topic.events"),
			{
				Config: testAccMDBKafkaTopicConfigStep2(clusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBKafkaTopicHasPartitions("events", 12),
					testAccCheckMDBKafkaTopicHasReplicationFactor("events", 1),
					testAccCheckMDBKafkaTopicHasConfig("events", &kafka.TopicConfig2_8{
						FlushMs:      &wrappers.Int64Value{Value: 4000},
						SegmentBytes: &wrappers.Int64Value{Value: 52428800},
					}),
					testAccCheckMDBKafkaClusterDoesNotHaveTopic("transactions"),
				),
			},
		},
	})
}

func mdbKafkaTopicImportStep(name string) resource.TestStep {
	return resource.TestStep{
		ResourceName:      name,
		ImportState:       true,
		ImportStateVerify: true,
	}
}

func testAccMDBKafkaTopicConfigStep0(name string) string {
	return fmt.Sprintf(kfVPCDependencies+`
resource "yandex_mdb_kafka_cluster" "foo" {
	name        = "%s"
	description = "Kafka Topic Terraform Test"
	environment = "PRODUCTION"
	network_id  = yandex_vpc_network.mdb-kafka-test-net.id
	subnet_ids = [yandex_vpc_subnet.mdb-kafka-test-subnet-a.id]

	config {
	  version          = "2.8"
	  brokers_count    = 1
	  zones            = ["ru-central1-a"]
	  unmanaged_topics = true
	  kafka {
		resources {
		  resource_preset_id = "s2.micro"
		  disk_type_id       = "network-hdd"
		  disk_size          = 16
		}

		kafka_config {
		  log_segment_bytes = 104857600
		}
	  }
	}
}
`, name)
}

func testAccMDBKafkaTopicConfigStep1(name string) string {
	return testAccMDBKafkaTopicConfigStep0(name) + `
resource "yandex_mdb_kafka_topic" events {
  cluster_id         = yandex_mdb_kafka_cluster.foo.id
  name               = "events"
  partitions         = 6
  replication_factor = 1
  topic_config {
    delete_retention_ms = 86400000
    flush_ms            = 2000
  }
}

resource "yandex_mdb_kafka_topic" transactions {
  cluster_id         = yandex_mdb_kafka_cluster.foo.id
  name               = "transactions"
  partitions         = 6
  replication_factor = 1
}
`
}

func testAccMDBKafkaTopicConfigStep2(name string) string {
	return testAccMDBKafkaTopicConfigStep0(name) + `
resource "yandex_mdb_kafka_topic" events {
  cluster_id         = yandex_mdb_kafka_cluster.foo.id
  name               = "events"
  partitions         = 12
  replication_factor = 1
  topic_config {
    flush_ms      = 4000
    segment_bytes = 52428800
  }
}
`
}

func testAccLoadKafkaTopic(s *terraform.State, topicName string) (*kafka.Topic, error) {
	rs, ok := s.RootModule().Resources[kafkaClusterResourceName]
	if !ok {
		return nil, fmt.Errorf("resource %q not found", kafkaClusterResourceName)
	}

	if rs.Primary.ID == "" {
		return nil, fmt.Errorf("no ID is set")
	}

	config := testAccProvider.Meta().(*Config)
	return config.sdk.MDB().Kafka().Topic().Get(context.Background(), &kafka.GetTopicRequest{
		ClusterId: rs.Primary.ID,
		TopicName: topicName,
	})
}

func testAccCheckMDBKafkaClusterDoesNotHaveTopic(topicName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := testAccLoadKafkaTopic(s, topicName)
		if err == nil {
			return fmt.Errorf("expected topic %q to be absent but it exists", topicName)
		}
		if !isStatusWithCode(err, codes.NotFound) {
			return err
		}
		return nil
	}
}

func testAccCheckMDBKafkaClusterHasTopic(topicName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := testAccLoadKafkaTopic(s, topicName)
		return err
	}
}

func testAccCheckMDBKafkaTopicHasPartitions(topicName string, partitions int64) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		topic, err := testAccLoadKafkaTopic(s, topicName)
		if err != nil {
			return err
		}
		v := topic.GetPartitions().GetValue()
		if v != partitions {
			return fmt.Errorf("topic %v has %v partitions, expected: %v", topicName, v, partitions)
		}
		return nil
	}
}

func testAccCheckMDBKafkaTopicHasReplicationFactor(topicName string, replicas int64) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		topic, err := testAccLoadKafkaTopic(s, topicName)
		if err != nil {
			return err
		}
		v := topic.GetReplicationFactor().GetValue()
		if v != replicas {
			return fmt.Errorf("topic %v has replication factor %v, expected: %v", topicName, v, replicas)
		}
		return nil
	}
}

func testAccCheckMDBKafkaTopicHasConfig(topicName string, topicConfig *kafka.TopicConfig2_8) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		topic, err := testAccLoadKafkaTopic(s, topicName)
		if err != nil {
			return err
		}
		actualTopicConfig := topic.GetTopicConfig_2_8()
		if !reflect.DeepEqual(topicConfig, actualTopicConfig) {
			return fmt.Errorf("topic %q has config %v, expected: %v", topicName, actualTopicConfig, topicConfig)
		}
		return nil
	}
}
