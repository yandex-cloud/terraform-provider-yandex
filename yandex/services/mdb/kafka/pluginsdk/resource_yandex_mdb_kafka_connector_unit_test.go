package pluginsdk_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/kafka/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
	svc_kafka "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/kafka/pluginsdk"
)

func TestKafkaConnectorMirrorCheckpoint(t *testing.T) {
	for _, tt := range []struct {
		name     string
		policy   kafka.ConnectorConfigMirrorCheckpoint_ReplicationPolicy
		location kafka.ConnectorConfigMirrorCheckpoint_OffsetSyncsTopicLocation
	}{
		{name: "inherited"},
		{name: "identity target", policy: kafka.ConnectorConfigMirrorCheckpoint_IDENTITY, location: kafka.ConnectorConfigMirrorCheckpoint_TARGET},
		{name: "default source", policy: kafka.ConnectorConfigMirrorCheckpoint_DEFAULT, location: kafka.ConnectorConfigMirrorCheckpoint_SOURCE},
	} {
		t.Run(tt.name, func(t *testing.T) {
			config := map[string]any{
				"source_connector_name": "mirror-source",
				"groups":                []any{"orders.reader", "billing-*"},
			}
			if tt.policy != 0 {
				config["replication_policy"] = tt.policy.String()
				config["offset_syncs_topic_location"] = tt.location.String()
			}
			raw := kafkaConnectorMirrorCheckpointRaw(config)
			raw["tasks_max"] = 1
			raw["properties"] = map[string]any{"sync.group.offsets.enabled": "false"}
			resource := svc_kafka.ResourceYandexMDBKafkaConnector()
			data := schema.TestResourceDataRaw(t, resource.Schema, raw)
			created, err := svc_kafka.BuildKafkaConnectorSpec(data)
			require.NoError(t, err)
			checkpoint := created.GetConnectorConfigMirrorCheckpoint()
			require.NotNil(t, checkpoint)
			assert.Equal(t, "mirror-source", checkpoint.SourceConnectorName)
			assert.ElementsMatch(t, []string{"orders.reader", "billing-*"}, checkpoint.Groups)
			assert.Equal(t, tt.policy, checkpoint.ReplicationPolicy)
			assert.Equal(t, tt.location, checkpoint.OffsetSyncsTopicLocation)
			assert.EqualValues(t, 1, created.TasksMax.Value)
			assert.Equal(t, "false", created.Properties["sync.group.offsets.enabled"])

			updated, err := svc_kafka.BuildKafkaConnectorUpdateSpec(data)
			require.NoError(t, err)
			assert.Equal(t, &kafka.UpdateConnectorConfigMirrorCheckpointSpec{
				Groups:                   checkpoint.Groups,
				ReplicationPolicy:        tt.policy,
				OffsetSyncsTopicLocation: tt.location,
			}, updated.GetConnectorConfigMirrorCheckpoint())

			response := &kafka.ConnectorConfigMirrorCheckpoint{
				SourceConnectorName:      checkpoint.SourceConnectorName,
				Groups:                   []string{"billing-*", "orders.reader"},
				ReplicationPolicy:        tt.policy,
				OffsetSyncsTopicLocation: tt.location,
			}
			data.SetId("cluster-id:checkpoint")
			require.NoError(t, data.Set("connector_config_mirror_checkpoint", svc_kafka.FlattenKafkaConnectorMirrorCheckpoint(response)))
			diff, err := resource.Diff(context.Background(), data.State(), terraform.NewResourceConfigRaw(raw), nil)
			require.NoError(t, err)
			assert.True(t, diff.Empty(), "%+v", diff)
		})
	}
}

