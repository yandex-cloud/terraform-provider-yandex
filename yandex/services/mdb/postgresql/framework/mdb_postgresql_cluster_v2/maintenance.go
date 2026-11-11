package mdb_postgresql_cluster_v2

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	mdb "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
	"google.golang.org/genproto/googleapis/type/dayofweek"
	"google.golang.org/genproto/googleapis/type/timeofday"
	"google.golang.org/protobuf/types/known/durationpb"
)

var maintenanceDays = []string{"MON", "TUE", "WED", "THU", "FRI", "SAT", "SUN"}
var maintenanceSlotTypes = map[string]attr.Type{
	"day": types.StringType, "start_time": types.StringType, "duration": types.StringType, "allow_temporary_unavailability": types.BoolType,
}
var maintenanceTypes = map[string]attr.Type{"type": types.StringType, "slot": types.SetType{ElemType: types.ObjectType{AttrTypes: maintenanceSlotTypes}}}

func maintenanceSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional: true, Computed: true, Description: "Maintenance schedule. Conflicts with maintenance_window. Times are UTC. Omitting both attributes preserves the existing schedule; set ANYTIME explicitly to remove time restrictions.",
		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{Required: true, Description: "ANYTIME or WEEKLY. ANYTIME forbids slots; WEEKLY requires at least one.", Validators: []validator.String{stringvalidator.OneOf("ANYTIME", "WEEKLY")}},
			"slot": schema.SetNestedAttribute{Optional: true, Description: "Unordered weekly slots. Must not overlap, including across the end of the week. At least one slot must permit temporary write unavailability.", NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
				"day":                            schema.StringAttribute{Required: true, Description: "UTC weekday: MON, TUE, WED, THU, FRI, SAT or SUN.", Validators: []validator.String{stringvalidator.OneOf(maintenanceDays...)}},
				"start_time":                     schema.StringAttribute{Required: true, Description: "UTC start time in HH:MM:00 format, with minute precision."},
				"duration":                       schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{maintenanceDurationPlanModifier{}}, Description: "Duration from 1h through 24h with minute precision, for example 3h or 90m."},
				"allow_temporary_unavailability": schema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), Description: "Allow maintenance that may temporarily interrupt writes. Defaults to false; at least one weekly slot must allow this."},
			}}},
		},
	}
}

func maintenanceKnown(value attr.Value) bool {
	if value.IsUnknown() {
		return false
	}
	if value.IsNull() {
		return true
	}
	switch v := value.(type) {
	case types.Object:
		for _, a := range v.Attributes() {
			if !maintenanceKnown(a) {
				return false
			}
		}
	case types.Set:
		for _, a := range v.Elements() {
			if !maintenanceKnown(a) {
				return false
			}
		}
	}
	return true
}
func maintenanceString(ctx context.Context, object types.Object, key string) string {
	v, ok := object.Attributes()[key].(basetypes.StringValuable)
	if !ok {
		return ""
	}
	s, _ := v.ToStringValue(ctx)
	return s.ValueString()
}
func maintenanceAnytime() *mdb.MaintenanceWindows {
	return &mdb.MaintenanceWindows{Policy: &mdb.MaintenanceWindows_Anytime{Anytime: &mdb.AnytimeMaintenanceWindow{}}}
}

func expandMaintenance(ctx context.Context, modern, legacy types.Object, diagnostics *diag.Diagnostics) *mdb.MaintenanceWindows {
	if !modern.IsNull() {
		if !maintenanceKnown(modern) {
			diagnostics.AddError("Unknown maintenance schedule", "maintenance_windows must be known before sending it to the API")
			return nil
		}
		result, err := expandModernMaintenance(ctx, modern)
		if err != nil {
			diagnostics.AddError("Invalid maintenance schedule", err.Error())
		}
		return result
	}
	if legacy.IsNull() || legacy.IsUnknown() {
		return nil
	}
	if !maintenanceKnown(legacy) {
		diagnostics.AddError("Unknown maintenance window", "maintenance_window must be known before sending it to the API")
		return nil
	}
	switch maintenanceString(ctx, legacy, "type") {
	case "ANYTIME":
		return maintenanceAnytime()
	case "WEEKLY":
		day := maintenanceDay(maintenanceString(ctx, legacy, "day"))
		hour := legacy.Attributes()["hour"].(types.Int64).ValueInt64()
		if day == 0 || hour < 1 || hour > 24 {
			diagnostics.AddError("Invalid maintenance window", "WEEKLY requires a weekday and hour between 1 and 24")
			return nil
		}
		return &mdb.MaintenanceWindows{Policy: &mdb.MaintenanceWindows_WeeklyMaintenanceSchedule{WeeklyMaintenanceSchedule: &mdb.WeeklyMaintenanceSchedule{Slots: []*mdb.MaintenanceWindowSlot{{Day: day, StartTime: &timeofday.TimeOfDay{Hours: int32(hour - 1)}, Duration: durationpb.New(time.Hour), AllowTemporaryUnavailability: true}}}}}
	default:
		diagnostics.AddError("Invalid maintenance window", "type must be ANYTIME or WEEKLY")
		return nil
	}
}
func maintenanceDay(s string) dayofweek.DayOfWeek {
	for i, d := range maintenanceDays {
		if d == s {
			return dayofweek.DayOfWeek(i + 1)
		}
	}
	return 0
}

