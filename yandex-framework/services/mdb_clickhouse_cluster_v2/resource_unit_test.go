package mdb_clickhouse_cluster_v2

import (
	"context"
	"slices"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	clickhouseConfig "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1/config"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_clickhouse_cluster_v2/models"
	"google.golang.org/protobuf/proto"
)

func TestClickHouseClusterAdminPasswordWoSchema(t *testing.T) {
	var resp frameworkresource.SchemaResponse
	NewClickHouseClusterResourceV2().Schema(context.Background(), frameworkresource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %#v", resp.Diagnostics)
	}

	legacyPassword := resp.Schema.Attributes["admin_password"].(schema.StringAttribute)
	if !legacyPassword.IsOptional() || legacyPassword.IsRequired() || !legacyPassword.IsSensitive() || len(legacyPassword.Validators) != 1 {
		t.Fatal("admin_password must remain optional, sensitive, and conflict with admin_password_wo")
	}

	writeOnlyPassword := resp.Schema.Attributes["admin_password_wo"].(schema.StringAttribute)
	if !writeOnlyPassword.IsOptional() || !writeOnlyPassword.IsWriteOnly() || !writeOnlyPassword.IsSensitive() || len(writeOnlyPassword.Validators) != 1 {
		t.Fatal("admin_password_wo must be optional, write-only, sensitive, and require its version")
	}

	version := resp.Schema.Attributes["admin_password_wo_version"].(schema.Int64Attribute)
	if !version.IsOptional() || version.IsWriteOnly() || len(version.Validators) != 1 {
		t.Fatal("admin_password_wo_version must be an optional state attribute requiring admin_password_wo")
	}
}

func TestClickHouseClusterAdminPasswordForCreate(t *testing.T) {
	plan := &models.ClusterResource{AdminPassword: types.StringValue("legacy-password")}
	if got := clickHouseClusterAdminPasswordForCreate(plan, types.StringValue("write-only-password")); got != "write-only-password" {
		t.Fatalf("password = %q, want write-only-password", got)
	}
	if got := clickHouseClusterAdminPasswordForCreate(plan, types.StringNull()); got != "legacy-password" {
		t.Fatalf("password = %q, want legacy-password", got)
	}
}

func TestClickHouseClusterAdminPasswordChange(t *testing.T) {
	t.Run("version change rotates write-only password", func(t *testing.T) {
		state := &models.ClusterResource{AdminPassword: types.StringNull(), AdminPasswordWoVersion: types.Int64Value(1)}
		plan := &models.ClusterResource{AdminPassword: types.StringNull(), AdminPasswordWoVersion: types.Int64Value(2)}

		password, changed, diags := clickHouseClusterAdminPasswordChange(plan, state, types.StringValue("rotated-password"))
		if diags.HasError() || !changed || password != "rotated-password" {
			t.Fatalf("password change = (%q, %t, %#v), want rotated password", password, changed, diags)
		}
	})

	t.Run("same version ignores write-only password", func(t *testing.T) {
		state := &models.ClusterResource{AdminPassword: types.StringNull(), AdminPasswordWoVersion: types.Int64Value(1)}
		plan := &models.ClusterResource{AdminPassword: types.StringNull(), AdminPasswordWoVersion: types.Int64Value(1)}

		password, changed, diags := clickHouseClusterAdminPasswordChange(plan, state, types.StringValue("different-password"))
		if diags.HasError() || changed || password != "" {
			t.Fatalf("password change = (%q, %t, %#v), want no change", password, changed, diags)
		}
	})

	t.Run("version change requires write-only password", func(t *testing.T) {
		state := &models.ClusterResource{AdminPassword: types.StringNull(), AdminPasswordWoVersion: types.Int64Value(1)}
		plan := &models.ClusterResource{AdminPassword: types.StringNull(), AdminPasswordWoVersion: types.Int64Value(2)}

		_, _, diags := clickHouseClusterAdminPasswordChange(plan, state, types.StringNull())
		if !diags.HasError() {
			t.Fatal("clickHouseClusterAdminPasswordChange() diagnostics has no error")
		}
	})

	t.Run("legacy password change remains supported", func(t *testing.T) {
		state := &models.ClusterResource{AdminPassword: types.StringValue("old-password"), AdminPasswordWoVersion: types.Int64Null()}
		plan := &models.ClusterResource{AdminPassword: types.StringValue("new-password"), AdminPasswordWoVersion: types.Int64Null()}

		password, changed, diags := clickHouseClusterAdminPasswordChange(plan, state, types.StringNull())
		if diags.HasError() || !changed || password != "new-password" {
			t.Fatalf("password change = (%q, %t, %#v), want legacy password", password, changed, diags)
		}
	})
}

