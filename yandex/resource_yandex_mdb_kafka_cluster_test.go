package yandex

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"

	terraform2 "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/kafka/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/mocks"
)

const currentDefaultKafkaVersion = "3.5"

const kfResource = "yandex_mdb_kafka_cluster.foo"

const kfVPCDependencies = `
resource "yandex_vpc_network" "mdb-kafka-test-net" {}

resource "yandex_vpc_subnet" "mdb-kafka-test-subnet-a" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.mdb-kafka-test-net.id
  v4_cidr_blocks = ["10.1.0.0/24"]
}

resource "yandex_vpc_subnet" "mdb-kafka-test-subnet-b" {
  zone           = "ru-central1-b"
  network_id     = yandex_vpc_network.mdb-kafka-test-net.id
  v4_cidr_blocks = ["10.2.0.0/24"]
}

resource "yandex_vpc_subnet" "mdb-kafka-test-subnet-d" {
  zone           = "ru-central1-d"
  network_id     = yandex_vpc_network.mdb-kafka-test-net.id
  v4_cidr_blocks = ["10.3.0.0/24"]
}
`

var Versions3x = []string{"3.0", "3.1", "3.2", "3.3", "3.4", "3.5", "3.6"}

func init() {
	resource.AddTestSweepers("yandex_mdb_kafka_cluster", &resource.Sweeper{
		Name: "yandex_mdb_kafka_cluster",
		F:    testSweepMDBKafkaCluster,
	})
}

