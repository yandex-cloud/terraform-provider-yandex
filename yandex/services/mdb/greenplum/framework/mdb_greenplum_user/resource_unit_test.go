package mdb_greenplum_user

import (
	"context"
	"slices"
	"testing"

	frameworkdatasource "github.com/hashicorp/terraform-plugin-framework/datasource"
	frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/greenplum/v1"
)

func TestGreenplumUserPasswordWoSchema(t *testing.T) {
	var resp frameworkresource.SchemaResponse
	NewResource().Schema(context.Background(), frameworkresource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %#v", resp.Diagnostics)
	}

	legacyPassword := resp.Schema.Attributes["password"].(schema.StringAttribute)
	if !legacyPassword.IsOptional() || legacyPassword.IsRequired() || !legacyPassword.IsSensitive() || len(legacyPassword.Validators) != 1 {
		t.Fatal("password must remain optional and sensitive and require exactly one of password or password_wo")
	}

	writeOnlyPassword := resp.Schema.Attributes["password_wo"].(schema.StringAttribute)
	if !writeOnlyPassword.IsOptional() || !writeOnlyPassword.IsWriteOnly() || !writeOnlyPassword.IsSensitive() || len(writeOnlyPassword.Validators) != 1 {
		t.Fatal("password_wo must be optional, write-only, sensitive, and require its version")
	}

	version := resp.Schema.Attributes["password_wo_version"].(schema.Int64Attribute)
	if !version.IsOptional() || version.IsWriteOnly() || len(version.Validators) != 1 {
		t.Fatal("password_wo_version must be an optional state attribute requiring the password")
	}
}

func TestGreenplumUserDataSourceExcludesWriteOnlyPassword(t *testing.T) {
	var resp frameworkdatasource.SchemaResponse
	NewDataSource().Schema(context.Background(), frameworkdatasource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %#v", resp.Diagnostics)
	}

	if _, ok := resp.Schema.Attributes["password_wo"]; ok {
		t.Fatal("data source schema unexpectedly contains password_wo")
	}
	if _, ok := resp.Schema.Attributes["password_wo_version"]; ok {
		t.Fatal("data source schema unexpectedly contains password_wo_version")
	}
}

func TestGreenplumUserPasswordForCreate(t *testing.T) {
	legacy := "legacy-password"
	plan := &User{Password: &legacy}
	if got := greenplumUserPasswordForCreate(plan, types.StringValue("write-only-password")); got != "write-only-password" {
		t.Fatalf("password = %q, want write-only-password", got)
	}
	if got := greenplumUserPasswordForCreate(plan, types.StringNull()); got != "legacy-password" {
		t.Fatalf("password = %q, want legacy-password", got)
	}
}

func TestGreenplumUserPasswordChange(t *testing.T) {
	t.Run("version change rotates write-only password", func(t *testing.T) {
		state := &User{PasswordWoVersion: types.Int64Value(1)}
		plan := &User{PasswordWoVersion: types.Int64Value(2)}

		password, changed, diags := greenplumUserPasswordChange(plan, state, types.StringValue("rotated-password"))
		if diags.HasError() || !changed || password != "rotated-password" {
			t.Fatalf("password change = (%q, %t, %#v), want rotated password", password, changed, diags)
		}
	})

	t.Run("same version ignores write-only password", func(t *testing.T) {
		state := &User{PasswordWoVersion: types.Int64Value(1)}
		plan := &User{PasswordWoVersion: types.Int64Value(1)}

		password, changed, diags := greenplumUserPasswordChange(plan, state, types.StringValue("different-password"))
		if diags.HasError() || changed || password != "" {
			t.Fatalf("password change = (%q, %t, %#v), want no change", password, changed, diags)
		}
	})

	t.Run("version change requires write-only password", func(t *testing.T) {
		state := &User{PasswordWoVersion: types.Int64Value(1)}
		plan := &User{PasswordWoVersion: types.Int64Value(2)}

		_, _, diags := greenplumUserPasswordChange(plan, state, types.StringNull())
		if !diags.HasError() {
			t.Fatal("greenplumUserPasswordChange() diagnostics has no error")
		}
	})

	t.Run("legacy password change remains supported", func(t *testing.T) {
		oldPassword := "old-password"
		newPassword := "new-password"
		state := &User{Password: &oldPassword, PasswordWoVersion: types.Int64Null()}
		plan := &User{Password: &newPassword, PasswordWoVersion: types.Int64Null()}

		password, changed, diags := greenplumUserPasswordChange(plan, state, types.StringNull())
		if diags.HasError() || !changed || password != "new-password" {
			t.Fatalf("password change = (%q, %t, %#v), want legacy password", password, changed, diags)
		}
	})
}

func TestGreenplumUserPasswordUpdatePathAndState(t *testing.T) {
	paths := getUpdatePaths(&greenplum.User{}, &greenplum.User{}, true)
	if !slices.Contains(paths, "user.password") {
		t.Fatalf("update paths = %v, want user.password", paths)
	}

	legacy := "legacy-password"
	state := User{Password: &legacy, PasswordWo: types.StringValue("must-not-survive"), PasswordWoVersion: types.Int64Value(2)}
	userToState(&greenplum.User{Name: "alice", ResourceGroup: "analytics"}, &state)
	if !state.PasswordWo.IsNull() || state.PasswordWoVersion.ValueInt64() != 2 || state.Password == nil || *state.Password != legacy {
		t.Fatalf("password state was not preserved safely: %#v", state)
	}
}
