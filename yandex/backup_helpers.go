package yandex

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	backuppb "github.com/yandex-cloud/go-genproto/yandex/cloud/backup/v1"
	sdkoperation "github.com/yandex-cloud/go-sdk/operation"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	resourceYandexBackupCompressionValues = []string{
		backuppb.PolicySettings_NORMAL.String(),
		backuppb.PolicySettings_HIGH.String(),
		backuppb.PolicySettings_MAX.String(),
		backuppb.PolicySettings_OFF.String(),
	}

	resourceYandexBackupPolicyFormatValues = []string{
		backuppb.Format_AUTO.String(),
		backuppb.Format_VERSION_11.String(),
		backuppb.Format_VERSION_12.String(),
	}

	resourceYandexBackupVSSProviderValues = []string{
		backuppb.PolicySettings_VolumeShadowCopyServiceSettings_NATIVE.String(),
		backuppb.PolicySettings_VolumeShadowCopyServiceSettings_TARGET_SYSTEM_DEFINED.String(),
	}

	resourceYandexBackupCBTValues = []string{
		backuppb.PolicySettings_DO_NOT_USE.String(),
		backuppb.PolicySettings_USE_IF_ENABLED.String(),
		backuppb.PolicySettings_ENABLE_AND_USE.String(),
	}

	resourceYandexBackupDayValues = []string{
		backuppb.PolicySettings_MONDAY.String(),
		backuppb.PolicySettings_TUESDAY.String(),
		backuppb.PolicySettings_WEDNESDAY.String(),
		backuppb.PolicySettings_THURSDAY.String(),
		backuppb.PolicySettings_FRIDAY.String(),
		backuppb.PolicySettings_SATURDAY.String(),
		backuppb.PolicySettings_SUNDAY.String(),
	}

	resourceYandexBackupSchedulingBackupSetTypeValues = []string{
		backuppb.PolicySettings_Scheduling_BackupSet_TYPE_AUTO.String(),
		backuppb.PolicySettings_Scheduling_BackupSet_TYPE_FULL.String(),
		backuppb.PolicySettings_Scheduling_BackupSet_TYPE_INCREMENTAL.String(),
		backuppb.PolicySettings_Scheduling_BackupSet_TYPE_DIFFERENTIAL.String(),
	}

	resourceYandexBackupTypeValues = []string{
		backuppb.PolicySettings_Scheduling_ALWAYS_INCREMENTAL.String(),
		backuppb.PolicySettings_Scheduling_ALWAYS_FULL.String(),
		backuppb.PolicySettings_Scheduling_WEEKLY_FULL_DAILY_INCREMENTAL.String(),
		backuppb.PolicySettings_Scheduling_WEEKLY_INCREMENTAL.String(),
		backuppb.PolicySettings_Scheduling_CUSTOM.String(),
	}

	resourceYandexBackupRepeatPeriodValues = []string{
		backuppb.PolicySettings_HOURLY.String(),
		backuppb.PolicySettings_DAILY.String(),
		backuppb.PolicySettings_WEEKLY.String(),
		backuppb.PolicySettings_MONTHLY.String(),
	}
)

var (
	resourceYandexBackupIntervalTemplate       = regexp.MustCompile(`^(\d+)([smMhdw])$`)
	resourceYandexBackupBadArchiveNameTemplate = regexp.MustCompile(`^.*\[.+\]\d*$`)
)

var errBackupPolicyBindingsNotFound = errors.New("backup policy bindings not found")

func resourceYandexBackupPolicySettingsTimeOfDayValidateFunc(v any, _ string) ([]string, []error) {
	const timeFormat = "15:04"

	_, err := time.Parse(timeFormat, v.(string))
	if err != nil {
		return nil, []error{err}
	}

	return nil, nil
}

func resourceYandexBackupSchedulingIntervalValidateFunc(v any, _ string) (warnings []string, errs []error) {
	value, ok := v.(int)
	if !ok {
		panic(fmt.Errorf("value expected to be string, but got %T", v))
	}

	if value == 0 {
		return warnings, errs
	}

	if value > yandexBackupPolicySchedulingIntervalMaxAvailableValue {
		errs = append(errs, fmt.Errorf("value %d cannot exceed 9999 days", value))
	}

	return warnings, errs
}

