package mdb_clickhouse_user

import (
	"context"
	"testing"

	frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/chcommon/usersettings"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestValidateAuthConfiguration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		authMethod       clickhouse.AuthMethod
		password         string
		generatePassword bool
		wantErr          bool
	}{
		{
			name:       "password auth with password succeeds",
			authMethod: clickhouse.AuthMethod_AUTH_METHOD_PASSWORD,
			password:   "secret",
		},
		{
			name:             "password auth with generated password succeeds",
			authMethod:       clickhouse.AuthMethod_AUTH_METHOD_PASSWORD,
			generatePassword: true,
		},
		{
			name:       "password auth with neither fails",
			authMethod: clickhouse.AuthMethod_AUTH_METHOD_PASSWORD,
			wantErr:    true,
		},
		{
			name:             "password auth with both fails",
			authMethod:       clickhouse.AuthMethod_AUTH_METHOD_PASSWORD,
			password:         "secret",
			generatePassword: true,
			wantErr:          true,
		},
		{
			name:       "iam auth with neither succeeds",
			authMethod: clickhouse.AuthMethod_AUTH_METHOD_IAM,
		},
		{
			name:       "iam auth with password fails",
			authMethod: clickhouse.AuthMethod_AUTH_METHOD_IAM,
			password:   "secret",
			wantErr:    true,
		},
		{
			name:             "iam auth with generated password fails",
			authMethod:       clickhouse.AuthMethod_AUTH_METHOD_IAM,
			generatePassword: true,
			wantErr:          true,
		},
		{
			name:             "iam auth with both fails",
			authMethod:       clickhouse.AuthMethod_AUTH_METHOD_IAM,
			password:         "secret",
			generatePassword: true,
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateAuthConfiguration(&clickhouse.UserSpec{
				Password:         tt.password,
				GeneratePassword: wrapperspb.Bool(tt.generatePassword),
				AuthMethod:       tt.authMethod,
			})

			if (err != nil) != tt.wantErr {
				t.Fatalf("validateAuthConfiguration() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetUpdatePathsAuthTransitions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		state           ResourceUser
		plan            ResourceUser
		passwordChanged bool
		want            []string
	}{
		{
			name: "password to iam updates only auth method",
			state: testResourceUserAuth(
				clickhouse.AuthMethod_AUTH_METHOD_PASSWORD,
				types.StringValue("old-secret"),
				false,
			),
			plan: testResourceUserAuth(
				clickhouse.AuthMethod_AUTH_METHOD_IAM,
				types.StringNull(),
				false,
			),
			want: []string{"auth_method"},
		},
		{
			name: "generated password to iam updates only auth method",
			state: testResourceUserAuth(
				clickhouse.AuthMethod_AUTH_METHOD_PASSWORD,
				types.StringNull(),
				true,
			),
			plan: testResourceUserAuth(
				clickhouse.AuthMethod_AUTH_METHOD_IAM,
				types.StringNull(),
				false,
			),
			want: []string{"auth_method"},
		},
		{
			name: "iam to password updates auth method and password",
			state: testResourceUserAuth(
				clickhouse.AuthMethod_AUTH_METHOD_IAM,
				types.StringNull(),
				false,
			),
			plan: testResourceUserAuth(
				clickhouse.AuthMethod_AUTH_METHOD_PASSWORD,
				types.StringValue("new-secret"),
				false,
			),
			passwordChanged: true,
			want:            []string{"auth_method", "password"},
		},
		{
			name: "iam to generated password updates auth method and generate password",
			state: testResourceUserAuth(
				clickhouse.AuthMethod_AUTH_METHOD_IAM,
				types.StringNull(),
				false,
			),
			plan: testResourceUserAuth(
				clickhouse.AuthMethod_AUTH_METHOD_PASSWORD,
				types.StringNull(),
				true,
			),
			want: []string{"auth_method", "generate_password"},
		},
		{
			name: "password to generated password does not update empty password",
			state: testResourceUserAuth(
				clickhouse.AuthMethod_AUTH_METHOD_PASSWORD,
				types.StringValue("old-secret"),
				false,
			),
			plan: testResourceUserAuth(
				clickhouse.AuthMethod_AUTH_METHOD_PASSWORD,
				types.StringNull(),
				true,
			),
			want: []string{"generate_password"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := getUpdatePaths(&tt.plan, &tt.state, tt.passwordChanged)
			assertSamePaths(t, got, tt.want)
		})
	}
}

func TestClickHouseUserPasswordWoSchema(t *testing.T) {
	t.Parallel()

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

func TestClickHouseUserPasswordForCreate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		plan       ResourceUser
		passwordWo types.String
		want       string
	}{
		{
			name: "write-only password takes precedence",
			plan: ResourceUser{
				Password: types.StringValue("legacy-password"),
			},
			passwordWo: types.StringValue("write-only-password"),
			want:       "write-only-password",
		},
		{
			name: "legacy password is used without write-only password",
			plan: ResourceUser{
				Password: types.StringValue("legacy-password"),
			},
			passwordWo: types.StringNull(),
			want:       "legacy-password",
		},
		{
			name: "stale write-only password version is not a password",
			plan: ResourceUser{
				Password:          types.StringNull(),
				PasswordWoVersion: types.Int64Value(1),
			},
			passwordWo: types.StringNull(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := clickHouseUserPasswordForCreate(&tt.plan, tt.passwordWo); got != tt.want {
				t.Fatalf("clickHouseUserPasswordForCreate() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestClickHouseUserPasswordChange(t *testing.T) {
	t.Parallel()

	state := ResourceUser{
		Password:          types.StringNull(),
		PasswordWoVersion: types.Int64Value(1),
	}
	plan := state

	password, changed, diags := clickHouseUserPasswordChange(&plan, &state, types.StringValue("ignored-password"))
	if diags.HasError() || changed || password != "" {
		t.Fatalf("unchanged password = %q, changed = %t, diagnostics = %#v", password, changed, diags)
	}

	plan.PasswordWoVersion = types.Int64Value(2)
	password, changed, diags = clickHouseUserPasswordChange(&plan, &state, types.StringValue("rotated-password"))
	if diags.HasError() || !changed || password != "rotated-password" {
		t.Fatalf("rotated password = %q, changed = %t, diagnostics = %#v", password, changed, diags)
	}

	_, _, diags = clickHouseUserPasswordChange(&plan, &state, types.StringNull())
	if !diags.HasError() {
		t.Fatal("missing write-only password must produce an error when the version changes")
	}
}

func TestValidateAuthConfigurationWithWriteOnlyPassword(t *testing.T) {
	t.Parallel()

	err := validateAuthConfigurationWithPassword(&clickhouse.UserSpec{
		GeneratePassword: wrapperspb.Bool(false),
		AuthMethod:       clickhouse.AuthMethod_AUTH_METHOD_PASSWORD,
	}, true)
	if err != nil {
		t.Fatalf("write-only password auth validation failed: %v", err)
	}
}

func TestGetUpdatePathsPasswordWoVersionChange(t *testing.T) {
	t.Parallel()

	state := testResourceUserAuth(clickhouse.AuthMethod_AUTH_METHOD_PASSWORD, types.StringNull(), false)
	plan := state
	paths := getUpdatePaths(&plan, &state, true)
	assertSamePaths(t, paths, []string{"password"})
}

func TestClickHouseUserWriteOnlyPasswordState(t *testing.T) {
	t.Parallel()

	state := testResourceUserAuth(clickhouse.AuthMethod_AUTH_METHOD_PASSWORD, types.StringNull(), false)
	state.PasswordWo = types.StringValue("must-not-survive")
	state.PasswordWoVersion = types.Int64Value(2)
	diags := userToState(context.Background(), &clickhouse.User{
		Name:       "alice",
		ClusterId:  "cluster-id",
		Settings:   &clickhouse.UserSettings{},
		AuthMethod: clickhouse.AuthMethod_AUTH_METHOD_PASSWORD,
	}, &state)
	if diags.HasError() {
		t.Fatalf("userToState() diagnostics: %#v", diags)
	}
	if !state.PasswordWo.IsNull() || state.PasswordWoVersion.ValueInt64() != 2 {
		t.Fatalf("write-only password state was not preserved safely: %#v", state)
	}
}

func TestGetAuthMethodNameNormalizesUnspecified(t *testing.T) {
	t.Parallel()

	got := getAuthMethodName(clickhouse.AuthMethod_AUTH_METHOD_UNSPECIFIED)
	if got.ValueString() != defaultUserAuthMethod {
		t.Fatalf("getAuthMethodName() = %q, want %q", got.ValueString(), defaultUserAuthMethod)
	}
}

func TestAuthMethodUsesTerraformNames(t *testing.T) {
	t.Parallel()

	if got := getAuthMethodName(clickhouse.AuthMethod_AUTH_METHOD_IAM); got.ValueString() != "iam" {
		t.Fatalf("getAuthMethodName() = %q, want %q", got.ValueString(), "iam")
	}

	if got := getAuthMethodValue(types.StringValue("iam")); got != clickhouse.AuthMethod_AUTH_METHOD_IAM {
		t.Fatalf("getAuthMethodValue() = %s, want %s", got.String(), clickhouse.AuthMethod_AUTH_METHOD_IAM.String())
	}
}

func TestDatasourceUserSupportsAuthMethodFlattening(t *testing.T) {
	t.Parallel()

	ds := &DatasourceUser{}
	var state interface{ SetAuthMethod(types.String) } = ds
	authMethod := "iam"

	state.SetAuthMethod(types.StringValue(authMethod))

	if ds.AuthMethod.ValueString() != authMethod {
		t.Fatalf("auth method = %q, want %q", ds.AuthMethod.ValueString(), authMethod)
	}
}

func testResourceUserAuth(authMethod clickhouse.AuthMethod, password types.String, generatePassword bool) ResourceUser {
	return ResourceUser{
		AuthMethod:        getAuthMethodName(authMethod),
		Password:          password,
		GeneratePassword:  types.BoolValue(generatePassword),
		Permissions:       types.SetNull(permissionType),
		Settings:          types.ObjectNull(usersettings.AttrTypes),
		Quotas:            types.SetNull(quotaType),
		ConnectionManager: types.ObjectNull(connectionManagerType),
	}
}

func assertSamePaths(t *testing.T, got, want []string) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("paths = %v, want %v", got, want)
	}

	counts := make(map[string]int, len(got))
	for _, path := range got {
		counts[path]++
	}

	for _, path := range want {
		counts[path]--
		if counts[path] < 0 {
			t.Fatalf("paths = %v, want %v", got, want)
		}
	}
}
