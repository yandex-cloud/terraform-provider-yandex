package mdb_mysql_cluster_v2

import (
	"context"
	"reflect"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	providerschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"
	msconfig "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1/config"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/internal/provider/framework/config"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/mysql/testfixtures"
	"google.golang.org/genproto/googleapis/type/timeofday"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestYandexProvider_MDBMySQLClusterPrepateUpdateRequestBasic(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cluster := baseCluster

	cluster.Version = types.StringValue("8.0")
	cluster.BackupWindowStart = types.ObjectValueMust(expectedBWSAttrs, map[string]attr.Value{
		"hours":   types.Int64Value(2),
		"minutes": types.Int64Value(0),
	})

	cluster.Name = types.StringValue("test-cluster-new")
	cluster.DeletionProtection = types.BoolValue(false)
	cluster.SecurityGroupIds = types.SetValueMust(types.StringType, []attr.Value{
		types.StringValue("test-sg-new"),
	})
	cluster.MaintenanceWindow = types.ObjectNull(expectedMWAttrs)
	cluster.MySQLConfig = NewMsSettingsMapValueMust(
		map[string]attr.Value{"max_connections": types.Int64Value(25)},
	)
	cluster.DiskSizeAutoscaling = types.ObjectNull(expectedDSAAttrs)
	cluster.PerformanceDiagnostics = types.ObjectNull(expectedPDAttrs)

	req, diags := prepareUpdateRequest(ctx, &baseCluster, &cluster)
	if diags.HasError() {
		t.Fatalf(
			"Unexpected expand diagnostics status: expected without error, actual with errors: %v",
			diags.Errors(),
		)
	}

	expectedUpdateReq := &mysql.UpdateClusterRequest{
		ClusterId: "test-id",
		Name:      "test-cluster-new",
		ConfigSpec: &mysql.ConfigSpec{
			BackupWindowStart: &timeofday.TimeOfDay{
				Hours:   2,
				Minutes: 0,
			},
			MysqlConfig: &mysql.ConfigSpec_MysqlConfig_8_0{
				MysqlConfig_8_0: &msconfig.MysqlConfig8_0{
					MaxConnections: wrapperspb.Int64(25),
				},
			},
		},
		MaintenanceWindow:  nil,
		SecurityGroupIds:   []string{"test-sg-new"},
		DeletionProtection: false,
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: []string{
				"name",
				"config_spec.backup_window_start.hours",
				"config_spec.backup_window_start.minutes",
				"config_spec.mysql_config_8_0.max_connections",
				"config_spec.mysql_config_8_0.default_authentication_plugin",
				"config_spec.mysql_config_8_0.innodb_print_all_deadlocks",
				"security_group_ids",
				"deletion_protection",
				"maintenance_window",
			},
		},
	}

	sort.Strings(req.UpdateMask.Paths)
	sort.Strings(expectedUpdateReq.UpdateMask.Paths)

	for i, path := range req.UpdateMask.Paths {
		if path != expectedUpdateReq.UpdateMask.Paths[i] {
			t.Fatalf("Unexpected update mask paths: expected %s, actual %s", expectedUpdateReq.UpdateMask, req.UpdateMask)
		}
	}

	if !reflect.DeepEqual(req, expectedUpdateReq) {
		t.Fatalf("Unexpected update request:\nexpected %s\nactual %s", req, expectedUpdateReq)
	}
}

func TestYandexProvider_MDBMySQLClusterPrepateUpdateVersionRequest(t *testing.T) {
	t.Parallel()

	cluster := baseCluster

	cluster.Version = types.StringValue("9.0")

	req, diags := prepareVersionUpdateRequest(&baseCluster, &cluster)
	if diags.HasError() {
		t.Fatalf(
			"Unexpected expand diagnostics status: expected without error, actual with errors: %v",
			diags.Errors(),
		)
	}

	expectedUpdateReq := &mysql.UpdateClusterRequest{
		ClusterId: "test-id",
		ConfigSpec: &mysql.ConfigSpec{
			Version: "9.0",
		},
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: []string{
				"config_spec.version",
			},
		},
	}

	for i, path := range req.UpdateMask.Paths {
		if path != expectedUpdateReq.UpdateMask.Paths[i] {
			t.Fatalf("Unexpected update mask paths: expected %s, actual %s", expectedUpdateReq.UpdateMask, req.UpdateMask)
		}
	}

	if !reflect.DeepEqual(req, expectedUpdateReq) {
		t.Fatalf("Unexpected update request:\nexpected %s\nactual %s", req, expectedUpdateReq)
	}
}

