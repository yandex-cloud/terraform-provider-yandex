package yandex

import (
	"fmt"
	"sort"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	mdb "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
	"google.golang.org/genproto/googleapis/type/dayofweek"
	"google.golang.org/genproto/googleapis/type/timeofday"
	"google.golang.org/protobuf/types/known/durationpb"
)

// rawPGMaintenanceForm distinguishes configured blocks from Optional+Computed
// values inherited from state. Unknown blocks still select their representation.
func rawPGMaintenanceForm(d mdbcommon.RawConfigProvider) string {
	raw := d.GetRawConfig()
	if raw.IsNull() || !raw.IsKnown() {
		return ""
	}
	for _, name := range []string{"maintenance_windows", "maintenance_window"} {
		if !raw.Type().HasAttribute(name) {
			continue
		}
		v := raw.GetAttr(name)
		if !v.IsNull() && (!v.IsKnown() || v.LengthInt() > 0) {
			return name
		}
	}
	return ""
}

func customizePGMaintenanceWindowsDiff(d *schema.ResourceDiff) error {
	form := rawPGMaintenanceForm(d)
	if form == "" {
		return nil
	}
	other := "maintenance_window"
	if form == other {
		other = "maintenance_windows"
	}
	if err := d.SetNew(other, []interface{}{}); err != nil {
		return err
	}
	if form == "maintenance_windows" {
		value, known := mdbcommon.LookupRawConfigPath(d, form)
		if !known || !value.IsWhollyKnown() || !d.NewValueKnown(form) {
			return nil
		}
		if blocks := d.Get(form).([]interface{}); len(blocks) > 0 {
			return validatePGMaintenanceWindows(blocks[0].(map[string]interface{}))
		}
	}
	return nil
}

func expandPGMaintenanceWindows(d *schema.ResourceData) (*mdb.MaintenanceWindows, error) {
	form := rawPGMaintenanceForm(d)
	// ResourceData built by legacy SDK helpers has no raw configuration. This
	// fallback is also useful for restored state; explicit configuration wins.
	if form == "" {
		if !d.GetRawConfig().IsNull() {
			return nil, nil
		}
		if blocks := d.Get("maintenance_windows").([]interface{}); len(blocks) > 0 {
			form = "maintenance_windows"
		} else {
			form = "maintenance_window"
		}
	}
	blocks := d.Get(form).([]interface{})
	if len(blocks) == 0 {
		return nil, nil
	}
	w := blocks[0].(map[string]interface{})
	if form == "maintenance_windows" {
		if err := validatePGMaintenanceWindows(w); err != nil {
			return nil, err
		}
		if w["type"] == "ANYTIME" {
			return pgAnytimeMaintenanceWindows(), nil
		}
		slots := w["slot"].(*schema.Set).List()
		schedule := &mdb.WeeklyMaintenanceSchedule{}
		for _, raw := range slots {
			s := raw.(map[string]interface{})
			t, _ := time.Parse("15:04:05", s["start_time"].(string))
			duration, _ := time.ParseDuration(s["duration"].(string))
			day := 0
			for i, name := range pgMaintenanceDays {
				if name == s["day"] {
					day = i + 1
					break
				}
			}
			schedule.Slots = append(schedule.Slots, &mdb.MaintenanceWindowSlot{
				Day: dayofweek.DayOfWeek(day), StartTime: &timeofday.TimeOfDay{Hours: int32(t.Hour()), Minutes: int32(t.Minute())},
				Duration: durationpb.New(duration), AllowTemporaryUnavailability: s["allow_temporary_unavailability"].(bool),
			})
		}
		return &mdb.MaintenanceWindows{Policy: &mdb.MaintenanceWindows_WeeklyMaintenanceSchedule{WeeklyMaintenanceSchedule: schedule}}, nil
	}
	if w["type"] == "ANYTIME" {
		if w["day"].(string) != "" || w["hour"].(int) != 0 {
			return nil, fmt.Errorf("maintenance_window: ANYTIME must not specify day or hour")
		}
		return pgAnytimeMaintenanceWindows(), nil
	}
	day := 0
	for i, name := range pgMaintenanceDays {
		if name == w["day"] {
			day = i + 1
			break
		}
	}
	hour := w["hour"].(int)
	if w["type"] != "WEEKLY" || day == 0 || hour < 1 || hour > 24 {
		return nil, fmt.Errorf("maintenance_window: WEEKLY requires a weekday and hour between 1 and 24")
	}
	return &mdb.MaintenanceWindows{Policy: &mdb.MaintenanceWindows_WeeklyMaintenanceSchedule{WeeklyMaintenanceSchedule: &mdb.WeeklyMaintenanceSchedule{Slots: []*mdb.MaintenanceWindowSlot{{
		Day: dayofweek.DayOfWeek(day), StartTime: &timeofday.TimeOfDay{Hours: int32(hour - 1)},
		Duration: durationpb.New(time.Hour), AllowTemporaryUnavailability: true,
	}}}}}, nil
}

