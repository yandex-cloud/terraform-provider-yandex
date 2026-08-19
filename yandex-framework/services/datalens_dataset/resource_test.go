package datalens_dataset

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datalens/wire"
)

func TestMarshalDataset_basic(t *testing.T) {
	plan := &datasetModel{
		Name:       types.StringValue("ds-1"),
		WorkbookId: types.StringValue("wb"),
		Dataset: &datasetContentModel{
			Description: types.StringValue("d"),
			AvatarRelations: []avatarRelationModel{{
				Id:            types.StringValue("r1"),
				LeftAvatarId:  types.StringValue("a1"),
				RightAvatarId: types.StringValue("a2"),
				JoinType:      types.StringValue("inner"),
				Conditions: []conditionModel{{
					Type:     types.StringValue("binary"),
					Operator: types.StringValue("eq"),
					Left:     &joinPartModel{CalcMode: types.StringValue("direct"), Source: types.StringValue("col")},
					Right:    &joinPartModel{CalcMode: types.StringValue("direct"), Source: types.StringValue("col")},
				}},
			}},
			SourceAvatars: []sourceAvatarModel{{
				Id:       types.StringValue("a1"),
				SourceId: types.StringValue("s1"),
				Title:    types.StringValue("avatar 1"),
			}},
			Sources: []dataSourceModel{{
				Id:           types.StringValue("s1"),
				Title:        types.StringValue("table"),
				SourceType:   types.StringValue("CH_TABLE"),
				ConnectionId: types.StringValue("conn"),
				Parameters: &sourceParametersModel{
					TableName: types.StringValue("t"),
					DbName:    types.StringValue("db"),
				},
			}},
			ResultSchema: []resultSchemaFieldModel{{
				Guid:     types.StringValue("g1"),
				Title:    types.StringValue("Country"),
				DataType: types.StringValue("string"),
				Type:     types.StringValue("DIMENSION"),
			}},
		},
	}
	body, err := wire.Marshal(plan)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if body["name"] != "ds-1" {
		t.Errorf("name: got %v", body["name"])
	}
	if body["workbook_id"] != "wb" {
		t.Errorf("workbook_id: got %v", body["workbook_id"])
	}
	dataset, _ := body["dataset"].(map[string]any)
	if dataset == nil {
		t.Fatalf("dataset block missing: %+v", body)
	}
	if dataset["description"] != "d" {
		t.Errorf("dataset.description: got %v", dataset["description"])
	}
	rels, ok := dataset["avatar_relations"].([]any)
	if !ok || len(rels) != 1 {
		t.Fatalf("avatar_relations: got %+v", dataset["avatar_relations"])
	}
	rel := rels[0].(map[string]any)
	if rel["join_type"] != "inner" {
		t.Errorf("join_type: got %v", rel["join_type"])
	}
	conds := rel["conditions"].([]any)
	if len(conds) != 1 {
		t.Errorf("conditions: got %d", len(conds))
	}
	cond := conds[0].(map[string]any)
	leftPart := cond["left"].(map[string]any)
	if leftPart["calc_mode"] != "direct" {
		t.Errorf("left.calc_mode: got %v", leftPart["calc_mode"])
	}
}

func TestPopulateDatasetFromResponse(t *testing.T) {
	model := &datasetModel{}
	resp := map[string]interface{}{
		"id":          "ds-1",
		"name":        "ds-1",
		"key":         "/x",
		"is_favorite": true,
		"dataset": map[string]interface{}{
			"description":             "hello",
			"load_preview_by_default": true,
			"avatar_relations": []interface{}{
				map[string]interface{}{
					"id":              "r1",
					"left_avatar_id":  "a1",
					"right_avatar_id": "a2",
					"join_type":       "inner",
					"conditions": []interface{}{
						map[string]interface{}{
							"type":     "binary",
							"operator": "eq",
							"left":     map[string]interface{}{"calc_mode": "direct", "source": "col"},
							"right":    map[string]interface{}{"calc_mode": "direct", "source": "col"},
						},
					},
				},
			},
			"result_schema": []interface{}{
				map[string]interface{}{
					"guid":      "g1",
					"title":     "Country",
					"data_type": "string",
					"type":      "DIMENSION",
				},
			},
		},
	}
	wire.Unmarshal(resp, model)
	if model.Id.ValueString() != "ds-1" {
		t.Errorf("id: got %q", model.Id.ValueString())
	}
	if model.IsFavorite.ValueBool() != true {
		t.Errorf("is_favorite: got %v", model.IsFavorite.ValueBool())
	}
	if model.Dataset == nil {
		t.Fatal("dataset is nil")
	}
	if model.Dataset.LoadPreviewByDefault.ValueBool() != true {
		t.Errorf("load_preview_by_default: got %v", model.Dataset.LoadPreviewByDefault.ValueBool())
	}
	if len(model.Dataset.AvatarRelations) != 1 {
		t.Errorf("avatar_relations: got %d", len(model.Dataset.AvatarRelations))
	}
	if len(model.Dataset.ResultSchema) != 1 {
		t.Errorf("result_schema: got %d", len(model.Dataset.ResultSchema))
	}
}

// TestPopulateRls2_mapShape verifies we read DataLens's spec-shaped map of
// `{<field_guid>: [entry, ...]}` directly into our typed Map<List<Object>>.
func TestPopulateRls2_mapShape(t *testing.T) {
	model := &datasetModel{}
	resp := map[string]interface{}{
		"id":   "ds-1",
		"name": "ds-1",
		"dataset": map[string]interface{}{
			"rls2": map[string]interface{}{
				"f-country": []interface{}{
					map[string]interface{}{
						"pattern_type":  "value",
						"allowed_value": "RU",
						"subject": map[string]interface{}{
							"subject_id":   "user-1",
							"subject_type": "user",
						},
					},
				},
			},
		},
	}
	if err := wire.Unmarshal(resp, model); err != nil {
		t.Fatalf("unmarshalDatasetResponse: %v", err)
	}
	if model.Dataset == nil || len(model.Dataset.Rls2) != 1 {
		t.Fatalf("expected 1 rls2 entry, got %+v", model.Dataset)
	}
	bucket := model.Dataset.Rls2["f-country"]
	if len(bucket) != 1 {
		t.Fatalf("f-country bucket: %v", bucket)
	}
	entry := bucket[0]
	if entry.AllowedValue.ValueString() != "RU" {
		t.Errorf("allowed_value: %v", entry.AllowedValue)
	}
	if entry.Subject == nil || entry.Subject.SubjectId.ValueString() != "user-1" {
		t.Errorf("subject_id: %v", entry.Subject)
	}
}
