package yandex

import (
	"strings"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	clickhouseConfig "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1/config"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
)

type testClickHouseClusterRawConfig struct {
	value   cty.Value
	values  map[string]any
	changes map[string]bool
}

func (c testClickHouseClusterRawConfig) GetRawConfig() cty.Value {
	return c.value
}

func (c testClickHouseClusterRawConfig) GetOk(key string) (any, bool) {
	value, ok := c.values[key]
	return value, ok
}

func (c testClickHouseClusterRawConfig) HasChange(key string) bool {
	return c.changes[key]
}

func TestMDBClickHouseClusterAdminPasswordWoSchema(t *testing.T) {
	resource := resourceYandexMDBClickHouseCluster()
	if err := resource.InternalValidate(nil, true); err != nil {
		t.Fatalf("resource schema validation: %v", err)
	}

	legacyPassword := resource.Schema["admin_password"]
	if !legacyPassword.Optional || legacyPassword.Required || !legacyPassword.Sensitive {
		t.Fatal("admin_password must be optional and sensitive")
	}

	writeOnlyPassword := resource.Schema["admin_password_wo"]
	if !writeOnlyPassword.Optional || !writeOnlyPassword.WriteOnly || !writeOnlyPassword.Sensitive || len(writeOnlyPassword.RequiredWith) != 1 {
		t.Fatal("admin_password_wo must be optional, write-only, sensitive, and require its version")
	}

	version := resource.Schema["admin_password_wo_version"]
	if !version.Optional || version.WriteOnly || len(version.RequiredWith) != 1 {
		t.Fatal("admin_password_wo_version must be an optional state attribute requiring the password")
	}
}

func TestMDBClickHouseClusterDataSourceExcludesAdminPasswordWo(t *testing.T) {
	dataSource := dataSourceYandexMDBClickHouseCluster()
	if _, ok := dataSource.Schema["admin_password_wo"]; ok {
		t.Fatal("ClickHouse cluster data source must not contain admin_password_wo")
	}
	if _, ok := dataSource.Schema["admin_password_wo_version"]; ok {
		t.Fatal("ClickHouse cluster data source must not contain admin_password_wo_version")
	}
	if err := dataSource.InternalValidate(nil, false); err != nil {
		t.Fatalf("data source schema validation: %v", err)
	}
}