func TestKafkaConnectorMirrorCheckpointUpdateDiff(t *testing.T) {
	resource := svc_kafka.ResourceYandexMDBKafkaConnector()
	initial := kafkaConnectorMirrorCheckpointRaw(map[string]any{
		"source_connector_name":       "mirror-source",
		"groups":                      []any{"old-group"},
		"replication_policy":          "IDENTITY",
		"offset_syncs_topic_location": "TARGET",
	})
	data := schema.TestResourceDataRaw(t, resource.Schema, initial)
	data.SetId("cluster-id:checkpoint")
	updated := kafkaConnectorMirrorCheckpointRaw(map[string]any{
		"source_connector_name": "mirror-source",
		"groups":                []any{"new-group"},
	})
	diff, err := resource.Diff(context.Background(), data.State(), terraform.NewResourceConfigRaw(updated), nil)
	require.NoError(t, err)
	require.False(t, diff.RequiresNew())
	updatedData, err := schema.InternalMap(resource.Schema).Data(data.State(), diff)
	require.NoError(t, err)
	var paths []string
	for field, path := range svc_kafka.MdbKafkaConnectorUpdateFieldsMap {
		if updatedData.HasChange(field) {
			paths = append(paths, path)
		}
	}
	assert.ElementsMatch(t, []string{
		"connector_spec.connector_config_mirror_checkpoint.groups",
		"connector_spec.connector_config_mirror_checkpoint.replication_policy",
		"connector_spec.connector_config_mirror_checkpoint.offset_syncs_topic_location",
	}, paths)
	spec, err := svc_kafka.BuildKafkaConnectorUpdateSpec(updatedData)
	require.NoError(t, err)
	assert.Equal(t, &kafka.UpdateConnectorConfigMirrorCheckpointSpec{Groups: []string{"new-group"}}, spec.GetConnectorConfigMirrorCheckpoint())

	updated["connector_config_mirror_checkpoint"].([]any)[0].(map[string]any)["source_connector_name"] = "another-source"
	diff, err = resource.Diff(context.Background(), data.State(), terraform.NewResourceConfigRaw(updated), nil)
	require.NoError(t, err)
	assert.True(t, diff.RequiresNew())
	assert.NotContains(t, svc_kafka.MdbKafkaConnectorUpdateFieldsMap, "connector_config_mirror_checkpoint.0.source_connector_name")

	mirrorMaker := kafkaConnectorMirrorMakerRaw(nil, nil)
	mirrorMaker["name"] = "checkpoint"
	diff, err = resource.Diff(context.Background(), data.State(), terraform.NewResourceConfigRaw(mirrorMaker), nil)
	require.NoError(t, err)
	assert.True(t, diff.RequiresNew())
	mirrorMakerData := schema.TestResourceDataRaw(t, resource.Schema, mirrorMaker)
	mirrorMakerData.SetId("cluster-id:checkpoint")
	diff, err = resource.Diff(context.Background(), mirrorMakerData.State(), terraform.NewResourceConfigRaw(initial), nil)
	require.NoError(t, err)
	assert.True(t, diff.RequiresNew())
}

func TestKafkaConnectorMirrorCheckpointValidation(t *testing.T) {
	resource := svc_kafka.ResourceYandexMDBKafkaConnector()
	for _, tt := range []struct {
		name  string
		field string
		value any
	}{
		{name: "empty source", field: "source_connector_name", value: ""},
		{name: "no groups", field: "groups", value: []any{}},
		{name: "empty group", field: "groups", value: []any{""}},
		{name: "invalid policy", field: "replication_policy", value: "CUSTOM"},
		{name: "invalid location", field: "offset_syncs_topic_location", value: "REMOTE"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			config := map[string]any{"source_connector_name": "source", "groups": []any{"group"}}
			config[tt.field] = tt.value
			diagnostics := resource.Validate(terraform.NewResourceConfigRaw(kafkaConnectorMirrorCheckpointRaw(config)))
			assert.True(t, diagnostics.HasError(), "%+v", diagnostics)
		})
	}
	raw := kafkaConnectorMirrorMakerRaw(nil, nil)
	raw["connector_config_mirror_checkpoint"] = []any{map[string]any{"source_connector_name": "source", "groups": []any{"group"}}}
	data := schema.TestResourceDataRaw(t, resource.Schema, raw)
	_, err := svc_kafka.BuildKafkaConnectorSpec(data)
	assert.ErrorContains(t, err, "only one connector-specific config")
	_, err = svc_kafka.BuildKafkaConnectorUpdateSpec(data)
	assert.ErrorContains(t, err, "only one connector-specific config")
}

