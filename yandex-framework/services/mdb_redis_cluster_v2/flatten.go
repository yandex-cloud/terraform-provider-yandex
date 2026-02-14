package mdb_redis_cluster_v2

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/redis/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
)

func flattenAutoscaling(ctx context.Context, r *redis.DiskSizeAutoscaling) (types.Object, diag.Diagnostics) {
	if r == nil {
		return types.ObjectNull(DiskSizeAutoscalingType.AttributeTypes()), nil
	}
	a := DiskSizeAutoscaling{
		DiskSizeLimit:           types.Int64Value(datasize.ToGigabytes(r.GetDiskSizeLimit().GetValue())),
		PlannedUsageThreshold:   types.Int64Value(r.GetPlannedUsageThreshold().GetValue()),
		EmergencyUsageThreshold: types.Int64Value(r.GetEmergencyUsageThreshold().GetValue()),
	}

	return types.ObjectValueFrom(ctx, DiskSizeAutoscalingType.AttributeTypes(), a)
}

func flattenModules(ctx context.Context, r *redis.ValkeyModules) (types.Object, diag.Diagnostics) {
	// Build attribute map for ObjectValue
	attrs := map[string]attr.Value{}

	// Flatten ValkeySearch - default to enabled=false if not present
	if r.GetValkeySearch() != nil {
		searchAttrs := map[string]attr.Value{
			"enabled":        types.BoolValue(r.ValkeySearch.Enabled),
			"reader_threads": types.Int64Value(r.ValkeySearch.GetReaderThreads().GetValue()),
			"writer_threads": types.Int64Value(r.ValkeySearch.GetWriterThreads().GetValue()),
		}
		searchObj, diags := types.ObjectValue(
			ValkeyModulesType.AttrTypes["valkey_search"].(types.ObjectType).AttrTypes,
			searchAttrs,
		)
		if diags.HasError() {
			return types.ObjectNull(ValkeyModulesType.AttributeTypes()), diags
		}
		attrs["valkey_search"] = searchObj
	} else {
		// Default: enabled=false, threads=0
		searchAttrs := map[string]attr.Value{
			"enabled":        types.BoolValue(false),
			"reader_threads": types.Int64Value(0),
			"writer_threads": types.Int64Value(0),
		}
		searchObj, diags := types.ObjectValue(
			ValkeyModulesType.AttrTypes["valkey_search"].(types.ObjectType).AttrTypes,
			searchAttrs,
		)
		if diags.HasError() {
			return types.ObjectNull(ValkeyModulesType.AttributeTypes()), diags
		}
		attrs["valkey_search"] = searchObj
	}

	// Flatten ValkeyJson - default to enabled=false if not present
	if r.GetValkeyJson() != nil {
		jsonAttrs := map[string]attr.Value{
			"enabled": types.BoolValue(r.ValkeyJson.Enabled),
		}
		jsonObj, diags := types.ObjectValue(
			ValkeyModulesType.AttrTypes["valkey_json"].(types.ObjectType).AttrTypes,
			jsonAttrs,
		)
		if diags.HasError() {
			return types.ObjectNull(ValkeyModulesType.AttributeTypes()), diags
		}
		attrs["valkey_json"] = jsonObj
	} else {
		// Default: enabled=false
		jsonAttrs := map[string]attr.Value{
			"enabled": types.BoolValue(false),
		}
		jsonObj, diags := types.ObjectValue(
			ValkeyModulesType.AttrTypes["valkey_json"].(types.ObjectType).AttrTypes,
			jsonAttrs,
		)
		if diags.HasError() {
			return types.ObjectNull(ValkeyModulesType.AttributeTypes()), diags
		}
		attrs["valkey_json"] = jsonObj
	}

	// Flatten ValkeyBloom - default to enabled=false if not present
	if r.GetValkeyBloom() != nil {
		bloomAttrs := map[string]attr.Value{
			"enabled": types.BoolValue(r.ValkeyBloom.Enabled),
		}
		bloomObj, diags := types.ObjectValue(
			ValkeyModulesType.AttrTypes["valkey_bloom"].(types.ObjectType).AttrTypes,
			bloomAttrs,
		)
		if diags.HasError() {
			return types.ObjectNull(ValkeyModulesType.AttributeTypes()), diags
		}
		attrs["valkey_bloom"] = bloomObj
	} else {
		// Default: enabled=false
		bloomAttrs := map[string]attr.Value{
			"enabled": types.BoolValue(false),
		}
		bloomObj, diags := types.ObjectValue(
			ValkeyModulesType.AttrTypes["valkey_bloom"].(types.ObjectType).AttrTypes,
			bloomAttrs,
		)
		if diags.HasError() {
			return types.ObjectNull(ValkeyModulesType.AttributeTypes()), diags
		}
		attrs["valkey_bloom"] = bloomObj
	}

	return types.ObjectValue(ValkeyModulesType.AttributeTypes(), attrs)
}