func expandModernMaintenance(ctx context.Context, object types.Object) (*mdb.MaintenanceWindows, error) {
	policy := maintenanceString(ctx, object, "type")
	slots := object.Attributes()["slot"].(types.Set).Elements()
	if policy == "ANYTIME" {
		if len(slots) > 0 {
			return nil, fmt.Errorf("ANYTIME must not contain slots")
		}
		return maintenanceAnytime(), nil
	}
	if policy != "WEEKLY" || len(slots) == 0 {
		return nil, fmt.Errorf("WEEKLY requires at least one slot")
	}
	result := &mdb.WeeklyMaintenanceSchedule{}
	type interval struct{ start, end time.Duration }
	intervals := []interval{}
	allow := false
	for _, raw := range slots {
		s := raw.(types.Object)
		day := maintenanceDay(maintenanceString(ctx, s, "day"))
		start := maintenanceString(ctx, s, "start_time")
		t, err := time.Parse("15:04:05", start)
		if day == 0 || err != nil || t.Format("15:04:05") != start || t.Second() != 0 {
			return nil, fmt.Errorf("slot requires a weekday and UTC start_time in HH:MM:00 format")
		}
		duration, err := time.ParseDuration(maintenanceString(ctx, s, "duration"))
		if err != nil || duration < time.Hour || duration > 24*time.Hour || duration%time.Minute != 0 {
			return nil, fmt.Errorf("slot duration must be between 1h and 24h with minute precision")
		}
		unavailable := s.Attributes()["allow_temporary_unavailability"].(types.Bool).ValueBool()
		allow = allow || unavailable
		result.Slots = append(result.Slots, &mdb.MaintenanceWindowSlot{Day: day, StartTime: &timeofday.TimeOfDay{Hours: int32(t.Hour()), Minutes: int32(t.Minute())}, Duration: durationpb.New(duration), AllowTemporaryUnavailability: unavailable})
		begin := time.Duration(day-1)*24*time.Hour + time.Duration(t.Hour())*time.Hour + time.Duration(t.Minute())*time.Minute
		intervals = append(intervals, interval{begin, begin + duration})
	}
	if !allow {
		return nil, fmt.Errorf("at least one weekly slot must allow temporary unavailability")
	}
	sort.Slice(intervals, func(i, j int) bool { return intervals[i].start < intervals[j].start })
	for i := 1; i < len(intervals); i++ {
		if intervals[i].start < intervals[i-1].end {
			return nil, fmt.Errorf("weekly slots must not overlap")
		}
	}
	if intervals[len(intervals)-1].end > intervals[0].start+7*24*time.Hour {
		return nil, fmt.Errorf("weekly slots must not overlap across the end of the week")
	}
	// Canonical ordering makes API equality independent of Terraform set order.
	sort.Slice(result.Slots, func(i, j int) bool {
		a, b := result.Slots[i], result.Slots[j]
		if a.Day != b.Day {
			return a.Day < b.Day
		}
		if a.StartTime.Hours != b.StartTime.Hours {
			return a.StartTime.Hours < b.StartTime.Hours
		}
		return a.StartTime.Minutes < b.StartTime.Minutes
	})
	return &mdb.MaintenanceWindows{Policy: &mdb.MaintenanceWindows_WeeklyMaintenanceSchedule{WeeklyMaintenanceSchedule: result}}, nil
}

