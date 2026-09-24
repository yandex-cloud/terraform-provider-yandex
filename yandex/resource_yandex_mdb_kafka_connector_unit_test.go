package yandex

import (
	"fmt"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/kafka/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
)

func TestKafkaConnectorSASLPassword(t *testing.T) {
	prefix := kafkaConnectorSASLPasswordPrefixes[0]
	legacy := testKafkaConnectorPasswordConfig{
		value:  cty.EmptyObjectVal,
		values: map[string]any{prefix + "sasl_password": "legacy-secret"},
	}
	assert.Equal(t, "legacy-secret", kafkaConnectorSASLPassword(legacy, prefix))

	writeOnly := testKafkaConnectorPasswordConfig{
		value: kafkaConnectorRawConfig(map[string]cty.Value{
			"sasl_password_wo": cty.StringVal("write-only-secret"),
		}),
		values: map[string]any{prefix + "sasl_password": "legacy-secret"},
	}
	assert.Equal(t, "write-only-secret", kafkaConnectorSASLPassword(writeOnly, prefix))
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
	resourceData := testResourceDataRawWithWriteOnly(t, resourceYandexMDBKafkaConnector(), raw)

	connectorSpec, err := buildKafkaConnectorSpec(resourceData)

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
			validator: validateKafkaConnectorSASLPasswordPair,
		},
		{
			name: "unknown write-only password defers pair validation",
			external: map[string]cty.Value{
				"sasl_password_wo":         cty.UnknownVal(cty.String),
				"sasl_password_wo_version": cty.NumberIntVal(1),
			},
			validator: validateKafkaConnectorSASLPasswordPair,
		},
		{
			name: "legacy and write-only passwords conflict",
			external: map[string]cty.Value{
				"sasl_password":            cty.StringVal("legacy-secret"),
				"sasl_password_wo":         cty.StringVal("source-secret"),
				"sasl_password_wo_version": cty.NumberIntVal(1),
			},
			validator: validateKafkaConnectorSASLPasswordConflict,
			wantError: "only one of `connector_config_mirrormaker.source_cluster.external_cluster.sasl_password` or " +
				"`connector_config_mirrormaker.source_cluster.external_cluster.sasl_password_wo` can be specified",
		},
		{
			name: "write-only password without version",
			external: map[string]cty.Value{
				"sasl_password_wo": cty.StringVal("source-secret"),
			},
			validator: validateKafkaConnectorSASLPasswordPair,
			wantError: "`connector_config_mirrormaker.source_cluster.external_cluster.sasl_password_wo` and " +
				"`connector_config_mirrormaker.source_cluster.external_cluster.sasl_password_wo_version` must be specified together",
		},
		{
			name: "unrelated unknown does not defer pair validation",
			external: map[string]cty.Value{
				"sasl_password_wo": cty.StringVal("source-secret"),
				"sasl_username":    cty.UnknownVal(cty.String),
			},
			validator: validateKafkaConnectorSASLPasswordPair,
			wantError: "`connector_config_mirrormaker.source_cluster.external_cluster.sasl_password_wo` and " +
				"`connector_config_mirrormaker.source_cluster.external_cluster.sasl_password_wo_version` must be specified together",
		},
		{
			name: "version without write-only password",
			external: map[string]cty.Value{
				"sasl_password_wo_version": cty.NumberIntVal(1),
			},
			validator: validateKafkaConnectorSASLPasswordPair,
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
	resourceData := schema.TestResourceDataRaw(t, resourceYandexMDBKafkaConnector().Schema, raw)
	mirrorMaker := &kafka.ConnectorConfigMirrorMaker{
		SourceCluster: kafkaExternalClusterConnection("source", "source.example:9092"),
		TargetCluster: kafkaExternalClusterConnection("target", "target.example:9092"),
	}

	flattened, err := flattenKafkaConnectorMirrormaker(mirrorMaker, resourceData)
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
	for _, prefix := range kafkaConnectorSASLPasswordPrefixes {
		assert.Equal(
			t,
			mdbKafkaConnectorUpdateFieldsMap[prefix+"sasl_password"],
			mdbKafkaConnectorUpdateFieldsMap[prefix+"sasl_password_wo_version"],
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

	redactedCreate := redactKafkaConnectorCreateRequest(createRequest)

	assert.Equal(t, sourcePassword, createRequest.ConnectorSpec.GetConnectorConfigMirrormaker().SourceCluster.GetExternalCluster().SaslPassword)
	assert.Equal(t, targetPassword, createRequest.ConnectorSpec.GetConnectorConfigMirrormaker().TargetCluster.GetExternalCluster().SaslPassword)
	assert.Equal(t, redactedKafkaSecret, redactedCreate.ConnectorSpec.GetConnectorConfigMirrormaker().SourceCluster.GetExternalCluster().SaslPassword)
	assert.Equal(t, redactedKafkaSecret, redactedCreate.ConnectorSpec.GetConnectorConfigMirrormaker().TargetCluster.GetExternalCluster().SaslPassword)
	assert.Equal(t, "property-secret", createRequest.ConnectorSpec.Properties["sasl.jaas.config"])
	assert.Equal(t, redactedKafkaSecret, redactedCreate.ConnectorSpec.Properties["sasl.jaas.config"])
	assert.NotContains(t, fmt.Sprintf("%+v", redactedCreate), sourcePassword)
	assert.NotContains(t, fmt.Sprintf("%+v", redactedCreate), targetPassword)
	assert.NotContains(t, fmt.Sprintf("%+v", redactedCreate), "property-secret")
	assert.NotSame(t, createRequest, redactedCreate)

	updateSpec := &kafka.UpdateConnectorSpec{Properties: map[string]string{"token": "update-property-secret"}}
	updateSpec.SetConnectorConfigMirrormaker(mirrorMaker)
	updateRequest := &kafka.UpdateConnectorRequest{ClusterId: "cluster-id", ConnectorName: "connector", ConnectorSpec: updateSpec}
	redactedUpdate := redactKafkaConnectorUpdateRequest(updateRequest)

	assert.Equal(t, sourcePassword, updateRequest.ConnectorSpec.GetConnectorConfigMirrormaker().SourceCluster.GetExternalCluster().SaslPassword)
	assert.Equal(t, redactedKafkaSecret, redactedUpdate.ConnectorSpec.GetConnectorConfigMirrormaker().SourceCluster.GetExternalCluster().SaslPassword)
	assert.Equal(t, "update-property-secret", updateRequest.ConnectorSpec.Properties["token"])
	assert.Equal(t, redactedKafkaSecret, redactedUpdate.ConnectorSpec.Properties["token"])
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

	redactedS3 := redactKafkaConnectorCreateRequest(s3Request)

	assert.Equal(t, s3Secret, s3Request.ConnectorSpec.GetConnectorConfigS3Sink().S3Connection.GetExternalS3().SecretAccessKey)
	assert.Equal(t, redactedKafkaSecret, redactedS3.ConnectorSpec.GetConnectorConfigS3Sink().S3Connection.GetExternalS3().SecretAccessKey)
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

	redactedIceberg := redactKafkaConnectorUpdateRequest(icebergRequest)

	assert.Equal(t, icebergSecret, icebergRequest.ConnectorSpec.GetConnectorConfigIcebergSink().S3Connection.GetExternalS3().SecretAccessKey)
	assert.Equal(t, redactedKafkaSecret, redactedIceberg.ConnectorSpec.GetConnectorConfigIcebergSink().S3Connection.GetExternalS3().SecretAccessKey)
	assert.NotContains(t, fmt.Sprintf("%+v", redactedIceberg), icebergSecret)
}

func TestKafkaConnectorDataSourceExcludesPasswordWo(t *testing.T) {
	dataSource := dataSourceYandexMDBKafkaConnector()
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
