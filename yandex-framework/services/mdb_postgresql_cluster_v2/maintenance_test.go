package mdb_postgresql_cluster_v2

import (
	"context"
	"testing"
	"time"

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
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
	mdb "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

func maintenanceSlot(day, start, duration string, allow bool) types.Object {
	return types.ObjectValueMust(maintenanceSlotTypes, map[string]attr.Value{"day": types.StringValue(day), "start_time": types.StringValue(start), "duration": types.StringValue(duration), "allow_temporary_unavailability": types.BoolValue(allow)})
}
func maintenanceWeekly(slots ...attr.Value) types.Object {
	return types.ObjectValueMust(maintenanceTypes, map[string]attr.Value{"type": types.StringValue("WEEKLY"), "slot": types.SetValueMust(types.ObjectType{AttrTypes: maintenanceSlotTypes}, slots)})
}
func maintenanceLegacy(hour int64) types.Object {
	return types.ObjectValueMust(mdbcommon.MaintenanceWindowType.AttrTypes, map[string]attr.Value{"type": types.StringValue("WEEKLY"), "day": types.StringValue("MON"), "hour": types.Int64Value(hour)})
}

func maintenanceTestState(t *testing.T, modern, legacy types.Object) tfsdk.State {
	t.Helper()
	ctx := context.Background()
	var sr resource.SchemaResponse
	(&clusterResource{}).Schema(ctx, resource.SchemaRequest{}, &sr)
	attrs := map[string]attr.Value{}
	ot := sr.Schema.Type().(basetypes.ObjectType)
	for k, tp := range ot.AttrTypes {
		v, err := tp.ValueFromTerraform(ctx, tftypes.NewValue(tp.TerraformType(ctx), nil))
		require.NoError(t, err)
		attrs[k] = v
	}
	attrs["id"] = types.StringValue("test")
	attrs["config"] = baseConfig
	attrs["hosts"] = types.MapValueMust(hostType, map[string]attr.Value{})
	attrs["maintenance_windows"], attrs["maintenance_window"] = modern, legacy
	raw, err := types.ObjectValueMust(ot.AttrTypes, attrs).ToTerraformValue(ctx)
	require.NoError(t, err)
	return tfsdk.State{Schema: sr.Schema, Raw: raw}
}

func TestMaintenanceLegacyConversion(t *testing.T) {
	ctx := context.Background()
	for _, hour := range []int64{1, 3, 24} {
		var d diag.Diagnostics
		old := maintenanceLegacy(hour)
		w := expandMaintenance(ctx, types.ObjectNull(maintenanceTypes), old, &d)
		require.False(t, d.HasError(), "%v", d)
		s := w.GetWeeklyMaintenanceSchedule().Slots[0]
		require.Equal(t, int32(hour-1), s.StartTime.Hours)
		require.Equal(t, time.Hour, s.Duration.AsDuration())
		require.True(t, s.AllowTemporaryUnavailability)
		state := Cluster{MaintenanceWindow: old, MaintenanceWindows: types.ObjectNull(maintenanceTypes)}
		flattenMaintenance(ctx, &state, w, &d)
		require.True(t, state.MaintenanceWindow.Equal(old))
		require.True(t, state.MaintenanceWindows.IsNull())
	}
}

func TestMaintenanceValidation(t *testing.T) {
	cases := []struct {
		name  string
		slots []attr.Value
		valid bool
	}{
		{"empty", nil, false},
		{"valid", []attr.Value{maintenanceSlot("MON", "02:30:00", "3h", true)}, true},
		{"too short", []attr.Value{maintenanceSlot("MON", "02:30:00", "59m", true)}, false},
		{"too long", []attr.Value{maintenanceSlot("MON", "02:30:00", "24h1m", true)}, false},
		{"seconds", []attr.Value{maintenanceSlot("MON", "02:30:01", "3h", true)}, false},
		{"duration precision", []attr.Value{maintenanceSlot("MON", "02:30:00", "3h1s", true)}, false},
		{"all false", []attr.Value{maintenanceSlot("MON", "02:30:00", "3h", false)}, false},
		{"overlap", []attr.Value{maintenanceSlot("MON", "02:30:00", "3h", true), maintenanceSlot("MON", "05:00:00", "1h", false)}, false},
		{"week wrap", []attr.Value{maintenanceSlot("SUN", "23:00:00", "3h", true), maintenanceSlot("MON", "01:00:00", "1h", false)}, false},
		{"week adjacent", []attr.Value{maintenanceSlot("SUN", "23:00:00", "3h", true), maintenanceSlot("MON", "02:00:00", "1h", false)}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := expandModernMaintenance(context.Background(), maintenanceWeekly(tc.slots...))
			require.Equal(t, tc.valid, err == nil, "%v", err)
		})
	}
}

