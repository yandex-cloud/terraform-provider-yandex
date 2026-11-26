package mdb_clickhouse_cluster_v2

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/clickhouse/framework/mdb_clickhouse_cluster_v2/models"
)

func TestDictionaryPostgresqlV2ConfigValidation(t *testing.T) {
	t.Parallel()

	replicaType := types.ObjectType{AttrTypes: models.DictionaryPostgresqlReplicaV2AttrTypes}
	replicas := func(values ...models.DictionaryPostgresqlReplicaV2) types.List {
		result, diags := types.ListValueFrom(context.Background(), replicaType, values)
		require.False(t, diags.HasError(), "%v", diags)
		return result
	}
	cases := []struct {
		name      string
		overrides map[string]attr.Value
		wantError bool
	}{
		{"query with replicas", nil, false},
		{"neither table nor query", map[string]attr.Value{"query": types.StringNull()}, true},
		{"both table and query", map[string]attr.Value{"table": types.StringValue("cities")}, true},
		{"table with where", map[string]attr.Value{
			"table": types.StringValue("cities"), "query": types.StringNull(), "where": types.StringValue("active = true"),
		}, false},
		{"query with where", map[string]attr.Value{"where": types.StringValue("active = true")}, true},
		{"missing replicas", map[string]attr.Value{"replicas": types.ListNull(replicaType)}, true},
		{"empty replicas", map[string]attr.Value{"replicas": types.ListValueMust(replicaType, []attr.Value{})}, true},
		{"duplicate endpoints with different priorities and passwords", map[string]attr.Value{"replicas": replicas(
			models.DictionaryPostgresqlReplicaV2{Host: types.StringValue("pg"), Priority: types.Int64Value(1), Password: types.StringValue("first")},
			models.DictionaryPostgresqlReplicaV2{Host: types.StringValue("pg"), Port: types.Int64Value(5432), User: types.StringValue("user"), Priority: types.Int64Value(2), Password: types.StringValue("second")},
		)}, true},
		{"duplicate endpoints with default source port", map[string]attr.Value{"port": types.Int64Null(), "replicas": replicas(
			models.DictionaryPostgresqlReplicaV2{Host: types.StringValue("pg")},
			models.DictionaryPostgresqlReplicaV2{Host: types.StringValue("pg"), Port: types.Int64Value(5432)},
		)}, true},
		{"same host with different ports", map[string]attr.Value{"replicas": replicas(
			models.DictionaryPostgresqlReplicaV2{Host: types.StringValue("pg")},
			models.DictionaryPostgresqlReplicaV2{Host: types.StringValue("pg"), Port: types.Int64Value(5433)},
		)}, false},
		{"same host with different users", map[string]attr.Value{"replicas": replicas(
			models.DictionaryPostgresqlReplicaV2{Host: types.StringValue("pg")},
			models.DictionaryPostgresqlReplicaV2{Host: types.StringValue("pg"), User: types.StringValue("other")},
		)}, false},
		{"unknown inherited port", map[string]attr.Value{"port": types.Int64Unknown(), "replicas": replicas(
			models.DictionaryPostgresqlReplicaV2{Host: types.StringValue("pg")},
			models.DictionaryPostgresqlReplicaV2{Host: types.StringValue("pg"), Port: types.Int64Value(5432)},
		)}, false},
		{"unknown replica host", map[string]attr.Value{"replicas": replicas(models.DictionaryPostgresqlReplicaV2{Host: types.StringUnknown()})}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			config := dictionaryTestConfig(t, tc.overrides)
			dynamicValue := dictionaryDynamicValue(t, config.Raw)

			server := providerserver.NewProtocol6(dictionaryCompatibilityProvider{})()

			response, err := server.ValidateResourceConfig(ctx, &tfprotov6.ValidateResourceConfigRequest{
				TypeName: "yandex_mdb_clickhouse_cluster_v2", Config: dynamicValue,
			})

			require.NoError(t, err)
			require.NotNil(t, response)

			var errors []string
			for _, diagnostic := range response.Diagnostics {
				if diagnostic.Severity == tfprotov6.DiagnosticSeverityError {
					errors = append(errors, diagnostic.Summary+": "+diagnostic.Detail)
				}
			}

			require.Equal(t, tc.wantError, len(errors) > 0, "%v", errors)
		})
	}
}