func pgAnytimeMaintenanceWindows() *mdb.MaintenanceWindows {
	return &mdb.MaintenanceWindows{Policy: &mdb.MaintenanceWindows_Anytime{Anytime: &mdb.AnytimeMaintenanceWindow{}}}
}

func flattenPGMaintenanceWindows(w *mdb.MaintenanceWindows) ([]interface{}, error) {
	if w == nil {
		return nil, nil
	}
	switch policy := w.GetPolicy().(type) {
	case *mdb.MaintenanceWindows_Anytime:
		return []interface{}{map[string]interface{}{"type": "ANYTIME", "slot": []interface{}{}}}, nil
	case *mdb.MaintenanceWindows_WeeklyMaintenanceSchedule:
		slots := make([]interface{}, 0, len(policy.WeeklyMaintenanceSchedule.GetSlots()))
		for _, s := range policy.WeeklyMaintenanceSchedule.GetSlots() {
			if s.GetDay() < 1 || s.GetDay() > 7 || s.GetStartTime() == nil || s.GetDuration() == nil {
				return nil, fmt.Errorf("invalid PostgreSQL maintenance slot returned by API")
			}
			if err := s.GetDuration().CheckValid(); err != nil {
				return nil, err
			}
			t := s.GetStartTime()
			if t.GetSeconds() != 0 || t.GetNanos() != 0 {
				return nil, fmt.Errorf("PostgreSQL maintenance start_time returned by API must have minute precision")
			}
			slots = append(slots, map[string]interface{}{
				"day": pgMaintenanceDays[int(s.GetDay())-1], "start_time": fmt.Sprintf("%02d:%02d:00", t.GetHours(), t.GetMinutes()),
				"duration": s.GetDuration().AsDuration().String(), "allow_temporary_unavailability": s.GetAllowTemporaryUnavailability(),
			})
		}
		return []interface{}{map[string]interface{}{"type": "WEEKLY", "slot": slots}}, nil
	default:
		return nil, fmt.Errorf("unsupported PostgreSQL maintenance policy type")
	}
}

// Exact projection only: a lossy first-slot projection would hide external drift.
func flattenPGLegacyMaintenanceWindows(w *mdb.MaintenanceWindows) ([]interface{}, bool) {
	if w == nil {
		return nil, true
	}
	switch policy := w.GetPolicy().(type) {
	case *mdb.MaintenanceWindows_Anytime:
		return []interface{}{map[string]interface{}{"type": "ANYTIME"}}, true
	case *mdb.MaintenanceWindows_WeeklyMaintenanceSchedule:
		slots := policy.WeeklyMaintenanceSchedule.GetSlots()
		if len(slots) != 1 {
			return nil, false
		}
		s := slots[0]
		t := s.GetStartTime()
		if s.GetDay() < 1 || s.GetDay() > 7 || t == nil || t.GetHours() < 0 || t.GetHours() > 23 || t.GetMinutes() != 0 || t.GetSeconds() != 0 || t.GetNanos() != 0 || s.GetDuration() == nil || s.GetDuration().AsDuration() != time.Hour || !s.GetAllowTemporaryUnavailability() {
			return nil, false
		}
		return []interface{}{map[string]interface{}{"type": "WEEKLY", "day": pgMaintenanceDays[int(s.GetDay())-1], "hour": int(t.GetHours()) + 1}}, true
	}
	return nil, false
}