func TestMaintenancePlanForms(t *testing.T) {
	ctx := context.Background()
	legacy := maintenanceLegacy(3)
	modern := maintenanceWeekly(maintenanceSlot("THU", "01:30:00", "2h", true))
	ln, mn := types.ObjectNull(mdbcommon.MaintenanceWindowType.AttrTypes), types.ObjectNull(maintenanceTypes)
	for _, tc := range []struct {
		name                                                 string
		configModern, configLegacy, priorModern, priorLegacy types.Object
		create                                               bool
	}{
		{"legacy to modern", modern, ln, mn, legacy, false}, {"modern to legacy", mn, legacy, modern, ln, false},
		{"remove modern unmanaged", mn, ln, modern, ln, false}, {"remove legacy unmanaged", mn, ln, mn, legacy, false},
		{"create modern", modern, ln, mn, ln, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state := maintenanceTestState(t, tc.priorModern, tc.priorLegacy)
			cfg := maintenanceTestState(t, tc.configModern, tc.configLegacy)
			if tc.create {
				state.Raw = tftypes.NewValue(state.Raw.Type(), nil)
			}
			plan := maintenanceProtocolPlan(t, state, cfg)
			var got Cluster
			require.False(t, plan.Get(ctx, &got).HasError())
			wantModern, wantLegacy := tc.configModern, tc.configLegacy
			if tc.configModern.IsNull() && tc.configLegacy.IsNull() {
				wantModern, wantLegacy = tc.priorModern, tc.priorLegacy
			}
			require.True(t, got.MaintenanceWindows.Equal(wantModern))
			require.True(t, got.MaintenanceWindow.Equal(wantLegacy))
		})
	}
}

func TestMaintenanceRequestsAndSemanticDurations(t *testing.T) {
	ctx := context.Background()
	a := baseCluster
	a.MaintenanceWindow = types.ObjectNull(mdbcommon.MaintenanceWindowType.AttrTypes)
	a.MaintenanceWindows = maintenanceWeekly(maintenanceSlot("MON", "02:00:00", "3h", true), maintenanceSlot("THU", "01:30:00", "2h", false))
	b := a
	b.MaintenanceWindows = maintenanceWeekly(maintenanceSlot("THU", "01:30:00", "120m", false), maintenanceSlot("MON", "02:00:00", "180m", true))
	req, d := prepareUpdateRequest(ctx, &a, &b)
	require.False(t, d.HasError(), "%v", d)
	require.Empty(t, req.UpdateMask.Paths)
	require.Nil(t, req.MaintenanceWindow)
	require.Nil(t, req.MaintenanceWindows)
	b.MaintenanceWindows = maintenanceWeekly(maintenanceSlot("MON", "02:00:00", "4h", true))
	req, d = prepareUpdateRequest(ctx, &a, &b)
	require.False(t, d.HasError(), "%v", d)
	require.Equal(t, []string{"maintenance_windows"}, req.UpdateMask.Paths)
	require.Nil(t, req.MaintenanceWindow)
	require.NotNil(t, req.MaintenanceWindows)
	create, d := prepareCreateRequest(ctx, &a, &config.State{}, nil)
	require.False(t, d.HasError(), "%v", d)
	require.Nil(t, create.MaintenanceWindow)
	require.NotNil(t, create.MaintenanceWindows)
	a.Restore = types.ObjectValueMust(expectedRestoreAttrTypes, map[string]attr.Value{"backup_id": types.StringValue("backup"), "time": types.StringNull(), "time_inclusive": types.BoolValue(false)})
	restore, d := prepareRestoreRequest(ctx, &a, &config.State{}, nil)
	require.False(t, d.HasError(), "%v", d)
	require.Nil(t, restore.MaintenanceWindow)
	require.True(t, proto.Equal(create.MaintenanceWindows, restore.MaintenanceWindows))
}