func TestDictionaryPostgresqlV2SSLModePlan(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	prior := dictionaryTestConfig(t, nil)
	priorValue := dictionaryDynamicValue(t, prior.Raw)

	for _, tc := range []struct {
		name string
		mode types.String
		want types.String
	}{
		{"omitted keeps current mode", types.StringNull(), types.StringValue("REQUIRE")},
		{"explicit change", types.StringValue("DISABLE"), types.StringValue("DISABLE")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			config := dictionaryTestConfig(t, map[string]attr.Value{"ssl_mode": tc.mode})
			configValue := dictionaryDynamicValue(t, config.Raw)

			server := providerserver.NewProtocol6(dictionaryCompatibilityProvider{})()

			response, err := server.PlanResourceChange(ctx, &tfprotov6.PlanResourceChangeRequest{
				TypeName: "yandex_mdb_clickhouse_cluster_v2", PriorState: priorValue, Config: configValue, ProposedNewState: configValue,
			})

			require.NoError(t, err)
			require.NotNil(t, response)
			for _, diagnostic := range response.Diagnostics {
				require.NotEqual(t, tfprotov6.DiagnosticSeverityError, diagnostic.Severity, "%s: %s", diagnostic.Summary, diagnostic.Detail)
			}
			require.NotNil(t, response.PlannedState)

			raw, err := response.PlannedState.Unmarshal(config.Schema.Type().TerraformType(ctx))
			require.NoError(t, err)

			plan := tfsdk.Plan{Schema: config.Schema, Raw: raw}
			var mode types.String
			diags := plan.GetAttribute(ctx, path.Root("external_dictionary").AtMapKey("dict").AtName("source").AtName("postgresql_source_v2").AtName("ssl_mode"), &mode)
			require.False(t, diags.HasError(), "%v", diags)

			require.Equal(t, tc.want, mode)
		})
	}
}

func dictionaryTestConfig(t *testing.T, overrides map[string]attr.Value) tfsdk.Config {
	t.Helper()

	ctx := context.Background()
	var schemaResponse resource.SchemaResponse
	NewClickHouseClusterResourceV2().Schema(ctx, resource.SchemaRequest{}, &schemaResponse)
	require.False(t, schemaResponse.Diagnostics.HasError(), "%v", schemaResponse.Diagnostics)

	replicas, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: models.DictionaryPostgresqlReplicaV2AttrTypes}, []models.DictionaryPostgresqlReplicaV2{
		{Host: types.StringValue("pg")},
	})
	require.False(t, d.HasError(), "%v", d)

	source, d := types.ObjectValueFrom(ctx, models.DictionaryPostgresqlSourceV2AttrTypes, models.DictionaryPostgresqlSourceV2{
		Db: types.StringValue("db"), Query: types.StringValue("SELECT * FROM cities"), Port: types.Int64Value(5432),
		User: types.StringValue("user"), Replicas: replicas, SslMode: types.StringValue("REQUIRE"),
	})
	require.False(t, d.HasError(), "%v", d)
	sourceValues := source.Attributes()
	for name, value := range overrides {
		sourceValues[name] = value
	}

	configTypes := schemaResponse.Schema.Type().(types.ObjectType).AttrTypes
	values := nullDictionaryValues(configTypes)
	// Other cluster settings may be unresolved while Terraform validates the dictionary.
	for name, attribute := range schemaResponse.Schema.Attributes {
		if attribute.IsRequired() {
			value, err := configTypes[name].ValueFromTerraform(ctx, tftypes.NewValue(configTypes[name].TerraformType(ctx), tftypes.UnknownValue))
			require.NoError(t, err)
			values[name] = value
		}
	}
	maintenance, d := types.ObjectValueFrom(ctx, models.MaintenanceWindowAttrTypes, models.MaintenanceWindow{Type: types.StringValue("ANYTIME")})
	require.False(t, d.HasError(), "%v", d)
	values["maintenance_window"] = maintenance
	values["external_dictionary"] = makeDictMapTF(map[string]types.Object{
		"dict": makeDictTF(makeDictSourceTF("postgresql_source_v2", types.ObjectValueMust(models.DictionaryPostgresqlSourceV2AttrTypes, sourceValues))),
	})
	raw, err := types.ObjectValueMust(configTypes, values).ToTerraformValue(ctx)
	require.NoError(t, err)

	return tfsdk.Config{Schema: schemaResponse.Schema, Raw: raw}
}

func dictionaryDynamicValue(t *testing.T, raw tftypes.Value) *tfprotov6.DynamicValue {
	t.Helper()

	value, err := tfprotov6.NewDynamicValue(raw.Type(), raw)
	require.NoError(t, err)

	return &value
}

// Exercise the framework RPCs using the production resource.
type dictionaryCompatibilityProvider struct{}

func (dictionaryCompatibilityProvider) Metadata(_ context.Context, _ provider.MetadataRequest, response *provider.MetadataResponse) {
	response.TypeName = "yandex"
}
func (dictionaryCompatibilityProvider) Schema(context.Context, provider.SchemaRequest, *provider.SchemaResponse) {
}
func (dictionaryCompatibilityProvider) Configure(context.Context, provider.ConfigureRequest, *provider.ConfigureResponse) {
}
func (dictionaryCompatibilityProvider) Resources(context.Context) []func() resource.Resource {
	return []func() resource.Resource{NewClickHouseClusterResourceV2}
}
func (dictionaryCompatibilityProvider) DataSources(context.Context) []func() datasource.DataSource {
	return nil
}

func nullDictionaryValues(attrTypes map[string]attr.Type) map[string]attr.Value {
	values := make(map[string]attr.Value, len(attrTypes))
	for name, typ := range attrTypes {
		value, err := typ.ValueFromTerraform(context.Background(), tftypes.NewValue(typ.TerraformType(context.Background()), nil))
		if err != nil {
			panic(err)
		}
		values[name] = value
	}
	return values
}