func resourceYandexBackupIntervalValidationFunc(v any, _ string) (warnings []string, errs []error) {
	value, ok := v.(string)
	if !ok {
		panic(fmt.Errorf("value expected to be string, but got %T", v))
	}

	if value == "" {
		return warnings, errs
	}

	results := resourceYandexBackupIntervalTemplate.FindAllStringSubmatch(value, -1)
	if len(results) == 0 {
		regexpValue := resourceYandexBackupIntervalTemplate.String()
		errs = append(errs, fmt.Errorf("value %s does not match regexp: %s", value, regexpValue))
	}

	return warnings, errs
}

func resourceYandexBackupIntervalSchedulingAdjust(interval *backuppb.PolicySettings_Interval) *backuppb.PolicySettings_Interval {
	if interval == nil {
		return nil
	}

	var multiplier int64 = 1
	switch itype := interval.GetType(); itype {
	case backuppb.PolicySettings_Interval_MONTHS, backuppb.PolicySettings_Interval_WEEKS:
		if itype == backuppb.PolicySettings_Interval_MONTHS {
			multiplier *= 30
		} else {
			multiplier *= 7
		}
		fallthrough
	case backuppb.PolicySettings_Interval_DAYS:
		multiplier *= 24
		fallthrough
	case backuppb.PolicySettings_Interval_HOURS:
		multiplier *= 60
		fallthrough
	case backuppb.PolicySettings_Interval_MINUTES:
		multiplier *= 60
		fallthrough
	case backuppb.PolicySettings_Interval_SECONDS:
	}

	interval.Type = backuppb.PolicySettings_Interval_SECONDS
	interval.Count *= multiplier

	return interval
}

func prepareBackupPolicyUpdateRequest(d *schema.ResourceData, _ *Config) (*backuppb.UpdatePolicyRequest, error) {
	settings, err := expandYandexBackupPolicySettingsFromResource(d)
	if err != nil {
		return nil, err
	}

	request := &backuppb.UpdatePolicyRequest{
		PolicyId: d.Id(),
		Settings: settings,
	}

	return request, nil
}

func expandBackupPolicySettingsTimeOfDay(v any) (tod *backuppb.PolicySettings_TimeOfDay, err error) {
	const timeFormat = "15:04"

	value := v.(string)

	if value == "" {
		return nil, nil
	}

	ts, err := time.Parse(timeFormat, value)
	if err != nil {
		return tod, err
	}

	tod = &backuppb.PolicySettings_TimeOfDay{
		Hour:   int64(ts.Hour()),
		Minute: int64(ts.Minute()),
	}

	return tod, nil
}

// TODO: handle naming

func expandBackupPolicySettingsSchedulingBackupSetTime(v any) (*backuppb.PolicySettings_Scheduling_BackupSet_Time_, error) {
	valueSet := v.(*schema.Set)
	if valueSet.Len() == 0 {
		return nil, nil
	}

	value := valueSet.List()[0].(map[string]any)

	out := new(backuppb.PolicySettings_Scheduling_BackupSet_Time)

	if repeatAt, ok := value["repeat_at"]; ok {
		log.Printf("[WARN] repeatAt: %[1]T %[1]v", repeatAt)
		repeatAtValues := repeatAt.([]any)
		out.RepeatAt = make([]*backuppb.PolicySettings_TimeOfDay, 0, len(repeatAtValues))
		for _, repeatAtValue := range repeatAtValues {
			repeatAtParsedValue, err := expandBackupPolicySettingsTimeOfDay(repeatAtValue)
			if err != nil {
				return nil, fmt.Errorf("parsing Cloud Backup scheduling repeat_at: %w", err)
			}

			out.RepeatAt = append(out.RepeatAt, repeatAtParsedValue)
		}
	}

	weekdayValues := value["weekdays"].([]any)
	out.Weekdays = make([]backuppb.PolicySettings_Day, 0, len(weekdayValues))
	for _, weekdayValue := range weekdayValues {
		weekday := expandBackupPolicySettingsDay(weekdayValue)
		out.Weekdays = append(out.Weekdays, weekday)
	}

	if repeatEvery, ok := value["repeat_every"]; ok {
		out.RepeatEvery = expandBackupPolicySettingsInterval(repeatEvery)
	}

	out.Monthdays = expandInt64(value["monthdays"])
	out.IncludeLastDayOfMonth = value["include_last_day_of_month"].(bool)
	out.Months = expandInt64(value["months"])
	out.Type = expandBackupPolicySettingsRepeatPeriod(value["type"])

	return &backuppb.PolicySettings_Scheduling_BackupSet_Time_{
		Time: out,
	}, nil
}

