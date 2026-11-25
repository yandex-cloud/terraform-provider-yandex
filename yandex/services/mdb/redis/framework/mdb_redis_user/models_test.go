package mdb_redis_user

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/redis/v1"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func testPermissionsObject(t *testing.T, permissions Permissions) types.Object {
	t.Helper()

	value, diagnostics := types.ObjectValueFrom(context.Background(), permissionType.AttrTypes, permissions)
	if diagnostics.HasError() {
		t.Fatalf("permissions object diagnostics: %#v", diagnostics)
	}
	return value
}

func testPermissions(sanitizePayload string) Permissions {
	return Permissions{
		Commands:        types.StringValue("+ping"),
		Categories:      types.StringValue("+@all"),
		Patterns:        types.StringValue("~*"),
		PubSubChannels:  types.StringValue("&*"),
		SanitizePayload: types.StringValue(sanitizePayload),
		Databases:       types.StringValue("alldbs"),
	}
}

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
		Name:       "alice",
		ClusterId:  "cluster-id",
		Enabled:    true,
		AclOptions: "+@all",
		Permissions: &redis.Permissions{
			SanitizePayload: wrapperspb.String("skip-sanitize-payload"),
		},
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

	var permissions Permissions
	diagnostics = state.Permissions.As(context.Background(), &permissions, basetypes.ObjectAsOptions{})
	if diagnostics.HasError() {
		t.Fatalf("permissions diagnostics: %#v", diagnostics)
	}
	if got := permissions.SanitizePayload.ValueString(); got != "sanitize-payload" {
		t.Fatalf("sanitize_payload = %q, want canonical value", got)
	}
}

func TestUserToStatePreservesSanitizePayload(t *testing.T) {
	state := User{
		Permissions: testPermissionsObject(t, testPermissions("skip-sanitize-payload")),
	}
	user := &redis.User{
		Permissions: &redis.Permissions{
			Commands:        wrapperspb.String("+get"),
			SanitizePayload: wrapperspb.String("sanitize-payload"),
		},
	}

	diagnostics := userToState(context.Background(), user, &state)
	if diagnostics.HasError() {
		t.Fatalf("userToState() diagnostics: %#v", diagnostics)
	}

	var permissions Permissions
	diagnostics = state.Permissions.As(context.Background(), &permissions, basetypes.ObjectAsOptions{})
	if diagnostics.HasError() {
		t.Fatalf("permissions diagnostics: %#v", diagnostics)
	}
	if got := permissions.SanitizePayload.ValueString(); got != "skip-sanitize-payload" {
		t.Fatalf("sanitize_payload = %q, want preserved state value", got)
	}
	if got := permissions.Commands.ValueString(); got != "+get" {
		t.Fatalf("commands = %q, want API value", got)
	}
}

func TestUserToStateUsesCanonicalSanitizePayloadWithoutState(t *testing.T) {
	tests := map[string]types.Object{
		"zero":    {},
		"null":    types.ObjectNull(permissionType.AttrTypes),
		"unknown": types.ObjectUnknown(permissionType.AttrTypes),
	}

	for name, current := range tests {
		t.Run(name, func(t *testing.T) {
			state := User{Permissions: current}
			user := &redis.User{
				Permissions: &redis.Permissions{
					SanitizePayload: wrapperspb.String("skip-sanitize-payload"),
				},
			}

			diagnostics := userToState(context.Background(), user, &state)
			if diagnostics.HasError() {
				t.Fatalf("userToState() diagnostics: %#v", diagnostics)
			}

			var permissions Permissions
			diagnostics = state.Permissions.As(context.Background(), &permissions, basetypes.ObjectAsOptions{})
			if diagnostics.HasError() {
				t.Fatalf("permissions diagnostics: %#v", diagnostics)
			}
			if got := permissions.SanitizePayload.ValueString(); got != "sanitize-payload" {
				t.Fatalf("sanitize_payload = %q, want canonical value", got)
			}
		})
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
