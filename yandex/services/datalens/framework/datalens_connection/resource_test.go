package datalens_connection

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestFlattenCommonFieldsToMap(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		model    *connectionModel
		expected map[string]interface{}
	}{
		{
			name: "all_fields_set",
			model: &connectionModel{
				Type:        types.StringValue("ydb"),
				Name:        types.StringValue("my-conn"),
				Description: types.StringValue("some desc"),
			},
			expected: map[string]interface{}{
				"type":        "ydb",
				"name":        "my-conn",
				"description": "some desc",
			},
		},
		{
			name: "description_null",
			model: &connectionModel{
				Type:        types.StringValue("ydb"),
				Name:        types.StringValue("my-conn"),
				Description: types.StringNull(),
			},
			expected: map[string]interface{}{
				"type": "ydb",
				"name": "my-conn",
			},
		},
		{
			name: "description_unknown",
			model: &connectionModel{
				Type:        types.StringValue("ydb"),
				Name:        types.StringValue("my-conn"),
				Description: types.StringUnknown(),
			},
			expected: map[string]interface{}{
				"type": "ydb",
				"name": "my-conn",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m := make(map[string]interface{})
			flattenCommonFieldsToMap(c.model, m)
			if !reflect.DeepEqual(m, c.expected) {
				t.Errorf("flattenCommonFieldsToMap:\n  got  %v\n  want %v", m, c.expected)
			}
		})
	}
}

func TestPopulateModelFromResponse_YdbType(t *testing.T) {
	t.Parallel()

	apiResponse := map[string]interface{}{
		"id":                 "conn-123",
		"type":               "ydb",
		"name":               "my-connection",
		"description":        "test connection",
		"created_at":         "2025-01-01T00:00:00Z",
		"updated_at":         "2025-06-01T12:00:00Z",
		"host":               "ydb.example.com",
		"port":               float64(2135),
		"db_name":            "/path/to/db",
		"cloud_id":           "cloud-1",
		"folder_id":          "folder-1",
		"service_account_id": "sa-1",
	}

	model := &connectionModel{}
	r := &connectionResource{}
	r.populateModelFromResponse(model, apiResponse)

	assertEqual(t, "id", model.Id, types.StringValue("conn-123"))
	assertEqual(t, "type", model.Type, types.StringValue("ydb"))
	assertEqual(t, "name", model.Name, types.StringValue("my-connection"))
	assertEqual(t, "description", model.Description, types.StringValue("test connection"))
	assertEqual(t, "created_at", model.CreatedAt, types.StringValue("2025-01-01T00:00:00Z"))
	assertEqual(t, "updated_at", model.UpdatedAt, types.StringValue("2025-06-01T12:00:00Z"))

	if model.Ydb == nil {
		t.Fatal("model.Ydb should not be nil after populateModelFromResponse")
	}
	assertEqual(t, "ydb.host", model.Ydb.Host, types.StringValue("ydb.example.com"))
	assertEqual(t, "ydb.port", model.Ydb.Port, types.Int64Value(2135))
}

func TestPopulateModelFromResponse_NullDescription(t *testing.T) {
	t.Parallel()

	apiResponse := map[string]interface{}{
		"id":          "conn-123",
		"type":        "ydb",
		"name":        "my-connection",
		"description": nil,
	}

	model := &connectionModel{}
	r := &connectionResource{}
	r.populateModelFromResponse(model, apiResponse)

	assertEqual(t, "description", model.Description, types.StringNull())
}

func TestPopulateModelFromResponse_PreservesExistingState(t *testing.T) {
	t.Parallel()

	model := &connectionModel{
		OrganizationId: types.StringValue("org-from-state"),
		Ydb: &ydbConfigModel{
			WorkbookId: types.StringValue("wb-from-state"),
		},
	}

	apiResponse := map[string]interface{}{
		"id":                 "conn-123",
		"type":               "ydb",
		"name":               "my-connection",
		"host":               "ydb.example.com",
		"port":               float64(2135),
		"db_name":            "/path/to/db",
		"cloud_id":           "cloud-1",
		"folder_id":          "folder-1",
		"service_account_id": "sa-1",
	}

	r := &connectionResource{}
	r.populateModelFromResponse(model, apiResponse)

	assertEqual(t, "organization_id", model.OrganizationId, types.StringValue("org-from-state"))
	assertEqual(t, "ydb.workbook_id", model.Ydb.WorkbookId, types.StringValue("wb-from-state"))
	assertEqual(t, "ydb.host", model.Ydb.Host, types.StringValue("ydb.example.com"))
}

func assertEqual(t *testing.T, field string, got, want attr.Value) {
	t.Helper()
	if !got.Equal(want) {
		t.Errorf("field %q: got %v, want %v", field, got, want)
	}
}