func TestMySQL97ResetSettings(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	state := baseCluster
	state.Version = types.StringValue("9.7")
	state.MySQLConfig = NewMsSettingsMapValueMust(map[string]attr.Value{
		"replica_parallel_workers": types.StringValue("4"),
		"innodb_redo_log_capacity": types.StringValue("268435456"),
		"max_connections":          types.StringValue("100"),
	})
	plan := state
	plan.MySQLConfig = NewMsSettingsMapValueMust(map[string]attr.Value{
		"max_connections": types.StringValue("150"),
	})
	req, diags := prepareUpdateRequest(ctx, &state, &plan)
	require.False(t, diags.HasError(), "%v", diags)
	require.True(t, req.UpdateMask.IsValid(req))
	require.ElementsMatch(t, []string{
		"config_spec.mysql_config_9_7.replica_parallel_workers",
		"config_spec.mysql_config_9_7.innodb_redo_log_capacity",
		"config_spec.mysql_config_9_7.max_connections",
	}, req.UpdateMask.Paths)
	require.True(t, proto.Equal(&msconfig.MysqlConfig9_7{MaxConnections: wrapperspb.Int64(150)}, req.ConfigSpec.GetMysqlConfig_9_7()))

	plan.MySQLConfig = NewMsSettingsMapNull()
	req, diags = prepareUpdateRequest(ctx, &state, &plan)
	require.False(t, diags.HasError(), "%v", diags)
	require.Len(t, req.UpdateMask.Paths, 3)
	require.True(t, req.UpdateMask.IsValid(req))
	require.True(t, proto.Equal(&msconfig.MysqlConfig9_7{}, req.ConfigSpec.GetMysqlConfig_9_7()))
}

// Build a state with the real schema, preserving typed nulls for optional fields.
func mysqlUpdateState(t *testing.T, version string, settings map[string]attr.Value) tfsdk.State {
	t.Helper()
	ctx := context.Background()
	var sch resource.SchemaResponse
	NewMySQLClusterResourceV2().Schema(ctx, resource.SchemaRequest{}, &sch)
	objectType := sch.Schema.Type().(basetypes.ObjectType)
	attrs := make(map[string]attr.Value, len(objectType.AttrTypes))
	for name, tp := range objectType.AttrTypes {
		value, err := tp.ValueFromTerraform(ctx, tftypes.NewValue(tp.TerraformType(ctx), nil))
		require.NoError(t, err)
		attrs[name] = value
	}
	attrs["id"] = types.StringValue("test-id")
	attrs["name"] = types.StringValue("test-cluster")
	attrs["environment"] = types.StringValue("PRESTABLE")
	attrs["network_id"] = types.StringValue("test-network")
	attrs["folder_id"] = types.StringValue("test-folder")
	attrs["version"] = types.StringValue(version)
	attrs["mysql_config"] = NewMsSettingsMapValueMust(settings)
	attrs["hosts"] = types.MapValueMust(hostType, map[string]attr.Value{})
	attrs["resources"] = baseCluster.Resources
	raw, err := types.ObjectValueMust(objectType.AttrTypes, attrs).ToTerraformValue(ctx)
	require.NoError(t, err)
	return tfsdk.State{Schema: sch.Schema, Raw: raw}
}