func expandBackupPolicySettingsSchedulingBackupSetDelay(v any) *backuppb.PolicySettings_Scheduling_BackupSet_SinceLastExecTime_ {
	value := v.(int)
	return &backuppb.PolicySettings_Scheduling_BackupSet_SinceLastExecTime_{
		SinceLastExecTime: &backuppb.PolicySettings_Scheduling_BackupSet_SinceLastExecTime{
			Delay: &backuppb.PolicySettings_Interval{
				Type:  backuppb.PolicySettings_Interval_SECONDS,
				Count: int64(value),
			},
		},
	}
}

func expandBackupPolicySettingsSchedulingExecuteByInterval(v any) *backuppb.PolicySettings_Scheduling_BackupSet {
	value, ok := v.(int)
	if !ok || value == 0 {
		return nil
	}

	return &backuppb.PolicySettings_Scheduling_BackupSet{
		Setting: expandBackupPolicySettingsSchedulingBackupSetDelay(v),
	}
}

func expandBackupPolicySettingsSchedulingExecuteByTime(v any) (*backuppb.PolicySettings_Scheduling_BackupSet, error) {
	setting, err := expandBackupPolicySettingsSchedulingBackupSetTime(v)
	if err != nil {
		return nil, err
	}
	if setting == nil {
		return nil, nil
	}

	return &backuppb.PolicySettings_Scheduling_BackupSet{
		Setting: setting,
	}, nil
}

func expandBackupScheme(v any) backuppb.PolicySettings_Scheduling_Scheme {
	value := v.(string)
	value = strings.ToUpper(value)
	out := backuppb.PolicySettings_Scheduling_Scheme_value[value]
	return backuppb.PolicySettings_Scheduling_Scheme(out)
}

func expandBackupPolicySettingsRepeatPeriod(v any) backuppb.PolicySettings_RepeatePeriod {
	value := v.(string)
	value = strings.ToUpper(value)
	out := backuppb.PolicySettings_RepeatePeriod_value[value]
	return backuppb.PolicySettings_RepeatePeriod(out)
}

func expandBackupPolicySettingsDay(v any) backuppb.PolicySettings_Day {
	value := v.(string)
	value = strings.ToUpper(value)
	out := backuppb.PolicySettings_Day_value[value]
	return backuppb.PolicySettings_Day(out)
}

func expandBackupPolicyScheduling(v any) (scheduling *backuppb.PolicySettings_Scheduling, err error) {
	settingsSet := v.(*schema.Set)

	if settingsSet.Len() == 0 {
		return nil, nil
	}

	settings := settingsSet.List()[0].(map[string]any)

	scheduling = new(backuppb.PolicySettings_Scheduling)

	// TODO: deprecated, remove later
	var byTime, byInterval bool
	bs := expandBackupPolicySettingsSchedulingExecuteByInterval(settings["execute_by_interval"])
	if bs != nil {
		byInterval = true
		scheduling.BackupSets = append(scheduling.BackupSets, bs)
	}

	bs, err = expandBackupPolicySettingsSchedulingExecuteByTime(settings["execute_by_time"])
	if err != nil {
		return nil, fmt.Errorf("expanding Cloud Backup Policy Scheduling execute_by_time: %w", err)
	}
	if bs != nil {
		byTime = true
		scheduling.BackupSets = append(scheduling.BackupSets, bs)
	}

	if byTime && byInterval {
		return nil, fmt.Errorf("should be set exactly one of: execute_by_interval, execute_by_time")
	}
	////////////

	backupSets := settings["backup_sets"].(*schema.Set)
	for _, s := range backupSets.List() {
		bs, err := expandBackupPolicySchedulingBackupSet(s)
		if err != nil {
			return nil, err
		}
		scheduling.BackupSets = append(scheduling.BackupSets, bs)
	}

	if len(scheduling.BackupSets) == 0 {
		return nil, fmt.Errorf("at least one backup set should be specified")
	}

	scheduling.Enabled = settings["enabled"].(bool)
	scheduling.MaxParallelBackups = int64(settings["max_parallel_backups"].(int))
	scheduling.RandMaxDelay = expandBackupPolicySettingsInterval(settings["random_max_delay"])
	scheduling.Scheme = expandBackupScheme(settings["scheme"])
	scheduling.WeeklyBackupDay = expandBackupPolicySettingsDay(settings["weekly_backup_day"])

	return scheduling, nil
}

