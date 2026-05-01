package datalens_chart

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestChartRPCSuffix(t *testing.T) {
	cases := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{"wizard", "Wizard", false},
		{"ql", "QL", false},
		{"editor", "", true},
		{"", "", true},
		{"unknown", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := chartRPCSuffix(tc.in)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q", tc.in)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestMarshalChart_wizard(t *testing.T) {
	plan := &chartModel{
		Type:       types.StringValue("wizard"),
		Name:       types.StringValue("my-chart"),
		WorkbookId: types.StringValue("wb-1"),
		Annotation: &chartAnnotationModel{Description: types.StringValue("d")},
		Data: &chartDataModel{
			Visualization: &chartVisualizationModel{Id: types.StringValue("flatTable")},
			Wizard:        &chartWizardModel{DatasetsIds: []string{"ds-1"}},
		},
	}
	body, err := marshalChart(plan)
	if err != nil {
		t.Fatalf("marshalChart: %v", err)
	}
	// template + data.type are injected at api.go boundary via
	// injectChartConstants, not in marshalChart.
	injectChartConstants(body, "wizard")
	if body["template"] != "datalens" {
		t.Errorf("template: got %v", body["template"])
	}
	if body["name"] != "my-chart" {
		t.Errorf("name: got %v", body["name"])
	}
	ann, ok := body["annotation"].(map[string]any)
	if !ok || ann["description"] != "d" {
		t.Errorf("annotation: %v", body["annotation"])
	}
	data, ok := body["data"].(map[string]any)
	if !ok {
		t.Fatalf("data missing: %+v", body)
	}
	if data["type"] != "wizard" {
		t.Errorf("data.type: got %v", data["type"])
	}
	if v, ok := data["datasetsIds"].([]any); !ok || len(v) != 1 || v[0] != "ds-1" {
		t.Errorf("datasetsIds: got %v", data["datasetsIds"])
	}
	vis, ok := data["visualization"].(map[string]any)
	if !ok || vis["id"] != "flatTable" {
		t.Errorf("visualization: got %v", data["visualization"])
	}
}

func TestMarshalChart_ql(t *testing.T) {
	plan := &chartModel{
		Type:       types.StringValue("ql"),
		Name:       types.StringValue("ql-chart"),
		WorkbookId: types.StringValue("wb"),
		Data: &chartDataModel{
			Visualization: &chartVisualizationModel{Id: types.StringValue("metric")},
			Ql: &chartQLModel{
				ChartType:  types.StringValue("sql"),
				Connection: &chartQLConnRefModel{EntryId: types.StringValue("conn-1"), Type: types.StringValue("ydb")},
				QueryValue: types.StringValue("SELECT 1"),
				Params: []chartQLParamModel{{
					Name:         types.StringValue("p"),
					Type:         types.StringValue("string"),
					DefaultValue: types.StringValue("x"),
				}},
			},
		},
	}
	body, err := marshalChart(plan)
	if err != nil {
		t.Fatalf("marshalChart: %v", err)
	}
	data := body["data"].(map[string]any)
	if data["chartType"] != "sql" {
		t.Errorf("chartType: got %v", data["chartType"])
	}
	conn, ok := data["connection"].(map[string]any)
	if !ok || conn["entryId"] != "conn-1" || conn["type"] != "ydb" {
		t.Errorf("connection: got %v", data["connection"])
	}
	if data["queryValue"] != "SELECT 1" {
		t.Errorf("queryValue: got %v", data["queryValue"])
	}
	params, ok := data["params"].([]any)
	if !ok || len(params) != 1 {
		t.Fatalf("params: got %v", data["params"])
	}
}

// Note: variant block presence is no longer enforced client-side — the
// DataLens API rejects a wizard chart without wizard fields, etc.

func TestUnmarshalChart_ql(t *testing.T) {
	model := &chartModel{Type: types.StringValue("ql")}
	resp := map[string]interface{}{
		"entryId":   "ent-123",
		"key":       "/foo/bar",
		"createdAt": "2026-01-01T00:00:00Z",
		"updatedAt": "2026-01-02T00:00:00Z",
		"revId":     "rev1",
		"data": map[string]interface{}{
			"type":       "ql",
			"version":    "7",
			"chartType":  "sql",
			"queryValue": "SELECT 1",
			"connection": map[string]interface{}{
				"entryId": "conn-1",
				"type":    "ydb",
			},
			"visualization": map[string]interface{}{"id": "metric"},
			"params": []interface{}{
				map[string]interface{}{"name": "p", "type": "string", "defaultValue": "x"},
			},
		},
	}
	if err := unmarshalChartResponse(model, resp); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if model.Id.ValueString() != "ent-123" {
		t.Errorf("id: got %q", model.Id.ValueString())
	}
	if model.Name.ValueString() != "bar" {
		t.Errorf("name: got %q", model.Name.ValueString())
	}
	if model.Data == nil || model.Data.Version.ValueString() != "7" {
		t.Errorf("version: got %v", model.Data)
	}
	if model.Data.Visualization == nil || model.Data.Visualization.Id.ValueString() != "metric" {
		t.Errorf("visualization: got %+v", model.Data.Visualization)
	}
	if model.Data.Ql == nil {
		t.Fatal("ql block not populated")
	}
	if model.Data.Ql.QueryValue.ValueString() != "SELECT 1" {
		t.Errorf("queryValue: got %q", model.Data.Ql.QueryValue.ValueString())
	}
	if model.Data.Ql.Connection == nil || model.Data.Ql.Connection.EntryId.ValueString() != "conn-1" {
		t.Errorf("connection: got %+v", model.Data.Ql.Connection)
	}
	if len(model.Data.Ql.Params) != 1 {
		t.Errorf("params: got %d", len(model.Data.Ql.Params))
	}
}
