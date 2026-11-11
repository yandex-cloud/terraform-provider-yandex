package mdb_redis_cluster_v2

import (
	"context"
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

	configAttribute, ok := resp.Schema.Attributes["config"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatalf("config type = %T, want schema.SingleNestedAttribute", resp.Schema.Attributes["config"])
	}

	legacyPassword := configAttribute.Attributes["password"].(schema.StringAttribute)
	if !legacyPassword.IsOptional() || legacyPassword.IsRequired() || len(legacyPassword.Validators) != 1 {
		t.Fatal("config.password must remain optional and require exactly one password source")
	}

	writeOnlyPassword := configAttribute.Attributes["password_wo"].(schema.StringAttribute)
	if !writeOnlyPassword.IsOptional() || !writeOnlyPassword.IsWriteOnly() || !writeOnlyPassword.IsSensitive() || len(writeOnlyPassword.Validators) != 1 {
		t.Fatal("config.password_wo must be optional, write-only, sensitive, and require its version")
	}

	version := configAttribute.Attributes["password_wo_version"].(schema.Int64Attribute)
	if !version.IsOptional() || version.IsWriteOnly() || len(version.Validators) != 1 {
		t.Fatal("config.password_wo_version must be an optional state attribute requiring password_wo")
	}
}

func TestRedisClusterCreatePassword(t *testing.T) {
	config := &Config{configModel: configModel{Password: types.StringValue("legacy-password")}}

	if got := redisClusterCreatePassword(config, types.StringNull()); got != "legacy-password" {
		t.Fatalf("legacy password = %q, want %q", got, "legacy-password")
	}
	if got := redisClusterCreatePassword(config, types.StringValue("write-only-password")); got != "write-only-password" {
		t.Fatalf("write-only password = %q, want %q", got, "write-only-password")
	}
}

func TestRedisClusterPasswordChange(t *testing.T) {
	t.Run("version change rotates write-only password", func(t *testing.T) {
		state := &Config{configModel: configModel{Password: types.StringNull()}, PasswordWoVersion: types.Int64Value(1)}
		plan := &Config{configModel: configModel{Password: types.StringNull()}, PasswordWoVersion: types.Int64Value(2)}

		password, changed, diags := redisClusterPasswordChange(plan, state, types.StringValue("rotated-password"))

		if diags.HasError() || !changed || password != "rotated-password" {
			t.Fatalf("password change = (%q, %t, %#v), want rotated password", password, changed, diags)
		}
	})

	t.Run("same version ignores write-only password", func(t *testing.T) {
		state := &Config{configModel: configModel{Password: types.StringNull()}, PasswordWoVersion: types.Int64Value(1)}
		plan := &Config{configModel: configModel{Password: types.StringNull()}, PasswordWoVersion: types.Int64Value(1)}

		password, changed, diags := redisClusterPasswordChange(plan, state, types.StringValue("different-password"))

		if diags.HasError() || changed || password != "" {
			t.Fatalf("password change = (%q, %t, %#v), want no change", password, changed, diags)
		}
	})

	t.Run("version change requires write-only password", func(t *testing.T) {
		state := &Config{configModel: configModel{Password: types.StringNull()}, PasswordWoVersion: types.Int64Value(1)}
		plan := &Config{configModel: configModel{Password: types.StringNull()}, PasswordWoVersion: types.Int64Value(2)}

		_, _, diags := redisClusterPasswordChange(plan, state, types.StringNull())
		if !diags.HasError() {
			t.Fatal("redisClusterPasswordChange() diagnostics has no error")
		}
	})
}
