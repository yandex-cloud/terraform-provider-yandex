package mdb_mongodb_user

import (
	"context"
	"reflect"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	frameworkdatasource "github.com/hashicorp/terraform-plugin-framework/datasource"
	frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mongodb/v1"
)

func TestMongoDBUserPasswordWoSchema(t *testing.T) {
	var resp frameworkresource.SchemaResponse
	NewResource().Schema(context.Background(), frameworkresource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %#v", resp.Diagnostics)
	}

	legacyPassword := resp.Schema.Attributes["password"].(schema.StringAttribute)
	if !legacyPassword.IsOptional() || legacyPassword.IsRequired() || !legacyPassword.IsSensitive() || len(legacyPassword.Validators) != 1 {
		t.Fatal("password must remain optional, sensitive, and conflict with password_wo")
	}

	writeOnlyPassword := resp.Schema.Attributes["password_wo"].(schema.StringAttribute)
	if !writeOnlyPassword.IsOptional() || !writeOnlyPassword.IsWriteOnly() || !writeOnlyPassword.IsSensitive() || len(writeOnlyPassword.Validators) != 1 {
		t.Fatal("password_wo must be optional, write-only, sensitive, and require its version")
	}

	version := resp.Schema.Attributes["password_wo_version"].(schema.Int64Attribute)
	if !version.IsOptional() || version.IsWriteOnly() || len(version.Validators) != 1 {
		t.Fatal("password_wo_version must be an optional state attribute requiring password_wo")
	}
}

func TestMongoDBUserPasswordForCreate(t *testing.T) {
	plan := &User{Password: types.StringValue("legacy-password")}
	if got := mongodbUserPasswordForCreate(plan, types.StringValue("write-only-password")); got != "write-only-password" {
		t.Fatalf("password = %q, want write-only-password", got)
	}
	if got := mongodbUserPasswordForCreate(plan, types.StringNull()); got != "legacy-password" {
		t.Fatalf("password = %q, want legacy-password", got)
	}
	if got := mongodbUserPasswordForCreate(&User{AuthType: types.StringValue(authTypeIam)}, types.StringNull()); got != "" {
		t.Fatalf("IAM password = %q, want no password", got)
	}
}

func TestMongoDBUserFromStatePassword(t *testing.T) {
	state := &User{Password: types.StringValue("legacy-password"), Permission: types.SetNull(permissionType)}
	user, diags := userFromState(context.Background(), state, "write-only-password")
	if diags.HasError() {
		t.Fatalf("userFromState diagnostics: %#v", diags)
	}
	if user.Password != "write-only-password" {
		t.Fatalf("API password = %q, want write-only-password", user.Password)
	}
}

func TestMongoDBUserPasswordChange(t *testing.T) {
	tests := []struct {
		name        string
		plan        User
		state       User
		passwordWo  types.String
		want        string
		wantChanged bool
		wantError   bool
	}{
		{
			name: "version change rotates write-only password",
			plan: User{PasswordWoVersion: types.Int64Value(2)}, state: User{PasswordWoVersion: types.Int64Value(1)},
			passwordWo: types.StringValue("rotated-password"), want: "rotated-password", wantChanged: true,
		},
		{
			name: "same version ignores write-only password",
			plan: User{PasswordWoVersion: types.Int64Value(1)}, state: User{PasswordWoVersion: types.Int64Value(1)},
			passwordWo: types.StringValue("different-password"),
		},
		{
			name: "version change requires write-only password",
			plan: User{PasswordWoVersion: types.Int64Value(2)}, state: User{PasswordWoVersion: types.Int64Value(1)}, wantError: true,
		},
		{
			name: "unknown password cannot rotate",
			plan: User{PasswordWoVersion: types.Int64Value(2)}, state: User{PasswordWoVersion: types.Int64Value(1)},
			passwordWo: types.StringUnknown(), wantError: true,
		},
		{
			name: "legacy password update",
			plan: User{Password: types.StringValue("new-password")}, state: User{Password: types.StringValue("old-password")},
			want: "new-password", wantChanged: true,
		},
		{
			name: "unchanged legacy password",
			plan: User{Password: types.StringValue("legacy-password")}, state: User{Password: types.StringValue("legacy-password")},
			want: "legacy-password",
		},
		{
			name: "switch from legacy to write-only",
			plan: User{PasswordWoVersion: types.Int64Value(1)}, state: User{Password: types.StringValue("legacy-password")},
			passwordWo: types.StringValue("write-only-password"), want: "write-only-password", wantChanged: true,
		},
		{
			name: "switch from write-only to legacy",
			plan: User{Password: types.StringValue("legacy-password")}, state: User{PasswordWoVersion: types.Int64Value(1)},
			want: "legacy-password", wantChanged: true,
		},
		{name: "IAM without a password", plan: User{AuthType: types.StringValue(authTypeIam)}, state: User{AuthType: types.StringValue(authTypeIam)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			password, changed, diags := mongodbUserPasswordChange(&tt.plan, &tt.state, tt.passwordWo)
			if diags.HasError() != tt.wantError || password != tt.want || changed != tt.wantChanged {
				t.Fatalf("password change = (%q, %t, %#v), want (%q, %t), error: %t", password, changed, diags, tt.want, tt.wantChanged, tt.wantError)
			}
		})
	}
}

