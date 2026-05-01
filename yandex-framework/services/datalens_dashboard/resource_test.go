package datalens_dashboard

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestMarshalDashboard(t *testing.T) {
	plan := &dashboardModel{
		Entry: &dashboardEntryModel{
			Name:       types.StringValue("dash-1"),
			WorkbookId: types.StringValue("wb"),
			Annotation: &dashboardAnnotationModel{Description: types.StringValue("d")},
			Meta:       &dashboardMetaModel{Title: types.StringValue("title"), Locale: types.StringValue("en")},
			Data: &dashboardDataModel{
				Counter:       types.Int64Value(1),
				Salt:          types.StringValue("abc"),
				SchemeVersion: types.Int64Value(8),
				Settings: &dashboardSettingsModel{
					SilentLoading:      types.BoolValue(false),
					DependentSelectors: types.BoolValue(false),
					ExpandTOC:          types.BoolValue(false),
				},
				Tabs: []dashboardTabModel{{
					Id:    types.StringValue("t1"),
					Title: types.StringValue("Tab"),
				}},
			},
		},
	}
	body, err := marshalDashboard(plan)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := body["mode"]; ok {
		t.Errorf("mode: should not be sent, got %v", body["mode"])
	}
	entry, _ := body["entry"].(map[string]any)
	if entry == nil {
		t.Fatalf("entry missing: %+v", body)
	}
	if entry["workbookId"] != "wb" {
		t.Errorf("workbookId: got %v", entry["workbookId"])
	}
	if ann, ok := entry["annotation"].(map[string]any); !ok || ann["description"] != "d" {
		t.Errorf("annotation: %v", entry["annotation"])
	}
	meta, _ := entry["meta"].(map[string]any)
	if meta == nil || meta["title"] != "title" || meta["locale"] != "en" {
		t.Errorf("meta: %+v", entry["meta"])
	}
	data, _ := entry["data"].(map[string]any)
	if data == nil {
		t.Fatalf("data missing: %+v", entry)
	}
	if data["counter"] != int64(1) {
		t.Errorf("counter: got %v", data["counter"])
	}
	if data["salt"] != "abc" {
		t.Errorf("salt: got %v", data["salt"])
	}
	if data["schemeVersion"] != int64(8) {
		t.Errorf("schemeVersion: got %v", data["schemeVersion"])
	}
	tabs, _ := data["tabs"].([]any)
	if len(tabs) != 1 {
		t.Fatalf("tabs: got %d", len(tabs))
	}
	tab := tabs[0].(map[string]any)
	if tab["id"] != "t1" || tab["title"] != "Tab" {
		t.Errorf("tab: %+v", tab)
	}
}

func TestUnmarshalDashboard(t *testing.T) {
	model := &dashboardModel{}
	resp := map[string]interface{}{
		"entry": map[string]interface{}{
			"entryId":    "dash-1",
			"key":        "/x/My Dashboard",
			"createdAt":  "2026",
			"updatedAt":  "2027",
			"revId":      "r",
			"workbookId": "wb",
			"data": map[string]interface{}{
				"counter":            float64(1),
				"salt":               "abc",
				"schemeVersion":      float64(8),
				"accessDescription":  "",
				"supportDescription": "",
				"settings": map[string]interface{}{
					"silentLoading":      false,
					"dependentSelectors": false,
					"expandTOC":          false,
					"hideDashTitle":      false,
					"hideTabs":           false,
				},
				"tabs": []interface{}{
					map[string]interface{}{
						"id":    "t1",
						"title": "Tab",
						"items": []interface{}{
							map[string]interface{}{
								"id":        "w1",
								"type":      "widget",
								"namespace": "default",
								"data": map[string]interface{}{
									"hideTitle": false,
									"tabs": []interface{}{
										map[string]interface{}{
											"id":      "wt1",
											"title":   "Chart",
											"chartId": "ch1",
										},
									},
								},
							},
						},
					},
				},
			},
			"annotation": map[string]interface{}{"description": "hi"},
			"meta":       map[string]interface{}{"title": "T"},
		},
	}
	if err := unmarshalDashboardResponse(model, resp); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if model.Id.ValueString() != "dash-1" {
		t.Errorf("id: got %q", model.Id.ValueString())
	}
	if model.Entry.Name.ValueString() != "My Dashboard" {
		t.Errorf("entry.name: got %q", model.Entry.Name.ValueString())
	}
	if model.Entry.WorkbookId.ValueString() != "wb" {
		t.Errorf("entry.workbook_id: got %q", model.Entry.WorkbookId.ValueString())
	}
	if model.Entry == nil || model.Entry.Data == nil {
		t.Fatal("entry/data missing")
	}
	if model.Entry.Data.Counter.ValueInt64() != 1 {
		t.Errorf("counter: got %d", model.Entry.Data.Counter.ValueInt64())
	}
	if model.Entry.Data.Salt.ValueString() != "abc" {
		t.Errorf("salt: got %q", model.Entry.Data.Salt.ValueString())
	}
	if len(model.Entry.Data.Tabs) != 1 {
		t.Fatalf("tabs: got %d", len(model.Entry.Data.Tabs))
	}
	tab := model.Entry.Data.Tabs[0]
	if len(tab.Items) != 1 {
		t.Fatalf("items: got %d", len(tab.Items))
	}
	item := tab.Items[0]
	if item.Widget == nil {
		t.Errorf("widget block missing: %+v", item)
	}
	if len(item.Widget.Tabs) != 1 || item.Widget.Tabs[0].ChartId.ValueString() != "ch1" {
		t.Errorf("widget tabs: %+v", item.Widget.Tabs)
	}
}
