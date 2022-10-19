package yandex

import (
	"testing"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/kafka/v1"
)

func TestBuildKafkaConnectorSpecMirrormaker(t *testing.T) {
	raw := map[string]interface{}{
		"cluster_id": "cid1",
		"name":       "connector1",
		"tasks_max":  3,
		"properties": map[string]interface{}{
			"prop1": "prop_val1",
			"prop2": "prop_val2",
		},
		"connector_config_mirrormaker": []interface{}{
			map[string]interface{}{
				"topics":             "topics_*",
				"replication_factor": 3,
				"source_cluster": []interface{}{
					map[string]interface{}{
						"alias": "source",
						"external_cluster": []interface{}{
							map[string]interface{}{
								"bootstrap_servers": "bootsrtap_servers",
								"sasl_username":     "sasl_username",
								"sasl_password":     "sasl_password",
								"sasl_mechanism":    "sasl_mechanism",
								"security_protocol": "security_protocol",
							},
						},
					},
				},
				"target_cluster": []interface{}{
					map[string]interface{}{
						"alias": "target",
						"this_cluster": []interface{}{
							map[string]interface{}{},
						},
					},
				},
			},
		},
	}
	resourceData := schema.TestResourceDataRaw(t, resourceYandexMDBKafkaConnector().Schema, raw)
	connSpec, err := buildKafkaConnectorSpec(resourceData)
	require.NoError(t, err)

	expected := &kafka.ConnectorSpec{
		Name:       "connector1",
		TasksMax:   &wrappers.Int64Value{Value: int64(3)},
		Properties: map[string]string{"prop1": "prop_val1", "prop2": "prop_val2"},
	}
	expected.SetConnectorConfigMirrormaker(&kafka.ConnectorConfigMirrorMakerSpec{
		Topics:            "topics_*",
		ReplicationFactor: &wrappers.Int64Value{Value: int64(3)},
		SourceCluster: &kafka.ClusterConnectionSpec{
			Alias: "source",
			ClusterConnection: &kafka.ClusterConnectionSpec_ExternalCluster{
				ExternalCluster: &kafka.ExternalClusterConnectionSpec{
					BootstrapServers: "bootsrtap_servers",
					SaslUsername:     "sasl_username",
					SaslPassword:     "sasl_password",
					SaslMechanism:    "sasl_mechanism",
					SecurityProtocol: "security_protocol",
				},
			},
		},
		TargetCluster: &kafka.ClusterConnectionSpec{
			Alias:             "target",
			ClusterConnection: &kafka.ClusterConnectionSpec_ThisCluster{},
		},
	})
	assert.Equal(t, expected, connSpec)
}

func TestBuildKafkaConnectorSpecS3Sink(t *testing.T) {
	raw := map[string]interface{}{
		"cluster_id": "cid1",
		"name":       "connector1",
		"tasks_max":  3,
		"properties": map[string]interface{}{
			"prop1": "prop_val1",
			"prop2": "prop_val2",
		},
		"connector_config_s3_sink": []interface{}{
			map[string]interface{}{
				"topics":                "topics_*",
				"file_compression_type": "gzip",
				"file_max_records":      10,
				"s3_connection": []interface{}{
					map[string]interface{}{
						"bucket_name": "bucket1",
						"external_s3": []interface{}{
							map[string]interface{}{
								"endpoint":          "endpoint",
								"access_key_id":     "access_key_id",
								"secret_access_key": "secret_access_key",
								"region":            "region",
							},
						},
					},
				},
			},
		},
	}
	resourceData := schema.TestResourceDataRaw(t, resourceYandexMDBKafkaConnector().Schema, raw)
	connSpec, err := buildKafkaConnectorSpec(resourceData)
	require.NoError(t, err)

	expected := &kafka.ConnectorSpec{
		Name:       "connector1",
		TasksMax:   &wrappers.Int64Value{Value: int64(3)},
		Properties: map[string]string{"prop1": "prop_val1", "prop2": "prop_val2"},
	}
	expected.SetConnectorConfigS3Sink(&kafka.ConnectorConfigS3SinkSpec{
		Topics:              "topics_*",
		FileCompressionType: "gzip",
		FileMaxRecords:      &wrappers.Int64Value{Value: int64(10)},
		S3Connection: &kafka.S3ConnectionSpec{
			BucketName: "bucket1",
			Storage: &kafka.S3ConnectionSpec_ExternalS3{
				ExternalS3: &kafka.ExternalS3StorageSpec{
					AccessKeyId:     "access_key_id",
					SecretAccessKey: "secret_access_key",
					Endpoint:        "endpoint",
					Region:          "region",
				},
			},
		},
	})
	assert.Equal(t, expected, connSpec)
}

