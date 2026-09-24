package yandex

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/audittrails/v1"
)

func TestUnpackProtoManagementEventsFilterIntoResourceData(t *testing.T) {
	t.Parallel()

	resourceScope := &audittrails.Trail_Resource{
		Id:   "folder-id",
		Type: "resource-manager.folder",
	}
	fieldFilterRule := &audittrails.Trail_FieldFilterRule{
		Conditions: []*audittrails.Trail_FieldCondition{
			{
				Field:    "$.event_type",
				Operator: audittrails.Trail_FieldCondition_IN,
				Values:   []string{"yandex.cloud.audit.audittrails.CreateTrail"},
			},
		},
	}
	flatResourceScope := []interface{}{
		map[string]string{
			"resource_id":   "folder-id",
			"resource_type": "resource-manager.folder",
		},
	}
	flatFieldFilterRule := []interface{}{
		map[string]interface{}{
			"condition": []interface{}{
				map[string]interface{}{
					"field":    "$.event_type",
					"operator": "IN",
					"values":   []string{"yandex.cloud.audit.audittrails.CreateTrail"},
				},
			},
		},
	}

	tests := []struct {
		name     string
		filter   *audittrails.Trail_ManagementEventsFiltering
		expected []interface{}
	}{
		{
			name: "nil filter",
		},
		{
			name:   "empty filter",
			filter: &audittrails.Trail_ManagementEventsFiltering{},
		},
		{
			name: "resource scope",
			filter: &audittrails.Trail_ManagementEventsFiltering{
				ResourceScopes: []*audittrails.Trail_Resource{resourceScope},
			},
			expected: []interface{}{
				map[string]interface{}{"resource_scope": flatResourceScope},
			},
		},
		{
			name: "include rule without resource scope",
			filter: &audittrails.Trail_ManagementEventsFiltering{
				IncludeRules: []*audittrails.Trail_FieldFilterRule{fieldFilterRule},
			},
		},
		{
			name: "include rule with resource scope",
			filter: &audittrails.Trail_ManagementEventsFiltering{
				ResourceScopes: []*audittrails.Trail_Resource{resourceScope},
				IncludeRules:   []*audittrails.Trail_FieldFilterRule{fieldFilterRule},
			},
			expected: []interface{}{
				map[string]interface{}{
					"resource_scope": flatResourceScope,
					"include_rule":   flatFieldFilterRule,
				},
			},
		},
		{
			name: "exclude rule with resource scope",
			filter: &audittrails.Trail_ManagementEventsFiltering{
				ResourceScopes: []*audittrails.Trail_Resource{resourceScope},
				ExcludeRules:   []*audittrails.Trail_FieldFilterRule{fieldFilterRule},
			},
			expected: []interface{}{
				map[string]interface{}{
					"resource_scope": flatResourceScope,
					"exclude_rule":   flatFieldFilterRule,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			actual := unpackProtoManagementEventsFilterIntoResourceData(test.filter)

			require.Equal(t, test.expected, actual)
		})
	}
}