func testSweepMDBKafkaCluster(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := conf.sdk.MDB().Kafka().Cluster().List(conf.Context(), &kafka.ListClustersRequest{
		FolderId: conf.FolderID,
		PageSize: defaultMDBPageSize,
	})
	if err != nil {
		return fmt.Errorf("error getting Kafka clusters: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.Clusters {
		if !sweepMDBKafkaCluster(conf, c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Kafka cluster %q", c.Id))
		}
	}

	return result.ErrorOrNil()
}

func sweepMDBKafkaCluster(conf *Config, id string) bool {
	return sweepWithRetry(sweepMDBKafkaClusterOnce, conf, "Kafka cluster", id)
}

func sweepMDBKafkaClusterOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexMDBKafkaClusterDeleteTimeout)
	defer cancel()

	mask := field_mask.FieldMask{Paths: []string{"deletion_protection"}}
	op, err := conf.sdk.MDB().Kafka().Cluster().Update(ctx, &kafka.UpdateClusterRequest{
		ClusterId:          id,
		DeletionProtection: false,
		UpdateMask:         &mask,
	})
	err = handleSweepOperation(ctx, conf, op, err)
	if err != nil && !strings.EqualFold(errorMessage(err), "no changes detected") {
		return err
	}

	op, err = conf.sdk.MDB().Kafka().Cluster().Delete(ctx, &kafka.DeleteClusterRequest{
		ClusterId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}

func mdbKafkaClusterImportStep(name string) resource.TestStep {
	return resource.TestStep{
		ResourceName:      name,
		ImportState:       true,
		ImportStateVerify: true,
		ImportStateVerifyIgnore: []string{
			"host",       // prevent failure if health is not the same
			"user",       // passwords are not returned
			"topic",      // order may differs
			"subnet_ids", // subnets not returned
			"health",     // volatile value
		},
	}
}

func TestExpandKafkaClusterConfig(t *testing.T) {
	raw := map[string]interface{}{
		"folder_id":   "",
		"name":        "kafka-tf-name",
		"description": "kafka-tf-desc",
		"environment": "PRESTABLE",
		"labels":      map[string]interface{}{"label1": "val1", "label2": "val2"},
		"config": []interface{}{
			map[string]interface{}{
				"version":         "2.8",
				"brokers_count":   1,
				"zones":           []interface{}{"ru-central1-b", "ru-central1-d"},
				"schema_registry": true,
				"kafka": []interface{}{
					map[string]interface{}{
						"resources": []interface{}{
							map[string]interface{}{
								"resource_preset_id": "s2.micro",
								"disk_size":          20,
								"disk_type_id":       "network-ssd",
							},
						},
						"kafka_config": []interface{}{
							map[string]interface{}{
								"compression_type":                "COMPRESSION_TYPE_ZSTD",
								"log_flush_interval_messages":     1,
								"log_flush_interval_ms":           2,
								"log_flush_scheduler_interval_ms": 3,
								"log_retention_bytes":             4,
								"log_retention_hours":             5,
								"log_retention_minutes":           6,
								"log_retention_ms":                7,
								"log_segment_bytes":               8,
								"socket_send_buffer_bytes":        9,
								"socket_receive_buffer_bytes":     10,
								"auto_create_topics_enable":       true,
								"num_partitions":                  11,
								"default_replication_factor":      12,
								"message_max_bytes":               13,
								"replica_fetch_max_bytes":         14,
								"ssl_cipher_suites":               []interface{}{"TLS_DHE_RSA_WITH_AES_128_CBC_SHA", "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256"},
								"offsets_retention_minutes":       15,
								"sasl_enabled_mechanisms":         []interface{}{"SASL_MECHANISM_SCRAM_SHA_256", "SASL_MECHANISM_SCRAM_SHA_512"},
							},
						},
					},
				},
				"zookeeper": []interface{}{
					map[string]interface{}{
						"resources": []interface{}{
							map[string]interface{}{
								"resource_preset_id": "b2.medium",
								"disk_size":          32,
								"disk_type_id":       "network-ssd",
							},
						},
					},
				},
				"disk_size_autoscaling": []interface{}{
					map[string]interface{}{
						"disk_size_limit":           200,
						"planned_usage_threshold":   50,
						"emergency_usage_threshold": 70,
					},
				},
				"rest_api": []interface{}{
					map[string]interface{}{
						"enabled": true,
					},
				},
				"kafka_ui": []interface{}{
					map[string]interface{}{
						"enabled": true,
					},
				},
			},
		},
		"subnet_ids":         []interface{}{"rc1a-subnet", "rc1b-subnet", "rc1c-subnet"},
		"security_group_ids": []interface{}{"security-group-x", "security-group-y"},
		"host_group_ids":     []interface{}{"hg1", "hg2", "hg3"},
		"maintenance_window": []interface{}{
			map[string]interface{}{
				"type": "WEEKLY",
				"day":  "WED",
				"hour": 2,
			},
		},
		"topic": []interface{}{
			map[string]interface{}{
				"name":               "raw_events",
				"partitions":         12,
				"replication_factor": 1,
				"topic_config": []interface{}{
					map[string]interface{}{
						"cleanup_policy":        "CLEANUP_POLICY_COMPACT_AND_DELETE",
						"compression_type":      "COMPRESSION_TYPE_ZSTD",
						"min_insync_replicas":   1,
						"delete_retention_ms":   2,
						"file_delete_delay_ms":  3,
						"flush_messages":        4,
						"flush_ms":              5,
						"min_compaction_lag_ms": 6,
						"retention_bytes":       7,
						"retention_ms":          8,
						"segment_bytes":         9,
						"max_message_bytes":     16777216,
					},
				},
			},
			map[string]interface{}{
				"name":               "final",
				"partitions":         13,
				"replication_factor": 2,
			},
		},
		"user": []interface{}{
			map[string]interface{}{
				"name":     "alice",
				"password": "password",
				"permission": []interface{}{
					map[string]interface{}{
						"topic_name": "raw_events",
						"role":       "ACCESS_ROLE_PRODUCER",
					},
				},
			},
			map[string]interface{}{
				"name":     "bob",
				"password": "password",
				"permission": []interface{}{
					map[string]interface{}{
						"topic_name": "raw_events",
						"role":       "ACCESS_ROLE_CONSUMER",
					},
					map[string]interface{}{
						"topic_name": "final",
						"role":       "ACCESS_ROLE_PRODUCER",
					},
				},
			},
		},
	}
	resourceData := schema.TestResourceDataRaw(t, resourceYandexMDBKafkaCluster().Schema, raw)

	config := &Config{FolderID: "folder-777"}
	req, err := prepareKafkaCreateRequest(resourceData, config)
	if err != nil {
		require.NoError(t, err)
	}

	expected := &kafka.CreateClusterRequest{
		FolderId:    "folder-777",
		Name:        "kafka-tf-name",
		Description: "kafka-tf-desc",
		Labels:      map[string]string{"label1": "val1", "label2": "val2"},
		Environment: kafka.Cluster_PRESTABLE,
		ConfigSpec: &kafka.ConfigSpec{
			Version:        "2.8",
			BrokersCount:   &wrappers.Int64Value{Value: int64(1)},
			ZoneId:         []string{"ru-central1-b", "ru-central1-d"},
			SchemaRegistry: true,
			Kafka: &kafka.ConfigSpec_Kafka{
				Resources: &kafka.Resources{
					ResourcePresetId: "s2.micro",
					DiskSize:         21474836480,
					DiskTypeId:       "network-ssd",
				},
				KafkaConfig: &kafka.ConfigSpec_Kafka_KafkaConfig_2_8{
					KafkaConfig_2_8: &kafka.KafkaConfig2_8{
						CompressionType:             kafka.CompressionType_COMPRESSION_TYPE_ZSTD,
						LogFlushIntervalMessages:    &wrappers.Int64Value{Value: 1},
						LogFlushIntervalMs:          &wrappers.Int64Value{Value: 2},
						LogFlushSchedulerIntervalMs: &wrappers.Int64Value{Value: 3},
						LogRetentionBytes:           &wrappers.Int64Value{Value: 4},
						LogRetentionHours:           &wrappers.Int64Value{Value: 5},
						LogRetentionMinutes:         &wrappers.Int64Value{Value: 6},
						LogRetentionMs:              &wrappers.Int64Value{Value: 7},
						LogSegmentBytes:             &wrappers.Int64Value{Value: 8},
						SocketSendBufferBytes:       &wrappers.Int64Value{Value: 9},
						SocketReceiveBufferBytes:    &wrappers.Int64Value{Value: 10},
						AutoCreateTopicsEnable:      &wrappers.BoolValue{Value: true},
						NumPartitions:               &wrappers.Int64Value{Value: 11},
						DefaultReplicationFactor:    &wrappers.Int64Value{Value: 12},
						MessageMaxBytes:             &wrappers.Int64Value{Value: 13},
						ReplicaFetchMaxBytes:        &wrappers.Int64Value{Value: 14},
						SslCipherSuites:             []string{"TLS_DHE_RSA_WITH_AES_128_CBC_SHA", "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256"},
						OffsetsRetentionMinutes:     &wrappers.Int64Value{Value: 15},
						SaslEnabledMechanisms: []kafka.SaslMechanism{
							kafka.SaslMechanism_SASL_MECHANISM_SCRAM_SHA_256,
							kafka.SaslMechanism_SASL_MECHANISM_SCRAM_SHA_512,
						},
					},
				},
			},
			Zookeeper: &kafka.ConfigSpec_Zookeeper{
				Resources: &kafka.Resources{
					ResourcePresetId: "b2.medium",
					DiskSize:         34359738368,
					DiskTypeId:       "network-ssd",
				},
			},
			DiskSizeAutoscaling: &kafka.DiskSizeAutoscaling{
				DiskSizeLimit:           200 * 1024 * 1024 * 1024,
				PlannedUsageThreshold:   50,
				EmergencyUsageThreshold: 70,
			},
			RestApiConfig: &kafka.ConfigSpec_RestAPIConfig{
				Enabled: true,
			},
			KafkaUiConfig: &kafka.ConfigSpec_KafkaUIConfig{
				Enabled: true,
			},
		},
		SubnetId: []string{"rc1a-subnet", "rc1b-subnet", "rc1c-subnet"},
		TopicSpecs: []*kafka.TopicSpec{
			{
				Name:              "raw_events",
				Partitions:        &wrappers.Int64Value{Value: int64(12)},
				ReplicationFactor: &wrappers.Int64Value{Value: int64(1)},
				TopicConfig: &kafka.TopicSpec_TopicConfig_2_8{
					TopicConfig_2_8: &kafka.TopicConfig2_8{
						CleanupPolicy:      kafka.TopicConfig2_8_CLEANUP_POLICY_COMPACT_AND_DELETE,
						CompressionType:    kafka.CompressionType_COMPRESSION_TYPE_ZSTD,
						MinInsyncReplicas:  &wrappers.Int64Value{Value: int64(1)},
						DeleteRetentionMs:  &wrappers.Int64Value{Value: int64(2)},
						FileDeleteDelayMs:  &wrappers.Int64Value{Value: int64(3)},
						FlushMessages:      &wrappers.Int64Value{Value: int64(4)},
						FlushMs:            &wrappers.Int64Value{Value: int64(5)},
						MinCompactionLagMs: &wrappers.Int64Value{Value: int64(6)},
						RetentionBytes:     &wrappers.Int64Value{Value: int64(7)},
						RetentionMs:        &wrappers.Int64Value{Value: int64(8)},
						SegmentBytes:       &wrappers.Int64Value{Value: int64(9)},
						MaxMessageBytes:    &wrappers.Int64Value{Value: int64(16777216)},
					},
				},
			},
			{
				Name:              "final",
				Partitions:        &wrappers.Int64Value{Value: int64(13)},
				ReplicationFactor: &wrappers.Int64Value{Value: int64(2)},
			},
		},
		UserSpecs: []*kafka.UserSpec{
			{
				Name:     "alice",
				Password: "password",
				Permissions: []*kafka.Permission{
					{
						TopicName:  "raw_events",
						Role:       kafka.Permission_ACCESS_ROLE_PRODUCER,
						AllowHosts: nil,
					},
				},
			},
			{
				Name:     "bob",
				Password: "password",
				Permissions: []*kafka.Permission{
					{
						TopicName:  "final",
						Role:       kafka.Permission_ACCESS_ROLE_PRODUCER,
						AllowHosts: nil,
					},
					{
						TopicName:  "raw_events",
						Role:       kafka.Permission_ACCESS_ROLE_CONSUMER,
						AllowHosts: nil,
					},
				},
			},
		},
		SecurityGroupIds: []string{"security-group-x", "security-group-y"},
		HostGroupIds:     []string{"hg2", "hg1", "hg3"},
		MaintenanceWindow: &kafka.MaintenanceWindow{
			Policy: &kafka.MaintenanceWindow_WeeklyMaintenanceWindow{
				WeeklyMaintenanceWindow: &kafka.WeeklyMaintenanceWindow{
					Day:  kafka.WeeklyMaintenanceWindow_WED,
					Hour: 2,
				},
			},
		},
	}

	assert.Equal(t, expected, req)
}

func TestExpandKafkaKRaftCombineClusterConfig(t *testing.T) {
	raw := map[string]interface{}{
		"folder_id":   "",
		"name":        "kafka-tf-name",
		"environment": "PRESTABLE",
		"config": []interface{}{
			map[string]interface{}{
				"version":       "3.6",
				"brokers_count": 1,
				"zones":         []interface{}{"ru-central1-a", "ru-central1-b", "ru-central1-d"},
				"kafka": []interface{}{
					map[string]interface{}{
						"resources": []interface{}{
							map[string]interface{}{
								"resource_preset_id": "s2.micro",
								"disk_size":          20,
								"disk_type_id":       "network-ssd",
							},
						},
					},
				},
			},
		},
	}
	resourceData := schema.TestResourceDataRaw(t, resourceYandexMDBKafkaCluster().Schema, raw)

	config := &Config{FolderID: "folder-777"}
	req, err := prepareKafkaCreateRequest(resourceData, config)
	if err != nil {
		require.NoError(t, err)
	}

	expected := &kafka.CreateClusterRequest{
		FolderId:    "folder-777",
		Name:        "kafka-tf-name",
		Labels:      map[string]string{},
		Environment: kafka.Cluster_PRESTABLE,
		ConfigSpec: &kafka.ConfigSpec{
			Version:      "3.6",
			BrokersCount: &wrappers.Int64Value{Value: int64(1)},
			ZoneId:       []string{"ru-central1-a", "ru-central1-b", "ru-central1-d"},
			Kafka: &kafka.ConfigSpec_Kafka{
				Resources: &kafka.Resources{
					ResourcePresetId: "s2.micro",
					DiskSize:         21474836480,
					DiskTypeId:       "network-ssd",
				},
				KafkaConfig: &kafka.ConfigSpec_Kafka_KafkaConfig_3{
					KafkaConfig_3: &kafka.KafkaConfig3{},
				},
			},
		},
		SubnetId:  []string{},
		UserSpecs: []*kafka.UserSpec{},
	}

	assert.Equal(t, expected, req)
}

func TestExpandKafkaKRaftSplitClusterConfig(t *testing.T) {
	raw := map[string]interface{}{
		"folder_id":   "",
		"name":        "kafka-tf-name",
		"environment": "PRESTABLE",
		"config": []interface{}{
			map[string]interface{}{
				"version":       "3.6",
				"brokers_count": 1,
				"zones":         []interface{}{"ru-central1-a", "ru-central1-b", "ru-central1-d"},
				"kafka": []interface{}{
					map[string]interface{}{
						"resources": []interface{}{
							map[string]interface{}{
								"resource_preset_id": "s2.micro",
								"disk_size":          20,
								"disk_type_id":       "network-ssd",
							},
						},
					},
				},
				"kraft": []interface{}{
					map[string]interface{}{
						"resources": []interface{}{
							map[string]interface{}{
								"resource_preset_id": "b2.medium",
								"disk_size":          32,
								"disk_type_id":       "network-ssd",
							},
						},
					},
				},
			},
		},
	}
	resourceData := schema.TestResourceDataRaw(t, resourceYandexMDBKafkaCluster().Schema, raw)

	config := &Config{FolderID: "folder-777"}
	req, err := prepareKafkaCreateRequest(resourceData, config)
	if err != nil {
		require.NoError(t, err)
	}

	expected := &kafka.CreateClusterRequest{
		FolderId:    "folder-777",
		Name:        "kafka-tf-name",
		Labels:      map[string]string{},
		Environment: kafka.Cluster_PRESTABLE,
		ConfigSpec: &kafka.ConfigSpec{
			Version:      "3.6",
			BrokersCount: &wrappers.Int64Value{Value: int64(1)},
			ZoneId:       []string{"ru-central1-a", "ru-central1-b", "ru-central1-d"},
			Kafka: &kafka.ConfigSpec_Kafka{
				Resources: &kafka.Resources{
					ResourcePresetId: "s2.micro",
					DiskSize:         21474836480,
					DiskTypeId:       "network-ssd",
				},
				KafkaConfig: &kafka.ConfigSpec_Kafka_KafkaConfig_3{
					KafkaConfig_3: &kafka.KafkaConfig3{},
				},
			},
			Kraft: &kafka.ConfigSpec_KRaft{
				Resources: &kafka.Resources{
					ResourcePresetId: "b2.medium",
					DiskSize:         34359738368,
					DiskTypeId:       "network-ssd",
				},
			},
		},
		SubnetId:  []string{},
		UserSpecs: []*kafka.UserSpec{},
	}

	assert.Equal(t, expected, req)
}

func TestExpandKafka3xClusterConfig(t *testing.T) {
	for _, version := range Versions3x {
		raw := map[string]interface{}{
			"config": []interface{}{
				map[string]interface{}{
					"version": version,
					"kafka": []interface{}{
						map[string]interface{}{
							"kafka_config": []interface{}{
								map[string]interface{}{
									"compression_type":                "COMPRESSION_TYPE_ZSTD",
									"log_flush_interval_messages":     1,
									"log_flush_interval_ms":           2,
									"log_flush_scheduler_interval_ms": 3,
									"log_retention_bytes":             4,
									"log_retention_hours":             5,
									"log_retention_minutes":           6,
									"log_retention_ms":                7,
									"log_segment_bytes":               8,
									"socket_send_buffer_bytes":        9,
									"socket_receive_buffer_bytes":     10,
									"auto_create_topics_enable":       true,
									"num_partitions":                  11,
									"default_replication_factor":      12,
									"message_max_bytes":               13,
									"replica_fetch_max_bytes":         14,
									"ssl_cipher_suites":               []interface{}{"TLS_DHE_RSA_WITH_AES_128_CBC_SHA", "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256"},
									"offsets_retention_minutes":       15,
									"sasl_enabled_mechanisms":         []interface{}{"SASL_MECHANISM_SCRAM_SHA_256", "SASL_MECHANISM_SCRAM_SHA_512"},
								},
							},
						},
					},
				},
			},
			"topic": []interface{}{
				map[string]interface{}{
					"name":               "raw_events",
					"partitions":         12,
					"replication_factor": 1,
					"topic_config": []interface{}{
						map[string]interface{}{
							"cleanup_policy":        "CLEANUP_POLICY_COMPACT_AND_DELETE",
							"compression_type":      "COMPRESSION_TYPE_ZSTD",
							"min_insync_replicas":   1,
							"delete_retention_ms":   2,
							"file_delete_delay_ms":  3,
							"flush_messages":        4,
							"flush_ms":              5,
							"min_compaction_lag_ms": 6,
							"retention_bytes":       7,
							"retention_ms":          8,
							"segment_bytes":         9,
							"max_message_bytes":     16777216,
						},
					},
				},
			},
		}
		resourceData := schema.TestResourceDataRaw(t, resourceYandexMDBKafkaCluster().Schema, raw)

		config := &Config{FolderID: "folder-777"}
		req, err := prepareKafkaCreateRequest(resourceData, config)
		if err != nil {
			require.NoError(t, err)
		}

		assert.Equal(t, version, req.ConfigSpec.Version)

		assert.Equal(t, &kafka.ConfigSpec_Kafka_KafkaConfig_3{
			KafkaConfig_3: &kafka.KafkaConfig3{
				CompressionType:             kafka.CompressionType_COMPRESSION_TYPE_ZSTD,
				LogFlushIntervalMessages:    &wrappers.Int64Value{Value: 1},
				LogFlushIntervalMs:          &wrappers.Int64Value{Value: 2},
				LogFlushSchedulerIntervalMs: &wrappers.Int64Value{Value: 3},
				LogRetentionBytes:           &wrappers.Int64Value{Value: 4},
				LogRetentionHours:           &wrappers.Int64Value{Value: 5},
				LogRetentionMinutes:         &wrappers.Int64Value{Value: 6},
				LogRetentionMs:              &wrappers.Int64Value{Value: 7},
				LogSegmentBytes:             &wrappers.Int64Value{Value: 8},
				SocketSendBufferBytes:       &wrappers.Int64Value{Value: 9},
				SocketReceiveBufferBytes:    &wrappers.Int64Value{Value: 10},
				AutoCreateTopicsEnable:      &wrappers.BoolValue{Value: true},
				NumPartitions:               &wrappers.Int64Value{Value: 11},
				DefaultReplicationFactor:    &wrappers.Int64Value{Value: 12},
				MessageMaxBytes:             &wrappers.Int64Value{Value: 13},
				ReplicaFetchMaxBytes:        &wrappers.Int64Value{Value: 14},
				SslCipherSuites:             []string{"TLS_DHE_RSA_WITH_AES_128_CBC_SHA", "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256"},
				OffsetsRetentionMinutes:     &wrappers.Int64Value{Value: 15},
				SaslEnabledMechanisms: []kafka.SaslMechanism{
					kafka.SaslMechanism_SASL_MECHANISM_SCRAM_SHA_256,
					kafka.SaslMechanism_SASL_MECHANISM_SCRAM_SHA_512,
				},
			},
		}, req.ConfigSpec.Kafka.KafkaConfig)

		assert.Equal(t, &kafka.TopicSpec{
			Name:              "raw_events",
			Partitions:        &wrappers.Int64Value{Value: int64(12)},
			ReplicationFactor: &wrappers.Int64Value{Value: int64(1)},
			TopicConfig: &kafka.TopicSpec_TopicConfig_3{
				TopicConfig_3: &kafka.TopicConfig3{
					CleanupPolicy:      kafka.TopicConfig3_CLEANUP_POLICY_COMPACT_AND_DELETE,
					CompressionType:    kafka.CompressionType_COMPRESSION_TYPE_ZSTD,
					MinInsyncReplicas:  &wrappers.Int64Value{Value: int64(1)},
					DeleteRetentionMs:  &wrappers.Int64Value{Value: int64(2)},
					FileDeleteDelayMs:  &wrappers.Int64Value{Value: int64(3)},
					FlushMessages:      &wrappers.Int64Value{Value: int64(4)},
					FlushMs:            &wrappers.Int64Value{Value: int64(5)},
					MinCompactionLagMs: &wrappers.Int64Value{Value: int64(6)},
					RetentionBytes:     &wrappers.Int64Value{Value: int64(7)},
					RetentionMs:        &wrappers.Int64Value{Value: int64(8)},
					SegmentBytes:       &wrappers.Int64Value{Value: int64(9)},
					MaxMessageBytes:    &wrappers.Int64Value{Value: int64(16777216)},
				},
			},
		}, req.TopicSpecs[0])
	}
}

func TestPrepareKafkaCreateRequestWhenTopicConfigHasNotAllFieldsShouldConvertItProperly(t *testing.T) {
	raw := map[string]interface{}{
		"config": []interface{}{
			map[string]interface{}{
				"version": "2.8",
			},
		},
		"topic": []interface{}{
			map[string]interface{}{
				"name":               "raw_events",
				"partitions":         12,
				"replication_factor": 1,
				"topic_config": []interface{}{
					map[string]interface{}{
						"cleanup_policy": "CLEANUP_POLICY_COMPACT_AND_DELETE",
					},
				},
			},
		},
	}
	resourceData := schema.TestResourceDataRaw(t, resourceYandexMDBKafkaCluster().Schema, raw)

	config := &Config{FolderID: "folder-777"}
	req, err := prepareKafkaCreateRequest(resourceData, config)
	require.NoError(t, err)

	assert.Equal(t, &kafka.TopicSpec{
		Name:              "raw_events",
		Partitions:        &wrappers.Int64Value{Value: int64(12)},
		ReplicationFactor: &wrappers.Int64Value{Value: int64(1)},
		TopicConfig: &kafka.TopicSpec_TopicConfig_2_8{
			TopicConfig_2_8: &kafka.TopicConfig2_8{
				CleanupPolicy: kafka.TopicConfig2_8_CLEANUP_POLICY_COMPACT_AND_DELETE,
			},
		},
	}, req.TopicSpecs[0])
}

func TestKafkaClusterUpdateRequest(t *testing.T) {
	raw := map[string]interface{}{
		"name":        "new-name",
		"description": "new description",
		"labels":      map[string]interface{}{"label1": "val1", "label2": "val2"},
		"network_id":  "mynetwork_id",
		"config": []interface{}{
			map[string]interface{}{
				"version":       "2.8",
				"brokers_count": 1,
				"zones":         []interface{}{"ru-central1-b", "ru-central1-d"},
				"kafka": []interface{}{
					map[string]interface{}{
						"resources": []interface{}{
							map[string]interface{}{
								"resource_preset_id": "s2.micro",
								"disk_size":          20,
								"disk_type_id":       "network-ssd",
							},
						},
						"kafka_config": []interface{}{
							map[string]interface{}{
								"compression_type":                "COMPRESSION_TYPE_ZSTD",
								"log_flush_interval_messages":     1,
								"log_flush_interval_ms":           2,
								"log_flush_scheduler_interval_ms": 3,
								"log_retention_bytes":             4,
								"log_retention_hours":             5,
								"log_retention_minutes":           6,
								"log_retention_ms":                7,
								"log_segment_bytes":               8,
								"socket_send_buffer_bytes":        9,
								"socket_receive_buffer_bytes":     10,
								"auto_create_topics_enable":       true,
								"num_partitions":                  11,
								"default_replication_factor":      12,
								"message_max_bytes":               13,
								"replica_fetch_max_bytes":         14,
								"ssl_cipher_suites":               []interface{}{"TLS_DHE_RSA_WITH_AES_128_CBC_SHA", "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256"},
								"offsets_retention_minutes":       15,
								"sasl_enabled_mechanisms":         []interface{}{"SASL_MECHANISM_SCRAM_SHA_256", "SASL_MECHANISM_SCRAM_SHA_512"},
							},
						},
					},
				},
				"zookeeper": []interface{}{
					map[string]interface{}{
						"resources": []interface{}{
							map[string]interface{}{
								"resource_preset_id": "b2.medium",
								"disk_size":          32,
								"disk_type_id":       "network-ssd",
							},
						},
					},
				},
				"disk_size_autoscaling": []interface{}{
					map[string]interface{}{
						"disk_size_limit":           200,
						"planned_usage_threshold":   50,
						"emergency_usage_threshold": 70,
					},
				},
				"rest_api": []interface{}{
					map[string]interface{}{
						"enabled": true,
					},
				},
				"kafka_ui": []interface{}{
					map[string]interface{}{
						"enabled": true,
					},
				},
			},
		},
		"subnet_ids":         []interface{}{"rc1a-subnet", "rc1b-subnet", "rc1c-subnet"},
		"security_group_ids": []interface{}{"security-group-x", "security-group-y"},
		"host_group_ids":     []interface{}{"hg1", "hg2", "hg3"},
		"maintenance_window": []interface{}{
			map[string]interface{}{
				"type": "ANYTIME",
			},
		},
	}
	resourceData := schema.TestResourceDataRaw(t, resourceYandexMDBKafkaCluster().Schema, raw)

	config := &Config{}
	req, err := kafkaClusterUpdateRequestWithMask(resourceData, config)
	require.NoError(t, err)

	expected := &kafka.UpdateClusterRequest{
		Name:        "new-name",
		Description: "new description",
		Labels:      map[string]string{"label1": "val1", "label2": "val2"},
		NetworkId:   "mynetwork_id",
		ConfigSpec: &kafka.ConfigSpec{
			Version:      "2.8",
			BrokersCount: &wrappers.Int64Value{Value: int64(1)},
			ZoneId:       []string{"ru-central1-b", "ru-central1-d"},
			Kafka: &kafka.ConfigSpec_Kafka{
				Resources: &kafka.Resources{
					ResourcePresetId: "s2.micro",
					DiskSize:         21474836480,
					DiskTypeId:       "network-ssd",
				},
				KafkaConfig: &kafka.ConfigSpec_Kafka_KafkaConfig_2_8{
					KafkaConfig_2_8: &kafka.KafkaConfig2_8{
						CompressionType:             kafka.CompressionType_COMPRESSION_TYPE_ZSTD,
						LogFlushIntervalMessages:    &wrappers.Int64Value{Value: 1},
						LogFlushIntervalMs:          &wrappers.Int64Value{Value: 2},
						LogFlushSchedulerIntervalMs: &wrappers.Int64Value{Value: 3},
						LogRetentionBytes:           &wrappers.Int64Value{Value: 4},
						LogRetentionHours:           &wrappers.Int64Value{Value: 5},
						LogRetentionMinutes:         &wrappers.Int64Value{Value: 6},
						LogRetentionMs:              &wrappers.Int64Value{Value: 7},
						LogSegmentBytes:             &wrappers.Int64Value{Value: 8},
						SocketSendBufferBytes:       &wrappers.Int64Value{Value: 9},
						SocketReceiveBufferBytes:    &wrappers.Int64Value{Value: 10},
						AutoCreateTopicsEnable:      &wrappers.BoolValue{Value: true},
						NumPartitions:               &wrappers.Int64Value{Value: 11},
						DefaultReplicationFactor:    &wrappers.Int64Value{Value: 12},
						MessageMaxBytes:             &wrappers.Int64Value{Value: 13},
						ReplicaFetchMaxBytes:        &wrappers.Int64Value{Value: 14},
						SslCipherSuites:             []string{"TLS_DHE_RSA_WITH_AES_128_CBC_SHA", "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256"},
						OffsetsRetentionMinutes:     &wrappers.Int64Value{Value: 15},
						SaslEnabledMechanisms: []kafka.SaslMechanism{
							kafka.SaslMechanism_SASL_MECHANISM_SCRAM_SHA_256,
							kafka.SaslMechanism_SASL_MECHANISM_SCRAM_SHA_512,
						},
					},
				},
			},
			Zookeeper: &kafka.ConfigSpec_Zookeeper{
				Resources: &kafka.Resources{
					ResourcePresetId: "b2.medium",
					DiskSize:         34359738368,
					DiskTypeId:       "network-ssd",
				},
			},
			DiskSizeAutoscaling: &kafka.DiskSizeAutoscaling{
				DiskSizeLimit:           200 * 1024 * 1024 * 1024,
				PlannedUsageThreshold:   50,
				EmergencyUsageThreshold: 70,
			},
			RestApiConfig: &kafka.ConfigSpec_RestAPIConfig{
				Enabled: true,
			},
			KafkaUiConfig: &kafka.ConfigSpec_KafkaUIConfig{
				Enabled: true,
			},
		},
		SecurityGroupIds: []string{"security-group-x", "security-group-y"},
		MaintenanceWindow: &kafka.MaintenanceWindow{
			Policy: &kafka.MaintenanceWindow_Anytime{
				Anytime: &kafka.AnytimeMaintenanceWindow{},
			},
		},
		SubnetIds: []string{"rc1a-subnet", "rc1b-subnet", "rc1c-subnet"},
		UpdateMask: &field_mask.FieldMask{Paths: []string{
			"config_spec.brokers_count",
			"config_spec.disk_size_autoscaling.disk_size_limit",
			"config_spec.disk_size_autoscaling.emergency_usage_threshold",
			"config_spec.disk_size_autoscaling.planned_usage_threshold",
			"config_spec.kafka.kafka_config_2_8.auto_create_topics_enable",
			"config_spec.kafka.kafka_config_2_8.compression_type",
			"config_spec.kafka.kafka_config_2_8.default_replication_factor",
			"config_spec.kafka.kafka_config_2_8.log_flush_interval_messages",
			"config_spec.kafka.kafka_config_2_8.log_flush_interval_ms",
			"config_spec.kafka.kafka_config_2_8.log_flush_scheduler_interval_ms",
			"config_spec.kafka.kafka_config_2_8.log_retention_bytes",
			"config_spec.kafka.kafka_config_2_8.log_retention_hours",
			"config_spec.kafka.kafka_config_2_8.log_retention_minutes",
			"config_spec.kafka.kafka_config_2_8.log_retention_ms",
			"config_spec.kafka.kafka_config_2_8.log_segment_bytes",
			"config_spec.kafka.kafka_config_2_8.message_max_bytes",
			"config_spec.kafka.kafka_config_2_8.num_partitions",
			"config_spec.kafka.kafka_config_2_8.offsets_retention_minutes",
			"config_spec.kafka.kafka_config_2_8.replica_fetch_max_bytes",
			"config_spec.kafka.kafka_config_2_8.sasl_enabled_mechanisms",
			"config_spec.kafka.kafka_config_2_8.socket_receive_buffer_bytes",
			"config_spec.kafka.kafka_config_2_8.socket_send_buffer_bytes",
			"config_spec.kafka.kafka_config_2_8.ssl_cipher_suites",
			"config_spec.kafka.resources.disk_size",
			"config_spec.kafka.resources.disk_type_id",
			"config_spec.kafka.resources.resource_preset_id",
			"config_spec.kafka_ui_config.enabled",
			"config_spec.rest_api_config.enabled",
			"config_spec.version",
			"config_spec.zone_id",
			"config_spec.zookeeper.resources.disk_size",
			"config_spec.zookeeper.resources.disk_type_id",
			"config_spec.zookeeper.resources.resource_preset_id",
			"description",
			"labels",
			"maintenance_window",
			"name",
			"network_id",
			"security_group_ids",
			"subnet_ids",
		}},
	}

	assert.Equal(t, expected, req)
}

func TestKafka3xClusterUpdateRequest(t *testing.T) {
	for _, version := range Versions3x {
		raw := map[string]interface{}{
			"name":        "new-name",
			"description": "new description",
			"labels":      map[string]interface{}{"label1": "val1", "label2": "val2"},
			"config": []interface{}{
				map[string]interface{}{
					"version":       version,
					"brokers_count": 1,
					"zones":         []interface{}{"ru-central1-b", "ru-central1-d"},
					"kafka": []interface{}{
						map[string]interface{}{
							"resources": []interface{}{
								map[string]interface{}{
									"resource_preset_id": "s2.micro",
									"disk_size":          20,
									"disk_type_id":       "network-ssd",
								},
							},
							"kafka_config": []interface{}{
								map[string]interface{}{
									"compression_type":                "COMPRESSION_TYPE_ZSTD",
									"log_flush_interval_messages":     1,
									"log_flush_interval_ms":           2,
									"log_flush_scheduler_interval_ms": 3,
									"log_retention_bytes":             4,
									"log_retention_hours":             5,
									"log_retention_minutes":           6,
									"log_retention_ms":                7,
									"log_segment_bytes":               8,
									"socket_send_buffer_bytes":        9,
									"socket_receive_buffer_bytes":     10,
									"auto_create_topics_enable":       true,
									"num_partitions":                  11,
									"default_replication_factor":      12,
									"message_max_bytes":               13,
									"replica_fetch_max_bytes":         14,
									"ssl_cipher_suites":               []interface{}{"TLS_DHE_RSA_WITH_AES_128_CBC_SHA", "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256"},
									"offsets_retention_minutes":       15,
									"sasl_enabled_mechanisms":         []interface{}{"SASL_MECHANISM_SCRAM_SHA_256", "SASL_MECHANISM_SCRAM_SHA_512"},
								},
							},
						},
					},
					"zookeeper": []interface{}{
						map[string]interface{}{
							"resources": []interface{}{
								map[string]interface{}{
									"resource_preset_id": "b2.medium",
									"disk_size":          32,
									"disk_type_id":       "network-ssd",
								},
							},
						},
					},
					"kraft": []interface{}{
						map[string]interface{}{
							"resources": []interface{}{
								map[string]interface{}{
									"resource_preset_id": "b2.medium",
									"disk_size":          32,
									"disk_type_id":       "network-ssd",
								},
							},
						},
					},
					"disk_size_autoscaling": []interface{}{
						map[string]interface{}{
							"disk_size_limit":           200,
							"planned_usage_threshold":   50,
							"emergency_usage_threshold": 70,
						},
					},
				},
			},
			"subnet_ids":         []interface{}{"rc1a-subnet", "rc1b-subnet", "rc1c-subnet"},
			"security_group_ids": []interface{}{"security-group-x", "security-group-y"},
			"host_group_ids":     []interface{}{"hg1", "hg2", "hg3"},
			"maintenance_window": []interface{}{
				map[string]interface{}{
					"type": "ANYTIME",
				},
			},
		}
		resourceData := schema.TestResourceDataRaw(t, resourceYandexMDBKafkaCluster().Schema, raw)

		config := &Config{}
		req, err := kafkaClusterUpdateRequestWithMask(resourceData, config)
		require.NoError(t, err)

		expected := &kafka.UpdateClusterRequest{
			Name:        "new-name",
			Description: "new description",
			Labels:      map[string]string{"label1": "val1", "label2": "val2"},
			ConfigSpec: &kafka.ConfigSpec{
				Version:      version,
				BrokersCount: &wrappers.Int64Value{Value: int64(1)},
				ZoneId:       []string{"ru-central1-b", "ru-central1-d"},
				Kafka: &kafka.ConfigSpec_Kafka{
					Resources: &kafka.Resources{
						ResourcePresetId: "s2.micro",
						DiskSize:         21474836480,
						DiskTypeId:       "network-ssd",
					},
					KafkaConfig: &kafka.ConfigSpec_Kafka_KafkaConfig_3{
						KafkaConfig_3: &kafka.KafkaConfig3{
							CompressionType:             kafka.CompressionType_COMPRESSION_TYPE_ZSTD,
							LogFlushIntervalMessages:    &wrappers.Int64Value{Value: 1},
							LogFlushIntervalMs:          &wrappers.Int64Value{Value: 2},
							LogFlushSchedulerIntervalMs: &wrappers.Int64Value{Value: 3},
							LogRetentionBytes:           &wrappers.Int64Value{Value: 4},
							LogRetentionHours:           &wrappers.Int64Value{Value: 5},
							LogRetentionMinutes:         &wrappers.Int64Value{Value: 6},
							LogRetentionMs:              &wrappers.Int64Value{Value: 7},
							LogSegmentBytes:             &wrappers.Int64Value{Value: 8},
							SocketSendBufferBytes:       &wrappers.Int64Value{Value: 9},
							SocketReceiveBufferBytes:    &wrappers.Int64Value{Value: 10},
							AutoCreateTopicsEnable:      &wrappers.BoolValue{Value: true},
							NumPartitions:               &wrappers.Int64Value{Value: 11},
							DefaultReplicationFactor:    &wrappers.Int64Value{Value: 12},
							MessageMaxBytes:             &wrappers.Int64Value{Value: 13},
							ReplicaFetchMaxBytes:        &wrappers.Int64Value{Value: 14},
							SslCipherSuites:             []string{"TLS_DHE_RSA_WITH_AES_128_CBC_SHA", "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256"},
							OffsetsRetentionMinutes:     &wrappers.Int64Value{Value: 15},
							SaslEnabledMechanisms: []kafka.SaslMechanism{
								kafka.SaslMechanism_SASL_MECHANISM_SCRAM_SHA_256,
								kafka.SaslMechanism_SASL_MECHANISM_SCRAM_SHA_512,
							},
						},
					},
				},
				Zookeeper: &kafka.ConfigSpec_Zookeeper{
					Resources: &kafka.Resources{
						ResourcePresetId: "b2.medium",
						DiskSize:         34359738368,
						DiskTypeId:       "network-ssd",
					},
				},
				Kraft: &kafka.ConfigSpec_KRaft{
					Resources: &kafka.Resources{
						ResourcePresetId: "b2.medium",
						DiskSize:         34359738368,
						DiskTypeId:       "network-ssd",
					},
				},
				DiskSizeAutoscaling: &kafka.DiskSizeAutoscaling{
					DiskSizeLimit:           200 * 1024 * 1024 * 1024,
					PlannedUsageThreshold:   50,
					EmergencyUsageThreshold: 70,
				},
			},
			SecurityGroupIds: []string{"security-group-x", "security-group-y"},
			MaintenanceWindow: &kafka.MaintenanceWindow{
				Policy: &kafka.MaintenanceWindow_Anytime{
					Anytime: &kafka.AnytimeMaintenanceWindow{},
				},
			},
			SubnetIds: []string{"rc1a-subnet", "rc1b-subnet", "rc1c-subnet"},
			UpdateMask: &field_mask.FieldMask{Paths: []string{
				"config_spec.brokers_count",
				"config_spec.disk_size_autoscaling.disk_size_limit",
				"config_spec.disk_size_autoscaling.emergency_usage_threshold",
				"config_spec.disk_size_autoscaling.planned_usage_threshold",
				"config_spec.kafka.kafka_config_3.auto_create_topics_enable",
				"config_spec.kafka.kafka_config_3.compression_type",
				"config_spec.kafka.kafka_config_3.default_replication_factor",
				"config_spec.kafka.kafka_config_3.log_flush_interval_messages",
				"config_spec.kafka.kafka_config_3.log_flush_interval_ms",
				"config_spec.kafka.kafka_config_3.log_flush_scheduler_interval_ms",
				"config_spec.kafka.kafka_config_3.log_retention_bytes",
				"config_spec.kafka.kafka_config_3.log_retention_hours",
				"config_spec.kafka.kafka_config_3.log_retention_minutes",
				"config_spec.kafka.kafka_config_3.log_retention_ms",
				"config_spec.kafka.kafka_config_3.log_segment_bytes",
				"config_spec.kafka.kafka_config_3.message_max_bytes",
				"config_spec.kafka.kafka_config_3.num_partitions",
				"config_spec.kafka.kafka_config_3.offsets_retention_minutes",
				"config_spec.kafka.kafka_config_3.replica_fetch_max_bytes",
				"config_spec.kafka.kafka_config_3.sasl_enabled_mechanisms",
				"config_spec.kafka.kafka_config_3.socket_receive_buffer_bytes",
				"config_spec.kafka.kafka_config_3.socket_send_buffer_bytes",
				"config_spec.kafka.kafka_config_3.ssl_cipher_suites",
				"config_spec.kafka.resources.disk_size",
				"config_spec.kafka.resources.disk_type_id",
				"config_spec.kafka.resources.resource_preset_id",
				"config_spec.kraft.resources.disk_size",
				"config_spec.kraft.resources.disk_type_id",
				"config_spec.kraft.resources.resource_preset_id",
				"config_spec.version",
				"config_spec.zone_id",
				"config_spec.zookeeper.resources.disk_size",
				"config_spec.zookeeper.resources.disk_type_id",
				"config_spec.zookeeper.resources.resource_preset_id",
				"description",
				"labels",
				"maintenance_window",
				"name",
				"security_group_ids",
				"subnet_ids",
			}},
		}

		assert.Equal(t, expected, req)
	}
}

func TestUpdateKafkaClusterTopics(t *testing.T) {
	rawInitial := map[string]interface{}{
		"config": []interface{}{
			map[string]interface{}{"version": "2.8"},
		},
		"topic": []interface{}{
			map[string]interface{}{"name": "deletedTopic"},
			map[string]interface{}{
				"name":       "sameTopic",
				"partitions": 1,
			},
			map[string]interface{}{
				"name":               "updatedTopic",
				"partitions":         1,
				"replication_factor": 3,
				"topic_config": []interface{}{
					map[string]interface{}{
						"cleanup_policy": "CLEANUP_POLICY_COMPACT_AND_DELETE",
					},
				},
			},
		},
	}
	diffAttributes := map[string]*terraform2.ResourceAttrDiff{
		"topic.#":                               {New: "3"},
		"topic.0.name":                          {New: "sameTopic"},
		"topic.0.partitions":                    {New: "1"},
		"topic.1.name":                          {New: "updatedTopic"},
		"topic.1.partitions":                    {New: "2"},
		"topic.1.replication_factor":            {New: "3"},
		"topic.1.topic_config.#":                {New: "1"},
		"topic.1.topic_config.0.cleanup_policy": {New: "CLEANUP_POLICY_COMPACT"},
		"topic.2.name":                          {New: "newTopic"},
		"topic.2.partitions":                    {New: "2"},
		"topic.2.replication_factor":            {New: "3"},
		"topic.2.topic_config.#":                {New: "1"},
		"topic.2.topic_config.0.cleanup_policy": {New: "CLEANUP_POLICY_COMPACT_AND_DELETE"},
	}
	resourceData := CreateResourceData(t, resourceYandexMDBKafkaCluster().Schema, rawInitial, diffAttributes)

	ctrl := gomock.NewController(t)
	topicModifier := mocks.NewMockKafkaTopicModifier(ctrl)
	topicModifier.EXPECT().DeleteKafkaTopic(gomock.Any(), resourceData, "deletedTopic").Return(nil).Times(1)
	topicModifier.EXPECT().CreateKafkaTopic(gomock.Any(), resourceData, gomock.Any()).DoAndReturn(
		func(ctx context.Context, d *schema.ResourceData, topicSpec *kafka.TopicSpec) error {
			require.Equal(t, (&kafka.TopicSpec{
				Name:              "newTopic",
				Partitions:        &wrapperspb.Int64Value{Value: 2},
				ReplicationFactor: &wrapperspb.Int64Value{Value: 3},
				TopicConfig: &kafka.TopicSpec_TopicConfig_2_8{
					TopicConfig_2_8: &kafka.TopicConfig2_8{
						CleanupPolicy: kafka.TopicConfig2_8_CLEANUP_POLICY_COMPACT_AND_DELETE,
					},
				},
			}).String(), topicSpec.String())
			return nil
		}).Return(nil).Times(1)
	topicModifier.EXPECT().UpdateKafkaTopic(gomock.Any(), resourceData, gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, d *schema.ResourceData, topicSpec *kafka.TopicSpec, paths []string) error {
			require.Equal(t, []string{"topic_spec.partitions", "topic_spec.topic_config_2_8.cleanup_policy"}, paths)
			require.Equal(t, (&kafka.TopicSpec{
				Name:              "updatedTopic",
				Partitions:        &wrapperspb.Int64Value{Value: 2},
				ReplicationFactor: &wrapperspb.Int64Value{Value: 3},
				TopicConfig: &kafka.TopicSpec_TopicConfig_2_8{
					TopicConfig_2_8: &kafka.TopicConfig2_8{
						CleanupPolicy: kafka.TopicConfig2_8_CLEANUP_POLICY_COMPACT,
					},
				},
			}).String(), topicSpec.String())
			return nil
		}).Times(1)

	err := updateKafkaClusterTopics(resourceData, topicModifier)

	require.NoError(t, err)
}

