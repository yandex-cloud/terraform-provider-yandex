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

func TestBuildKafkaConnectorSpecIcebergSink(t *testing.T) {
	raw := map[string]interface{}{
		"cluster_id": "cid1",
		"name":       "connector1",
		"tasks_max":  3,
		"properties": map[string]interface{}{
			"prop1": "prop_val1",
			"prop2": "prop_val2",
		},
		"connector_config_iceberg_sink": []interface{}{
			map[string]interface{}{
				"topics":        "topics_*",
				"control_topic": "control-topic",
				"metastore_connection": []interface{}{
					map[string]interface{}{
						"catalog_uri": "thrift://metastore:9083",
						"warehouse":   "s3a://bucket/warehouse",
					},
				},
				"s3_connection": []interface{}{
					map[string]interface{}{
						"external_s3": []interface{}{
							map[string]interface{}{
								"endpoint":          "storage.yandexcloud.net",
								"access_key_id":     "access_key_id",
								"secret_access_key": "secret_access_key",
								"region":            "ru-central1",
							},
						},
					},
				},
				"static_tables": []interface{}{
					map[string]interface{}{
						"tables": "db.table1,db.table2",
					},
				},
				"tables_config": []interface{}{
					map[string]interface{}{
						"default_commit_branch":   "main",
						"default_id_columns":      "id,timestamp",
						"default_partition_by":    "year(timestamp)",
						"evolve_schema_enabled":   true,
						"schema_force_optional":   true,
						"schema_case_insensitive": true,
					},
				},
				"control_config": []interface{}{
					map[string]interface{}{
						"group_id_prefix":      "cg-control",
						"commit_interval_ms":   300000,
						"commit_timeout_ms":    30000,
						"commit_threads":       4,
						"transactional_prefix": "tx-",
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
	expected.SetConnectorConfigIcebergSink(&kafka.ConnectorConfigIcebergSinkSpec{
		TopicsSource: &kafka.ConnectorConfigIcebergSinkSpec_Topics{
			Topics: "topics_*",
		},
		ControlTopic: "control-topic",
		MetastoreConnection: &kafka.MetastoreConnectionSpec{
			CatalogUri: "thrift://metastore:9083",
			Warehouse:  "s3a://bucket/warehouse",
		},
		S3Connection: &kafka.IcebergS3ConnectionSpec{
			Storage: &kafka.IcebergS3ConnectionSpec_ExternalS3{
				ExternalS3: &kafka.ExternalIcebergS3StorageSpec{
					AccessKeyId:     "access_key_id",
					SecretAccessKey: "secret_access_key",
					Endpoint:        "storage.yandexcloud.net",
					Region:          "ru-central1",
				},
			},
		},
		TableRouting: &kafka.ConnectorConfigIcebergSinkSpec_StaticTables{
			StaticTables: &kafka.StaticTablesSpec{
				Tables: "db.table1,db.table2",
			},
		},
		TablesConfig: &kafka.IcebergTablesConfigSpec{
			DefaultCommitBranch:   "main",
			DefaultIdColumns:      "id,timestamp",
			DefaultPartitionBy:    "year(timestamp)",
			EvolveSchemaEnabled:   true,
			SchemaForceOptional:   true,
			SchemaCaseInsensitive: true,
		},
		ControlConfig: &kafka.IcebergControlSpec{
			GroupIdPrefix:       "cg-control",
			CommitIntervalMs:    &wrappers.Int64Value{Value: int64(300000)},
			CommitTimeoutMs:     &wrappers.Int64Value{Value: int64(30000)},
			CommitThreads:       &wrappers.Int64Value{Value: int64(4)},
			TransactionalPrefix: "tx-",
		},
	})
	assert.Equal(t, expected, connSpec)
}

func TestBuildKafkaConnectorSpecIcebergSinkWithTopicsRegex(t *testing.T) {
	raw := map[string]interface{}{
		"cluster_id": "cid1",
		"name":       "connector1",
		"tasks_max":  3,
		"properties": map[string]interface{}{},
		"connector_config_iceberg_sink": []interface{}{
			map[string]interface{}{
				"topics_regex": "topic-.*",
				"metastore_connection": []interface{}{
					map[string]interface{}{
						"catalog_uri": "thrift://metastore:9083",
						"warehouse":   "s3a://bucket/warehouse",
					},
				},
				"s3_connection": []interface{}{
					map[string]interface{}{
						"external_s3": []interface{}{
							map[string]interface{}{
								"endpoint":          "storage.yandexcloud.net",
								"access_key_id":     "access_key_id",
								"secret_access_key": "secret_access_key",
								"region":            "ru-central1",
							},
						},
					},
				},
				"dynamic_tables": []interface{}{
					map[string]interface{}{
						"route_field": "table_name",
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
		Properties: map[string]string{},
	}
	expected.SetConnectorConfigIcebergSink(&kafka.ConnectorConfigIcebergSinkSpec{
		TopicsSource: &kafka.ConnectorConfigIcebergSinkSpec_TopicsRegex{
			TopicsRegex: "topic-.*",
		},
		MetastoreConnection: &kafka.MetastoreConnectionSpec{
			CatalogUri: "thrift://metastore:9083",
			Warehouse:  "s3a://bucket/warehouse",
		},
		S3Connection: &kafka.IcebergS3ConnectionSpec{
			Storage: &kafka.IcebergS3ConnectionSpec_ExternalS3{
				ExternalS3: &kafka.ExternalIcebergS3StorageSpec{
					AccessKeyId:     "access_key_id",
					SecretAccessKey: "secret_access_key",
					Endpoint:        "storage.yandexcloud.net",
					Region:          "ru-central1",
				},
			},
		},
		TableRouting: &kafka.ConnectorConfigIcebergSinkSpec_DynamicTables{
			DynamicTables: &kafka.DynamicTablesSpec{
				RouteField: "table_name",
			},
		},
	})
	assert.Equal(t, expected, connSpec)
}

func TestBuildKafkaConnectorSpecIcebergSinkUpdate(t *testing.T) {
	raw := map[string]interface{}{
		"cluster_id": "cid1",
		"name":       "connector1",
		"tasks_max":  5,
		"properties": map[string]interface{}{
			"prop1": "prop_val1",
		},
		"connector_config_iceberg_sink": []interface{}{
			map[string]interface{}{
				"topics":        "updated_topics_*",
				"control_topic": "updated-control-topic",
				"metastore_connection": []interface{}{
					map[string]interface{}{
						"catalog_uri": "thrift://new-metastore:9083",
						"warehouse":   "s3a://new-bucket/warehouse",
					},
				},
				"s3_connection": []interface{}{
					map[string]interface{}{
						"external_s3": []interface{}{
							map[string]interface{}{
								"endpoint":          "new-storage.yandexcloud.net",
								"access_key_id":     "new_access_key_id",
								"secret_access_key": "new_secret_access_key",
								"region":            "ru-central1-b",
							},
						},
					},
				},
				"tables_config": []interface{}{
					map[string]interface{}{
						"default_commit_branch":   "develop",
						"default_id_columns":      "id",
						"default_partition_by":    "month(timestamp)",
						"evolve_schema_enabled":   false,
						"schema_force_optional":   false,
						"schema_case_insensitive": false,
					},
				},
				"control_config": []interface{}{
					map[string]interface{}{
						"group_id_prefix":      "new-cg-control",
						"commit_interval_ms":   600000,
						"commit_timeout_ms":    60000,
						"commit_threads":       8,
						"transactional_prefix": "new-tx-",
					},
				},
			},
		},
	}
	resourceData := schema.TestResourceDataRaw(t, resourceYandexMDBKafkaConnector().Schema, raw)
	connSpec, err := buildKafkaConnectorUpdateSpec(resourceData)
	require.NoError(t, err)

	expected := &kafka.UpdateConnectorSpec{
		TasksMax:   &wrappers.Int64Value{Value: int64(5)},
		Properties: map[string]string{"prop1": "prop_val1"},
	}
	expected.SetConnectorConfigIcebergSink(&kafka.UpdateConnectorConfigIcebergSinkSpec{
		TopicsSource: &kafka.UpdateConnectorConfigIcebergSinkSpec_Topics{
			Topics: "updated_topics_*",
		},
		ControlTopic: "updated-control-topic",
		MetastoreConnection: &kafka.MetastoreConnectionSpec{
			CatalogUri: "thrift://new-metastore:9083",
			Warehouse:  "s3a://new-bucket/warehouse",
		},
		S3Connection: &kafka.IcebergS3ConnectionSpec{
			Storage: &kafka.IcebergS3ConnectionSpec_ExternalS3{
				ExternalS3: &kafka.ExternalIcebergS3StorageSpec{
					AccessKeyId:     "new_access_key_id",
					SecretAccessKey: "new_secret_access_key",
					Endpoint:        "new-storage.yandexcloud.net",
					Region:          "ru-central1-b",
				},
			},
		},
		TablesConfig: &kafka.IcebergTablesConfigSpec{
			DefaultCommitBranch:   "develop",
			DefaultIdColumns:      "id",
			DefaultPartitionBy:    "month(timestamp)",
			EvolveSchemaEnabled:   false,
			SchemaForceOptional:   false,
			SchemaCaseInsensitive: false,
		},
		ControlConfig: &kafka.IcebergControlSpec{
			GroupIdPrefix:       "new-cg-control",
			CommitIntervalMs:    &wrappers.Int64Value{Value: int64(600000)},
			CommitTimeoutMs:     &wrappers.Int64Value{Value: int64(60000)},
			CommitThreads:       &wrappers.Int64Value{Value: int64(8)},
			TransactionalPrefix: "new-tx-",
		},
	})
	assert.Equal(t, expected, connSpec)
}

func TestBuildKafkaConnectorSpecIcebergSinkMinimal(t *testing.T) {
	raw := map[string]interface{}{
		"cluster_id": "cid1",
		"name":       "connector1",
		"connector_config_iceberg_sink": []interface{}{
			map[string]interface{}{
				"topics": "topics_*",
				"metastore_connection": []interface{}{
					map[string]interface{}{
						"catalog_uri": "thrift://metastore:9083",
						"warehouse":   "s3a://bucket/warehouse",
					},
				},
				"s3_connection": []interface{}{
					map[string]interface{}{
						"external_s3": []interface{}{
							map[string]interface{}{
								"endpoint":          "storage.yandexcloud.net",
								"access_key_id":     "access_key_id",
								"secret_access_key": "secret_access_key",
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
		Properties: map[string]string{},
	}
	expected.SetConnectorConfigIcebergSink(&kafka.ConnectorConfigIcebergSinkSpec{
		TopicsSource: &kafka.ConnectorConfigIcebergSinkSpec_Topics{
			Topics: "topics_*",
		},
		MetastoreConnection: &kafka.MetastoreConnectionSpec{
			CatalogUri: "thrift://metastore:9083",
			Warehouse:  "s3a://bucket/warehouse",
		},
		S3Connection: &kafka.IcebergS3ConnectionSpec{
			Storage: &kafka.IcebergS3ConnectionSpec_ExternalS3{
				ExternalS3: &kafka.ExternalIcebergS3StorageSpec{
					AccessKeyId:     "access_key_id",
					SecretAccessKey: "secret_access_key",
					Endpoint:        "storage.yandexcloud.net",
				},
			},
		},
	})
	assert.Equal(t, expected, connSpec)
}