func setPGMaintenanceWindowsState(d *schema.ResourceData, w *mdb.MaintenanceWindows, datasource bool) error {
	modern, err := flattenPGMaintenanceWindows(w)
	if err != nil {
		return err
	}
	legacy, exact := flattenPGLegacyMaintenanceWindows(w)
	form := rawPGMaintenanceForm(d)
	if form == "" && !datasource {
		if old := d.Get("maintenance_window").([]interface{}); len(old) > 0 {
			form = "maintenance_window"
		} else if modernState := d.Get("maintenance_windows").([]interface{}); len(modernState) > 0 {
			form = "maintenance_windows"
		} else if exact {
			// Preserve historical import/default state for schedules the old
			// interface can express exactly. A new configuration selects the
			// modern representation on the next refresh.
			form = "maintenance_window"
		}
	}
	if datasource {
		if err := d.Set("maintenance_windows", modern); err != nil {
			return err
		}
		return d.Set("maintenance_window", legacy)
	}
	if form == "maintenance_window" && exact {
		if err := d.Set("maintenance_window", legacy); err != nil {
			return err
		}
		return d.Set("maintenance_windows", []interface{}{})
	}
	if err := d.Set("maintenance_window", []interface{}{}); err != nil {
		return err
	}
	return d.Set("maintenance_windows", modern)
}

func pgMaintenanceWindowsDataSourceSchema() *schema.Schema {
	s := pgMaintenanceWindowsSchema()
	var computed func(*schema.Schema)
	computed = func(s *schema.Schema) {
		s.Optional, s.Required, s.Computed = false, false, true
		s.Default, s.ValidateFunc, s.DiffSuppressFunc = nil, nil, nil
		s.ConflictsWith = nil
		s.MaxItems, s.MinItems = 0, 0
		if r, ok := s.Elem.(*schema.Resource); ok {
			for _, child := range r.Schema {
				computed(child)
			}
		}
	}
	computed(s)
	return s
}

var pgMaintenanceDays = []string{"MON", "TUE", "WED", "THU", "FRI", "SAT", "SUN"}

func pgMaintenanceWindowsSchema() *schema.Schema {
	result := &schema.Schema{
		Type: schema.TypeList, Optional: true, Computed: true, MaxItems: 1,
		Description:   "Maintenance schedule. Conflicts with maintenance_window. Times are UTC. To remove time restrictions, explicitly select ANYTIME.",
		ConflictsWith: []string{"maintenance_window"},
		Elem: &schema.Resource{Schema: map[string]*schema.Schema{
			"type": {
				Type: schema.TypeString, Required: true,
				ValidateFunc: validation.StringInSlice([]string{"ANYTIME", "WEEKLY"}, false),
				Description:  "ANYTIME permits maintenance at any time. WEEKLY requires at least one slot.",
			},
			"slot": {
				Type: schema.TypeSet, Optional: true,
				Description: "Non-overlapping weekly slots. At least one slot must allow temporary write unavailability. Slot order is insignificant.",
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"day": {Type: schema.TypeString, Required: true,
						ValidateFunc: validation.StringInSlice(pgMaintenanceDays, false),
						Description:  "UTC weekday: MON, TUE, WED, THU, FRI, SAT or SUN."},
					"start_time": {Type: schema.TypeString, Required: true,
						ValidateFunc: validatePGMaintenanceStartTime,
						Description:  "Start time in UTC, in HH:MM:00 format (minute precision)."},
					"duration": {Type: schema.TypeString, Required: true,
						ValidateFunc: validatePGMaintenanceDuration,
						DiffSuppressFunc: func(_ string, old, new string, _ *schema.ResourceData) bool {
							oldDuration, oldErr := time.ParseDuration(old)
							newDuration, newErr := time.ParseDuration(new)
							return oldErr == nil && newErr == nil && oldDuration == newDuration
						},
						Description: "Duration between 1h and 24h, with minute precision, for example 3h or 90m."},
					"allow_temporary_unavailability": {Type: schema.TypeBool, Optional: true, Default: false,
						Description: "Allow maintenance that may temporarily interrupt writes. Must be true for at least one weekly slot."},
				}},
			},
		}},
	}
	slots := result.Elem.(*schema.Resource).Schema["slot"]
	hashSlot := schema.HashResource(slots.Elem.(*schema.Resource))
	slots.Set = func(v interface{}) int {
		// Nested StateFunc does not canonicalize the identity of a TypeSet
		// element. Hash equivalent duration spellings identically explicitly.
		original := v.(map[string]interface{})
		canonical := make(map[string]interface{}, len(original))
		for key, value := range original {
			canonical[key] = value
		}
		if value, ok := original["duration"].(string); ok {
			if duration, err := time.ParseDuration(value); err == nil {
				canonical["duration"] = duration.String()
			}
		}
		return hashSlot(canonical)
	}
	return result
}