func TestUpdateKafka3xClusterTopics(t *testing.T) {
	for _, version := range Versions3x {
		rawInitial := map[string]interface{}{
			"config": []interface{}{
				map[string]interface{}{"version": version},
			},
			"topic": []interface{}{
				map[string]interface{}{"name": "deletedTopic"},
				map[string]interface{}{
					"name":       "sameTopic",
					"partitions": 1,
				},
				map[string]interface{}{
					"name":               "updatedTopic",
					"partitions":         1,
					"replication_factor": 3,
					"topic_config": []interface{}{
						map[string]interface{}{
							"cleanup_policy": "CLEANUP_POLICY_COMPACT_AND_DELETE",
						},
					},
				},
			},
		}
		diffAttributes := map[string]*terraform2.ResourceAttrDiff{
			"topic.#":                               {New: "3"},
			"topic.0.name":                          {New: "sameTopic"},
			"topic.0.partitions":                    {New: "1"},
			"topic.1.name":                          {New: "updatedTopic"},
			"topic.1.partitions":                    {New: "2"},
			"topic.1.replication_factor":            {New: "3"},
			"topic.1.topic_config.#":                {New: "1"},
			"topic.1.topic_config.0.cleanup_policy": {New: "CLEANUP_POLICY_COMPACT"},
			"topic.2.name":                          {New: "newTopic"},
			"topic.2.partitions":                    {New: "2"},
			"topic.2.replication_factor":            {New: "3"},
			"topic.2.topic_config.#":                {New: "1"},
			"topic.2.topic_config.0.cleanup_policy": {New: "CLEANUP_POLICY_COMPACT_AND_DELETE"},
		}
		resourceData := CreateResourceData(t, resourceYandexMDBKafkaCluster().Schema, rawInitial, diffAttributes)

		ctrl := gomock.NewController(t)
		topicModifier := mocks.NewMockKafkaTopicModifier(ctrl)
		topicModifier.EXPECT().DeleteKafkaTopic(gomock.Any(), resourceData, "deletedTopic").Return(nil).Times(1)
		topicModifier.EXPECT().CreateKafkaTopic(gomock.Any(), resourceData, gomock.Any()).DoAndReturn(
			func(ctx context.Context, d *schema.ResourceData, topicSpec *kafka.TopicSpec) error {
				require.Equal(t, (&kafka.TopicSpec{
					Name:              "newTopic",
					Partitions:        &wrapperspb.Int64Value{Value: 2},
					ReplicationFactor: &wrapperspb.Int64Value{Value: 3},
					TopicConfig: &kafka.TopicSpec_TopicConfig_3{
						TopicConfig_3: &kafka.TopicConfig3{
							CleanupPolicy: kafka.TopicConfig3_CLEANUP_POLICY_COMPACT_AND_DELETE,
						},
					},
				}).String(), topicSpec.String())
				return nil
			}).Return(nil).Times(1)
		topicModifier.EXPECT().UpdateKafkaTopic(gomock.Any(), resourceData, gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, d *schema.ResourceData, topicSpec *kafka.TopicSpec, paths []string) error {
				require.Equal(t, []string{"topic_spec.partitions", "topic_spec.topic_config_3.cleanup_policy"}, paths)
				require.Equal(t, (&kafka.TopicSpec{
					Name:              "updatedTopic",
					Partitions:        &wrapperspb.Int64Value{Value: 2},
					ReplicationFactor: &wrapperspb.Int64Value{Value: 3},
					TopicConfig: &kafka.TopicSpec_TopicConfig_3{
						TopicConfig_3: &kafka.TopicConfig3{
							CleanupPolicy: kafka.TopicConfig3_CLEANUP_POLICY_COMPACT,
						},
					},
				}).String(), topicSpec.String())
				return nil
			}).Times(1)

		err := updateKafkaClusterTopics(resourceData, topicModifier)

		require.NoError(t, err)
	}
}