func TestMaintenanceDriftImportAndAnytimeEmpty(t *testing.T) {
	ctx := context.Background()
	var d diag.Diagnostics
	w := expandMaintenance(ctx, maintenanceWeekly(maintenanceSlot("MON", "02:30:00", "3h", true)), types.ObjectNull(mdbcommon.MaintenanceWindowType.AttrTypes), &d)
	for _, old := range []types.Object{maintenanceLegacy(3), types.ObjectNull(mdbcommon.MaintenanceWindowType.AttrTypes)} {
		state := Cluster{MaintenanceWindow: old, MaintenanceWindows: types.ObjectNull(maintenanceTypes)}
		flattenMaintenance(ctx, &state, w, &d)
		require.False(t, d.HasError(), "%v", d)
		require.True(t, state.MaintenanceWindow.IsNull())
		require.False(t, state.MaintenanceWindows.IsNull())
	}
	for _, slots := range []types.Set{types.SetNull(types.ObjectType{AttrTypes: maintenanceSlotTypes}), types.SetValueMust(types.ObjectType{AttrTypes: maintenanceSlotTypes}, []attr.Value{})} {
		modern := types.ObjectValueMust(maintenanceTypes, map[string]attr.Value{"type": types.StringValue("ANYTIME"), "slot": slots})
		state := Cluster{MaintenanceWindows: modern}
		flattenMaintenance(ctx, &state, maintenanceAnytime(), &d)
		require.True(t, state.MaintenanceWindows.Equal(modern))
	}
}

func TestMaintenanceConfigUnknownConflictAndAutoscaling(t *testing.T) {
	ctx := context.Background()
	for field, value := range map[string]attr.Value{
		"day": types.StringUnknown(), "start_time": types.StringUnknown(),
		"duration": types.StringUnknown(), "allow_temporary_unavailability": types.BoolUnknown(),
	} {
		attrs := maintenanceSlot("MON", "02:00:00", "3h", false).Attributes()
		attrs[field] = value
		modern := maintenanceWeekly(types.ObjectValueMust(maintenanceSlotTypes, attrs))
		state := maintenanceTestState(t, modern, types.ObjectNull(mdbcommon.MaintenanceWindowType.AttrTypes))
		response := resource.ValidateConfigResponse{}
		validateMaintenanceConfig(ctx, resource.ValidateConfigRequest{Config: tfsdk.Config{Schema: state.Schema, Raw: state.Raw}}, &response)
		require.False(t, response.Diagnostics.HasError(), "unknown %s must defer cross-slot validation: %v", field, response.Diagnostics)
		var diagnostics diag.Diagnostics
		require.Nil(t, expandMaintenance(ctx, modern, types.ObjectNull(mdbcommon.MaintenanceWindowType.AttrTypes), &diagnostics))
		require.True(t, diagnostics.HasError(), "unknown %s must never be sent to the API", field)
	}
	modern := maintenanceWeekly(maintenanceSlot("MON", "02:00:00", "3h", true))
	legacy := maintenanceLegacy(3)
	for _, tc := range []struct {
		name           string
		modern, legacy types.Object
		threshold      int64
		valid          bool
	}{
		{"modern weekly autoscaling", modern, types.ObjectNull(mdbcommon.MaintenanceWindowType.AttrTypes), 70, true},
		{"legacy weekly autoscaling", types.ObjectNull(maintenanceTypes), legacy, 70, true},
		{"no schedule autoscaling", types.ObjectNull(maintenanceTypes), types.ObjectNull(mdbcommon.MaintenanceWindowType.AttrTypes), 70, false},
		{"conflicting", modern, legacy, 0, false},
		{"unknown", types.ObjectUnknown(maintenanceTypes), types.ObjectNull(mdbcommon.MaintenanceWindowType.AttrTypes), 70, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state := maintenanceTestState(t, tc.modern, tc.legacy)
			state.SetAttribute(ctx, path.Root("config").AtName("disk_size_autoscaling"), types.ObjectValueMust(expectedDiskSizeAutoscalingAttrs, map[string]attr.Value{"disk_size_limit": types.Int64Value(100), "planned_usage_threshold": types.Int64Value(tc.threshold), "emergency_usage_threshold": types.Int64Value(90)}))
			resp := resource.ValidateConfigResponse{}
			validateMaintenanceConfig(ctx, resource.ValidateConfigRequest{Config: tfsdk.Config{Schema: state.Schema, Raw: state.Raw}}, &resp)
			require.Equal(t, tc.valid, !resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
		})
	}
}

type maintenanceProtocolProvider struct{ api *mdb.MaintenanceWindows }