func flattenMaintenance(ctx context.Context, state *Cluster, w *mdb.MaintenanceWindows, diagnostics *diag.Diagnostics) {
	legacy := types.ObjectNull(mdbcommon.MaintenanceWindowType.AttrTypes)
	modern := types.ObjectNull(maintenanceTypes)
	if w == nil {
		state.MaintenanceWindow, state.MaintenanceWindows = legacy, modern
		return
	}
	attrs := map[string]attr.Value{"type": types.StringValue("ANYTIME"), "slot": types.SetNull(types.ObjectType{AttrTypes: maintenanceSlotTypes})}
	switch p := w.GetPolicy().(type) {
	case *mdb.MaintenanceWindows_Anytime:
		if !state.MaintenanceWindows.IsNull() && !state.MaintenanceWindows.IsUnknown() {
			if slots, ok := state.MaintenanceWindows.Attributes()["slot"].(types.Set); ok && !slots.IsNull() && !slots.IsUnknown() && len(slots.Elements()) == 0 {
				attrs["slot"] = slots
			}
		}
		legacy = types.ObjectValueMust(mdbcommon.MaintenanceWindowType.AttrTypes, map[string]attr.Value{"type": types.StringValue("ANYTIME"), "day": types.StringNull(), "hour": types.Int64Null()})
	case *mdb.MaintenanceWindows_WeeklyMaintenanceSchedule:
		attrs["type"] = types.StringValue("WEEKLY")
		slots := []attr.Value{}
		for _, s := range p.WeeklyMaintenanceSchedule.GetSlots() {
			t := s.GetStartTime()
			d := s.GetDuration()
			if s.GetDay() < 1 || s.GetDay() > 7 || t == nil || d == nil || d.CheckValid() != nil || t.Hours < 0 || t.Hours > 23 || t.Minutes < 0 || t.Minutes > 59 || t.Seconds != 0 || t.Nanos != 0 {
				diagnostics.AddError("Invalid maintenance schedule", "API returned an invalid maintenance slot")
				return
			}
			slot := types.ObjectValueMust(maintenanceSlotTypes, map[string]attr.Value{
				"day":                            types.StringValue(maintenanceDays[int(s.Day)-1]),
				"start_time":                     types.StringValue(fmt.Sprintf("%02d:%02d:00", t.Hours, t.Minutes)),
				"duration":                       types.StringValue(d.AsDuration().String()),
				"allow_temporary_unavailability": types.BoolValue(s.AllowTemporaryUnavailability),
			})
			slots = append(slots, preserveMaintenanceSlotDuration(ctx, state.MaintenanceWindows, slot))
		}
		attrs["slot"] = types.SetValueMust(types.ObjectType{AttrTypes: maintenanceSlotTypes}, slots)
		if len(slots) == 1 {
			s := p.WeeklyMaintenanceSchedule.Slots[0]
			if s.StartTime.Minutes == 0 && s.Duration.AsDuration() == time.Hour && s.AllowTemporaryUnavailability {
				legacy = types.ObjectValueMust(mdbcommon.MaintenanceWindowType.AttrTypes, map[string]attr.Value{"type": types.StringValue("WEEKLY"), "day": types.StringValue(maintenanceDays[int(s.Day)-1]), "hour": types.Int64Value(int64(s.StartTime.Hours) + 1)})
			}
		}
	default:
		diagnostics.AddError("Invalid maintenance schedule", "API returned an unsupported maintenance policy")
		return
	}
	modern = types.ObjectValueMust(maintenanceTypes, attrs)
	// A selected modern representation stays modern. Import/default use the
	// legacy representation only when it represents the complete API schedule.
	if !state.MaintenanceWindows.IsNull() || legacy.IsNull() {
		state.MaintenanceWindow = types.ObjectNull(mdbcommon.MaintenanceWindowType.AttrTypes)
		state.MaintenanceWindows = modern
	} else {
		state.MaintenanceWindow = legacy
		state.MaintenanceWindows = types.ObjectNull(maintenanceTypes)
	}
}

func modifyMaintenancePlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var modern, legacy types.Object
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("maintenance_windows"), &modern)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("maintenance_window"), &legacy)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !modern.IsNull() {
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("maintenance_window"), types.ObjectNull(mdbcommon.MaintenanceWindowType.AttrTypes))...)
		return
	}
	if !legacy.IsNull() {
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("maintenance_windows"), types.ObjectNull(maintenanceTypes))...)
		return
	}
	if !req.State.Raw.IsNull() {
		resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("maintenance_windows"), &modern)...)
		resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("maintenance_window"), &legacy)...)
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("maintenance_windows"), modern)...)
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("maintenance_window"), legacy)...)
	} else {
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("maintenance_windows"), types.ObjectNull(maintenanceTypes))...)
	}
}

func validateMaintenanceConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var modern, legacy types.Object
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("maintenance_windows"), &modern)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("maintenance_window"), &legacy)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !modern.IsNull() && !legacy.IsNull() {
		resp.Diagnostics.AddError("Conflicting maintenance configuration", "Specify only one of maintenance_window and maintenance_windows")
		return
	}
	if !modern.IsNull() && maintenanceKnown(modern) {
		if _, err := expandModernMaintenance(ctx, modern); err != nil {
			resp.Diagnostics.AddAttributeError(path.Root("maintenance_windows"), "Invalid maintenance schedule", err.Error())
		}
	}
	var threshold types.Int64
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("config").AtName("disk_size_autoscaling").AtName("planned_usage_threshold"), &threshold)...)
	if threshold.IsNull() || threshold.IsUnknown() || threshold.ValueInt64() == 0 {
		return
	}
	selected := legacy
	if !modern.IsNull() {
		selected = modern
	}
	if !maintenanceKnown(selected) {
		return
	}
	if selected.IsNull() || maintenanceString(ctx, selected, "type") != "WEEKLY" {
		resp.Diagnostics.AddAttributeError(path.Root("config").AtName("disk_size_autoscaling").AtName("planned_usage_threshold"), "Missing weekly maintenance schedule", "A nonzero planned_usage_threshold requires a WEEKLY maintenance_window or maintenance_windows")
	}
}

// Preserve a duration's spelling only for the same complete slot. A scalar
// semantic-equality type underneath a Set can match the duration of a different
// slot, even when its day, start time, or unavailability flag differs.
func preserveMaintenanceSlotDuration(ctx context.Context, prior types.Object, slot types.Object) types.Object {
	if prior.IsNull() || prior.IsUnknown() {
		return slot
	}
	slots, ok := prior.Attributes()["slot"].(types.Set)
	if !ok || slots.IsNull() || slots.IsUnknown() {
		return slot
	}
	attrs := slot.Attributes()
	duration, err := time.ParseDuration(maintenanceString(ctx, slot, "duration"))
	if err != nil {
		return slot
	}
	for _, value := range slots.Elements() {
		previous, ok := value.(types.Object)
		if !ok || previous.IsNull() || !maintenanceKnown(previous) {
			continue
		}
		previousAttrs := previous.Attributes()
		if !attrs["day"].Equal(previousAttrs["day"]) || !attrs["start_time"].Equal(previousAttrs["start_time"]) || !attrs["allow_temporary_unavailability"].Equal(previousAttrs["allow_temporary_unavailability"]) {
			continue
		}
		previousDuration, err := time.ParseDuration(maintenanceString(ctx, previous, "duration"))
		if err == nil && previousDuration == duration {
			attrs["duration"] = previousAttrs["duration"]
			return types.ObjectValueMust(maintenanceSlotTypes, attrs)
		}
	}
	return slot
}

// Nested sets have no stable element index. Use only the known duration and
// prior schedule to choose a spelling, so resolving another unknown attribute
// cannot change a value that was already known in the initial plan.
type maintenanceDurationPlanModifier struct{}

func (maintenanceDurationPlanModifier) Description(context.Context) string {
	return "Preserves an unambiguous prior spelling of an equivalent duration."
}
func (m maintenanceDurationPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}
func (maintenanceDurationPlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.State.Raw.IsNull() || req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}
	duration, err := time.ParseDuration(req.PlanValue.ValueString())
	if err != nil {
		return
	}
	var prior types.Object
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("maintenance_windows"), &prior)...)
	if resp.Diagnostics.HasError() || prior.IsNull() || !maintenanceKnown(prior) {
		return
	}
	slots, ok := prior.Attributes()["slot"].(types.Set)
	if !ok {
		return
	}
	var spelling string
	for _, value := range slots.Elements() {
		slot, ok := value.(types.Object)
		if !ok || slot.IsNull() {
			continue
		}
		candidate := maintenanceString(ctx, slot, "duration")
		previousDuration, err := time.ParseDuration(candidate)
		if err != nil || previousDuration != duration {
			continue
		}
		if spelling != "" && spelling != candidate {
			// With multiple spellings there is no stable duration-only choice.
			// Keep the configuration; apply may update only its state representation.
			return
		}
		spelling = candidate
	}
	if spelling != "" {
		resp.PlanValue = types.StringValue(spelling)
	}
}
