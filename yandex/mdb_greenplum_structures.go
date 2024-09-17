package yandex

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/genproto/googleapis/type/timeofday"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/greenplum/v1"
)

func parseGreenplumEnv(e string) (greenplum.Cluster_Environment, error) {
	v, ok := greenplum.Cluster_Environment_value[e]
	if !ok {
		return 0, fmt.Errorf("value for 'environment' must be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeys(greenplum.Cluster_Environment_value)), e)
	}
	return greenplum.Cluster_Environment(v), nil
}
func getGreenplumConfigFieldName(version string) string {
	if version == "6.22" {
		return "greenplum_config_6_22"
	}
	return "greenplum_config_6"
}

func flattenGreenplumMasterSubcluster(r *greenplum.Resources) []map[string]interface{} {
	subcluster := map[string]interface{}{}
	resources := map[string]interface{}{}
	resources["resource_preset_id"] = r.ResourcePresetId
	resources["disk_type_id"] = r.DiskTypeId
	resources["disk_size"] = toGigabytes(r.DiskSize)
	subcluster["resources"] = []map[string]interface{}{resources}
	return []map[string]interface{}{subcluster}
}

func flattenGreenplumSegmentSubcluster(r *greenplum.Resources) []map[string]interface{} {
	subcluster := map[string]interface{}{}
	resources := map[string]interface{}{}
	resources["resource_preset_id"] = r.ResourcePresetId
	resources["disk_type_id"] = r.DiskTypeId
	resources["disk_size"] = toGigabytes(r.DiskSize)
	subcluster["resources"] = []map[string]interface{}{resources}
	return []map[string]interface{}{subcluster}
}

func flattenGreenplumHosts(masterHosts, segmentHosts []*greenplum.Host) ([]map[string]interface{}, []map[string]interface{}) {
	mHost := make([]map[string]interface{}, 0, len(masterHosts))
	for _, h := range masterHosts {
		mHost = append(mHost, map[string]interface{}{"fqdn": h.Name, "assign_public_ip": h.AssignPublicIp})
	}

	sHost := make([]map[string]interface{}, 0, len(segmentHosts))
	for _, h := range segmentHosts {
		sHost = append(sHost, map[string]interface{}{"fqdn": h.Name})
	}

	return mHost, sHost
}

func flattenGreenplumAccess(c *greenplum.GreenplumConfig) []map[string]interface{} {
	out := map[string]interface{}{}
	if c != nil && c.Access != nil {
		out["data_lens"] = c.Access.DataLens
		out["web_sql"] = c.Access.WebSql
		out["data_transfer"] = c.Access.DataTransfer
		out["yandex_query"] = c.Access.YandexQuery
	}
	return []map[string]interface{}{out}
}

func flattenGreenplumCloudStorage(c *greenplum.CloudStorage) []map[string]interface{} {
	out := map[string]interface{}{}
	if c != nil {
		out["enable"] = c.Enable
	}
	return []map[string]interface{}{out}
}

func flattenGreenplumMaintenanceWindow(mw *greenplum.MaintenanceWindow) ([]interface{}, error) {
	maintenanceWindow := map[string]interface{}{}
	if mw != nil {
		switch p := mw.GetPolicy().(type) {
		case *greenplum.MaintenanceWindow_Anytime:
			maintenanceWindow["type"] = "ANYTIME"
			// do nothing
		case *greenplum.MaintenanceWindow_WeeklyMaintenanceWindow:
			maintenanceWindow["type"] = "WEEKLY"
			maintenanceWindow["hour"] = p.WeeklyMaintenanceWindow.Hour
			maintenanceWindow["day"] = greenplum.WeeklyMaintenanceWindow_WeekDay_name[int32(p.WeeklyMaintenanceWindow.GetDay())]
		default:
			return nil, fmt.Errorf("unsupported greenplum maintenance policy type")
		}
	}

	return []interface{}{maintenanceWindow}, nil
}

func expandGreenplumMaintenanceWindow(d *schema.ResourceData) (*greenplum.MaintenanceWindow, error) {
	if _, ok := d.GetOkExists("maintenance_window"); !ok {
		return nil, nil
	}

	out := &greenplum.MaintenanceWindow{}
	typeMW, _ := d.GetOk("maintenance_window.0.type")
	if typeMW == "ANYTIME" {
		if hour, ok := d.GetOk("maintenance_window.0.hour"); ok && hour != "" {
			return nil, fmt.Errorf("hour should be not set, when using ANYTIME")
		}
		if day, ok := d.GetOk("maintenance_window.0.day"); ok && day != "" {
			return nil, fmt.Errorf("day should be not set, when using ANYTIME")
		}
		out.Policy = &greenplum.MaintenanceWindow_Anytime{
			Anytime: &greenplum.AnytimeMaintenanceWindow{},
		}
	} else if typeMW == "WEEKLY" {
		hour := d.Get("maintenance_window.0.hour").(int)
		dayString := d.Get("maintenance_window.0.day").(string)

		day, ok := greenplum.WeeklyMaintenanceWindow_WeekDay_value[dayString]
		if !ok || day == 0 {
			return nil, fmt.Errorf(`day value should be one of ("MON", "TUE", "WED", "THU", "FRI", "SAT", "SUN")`)
		}

		out.Policy = &greenplum.MaintenanceWindow_WeeklyMaintenanceWindow{
			WeeklyMaintenanceWindow: &greenplum.WeeklyMaintenanceWindow{
				Hour: int64(hour),
				Day:  greenplum.WeeklyMaintenanceWindow_WeekDay(day),
			},
		}
	} else {
		return nil, fmt.Errorf("maintenance_window.0.type should be ANYTIME or WEEKLY")
	}

	return out, nil
}

func flattenGreenplumClusterConfig(c *greenplum.ClusterConfigSet) (map[string]string, error) {
	var gpConfig interface{}

	if cf, ok := c.GreenplumConfig.(*greenplum.ClusterConfigSet_GreenplumConfigSet_6); ok {
		gpConfig = cf.GreenplumConfigSet_6.UserConfig
	} else if cf, ok := c.GreenplumConfig.(*greenplum.ClusterConfigSet_GreenplumConfigSet_6_22); ok {
		gpConfig = cf.GreenplumConfigSet_6_22.UserConfig
	}

	return flattenResourceGenerateMapS(gpConfig, false, mdbGreenplumSettingsFieldsInfo, false, true, nil)
}

func flattenGreenplumPoolerConfig(c *greenplum.ConnectionPoolerConfigSet) ([]interface{}, error) {
	if c == nil {
		return nil, nil
	}

	out := map[string]interface{}{}

	out["pooling_mode"] = c.EffectiveConfig.GetMode().String()
	out["pool_size"] = c.EffectiveConfig.GetSize().GetValue()
	out["pool_client_idle_timeout"] = c.EffectiveConfig.GetClientIdleTimeout().GetValue()

	return []interface{}{out}, nil
}

func flattenGreenplumPXFConfig(c *greenplum.PXFConfigSet) ([]interface{}, error) {
	if c == nil {
		return nil, nil
	}

	out := map[string]interface{}{}

	out["connection_timeout"] = c.EffectiveConfig.GetConnectionTimeout().GetValue()
	out["upload_timeout"] = c.EffectiveConfig.GetUploadTimeout().GetValue()

	out["max_threads"] = c.EffectiveConfig.GetMaxThreads().GetValue()
	out["pool_allow_core_thread_timeout"] = c.EffectiveConfig.GetPoolAllowCoreThreadTimeout().GetValue()
	out["pool_core_size"] = c.EffectiveConfig.GetPoolCoreSize().GetValue()
	out["pool_queue_capacity"] = c.EffectiveConfig.GetPoolQueueCapacity().GetValue()
	out["pool_max_size"] = c.EffectiveConfig.GetPoolMaxSize().GetValue()

	out["xmx"] = c.EffectiveConfig.GetXmx().GetValue()
	out["xms"] = c.EffectiveConfig.GetXms().GetValue()

	return []interface{}{out}, nil
}

func flattenGreenplumBackgroundActivities(c *greenplum.BackgroundActivitiesConfig) ([]interface{}, error) {
	if c == nil {
		return nil, nil
	}

	out := map[string]interface{}{}

	if c.AnalyzeAndVacuum != nil {
		av := map[string]interface{}{}
		if c.AnalyzeAndVacuum.Start != nil {
			av["start_time"] = fmt.Sprintf("%d:%d", c.AnalyzeAndVacuum.Start.Hours, c.AnalyzeAndVacuum.Start.Minutes)
		}
		if c.AnalyzeAndVacuum.AnalyzeTimeout != nil {
			av["analyze_timeout"] = c.AnalyzeAndVacuum.AnalyzeTimeout.Value
		}
		if c.AnalyzeAndVacuum.VacuumTimeout != nil {
			av["vacuum_timeout"] = c.AnalyzeAndVacuum.VacuumTimeout.Value
		}
		out["analyze_and_vacuum"] = []interface{}{av}
	}
	if c.QueryKillerScripts != nil && c.QueryKillerScripts.Idle != nil {
		qk := flattenGreenplumQueryKiller(c.QueryKillerScripts.Idle)
		out["query_killer_idle"] = []interface{}{qk}
	}
	if c.QueryKillerScripts != nil && c.QueryKillerScripts.IdleInTransaction != nil {
		qk := flattenGreenplumQueryKiller(c.QueryKillerScripts.IdleInTransaction)
		out["query_killer_idle_in_transaction"] = []interface{}{qk}
	}
	if c.QueryKillerScripts != nil && c.QueryKillerScripts.LongRunning != nil {
		qk := flattenGreenplumQueryKiller(c.QueryKillerScripts.LongRunning)
		out["query_killer_long_running"] = []interface{}{qk}
	}

	return []interface{}{out}, nil
}

func flattenGreenplumQueryKiller(c *greenplum.QueryKiller) interface{} {
	if c == nil {
		return nil
	}

	out := map[string]interface{}{}
	if c.Enable != nil {
		out["enable"] = c.Enable.Value
	}
	if c.MaxAge != nil {
		out["max_age"] = c.MaxAge.Value
	}
	if c.IgnoreUsers != nil {
		out["ignore_users"] = c.IgnoreUsers
	}
	return out
}

func expandGreenplumBackupWindowStart(d *schema.ResourceData) *timeofday.TimeOfDay {
	out := &timeofday.TimeOfDay{}

	if v, ok := d.GetOk("backup_window_start.0.hours"); ok {
		out.Hours = int32(v.(int))
	}

	if v, ok := d.GetOk("backup_window_start.0.minutes"); ok {
		out.Minutes = int32(v.(int))
	}

	return out
}

func expandGreenplumAccess(d *schema.ResourceData) *greenplum.Access {
	if _, ok := d.GetOkExists("access"); !ok {
		return nil
	}

	out := &greenplum.Access{}

	if v, ok := d.GetOk("access.0.data_lens"); ok {
		out.DataLens = v.(bool)
	}

	if v, ok := d.GetOk("access.0.web_sql"); ok {
		out.WebSql = v.(bool)
	}

	if v, ok := d.GetOk("access.0.data_transfer"); ok {
		out.DataTransfer = v.(bool)
	}

	if v, ok := d.GetOk("access.0.yandex_query"); ok {
		out.YandexQuery = v.(bool)
	}

	return out
}

func expandGreenplumCloudStorage(d *schema.ResourceData) *greenplum.CloudStorage {
	if _, ok := d.GetOk("cloud_storage"); !ok {
		return nil
	}

	out := &greenplum.CloudStorage{}

	if v, ok := d.GetOk("cloud_storage.0.enable"); ok {
		out.Enable = v.(bool)
	}

	return out
}

func expandGreenplumUpdatePath(d *schema.ResourceData, settingNames []string) []string {
	mdbGreenplumUpdateFieldsMap := map[string]string{
		"name":                   "name",
		"description":            "description",
		"user_password":          "user_password",
		"labels":                 "labels",
		"network_id":             "network_id",
		"access.0.data_lens":     "config.access.data_lens",
		"access.0.web_sql":       "config.access.web_sql",
		"access.0.data_transfer": "config.access.data_transfer",
		"access.0.yandex_query":  "config.access.yandex_query",
		"cloud_storage.0.enable": "cloud_storage",
		"backup_window_start":    "config.backup_window_start",
		"maintenance_window":     "maintenance_window",
		"deletion_protection":    "deletion_protection",
		"security_group_ids":     "security_group_ids",

		"pooler_config.0.pooling_mode":             "config_spec.pool.mode",
		"pooler_config.0.pool_size":                "config_spec.pool.size",
		"pooler_config.0.pool_client_idle_timeout": "config_spec.pool.client_idle_timeout",

		"pxf_config.0.connection_timeout":             "config_spec.pxf_config.connection_timeout",
		"pxf_config.0.upload_timeout":                 "config_spec.pxf_config.upload_timeout",
		"pxf_config.0.max_threads":                    "config_spec.pxf_config.max_threads",
		"pxf_config.0.pool_allow_core_thread_timeout": "config_spec.pxf_config.pool_allow_core_thread_timeout",
		"pxf_config.0.pool_core_size":                 "config_spec.pxf_config.pool_core_size",
		"pxf_config.0.pool_queue_capacity":            "config_spec.pxf_config.pool_queue_capacity",
		"pxf_config.0.pool_max_size":                  "config_spec.pxf_config.pool_max_size",
		"pxf_config.0.xmx":                            "config_spec.pxf_config.xmx",
		"pxf_config.0.xms":                            "config_spec.pxf_config.xms",

		"master_subcluster.0.resources.0.resource_preset_id": "master_config.resources.resource_preset_id",
		"master_subcluster.0.resources.0.disk_type_id":       "master_config.resources.disk_type_id",
		"master_subcluster.0.resources.0.disk_size":          "master_config.resources.disk_size",

		"segment_subcluster.0.resources.0.resource_preset_id": "segment_config.resources.resource_preset_id",
		"segment_subcluster.0.resources.0.disk_type_id":       "segment_config.resources.disk_type_id",
		"segment_subcluster.0.resources.0.disk_size":          "segment_config.resources.disk_size",

		"background_activities.0.analyze_and_vacuum.0.start_time":                 "config_spec.background_activities.analyze_and_vacuum.start",
		"background_activities.0.analyze_and_vacuum.0.analyze_timeout":            "config_spec.background_activities.analyze_and_vacuum.analyze_timeout",
		"background_activities.0.analyze_and_vacuum.0.vacuum_timeout":             "config_spec.background_activities.analyze_and_vacuum.vacuum_timeout",
		"background_activities.0.query_killer_idle.0.enable":                      "config_spec.background_activities.query_killer_scripts.idle.enable",
		"background_activities.0.query_killer_idle.0.max_age":                     "config_spec.background_activities.query_killer_scripts.idle.max_age",
		"background_activities.0.query_killer_idle.0.ignore_users":                "config_spec.background_activities.query_killer_scripts.idle.ignore_users",
		"background_activities.0.query_killer_idle_in_transaction.0.enable":       "config_spec.background_activities.query_killer_scripts.idle_in_transaction.enable",
		"background_activities.0.query_killer_idle_in_transaction.0.max_age":      "config_spec.background_activities.query_killer_scripts.idle_in_transaction.max_age",
		"background_activities.0.query_killer_idle_in_transaction.0.ignore_users": "config_spec.background_activities.query_killer_scripts.idle_in_transaction.ignore_users",
		"background_activities.0.query_killer_long_running.0.enable":              "config_spec.background_activities.query_killer_scripts.long_running.enable",
		"background_activities.0.query_killer_long_running.0.max_age":             "config_spec.background_activities.query_killer_scripts.long_running.max_age",
		"background_activities.0.query_killer_long_running.0.ignore_users":        "config_spec.background_activities.query_killer_scripts.long_running.ignore_users",
	}

	updatePath := []string{}
	for field, path := range mdbGreenplumUpdateFieldsMap {
		if d.HasChange(field) {
			updatePath = append(updatePath, path)
		}
	}

	// version is like 6.22, 6.25, etc.
	version := d.Get("version").(string)
	gpFieldName := getGreenplumConfigFieldName(version)

	for _, setting := range settingNames {
		field := fmt.Sprintf("greenplum_config.%s", setting)
		if d.HasChange(field) {
			path := fmt.Sprintf("config_spec.%s.%s", gpFieldName, setting)
			updatePath = append(updatePath, path)
		}
	}

	return updatePath
}

func expandGreenplumConfigSpec(d *schema.ResourceData) (*greenplum.ConfigSpec, []string, error) {
	poolerConfig, err := expandGreenplumPoolerConfig(d)
	if err != nil {
		return nil, nil, err
	}

	pxfConfig, err := expandGreenplumPXFConfig(d)
	if err != nil {
		return nil, nil, err
	}

	gpConfig1, gpConfig2, settingNames, err := expandGreenplumConfigSpecGreenplumConfig(d)
	if err != nil {
		return nil, nil, err
	}

	backgroundActivities, err := expandGreenplumBackgroundActivities(d)
	if err != nil {
		return nil, nil, err
	}

	configSpec := &greenplum.ConfigSpec{
		Pool:                 poolerConfig,
		PxfConfig:            pxfConfig,
		BackgroundActivities: backgroundActivities,
	}
	if gpConfig1 != nil {
		configSpec.GreenplumConfig = gpConfig1
	} else {
		configSpec.GreenplumConfig = gpConfig2
	}

	return configSpec, settingNames, nil
}

func expandGreenplumConfigSpecGreenplumConfig(d *schema.ResourceData) (*greenplum.ConfigSpec_GreenplumConfig_6_22, *greenplum.ConfigSpec_GreenplumConfig_6, []string, error) {
	version := d.Get("version").(string)
	if version == "6.22" {
		cfg := &greenplum.ConfigSpec_GreenplumConfig_6_22{
			GreenplumConfig_6_22: &greenplum.GreenplumConfig6_22{},
		}

		settingNames, err := expandResourceGenerateNonSkippedFields(mdbGreenplumSettingsFieldsInfo, d, cfg.GreenplumConfig_6_22, "greenplum_config.", true)
		if err != nil {
			return nil, nil, []string{}, err
		}
		return cfg, nil, settingNames, nil
	} else if version == "6.25" {
		cfg := &greenplum.ConfigSpec_GreenplumConfig_6{
			GreenplumConfig_6: &greenplum.GreenplumConfig6{},
		}

		settingNames, err := expandResourceGenerateNonSkippedFields(mdbGreenplumSettingsFieldsInfo, d, cfg.GreenplumConfig_6, "greenplum_config.", true)
		if err != nil {
			return nil, nil, []string{}, err
		}
		return nil, cfg, settingNames, nil
	}

	return nil, nil, nil, fmt.Errorf("unknown Greenplum version %s but '6.22' and '6.25' are only available", version)
}

func expandGreenplumPoolerConfig(d *schema.ResourceData) (*greenplum.ConnectionPoolerConfig, error) {
	pc := &greenplum.ConnectionPoolerConfig{}

	if v, ok := d.GetOk("pooler_config.0.pooling_mode"); ok {
		pm, err := parseGreenplumPoolingMode(v.(string))
		if err != nil {
			return nil, err
		}

		pc.Mode = pm
	}

	if v, ok := d.GetOk("pooler_config.0.pool_size"); ok {
		pc.Size = &wrappers.Int64Value{Value: int64(v.(int))}
	}

	if v, ok := d.GetOk("pooler_config.0.pool_client_idle_timeout"); ok {
		pc.ClientIdleTimeout = &wrappers.Int64Value{Value: int64(v.(int))}
	}

	return pc, nil
}

func expandGreenplumPXFConfig(d *schema.ResourceData) (*greenplum.PXFConfig, error) {
	pc := &greenplum.PXFConfig{}

	if v, ok := d.GetOk("pxf_config.0.connection_timeout"); ok {
		pc.ConnectionTimeout = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk("pxf_config.0.upload_timeout"); ok {
		pc.UploadTimeout = &wrappers.Int64Value{Value: int64(v.(int))}
	}

	if v, ok := d.GetOk("pxf_config.0.max_threads"); ok {
		pc.MaxThreads = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk("pxf_config.0.pool_allow_core_thread_timeout"); ok {
		pc.PoolAllowCoreThreadTimeout = &wrappers.BoolValue{Value: v.(bool)}
	}
	if v, ok := d.GetOk("pxf_config.0.pool_core_size"); ok {
		pc.PoolCoreSize = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk("pxf_config.0.pool_queue_capacity"); ok {
		pc.PoolQueueCapacity = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk("pxf_config.0.pool_max_size"); ok {
		pc.PoolMaxSize = &wrappers.Int64Value{Value: int64(v.(int))}
	}

	if v, ok := d.GetOk("pxf_config.0.xmx"); ok {
		pc.Xmx = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk("pxf_config.0.xms"); ok {
		pc.Xms = &wrappers.Int64Value{Value: int64(v.(int))}
	}

	return pc, nil
}

func expandGreenplumBackgroundActivities(d *schema.ResourceData) (*greenplum.BackgroundActivitiesConfig, error) {
	result := &greenplum.BackgroundActivitiesConfig{
		QueryKillerScripts: &greenplum.QueryKillerScripts{},
	}

	if _, ok := d.GetOk("background_activities.0.analyze_and_vacuum"); ok {
		analyzeAndVacuum, err := expandGreenplumAnalyzeAndVacuum(d)
		if err != nil {
			return nil, err
		}
		result.AnalyzeAndVacuum = analyzeAndVacuum
	}
	if _, ok := d.GetOk("background_activities.0.query_killer_idle.0"); ok {
		result.QueryKillerScripts.Idle = expandGreenplumQueryKillerScript(d, "background_activities.0.query_killer_idle.0")
	}
	if _, ok := d.GetOk("background_activities.0.query_killer_idle_in_transaction.0"); ok {
		result.QueryKillerScripts.IdleInTransaction = expandGreenplumQueryKillerScript(d, "background_activities.0.query_killer_idle_in_transaction.0")
	}
	if _, ok := d.GetOk("background_activities.0.query_killer_long_running.0"); ok {
		result.QueryKillerScripts.LongRunning = expandGreenplumQueryKillerScript(d, "background_activities.0.query_killer_long_running.0")
	}

	return result, nil
}

func expandGreenplumAnalyzeAndVacuum(d *schema.ResourceData) (*greenplum.AnalyzeAndVacuum, error) {
	result := &greenplum.AnalyzeAndVacuum{}
	if v, ok := d.GetOk("background_activities.0.analyze_and_vacuum.0.start_time"); ok {
		hours, minutes, err := parseTimeOfDay(v.(string))
		if err != nil {
			return nil, err
		}
		result.Start = &greenplum.BackgroundActivityStartAt{
			Hours:   int64(hours),
			Minutes: int64(minutes),
		}
	}
	if v, ok := d.GetOk("background_activities.0.analyze_and_vacuum.0.analyze_timeout"); ok {
		result.AnalyzeTimeout = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk("background_activities.0.analyze_and_vacuum.0.vacuum_timeout"); ok {
		result.VacuumTimeout = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	return result, nil
}

// parse `HH:MM`
func parseTimeOfDay(time string) (int, int, error) {
	parts := strings.SplitN(time, ":", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid time format: expected 'HH:MM'")
	}
	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid hours in time of day: %w", err)
	}
	if hours < 0 || hours > 24 {
		return 0, 0, fmt.Errorf("invalid hours in time of day: should be in range 0..23")
	}
	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid minutes in time of day: %w", err)
	}
	if minutes < 0 || minutes > 59 {
		return 0, 0, fmt.Errorf("invalid minutes in time of day: should be in range 0..59")
	}
	return hours, minutes, nil
}

func expandGreenplumQueryKillerScript(d *schema.ResourceData, prefix string) *greenplum.QueryKiller {
	result := &greenplum.QueryKiller{}
	if v, ok := d.GetOk(prefix + ".enable"); ok {
		result.Enable = &wrappers.BoolValue{Value: v.(bool)}
	}
	if v, ok := d.GetOk(prefix + ".max_age"); ok {
		result.MaxAge = &wrappers.Int64Value{Value: int64(v.(int))}
	}
	if v, ok := d.GetOk(prefix + ".ignore_users"); ok {
		vS := v.([]interface{})
		for _, user := range vS {
			result.IgnoreUsers = append(result.IgnoreUsers, user.(string))
		}
	}
	return result
}

func parseGreenplumPoolingMode(s string) (greenplum.ConnectionPoolerConfig_PoolMode, error) {
	v, ok := greenplum.ConnectionPoolerConfig_PoolMode_value[s]
	if !ok {
		return 0, fmt.Errorf("value for 'pooling_mode' must be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeys(greenplum.ConnectionPoolerConfig_PoolMode_value)), s)
	}

	return greenplum.ConnectionPoolerConfig_PoolMode(v), nil
}

var mdbGreenplumSettingsFieldsInfo = newObjectFieldsInfo().
	addType(greenplum.GreenplumConfig6_22{}).
	addType(greenplum.GreenplumConfig6{})
