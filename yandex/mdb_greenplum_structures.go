package yandex

import (
	"fmt"
	"log"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/greenplum/v1"
	"google.golang.org/genproto/googleapis/type/timeofday"
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
	if version == "6.17" {
		return "greenplum_config_6_17"
	}
	return "greenplum_config_6_19"
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
	}
	return []map[string]interface{}{out}
}

func flattenBackupWindowsStart(c *greenplum.GreenplumConfig) []map[string]interface{} {
	out := map[string]interface{}{}
	if c != nil && c.BackupWindowStart != nil {
		out["hours"] = c.BackupWindowStart.Hours
		out["minutes"] = c.BackupWindowStart.Minutes
	}
	return []map[string]interface{}{out}
}

func flattenGreenplumClusterConfig(c *greenplum.ClusterConfigSet) (map[string]string, error) {
	var gpConfig interface{}

	if cf, ok := c.GreenplumConfig.(*greenplum.ClusterConfigSet_GreenplumConfigSet_6_17); ok {
		gpConfig = cf.GreenplumConfigSet_6_17.UserConfig
	}
	if cf, ok := c.GreenplumConfig.(*greenplum.ClusterConfigSet_GreenplumConfigSet_6_19); ok {
		gpConfig = cf.GreenplumConfigSet_6_19.UserConfig
	}

	settings, err := flattenResourceGenerateMapS(gpConfig, false, mdbGreenplumSettingsFieldsInfo, false, true, nil)
	if err != nil {
		return nil, err
	}

	return settings, err
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

	return out
}

func expandGreenplumUpdatePath(d *schema.ResourceData, settingNames []string) []string {
	mdbGreenplumUpdateFieldsMap := map[string]string{
		"name":                         "name",
		"description":                  "description",
		"labels":                       "labels",
		"access.0.data_lens":           "config.access.data_lens",
		"access.0.web_sql":             "config.access.web_sql",
		"backup_window_start":          "config.backup_window_start",
		"deletion_protection":          "deletion_protection",
		"security_group_ids":           "security_group_ids",
		"pooler_config.0.pooling_mode": "config_spec.pool.mode",
		"pooler_config.0.pool_size":    "config_spec.pool.size",
		"pooler_config.0.pool_client_idle_timeout": "config_spec.pool.client_idle_timeout",
	}

	updatePath := []string{}
	for field, path := range mdbGreenplumUpdateFieldsMap {
		if d.HasChange(field) {
			updatePath = append(updatePath, path)
		}
	}

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

	gpConfig617, gpConfig619, settingNames, err := expandGreenplumConfigSpecGreenplumConfig(d)
	if err != nil {
		return nil, nil, err
	}

	configSpec := &greenplum.ConfigSpec{Pool: poolerConfig}
	if gpConfig617 != nil {
		configSpec.GreenplumConfig = gpConfig617
	} else {
		configSpec.GreenplumConfig = gpConfig619
	}

	return configSpec, settingNames, nil
}

func expandGreenplumConfigSpecGreenplumConfig(d *schema.ResourceData) (*greenplum.ConfigSpec_GreenplumConfig_6_17, *greenplum.ConfigSpec_GreenplumConfig_6_19, []string, error) {
	version := d.Get("version").(string)
	if version == "6.17" {
		cfg := &greenplum.ConfigSpec_GreenplumConfig_6_17{
			GreenplumConfig_6_17: &greenplum.GreenplumConfig6_17{},
		}
		fields, err := expandResourceGenerateNonSkippedFields(mdbGreenplumSettingsFieldsInfo, d, cfg.GreenplumConfig_6_17, "greenplum_config.", true)
		if err != nil {
			return nil, nil, nil, err
		}
		return cfg, nil, fields, nil
	} else if version == "6.19" {
		cfg := &greenplum.ConfigSpec_GreenplumConfig_6_19{
			GreenplumConfig_6_19: &greenplum.GreenplumConfig6_19{},
		}

		settingNames, err := expandResourceGenerateNonSkippedFields(mdbGreenplumSettingsFieldsInfo, d, cfg.GreenplumConfig_6_19, "greenplum_config.", true)
		if err != nil {
			return nil, nil, []string{}, err
		}
		log.Printf("[SPECIAL DEBUG] %v", cfg.GreenplumConfig_6_19)
		log.Printf("[SPECIAL DEBUG] %v", settingNames)
		return nil, cfg, settingNames, nil
	}

	return nil, nil, nil, fmt.Errorf("unknown Greenplum version: '%s' but '6.17' and '6.19' are only available", version)
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

func parseGreenplumPoolingMode(s string) (greenplum.ConnectionPoolerConfig_PoolMode, error) {
	v, ok := greenplum.ConnectionPoolerConfig_PoolMode_value[s]
	if !ok {
		return 0, fmt.Errorf("value for 'pooling_mode' must be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeys(greenplum.ConnectionPoolerConfig_PoolMode_value)), s)
	}

	return greenplum.ConnectionPoolerConfig_PoolMode(v), nil
}

var mdbGreenplumSettingsFieldsInfo = newObjectFieldsInfo().
	addType(greenplum.GreenplumConfig6_17{}).
	addType(greenplum.GreenplumConfig6_19{})