func expandBackupPolicySchedulingBackupSet(v any) (bs *backuppb.PolicySettings_Scheduling_BackupSet, err error) {
	settings := v.(map[string]any)

	bsInterval := expandBackupPolicySettingsSchedulingExecuteByInterval(settings["execute_by_interval"])
	bsTime, err := expandBackupPolicySettingsSchedulingExecuteByTime(settings["execute_by_time"])
	if err != nil {
		return nil, fmt.Errorf("expanding Cloud Backup Policy Scheduling execute_by_time: %w", err)
	}

	exactlyOne := (bsInterval == nil) != (bsTime == nil)
	if !exactlyOne {
		return nil, fmt.Errorf("should be set exactly one of: execute_by_interval, execute_by_time")
	}

	if bsInterval != nil {
		bsInterval.Type = expandBackupPolicySchedulingBackupSetType(settings["type"])
		return bsInterval, nil
	} else {
		bsTime.Type = expandBackupPolicySchedulingBackupSetType(settings["type"])
		return bsTime, nil
	}
}

func expandBackupPolicySchedulingBackupSetType(v any) backuppb.PolicySettings_Scheduling_BackupSet_Type {
	value := v.(string)
	value = strings.ToUpper(value)
	out := backuppb.PolicySettings_Scheduling_BackupSet_Type_value[value]
	return backuppb.PolicySettings_Scheduling_BackupSet_Type(out)
}

func expandYandexBackupPolicyRetentionRule(v any) (out *backuppb.PolicySettings_Retention_RetentionRule) {
	value := v.(map[string]any)

	maxAge, maxAgeOk := value["max_age"].(string)
	maxCount, maxCountOk := value["max_count"].(int)

	isMaxAgeSet := maxAgeOk && maxAge != ""
	isMaxCountSet := maxCountOk && maxCount > 0

	if isMaxAgeSet == isMaxCountSet {
		panic("expected to be set only one of the max_age, max_count in retention policy")
	}

	out = new(backuppb.PolicySettings_Retention_RetentionRule)
	if isMaxAgeSet {
		out.Condition = &backuppb.PolicySettings_Retention_RetentionRule_MaxAge{
			MaxAge: expandBackupPolicySettingsInterval(maxAge),
		}
	} else {
		out.Condition = &backuppb.PolicySettings_Retention_RetentionRule_MaxCount{
			MaxCount: int64(maxCount),
		}
	}

	repeatPeriodList := value["repeat_period"].([]any)
	out.BackupSet = make([]backuppb.PolicySettings_RepeatePeriod, 0, len(repeatPeriodList))

	for _, repeatPeriod := range repeatPeriodList {
		repeatPeriodValue := expandBackupPolicySettingsRepeatPeriod(repeatPeriod)
		out.BackupSet = append(out.BackupSet, repeatPeriodValue)
	}

	return out
}

func expandYandexBackupPolicyRetention(v any) (rts *backuppb.PolicySettings_Retention, err error) {
	vscheme := v.(*schema.Set)
	if vscheme.Len() == 0 {
		return nil, nil
	}

	value := vscheme.List()[0].(map[string]any)
	rts = new(backuppb.PolicySettings_Retention)
	rts.BeforeBackup = !value["after_backup"].(bool)
	rules := value["rules"].(*schema.Set).List()
	for _, rule := range rules {
		rulepb := expandYandexBackupPolicyRetentionRule(rule)

		rts.Rules = append(rts.Rules, rulepb)
	}

	return rts, nil
}

