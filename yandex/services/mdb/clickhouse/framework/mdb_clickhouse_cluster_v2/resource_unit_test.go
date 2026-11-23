package mdb_clickhouse_cluster_v2

import (
	"context"
	"slices"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	clickhouseConfig "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1/config"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/clickhouse/framework/mdb_clickhouse_cluster_v2/models"
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

func TestKafkaSaslPasswordPlanModifiers(t *testing.T) {
	t.Parallel()

	passwordAttribute := KafkaSchema().Attributes["sasl_password"].(schema.StringAttribute)
	if len(passwordAttribute.PlanModifiers) != 2 {
		t.Fatalf("sasl_password plan modifiers count = %d, want 2", len(passwordAttribute.PlanModifiers))
	}

	testCases := []struct {
		name   string
		config types.String
		state  types.String
		plan   types.String
		want   types.String
	}{
		{
			name:   "null state normalizes unknown plan to null",
			config: types.StringNull(),
			state:  types.StringNull(),
			plan:   types.StringUnknown(),
			want:   types.StringNull(),
		},
		{
			name:   "known state is preserved",
			config: types.StringNull(),
			state:  types.StringValue("old-password"),
			plan:   types.StringUnknown(),
			want:   types.StringValue("old-password"),
		},
		{
			name:   "known config remains known",
			config: types.StringValue("new-password"),
			state:  types.StringNull(),
			plan:   types.StringValue("new-password"),
			want:   types.StringValue("new-password"),
		},
		{
			name:   "unknown config remains unknown",
			config: types.StringUnknown(),
			state:  types.StringNull(),
			plan:   types.StringUnknown(),
			want:   types.StringUnknown(),
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			request := planmodifier.StringRequest{
				ConfigValue: testCase.config,
				StateValue:  testCase.state,
				PlanValue:   testCase.plan,
			}
			response := planmodifier.StringResponse{PlanValue: testCase.plan}

			for _, modifier := range passwordAttribute.PlanModifiers {
				modifier.PlanModifyString(context.Background(), request, &response)
				request.PlanValue = response.PlanValue
			}

			if !response.PlanValue.Equal(testCase.want) {
				t.Fatalf("planned sasl_password = %s, want %s", response.PlanValue, testCase.want)
			}
		})
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

func TestShardGroupWriteOnlySchemaAndConfig(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := schema.Schema{
		Description: "Test schema for shard group write-only passwords.",
		Blocks:      map[string]schema.Block{"shard_group": ShardGroupSchema()},
	}
	if diags := s.ValidateImplementation(ctx); diags.HasError() {
		t.Fatalf("Unexpected schema diagnostics: %v", diags.Errors())
	}
	replicaSchema := s.Blocks["shard_group"].(schema.ListNestedBlock).NestedObject.Blocks["external_shard"].(schema.ListNestedBlock).NestedObject.Blocks["replica"].(schema.ListNestedBlock)
	passwordSchema := replicaSchema.NestedObject.Attributes["password_wo"].(schema.StringAttribute)
	if !passwordSchema.WriteOnly || !passwordSchema.Sensitive {
		t.Fatal("password_wo must be write-only and sensitive")
	}
	p := path.Root("shard_group").AtListIndex(0).AtName("external_shard").AtListIndex(0).AtName("replica").AtListIndex(0)
	cases := []struct {
		name          string
		password, wo  types.String
		version       types.Int64
		expectedError bool
	}{
		{name: "valid", wo: types.StringValue("config-secret"), version: types.Int64Value(1)},
		{name: "conflicting_passwords", password: types.StringValue("legacy"), wo: types.StringValue("config-secret"), version: types.Int64Value(1), expectedError: true},
		{name: "missing_version", wo: types.StringValue("config-secret"), expectedError: true},
		{name: "missing_password", version: types.Int64Value(1), expectedError: true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			r := models.ExternalShardReplica{Host: types.StringValue("host"), Password: c.password, PasswordWo: c.wo, PasswordWoVersion: c.version}
			sibling := models.ExternalShardReplica{Host: types.StringValue("other-host"), Password: types.StringValue("sibling-secret")}
			cluster := testShardGroupCluster(t, testExternalShard(t, "external", r, sibling))
			value := types.ObjectValueMust(s.Type().(types.ObjectType).AttrTypes, map[string]attr.Value{"shard_group": cluster.ShardGroup})
			raw, err := value.ToTerraformValue(ctx)
			if err != nil {
				t.Fatal(err)
			}
			config := tfsdk.Config{Raw: raw, Schema: s}
			var diags diag.Diagnostics
			var groups types.List
			diags.Append(config.GetAttribute(ctx, path.Root("shard_group"), &groups)...)
			configured := externalReplicasByIdentity(ctx, groups, nil, &diags)
			if diags.HasError() || !configured[externalReplicaKey{"analytics", "external", "host"}].PasswordWo.Equal(c.wo) {
				t.Fatalf("Failed to read password from config: %v", diags)
			}
			for name, value := range map[string]types.String{"password": c.password, "password_wo": c.wo} {
				for _, v := range replicaSchema.NestedObject.Attributes[name].(schema.StringAttribute).Validators {
					var resp validator.StringResponse
					v.ValidateString(ctx, validator.StringRequest{Path: p.AtName(name), PathExpression: p.AtName(name).Expression(), Config: config, ConfigValue: value}, &resp)
					diags.Append(resp.Diagnostics...)
				}
			}
			for _, v := range replicaSchema.NestedObject.Attributes["password_wo_version"].(schema.Int64Attribute).Validators {
				var resp validator.Int64Response
				v.ValidateInt64(ctx, validator.Int64Request{Path: p.AtName("password_wo_version"), PathExpression: p.AtName("password_wo_version").Expression(), Config: config, ConfigValue: c.version}, &resp)
				diags.Append(resp.Diagnostics...)
			}
			if diags.HasError() != c.expectedError {
				t.Fatalf("Expected error %t, got diagnostics: %v", c.expectedError, diags.Errors())
			}
		})
	}
}