func TestKafkaConnectorMirrorCheckpointDataSource(t *testing.T) {
	dataSource := svc_kafka.DataSourceYandexMDBKafkaConnector()
	checkpoint := dataSource.Schema["connector_config_mirror_checkpoint"]
	assert.True(t, checkpoint.Computed)
	for _, field := range checkpoint.Elem.(*schema.Resource).Schema {
		assert.True(t, field.Computed)
		assert.False(t, field.Required)
	}
	data := schema.TestResourceDataRaw(t, dataSource.Schema, map[string]any{"cluster_id": "cluster-id", "name": "checkpoint"})
	require.NoError(t, data.Set("connector_config_mirror_checkpoint", svc_kafka.FlattenKafkaConnectorMirrorCheckpoint(&kafka.ConnectorConfigMirrorCheckpoint{
		SourceConnectorName:      "mirror-source",
		Groups:                   []string{"group"},
		ReplicationPolicy:        kafka.ConnectorConfigMirrorCheckpoint_IDENTITY,
		OffsetSyncsTopicLocation: kafka.ConnectorConfigMirrorCheckpoint_TARGET,
	})))
	assert.Equal(t, "mirror-source", data.Get("connector_config_mirror_checkpoint.0.source_connector_name"))
	assert.Equal(t, "IDENTITY", data.Get("connector_config_mirror_checkpoint.0.replication_policy"))
	assert.Equal(t, "TARGET", data.Get("connector_config_mirror_checkpoint.0.offset_syncs_topic_location"))
}

func kafkaConnectorMirrorCheckpointRaw(config map[string]any) map[string]any {
	return map[string]any{
		"cluster_id":                         "cluster-id",
		"name":                               "checkpoint",
		"connector_config_mirror_checkpoint": []any{config},
	}
}

func TestKafkaConnectorSASLPassword(t *testing.T) {
	prefix := svc_kafka.KafkaConnectorSASLPasswordPrefixes[0]
	legacy := testKafkaConnectorPasswordConfig{
		value:  cty.EmptyObjectVal,
		values: map[string]any{prefix + "sasl_password": "legacy-secret"},
	}
	assert.Equal(t, "legacy-secret", svc_kafka.KafkaConnectorSASLPassword(legacy, prefix))

	writeOnly := testKafkaConnectorPasswordConfig{
		value: kafkaConnectorRawConfig(map[string]cty.Value{
			"sasl_password_wo": cty.StringVal("write-only-secret"),
		}),
		values: map[string]any{prefix + "sasl_password": "legacy-secret"},
	}
	assert.Equal(t, "write-only-secret", svc_kafka.KafkaConnectorSASLPassword(writeOnly, prefix))
}

func TestBuildKafkaConnectorSpecMirrorMakerPasswordWo(t *testing.T) {
	raw := kafkaConnectorMirrorMakerRaw(
		map[string]any{
			"bootstrap_servers":        "source.example:9092",
			"sasl_password_wo":         "write-only-secret",
			"sasl_password_wo_version": 1,
		},
		nil,
	)
	resourceData := testResourceDataRawWithWriteOnly(t, svc_kafka.ResourceYandexMDBKafkaConnector(), raw)

	connectorSpec, err := svc_kafka.BuildKafkaConnectorSpec(resourceData)

	require.NoError(t, err)
	assert.Equal(
		t,
		"write-only-secret",
		connectorSpec.GetConnectorConfigMirrormaker().SourceCluster.GetExternalCluster().SaslPassword,
	)
}