func expandYandexBackupCompression(v any) backuppb.PolicySettings_Compression {
	value := v.(string)
	out := backuppb.PolicySettings_Compression_value[value]
	return backuppb.PolicySettings_Compression(out)
}

func expandYandexBackupFormat(v any) backuppb.Format {
	value := v.(string)
	out := backuppb.Format_value[value]
	return backuppb.Format(out)
}

func expandYandexBackupCBT(v any) backuppb.PolicySettings_ChangedBlockTracking {
	value := v.(string)
	out := backuppb.PolicySettings_ChangedBlockTracking_value[value]
	return backuppb.PolicySettings_ChangedBlockTracking(out)
}

func expandYandexBackupVSSProvider(v any) backuppb.PolicySettings_VolumeShadowCopyServiceSettings_VSSProvider {
	value := v.(string)
	provider := backuppb.PolicySettings_VolumeShadowCopyServiceSettings_VSSProvider_value[value]
	return backuppb.PolicySettings_VolumeShadowCopyServiceSettings_VSSProvider(provider)
}

func expandYandexBackupPolicyRetriesConfiguration(v any) *backuppb.PolicySettings_RetriesConfiguration {
	valueSet := v.(*schema.Set)
	if valueSet.Len() == 0 {
		return nil
	}

	value := valueSet.List()[0].(map[string]any)
	out := new(backuppb.PolicySettings_RetriesConfiguration)
	out.Enabled = value["enabled"].(bool)
	out.MaxAttempts = int64(value["max_attempts"].(int))
	out.Interval = expandBackupPolicySettingsInterval(value["interval"])

	return out
}

func expandYandexBackupPolicySettingsFromResource(d *schema.ResourceData) (settings *backuppb.PolicySettings, err error) {
	retentionRules, err := expandYandexBackupPolicyRetention(d.Get("retention"))
	if err != nil {
		return nil, fmt.Errorf("preparing Cloud Backup Policy retention: %w", err)
	}
	scheduling, err := expandBackupPolicyScheduling(d.Get("scheduling"))
	if err != nil {
		return nil, fmt.Errorf("preparing Cloud Backup Policy scheduling: %w", err)
	}
	splittingBytes, err := strconv.ParseInt(d.Get("splitting_bytes").(string), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parsing splitting bytes value: %w", err)
	}

	splitting := &backuppb.PolicySettings_Splitting{
		Size: splittingBytes,
	}
	archive := &backuppb.PolicySettings_ArchiveProperties{
		Name: d.Get("archive_name").(string),
	}
	performanceWindow := &backuppb.PolicySettings_PerformanceWindow{
		Enabled: d.Get("performance_window_enabled").(bool),
	}
	vss := &backuppb.PolicySettings_VolumeShadowCopyServiceSettings{
		Enabled:  true,
		Provider: expandYandexBackupVSSProvider(d.Get("vss_provider")),
	}

	settings = &backuppb.PolicySettings{
		Retention:         retentionRules,
		Scheduling:        scheduling,
		Splitting:         splitting,
		Archive:           archive,
		PerformanceWindow: performanceWindow,
		Vss:               vss,

		MultiVolumeSnapshottingEnabled: d.Get("multi_volume_snapshotting_enabled").(bool),
		PreserveFileSecuritySettings:   d.Get("preserve_file_security_settings").(bool),
		SilentModeEnabled:              d.Get("silent_mode_enabled").(bool),
		FastBackupEnabled:              d.Get("fast_backup_enabled").(bool),
		QuiesceSnapshottingEnabled:     d.Get("quiesce_snapshotting_enabled").(bool),

		Compression:          expandYandexBackupCompression(d.Get("compression")),
		Format:               expandYandexBackupFormat(d.Get("format")),
		Reattempts:           expandYandexBackupPolicyRetriesConfiguration(d.Get("reattempts")),
		VmSnapshotReattempts: expandYandexBackupPolicyRetriesConfiguration(d.Get("vm_snapshot_reattempts")),
		Cbt:                  expandYandexBackupCBT(d.Get("cbt")),
	}

	return settings, nil
}