func TestMySQL97UpdateOperations(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name             string
		settings         map[string]attr.Value
		previousSettings map[string]attr.Value
		failOperation    string
		wantEvents       []string
	}{
		{
			name: "failed upgrade", settings: map[string]attr.Value{"max_connections": types.StringValue("200")},
			failOperation: "version", wantEvents: []string{"update:version", "wait:version"},
		},
		{
			name: "failed settings", settings: map[string]attr.Value{"max_connections": types.StringValue("200")},
			failOperation: "settings", wantEvents: []string{"update:version", "wait:version", "get", "update:settings", "wait:settings"},
		},
		{
			name: "removed setting", settings: map[string]attr.Value{"slave_parallel_workers": types.StringValue("4")},
		},
		{
			name: "unchanged removed setting", settings: map[string]attr.Value{"slave_parallel_workers": types.StringValue("4")},
			previousSettings: map[string]attr.Value{"slave_parallel_workers": types.StringValue("4")},
		},
		{
			name: "invalid value", settings: map[string]attr.Value{"max_connections": types.StringValue("invalid")},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			previous := tc.previousSettings
			if previous == nil {
				previous = map[string]attr.Value{"max_connections": types.StringValue("100")}
			}
			state := mysqlUpdateState(t, "8.4", previous)
			planned := mysqlUpdateState(t, "9.7", tc.settings)
			api := testfixtures.UpgradeCluster()
			api.FailOperation = tc.failOperation
			api.UpgradedConfig.GetMysqlConfig_9_7().UserConfig = &msconfig.MysqlConfig9_7{MaxConnections: wrapperspb.Int64(100)}
			r := &clusterResource{providerConfig: &providerconfig.Config{SDKv2: api.SDK(t)}}
			response := resource.UpdateResponse{State: state}
			r.Update(ctx, resource.UpdateRequest{State: state, Plan: tfsdk.Plan{Schema: planned.Schema, Raw: planned.Raw}}, &response)
			require.True(t, response.Diagnostics.HasError(), "expected the operation or settings validation to fail")
			require.Equal(t, tc.wantEvents, api.Events)
			if len(tc.wantEvents) == 0 {
				require.Empty(t, api.Requests, "invalid settings must be rejected before any Update RPC")
			}
		})
	}
}

type mysqlPlanProvider struct{}

func (*mysqlPlanProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "yandex"
}
func (*mysqlPlanProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = providerschema.Schema{}
}
func (*mysqlPlanProvider) Configure(context.Context, provider.ConfigureRequest, *provider.ConfigureResponse) {
}
func (*mysqlPlanProvider) DataSources(context.Context) []func() datasource.DataSource { return nil }
func (*mysqlPlanProvider) Resources(context.Context) []func() resource.Resource {
	return []func() resource.Resource{NewMySQLClusterResourceV2}
}

func TestMySQL97UnmanagedConfigUpgrade(t *testing.T) {
	t.Parallel()
	for _, version := range []string{"8.4", "9.7"} {
		t.Run(version, func(t *testing.T) {
			ctx := context.Background()
			state := mysqlUpdateState(t, "8.4", map[string]attr.Value{"slave_parallel_workers": types.StringValue("4")})
			configured := mysqlUpdateState(t, version, nil)
			require.False(t, configured.SetAttribute(ctx, path.Root("mysql_config"), NewMsSettingsMapNull()).HasError())
			// Terraform proposes prior values for omitted Optional+Computed attributes.
			proposed := state
			require.False(t, proposed.SetAttribute(ctx, path.Root("version"), types.StringValue(version)).HasError())
			plan := mysqlResourcePlan(t, state, configured, proposed)
			var old, next Cluster
			require.False(t, state.Get(ctx, &old).HasError())
			require.False(t, plan.Get(ctx, &next).HasError())
			if version == "9.7" {
				require.True(t, next.MySQLConfig.IsUnknown(), "renamed settings must be read after the upgrade")
			} else {
				require.True(t, next.MySQLConfig.Equal(old.MySQLConfig))
			}
			req, diags := prepareUpdateRequest(ctx, &old, &next)
			require.False(t, diags.HasError(), "%v", diags)
			for _, field := range req.UpdateMask.Paths {
				require.NotContains(t, field, "mysql_config", "unmanaged settings must be migrated by the server")
			}
		})
	}
}

