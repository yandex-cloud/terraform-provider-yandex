package mdb_redis_user

import (
	"context"
	"slices"
	"testing"

	frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestPasswordWoSchema(t *testing.T) {
	var resp frameworkresource.SchemaResponse
	NewResource().Schema(context.Background(), frameworkresource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %#v", resp.Diagnostics)
	}

	passwords := resp.Schema.Attributes["passwords"].(schema.SetAttribute)
	if !passwords.IsOptional() || passwords.IsRequired() || len(passwords.Validators) != 2 {
		t.Fatal("passwords must remain optional, size-limited, and mutually exclusive with password_wo")
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

func TestGetUpdatePathsPasswordWoVersion(t *testing.T) {
	state := User{
		Permissions:       types.ObjectNull(permissionType.AttributeTypes()),
		Enabled:           types.BoolValue(true),
		Passwords:         types.SetNull(types.StringType),
		PasswordWoVersion: types.Int64Value(1),
	}
	plan := state
	plan.PasswordWoVersion = types.Int64Value(2)

	paths := getUpdatePaths(context.Background(), nil, plan, state)
	if !slices.Contains(paths, "passwords") {
		t.Fatalf("update paths = %v, want passwords", paths)
	}
}
