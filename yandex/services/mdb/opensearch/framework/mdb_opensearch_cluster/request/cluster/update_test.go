package cluster

import (
	"context"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestPrepareConfigChangeAdminPasswordWo(t *testing.T) {
	t.Run("version change rotates write-only password", func(t *testing.T) {
		state := passwordTestConfig(types.StringNull(), types.StringNull(), types.Int64Value(1))
		plan := passwordTestConfig(types.StringNull(), types.StringNull(), types.Int64Value(2))
		update, mask, diags := prepareConfigChange(context.Background(), plan, state, types.StringValue("rotated-password"))

		if diags.HasError() {
			t.Fatalf("prepareConfigChange() diagnostics: %#v", diags)
		}
		if update.AdminPassword != "rotated-password" {
			t.Fatalf("AdminPassword = %q, want %q", update.AdminPassword, "rotated-password")
		}
		if !slices.Contains(mask, "config_spec.admin_password") {
			t.Fatalf("update mask = %v, want config_spec.admin_password", mask)
		}
	})

	t.Run("version change without write-only password returns an error", func(t *testing.T) {
		state := passwordTestConfig(types.StringNull(), types.StringNull(), types.Int64Value(1))
		plan := passwordTestConfig(types.StringNull(), types.StringNull(), types.Int64Value(2))
		update, mask, diags := prepareConfigChange(context.Background(), plan, state, types.StringNull())

		if !diags.HasError() {
			t.Fatal("prepareConfigChange() diagnostics has no error")
		}
		if update != nil {
			t.Fatalf("update = %#v, want nil", update)
		}
		if mask != nil {
			t.Fatalf("update mask = %v, want nil", mask)
		}
	})

	t.Run("same version ignores write-only password", func(t *testing.T) {
		state := passwordTestConfig(types.StringNull(), types.StringNull(), types.Int64Value(1))
		plan := passwordTestConfig(types.StringNull(), types.StringNull(), types.Int64Value(1))
		update, mask, diags := prepareConfigChange(context.Background(), plan, state, types.StringValue("different-password"))

		if diags.HasError() {
			t.Fatalf("prepareConfigChange() diagnostics: %#v", diags)
		}
		if update.AdminPassword != "" {
			t.Fatalf("AdminPassword = %q, want empty", update.AdminPassword)
		}
		if slices.Contains(mask, "config_spec.admin_password") {
			t.Fatalf("update mask = %v, must not contain config_spec.admin_password", mask)
		}
	})

	t.Run("switching back to legacy password does not clear it", func(t *testing.T) {
		state := passwordTestConfig(types.StringNull(), types.StringNull(), types.Int64Value(2))
		plan := passwordTestConfig(types.StringValue("legacy-password"), types.StringNull(), types.Int64Null())
		update, mask, diags := prepareConfigChange(context.Background(), plan, state, types.StringNull())

		if diags.HasError() {
			t.Fatalf("prepareConfigChange() diagnostics: %#v", diags)
		}
		if update.AdminPassword != "legacy-password" {
			t.Fatalf("AdminPassword = %q, want %q", update.AdminPassword, "legacy-password")
		}
		if !slices.Contains(mask, "config_spec.admin_password") {
			t.Fatalf("update mask = %v, want config_spec.admin_password", mask)
		}
	})
}
