package yandex

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/hashicorp/go-cty/cty"
	ctyjson "github.com/hashicorp/go-cty/cty/json"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
	mdb "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

func pgMaintenanceTestResource() *schema.Resource {
	r := resourceYandexMDBPostgreSQLCluster()
	return &schema.Resource{
		Schema: map[string]*schema.Schema{"maintenance_window": r.Schema["maintenance_window"], "maintenance_windows": r.Schema["maintenance_windows"]},
		CustomizeDiff: func(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
			return customizePGMaintenanceWindowsDiff(d)
		},
	}
}

func pgMaintenanceRawConfig(t *testing.T, r *schema.Resource, config map[string]interface{}) cty.Value {
	t.Helper()
	b, err := json.Marshal(config)
	require.NoError(t, err)
	v, err := ctyjson.Unmarshal(b, r.CoreConfigSchema().ImpliedType())
	require.NoError(t, err)
	return v
}

func pgMaintenanceTestData(t *testing.T, r *schema.Resource, config map[string]interface{}) *schema.ResourceData {
	t.Helper()
	d := schema.TestResourceDataRaw(t, r.Schema, config)
	d.SetId("pg-test")
	state := d.State()
	state.RawConfig = pgMaintenanceRawConfig(t, r, config)
	return r.Data(state)
}

func pgLegacyConfig(hour int) map[string]interface{} {
	return map[string]interface{}{"maintenance_window": []interface{}{map[string]interface{}{"type": "WEEKLY", "day": "MON", "hour": hour}}}
}

func pgModernConfig() map[string]interface{} {
	return map[string]interface{}{"maintenance_windows": []interface{}{map[string]interface{}{"type": "WEEKLY", "slot": []interface{}{
		map[string]interface{}{"day": "MON", "start_time": "02:30:00", "duration": "3h", "allow_temporary_unavailability": true},
		map[string]interface{}{"day": "THU", "start_time": "01:30:00", "duration": "2h", "allow_temporary_unavailability": false},
	}}}}
}

func TestPGMaintenanceAPILegacyConversion(t *testing.T) {
	r := pgMaintenanceTestResource()
	for _, hour := range []int{1, 3, 24} {
		d := pgMaintenanceTestData(t, r, pgLegacyConfig(hour))
		w, err := expandPGMaintenanceWindows(d)
		require.NoError(t, err)
		slots := w.GetWeeklyMaintenanceSchedule().GetSlots()
		require.Len(t, slots, 1)
		require.Equal(t, int32(hour-1), slots[0].GetStartTime().GetHours())
		require.Equal(t, time.Hour, slots[0].GetDuration().AsDuration())
		require.True(t, slots[0].GetAllowTemporaryUnavailability())
		require.NoError(t, setPGMaintenanceWindowsState(d, w, false))
		require.Equal(t, hour, d.Get("maintenance_window.0.hour"))
		require.Empty(t, d.Get("maintenance_windows"))
		state := d.State()
		state.RawConfig = pgMaintenanceRawConfig(t, r, pgLegacyConfig(hour))
		diff, err := r.Diff(context.Background(), state, terraform.NewResourceConfigRaw(pgLegacyConfig(hour)), nil)
		require.NoError(t, err)
		require.True(t, diff == nil || diff.Empty(), "legacy upgrade must not produce a diff: %#v", diff)
	}
	config := map[string]interface{}{"maintenance_window": []interface{}{map[string]interface{}{"type": "ANYTIME"}}}
	d := pgMaintenanceTestData(t, r, config)
	w, err := expandPGMaintenanceWindows(d)
	require.NoError(t, err)
	require.NotNil(t, w.GetAnytime())
	require.NoError(t, setPGMaintenanceWindowsState(d, w, false))
	require.Equal(t, "ANYTIME", d.Get("maintenance_window.0.type"))
}