func TestValidateKafkaConnectorSASLPasswordFields(t *testing.T) {
	tests := []struct {
		name      string
		external  map[string]cty.Value
		validator func(d mdbcommon.RawConfigProvider) error
		wantError string
	}{
		{
			name: "write-only password with version",
			external: map[string]cty.Value{
				"sasl_password_wo":         cty.StringVal("source-secret"),
				"sasl_password_wo_version": cty.NumberIntVal(1),
			},
			validator: svc_kafka.ValidateKafkaConnectorSASLPasswordPair,
		},
		{
			name: "unknown write-only password defers pair validation",
			external: map[string]cty.Value{
				"sasl_password_wo":         cty.UnknownVal(cty.String),
				"sasl_password_wo_version": cty.NumberIntVal(1),
			},
			validator: svc_kafka.ValidateKafkaConnectorSASLPasswordPair,
		},
		{
			name: "legacy and write-only passwords conflict",
			external: map[string]cty.Value{
				"sasl_password":            cty.StringVal("legacy-secret"),
				"sasl_password_wo":         cty.StringVal("source-secret"),
				"sasl_password_wo_version": cty.NumberIntVal(1),
			},
			validator: svc_kafka.ValidateKafkaConnectorSASLPasswordConflict,
			wantError: "only one of `connector_config_mirrormaker.source_cluster.external_cluster.sasl_password` or " +
				"`connector_config_mirrormaker.source_cluster.external_cluster.sasl_password_wo` can be specified",
		},
		{
			name: "write-only password without version",
			external: map[string]cty.Value{
				"sasl_password_wo": cty.StringVal("source-secret"),
			},
			validator: svc_kafka.ValidateKafkaConnectorSASLPasswordPair,
			wantError: "`connector_config_mirrormaker.source_cluster.external_cluster.sasl_password_wo` and " +
				"`connector_config_mirrormaker.source_cluster.external_cluster.sasl_password_wo_version` must be specified together",
		},
		{
			name: "unrelated unknown does not defer pair validation",
			external: map[string]cty.Value{
				"sasl_password_wo": cty.StringVal("source-secret"),
				"sasl_username":    cty.UnknownVal(cty.String),
			},
			validator: svc_kafka.ValidateKafkaConnectorSASLPasswordPair,
			wantError: "`connector_config_mirrormaker.source_cluster.external_cluster.sasl_password_wo` and " +
				"`connector_config_mirrormaker.source_cluster.external_cluster.sasl_password_wo_version` must be specified together",
		},
		{
			name: "version without write-only password",
			external: map[string]cty.Value{
				"sasl_password_wo_version": cty.NumberIntVal(1),
			},
			validator: svc_kafka.ValidateKafkaConnectorSASLPasswordPair,
			wantError: "`connector_config_mirrormaker.source_cluster.external_cluster.sasl_password_wo` and " +
				"`connector_config_mirrormaker.source_cluster.external_cluster.sasl_password_wo_version` must be specified together",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := testKafkaConnectorPasswordConfig{value: kafkaConnectorRawConfig(tt.external)}

			err := tt.validator(config)
			if tt.wantError == "" {
				require.NoError(t, err)
				return
			}
			require.EqualError(t, err, tt.wantError)
		})
	}
}

type testKafkaConnectorPasswordConfig struct {
	value  cty.Value
	values map[string]any
}

func (c testKafkaConnectorPasswordConfig) GetRawConfig() cty.Value {
	return c.value
}

func (c testKafkaConnectorPasswordConfig) GetOk(key string) (any, bool) {
	value, ok := c.values[key]
	return value, ok
}

func kafkaConnectorRawConfig(external map[string]cty.Value) cty.Value {
	return cty.ObjectVal(map[string]cty.Value{
		"connector_config_mirrormaker": cty.TupleVal([]cty.Value{
			cty.ObjectVal(map[string]cty.Value{
				"source_cluster": cty.TupleVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"external_cluster": cty.TupleVal([]cty.Value{cty.ObjectVal(external)}),
					}),
				}),
			}),
		}),
	})
}

func TestFlattenKafkaConnectorMirrormakerPreservesPasswordState(t *testing.T) {
	raw := kafkaConnectorMirrorMakerRaw(
		map[string]any{
			"bootstrap_servers":        "source.example:9092",
			"sasl_password_wo":         "source-secret",
			"sasl_password_wo_version": 4,
		},
		map[string]any{
			"bootstrap_servers": "target.example:9092",
			"sasl_password":     "legacy-target-secret",
		},
	)
	resourceData := schema.TestResourceDataRaw(t, svc_kafka.ResourceYandexMDBKafkaConnector().Schema, raw)
	mirrorMaker := &kafka.ConnectorConfigMirrorMaker{
		SourceCluster: kafkaExternalClusterConnection("source", "source.example:9092"),
		TargetCluster: kafkaExternalClusterConnection("target", "target.example:9092"),
	}

	flattened, err := svc_kafka.FlattenKafkaConnectorMirrormaker(mirrorMaker, resourceData)
	require.NoError(t, err)
	require.Len(t, flattened, 1)

	source := flattenedKafkaExternalCluster(t, flattened[0], "source_cluster")
	assert.Equal(t, 4, source["sasl_password_wo_version"])
	assert.NotContains(t, source, "sasl_password_wo")
	assert.NotContains(t, source, "sasl_password")

	target := flattenedKafkaExternalCluster(t, flattened[0], "target_cluster")
	assert.Equal(t, "legacy-target-secret", target["sasl_password"])
}

func TestKafkaConnectorPasswordWoUpdateMask(t *testing.T) {
	for _, prefix := range svc_kafka.KafkaConnectorSASLPasswordPrefixes {
		assert.Equal(
			t,
			svc_kafka.MdbKafkaConnectorUpdateFieldsMap[prefix+"sasl_password"],
			svc_kafka.MdbKafkaConnectorUpdateFieldsMap[prefix+"sasl_password_wo_version"],
		)
	}
}

