package mdb_redis_cluster_v2

import (
	"context"

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
	if r == nil {
		return types.ObjectNull(ValkeyModulesType.AttributeTypes()), nil
	}
	a := ValkeyModules{}

	// Only create module objects if the API returns non-nil values
	if r.ValkeySearch != nil {
		a.ValkeySearch = &ValkeySearch{
			Enabled: types.BoolValue(r.ValkeySearch.GetEnabled()),
		}
		if r.ValkeySearch.ReaderThreads != nil {
			a.ValkeySearch.ReaderThreads = types.Int64Value(r.ValkeySearch.GetReaderThreads().GetValue())
		}
		if r.ValkeySearch.WriterThreads != nil {
			a.ValkeySearch.WriterThreads = types.Int64Value(r.ValkeySearch.GetWriterThreads().GetValue())
		}
	}

	if r.ValkeyJson != nil {
		a.ValkeyJson = &ValkeyJson{
			Enabled: types.BoolValue(r.ValkeyJson.GetEnabled()),
		}
	}

	if r.ValkeyBloom != nil {
		a.ValkeyBloom = &ValkeyBloom{
			Enabled: types.BoolValue(r.ValkeyBloom.GetEnabled()),
		}
	}

	// If no modules are present, return null
	if a.ValkeySearch == nil && a.ValkeyJson == nil && a.ValkeyBloom == nil {
		return types.ObjectNull(ValkeyModulesType.AttributeTypes()), nil
	}

	// If all modules are disabled (default values), return null to maintain consistency
	// with the plan when no modules are specified in the configuration
	allDisabled := true
	if a.ValkeySearch != nil && a.ValkeySearch.Enabled.ValueBool() {
		allDisabled = false
	}
	if a.ValkeyJson != nil && a.ValkeyJson.Enabled.ValueBool() {
		allDisabled = false
	}
	if a.ValkeyBloom != nil && a.ValkeyBloom.Enabled.ValueBool() {
		allDisabled = false
	}

	if allDisabled {
		return types.ObjectNull(ValkeyModulesType.AttributeTypes()), nil
	}

	return types.ObjectValueFrom(ctx, ValkeyModulesType.AttributeTypes(), a)
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