func TestClickHouseClusterAdminPasswordUpdatePath(t *testing.T) {
	state := &models.ClusterResource{}
	plan := &models.ClusterResource{}
	var diags diag.Diagnostics

	configSpec, paths := prepareClusterConfigSpec(context.Background(), plan, state, "rotated-password", true, &diags)
	if diags.HasError() {
		t.Fatalf("prepareClusterConfigSpec() diagnostics: %#v", diags)
	}
	if !slices.Contains(paths, "config_spec.admin_password") || configSpec.GetAdminPassword() != "rotated-password" {
		t.Fatalf("config password = %q, paths = %v", configSpec.GetAdminPassword(), paths)
	}
}

func TestClickHouseClusterAdminPasswordRedaction(t *testing.T) {
	createReq := &clickhouse.CreateClusterRequest{
		ConfigSpec: clickHouseConfigSpecWithPasswords("create"),
		UserSpecs:  []*clickhouse.UserSpec{{Password: "create-user-secret"}},
	}
	redactedCreate := redactClickHouseCreateClusterRequest(createReq)
	assertClickHouseConfigSpecPasswordsRedacted(t, createReq.ConfigSpec, redactedCreate.ConfigSpec, "create", redactedCreate)
	if createReq.UserSpecs[0].GetPassword() != "create-user-secret" || redactedCreate.UserSpecs[0].GetPassword() != redactedClickHousePassword {
		t.Fatal("create user password was not safely redacted")
	}
	if dump := validate.ProtoDump(redactedCreate); strings.Contains(dump, "create-user-secret") {
		t.Fatalf("redacted create request contains the user password: %s", dump)
	}

	updateReq := &clickhouse.UpdateClusterRequest{ConfigSpec: clickHouseConfigSpecWithPasswords("update")}
	redactedUpdate := redactClickHouseUpdateClusterRequest(updateReq)
	assertClickHouseConfigSpecPasswordsRedacted(t, updateReq.ConfigSpec, redactedUpdate.ConfigSpec, "update", redactedUpdate)

	restoreReq := &clickhouse.RestoreClusterRequest{ConfigSpec: clickHouseConfigSpecWithPasswords("restore")}
	redactedRestore := redactClickHouseRestoreClusterRequest(restoreReq)
	assertClickHouseConfigSpecPasswordsRedacted(t, restoreReq.ConfigSpec, redactedRestore.ConfigSpec, "restore", redactedRestore)
}

func clickHouseConfigSpecWithPasswords(prefix string) *clickhouse.ConfigSpec {
	return &clickhouse.ConfigSpec{
		AdminPassword: prefix + "-admin-secret",
		Clickhouse: &clickhouse.ConfigSpec_Clickhouse{Config: &clickhouseConfig.ClickhouseConfig{
			Kafka:       &clickhouseConfig.ClickhouseConfig_Kafka{SaslPassword: prefix + "-kafka-secret"},
			KafkaTopics: []*clickhouseConfig.ClickhouseConfig_KafkaTopic{{Settings: &clickhouseConfig.ClickhouseConfig_Kafka{SaslPassword: prefix + "-kafka-topic-secret"}}},
			Rabbitmq:    &clickhouseConfig.ClickhouseConfig_Rabbitmq{Password: prefix + "-rabbitmq-secret"},
		}},
	}
}

func assertClickHouseConfigSpecPasswordsRedacted(t *testing.T, original, redacted *clickhouse.ConfigSpec, prefix string, request proto.Message) {
	t.Helper()
	originalConfig := original.GetClickhouse().GetConfig()
	redactedConfig := redacted.GetClickhouse().GetConfig()
	originalPasswords := []string{
		original.GetAdminPassword(),
		originalConfig.GetKafka().GetSaslPassword(),
		originalConfig.GetKafkaTopics()[0].GetSettings().GetSaslPassword(),
		originalConfig.GetRabbitmq().GetPassword(),
	}
	redactedPasswords := []string{
		redacted.GetAdminPassword(),
		redactedConfig.GetKafka().GetSaslPassword(),
		redactedConfig.GetKafkaTopics()[0].GetSettings().GetSaslPassword(),
		redactedConfig.GetRabbitmq().GetPassword(),
	}
	secrets := []string{
		prefix + "-admin-secret",
		prefix + "-kafka-secret",
		prefix + "-kafka-topic-secret",
		prefix + "-rabbitmq-secret",
	}
	for i, secret := range secrets {
		if originalPasswords[i] != secret || redactedPasswords[i] != redactedClickHousePassword {
			t.Fatalf("request redaction mutated the source or did not redact the clone: original=%q redacted=%q", originalPasswords[i], redactedPasswords[i])
		}
		if dump := validate.ProtoDump(request); strings.Contains(dump, secret) {
			t.Fatalf("redacted request contains the password: %s", dump)
		}
	}
}