func flattenBackupPolicySettingsRetriesConfiguration(d *schema.ResourceData, retrySettings *backuppb.PolicySettings_RetriesConfiguration, key string) (err error) {
	if retrySettings == nil {
		return nil
	}

	result := map[string]any{
		"enabled":      retrySettings.Enabled,
		"max_attempts": retrySettings.MaxAttempts,
		"interval":     flattenBackupInterval(retrySettings.Interval),
	}

	return d.Set(key, []any{result})
}

func flattenBackupPolicySettingsVSS(d *schema.ResourceData, vss *backuppb.PolicySettings_VolumeShadowCopyServiceSettings) (err error) {
	if vss == nil {
		return nil
	}

	return d.Set("vss_provider", vss.GetProvider().String())
}

func flattenBackupPolicySettingsRetention(d *schema.ResourceData, retention *backuppb.PolicySettings_Retention) (err error) {
	if retention == nil {
		return nil
	}

	result := make(map[string]any, 2)
	result["after_backup"] = !retention.BeforeBackup

	resultRules := make([]any, 0, len(retention.Rules))
	for _, rule := range retention.Rules {
		resultRule := make(map[string]any, 2)
		switch cond := rule.Condition.(type) {
		case *backuppb.PolicySettings_Retention_RetentionRule_MaxAge:
			resultRule["max_age"] = flattenBackupInterval(cond.MaxAge)
		case *backuppb.PolicySettings_Retention_RetentionRule_MaxCount:
			resultRule["max_count"] = cond.MaxCount
		}

		backupSets := make([]string, 0, len(rule.BackupSet))
		for _, bs := range rule.BackupSet {
			backupSets = append(backupSets, bs.String())
		}
		resultRule["repeat_period"] = backupSets

		resultRules = append(resultRules, resultRule)
	}

	result["rules"] = schema.NewSet(storageBucketS3SetFunc("max_age", "max_count", "repeat_period"), resultRules)

	return d.Set("retention", []any{result})
}

func flattenYandexBackupPolicyScheduling(d *schema.ResourceData, scheduling *backuppb.PolicySettings_Scheduling) (err error) {
	if scheduling == nil {
		return nil
	}

	result := make(map[string]any, 6)
	result["enabled"] = scheduling.Enabled
	result["max_parallel_backups"] = scheduling.MaxParallelBackups
	result["random_max_delay"] = flattenBackupInterval(scheduling.RandMaxDelay)
	result["scheme"] = scheduling.Scheme.String()
	result["weekly_backup_day"] = scheduling.WeeklyBackupDay.String()

	bss, err := flattenYandexBackupPolicySchedulingBackupSet(scheduling.GetBackupSets())
	if err != nil {
		return err
	}
	result["backup_sets"] = bss

	return d.Set("scheduling", []any{result})
}

func flattenYandexBackupPolicySchedulingBackupSet(bss []*backuppb.PolicySettings_Scheduling_BackupSet) ([]map[string]any, error) {
	if len(bss) == 0 {
		return nil, fmt.Errorf("expected to have at least one scheduling backup set")
	}

	result := make([]map[string]any, len(bss))
	for i, bs := range bss {
		result[i] = make(map[string]any, 2)
		switch typedBS := bs.Setting.(type) {
		case *backuppb.PolicySettings_Scheduling_BackupSet_SinceLastExecTime_:
			delay := resourceYandexBackupIntervalSchedulingAdjust(typedBS.SinceLastExecTime.GetDelay())
			result[i]["execute_by_interval"] = delay.Count
		case *backuppb.PolicySettings_Scheduling_BackupSet_Time_:
			timeRule := typedBS.Time

			repeatAt := make([]string, 0, len(timeRule.RepeatAt))
			for _, repeatAtValue := range timeRule.RepeatAt {
				repeatAt = append(repeatAt, flattenBackupPolicySettingsTimeOfDay(repeatAtValue))
			}

			schemaSetFunc := storageBucketS3SetFunc("weekdays", "repeat_at", "repeat_every", "monthdays", "include_last_day_of_month", "months", "type")
			item := schema.NewSet(schemaSetFunc, []any{
				map[string]any{
					"repeat_at":                 repeatAt,
					"weekdays":                  asStringSlice(timeRule.Weekdays...),
					"repeat_every":              flattenBackupInterval(timeRule.RepeatEvery),
					"monthdays":                 asAnySlice(timeRule.Monthdays...),
					"include_last_day_of_month": timeRule.IncludeLastDayOfMonth,
					"months":                    asAnySlice(timeRule.Months...),
					"type":                      timeRule.Type.String(),
				},
			})

			result[i]["execute_by_time"] = item
		}
		result[i]["type"] = bs.Type.String()
	}

	return result, nil
}