func TestPGMaintenanceAPIModernRoundtripAndDrift(t *testing.T) {
	r := pgMaintenanceTestResource()
	d := pgMaintenanceTestData(t, r, pgModernConfig())
	w, err := expandPGMaintenanceWindows(d)
	require.NoError(t, err)
	require.Len(t, w.GetWeeklyMaintenanceSchedule().GetSlots(), 2)
	require.NoError(t, setPGMaintenanceWindowsState(d, w, false))
	got, err := expandPGMaintenanceWindows(d)
	require.NoError(t, err)
	require.True(t, proto.Equal(w, got))
	state := d.State()
	state.RawConfig = pgMaintenanceRawConfig(t, r, pgModernConfig())
	diff, err := r.Diff(context.Background(), state, terraform.NewResourceConfigRaw(pgModernConfig()), nil)
	require.NoError(t, err)
	require.True(t, diff == nil || diff.Empty(), "%#v", diff)

	legacy := pgMaintenanceTestData(t, r, pgLegacyConfig(3))
	require.NoError(t, setPGMaintenanceWindowsState(legacy, w, false))
	require.Empty(t, legacy.Get("maintenance_window"))
	require.Equal(t, 2, legacy.Get("maintenance_windows.0.slot").(*schema.Set).Len())
	state = legacy.State()
	state.RawConfig = pgMaintenanceRawConfig(t, r, pgLegacyConfig(3))
	diff, err = r.Diff(context.Background(), state, terraform.NewResourceConfigRaw(pgLegacyConfig(3)), nil)
	require.NoError(t, err)
	require.False(t, diff.Empty(), "external multi-slot drift must remain visible")
}

func TestPGMaintenanceAPIImportAndUnmanaged(t *testing.T) {
	r := pgMaintenanceTestResource()
	legacy := pgMaintenanceTestData(t, r, pgLegacyConfig(24))
	w, err := expandPGMaintenanceWindows(legacy)
	require.NoError(t, err)
	imported := r.Data(&terraform.InstanceState{ID: "pg-test"})
	require.NoError(t, setPGMaintenanceWindowsState(imported, w, false))
	require.Equal(t, 24, imported.Get("maintenance_window.0.hour"))
	modern := pgMaintenanceTestData(t, r, pgModernConfig())
	w, err = expandPGMaintenanceWindows(modern)
	require.NoError(t, err)
	imported = r.Data(&terraform.InstanceState{ID: "pg-test"})
	require.NoError(t, setPGMaintenanceWindowsState(imported, w, false))
	require.Equal(t, 2, imported.Get("maintenance_windows.0.slot").(*schema.Set).Len())
	state := imported.State()
	state.RawConfig = pgMaintenanceRawConfig(t, r, map[string]interface{}{})
	diff, err := r.Diff(context.Background(), state, terraform.NewResourceConfigRaw(map[string]interface{}{}), nil)
	require.NoError(t, err)
	require.True(t, diff == nil || diff.Empty(), "omitting both blocks must leave schedule unmanaged")
	w, err = expandPGMaintenanceWindows(r.Data(state))
	require.NoError(t, err)
	require.Nil(t, w, "unmanaged schedule must not be sent on unrelated updates")
}

func TestPGMaintenanceAPISwitchRepresentations(t *testing.T) {
	r := pgMaintenanceTestResource()
	for _, configs := range [][2]map[string]interface{}{{pgLegacyConfig(1), pgModernConfig()}, {pgModernConfig(), pgLegacyConfig(24)}} {
		d := pgMaintenanceTestData(t, r, configs[0])
		w, err := expandPGMaintenanceWindows(d)
		require.NoError(t, err)
		require.NoError(t, setPGMaintenanceWindowsState(d, w, false))
		state := d.State()
		state.RawConfig = pgMaintenanceRawConfig(t, r, configs[1])
		diff, err := r.Diff(context.Background(), state, terraform.NewResourceConfigRaw(configs[1]), nil)
		require.NoError(t, err)
		require.False(t, diff.Empty())
		planned := state.MergeDiff(diff)
		planned.RawConfig = state.RawConfig
		pd := r.Data(planned)
		w, err = expandPGMaintenanceWindows(pd)
		require.NoError(t, err)
		require.NoError(t, setPGMaintenanceWindowsState(pd, w, false))
		finalState := pd.State()
		finalState.RawConfig = state.RawConfig
		finalDiff, err := r.Diff(context.Background(), finalState, terraform.NewResourceConfigRaw(configs[1]), nil)
		require.NoError(t, err)
		require.True(t, finalDiff == nil || finalDiff.Empty(), "switch must converge: %#v", finalDiff)
	}
}