func TestValidateClickHouseClusterAdminPasswordConflict(t *testing.T) {
	tests := []struct {
		name      string
		config    map[string]cty.Value
		wantError bool
	}{
		{
			name: "legacy and write-only passwords conflict",
			config: map[string]cty.Value{
				"admin_password":    cty.StringVal("legacy-password"),
				"admin_password_wo": cty.StringVal("write-only-password"),
			},
			wantError: true,
		},
		{name: "legacy password", config: map[string]cty.Value{"admin_password": cty.StringVal("legacy-password")}},
		{name: "write-only password", config: map[string]cty.Value{"admin_password_wo": cty.StringVal("write-only-password")}},
		{
			name: "null legacy password is absent",
			config: map[string]cty.Value{
				"admin_password":    cty.NullVal(cty.String),
				"admin_password_wo": cty.StringVal("write-only-password"),
			},
		},
		{
			name: "unknown legacy password is absent",
			config: map[string]cty.Value{
				"admin_password":    cty.UnknownVal(cty.String),
				"admin_password_wo": cty.StringVal("write-only-password"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := testClickHouseClusterRawConfig{value: cty.ObjectVal(tt.config)}
			err := validateClickHouseClusterAdminPasswordConflict(raw)
			if gotError := err != nil; gotError != tt.wantError {
				t.Fatalf("validateClickHouseClusterAdminPasswordConflict() error = %v, wantError %t", err, tt.wantError)
			}
		})
	}
}

func TestValidateClickHouseClusterAdminPasswordPair(t *testing.T) {
	tests := []struct {
		name      string
		config    map[string]cty.Value
		wantError bool
	}{
		{
			name: "write-only password with version",
			config: map[string]cty.Value{
				"admin_password_wo":         cty.StringVal("write-only-password"),
				"admin_password_wo_version": cty.NumberIntVal(1),
			},
		},
		{name: "write-only password without version", config: map[string]cty.Value{"admin_password_wo": cty.StringVal("write-only-password")}, wantError: true},
		{name: "version without write-only password", config: map[string]cty.Value{"admin_password_wo_version": cty.NumberIntVal(1)}, wantError: true},
		{
			name: "null pair is absent",
			config: map[string]cty.Value{
				"admin_password_wo":         cty.NullVal(cty.String),
				"admin_password_wo_version": cty.NullVal(cty.Number),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := testClickHouseClusterRawConfig{value: cty.ObjectVal(tt.config)}
			err := validateClickHouseClusterAdminPasswordPair(raw)
			if gotError := err != nil; gotError != tt.wantError {
				t.Fatalf("validateClickHouseClusterAdminPasswordPair() error = %v, wantError %t", err, tt.wantError)
			}
		})
	}
}

func TestClickHouseClusterAdminPassword(t *testing.T) {
	legacy := testClickHouseClusterRawConfig{
		value:  cty.EmptyObjectVal,
		values: map[string]any{"admin_password": "legacy-password"},
	}
	if password := clickHouseClusterAdminPassword(legacy); password != "legacy-password" {
		t.Fatalf("legacy password = %q", password)
	}

	writeOnly := testClickHouseClusterRawConfig{
		value: cty.ObjectVal(map[string]cty.Value{
			"admin_password_wo": cty.StringVal("write-only-password"),
		}),
		values: map[string]any{"admin_password": "legacy-password"},
	}
	if password := clickHouseClusterAdminPassword(writeOnly); password != "write-only-password" {
		t.Fatalf("write-only password = %q", password)
	}
}

func TestClickHouseClusterAdminPasswordForUpdate(t *testing.T) {
	config := testClickHouseClusterRawConfig{
		value: cty.ObjectVal(map[string]cty.Value{
			"admin_password_wo": cty.StringVal("write-only-password"),
		}),
		changes: map[string]bool{"description": true},
	}
	if password := clickHouseClusterAdminPasswordForUpdate(config); password != "" {
		t.Fatalf("unrelated update password = %q, want empty", password)
	}

	config.changes["admin_password_wo_version"] = true
	if password := clickHouseClusterAdminPasswordForUpdate(config); password != "write-only-password" {
		t.Fatalf("password version update password = %q, want write-only-password", password)
	}
}

func TestRedactClickHouseClusterUpdateRequest(t *testing.T) {
	secrets := []string{"admin-secret", "kafka-secret", "kafka-topic-secret", "rabbitmq-secret"}
	req := &clickhouse.UpdateClusterRequest{ConfigSpec: &clickhouse.ConfigSpec{
		AdminPassword: secrets[0],
		Clickhouse: &clickhouse.ConfigSpec_Clickhouse{Config: &clickhouseConfig.ClickhouseConfig{
			Kafka:       &clickhouseConfig.ClickhouseConfig_Kafka{SaslPassword: secrets[1]},
			KafkaTopics: []*clickhouseConfig.ClickhouseConfig_KafkaTopic{{Settings: &clickhouseConfig.ClickhouseConfig_Kafka{SaslPassword: secrets[2]}}},
			Rabbitmq:    &clickhouseConfig.ClickhouseConfig_Rabbitmq{Password: secrets[3]},
		}},
	}}

	redacted := redactClickHouseClusterUpdateRequest(req)
	originalConfig := req.ConfigSpec.GetClickhouse().GetConfig()
	redactedConfig := redacted.ConfigSpec.GetClickhouse().GetConfig()
	originalPasswords := []string{
		req.ConfigSpec.GetAdminPassword(),
		originalConfig.GetKafka().GetSaslPassword(),
		originalConfig.GetKafkaTopics()[0].GetSettings().GetSaslPassword(),
		originalConfig.GetRabbitmq().GetPassword(),
	}
	redactedPasswords := []string{
		redacted.ConfigSpec.GetAdminPassword(),
		redactedConfig.GetKafka().GetSaslPassword(),
		redactedConfig.GetKafkaTopics()[0].GetSettings().GetSaslPassword(),
		redactedConfig.GetRabbitmq().GetPassword(),
	}
	dump := validate.ProtoDump(redacted)
	for i, secret := range secrets {
		if originalPasswords[i] != secret || redactedPasswords[i] != redactedClickHousePassword {
			t.Fatalf("request redaction mutated the source or did not redact the clone: original=%q redacted=%q", originalPasswords[i], redactedPasswords[i])
		}
		if strings.Contains(dump, secret) {
			t.Fatalf("redacted request contains the password: %s", dump)
		}
	}
}