func flattenYandBackupPolicySettings(d *schema.ResourceData, settings *backuppb.PolicySettings) (err error) {
	if settings == nil {
		return nil
	}

	setF := func(key string, value any) {
		if err != nil {
			return
		}

		err = d.Set(key, value)
	}

	splittingBytesStr := strconv.FormatInt(settings.GetSplitting().Size, 10)

	setF("compression", settings.Compression.String())
	setF("format", settings.Format.String())
	setF("multi_volume_snapshotting_enabled", settings.MultiVolumeSnapshottingEnabled)
	setF("preserve_file_security_settings", settings.PreserveFileSecuritySettings)
	setF("silent_mode_enabled", settings.SilentModeEnabled)
	setF("splitting_bytes", splittingBytesStr)
	setF("archive_name", settings.GetArchive().GetName())
	setF("performance_window_enabled", settings.GetPerformanceWindow().GetEnabled())
	setF("cbt", settings.Cbt.String())
	setF("fast_backup_enabled", settings.FastBackupEnabled)
	setF("quiesce_snapshotting_enabled", settings.QuiesceSnapshottingEnabled)

	if err != nil {
		return err
	}

	if err = flattenBackupPolicySettingsVSS(d, settings.Vss); err != nil {
		return err
	}

	if err = flattenBackupPolicySettingsRetriesConfiguration(d, settings.Reattempts, "reattempts"); err != nil {
		return err
	}

	if err = flattenBackupPolicySettingsRetriesConfiguration(d, settings.VmSnapshotReattempts, "vm_snapshot_reattempts"); err != nil {
		return err
	}

	if err = flattenBackupPolicySettingsRetention(d, settings.Retention); err != nil {
		return err
	}

	return flattenYandexBackupPolicyScheduling(d, settings.Scheduling)
}

func flattenBackupPolicy(d *schema.ResourceData, resource *backuppb.Policy) (err error) {
	setF := func(key string, value any) {
		if err != nil {
			return
		}

		err = d.Set(key, value)
	}

	d.SetId(resource.Id)

	setF("name", resource.Name)
	setF("folder_id", resource.FolderId)
	setF("enabled", resource.Enabled)
	setF("created_at", getTimestamp(resource.CreatedAt))
	setF("updated_at", getTimestamp(resource.UpdatedAt))

	if err != nil {
		return err
	}

	return flattenYandBackupPolicySettings(d, resource.Settings)
}

func flattenBackupPolicySettingsTimeOfDay(tod *backuppb.PolicySettings_TimeOfDay) string {
	if tod == nil {
		return ""
	}

	return fmt.Sprintf("%02d:%02d", tod.Hour, tod.Minute)
}

func asAnySlice[T any](values ...T) []any {
	out := make([]any, 0, len(values))
	for _, value := range values {
		out = append(out, value)
	}

	return out
}

func flattenBackupInterval(interval *backuppb.PolicySettings_Interval) string {
	if interval == nil {
		return ""
	}

	var intervalTypeStr string
	switch interval.GetType() {
	case backuppb.PolicySettings_Interval_SECONDS:
		intervalTypeStr = "s"
	case backuppb.PolicySettings_Interval_MINUTES:
		intervalTypeStr = "m"
	case backuppb.PolicySettings_Interval_HOURS:
		intervalTypeStr = "h"
	case backuppb.PolicySettings_Interval_DAYS:
		intervalTypeStr = "d"
	case backuppb.PolicySettings_Interval_WEEKS:
		intervalTypeStr = "w"
	case backuppb.PolicySettings_Interval_MONTHS:
		intervalTypeStr = "M"
	}

	return fmt.Sprintf("%d%s", interval.GetCount(), intervalTypeStr)
}