func TestMongoDBUserPasswordUpdatePaths(t *testing.T) {
	if paths := getUpdatePaths(&mongodb.UserSpec{}, &mongodb.UserSpec{}, true); !slices.Equal(paths, []string{"password"}) {
		t.Fatalf("update paths = %v, want password", paths)
	}
	if paths := getUpdatePaths(&mongodb.UserSpec{}, &mongodb.UserSpec{Password: "old-password"}, false); len(paths) != 0 {
		t.Fatalf("update paths = %v, want no change", paths)
	}
	plan := &mongodb.UserSpec{
		Permissions: []*mongodb.Permission{{DatabaseName: "testdb", Roles: []string{"readWrite"}}},
	}
	if paths := getUpdatePaths(plan, &mongodb.UserSpec{}, false); !slices.Equal(paths, []string{"permissions"}) {
		t.Fatalf("update paths = %v, want permissions", paths)
	}
}

func TestMongoDBUserWriteOnlyPasswordState(t *testing.T) {
	state := User{
		Password: types.StringValue("legacy-password"), PasswordWo: types.StringValue("must-not-survive"),
		PasswordWoVersion: types.Int64Value(2), Timeouts: mongodbUserTestTimeouts(),
	}
	if diags := userToState(&mongodb.User{Name: "alice", ClusterId: "cluster-id"}, &state); diags.HasError() {
		t.Fatalf("userToState diagnostics: %#v", diags)
	}
	if !state.PasswordWo.IsNull() || state.Password.ValueString() != "legacy-password" || state.PasswordWoVersion.ValueInt64() != 2 {
		t.Fatalf("password state was not preserved safely: %#v", state)
	}

	var resp frameworkresource.SchemaResponse
	NewResource().Schema(context.Background(), frameworkresource.SchemaRequest{}, &resp)
	tfState := tfsdk.State{Schema: resp.Schema}
	if diags := tfState.Set(context.Background(), state); diags.HasError() {
		t.Fatalf("resource state diagnostics: %#v", diags)
	}
}

func TestMongoDBUserDataSourceState(t *testing.T) {
	ctx := context.Background()
	var resp frameworkdatasource.SchemaResponse
	NewDataSource().Schema(ctx, frameworkdatasource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %#v", resp.Diagnostics)
	}
	if _, ok := resp.Schema.Attributes["password_wo"]; ok {
		t.Fatal("data source schema unexpectedly contains password_wo")
	}
	if _, ok := resp.Schema.Attributes["password_wo_version"]; ok {
		t.Fatal("data source schema unexpectedly contains password_wo_version")
	}

	state := dataSourceUser{Timeouts: mongodbUserTestTimeouts()}
	if diags := dataSourceUserToState(&mongodb.User{Name: "alice", ClusterId: "cluster-id"}, &state); diags.HasError() {
		t.Fatalf("dataSourceUserToState diagnostics: %#v", diags)
	}
	tfState := tfsdk.State{Schema: resp.Schema}
	if diags := tfState.Set(ctx, state); diags.HasError() {
		t.Fatalf("data source state diagnostics: %#v", diags)
	}
	config := tfsdk.Config{Schema: resp.Schema, Raw: tfState.Raw}
	var actual dataSourceUser
	if diags := config.Get(ctx, &actual); diags.HasError() {
		t.Fatalf("data source config diagnostics: %#v", diags)
	}
	want := dataSourceUser{
		Name: types.StringValue("alice"), ClusterID: types.StringValue("cluster-id"), AuthType: types.StringValue(authTypePassword),
		Permission: types.SetValueMust(permissionType, []attr.Value{}), Timeouts: mongodbUserTestTimeouts(),
	}
	if !reflect.DeepEqual(want, actual) {
		t.Fatalf("data source state round trip = %#v, want %#v", actual, want)
	}
}

func mongodbUserTestTimeouts() timeouts.Value {
	return timeouts.Value{Object: types.ObjectNull(map[string]attr.Type{
		"create": types.StringType, "update": types.StringType, "delete": types.StringType,
	})}
}