func TestRedactKafkaConnectorRequests(t *testing.T) {
	const (
		sourcePassword = "source-secret"
		targetPassword = "target-secret"
	)
	mirrorMaker := &kafka.ConnectorConfigMirrorMakerSpec{
		SourceCluster: kafkaExternalClusterConnectionSpec("source", sourcePassword),
		TargetCluster: kafkaExternalClusterConnectionSpec("target", targetPassword),
	}
	createSpec := &kafka.ConnectorSpec{Name: "connector"}
	createSpec.Properties = map[string]string{"sasl.jaas.config": "property-secret"}
	createSpec.SetConnectorConfigMirrormaker(mirrorMaker)
	createRequest := &kafka.CreateConnectorRequest{ClusterId: "cluster-id", ConnectorSpec: createSpec}

	redactedCreate := svc_kafka.RedactKafkaConnectorCreateRequest(createRequest)

	assert.Equal(t, sourcePassword, createRequest.ConnectorSpec.GetConnectorConfigMirrormaker().SourceCluster.GetExternalCluster().SaslPassword)
	assert.Equal(t, targetPassword, createRequest.ConnectorSpec.GetConnectorConfigMirrormaker().TargetCluster.GetExternalCluster().SaslPassword)
	assert.Equal(t, svc_kafka.RedactedKafkaSecret, redactedCreate.ConnectorSpec.GetConnectorConfigMirrormaker().SourceCluster.GetExternalCluster().SaslPassword)
	assert.Equal(t, svc_kafka.RedactedKafkaSecret, redactedCreate.ConnectorSpec.GetConnectorConfigMirrormaker().TargetCluster.GetExternalCluster().SaslPassword)
	assert.Equal(t, "property-secret", createRequest.ConnectorSpec.Properties["sasl.jaas.config"])
	assert.Equal(t, svc_kafka.RedactedKafkaSecret, redactedCreate.ConnectorSpec.Properties["sasl.jaas.config"])
	assert.NotContains(t, fmt.Sprintf("%+v", redactedCreate), sourcePassword)
	assert.NotContains(t, fmt.Sprintf("%+v", redactedCreate), targetPassword)
	assert.NotContains(t, fmt.Sprintf("%+v", redactedCreate), "property-secret")
	assert.NotSame(t, createRequest, redactedCreate)

	updateSpec := &kafka.UpdateConnectorSpec{Properties: map[string]string{"token": "update-property-secret"}}
	updateSpec.SetConnectorConfigMirrormaker(mirrorMaker)
	updateRequest := &kafka.UpdateConnectorRequest{ClusterId: "cluster-id", ConnectorName: "connector", ConnectorSpec: updateSpec}
	redactedUpdate := svc_kafka.RedactKafkaConnectorUpdateRequest(updateRequest)

	assert.Equal(t, sourcePassword, updateRequest.ConnectorSpec.GetConnectorConfigMirrormaker().SourceCluster.GetExternalCluster().SaslPassword)
	assert.Equal(t, svc_kafka.RedactedKafkaSecret, redactedUpdate.ConnectorSpec.GetConnectorConfigMirrormaker().SourceCluster.GetExternalCluster().SaslPassword)
	assert.Equal(t, "update-property-secret", updateRequest.ConnectorSpec.Properties["token"])
	assert.Equal(t, svc_kafka.RedactedKafkaSecret, redactedUpdate.ConnectorSpec.Properties["token"])
	assert.NotContains(t, fmt.Sprintf("%+v", redactedUpdate), sourcePassword)
	assert.NotContains(t, fmt.Sprintf("%+v", redactedUpdate), "update-property-secret")
}