// Test that a Kafka Cluster can be created, updated and destroyed in single zone mode
func TestAccMDBKafkaCluster_single(t *testing.T) {
	t.Parallel()

	var r kafka.Cluster
	kfName := acctest.RandomWithPrefix("tf-kafka")
	kfDesc := "Kafka Cluster Terraform Test"
	kfDescUpdated := "Kafka Cluster Terraform Test (updated)"
	folderID := getExampleFolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMDBKafkaClusterDestroy,
		Steps: []resource.TestStep{
			// Create Kafka Cluster
			{
				Config: testAccMDBKafkaClusterConfigMain(kfName, kfDesc, "PRESTABLE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBKafkaClusterExists(kfResource, &r, 1),
					resource.TestCheckResourceAttr(kfResource, "name", kfName),
					resource.TestCheckResourceAttr(kfResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(kfResource, "description", kfDesc),
					resource.TestCheckResourceAttr(kfResource, "deletion_protection", "false"),
					resource.TestCheckResourceAttr(kfResource, "config.0.version", currentDefaultKafkaVersion),
					resource.TestCheckResourceAttr(kfResource, "config.0.access.0.data_transfer", "true"),
					resource.TestCheckResourceAttr(kfResource, "config.0.rest_api.0.enabled", "false"),
					resource.TestCheckResourceAttr(kfResource, "config.0.kafka_ui.0.enabled", "false"),
					testAccCheckMDBKafkaClusterContainsLabel(&r, "test_key", "test_value"),
					testAccCheckMDBKafkaConfigKafkaHasResources(&r, "s2.micro", "network-hdd", 16*1024*1024*1024),
					testAccCheckMDBKafkaClusterHasTopics(kfResource, []string{"raw_events", "final"}),
					testAccCheckMDBKafkaClusterHasUsers(kfResource, map[string][]string{"alice": {"raw_events"}, "bob": {"raw_events", "final"}}),
					testAccCheckMDBKafkaClusterCompressionType(&r, kafka.CompressionType_COMPRESSION_TYPE_ZSTD),
					testAccCheckMDBKafkaClusterLogRetentionBytes(&r, 1073741824),
					testAccCheckMDBKafkaClusterSaslEnabledMechanisms(&r, []kafka.SaslMechanism{kafka.SaslMechanism_SASL_MECHANISM_SCRAM_SHA_256}),
					testAccCheckMDBKafkaTopicMaxMessageBytes(kfResource, "raw_events", 777216),
					testAccCheckMDBKafkaTopicConfig(kfResource, "raw_events", &kafka.TopicConfig3{
						CleanupPolicy:   kafka.TopicConfig3_CLEANUP_POLICY_COMPACT_AND_DELETE,
						MaxMessageBytes: &wrappers.Int64Value{Value: 777216},
						SegmentBytes:    &wrappers.Int64Value{Value: 134217728},
						FlushMs:         &wrappers.Int64Value{Value: 9223372036854775807},
					}),
					testAccCheckCreatedAtAttr(kfResource),
				),
			},
			mdbKafkaClusterImportStep(kfResource),
			// Change some options
			{
				Config: testAccMDBKafkaClusterConfigUpdated(kfName, kfDescUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBKafkaClusterExists(kfResource, &r, 1),
					resource.TestCheckResourceAttr(kfResource, "name", kfName),
					resource.TestCheckResourceAttr(kfResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(kfResource, "description", kfDescUpdated),
					resource.TestCheckResourceAttr(kfResource, "config.0.access.0.data_transfer", "false"),
					resource.TestCheckResourceAttr(kfResource, "config.0.rest_api.0.enabled", "true"),
					resource.TestCheckResourceAttr(kfResource, "config.0.kafka_ui.0.enabled", "true"),
					resource.TestCheckResourceAttr(kfResource, "config.0.schema_registry", "true"),
					resource.TestCheckResourceAttr(kfResource, "config.0.version", currentDefaultKafkaVersion),
					testAccCheckMDBKafkaClusterContainsLabel(&r, "new_key", "new_value"),
					testAccCheckMDBKafkaClusterHasTopics(kfResource, []string{"raw_events", "new_topic"}),
					testAccCheckMDBKafkaClusterHasUsers(kfResource, map[string][]string{"alice": {"raw_events", "raw_events"}, "charlie": {"raw_events", "new_topic"}}),
					testAccCheckMDBKafkaClusterCompressionType(&r, kafka.CompressionType_COMPRESSION_TYPE_ZSTD),
					testAccCheckMDBKafkaClusterLogRetentionBytes(&r, 2147483648),
					testAccCheckMDBKafkaClusterLogSegmentBytes(&r, 268435456),
					testAccCheckMDBKafkaClusterSaslEnabledMechanisms(&r, []kafka.SaslMechanism{kafka.SaslMechanism_SASL_MECHANISM_SCRAM_SHA_256, kafka.SaslMechanism_SASL_MECHANISM_SCRAM_SHA_512}),
					testAccCheckMDBKafkaTopicConfig(kfResource, "raw_events", &kafka.TopicConfig3{
						CleanupPolicy:   kafka.TopicConfig3_CLEANUP_POLICY_DELETE,
						MaxMessageBytes: &wrappers.Int64Value{Value: 554432},
						SegmentBytes:    &wrappers.Int64Value{Value: 268435456},
						FlushMs:         &wrappers.Int64Value{Value: 9223372036854775807},
					}),
					testAccCheckCreatedAtAttr(kfResource),
				),
			},
		},
	})
}