func TestPGMaintenanceAPIDataSource(t *testing.T) {
	r := dataSourceYandexMDBPostgreSQLCluster()
	require.NoError(t, r.InternalValidate(nil, false))
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{})
	w, err := expandPGMaintenanceWindows(pgMaintenanceTestData(t, pgMaintenanceTestResource(), pgModernConfig()))
	require.NoError(t, err)
	require.NoError(t, setPGMaintenanceWindowsState(d, w, true))
	require.Equal(t, 2, d.Get("maintenance_windows.0.slot").(*schema.Set).Len())
	require.Empty(t, d.Get("maintenance_window"))
	require.NoError(t, setPGMaintenanceWindowsState(d, &mdb.MaintenanceWindows{Policy: &mdb.MaintenanceWindows_Anytime{Anytime: &mdb.AnytimeMaintenanceWindow{}}}, true))
	require.Equal(t, "ANYTIME", d.Get("maintenance_window.0.type"))
	require.Equal(t, "ANYTIME", d.Get("maintenance_windows.0.type"))
}

func TestPGMaintenanceAPISingleSlotDrift(t *testing.T) {
	r := pgMaintenanceTestResource()
	base, err := expandPGMaintenanceWindows(pgMaintenanceTestData(t, r, pgLegacyConfig(3)))
	require.NoError(t, err)
	for _, mutate := range []func(*mdb.MaintenanceWindowSlot){
		func(s *mdb.MaintenanceWindowSlot) { s.Duration = durationpb.New(2 * time.Hour) },
		func(s *mdb.MaintenanceWindowSlot) { s.AllowTemporaryUnavailability = false },
		func(s *mdb.MaintenanceWindowSlot) { s.StartTime.Minutes = 30 },
	} {
		w := proto.Clone(base).(*mdb.MaintenanceWindows)
		mutate(w.GetWeeklyMaintenanceSchedule().Slots[0])
		legacy, exact := flattenPGLegacyMaintenanceWindows(w)
		require.False(t, exact)
		require.Nil(t, legacy)
		d := pgMaintenanceTestData(t, r, pgLegacyConfig(3))
		require.NoError(t, setPGMaintenanceWindowsState(d, w, false))
		require.Empty(t, d.Get("maintenance_window"))
		require.Equal(t, 1, d.Get("maintenance_windows.0.slot").(*schema.Set).Len())
	}
}

func TestPGMaintenanceAPIUnknownAndValidation(t *testing.T) {
	require.NoError(t, resourceYandexMDBPostgreSQLCluster().InternalValidate(nil, true))
	r := pgMaintenanceTestResource()
	conflicting := pgModernConfig()
	conflicting["maintenance_window"] = pgLegacyConfig(3)["maintenance_window"]
	require.True(t, r.Validate(terraform.NewResourceConfigRaw(conflicting)).HasError())
	for _, field := range []string{"day", "start_time", "duration", "allow_temporary_unavailability"} {
		config := pgModernConfig()
		// A false-only schedule is invalid once known, but an unknown value
		// must not be treated as false (or as a zero/empty slot value).
		slots := config["maintenance_windows"].([]interface{})[0].(map[string]interface{})["slot"].([]interface{})
		slots[0].(map[string]interface{})["allow_temporary_unavailability"] = false
		raw := pgMaintenanceRawConfig(t, r, config).AsValueMap()
		block := raw["maintenance_windows"].Index(cty.NumberIntVal(0)).AsValueMap()
		values := block["slot"].AsValueSlice()
		slot := values[0].AsValueMap()
		slot[field] = cty.UnknownVal(slot[field].Type())
		values[0] = cty.ObjectVal(slot)
		block["slot"] = cty.SetVal(values)
		raw["maintenance_windows"] = cty.ListVal([]cty.Value{cty.ObjectVal(block)})
		rawConfig := cty.ObjectVal(raw)
		state := &terraform.InstanceState{ID: "pg-test", Attributes: map[string]string{}, RawConfig: rawConfig}
		_, err := r.Diff(context.Background(), state, terraform.NewResourceConfigShimmed(rawConfig, r.CoreConfigSchema()), nil)
		require.NoError(t, err, "unknown %s must defer validation", field)
	}
	invalid := pgModernConfig()
	invalid["maintenance_windows"].([]interface{})[0].(map[string]interface{})["slot"].([]interface{})[0].(map[string]interface{})["allow_temporary_unavailability"] = false
	state := &terraform.InstanceState{ID: "pg-test", RawConfig: pgMaintenanceRawConfig(t, r, invalid)}
	_, err := r.Diff(context.Background(), state, terraform.NewResourceConfigRaw(invalid), nil)
	require.ErrorContains(t, err, "at least one slot")
}

