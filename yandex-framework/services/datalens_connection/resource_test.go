package datalens_connection

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestMarshalConnection_Ydb(t *testing.T) {
	t.Parallel()

	model := &connectionModel{
		Type:        types.StringValue("ydb"),
		Name:        types.StringValue("my-conn"),
		Description: types.StringValue("desc"),
		Ydb: &ydbConfigModel{
			WorkbookId:       types.StringValue("wb-1"),
			Host:             types.StringValue("ydb.example.com"),
			Port:             types.Int64Value(2135),
			DbName:           types.StringValue("/db"),
			CloudId:          types.StringValue("cloud-1"),
			FolderId:         types.StringValue("folder-1"),
			ServiceAccountId: types.StringValue("sa-1"),
		},
	}
	got, err := marshalConnection(model)
	if err != nil {
		t.Fatalf("marshalConnection: %v", err)
	}

	for k, want := range map[string]any{
		"type":               "ydb",
		"name":               "my-conn",
		"description":        "desc",
		"workbook_id":        "wb-1",
		"host":               "ydb.example.com",
		"port":               int64(2135),
		"db_name":            "/db",
		"cloud_id":           "cloud-1",
		"folder_id":          "folder-1",
		"service_account_id": "sa-1",
	} {
		if got[k] != want {
			t.Errorf("%s: got %v, want %v", k, got[k], want)
		}
	}
}

func TestMarshalConnection_OmitsNullFields(t *testing.T) {
	t.Parallel()

	model := &connectionModel{
		Type:        types.StringValue("ydb"),
		Name:        types.StringValue("c"),
		Description: types.StringNull(),
		Ydb:         &ydbConfigModel{Host: types.StringValue("h"), Port: types.Int64Value(1)},
	}
	got, _ := marshalConnection(model)
	if _, ok := got["description"]; ok {
		t.Errorf("description should not be marshalled when null")
	}
}

func TestUnmarshalConnection_DbTypeAlias(t *testing.T) {
	t.Parallel()

	apiResponse := map[string]any{
		"id":                 "conn-123",
		"db_type":            "ydb", // get response uses db_type
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
	if err := unmarshalConnection(apiResponse, model); err != nil {
		t.Fatalf("unmarshalConnection: %v", err)
	}

	assertEqual(t, "id", model.Id, types.StringValue("conn-123"))
	assertEqual(t, "type", model.Type, types.StringValue("ydb"))
	assertEqual(t, "name", model.Name, types.StringValue("my-connection"))
	assertEqual(t, "description", model.Description, types.StringValue("test connection"))
	assertEqual(t, "created_at", model.CreatedAt, types.StringValue("2025-01-01T00:00:00Z"))
	assertEqual(t, "updated_at", model.UpdatedAt, types.StringValue("2025-06-01T12:00:00Z"))

	if model.Ydb == nil {
		t.Fatal("model.Ydb should not be nil after unmarshal")
	}
	assertEqual(t, "ydb.host", model.Ydb.Host, types.StringValue("ydb.example.com"))
	assertEqual(t, "ydb.port", model.Ydb.Port, types.Int64Value(2135))
}

func TestUnmarshalConnection_NullDescription(t *testing.T) {
	t.Parallel()

	apiResponse := map[string]any{
		"id":          "conn-123",
		"db_type":     "ydb",
		"name":        "my-connection",
		"description": nil,
	}
	model := &connectionModel{}
	if err := unmarshalConnection(apiResponse, model); err != nil {
		t.Fatalf("unmarshalConnection: %v", err)
	}
	assertEqual(t, "description", model.Description, types.StringNull())
}

func TestUnmarshalConnection_EmptyDescriptionStaysNullViaTag(t *testing.T) {
	t.Parallel()

	apiResponse := map[string]any{
		"id":          "conn-123",
		"db_type":     "ydb",
		"name":        "n",
		"description": "",
	}
	model := &connectionModel{Description: types.StringNull()}
	if err := unmarshalConnection(apiResponse, model); err != nil {
		t.Fatalf("unmarshalConnection: %v", err)
	}
	if !model.Description.IsNull() {
		t.Errorf("description should be null when API returns \"\" (nullIfEmpty), got %q", model.Description.ValueString())
	}
}

func assertEqual(t *testing.T, field string, got, want attr.Value) {
	t.Helper()
	if !got.Equal(want) {
		t.Errorf("field %q: got %v, want %v", field, got, want)
	}
}
