package mdb_greenplum_cluster_v2

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	providerschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/greenplum/v1"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func poolDiscardTestObject(t *testing.T, value types.Bool) types.Object {
	t.Helper()
	model := NewYandexMdbGreenplumClusterV2ClusterConfigPoolModel()
	model.Mode = types.StringValue("SESSION")
	model.PoolDiscard = value
	obj, diags := types.ObjectValueFrom(context.Background(), yandexMdbGreenplumClusterV2ClusterConfigPoolModelType.AttrTypes, model)
	require.False(t, diags.HasError(), "%v", diags)
	return obj
}

func TestGreenplumClusterPoolDiscardExpand(t *testing.T) {
	for _, tt := range []struct {
		name  string
		value types.Bool
		want  *wrapperspb.BoolValue
	}{
		{"true", types.BoolValue(true), wrapperspb.Bool(true)},
		{"false", types.BoolValue(false), wrapperspb.Bool(false)},
		{"omitted", types.BoolNull(), nil},
		{"unknown", types.BoolUnknown(), nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			model := NewYandexMdbGreenplumClusterV2ClusterConfigModel()
			model.Pool = poolDiscardTestObject(t, tt.value)
			config, diags := types.ObjectValueFrom(ctx, yandexMdbGreenplumClusterV2ClusterConfigModelType.AttrTypes, model)
			require.False(t, diags.HasError(), "%v", diags)
			create := expandYandexMdbGreenplumClusterV2ClusterConfig_create(ctx, config, &diags)
			update := expandYandexMdbGreenplumClusterV2ClusterConfig_update(ctx, config, &diags)
			require.False(t, diags.HasError(), "%v", diags)
			require.Equal(t, tt.want, create.GetPool().GetPoolDiscard())
			require.Equal(t, tt.want, update.GetPool().GetPoolDiscard())
		})
	}
}

func TestGreenplumClusterPoolDiscardFlatten(t *testing.T) {
	for _, value := range []bool{true, false} {
		t.Run(fmt.Sprint(value), func(t *testing.T) {
			ctx := context.Background()
			var diags diag.Diagnostics
			config := &greenplum.ClusterConfigSet{Pool: &greenplum.ConnectionPoolerConfigSet{
				UserConfig:      &greenplum.ConnectionPoolerConfig{Mode: greenplum.ConnectionPoolerConfig_SESSION, PoolDiscard: wrapperspb.Bool(!value)},
				EffectiveConfig: &greenplum.ConnectionPoolerConfig{PoolDiscard: wrapperspb.Bool(value)},
			}}
			config.DbmsConfigSet = &greenplum.DBMSConfigSet{UserConfig: &greenplum.DBMSConfig{MaxConnections: wrapperspb.Int64(123)}}
			// A fresh model represents reading an imported cluster without prior configuration.
			obj := flattenYandexMdbGreenplumClusterV2ClusterConfig(ctx, config, NewYandexMdbGreenplumClusterV2ClusterConfigModel(), &diags)
			require.False(t, diags.HasError(), "%v", diags)
			pool := obj.Attributes()["pool"].(types.Object)
			require.Equal(t, types.BoolValue(value), pool.Attributes()["pool_discard"])
			require.Equal(t, types.StringValue("SESSION"), pool.Attributes()["mode"])

			dsModel := NewYandexMdbGreenplumClusterV2DatasourceModel()
			dsModel.Timeouts = timeouts.Value{Object: types.ObjectNull(timeouts.AttributesAll(ctx).GetType().(timeouts.Type).AttrTypes)}
			ds := flattenYandexMdbGreenplumClusterV2Datasource(ctx, &greenplum.Cluster{ClusterConfig: config}, dsModel, dsModel.Timeouts, &diags)
			require.False(t, diags.HasError(), "%v", diags)
			dsState := tfsdk.State{Schema: YandexMdbGreenplumClusterV2DatasourceSchema(ctx)}
			stateDiags := dsState.Set(ctx, ds)
			require.False(t, stateDiags.HasError(), "%v", stateDiags)
			var maxConnections types.Int64
			stateDiags = dsState.GetAttribute(ctx, path.Root("cluster_config").AtName("greenplum_config_set_6").AtName("max_connections"), &maxConnections)
			require.False(t, stateDiags.HasError(), "%v", stateDiags)
			require.Equal(t, types.Int64Value(123), maxConnections)
			dsPool := ds.Attributes()["cluster_config"].(types.Object).Attributes()["pool"].(types.Object)
			require.Equal(t, types.BoolValue(value), dsPool.Attributes()["pool_discard"])
		})
	}
	t.Run("missing effective value remains null", func(t *testing.T) {
		var diags diag.Diagnostics
		obj := flattenYandexMdbGreenplumClusterV2ClusterConfigPool(context.Background(), &greenplum.ConnectionPoolerConfigSet{}, NewYandexMdbGreenplumClusterV2ClusterConfigPoolModel(), &diags)
		require.False(t, diags.HasError(), "%v", diags)
		require.Equal(t, types.BoolNull(), obj.Attributes()["pool_discard"])
	})
}