func (*maintenanceProtocolProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "yandex"
}
func (*maintenanceProtocolProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = providerschema.Schema{}
}
func (*maintenanceProtocolProvider) Configure(context.Context, provider.ConfigureRequest, *provider.ConfigureResponse) {
}
func (*maintenanceProtocolProvider) DataSources(context.Context) []func() datasource.DataSource {
	return nil
}
func (p *maintenanceProtocolProvider) Resources(context.Context) []func() resource.Resource {
	return []func() resource.Resource{func() resource.Resource { return &maintenanceProtocolResource{api: p.api} }}
}

type maintenanceProtocolResource struct {
	clusterResource
	api *mdb.MaintenanceWindows
}

func (r *maintenanceProtocolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	resp.State = req.State
	var model Cluster
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	flattenMaintenance(ctx, &model, r.api, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("maintenance_windows"), model.MaintenanceWindows)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("maintenance_window"), model.MaintenanceWindow)...)
}

func TestMaintenanceProtocolPlanEquivalentDurations(t *testing.T) {
	for _, duration := range []string{"180m", "3h0m0s"} {
		t.Run(duration, func(t *testing.T) {
			prior := maintenanceWeekly(maintenanceSlot("MON", "02:00:00", duration, true), maintenanceSlot("THU", "01:30:00", "2h", false))
			configured := maintenanceWeekly(maintenanceSlot("THU", "01:30:00", "120m", false), maintenanceSlot("MON", "02:00:00", "3h", true))
			got := maintenanceProtocolPlanWindows(t, prior, configured)
			require.True(t, prior.Equal(got), "Framework planning must preserve the prior duration spelling and set elements")
		})
	}
}

// Equal durations in different slots must never borrow each other's spelling.
func TestMaintenanceProtocolRead(t *testing.T) {
	ctx := context.Background()
	for _, mode := range []string{"unchanged", "reordered", "changed-duration", "changed-start", "changed-flag"} {
		t.Run(mode, func(t *testing.T) {
			modern := maintenanceWeekly(maintenanceSlot("MON", "02:00:00", "180m", true), maintenanceSlot("THU", "01:30:00", "3h0m", false))
			state := maintenanceTestState(t, modern, types.ObjectNull(mdbcommon.MaintenanceWindowType.AttrTypes))
			api, err := expandModernMaintenance(ctx, modern)
			require.NoError(t, err)
			slots := api.GetWeeklyMaintenanceSchedule().Slots
			switch mode {
			case "reordered":
				slots[0], slots[1] = slots[1], slots[0]
			case "changed-duration":
				slots[0].Duration = durationpb.New(4 * time.Hour)
			case "changed-start":
				slots[1].StartTime.Hours = 6
			case "changed-flag":
				slots[1].AllowTemporaryUnavailability = true
			}
			server := providerserver.NewProtocol6(&maintenanceProtocolProvider{api: api})()
			dynamic, err := tfprotov6.NewDynamicValue(state.Raw.Type(), state.Raw)
			require.NoError(t, err)
			read, err := server.ReadResource(ctx, &tfprotov6.ReadResourceRequest{TypeName: "yandex_mdb_postgresql_cluster_v2", CurrentState: &dynamic})
			require.NoError(t, err)
			for _, d := range read.Diagnostics {
				require.NotEqual(t, tfprotov6.DiagnosticSeverityError, d.Severity, "%s: %s", d.Summary, d.Detail)
			}
			got, err := read.NewState.Unmarshal(state.Raw.Type())
			require.NoError(t, err)
			if mode == "unchanged" || mode == "reordered" {
				require.True(t, state.Raw.Equal(got), "unchanged repeated durations must keep each slot spelling: %s", got)
				return
			}
			require.False(t, state.Raw.Equal(got), "real remote change must be visible")
			result := tfsdk.State{Schema: state.Schema, Raw: got}
			var windows types.Object
			require.False(t, result.GetAttribute(ctx, path.Root("maintenance_windows"), &windows).HasError())
			for _, raw := range windows.Attributes()["slot"].(types.Set).Elements() {
				slot := raw.(types.Object)
				day := maintenanceString(ctx, slot, "day")
				duration := maintenanceString(ctx, slot, "duration")
				if day == "MON" && mode != "changed-duration" {
					require.Equal(t, "180m", duration)
				}
				if day == "THU" && mode == "changed-duration" {
					require.Equal(t, "3h0m", duration)
				}
				if day == "MON" && mode == "changed-duration" {
					parsed, err := time.ParseDuration(duration)
					require.NoError(t, err)
					require.Equal(t, 4*time.Hour, parsed)
				}
			}
		})
	}
}