// Test that a Kafka Cluster can be created, updated and destroyed in high availability mode
func TestAccMDBKafkaCluster_HA(t *testing.T) {
	t.Parallel()

	var r kafka.Cluster
	kfName := acctest.RandomWithPrefix("tf-kafka")
	kfDesc := "Kafka Cluster Terraform Test"
	kfDescUpdated := "Kafka Cluster Terraform Test (updated)"
	folderID := getExampleFolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMDBKafkaClusterDestroy,
		Steps: []resource.TestStep{
			// Create Kafka Cluster
			{
				Config: testAccMDBKafkaClusterConfigMainHA(kfName, kfDesc),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBKafkaClusterExists(kfResource, &r, 1),
					resource.TestCheckResourceAttr(kfResource, "name", kfName),
					resource.TestCheckResourceAttr(kfResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(kfResource, "description", kfDesc),
					resource.TestCheckResourceAttr(kfResource, "config.0.version", currentDefaultKafkaVersion),
					resource.TestCheckResourceAttr(kfResource, "config.0.rest_api.0.enabled", "true"),
					resource.TestCheckResourceAttr(kfResource, "config.0.kafka_ui.0.enabled", "true"),
					testAccCheckMDBKafkaClusterContainsLabel(&r, "test_key", "test_value"),
					testAccCheckMDBKafkaConfigKafkaHasResources(&r, "s2.micro", "network-hdd", 17179869184),
					testAccCheckMDBKafkaClusterHasTopics(kfResource, []string{"raw_events", "final"}),
					testAccCheckMDBKafkaClusterHasUsers(kfResource, map[string][]string{"alice": {"raw_events"}, "bob": {"raw_events", "final"}}),
					testAccCheckMDBKafkaConfigZones(&r, []string{"ru-central1-a", "ru-central1-b"}),
					testAccCheckMDBKafkaConfigBrokersCount(&r, 1),
					testAccCheckMDBKafkaClusterCompressionType(&r, kafka.CompressionType_COMPRESSION_TYPE_ZSTD),
					testAccCheckMDBKafkaClusterLogRetentionBytes(&r, 1073741824),
					testAccCheckMDBKafkaClusterSaslEnabledMechanisms(&r, []kafka.SaslMechanism{kafka.SaslMechanism_SASL_MECHANISM_SCRAM_SHA_256}),
					testAccCheckMDBKafkaTopicConfig(kfResource, "raw_events", &kafka.TopicConfig3{MaxMessageBytes: &wrappers.Int64Value{Value: 777216}, SegmentBytes: &wrappers.Int64Value{Value: 134217728}}),
					testAccCheckCreatedAtAttr(kfResource),
				),
			},
			mdbKafkaClusterImportStep(kfResource),
			// Change some options
			{
				Config: testAccMDBKafkaClusterConfigUpdatedHA(kfName, kfDescUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMDBKafkaClusterExists(kfResource, &r, 1),
					resource.TestCheckResourceAttr(kfResource, "name", kfName),
					resource.TestCheckResourceAttr(kfResource, "folder_id", folderID),
					resource.TestCheckResourceAttr(kfResource, "description", kfDescUpdated),
					resource.TestCheckResourceAttr(kfResource, "config.0.version", currentDefaultKafkaVersion),
					resource.TestCheckResourceAttr(kfResource, "config.0.rest_api.0.enabled", "true"),
					resource.TestCheckResourceAttr(kfResource, "config.0.kafka_ui.0.enabled", "true"),
					testAccCheckMDBKafkaClusterContainsLabel(&r, "new_key", "new_value"),
					testAccCheckMDBKafkaConfigZones(&r, []string{"ru-central1-a", "ru-central1-b", "ru-central1-d"}),
					testAccCheckMDBKafkaConfigBrokersCount(&r, 2),
					testAccCheckMDBKafkaClusterHasTopics(kfResource, []string{"raw_events", "new_topic"}),
					testAccCheckMDBKafkaClusterHasUsers(kfResource, map[string][]string{"alice": {"raw_events"}, "charlie": {"raw_events", "new_topic"}}),
					testAccCheckMDBKafkaClusterCompressionType(&r, kafka.CompressionType_COMPRESSION_TYPE_ZSTD),
					testAccCheckMDBKafkaClusterLogRetentionBytes(&r, 2147483648),
					testAccCheckMDBKafkaClusterLogSegmentBytes(&r, 268435456),
					testAccCheckMDBKafkaClusterSaslEnabledMechanisms(&r, nil),
					testAccCheckMDBKafkaTopicConfig(kfResource, "raw_events", &kafka.TopicConfig3{MaxMessageBytes: &wrappers.Int64Value{Value: 554432}, SegmentBytes: &wrappers.Int64Value{Value: 268435456}, RetentionBytes: &wrappers.Int64Value{Value: 1073741824}}),
					testAccCheckCreatedAtAttr(kfResource),
				),
			},
		},
	})
}