// Exercise the actual Framework planning pipeline, including nested computed attributes.
type poolDiscardTestProvider struct{}

func (*poolDiscardTestProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "yandex"
}
func (*poolDiscardTestProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = providerschema.Schema{}
}
func (*poolDiscardTestProvider) Configure(context.Context, provider.ConfigureRequest, *provider.ConfigureResponse) {
}
func (*poolDiscardTestProvider) DataSources(context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{NewDataSource}
}
func (*poolDiscardTestProvider) Resources(context.Context) []func() resource.Resource {
	return []func() resource.Resource{NewResource}
}

func poolDiscardTestState(t *testing.T, pool types.Object) tfsdk.State {
	t.Helper()
	ctx := context.Background()
	s := YandexMdbGreenplumClusterV2ResourceSchema(ctx)
	objType := s.Type().(types.ObjectType)
	attrs := map[string]attr.Value{}
	for k, tp := range objType.AttrTypes {
		v, err := tp.ValueFromTerraform(ctx, tftypes.NewValue(tp.TerraformType(ctx), nil))
		require.NoError(t, err)
		attrs[k] = v
	}
	model := NewYandexMdbGreenplumClusterV2ClusterConfigModel()
	model.Pool = pool
	config, diags := types.ObjectValueFrom(ctx, yandexMdbGreenplumClusterV2ClusterConfigModelType.AttrTypes, model)
	require.False(t, diags.HasError(), "%v", diags)
	attrs["cluster_config"] = config
	attrs["id"] = types.StringValue("cluster-id")
	attrs["name"] = types.StringValue("test-cluster")
	raw, err := types.ObjectValueMust(objType.AttrTypes, attrs).ToTerraformValue(ctx)
	require.NoError(t, err)
	return tfsdk.State{Schema: s, Raw: raw}
}

func TestGreenplumClusterPoolDiscardPlan(t *testing.T) {
	ctx := context.Background()
	server := providerserver.NewProtocol6(&poolDiscardTestProvider{})()
	// GetProviderSchema validates both resource and data source schemas.
	schemas, err := server.GetProviderSchema(ctx, &tfprotov6.GetProviderSchemaRequest{})
	require.NoError(t, err)
	for _, d := range schemas.Diagnostics {
		require.NotEqual(t, tfprotov6.DiagnosticSeverityError, d.Severity, "%s: %s", d.Summary, d.Detail)
	}
	for _, previous := range []bool{true, false} {
		for _, next := range []string{"omitted", "no pool", "true", "false", "change sibling"} {
			t.Run(fmt.Sprintf("%t to %s", previous, next), func(t *testing.T) {
				prior := poolDiscardTestState(t, poolDiscardTestObject(t, types.BoolValue(previous)))
				value := types.BoolNull()
				want := previous
				if next == "true" || next == "false" {
					want = next == "true"
					value = types.BoolValue(want)
				}
				pool := poolDiscardTestObject(t, value)
				if next == "no pool" {
					pool = types.ObjectNull(yandexMdbGreenplumClusterV2ClusterConfigPoolModelType.AttrTypes)
				}
				if next == "change sibling" {
					attrs := pool.Attributes()
					attrs["size"] = types.Int64Value(20)
					pool = types.ObjectValueMust(pool.AttributeTypes(ctx), attrs)
				}
				configured := poolDiscardTestState(t, pool)
				priorDynamic, err := tfprotov6.NewDynamicValue(prior.Raw.Type(), prior.Raw)
				require.NoError(t, err)
				configDynamic, err := tfprotov6.NewDynamicValue(configured.Raw.Type(), configured.Raw)
				require.NoError(t, err)
				response, err := server.PlanResourceChange(ctx, &tfprotov6.PlanResourceChangeRequest{
					TypeName: "yandex_mdb_greenplum_cluster_v2", PriorState: &priorDynamic, Config: &configDynamic, ProposedNewState: &configDynamic,
				})
				require.NoError(t, err)
				for _, d := range response.Diagnostics {
					require.NotEqual(t, tfprotov6.DiagnosticSeverityError, d.Severity, "%s: %s", d.Summary, d.Detail)
				}
				raw, err := response.PlannedState.Unmarshal(prior.Raw.Type())
				require.NoError(t, err)
				plan := tfsdk.Plan{Schema: prior.Schema, Raw: raw}
				var got types.Bool
				diags := plan.GetAttribute(ctx, path.Root("cluster_config").AtName("pool").AtName("pool_discard"), &got)
				require.False(t, diags.HasError(), "%v", diags)
				require.Equal(t, types.BoolValue(want), got)
				if want == previous && next != "change sibling" {
					var plannedPool types.Object
					diags = plan.GetAttribute(ctx, path.Root("cluster_config").AtName("pool"), &plannedPool)
					require.False(t, diags.HasError(), "%v", diags)
					require.True(t, poolDiscardTestObject(t, types.BoolValue(previous)).Equal(plannedPool), "unchanged pool must not produce a diff")
				}
			})
		}
	}
}
