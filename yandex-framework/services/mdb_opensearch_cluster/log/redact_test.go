package log

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/opensearch/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_opensearch_cluster/model"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestRedactModel(t *testing.T) {
	ctx := context.Background()
	config, diags := types.ObjectValue(model.ConfigAttrTypes, map[string]attr.Value{
		"version":                   types.StringValue("2"),
		"admin_password":            types.StringValue("legacy-password"),
		"admin_password_wo":         types.StringNull(),
		"admin_password_wo_version": types.Int64Value(1),
		"opensearch":                types.ObjectNull(model.OpenSearchSubConfigAttrTypes),
		"dashboards":                types.ObjectNull(model.DashboardsSubConfigAttrTypes),
		"access":                    types.ObjectNull(model.ConfigAttrTypes["access"].(types.ObjectType).AttrTypes),
		"audit_log":                 types.ObjectNull(model.AuditLogTypes),
	})
	if diags.HasError() {
		t.Fatalf("create config object: %#v", diags)
	}
	source := &model.OpenSearch{Config: config}

	redacted := RedactModel(ctx, source)
	redactedConfig, parseDiags := model.ParseConfig(ctx, redacted)
	if parseDiags.HasError() {
		t.Fatalf("parse redacted config: %#v", parseDiags)
	}
	if got := redactedConfig.AdminPassword.ValueString(); got != RedactedPassword {
		t.Fatalf("admin_password = %q, want %q", got, RedactedPassword)
	}
	if !redactedConfig.AdminPasswordWo.IsNull() {
		t.Fatalf("admin_password_wo = %#v, want null", redactedConfig.AdminPasswordWo)
	}

	sourceConfig, parseDiags := model.ParseConfig(ctx, source)
	if parseDiags.HasError() {
		t.Fatalf("parse source config: %#v", parseDiags)
	}
	if got := sourceConfig.AdminPassword.ValueString(); got != "legacy-password" {
		t.Fatalf("source admin_password = %q, want original value", got)
	}
}

func TestRedactAdminPasswordList(t *testing.T) {
	ctx := context.Background()
	configType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"version":        types.StringType,
		"admin_password": types.StringType,
	}}
	config := types.ObjectValueMust(configType.AttrTypes, map[string]attr.Value{
		"version":        types.StringValue("2"),
		"admin_password": types.StringValue("legacy-password"),
	})
	source := types.ListValueMust(configType, []attr.Value{config})

	redacted := RedactAdminPasswordList(ctx, source)

	redactedConfig := redacted.Elements()[0].(types.Object)
	if got := redactedConfig.Attributes()["admin_password"].(types.String).ValueString(); got != RedactedPassword {
		t.Fatalf("admin_password = %q, want %q", got, RedactedPassword)
	}
	sourceConfig := source.Elements()[0].(types.Object)
	if got := sourceConfig.Attributes()["admin_password"].(types.String).ValueString(); got != "legacy-password" {
		t.Fatalf("source admin_password = %q, want original value", got)
	}
}

func TestRedactClusterRequests(t *testing.T) {
	t.Run("create request", func(t *testing.T) {
		source := &opensearch.CreateClusterRequest{
			ConfigSpec: &opensearch.ConfigCreateSpec{AdminPassword: "secret-password"},
		}

		redacted := RedactCreateClusterRequest(source)

		if got := redacted.ConfigSpec.AdminPassword; got != RedactedPassword {
			t.Fatalf("redacted password = %q, want %q", got, RedactedPassword)
		}
		if got := source.ConfigSpec.AdminPassword; got != "secret-password" {
			t.Fatalf("source password = %q, want original value", got)
		}
	})

	t.Run("update request", func(t *testing.T) {
		source := &opensearch.UpdateClusterRequest{
			ConfigSpec: &opensearch.ConfigUpdateSpec{AdminPassword: "secret-password"},
		}

		redacted := RedactUpdateClusterRequest(source)

		if got := redacted.ConfigSpec.AdminPassword; got != RedactedPassword {
			t.Fatalf("redacted password = %q, want %q", got, RedactedPassword)
		}
		if got := source.ConfigSpec.AdminPassword; got != "secret-password" {
			t.Fatalf("source password = %q, want original value", got)
		}
	})
}