func TestMySQL97UpgradeWithUnchangedSettings(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name        string
		omitConfig  bool
		wantUpdates int
	}{
		{"managed", false, 2},
		{"omitted", true, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			api := testfixtures.UpgradeWithDefaultRedoCapacity(wrapperspb.Int64(100))
			r := &clusterResource{providerConfig: &providerconfig.Config{SDKv2: api.SDK(t)}}
			state := mysqlUpdateState(t, "8.4", map[string]attr.Value{"max_connections": types.StringValue("100")})
			read := resource.ReadResponse{State: state}
			r.Read(ctx, resource.ReadRequest{State: state}, &read)
			require.False(t, read.Diagnostics.HasError(), "%v", read.Diagnostics)
			proposed := read.State
			require.False(t, proposed.SetAttribute(ctx, path.Root("version"), types.StringValue("9.7")).HasError())
			configured := proposed
			want := &msconfig.MysqlConfig9_7{MaxConnections: wrapperspb.Int64(100)}
			if tc.omitConfig {
				require.False(t, configured.SetAttribute(ctx, path.Root("mysql_config"), NewMsSettingsMapNull()).HasError())
				want = api.UpgradedConfig.GetMysqlConfig_9_7().GetUserConfig()
			}
			plan := mysqlResourcePlan(t, read.State, configured, proposed)
			response := resource.UpdateResponse{State: read.State}
			r.Update(ctx, resource.UpdateRequest{State: read.State, Plan: plan}, &response)
			require.False(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)
			require.True(t, proto.Equal(want, api.Cluster.Config.GetMysqlConfig_9_7().GetUserConfig()), "unexpected user_config after first apply")
			require.Len(t, api.Requests, tc.wantUpdates)
			next := mysqlResourcePlan(t, response.State, configured, response.State)
			require.True(t, next.Raw.Equal(response.State.Raw))
		})
	}
}

func mysqlSettingsValues(settings map[string]string) map[string]attr.Value {
	result := make(map[string]attr.Value)
	for name, value := range settings {
		result[name] = types.StringValue(value)
	}
	return result
}

func TestMySQL97UpgradeUsesActualSettings(t *testing.T) {
	t.Parallel()
	for _, tc := range testfixtures.UpgradeCases() {
		t.Run(tc.Name, func(t *testing.T) {
			ctx := context.Background()
			api := testfixtures.UpgradeCluster()
			api.GetError = tc.GetError
			if tc.AlreadyReset {
				api.UpgradedConfig.GetMysqlConfig_9_7().UserConfig = &msconfig.MysqlConfig9_7{MaxConnections: wrapperspb.Int64(100)}
			}
			settings := mysqlSettingsValues(tc.Settings)
			previous := mysqlSettingsValues(testfixtures.UpgradeSettings("8.4"))
			previous["slave_parallel_type"] = types.Int64Value(int64(msconfig.MysqlConfig8_4_LOGICAL_CLOCK))
			state := mysqlUpdateState(t, "8.4", previous)
			planned := mysqlUpdateState(t, "9.7", settings)
			if tc.Description != "" {
				require.False(t, planned.SetAttribute(ctx, path.Root("description"), types.StringValue(tc.Description)).HasError())
			}
			r := &clusterResource{providerConfig: &providerconfig.Config{SDKv2: api.SDK(t)}}
			response := resource.UpdateResponse{State: state}
			r.Update(ctx, resource.UpdateRequest{State: state, Plan: tfsdk.Plan{Schema: planned.Schema, Raw: planned.Raw}}, &response)
			require.Equal(t, tc.GetError != nil, response.Diagnostics.HasError(), "%v", response.Diagnostics)
			wantEvents := []string{"update:version", "wait:version", "get"}
			if len(tc.WantPaths) == 0 {
				require.Len(t, api.Requests, 1, "a settings no-op must not reach the API")
			} else {
				require.Len(t, api.Requests, 2)
				require.ElementsMatch(t, tc.WantPaths, api.Requests[1].UpdateMask.Paths)
				wantEvents = append(wantEvents, "update:settings", "wait:settings")
			}
			if tc.GetError == nil {
				wantEvents = append(wantEvents, "get")
				var result Cluster
				require.False(t, response.State.Get(ctx, &result).HasError())
				require.Equal(t, "9.7", result.Version.ValueString())
				require.True(t, result.MySQLConfig.Equal(NewMsSettingsMapValueMust(settings)))
			}
			require.Equal(t, wantEvents, api.Events)
		})
	}
}