func TestPGMaintenanceAPIRequests(t *testing.T) {
	r := resourceYandexMDBPostgreSQLCluster()
	for _, schedule := range []map[string]interface{}{pgLegacyConfig(24), pgModernConfig()} {
		config := map[string]interface{}{
			"name": "maintenance-test", "folder_id": "folder", "network_id": "network", "environment": "PRODUCTION",
			"config": []interface{}{map[string]interface{}{"version": "15", "resources": []interface{}{map[string]interface{}{"resource_preset_id": "s2.micro", "disk_type_id": "network-ssd", "disk_size": 16}}}},
			"host":   []interface{}{map[string]interface{}{"zone": "ru-central1-a", "subnet_id": "subnet"}},
		}
		for k, v := range schedule {
			config[k] = v
		}
		d := pgMaintenanceTestData(t, r, config)
		create, err := prepareCreatePostgreSQLRequest(d, &Config{})
		require.NoError(t, err)
		require.Nil(t, create.MaintenanceWindow)
		require.NotNil(t, create.MaintenanceWindows)
		restore, err := prepareRestorePostgreSQLRequest(d, create, "backup")
		require.NoError(t, err)
		require.Nil(t, restore.MaintenanceWindow)
		require.True(t, proto.Equal(create.MaintenanceWindows, restore.MaintenanceWindows))
		update, err := prepareUpdatePostgreSQLClusterParamsRequest(d, &Config{})
		require.NoError(t, err)
		require.Nil(t, update.MaintenanceWindow)
		require.True(t, proto.Equal(create.MaintenanceWindows, update.MaintenanceWindows))
		require.NotContains(t, update.GetUpdateMask().GetPaths(), "maintenance_window")

		for _, managed := range []bool{true, false} {
			before := d.State()
			desired := make(map[string]interface{}, len(config))
			for key, value := range config {
				desired[key] = value
			}
			delete(desired, "maintenance_window")
			delete(desired, "maintenance_windows")
			desired["name"] = "renamed-cluster"
			if managed {
				desired["maintenance_windows"] = []interface{}{map[string]interface{}{"type": "ANYTIME"}}
			}
			before.RawConfig = pgMaintenanceRawConfig(t, r, desired)
			diff, err := r.Diff(context.Background(), before, terraform.NewResourceConfigRaw(desired), nil)
			require.NoError(t, err)
			called := false
			r.Update = func(data *schema.ResourceData, _ interface{}) error {
				called = true
				req, err := prepareUpdatePostgreSQLClusterParamsRequest(data, &Config{})
				if err != nil {
					return err
				}
				paths := req.GetUpdateMask().GetPaths()
				require.NotContains(t, paths, "maintenance_window")
				count := 0
				for _, p := range paths {
					if p == "maintenance_windows" {
						count++
					}
				}
				if managed {
					require.Equal(t, 1, count, "schedule update must send the new mask path exactly once")
					require.NotNil(t, req.GetMaintenanceWindows().GetAnytime())
				} else {
					require.Zero(t, count, "unmanaged schedule must not be updated")
					require.Nil(t, req.MaintenanceWindows)
				}
				return nil
			}
			_, diagnostics := r.Apply(context.Background(), before, diff, nil)
			require.False(t, diagnostics.HasError(), "%v", diagnostics)
			require.True(t, called)
		}
	}
}