// A known duration must retain the same spelling when a different field
// resolves during apply, including when the prior spellings are ambiguous.
func TestMaintenanceProtocolPlanUnknownStability(t *testing.T) {
	for _, ambiguous := range []bool{false, true} {
		for field, unknown := range map[string]attr.Value{
			"day":                            types.StringUnknown(),
			"start_time":                     types.StringUnknown(),
			"allow_temporary_unavailability": types.BoolUnknown(),
		} {
			name := field
			if ambiguous {
				name += "/ambiguous"
			}
			t.Run(name, func(t *testing.T) {
				priorOther, configuredOther := "2h", "120m"
				wantDuration, wantOther := "180m", "2h"
				if ambiguous {
					priorOther, configuredOther = "3h0m", "180m"
					wantDuration, wantOther = "3h", "180m"
				}
				prior := maintenanceWeekly(maintenanceSlot("MON", "02:00:00", "180m", true), maintenanceSlot("THU", "01:30:00", priorOther, false))
				configuredAttrs := maintenanceSlot("MON", "02:00:00", "3h", true).Attributes()
				configuredAttrs[field] = unknown
				configured := maintenanceWeekly(types.ObjectValueMust(maintenanceSlotTypes, configuredAttrs), maintenanceSlot("THU", "01:30:00", configuredOther, false))
				wantAttrs := maintenanceSlot("MON", "02:00:00", wantDuration, true).Attributes()
				wantAttrs[field] = unknown
				want := maintenanceWeekly(types.ObjectValueMust(maintenanceSlotTypes, wantAttrs), maintenanceSlot("THU", "01:30:00", wantOther, false))
				require.True(t, want.Equal(maintenanceProtocolPlanWindows(t, prior, configured)), "initial plan must preserve unknown fields and normalize only known durations")

				resolved := maintenanceWeekly(maintenanceSlot("MON", "02:00:00", "3h", true), maintenanceSlot("THU", "01:30:00", configuredOther, false))
				wantResolved := maintenanceWeekly(maintenanceSlot("MON", "02:00:00", wantDuration, true), maintenanceSlot("THU", "01:30:00", wantOther, false))
				require.True(t, wantResolved.Equal(maintenanceProtocolPlanWindows(t, prior, resolved)), "resolving another field must not change known duration spellings")
			})
		}
	}
}

func maintenanceProtocolPlanWindows(t *testing.T, priorWindows, configuredWindows types.Object) types.Object {
	t.Helper()
	legacyNull := types.ObjectNull(mdbcommon.MaintenanceWindowType.AttrTypes)
	plan := maintenanceProtocolPlan(t, maintenanceTestState(t, priorWindows, legacyNull), maintenanceTestState(t, configuredWindows, legacyNull))
	var got types.Object
	require.False(t, plan.GetAttribute(context.Background(), path.Root("maintenance_windows"), &got).HasError())
	return got
}

func maintenanceProtocolPlan(t *testing.T, prior, configured tfsdk.State) tfsdk.Plan {
	t.Helper()
	ctx := context.Background()
	priorDynamic, err := tfprotov6.NewDynamicValue(prior.Raw.Type(), prior.Raw)
	require.NoError(t, err)
	configDynamic, err := tfprotov6.NewDynamicValue(configured.Raw.Type(), configured.Raw)
	require.NoError(t, err)
	server := providerserver.NewProtocol6(&maintenanceProtocolProvider{})()
	response, err := server.PlanResourceChange(ctx, &tfprotov6.PlanResourceChangeRequest{
		TypeName:         "yandex_mdb_postgresql_cluster_v2",
		PriorState:       &priorDynamic,
		Config:           &configDynamic,
		ProposedNewState: &configDynamic,
	})
	require.NoError(t, err)
	for _, d := range response.Diagnostics {
		require.NotEqual(t, tfprotov6.DiagnosticSeverityError, d.Severity, "%s: %s", d.Summary, d.Detail)
	}
	raw, err := response.PlannedState.Unmarshal(prior.Raw.Type())
	require.NoError(t, err)
	return tfsdk.Plan{Schema: prior.Schema, Raw: raw}
}
