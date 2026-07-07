package mdb_clickhouse_user

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
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
		name  string
		state ResourceUser
		plan  ResourceUser
		want  []string
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
			want: []string{"auth_method", "password"},
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

			got := getUpdatePaths(&tt.plan, &tt.state)
			assertSamePaths(t, got, tt.want)
		})
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
		Settings:          types.ObjectNull(settingsType),
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