func TestBuildKafkaConnectorSpecMirrormakerUpdate(t *testing.T) {
	raw := map[string]interface{}{
		"cluster_id": "cid1",
		"name":       "connector1",
		"tasks_max":  3,
		"properties": map[string]interface{}{
			"prop1": "prop_val1",
			"prop2": "prop_val2",
		},
		"connector_config_mirrormaker": []interface{}{
			map[string]interface{}{
				"topics":             "topics_*",
				"replication_factor": 3,
				"source_cluster": []interface{}{
					map[string]interface{}{
						"alias": "source",
						"external_cluster": []interface{}{
							map[string]interface{}{
								"bootstrap_servers": "bootsrtap_servers",
								"sasl_username":     "sasl_username",
								"sasl_password":     "sasl_password",
								"sasl_mechanism":    "sasl_mechanism",
								"security_protocol": "security_protocol",
							},
						},
					},
				},
				"target_cluster": []interface{}{
					map[string]interface{}{
						"alias": "target",
						"this_cluster": []interface{}{
							map[string]interface{}{},
						},
					},
				},
			},
		},
	}
	resourceData := schema.TestResourceDataRaw(t, resourceYandexMDBKafkaConnector().Schema, raw)
	connSpec, err := buildKafkaConnectorUpdateSpec(resourceData)
	require.NoError(t, err)

	expected := &kafka.UpdateConnectorSpec{
		TasksMax:   &wrappers.Int64Value{Value: int64(3)},
		Properties: map[string]string{"prop1": "prop_val1", "prop2": "prop_val2"},
	}
	expected.SetConnectorConfigMirrormaker(&kafka.ConnectorConfigMirrorMakerSpec{
		Topics:            "topics_*",
		ReplicationFactor: &wrappers.Int64Value{Value: int64(3)},
		SourceCluster: &kafka.ClusterConnectionSpec{
			Alias: "source",
			ClusterConnection: &kafka.ClusterConnectionSpec_ExternalCluster{
				ExternalCluster: &kafka.ExternalClusterConnectionSpec{
					BootstrapServers: "bootsrtap_servers",
					SaslUsername:     "sasl_username",
					SaslPassword:     "sasl_password",
					SaslMechanism:    "sasl_mechanism",
					SecurityProtocol: "security_protocol",
				},
			},
		},
		TargetCluster: &kafka.ClusterConnectionSpec{
			Alias:             "target",
			ClusterConnection: &kafka.ClusterConnectionSpec_ThisCluster{},
		},
	})
	assert.Equal(t, expected, connSpec)
}

func TestBuildKafkaConnectorSpecS3SinkUpdate(t *testing.T) {
	raw := map[string]interface{}{
		"cluster_id": "cid1",
		"name":       "connector1",
		"tasks_max":  3,
		"properties": map[string]interface{}{
			"prop1": "prop_val1",
			"prop2": "prop_val2",
		},
		"connector_config_s3_sink": []interface{}{
			map[string]interface{}{
				"topics":                "topics_*",
				"file_compression_type": "gzip",
				"file_max_records":      10,
				"s3_connection": []interface{}{
					map[string]interface{}{
						"bucket_name": "bucket1",
						"external_s3": []interface{}{
							map[string]interface{}{
								"endpoint":          "endpoint",
								"access_key_id":     "access_key_id",
								"secret_access_key": "secret_access_key",
								"region":            "region",
							},
						},
					},
				},
			},
		},
	}
	resourceData := schema.TestResourceDataRaw(t, resourceYandexMDBKafkaConnector().Schema, raw)
	connSpec, err := buildKafkaConnectorUpdateSpec(resourceData)
	require.NoError(t, err)

	expected := &kafka.UpdateConnectorSpec{
		TasksMax:   &wrappers.Int64Value{Value: int64(3)},
		Properties: map[string]string{"prop1": "prop_val1", "prop2": "prop_val2"},
	}
	expected.SetConnectorConfigS3Sink(&kafka.UpdateConnectorConfigS3SinkSpec{
		Topics:         "topics_*",
		FileMaxRecords: &wrappers.Int64Value{Value: int64(10)},
		S3Connection: &kafka.S3ConnectionSpec{
			BucketName: "bucket1",
			Storage: &kafka.S3ConnectionSpec_ExternalS3{
				ExternalS3: &kafka.ExternalS3StorageSpec{
					AccessKeyId:     "access_key_id",
					SecretAccessKey: "secret_access_key",
					Endpoint:        "endpoint",
					Region:          "region",
				},
			},
		},
	})
	assert.Equal(t, expected, connSpec)
}

