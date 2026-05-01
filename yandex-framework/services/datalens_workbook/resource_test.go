package datalens_workbook

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datalens/wire"
)

func TestUnmarshalWorkbook(t *testing.T) {
	model := &workbookModel{}
	resp := map[string]interface{}{
		"workbookId":   "wb-1",
		"collectionId": nil,
		"title":        "T",
		"description":  "D",
		"tenantId":     "ten-1",
		"status":       "active",
		"createdBy":    "u1",
		"createdAt":    "2026-01-01T00:00:00Z",
		"updatedBy":    "u2",
		"updatedAt":    "2026-01-02T00:00:00Z",
	}
	if err := wire.Unmarshal(resp, model); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if model.Id.ValueString() != "wb-1" {
		t.Errorf("id: got %q", model.Id.ValueString())
	}
	if !model.CollectionId.IsNull() {
		t.Errorf("collection_id should be null, got %q", model.CollectionId.ValueString())
	}
	if model.Title.ValueString() != "T" {
		t.Errorf("title: got %q", model.Title.ValueString())
	}
	if model.Description.ValueString() != "D" {
		t.Errorf("description: got %q", model.Description.ValueString())
	}
	if model.Status.ValueString() != "active" {
		t.Errorf("status: got %q", model.Status.ValueString())
	}
}

func TestUnmarshalWorkbook_emptyDescriptionStaysNull(t *testing.T) {
	model := &workbookModel{Description: types.StringNull()}
	resp := map[string]interface{}{
		"workbookId":  "wb-2",
		"title":       "T",
		"description": "",
	}
	if err := wire.Unmarshal(resp, model); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !model.Description.IsNull() {
		t.Errorf("description should be null when API returns \"\" (nullIfEmpty), got %q", model.Description.ValueString())
	}
}

func TestUnmarshalWorkbook_collectionId(t *testing.T) {
	model := &workbookModel{}
	resp := map[string]interface{}{
		"workbookId":   "wb-3",
		"collectionId": "coll-1",
		"title":        "T",
	}
	if err := wire.Unmarshal(resp, model); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if model.CollectionId.ValueString() != "coll-1" {
		t.Errorf("collection_id: got %q", model.CollectionId.ValueString())
	}
}

func TestMarshalWorkbook_OmitsNullAndUntagged(t *testing.T) {
	model := &workbookModel{
		Id:             types.StringValue("wb-1"),
		OrganizationId: types.StringValue("org-1"), // wire:"-" → must be omitted
		Title:          types.StringValue("T"),
		Description:    types.StringNull(), // null → must be omitted
		CollectionId:   types.StringValue("c1"),
	}
	got, err := wire.Marshal(model)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := map[string]any{
		"workbookId":   "wb-1",
		"title":        "T",
		"collectionId": "c1",
	}
	if len(got) != len(want) {
		t.Errorf("unexpected key count: got %d (%#v), want %d", len(got), got, len(want))
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s: got %v, want %v", k, got[k], v)
		}
	}
	if _, ok := got["organization_id"]; ok {
		t.Errorf("organization_id should not be marshalled (wire:\"-\")")
	}
	if _, ok := got["description"]; ok {
		t.Errorf("description should not be marshalled when null")
	}
}