func flattenBackupPolicyApplication(d *schema.ResourceData, pa *backuppb.PolicyApplication) (err error) {
	if pa == nil {
		return nil
	}

	setF := func(key string, value any) {
		if err != nil {
			return
		}

		err = d.Set(key, value)
	}

	setF("policy_id", pa.PolicyId)
	setF("instance_id", pa.ComputeInstanceId)
	setF("enabled", pa.Enabled)
	setF("processing", pa.IsProcessing)
	setF("created_at", getTimestamp(pa.CreatedAt))
	return nil
}

func asStringSlice[T fmt.Stringer](values ...T) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, value.String())
	}

	return out
}

func checkBackupProviderActivated(ctx context.Context, config *Config) error {
	iterator := config.sdk.Backup().Provider().ProviderActivatedIterator(
		ctx,
		&backuppb.ListActivatedProvidersRequest{FolderId: config.FolderID},
	)
	providerNames, err := iterator.TakeAll()
	if err != nil {
		return err
	}

	if len(providerNames) == 0 {
		return fmt.Errorf("the specified folder has no activated backup providers, please, activate provider and try again")
	}
	return nil
}

func getPolicyByName(ctx context.Context, config *Config, name string) (*backuppb.Policy, error) {
	var res *backuppb.Policy
	iterator := config.sdk.Backup().Policy().PolicyIterator(ctx, &backuppb.ListPoliciesRequest{
		FolderId: config.FolderID,
	})
	for iterator.Next() {
		err := iterator.Error()
		if err != nil {
			return nil, err
		}
		policy := iterator.Value()
		if policy.Name == name {
			if res != nil {
				return nil, fmt.Errorf("more then one policy with name %q exists, use policy id or rename policy instead", name)
			}
			res = policy
		}
	}
	if res == nil {
		return nil, status.Error(codes.NotFound, "policy does not exist")
	}
	return res, nil
}

func getBackupPolicyApplication(ctx context.Context, config *Config, policyID, instanceID string) (*backuppb.PolicyApplication, error) {
	iterator := config.sdk.Backup().Policy().PolicyApplicationsIterator(ctx, &backuppb.ListApplicationsRequest{
		Id: &backuppb.ListApplicationsRequest_ComputeInstanceId{
			ComputeInstanceId: instanceID,
		},
		ShowProcessing: true,
	})

	for iterator.Next() {
		app := iterator.Value()
		if app.GetPolicyId() == policyID {
			return app, nil
		}
	}
	return nil, errBackupPolicyBindingsNotFound
}

func createBackupPolicyBindingsWithRetry(ctx context.Context, config *Config, policyID, instanceID string) (op *sdkoperation.Operation, err error) {
	const (
		firstRetryInterval = 100 * time.Second
		retryInterval      = 20 * time.Second
	)

	isRetryableError := func(op *sdkoperation.Operation, err error) bool {
		if err != nil {
			s, _ := status.FromError(err)
			if s.Code() == codes.NotFound {
				return true
			}
		} else if op.Failed() {
			return true
		}
		return false
	}

	request := &backuppb.ApplyPolicyRequest{
		PolicyId:          policyID,
		ComputeInstanceId: instanceID,
	}

	firstRetry := true
	for i := 0; i < config.MaxRetries; i++ {
		log.Printf("[INFO]: Try to bind policy_id=%q with instance_id=%q, attempt=%v", policyID, instanceID, i+1)
		op, err = config.sdk.WrapOperation(config.sdk.Backup().Policy().Apply(ctx, request))
		if isRetryableError(op, err) {
			log.Printf("[INFO]: Unable to bind policy_id=%q with instance_id=%q: %s", policyID, instanceID, err)
			if firstRetry {
				time.Sleep(firstRetryInterval)
			} else {
				time.Sleep(retryInterval)
			}
			firstRetry = false
			continue
		}
		break
	}
	return
}

func parseBackupPolicyBindingsID(id string) (policyID, instanceID string, err error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid or empty backup policy bindings id=%q", id)
	}
	return parts[0], parts[1], nil
}

func makeBackupPolicyBindingsID(policyID, instanceID string) string {
	return fmt.Sprintf("%s:%s", policyID, instanceID)
}