func TestMySQL97RetryAfterSettingsFailure(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	api := testfixtures.UpgradeCluster()
	api.FailOperation = "settings"
	r := &clusterResource{providerConfig: &providerconfig.Config{SDKv2: api.SDK(t)}}
	state := mysqlUpdateState(t, "8.4", mysqlSettingsValues(testfixtures.UpgradeSettings("8.4")))
	planned := mysqlUpdateState(t, "9.7", map[string]attr.Value{"max_connections": types.StringValue("100")})
	plan := tfsdk.Plan{Schema: planned.Schema, Raw: planned.Raw}
	first := resource.UpdateResponse{State: state}
	r.Update(ctx, resource.UpdateRequest{State: state, Plan: plan}, &first)
	require.True(t, first.Diagnostics.HasError(), "the second operation must fail")
	require.Len(t, api.Requests, 2)
	require.Equal(t, "9.7", api.Cluster.Config.Version, "the successful upgrade is not rolled back")
	require.Equal(t, int64(4), api.Cluster.Config.GetMysqlConfig_9_7().GetUserConfig().GetReplicaParallelWorkers().GetValue())

	read := resource.ReadResponse{State: first.State}
	r.Read(ctx, resource.ReadRequest{State: first.State}, &read)
	require.False(t, read.Diagnostics.HasError(), "%v", read.Diagnostics)
	var refreshed Cluster
	require.False(t, read.State.Get(ctx, &refreshed).HasError())
	require.Equal(t, "9.7", refreshed.Version.ValueString())
	require.True(t, refreshed.MySQLConfig.Equal(NewMsSettingsMapValueMust(mysqlSettingsValues(testfixtures.UpgradeSettings("9.7")))), "Read must update the settings schema together with the version: %v", refreshed.MySQLConfig)

	api.FailOperation = ""
	proposed := read.State
	require.False(t, proposed.SetAttribute(ctx, path.Root("mysql_config"), NewMsSettingsMapValueMust(map[string]attr.Value{"max_connections": types.StringValue("100")})).HasError())
	retryPlan := mysqlResourcePlan(t, read.State, planned, proposed)
	retry := resource.UpdateResponse{State: read.State}
	r.Update(ctx, resource.UpdateRequest{State: read.State, Plan: retryPlan}, &retry)
	require.False(t, retry.Diagnostics.HasError(), "%v", retry.Diagnostics)
	require.Len(t, api.Requests, 3, "retry must only update settings, without another upgrade")
	req := api.Requests[2]
	require.True(t, req.UpdateMask.IsValid(req), "old setting names must not survive into the retry mask")
	require.Contains(t, req.UpdateMask.Paths, "config_spec.mysql_config_9_7.replica_parallel_workers")
	require.NotContains(t, req.UpdateMask.Paths, "config_spec.version")
	user := api.Cluster.Config.GetMysqlConfig_9_7().GetUserConfig()
	require.Nil(t, user.ReplicaParallelWorkers)
	require.Nil(t, user.RplSemiSyncSourceWaitForReplicaCount)
	require.Nil(t, user.InnodbRedoLogCapacity)
	require.Equal(t, int64(100), user.GetMaxConnections().GetValue())
	var result Cluster
	require.False(t, retry.State.Get(ctx, &result).HasError())
	require.True(t, result.MySQLConfig.Equal(NewMsSettingsMapValueMust(map[string]attr.Value{"max_connections": types.StringValue("100")})))
	nextPlan := mysqlResourcePlan(t, retry.State, planned, retry.State)
	require.True(t, nextPlan.Raw.Equal(retry.State.Raw), "the same configuration must produce no changes after recovery")
}

func mysqlResourcePlan(t *testing.T, state, configured, proposed tfsdk.State) tfsdk.Plan {
	t.Helper()
	ctx := context.Background()
	prior, err := tfprotov6.NewDynamicValue(state.Raw.Type(), state.Raw)
	require.NoError(t, err)
	config, err := tfprotov6.NewDynamicValue(configured.Raw.Type(), configured.Raw)
	require.NoError(t, err)
	proposal, err := tfprotov6.NewDynamicValue(proposed.Raw.Type(), proposed.Raw)
	require.NoError(t, err)
	server := providerserver.NewProtocol6(&mysqlPlanProvider{})()
	response, err := server.PlanResourceChange(ctx, &tfprotov6.PlanResourceChangeRequest{
		TypeName: "yandex_mdb_mysql_cluster_v2", PriorState: &prior, Config: &config, ProposedNewState: &proposal,
	})
	require.NoError(t, err)
	for _, d := range response.Diagnostics {
		require.NotEqual(t, tfprotov6.DiagnosticSeverityError, d.Severity, "%s: %s", d.Summary, d.Detail)
	}
	raw, err := response.PlannedState.Unmarshal(state.Raw.Type())
	require.NoError(t, err)
	return tfsdk.Plan{Schema: state.Schema, Raw: raw}
}
