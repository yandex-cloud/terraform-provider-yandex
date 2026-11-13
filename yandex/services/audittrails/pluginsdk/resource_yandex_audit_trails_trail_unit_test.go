package pluginsdk

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/audittrails/v1"
	"google.golang.org/protobuf/proto"
)

type auditTrailsTrailTestDestination struct {
	name        string
	config      []interface{}
	destination *audittrails.Trail_Destination
}

func auditTrailsTrailTestDestinations() []auditTrailsTrailTestDestination {
	return []auditTrailsTrailTestDestination{
		{
			name:   "monium_destination",
			config: []interface{}{map[string]interface{}{}},
			destination: &audittrails.Trail_Destination{
				Destination: &audittrails.Trail_Destination_Monium{Monium: &audittrails.Trail_Monium{}},
			},
		},
		{
			name: "storage_destination",
			config: []interface{}{map[string]interface{}{
				"bucket_name":        "bucket",
				"object_prefix":      "events/",
				"aggregation_period": "5m0s",
			}},
		},
		{
			name:   "logging_destination",
			config: []interface{}{map[string]interface{}{"log_group_id": "log-group-id"}},
		},
		{
			name: "data_stream_destination",
			config: []interface{}{map[string]interface{}{
				"database_id": "database-id",
				"stream_name": "events",
				"codec":       "GZIP",
			}},
		},
	}
}

func auditTrailsTrailDestinationTestConfig(destinations ...auditTrailsTrailTestDestination) map[string]interface{} {
	config := map[string]interface{}{
		"name":               "trail",
		"folder_id":          "folder-id",
		"service_account_id": "service-account-id",
		"filtering_policy": []interface{}{map[string]interface{}{
			"management_events_filter": []interface{}{map[string]interface{}{
				"resource_scope": []interface{}{map[string]interface{}{
					"resource_id":   "folder-id",
					"resource_type": "resource-manager.folder",
				}},
			}},
		}},
	}
	for _, destination := range destinations {
		config[destination.name] = destination.config
	}
	return config
}

func TestAuditTrailsTrailDestinationValidation(t *testing.T) {
	resource := ResourceYandexAuditTrailsTrail()
	destinations := auditTrailsTrailTestDestinations()

	for i, destination := range destinations {
		t.Run(destination.name, func(t *testing.T) {
			diagnostics := resource.Validate(terraform.NewResourceConfigRaw(auditTrailsTrailDestinationTestConfig(destination)))
			require.False(t, diagnostics.HasError(), "diagnostics: %#v", diagnostics)
		})

		for _, other := range destinations[i+1:] {
			t.Run(destination.name+"_and_"+other.name, func(t *testing.T) {
				config := auditTrailsTrailDestinationTestConfig(destination, other)
				diagnostics := resource.Validate(terraform.NewResourceConfigRaw(config))
				require.True(t, diagnostics.HasError(), "two destinations must be rejected; diagnostics: %#v", diagnostics)
			})
		}
	}

	t.Run("no_destination", func(t *testing.T) {
		diagnostics := resource.Validate(terraform.NewResourceConfigRaw(auditTrailsTrailDestinationTestConfig()))
		require.True(t, diagnostics.HasError(), "a destination is required; diagnostics: %#v", diagnostics)
	})

	t.Run("repeated_destination_block", func(t *testing.T) {
		for _, destination := range destinations {
			t.Run(destination.name, func(t *testing.T) {
				config := auditTrailsTrailDestinationTestConfig(destination)
				config[destination.name] = []interface{}{destination.config[0], destination.config[0]}
				diagnostics := resource.Validate(terraform.NewResourceConfigRaw(config))
				require.True(t, diagnostics.HasError(), "multiple destination blocks must be rejected; diagnostics: %#v", diagnostics)
			})
		}
	})
}

func TestAuditTrailsTrailMoniumDestinationRoundTrip(t *testing.T) {
	resource := ResourceYandexAuditTrailsTrail()
	monium := auditTrailsTrailTestDestinations()[0]
	config := auditTrailsTrailDestinationTestConfig(monium)
	data := schema.TestResourceDataRaw(t, resource.Schema, config)

	destination, err := packResourceDataIntoDestination(data)
	require.NoError(t, err)
	require.True(t, proto.Equal(monium.destination, destination), "destination: %v", destination)
	require.NotNil(t, destination.GetMonium())

	trail := &audittrails.Trail{
		Id:               "trail-id",
		Name:             "trail",
		FolderId:         "folder-id",
		ServiceAccountId: "service-account-id",
		Destination:      destination,
		FilteringPolicy: &audittrails.Trail_FilteringPolicy{
			ManagementEventsFilter: &audittrails.Trail_ManagementEventsFiltering{
				ResourceScopes: []*audittrails.Trail_Resource{{Id: "folder-id", Type: "resource-manager.folder"}},
			},
		},
	}
	diagnostics := unpackProtoTrailIntoResourceData(trail, data)
	require.False(t, diagnostics.HasError(), "diagnostics: %#v", diagnostics)

	// An empty nested block is persisted by its count, without any child attributes.
	state := data.State()
	require.Equal(t, "1", state.Attributes["monium_destination.#"])
	refreshedData := resource.Data(state)
	_, exists := refreshedData.GetOk("monium_destination")
	require.True(t, exists)
	destination, err = packResourceDataIntoDestination(refreshedData)
	require.NoError(t, err)
	require.True(t, proto.Equal(monium.destination, destination), "destination after state round trip: %v", destination)

	diff, err := resource.Diff(context.Background(), state, terraform.NewResourceConfigRaw(config), nil)
	require.NoError(t, err)
	if diff != nil {
		for key, change := range diff.Attributes {
			for _, candidate := range auditTrailsTrailTestDestinations() {
				require.False(t, strings.HasPrefix(key, candidate.name+"."), "unexpected destination diff %s: %#v", key, change)
			}
		}
	}
}