func testAccCheckMDBKafkaClusterExists(n string, r *kafka.Cluster, hosts int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.sdk.MDB().Kafka().Cluster().Get(context.Background(), &kafka.GetClusterRequest{
			ClusterId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Kafka Cluster not found")
		}

		*r = *found
		return nil
	}
}

func testAccCheckMDBKafkaClusterDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_mdb_kafka_cluster" {
			continue
		}

		_, err := config.sdk.MDB().Kafka().Cluster().Get(context.Background(), &kafka.GetClusterRequest{
			ClusterId: rs.Primary.ID,
		})

		if err == nil {
			return fmt.Errorf("kafka Cluster still exists")
		}
	}

	return nil
}

func testAccMDBKafkaClusterConfigMain(name, desc, environment string) string {
	return fmt.Sprintf(kfVPCDependencies+`
resource "yandex_mdb_kafka_cluster" "foo" {
	name        = "%s"
	description = "%s"
	environment = "%s"
	network_id  = yandex_vpc_network.mdb-kafka-test-net.id
	labels = {
	  test_key = "test_value"
	}
	subnet_ids = [yandex_vpc_subnet.mdb-kafka-test-subnet-a.id]
	deletion_protection = false

	config {
	  version          = "%s"
	  brokers_count    = 1
	  zones            = ["ru-central1-a"]
	  assign_public_ip = false
	  schema_registry  = false	
      access {
	    data_transfer  = true
      }
      rest_api {
	    enabled = false
      }
      kafka_ui {
	    enabled = false
      }
	  kafka {
		resources {
		  resource_preset_id = "s2.micro"
		  disk_type_id       = "network-hdd"
		  disk_size          = 16
		}

		kafka_config {
		  compression_type    		 = "COMPRESSION_TYPE_ZSTD"
		  log_retention_bytes 		 = 1073741824
		  sasl_enabled_mechanisms    = ["SASL_MECHANISM_SCRAM_SHA_256"]
		}
	  }
	}

	topic {
	  name               = "raw_events"
	  partitions         = 1
	  replication_factor = 1
	  topic_config {
		cleanup_policy    = "CLEANUP_POLICY_COMPACT_AND_DELETE"
		max_message_bytes = 777216
		segment_bytes     = 134217728
		flush_ms          = 9223372036854775807
	  }
	}

	topic {
	  name               = "final"
	  partitions         = 2
	  replication_factor = 1
	  topic_config {
		compression_type = "COMPRESSION_TYPE_ZSTD"
		segment_bytes    = 134217728
	  }
	}

	user {
	  name     = "alice"
	  password = "password"
	  permission {
		topic_name = "raw_events"
		role       = "ACCESS_ROLE_PRODUCER"
	  }
	}

	user {
	  name     = "bob"
	  password = "password"
	  permission {
		topic_name = "raw_events"
		role       = "ACCESS_ROLE_CONSUMER"
	  }
	  permission {
		topic_name = "final"
		role       = "ACCESS_ROLE_PRODUCER"
	  }
	}
}
`, name, desc, environment, currentDefaultKafkaVersion)
}