func TestBuildKafkaConnectorSpecWhenBothConfigsThenError(t *testing.T) {
	raw := map[string]interface{}{
		"cluster_id": "cid1",
		"name":       "connector1",
		"tasks_max":  3,
		"properties": map[string]interface{}{
			"prop1": "prop_val1",
			"prop2": "prop_val2",
		},
		"connector_config_mirrormaker": []interface{}{
			map[string]interface{}{
				"topics":             "topics_*",
				"replication_factor": 3,
				"source_cluster": []interface{}{
					map[string]interface{}{
						"alias": "source",
						"external_cluster": []interface{}{
							map[string]interface{}{
								"bootstrap_servers": "bootsrtap_servers",
								"sasl_username":     "sasl_username",
								"sasl_password":     "sasl_password",
								"sasl_mechanism":    "sasl_mechanism",
								"security_protocol": "security_protocol",
							},
						},
					},
				},
				"target_cluster": []interface{}{
					map[string]interface{}{
						"alias": "target",
						"this_cluster": []interface{}{
							map[string]interface{}{},
						},
					},
				},
			},
		},
		"connector_config_s3_sink": []interface{}{
			map[string]interface{}{
				"topics":                "topics_*",
				"file_compression_type": "gzip",
				"file_max_records":      10,
				"s3_connection": []interface{}{
					map[string]interface{}{
						"bucket_name": "bucket1",
						"external_s3": []interface{}{
							map[string]interface{}{
								"endpoint":          "endpoint",
								"access_key_id":     "access_key_id",
								"secret_access_key": "secret_access_key",
								"region":            "region",
							},
						},
					},
				},
			},
		},
	}
	resourceData := schema.TestResourceDataRaw(t, resourceYandexMDBKafkaConnector().Schema, raw)
	_, err := buildKafkaConnectorSpec(resourceData)
	require.Error(t, err)
	require.Equal(t, "must be specified only one connector-specific config", err.Error())
}

func TestBuildKafkaConnectorSpecUpdateWhenBothConfigsThenError(t *testing.T) {
	raw := map[string]interface{}{
		"cluster_id": "cid1",
		"name":       "connector1",
		"tasks_max":  3,
		"properties": map[string]interface{}{
			"prop1": "prop_val1",
			"prop2": "prop_val2",
		},
		"connector_config_mirrormaker": []interface{}{
			map[string]interface{}{
				"topics":             "topics_*",
				"replication_factor": 3,
				"source_cluster": []interface{}{
					map[string]interface{}{
						"alias": "source",
						"external_cluster": []interface{}{
							map[string]interface{}{
								"bootstrap_servers": "bootsrtap_servers",
								"sasl_username":     "sasl_username",
								"sasl_password":     "sasl_password",
								"sasl_mechanism":    "sasl_mechanism",
								"security_protocol": "security_protocol",
							},
						},
					},
				},
				"target_cluster": []interface{}{
					map[string]interface{}{
						"alias": "target",
						"this_cluster": []interface{}{
							map[string]interface{}{},
						},
					},
				},
			},
		},
		"connector_config_s3_sink": []interface{}{
			map[string]interface{}{
				"topics":                "topics_*",
				"file_compression_type": "gzip",
				"file_max_records":      10,
				"s3_connection": []interface{}{
					map[string]interface{}{
						"bucket_name": "bucket1",
						"external_s3": []interface{}{
							map[string]interface{}{
								"endpoint":          "endpoint",
								"access_key_id":     "access_key_id",
								"secret_access_key": "secret_access_key",
								"region":            "region",
							},
						},
					},
				},
			},
		},
	}
	resourceData := schema.TestResourceDataRaw(t, resourceYandexMDBKafkaConnector().Schema, raw)
	_, err := buildKafkaConnectorUpdateSpec(resourceData)
	require.Error(t, err)
	require.Equal(t, "must be specified only one connector-specific config", err.Error())
}

func TestBuildKafkaConnectorSpecWhenNoConfigsThenError(t *testing.T) {
	raw := map[string]interface{}{
		"cluster_id": "cid1",
		"name":       "connector1",
		"tasks_max":  3,
		"properties": map[string]interface{}{
			"prop1": "prop_val1",
			"prop2": "prop_val2",
		},
	}
	resourceData := schema.TestResourceDataRaw(t, resourceYandexMDBKafkaConnector().Schema, raw)
	_, err := buildKafkaConnectorSpec(resourceData)
	require.Error(t, err)
	require.Equal(t, "connector-specific config must be specified", err.Error())
}
