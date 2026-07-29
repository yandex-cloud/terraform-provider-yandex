package legacy

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_opensearch_cluster/model"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestTransformConfigV1AddsCurrentAttributes(t *testing.T) {
	ctx := context.Background()
	oldOpenSearchNodeType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"name":             types.StringType,
		"resources":        types.ObjectType{AttrTypes: model.NodeResourceAttrTypes},
		"hosts_count":      types.Int64Type,
		"zone_ids":         types.SetType{ElemType: types.StringType},
		"subnet_ids":       types.ListType{ElemType: types.StringType},
		"assign_public_ip": types.BoolType,
		"roles":            types.SetType{ElemType: types.StringType},
	}}
	oldOpenSearchNode := types.ObjectValueMust(oldOpenSearchNodeType.AttrTypes, map[string]attr.Value{
		"name":             types.StringValue("data"),
		"resources":        types.ObjectNull(model.NodeResourceAttrTypes),
		"hosts_count":      types.Int64Value(1),
		"zone_ids":         types.SetValueMust(types.StringType, []attr.Value{types.StringValue("ru-central1-a")}),
		"subnet_ids":       types.ListNull(types.StringType),
		"assign_public_ip": types.BoolValue(false),
		"roles":            types.SetValueMust(types.StringType, []attr.Value{types.StringValue("DATA")}),
	})
	oldOpenSearchAttrTypes := map[string]attr.Type{
		"node_groups": types.ListType{ElemType: oldOpenSearchNodeType},
		"plugins":     types.SetType{ElemType: types.StringType},
	}
	oldOpenSearch := types.ObjectValueMust(oldOpenSearchAttrTypes, map[string]attr.Value{
		"node_groups": types.ListValueMust(oldOpenSearchNodeType, []attr.Value{oldOpenSearchNode}),
		"plugins":     types.SetNull(types.StringType),
	})
	oldConfigAttrTypes := map[string]attr.Type{
		"version":        types.StringType,
		"admin_password": types.StringType,
		"opensearch":     types.ObjectType{AttrTypes: oldOpenSearchAttrTypes},
		"dashboards":     types.ObjectType{AttrTypes: model.DashboardsSubConfigAttrTypes},
		"access":         model.ConfigAttrTypes["access"],
	}
	oldConfig := types.ObjectValueMust(oldConfigAttrTypes, map[string]attr.Value{
		"version":        types.StringValue("2"),
		"admin_password": types.StringValue("legacy-password"),
		"opensearch":     oldOpenSearch,
		"dashboards":     types.ObjectNull(model.DashboardsSubConfigAttrTypes),
		"access":         types.ObjectNull(model.ConfigAttrTypes["access"].(types.ObjectType).AttrTypes),
	})

	got, diags := transformConfigV1(ctx, oldConfig)

	if diags.HasError() {
		t.Fatalf("transformConfigV1() diagnostics: %#v", diags)
	}
	if !got.Type(ctx).Equal(types.ObjectType{AttrTypes: model.ConfigAttrTypes}) {
		t.Fatalf("config type = %#v, want current model.ConfigAttrTypes", got.Type(ctx))
	}

	state := &model.OpenSearch{Config: got}
	config, parseDiags := model.ParseConfig(ctx, state)
	if parseDiags.HasError() {
		t.Fatalf("ParseConfig() diagnostics: %#v", parseDiags)
	}
	if config.AdminPassword.ValueString() != "legacy-password" {
		t.Fatalf("admin_password = %q, want legacy password", config.AdminPassword.ValueString())
	}
	if !config.AdminPasswordWo.IsNull() || !config.AdminPasswordWoVersion.IsNull() {
		t.Fatal("write-only password attributes must be null after v1 state upgrade")
	}

	openSearch, parseDiags := model.ParseOpenSearchSubConfig(ctx, config)
	if parseDiags.HasError() {
		t.Fatalf("ParseOpenSearchSubConfig() diagnostics: %#v", parseDiags)
	}
	var nodeGroups []model.OpenSearchNode
	parseDiags.Append(openSearch.NodeGroups.ElementsAs(ctx, &nodeGroups, false)...)
	if parseDiags.HasError() {
		t.Fatalf("parse upgraded node groups: %#v", parseDiags)
	}
	if len(nodeGroups) != 1 || !nodeGroups[0].DiskSizeAutoscaling.IsNull() {
		t.Fatalf("upgraded node groups = %#v, want one group with null disk_size_autoscaling", nodeGroups)
	}
}