func testAccMDBKafkaClusterConfigUpdated(name, desc string) string {
	return fmt.Sprintf(kfVPCDependencies+`
resource "yandex_mdb_kafka_cluster" "foo" {
	name        = "%s"
	description = "%s"
	environment = "PRESTABLE"
	network_id  = yandex_vpc_network.mdb-kafka-test-net.id
	labels = {
		test_key = "test_value"
		new_key = "new_value"
	}
	subnet_ids = [yandex_vpc_subnet.mdb-kafka-test-subnet-a.id]

	config {
		version = "%s"
		brokers_count = 1
		zones = ["ru-central1-a"]
		assign_public_ip = false
		schema_registry  = true
        access {
	        data_transfer = false
        }
        rest_api {
	        enabled = true
        }
        kafka_ui {
	        enabled = true
        }
		kafka {
			resources {
				resource_preset_id = "s2.micro"
				disk_type_id       = "network-hdd"
                disk_size          = 16
			}
			kafka_config {
				compression_type    	   = "COMPRESSION_TYPE_ZSTD"
				log_retention_bytes 	   = 2147483648
				log_segment_bytes   	   = 268435456
           		sasl_enabled_mechanisms    = ["SASL_MECHANISM_SCRAM_SHA_512","SASL_MECHANISM_SCRAM_SHA_256"]
			}
		}
	}

	topic {
		name = "raw_events"
		partitions = 1
		replication_factor = 1

		topic_config {
			cleanup_policy = "CLEANUP_POLICY_DELETE"
	 		max_message_bytes = 554432
			segment_bytes = 268435456
			flush_ms      = 9223372036854775807
		}
	}

	topic {
		name = "new_topic"
		partitions = 1
		replication_factor = 1
	}

	user {
		name = "alice"
		password = "password"
		permission {
			topic_name = "raw_events"
			role = "ACCESS_ROLE_PRODUCER"
		}
		permission {
			topic_name = "raw_events"
			role = "ACCESS_ROLE_CONSUMER"
		}
	}

	user {
		name = "charlie"
		password = "password"
		permission {
			topic_name = "raw_events"
			role = "ACCESS_ROLE_CONSUMER"
		}
		permission {
			topic_name = "new_topic"
			role = "ACCESS_ROLE_PRODUCER"
		}
	}
}
`, name, desc, currentDefaultKafkaVersion)
}

func testAccCheckMDBKafkaClusterContainsLabel(r *kafka.Cluster, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		v, ok := r.Labels[key]
		if !ok {
			return fmt.Errorf("expected label with key '%s' not found", key)
		}
		if v != value {
			return fmt.Errorf("incorrect label value for key '%s': expected '%s' but found '%s'", key, value, v)
		}
		return nil
	}
}

func testAccCheckMDBKafkaClusterCompressionType(r *kafka.Cluster, value kafka.CompressionType) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		v := r.Config.Kafka.GetKafkaConfig_3().CompressionType
		if v != value {
			return fmt.Errorf("incorrect compression_type value: expected '%s' but found '%s'", value, v)
		}
		return nil
	}
}