func TestPGMaintenanceWindowsSetSemantics(t *testing.T) {
	resource := pgMaintenanceTestResource()
	makeConfig := func(reverse bool, duration string) map[string]interface{} {
		slots := []interface{}{
			map[string]interface{}{"day": "MON", "start_time": "02:00:00", "duration": duration, "allow_temporary_unavailability": true},
			map[string]interface{}{"day": "THU", "start_time": "01:30:00", "duration": "2h", "allow_temporary_unavailability": false},
		}
		if reverse {
			slots[0], slots[1] = slots[1], slots[0]
		}
		return map[string]interface{}{
			"maintenance_windows": []interface{}{map[string]interface{}{"type": "WEEKLY", "slot": slots}},
		}
	}
	for _, oldDuration := range []string{"3h", "3h0m0s", "180m"} {
		d := schema.TestResourceDataRaw(t, resource.Schema, makeConfig(false, oldDuration))
		d.SetId("maintenance-test")
		// Model refresh clearing the unselected computed legacy representation.
		if err := d.Set("maintenance_window", []interface{}{}); err != nil {
			t.Fatal(err)
		}
		for _, config := range []map[string]interface{}{makeConfig(true, "3h"), makeConfig(false, "180m"), makeConfig(false, "10800s")} {
			diff, err := resource.Diff(context.Background(), d.State(), terraform.NewResourceConfigRaw(config), nil)
			if err != nil {
				t.Fatal(err)
			}
			if diff != nil && !diff.Empty() {
				t.Fatalf("equivalent schedules must not produce a diff: %#v", diff)
			}
		}
		diff, err := resource.Diff(context.Background(), d.State(), terraform.NewResourceConfigRaw(makeConfig(false, "4h")), nil)
		if err != nil {
			t.Fatal(err)
		}
		if diff == nil || diff.Empty() {
			t.Fatal("changing the duration must produce a diff")
		}
		changedAvailability := makeConfig(false, "3h")
		changedAvailability["maintenance_windows"].([]interface{})[0].(map[string]interface{})["slot"].([]interface{})[1].(map[string]interface{})["allow_temporary_unavailability"] = true
		diff, err = resource.Diff(context.Background(), d.State(), terraform.NewResourceConfigRaw(changedAvailability), nil)
		if err != nil {
			t.Fatal(err)
		}
		if diff == nil || diff.Empty() {
			t.Fatal("changing temporary unavailability permission must produce a diff")
		}
	}
}

func TestPGMaintenanceWindowsValidation(t *testing.T) {
	slot := func(day, start, duration string, unavailable bool) map[string]interface{} {
		return map[string]interface{}{"day": day, "start_time": start, "duration": duration, "allow_temporary_unavailability": unavailable}
	}
	cases := []struct {
		name   string
		policy string
		slots  []interface{}
		valid  bool
	}{
		{"anytime", "ANYTIME", nil, true},
		{"empty weekly", "WEEKLY", nil, false},
		{"anytime with slots", "ANYTIME", []interface{}{slot("MON", "02:00:00", "1h", true)}, false},
		{"single", "WEEKLY", []interface{}{slot("MON", "02:00:00", "3h", true)}, true},
		{"all false", "WEEKLY", []interface{}{slot("MON", "02:00:00", "3h", false)}, false},
		{"touching", "WEEKLY", []interface{}{slot("MON", "02:00:00", "3h", true), slot("MON", "05:00:00", "1h", false)}, true},
		{"overlap", "WEEKLY", []interface{}{slot("MON", "02:00:00", "3h", true), slot("MON", "04:59:00", "1h", false)}, false},
		{"next day overlap", "WEEKLY", []interface{}{slot("MON", "23:00:00", "3h", true), slot("TUE", "01:00:00", "1h", false)}, false},
		{"wrap overlap", "WEEKLY", []interface{}{slot("SUN", "23:00:00", "3h", true), slot("MON", "01:00:00", "1h", false)}, false},
		{"wrap touching", "WEEKLY", []interface{}{slot("SUN", "23:00:00", "3h", true), slot("MON", "02:00:00", "1h", false)}, true},
		{"single wrap", "WEEKLY", []interface{}{slot("SUN", "23:00:00", "24h", true)}, true},
		{"short", "WEEKLY", []interface{}{slot("MON", "02:00:00", "59m", true)}, false},
		{"long", "WEEKLY", []interface{}{slot("MON", "02:00:00", "24h1m", true)}, false},
		{"fractional duration", "WEEKLY", []interface{}{slot("MON", "02:00:00", "1h1s", true)}, false},
		{"seconds", "WEEKLY", []interface{}{slot("MON", "02:00:01", "1h", true)}, false},
		{"non padded", "WEEKLY", []interface{}{slot("MON", "2:00:00", "1h", true)}, false},
		{"24 hours", "WEEKLY", []interface{}{slot("MON", "24:00:00", "1h", true)}, false},
	}
	slotSchema := pgMaintenanceWindowsSchema().Elem.(*schema.Resource).Schema["slot"].Elem.(*schema.Resource)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := map[string]interface{}{"type": tc.policy, "slot": schema.NewSet(schema.HashResource(slotSchema), tc.slots)}
			err := validatePGMaintenanceWindows(w)
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%v, error=%v", tc.valid, err)
			}
		})
	}
}
