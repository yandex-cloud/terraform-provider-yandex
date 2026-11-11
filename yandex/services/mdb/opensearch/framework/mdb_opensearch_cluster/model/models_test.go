package model

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/opensearch/v1"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestConfigToStatePreservesAdminPasswordWoVersion(t *testing.T) {
	ctx := context.Background()
	state := &OpenSearch{
		Config: resourceConfigValue(t, types.Int64Value(2)),
	}
	apiConfig := &opensearch.ClusterConfig{
		Version:    "2",
		Opensearch: &opensearch.OpenSearch{},
	}

	got, diags := configToState(ctx, apiConfig, state)

	if diags.HasError() {
		t.Fatalf("configToState() diagnostics: %#v", diags)
	}
	state.Config = got
	config, parseDiags := ParseConfig(ctx, state)
	if parseDiags.HasError() {
		t.Fatalf("ParseConfig() diagnostics: %#v", parseDiags)
	}
	if gotVersion := config.AdminPasswordWoVersion.ValueInt64(); gotVersion != 2 {
		t.Fatalf("admin_password_wo_version = %d, want 2", gotVersion)
	}
	if !config.AdminPasswordWo.IsNull() {
		t.Fatalf("admin_password_wo = %#v, want null", config.AdminPasswordWo)
	}
}

func TestConfigToStateKeepsDataSourceConfigShape(t *testing.T) {
	ctx := context.Background()
	state := &OpenSearch{
		Config: types.ObjectNull(dataSourceConfigAttrTypes),
	}
	apiConfig := &opensearch.ClusterConfig{
		Version:    "2",
		Opensearch: &opensearch.OpenSearch{},
	}

	got, diags := configToState(ctx, apiConfig, state)

	if diags.HasError() {
		t.Fatalf("configToState() diagnostics: %#v", diags)
	}
	if _, ok := got.AttributeTypes(ctx)["admin_password_wo"]; ok {
		t.Fatal("data source config unexpectedly contains admin_password_wo")
	}
}

func resourceConfigValue(t *testing.T, adminPasswordWoVersion types.Int64) types.Object {
	t.Helper()

	openSearch := types.ObjectValueMust(OpenSearchSubConfigAttrTypes, map[string]attr.Value{
		"node_groups": types.ListValueMust(OpenSearchNodeType, nil),
		"plugins":     types.SetNull(types.StringType),
		"config":      types.ObjectNull(OpenSearchConfig2Types),
	})

	return types.ObjectValueMust(ConfigAttrTypes, map[string]attr.Value{
		"version":                   types.StringValue("2"),
		"admin_password":            types.StringNull(),
		"admin_password_wo":         types.StringNull(),
		"admin_password_wo_version": adminPasswordWoVersion,
		"opensearch":                openSearch,
		"dashboards":                types.ObjectNull(DashboardsSubConfigAttrTypes),
		"access":                    types.ObjectNull(accessAttrTypes),
		"audit_log":                 types.ObjectNull(AuditLogTypes),
	})
}
