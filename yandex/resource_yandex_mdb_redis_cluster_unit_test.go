package yandex

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type testRedisClusterRawConfig struct {
	value cty.Value
}

func (c testRedisClusterRawConfig) GetRawConfig() cty.Value {
	return c.value
}

func TestMDBRedisClusterPasswordWoSchema(t *testing.T) {
	resource := resourceYandexMDBRedisCluster()
	if err := resource.InternalValidate(nil, true); err != nil {
		t.Fatalf("resource schema validation: %v", err)
	}
	configSchema := resource.Schema["config"].Elem.(*schema.Resource).Schema

	legacyPassword := configSchema["password"]
	if !legacyPassword.Optional || legacyPassword.Required || !legacyPassword.Sensitive {
		t.Fatal("config.password must remain optional and sensitive")
	}

	writeOnlyPassword := configSchema["password_wo"]
	if !writeOnlyPassword.Optional || !writeOnlyPassword.WriteOnly || !writeOnlyPassword.Sensitive {
		t.Fatal("config.password_wo must be optional, write-only, and sensitive")
	}
	if len(writeOnlyPassword.RequiredWith) != 1 {
		t.Fatalf("config.password_wo RequiredWith = %v, want password_wo_version", writeOnlyPassword.RequiredWith)
	}

	version := configSchema["password_wo_version"]
	if !version.Optional || version.WriteOnly {
		t.Fatal("config.password_wo_version must be an optional state attribute")
	}
	if len(version.RequiredWith) != 1 {
		t.Fatalf("config.password_wo_version RequiredWith = %v, want password_wo", version.RequiredWith)
	}
}

func TestFlattenRedisConfigStatePasswords(t *testing.T) {
	writeOnlyState := flattenRedisConfigState(redisConfig{}, "", 2, true, nil)
	if _, ok := writeOnlyState["password"]; ok {
		t.Fatalf("write-only state contains legacy password: %#v", writeOnlyState["password"])
	}
	if got := writeOnlyState["password_wo_version"]; got != 2 {
		t.Fatalf("password_wo_version = %#v, want 2", got)
	}

	legacyState := flattenRedisConfigState(redisConfig{}, "legacy-password", nil, false, nil)
	if got := legacyState["password"]; got != "legacy-password" {
		t.Fatalf("password = %#v, want legacy-password", got)
	}
	if _, ok := legacyState["password_wo_version"]; ok {
		t.Fatalf("legacy state contains write-only password version: %#v", legacyState["password_wo_version"])
	}
}

func TestValidateRedisClusterPasswordConflict(t *testing.T) {
	tests := []struct {
		name      string
		config    map[string]cty.Value
		wantError bool
	}{
		{
			name: "legacy and write-only passwords conflict",
			config: map[string]cty.Value{
				"password":    cty.StringVal("legacy-password"),
				"password_wo": cty.StringVal("write-only-password"),
			},
			wantError: true,
		},
		{name: "legacy password", config: map[string]cty.Value{"password": cty.StringVal("legacy-password")}},
		{name: "write-only password", config: map[string]cty.Value{"password_wo": cty.StringVal("write-only-password")}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := testRedisClusterRawConfig{value: redisClusterRawConfig(tt.config)}
			err := validateRedisClusterPasswordConflict(raw)
			if gotError := err != nil; gotError != tt.wantError {
				t.Fatalf("validateRedisClusterPasswordConflict() error = %v, wantError %t", err, tt.wantError)
			}
		})
	}
}

func TestValidateRedisClusterPasswordPair(t *testing.T) {
	tests := []struct {
		name      string
		config    map[string]cty.Value
		wantError bool
	}{
		{
			name: "write-only password with version",
			config: map[string]cty.Value{
				"password_wo":         cty.StringVal("write-only-password"),
				"password_wo_version": cty.NumberIntVal(1),
			},
		},
		{name: "write-only password without version", config: map[string]cty.Value{"password_wo": cty.StringVal("write-only-password")}, wantError: true},
		{name: "version without write-only password", config: map[string]cty.Value{"password_wo_version": cty.NumberIntVal(1)}, wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := testRedisClusterRawConfig{value: redisClusterRawConfig(tt.config)}
			err := validateRedisClusterPasswordPair(raw)
			if gotError := err != nil; gotError != tt.wantError {
				t.Fatalf("validateRedisClusterPasswordPair() error = %v, wantError %t", err, tt.wantError)
			}
		})
	}
}

func redisClusterRawConfig(config map[string]cty.Value) cty.Value {
	return cty.ObjectVal(map[string]cty.Value{
		"config": cty.TupleVal([]cty.Value{cty.ObjectVal(config)}),
	})
}
