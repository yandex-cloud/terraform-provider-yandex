package cluster

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_opensearch_cluster/model"
)

func TestPrepareConfigCreateSpecPassword(t *testing.T) {
	t.Run("legacy password", func(t *testing.T) {
		plan := openSearchTestModel(t, passwordTestConfig(
			types.StringValue("legacy-password"),
			types.StringNull(),
			types.Int64Null(),
		))

		spec, diags := prepareConfigCreateSpec(context.Background(), plan, types.StringNull())

		if diags.HasError() {
			t.Fatalf("prepareConfigCreateSpec() diagnostics: %#v", diags)
		}
		if got := spec.AdminPassword; got != "legacy-password" {
			t.Fatalf("AdminPassword = %q, want %q", got, "legacy-password")
		}
	})

	t.Run("write-only password", func(t *testing.T) {
		plan := openSearchTestModel(t, passwordTestConfig(
			types.StringNull(),
			types.StringNull(),
			types.Int64Value(1),
		))

		spec, diags := prepareConfigCreateSpec(context.Background(), plan, types.StringValue("write-only-password"))

		if diags.HasError() {
			t.Fatalf("prepareConfigCreateSpec() diagnostics: %#v", diags)
		}
		if got := spec.AdminPassword; got != "write-only-password" {
			t.Fatalf("AdminPassword = %q, want %q", got, "write-only-password")
		}
	})
}

func passwordTestConfig(adminPassword, adminPasswordWo types.String, adminPasswordWoVersion types.Int64) *model.Config {
	openSearch := types.ObjectValueMust(model.OpenSearchSubConfigAttrTypes, map[string]attr.Value{
		"node_groups": types.ListValueMust(model.OpenSearchNodeType, nil),
		"plugins":     types.SetNull(types.StringType),
		"config":      types.ObjectNull(model.OpenSearchConfig2Types),
	})

	return &model.Config{
		Version:                types.StringValue("2"),
		AdminPassword:          adminPassword,
		AdminPasswordWo:        adminPasswordWo,
		AdminPasswordWoVersion: adminPasswordWoVersion,
		OpenSearch:             openSearch,
		Dashboards:             types.ObjectNull(model.DashboardsSubConfigAttrTypes),
		Access:                 types.ObjectNull(model.ConfigAttrTypes["access"].(types.ObjectType).AttrTypes),
		AuditLog:               types.ObjectNull(model.AuditLogTypes),
	}
}

func openSearchTestModel(t *testing.T, config *model.Config) *model.OpenSearch {
	t.Helper()

	configValue, diags := types.ObjectValueFrom(context.Background(), model.ConfigAttrTypes, config)
	if diags.HasError() {
		t.Fatalf("create config value: %#v", diags)
	}
	return &model.OpenSearch{Config: configValue}
}
