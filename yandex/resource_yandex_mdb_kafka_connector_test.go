package yandex

import (
	"testing"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/kafka/v1"
)

func TestBuildKafkaConnectorSpec(t *testing.T) {
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