func validatePGMaintenanceStartTime(v interface{}, key string) ([]string, []error) {
	s := v.(string)
	t, err := time.Parse("15:04:05", s)
	if err != nil || t.Format("15:04:05") != s || t.Second() != 0 {
		return nil, []error{fmt.Errorf("%s must be a UTC time in HH:MM:00 format, between 00:00:00 and 23:59:00", key)}
	}
	return nil, nil
}

func validatePGMaintenanceDuration(v interface{}, key string) ([]string, []error) {
	d, err := time.ParseDuration(v.(string))
	if err != nil || d < time.Hour || d > 24*time.Hour || d%time.Minute != 0 {
		return nil, []error{fmt.Errorf("%s must be a duration between 1h and 24h with minute precision", key)}
	}
	return nil, nil
}

// Called only once all slot values are known; unknown computed expressions must
// not be mistaken for zero values while Terraform is planning.
func validatePGMaintenanceWindows(w map[string]interface{}) error {
	slots := w["slot"].(*schema.Set).List()
	if w["type"] == "ANYTIME" {
		if len(slots) != 0 {
			return fmt.Errorf("maintenance_windows: ANYTIME must not contain slots")
		}
		return nil
	}
	if len(slots) == 0 {
		return fmt.Errorf("maintenance_windows: WEEKLY requires at least one slot")
	}
	type interval struct{ start, end time.Duration }
	intervals := make([]interval, 0, len(slots))
	allowsUnavailability := false
	for _, raw := range slots {
		s := raw.(map[string]interface{})
		if _, errs := validatePGMaintenanceStartTime(s["start_time"], "start_time"); len(errs) != 0 {
			return errs[0]
		}
		if _, errs := validatePGMaintenanceDuration(s["duration"], "duration"); len(errs) != 0 {
			return errs[0]
		}
		day := -1
		for i, name := range pgMaintenanceDays {
			if name == s["day"] {
				day = i
				break
			}
		}
		if day < 0 {
			return fmt.Errorf("maintenance_windows: invalid weekday %q", s["day"])
		}
		t, _ := time.Parse("15:04:05", s["start_time"].(string))
		duration, _ := time.ParseDuration(s["duration"].(string))
		start := time.Duration(day)*24*time.Hour + time.Duration(t.Hour())*time.Hour + time.Duration(t.Minute())*time.Minute
		intervals = append(intervals, interval{start, start + duration})
		allowsUnavailability = allowsUnavailability || s["allow_temporary_unavailability"].(bool)
	}
	if !allowsUnavailability {
		return fmt.Errorf("maintenance_windows: at least one slot must allow temporary unavailability")
	}
	sort.Slice(intervals, func(i, j int) bool { return intervals[i].start < intervals[j].start })
	for i := 1; i < len(intervals); i++ {
		if intervals[i].start < intervals[i-1].end {
			return fmt.Errorf("maintenance_windows: weekly slots must not overlap")
		}
	}
	if intervals[len(intervals)-1].end > intervals[0].start+7*24*time.Hour {
		return fmt.Errorf("maintenance_windows: weekly slots must not overlap across the end of the week")
	}
	return nil
}
