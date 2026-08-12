package yandex

import (
	"slices"
	"testing"

	"github.com/hashicorp/go-cty/cty"
)

type testGreenplumClusterRawConfig struct {
	value   cty.Value
	values  map[string]interface{}
	changes map[string]bool
}

func (c testGreenplumClusterRawConfig) GetRawConfig() cty.Value {
	return c.value
}

func (c testGreenplumClusterRawConfig) GetOk(key string) (interface{}, bool) {
	value, ok := c.values[key]
	return value, ok
}

func (c testGreenplumClusterRawConfig) HasChange(key string) bool {
	return c.changes[key]
}

func TestMDBGreenplumClusterPasswordWoSchema(t *testing.T) {
	resource := resourceYandexMDBGreenplumCluster()
	if err := resource.InternalValidate(nil, true); err != nil {
		t.Fatalf("resource schema validation: %v", err)
	}

	legacyPassword := resource.Schema["user_password"]
	if !legacyPassword.Optional || legacyPassword.Required || !legacyPassword.Sensitive || len(legacyPassword.AtLeastOneOf) != 2 {
		t.Fatal("user_password must be optional, sensitive, and require one password form")
	}

	writeOnlyPassword := resource.Schema["user_password_wo"]
	if !writeOnlyPassword.Optional || !writeOnlyPassword.WriteOnly || !writeOnlyPassword.Sensitive || len(writeOnlyPassword.RequiredWith) != 1 {
		t.Fatal("user_password_wo must be optional, write-only, sensitive, and require its version")
	}

	version := resource.Schema["user_password_wo_version"]
	if !version.Optional || version.WriteOnly || len(version.RequiredWith) != 1 {
		t.Fatal("user_password_wo_version must be an optional state attribute requiring the password")
	}
}

func TestValidateGreenplumClusterPasswordConflict(t *testing.T) {
	tests := []struct {
		name      string
		config    map[string]cty.Value
		wantError bool
	}{
		{
			name: "legacy and write-only passwords conflict",
			config: map[string]cty.Value{
				"user_password":    cty.StringVal("legacy-password"),
				"user_password_wo": cty.StringVal("write-only-password"),
			},
			wantError: true,
		},
		{name: "legacy password", config: map[string]cty.Value{"user_password": cty.StringVal("legacy-password")}},
		{name: "write-only password", config: map[string]cty.Value{"user_password_wo": cty.StringVal("write-only-password")}},
		{
			name: "null legacy password is absent",
			config: map[string]cty.Value{
				"user_password":    cty.NullVal(cty.String),
				"user_password_wo": cty.StringVal("write-only-password"),
			},
		},
		{
			name: "unknown legacy password is absent",
			config: map[string]cty.Value{
				"user_password":    cty.UnknownVal(cty.String),
				"user_password_wo": cty.StringVal("write-only-password"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := testGreenplumClusterRawConfig{value: cty.ObjectVal(tt.config)}
			err := validateGreenplumClusterPasswordConflict(raw)
			if gotError := err != nil; gotError != tt.wantError {
				t.Fatalf("validateGreenplumClusterPasswordConflict() error = %v, wantError %t", err, tt.wantError)
			}
		})
	}
}

func TestValidateGreenplumClusterPasswordPair(t *testing.T) {
	tests := []struct {
		name      string
		config    map[string]cty.Value
		wantError bool
	}{
		{
			name: "write-only password with version",
			config: map[string]cty.Value{
				"user_password_wo":         cty.StringVal("write-only-password"),
				"user_password_wo_version": cty.NumberIntVal(1),
			},
		},
		{name: "write-only password without version", config: map[string]cty.Value{"user_password_wo": cty.StringVal("write-only-password")}, wantError: true},
		{name: "version without write-only password", config: map[string]cty.Value{"user_password_wo_version": cty.NumberIntVal(1)}, wantError: true},
		{
			name: "null pair is absent",
			config: map[string]cty.Value{
				"user_password_wo":         cty.NullVal(cty.String),
				"user_password_wo_version": cty.NullVal(cty.Number),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := testGreenplumClusterRawConfig{value: cty.ObjectVal(tt.config)}
			err := validateGreenplumClusterPasswordPair(raw)
			if gotError := err != nil; gotError != tt.wantError {
				t.Fatalf("validateGreenplumClusterPasswordPair() error = %v, wantError %t", err, tt.wantError)
			}
		})
	}
}

func TestGreenplumClusterPasswordForUpdate(t *testing.T) {
	config := testGreenplumClusterRawConfig{
		value: cty.ObjectVal(map[string]cty.Value{
			"user_password_wo": cty.StringVal("write-only-password"),
		}),
		changes: map[string]bool{"description": true},
	}
	if password := greenplumClusterPasswordForUpdate(config); password != "" {
		t.Fatalf("unrelated update password = %q, want empty", password)
	}

	config.changes["user_password_wo_version"] = true
	if password := greenplumClusterPasswordForUpdate(config); password != "write-only-password" {
		t.Fatalf("password version update password = %q, want write-only-password", password)
	}
}

func TestExpandGreenplumConfigUpdateMask(t *testing.T) {
	const logStatementPath = "config_spec.dbms_config.log_statement"

	tests := []struct {
		name    string
		changes map[string]bool
		want    []string
	}{
		{
			name:    "state defaults are omitted",
			changes: map[string]bool{"user_password_wo_version": true},
		},
		{
			name:    "changed user setting is preserved",
			changes: map[string]bool{"greenplum_config.log_statement": true},
			want:    []string{logStatementPath},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandGreenplumConfigUpdateMask(testGreenplumClusterRawConfig{changes: tt.changes}, []string{
				"gp_autostats_mode",
				"log_error_verbosity",
				"log_min_messages",
				"log_statement",
			})

			if !slices.Equal(got.GetPaths(), tt.want) {
				t.Fatalf("DBMS config mask = %v, want %v", got.GetPaths(), tt.want)
			}
		})
	}
}