func testAccCheckMDBKafkaClusterLogRetentionBytes(r *kafka.Cluster, value int64) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		v := r.Config.Kafka.GetKafkaConfig_3().LogRetentionBytes
		if v.GetValue() != value {
			return fmt.Errorf("incorrect log_retention_bytes value: expected '%v' but found '%v'", value, v.GetValue())
		}
		return nil
	}
}

func testAccCheckMDBKafkaClusterLogSegmentBytes(r *kafka.Cluster, value int64) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		v := r.Config.Kafka.GetKafkaConfig_3().LogSegmentBytes
		if v.GetValue() != value {
			return fmt.Errorf("incorrect log_segment_bytes value: expected '%v' but found '%v'", value, v.GetValue())
		}
		return nil
	}
}

func testAccCheckMDBKafkaClusterSaslEnabledMechanisms(r *kafka.Cluster, value []kafka.SaslMechanism) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		v := r.Config.Kafka.GetKafkaConfig_3().SaslEnabledMechanisms
		if !reflect.DeepEqual(v, value) {
			return fmt.Errorf("incorrect saslEnabledMechanisms value: expected '%v' but found '%v'", value, v)
		}
		return nil
	}
}

func testAccCheckMDBKafkaConfigKafkaHasResources(r *kafka.Cluster, resourcePresetID string, diskType string, diskSize int64) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := r.Config.Kafka.Resources
		if rs.ResourcePresetId != resourcePresetID {
			return fmt.Errorf("expected resource preset id '%s', got '%s'", resourcePresetID, rs.ResourcePresetId)
		}
		if rs.DiskTypeId != diskType {
			return fmt.Errorf("expected disk type '%s', got '%s'", diskType, rs.DiskTypeId)
		}
		if rs.DiskSize != diskSize {
			return fmt.Errorf("expected disk size '%d', got '%d'", diskSize, rs.DiskSize)
		}
		return nil
	}
}

func testAccCheckMDBKafkaConfigBrokersCount(r *kafka.Cluster, brokers int64) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if r.Config.BrokersCount.GetValue() != brokers {
			return fmt.Errorf("expected brokers '%v', got '%v'", brokers, r.Config.BrokersCount.GetValue())
		}
		return nil
	}
}

func testAccCheckMDBKafkaConfigZones(r *kafka.Cluster, zones []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !reflect.DeepEqual(r.Config.ZoneId, zones) {
			return fmt.Errorf("expected zones '%s', got '%s'", zones, r.Config.ZoneId)
		}
		return nil
	}
}

func testAccCheckMDBKafkaClusterHasTopics(r string, topics []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("not found: %s", r)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		resp, err := config.sdk.MDB().Kafka().Topic().List(context.Background(), &kafka.ListTopicsRequest{
			ClusterId: rs.Primary.ID,
			PageSize:  defaultMDBPageSize,
		})
		if err != nil {
			return err
		}
		tpcs := []string{}
		for _, d := range resp.Topics {
			tpcs = append(tpcs, d.Name)
		}

		if len(tpcs) != len(topics) {
			return fmt.Errorf("expected topics %v, found %v", topics, tpcs)
		}

		sort.Strings(tpcs)
		sort.Strings(topics)
		if fmt.Sprintf("%v", tpcs) != fmt.Sprintf("%v", topics) {
			return fmt.Errorf("cluster has wrong topics, %v. Expected %v", tpcs, topics)
		}

		return nil
	}
}

func testAccCheckMDBKafkaTopicMaxMessageBytes(r string, topic string, value int64) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("not found: %s", r)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		resp, err := config.sdk.MDB().Kafka().Topic().Get(context.Background(), &kafka.GetTopicRequest{
			ClusterId: rs.Primary.ID,
			TopicName: topic,
		})
		if err != nil {
			return err
		}
		v := resp.GetTopicConfig_3().MaxMessageBytes.GetValue()
		if v != value {
			return fmt.Errorf("MaxMessageByte for topic %v has value: %v, expected: %v", topic, v, value)
		}
		return nil
	}
}

func testAccCheckMDBKafkaTopicConfig(r string, topicName string, topicConfig *kafka.TopicConfig3) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("not found: %s", r)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		resp, err := config.sdk.MDB().Kafka().Topic().Get(context.Background(), &kafka.GetTopicRequest{
			ClusterId: rs.Primary.ID,
			TopicName: topicName,
		})
		if err != nil {
			return err
		}
		actualTopicConfig := resp.GetTopicConfig_3()
		if !reflect.DeepEqual(topicConfig, actualTopicConfig) {
			return fmt.Errorf("topic %v differs, actual: %v, expected %v", topicName, actualTopicConfig, topicConfig)
		}
		return nil
	}
}

func testAccCheckMDBKafkaClusterHasUsers(r string, perms map[string][]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("not found: %s", r)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		resp, err := config.sdk.MDB().Kafka().User().List(context.Background(), &kafka.ListUsersRequest{
			ClusterId: rs.Primary.ID,
			PageSize:  defaultMDBPageSize,
		})
		if err != nil {
			return err
		}
		users := resp.Users

		if len(users) != len(perms) {
			return fmt.Errorf("expected %d users, found %d", len(perms), len(users))
		}

		for _, u := range users {
			ps, ok := perms[u.Name]
			if !ok {
				return fmt.Errorf("unexpected user: %s", u.Name)
			}

			ups := []string{}
			for _, p := range u.Permissions {
				ups = append(ups, p.TopicName)
			}

			sort.Strings(ps)
			sort.Strings(ups)
			if fmt.Sprintf("%v", ps) != fmt.Sprintf("%v", ups) {
				return fmt.Errorf("user %s has wrong permissions, %v. Expected %v", u.Name, ups, ps)
			}
		}

		return nil
	}
}

func testAccMDBKafkaClusterConfigMainHA(name, desc string) string {
	return fmt.Sprintf(kfVPCDependencies+`
resource "yandex_mdb_kafka_cluster" "foo" {
	name        = "%s"
	description = "%s"
	environment = "PRODUCTION"
	network_id  = yandex_vpc_network.mdb-kafka-test-net.id
	labels = {
	  test_key = "test_value"
	}
	subnet_ids = [
	  yandex_vpc_subnet.mdb-kafka-test-subnet-a.id,
	  yandex_vpc_subnet.mdb-kafka-test-subnet-b.id,
	  yandex_vpc_subnet.mdb-kafka-test-subnet-d.id
	]

	config {
	  version          = "%s"
	  brokers_count    = 1
	  zones            = ["ru-central1-a", "ru-central1-b"]
	  assign_public_ip = false
	  schema_registry  = false
      rest_api {
	    enabled = true
      }
      kafka_ui {
	    enabled = true
      }
	  kafka {
		resources {
		  resource_preset_id = "s2.micro"
		  disk_type_id       = "network-hdd"
		  disk_size          = 16
		}
		kafka_config {
		  compression_type    		 = "COMPRESSION_TYPE_ZSTD"
		  log_retention_bytes 		 = 1073741824
		  sasl_enabled_mechanisms    = ["SASL_MECHANISM_SCRAM_SHA_256"]
		}
	  }
	}

	topic {
	  name               = "raw_events"
	  partitions         = 1
	  replication_factor = 1

	  topic_config {
		max_message_bytes = 777216
		segment_bytes     = 134217728
	  }
	}

	topic {
	  name               = "final"
	  partitions         = 2
	  replication_factor = 1
	}

	user {
	  name     = "alice"
	  password = "password"
	  permission {
		topic_name = "raw_events"
		role       = "ACCESS_ROLE_PRODUCER"
	  }
	}

	user {
	  name     = "bob"
	  password = "password"
	  permission {
		topic_name = "raw_events"
		role       = "ACCESS_ROLE_CONSUMER"
	  }
	  permission {
		topic_name = "final"
		role       = "ACCESS_ROLE_PRODUCER"
	  }
	}
}
`, name, desc, currentDefaultKafkaVersion)
}

func testAccMDBKafkaClusterConfigUpdatedHA(name, desc string) string {
	return fmt.Sprintf(kfVPCDependencies+`
resource "yandex_mdb_kafka_cluster" "foo" {
	name        = "%s"
	description = "%s"
	environment = "PRODUCTION"
	network_id  = yandex_vpc_network.mdb-kafka-test-net.id
	labels = {
	  test_key = "test_value"
	  new_key  = "new_value"
	}
	subnet_ids = [
	  yandex_vpc_subnet.mdb-kafka-test-subnet-a.id,
	  yandex_vpc_subnet.mdb-kafka-test-subnet-b.id,
	  yandex_vpc_subnet.mdb-kafka-test-subnet-d.id
	]

	config {
	  version          = "%s"
	  brokers_count    = 2
	  zones            = ["ru-central1-a", "ru-central1-b", "ru-central1-d"]
	  assign_public_ip = false
      schema_registry  = false
	  kafka {
		resources {
		  resource_preset_id = "s2.micro"
		  disk_type_id       = "network-hdd"
		  disk_size          = 16
		}
		kafka_config {
		  compression_type    		 = "COMPRESSION_TYPE_ZSTD"
		  log_retention_bytes 		 = 2147483648
		  log_segment_bytes   		 = 268435456
		}
	  }
	}

	topic {
	  name               = "raw_events"
	  partitions         = 2
	  replication_factor = 1
	  topic_config {
		max_message_bytes = 554432
		segment_bytes     = 268435456
		retention_bytes   = 1073741824
	  }
	}

	topic {
	  name               = "new_topic"
	  partitions         = 1
	  replication_factor = 1
	}

	user {
	  name     = "alice"
	  password = "password"
	  permission {
		topic_name = "raw_events"
		role       = "ACCESS_ROLE_PRODUCER"
	  }
	}

	user {
	  name     = "charlie"
	  password = "password"
	  permission {
		topic_name = "raw_events"
		role       = "ACCESS_ROLE_CONSUMER"
	  }
	  permission {
		topic_name = "new_topic"
		role       = "ACCESS_ROLE_PRODUCER"
	  }
	}
}
`, name, desc, currentDefaultKafkaVersion)
}
