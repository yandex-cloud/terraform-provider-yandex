package mdb_redis_user

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/redis/v1"
)

func TestDataSourceUserModelMatchesSchema(t *testing.T) {
	var response datasource.SchemaResponse
	NewDataSource().Schema(context.Background(), datasource.SchemaRequest{}, &response)
	if response.Diagnostics.HasError() {
		t.Fatalf("data source schema diagnostics: %#v", response.Diagnostics)
	}

	modelType := reflect.TypeOf(dataSourceUser{})
	modelAttributes := make(map[string]struct{}, modelType.NumField())
	for i := 0; i < modelType.NumField(); i++ {
		modelAttributes[modelType.Field(i).Tag.Get("tfsdk")] = struct{}{}
	}

	if len(modelAttributes) != len(response.Schema.Attributes) {
		t.Fatalf("data source model attributes = %v, schema attributes = %v", modelAttributes, response.Schema.Attributes)
	}
	for name := range response.Schema.Attributes {
		if _, ok := modelAttributes[name]; !ok {
			t.Fatalf("data source schema attribute %q is missing from model", name)
		}
	}
}

func TestDataSourceUserToState(t *testing.T) {
	state := dataSourceUser{}
	user := &redis.User{
		Name:        "alice",
		ClusterId:   "cluster-id",
		Enabled:     true,
		AclOptions:  "+@all",
		Permissions: &redis.Permissions{},
	}

	diagnostics := dataSourceUserToState(context.Background(), user, &state)
	if diagnostics.HasError() {
		t.Fatalf("dataSourceUserToState() diagnostics: %#v", diagnostics)
	}
	if state.Name.ValueString() != "alice" || state.ClusterID.ValueString() != "cluster-id" {
		t.Fatalf("data source identity = (%q, %q), want (alice, cluster-id)", state.Name.ValueString(), state.ClusterID.ValueString())
	}
	if !state.Passwords.Equal(types.SetNull(types.StringType)) {
		t.Fatalf("passwords = %#v, want null", state.Passwords)
	}
}

func TestUserFromStatePasswordWo(t *testing.T) {
	state := &User{
		Name:        types.StringValue("alice"),
		Passwords:   types.SetNull(types.StringType),
		Enabled:     types.BoolValue(true),
		Permissions: types.ObjectNull(permissionType.AttributeTypes()),
	}

	user, diagnostics := userFromState(context.Background(), state, types.StringValue("write-only-password"))
	if diagnostics.HasError() {
		t.Fatalf("userFromState() diagnostics: %#v", diagnostics)
	}
	if len(user.Passwords) != 1 || user.Passwords[0] != "write-only-password" {
		t.Fatalf("passwords = %v, want write-only password", user.Passwords)
	}
}
