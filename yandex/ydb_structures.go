package yandex

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/ydb/v1"
)

func flattenYDBLocation(database *ydb.Database) ([]map[string]interface{}, error) {
	res := map[string]interface{}{}

	if t := database.GetRegionalDatabase(); t != nil {
		res["region"] = []map[string]interface{}{{"id": t.RegionId}}
	}

	if len(res) == 0 {
		return nil, nil
	}

	return []map[string]interface{}{res}, nil
}

func expandYDBLocationSpec(d *schema.ResourceData) (ydb.CreateDatabaseRequest_DatabaseType, error) {
	var db ydb.CreateDatabaseRequest

	if _, ok := d.GetOk("location.0.region"); ok {
		v := d.Get("location.0.region.0.id").(string)
		db.DatabaseType = &ydb.CreateDatabaseRequest_RegionalDatabase{
			RegionalDatabase: &ydb.RegionalDatabase{
				RegionId: v,
			},
		}
	}

	if db.DatabaseType != nil {
		return db.DatabaseType, nil
	}

	return &ydb.CreateDatabaseRequest_DedicatedDatabase{DedicatedDatabase: &ydb.DedicatedDatabase{}}, nil
}

func flattenYDBStorageConfig(storageConfig *ydb.StorageConfig) ([]map[string]interface{}, error) {
	if storageConfig == nil {
		return nil, nil
	}

	result := make([]map[string]interface{}, 0, len(storageConfig.StorageOptions))
	for _, option := range storageConfig.StorageOptions {
		result = append(result, map[string]interface{}{
			"storage_type_id": option.StorageTypeId,
			"group_count":     int(option.GroupCount),
		})
	}

	return result, nil
}

func expandYDBStorageConfigSpec(d *schema.ResourceData) (*ydb.StorageConfig, error) {
	storageConfig := d.Get("storage_config").([]interface{})
	if storageConfig == nil {
		return nil, nil
	}

	storageOptions := make([]*ydb.StorageOption, 0, len(storageConfig))
	for _, option := range storageConfig {
		storageOption := option.(map[string]interface{})
		storageOptions = append(storageOptions, &ydb.StorageOption{
			StorageTypeId: storageOption["storage_type_id"].(string),
			GroupCount:    int64(storageOption["group_count"].(int)),
		})
	}

	return &ydb.StorageConfig{StorageOptions: storageOptions}, nil
}

func flattenYDBScalePolicy(database *ydb.Database) ([]map[string]interface{}, error) {
	res := map[string]interface{}{}

	if sp := database.GetScalePolicy().GetFixedScale(); sp != nil {
		res["fixed_scale"] = []map[string]interface{}{{"size": int(sp.Size)}}
	}

	if sp := database.GetScalePolicy().GetAutoScale(); sp != nil {
		res["auto_scale"] = []map[string]any{{
			"max_size":        int(sp.MaxSize),
			"min_size":        int(sp.MinSize),
			"target_tracking": flattenYDBScalePolicyAutoScaleTargetTracking(sp.GetTargetTracking()),
		}}
	}

	if len(res) == 0 {
		return nil, nil
	}

	return []map[string]interface{}{res}, nil
}

func flattenYDBScalePolicyAutoScaleTargetTracking(v *ydb.ScalePolicy_AutoScale_TargetTracking) []map[string]any {
	if v == nil {
		return nil
	}

	return []map[string]any{
		{
			"cpu_utilization_percent": int(v.GetCpuUtilizationPercent()),
		},
	}
}

func expandYDBScalePolicySpec(d *schema.ResourceData) (*ydb.ScalePolicy, error) {
	if _, ok := d.GetOk("scale_policy.0.fixed_scale"); ok {
		v := d.Get("scale_policy.0.fixed_scale.0.size").(int)
		return &ydb.ScalePolicy{
			ScaleType: &ydb.ScalePolicy_FixedScale_{
				FixedScale: &ydb.ScalePolicy_FixedScale{
					Size: int64(v),
				},
			},
		}, nil
	}

	if _, ok := d.GetOk("scale_policy.0.auto_scale"); ok {
		minSize := d.Get("scale_policy.0.auto_scale.0.min_size").(int)
		maxSize := d.Get("scale_policy.0.auto_scale.0.max_size").(int)
		targetTrackingCPUUtilizationPercent := d.Get("scale_policy.0.auto_scale.0.target_tracking.0.cpu_utilization_percent").(int)
		return &ydb.ScalePolicy{
			ScaleType: &ydb.ScalePolicy_AutoScale_{
				AutoScale: &ydb.ScalePolicy_AutoScale{
					MinSize: int64(minSize),
					MaxSize: int64(maxSize),
					AutoScaleType: &ydb.ScalePolicy_AutoScale_TargetTracking_{
						TargetTracking: &ydb.ScalePolicy_AutoScale_TargetTracking{
							Target: &ydb.ScalePolicy_AutoScale_TargetTracking_CpuUtilizationPercent{
								CpuUtilizationPercent: int64(targetTrackingCPUUtilizationPercent),
							},
						},
					},
				},
			},
		}, nil
	}

	return nil, nil
}

func changeYDBnetworkIdSpec(d *schema.ResourceData) (string, error) {
	if _, ok := d.GetOk("network_id"); ok {
		v := d.Get("network_id").(string)
		return v, nil
	}
	return "", nil
}

func changeYDBsubnetIdsSpec(d *schema.ResourceData) ([]string, error) {
	if _, ok := d.GetOk("subnet_ids"); ok {
		v := d.Get("subnet_ids").(*schema.Set)
		var subnets []string
		for _, k := range v.List() {
			subnets = append(subnets, k.(string))
		}
		return subnets, nil
	}
	return nil, nil
}

func changeYDBsecurityGroupIdsSpec(d *schema.ResourceData) ([]string, error) {
	if _, ok := d.GetOk("security_group_ids"); ok {
		v := d.Get("security_group_ids").(*schema.Set)
		var securityGroups []string
		for _, k := range v.List() {
			securityGroups = append(securityGroups, k.(string))
		}
		return securityGroups, nil
	}
	return nil, nil
}