func TestRedactKafkaConnectorStorageSecrets(t *testing.T) {
	const (
		s3Secret      = "s3-secret"
		icebergSecret = "iceberg-secret"
	)

	s3Spec := &kafka.ConnectorSpec{Name: "s3-connector"}
	s3Spec.SetConnectorConfigS3Sink(&kafka.ConnectorConfigS3SinkSpec{
		S3Connection: &kafka.S3ConnectionSpec{
			Storage: &kafka.S3ConnectionSpec_ExternalS3{
				ExternalS3: &kafka.ExternalS3StorageSpec{SecretAccessKey: s3Secret},
			},
		},
	})
	s3Request := &kafka.CreateConnectorRequest{ConnectorSpec: s3Spec}

	redactedS3 := svc_kafka.RedactKafkaConnectorCreateRequest(s3Request)

	assert.Equal(t, s3Secret, s3Request.ConnectorSpec.GetConnectorConfigS3Sink().S3Connection.GetExternalS3().SecretAccessKey)
	assert.Equal(t, svc_kafka.RedactedKafkaSecret, redactedS3.ConnectorSpec.GetConnectorConfigS3Sink().S3Connection.GetExternalS3().SecretAccessKey)
	assert.NotContains(t, fmt.Sprintf("%+v", redactedS3), s3Secret)

	icebergSpec := &kafka.UpdateConnectorSpec{}
	icebergSpec.SetConnectorConfigIcebergSink(&kafka.UpdateConnectorConfigIcebergSinkSpec{
		S3Connection: &kafka.IcebergS3ConnectionSpec{
			Storage: &kafka.IcebergS3ConnectionSpec_ExternalS3{
				ExternalS3: &kafka.ExternalIcebergS3StorageSpec{SecretAccessKey: icebergSecret},
			},
		},
	})
	icebergRequest := &kafka.UpdateConnectorRequest{ConnectorSpec: icebergSpec}

	redactedIceberg := svc_kafka.RedactKafkaConnectorUpdateRequest(icebergRequest)

	assert.Equal(t, icebergSecret, icebergRequest.ConnectorSpec.GetConnectorConfigIcebergSink().S3Connection.GetExternalS3().SecretAccessKey)
	assert.Equal(t, svc_kafka.RedactedKafkaSecret, redactedIceberg.ConnectorSpec.GetConnectorConfigIcebergSink().S3Connection.GetExternalS3().SecretAccessKey)
	assert.NotContains(t, fmt.Sprintf("%+v", redactedIceberg), icebergSecret)
}

func TestKafkaConnectorDataSourceExcludesPasswordWo(t *testing.T) {
	dataSource := svc_kafka.DataSourceYandexMDBKafkaConnector()
	mirrorMaker := dataSource.Schema["connector_config_mirrormaker"].Elem.(*schema.Resource)
	sourceCluster := mirrorMaker.Schema["source_cluster"].Elem.(*schema.Resource)
	externalCluster := sourceCluster.Schema["external_cluster"].Elem.(*schema.Resource)

	assert.NotContains(t, externalCluster.Schema, "sasl_password_wo")
	assert.NotContains(t, externalCluster.Schema, "sasl_password_wo_version")
}

func kafkaConnectorMirrorMakerRaw(sourceExternal, targetExternal map[string]any) map[string]any {
	sourceCluster := map[string]any{"alias": "source"}
	if sourceExternal == nil {
		sourceCluster["this_cluster"] = []any{map[string]any{}}
	} else {
		sourceCluster["external_cluster"] = []any{sourceExternal}
	}
	targetCluster := map[string]any{"alias": "target"}
	if targetExternal == nil {
		targetCluster["this_cluster"] = []any{map[string]any{}}
	} else {
		targetCluster["external_cluster"] = []any{targetExternal}
	}

	return map[string]any{
		"cluster_id": "cluster-id",
		"name":       "connector",
		"connector_config_mirrormaker": []any{
			map[string]any{
				"topics":             "topics-*",
				"replication_factor": 1,
				"source_cluster":     []any{sourceCluster},
				"target_cluster":     []any{targetCluster},
			},
		},
	}
}

func kafkaExternalClusterConnection(alias, bootstrapServers string) *kafka.ClusterConnection {
	return &kafka.ClusterConnection{
		Alias: alias,
		ClusterConnection: &kafka.ClusterConnection_ExternalCluster{
			ExternalCluster: &kafka.ExternalClusterConnection{BootstrapServers: bootstrapServers},
		},
	}
}

func kafkaExternalClusterConnectionSpec(alias, password string) *kafka.ClusterConnectionSpec {
	return &kafka.ClusterConnectionSpec{
		Alias: alias,
		ClusterConnection: &kafka.ClusterConnectionSpec_ExternalCluster{
			ExternalCluster: &kafka.ExternalClusterConnectionSpec{SaslPassword: password},
		},
	}
}

func flattenedKafkaExternalCluster(t *testing.T, mirrorMaker map[string]any, field string) map[string]any {
	t.Helper()
	cluster, ok := mirrorMaker[field].([]map[string]any)
	require.True(t, ok)
	require.Len(t, cluster, 1)
	external, ok := cluster[0]["external_cluster"].([]map[string]any)
	require.True(t, ok)
	require.Len(t, external, 1)
	return external[0]
}