func flattenAccess(ctx context.Context, r *redis.Access) (types.Object, diag.Diagnostics) {
	if r == nil {
		return types.ObjectNull(AccessType.AttributeTypes()), nil
	}
	a := Access{
		WebSql:   types.BoolValue(r.WebSql),
		DataLens: types.BoolValue(r.DataLens),
	}
	return types.ObjectValueFrom(ctx, AccessType.AttributeTypes(), a)
}

func FlattenConfig(cc *redis.ClusterConfig) Config {
	c := cc.Redis.EffectiveConfig

	res := Config{
		MaxmemoryPolicy:      types.StringValue(c.GetMaxmemoryPolicy().String()),
		Timeout:              types.Int64Value(c.GetTimeout().GetValue()),
		NotifyKeyspaceEvents: types.StringValue(c.GetNotifyKeyspaceEvents()),
		SlowlogLogSlowerThan: types.Int64Value(c.GetSlowlogLogSlowerThan().GetValue()),
		SlowlogMaxLen:        types.Int64Value(c.GetSlowlogMaxLen().GetValue()),
		Databases:            types.Int64Value(c.GetDatabases().GetValue()),
		MaxmemoryPercent:     types.Int64Value(c.GetMaxmemoryPercent().GetValue()),
		ClientOutputBufferLimitNormal: types.StringValue(limitToStr(
			c.GetClientOutputBufferLimitNormal().GetHardLimit(),
			c.GetClientOutputBufferLimitNormal().GetSoftLimit(),
			c.GetClientOutputBufferLimitNormal().GetSoftSeconds(),
		)),
		ClientOutputBufferLimitPubsub: types.StringValue(limitToStr(
			c.GetClientOutputBufferLimitPubsub().GetHardLimit(),
			c.GetClientOutputBufferLimitPubsub().GetSoftLimit(),
			c.GetClientOutputBufferLimitPubsub().GetSoftSeconds(),
		)),
		UseLuajit:                       types.BoolValue(c.GetUseLuajit().GetValue()),
		IoThreadsAllowed:                types.BoolValue(c.GetIoThreadsAllowed().GetValue()),
		Version:                         types.StringValue(cc.Version),
		LuaTimeLimit:                    types.Int64Value(c.GetLuaTimeLimit().GetValue()),
		ReplBacklogSizePercent:          types.Int64Value(c.GetReplBacklogSizePercent().GetValue()),
		ClusterRequireFullCoverage:      types.BoolValue(c.GetClusterRequireFullCoverage().GetValue()),
		ClusterAllowReadsWhenDown:       types.BoolValue(c.GetClusterAllowReadsWhenDown().GetValue()),
		ClusterAllowPubsubshardWhenDown: types.BoolValue(c.GetClusterAllowPubsubshardWhenDown().GetValue()),
		LfuDecayTime:                    types.Int64Value(c.GetLfuDecayTime().GetValue()),
		LfuLogFactor:                    types.Int64Value(c.GetLfuLogFactor().GetValue()),
		TurnBeforeSwitchover:            types.BoolValue(c.GetTurnBeforeSwitchover().GetValue()),
		AllowDataLoss:                   types.BoolValue(c.GetAllowDataLoss().GetValue()),
		BackupRetainPeriodDays:          types.Int64Value(cc.GetBackupRetainPeriodDays().GetValue()),
		ZsetMaxListpackEntries:          types.Int64Value(c.GetZsetMaxListpackEntries().GetValue()),
	}

	return res
}
