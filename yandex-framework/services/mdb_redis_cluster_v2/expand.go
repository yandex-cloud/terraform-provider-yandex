package mdb_redis_cluster_v2

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/redis/v1"
	config "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/redis/v1/config"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	utils "github.com/yandex-cloud/terraform-provider-yandex/pkg/wrappers"
)

func expandAutoscaling(ctx context.Context, o types.Object) (*redis.DiskSizeAutoscaling, diag.Diagnostics) {

	if !utils.IsPresent(o) {
		return nil, nil
	}
	d := &DiskSizeAutoscaling{}
	diags := o.As(ctx, d, baseOptions)
	if diags.HasError() {
		return nil, diags
	}

	rs := &redis.DiskSizeAutoscaling{
		DiskSizeLimit: &wrappers.Int64Value{Value: datasize.ToBytes(d.DiskSizeLimit.ValueInt64())},
	}

	if utils.IsPresent(d.PlannedUsageThreshold) {
		rs.PlannedUsageThreshold = &wrappers.Int64Value{Value: d.PlannedUsageThreshold.ValueInt64()}
	}
	if utils.IsPresent(d.EmergencyUsageThreshold) {
		rs.EmergencyUsageThreshold = &wrappers.Int64Value{Value: d.EmergencyUsageThreshold.ValueInt64()}
	}
	return rs, diags
}

func expandAccess(ctx context.Context, a types.Object) (*redis.Access, diag.Diagnostics) {
	if !utils.IsPresent(a) {
		return nil, nil
	}

	access := &Access{}
	diags := a.As(ctx, access, baseOptions)
	if diags.HasError() {
		return nil, diags
	}
	result := &redis.Access{}

	if utils.IsPresent(access.WebSql) {
		result.WebSql = access.WebSql.ValueBool()
	}
	if utils.IsPresent(access.DataLens) {
		result.DataLens = access.DataLens.ValueBool()
	}

	return result, diags
}

func expandRedisConfig(d *Config) (*config.RedisConfig, error) {
	c := config.RedisConfig{
		Timeout:                         utils.Int64FromTF(d.Timeout),
		Databases:                       utils.Int64FromTF(d.Databases),
		SlowlogLogSlowerThan:            utils.Int64FromTF(d.SlowlogLogSlowerThan),
		SlowlogMaxLen:                   utils.Int64FromTF(d.SlowlogMaxLen),
		NotifyKeyspaceEvents:            utils.StringFromTF(d.NotifyKeyspaceEvents),
		MaxmemoryPercent:                utils.Int64FromTF(d.MaxmemoryPercent),
		LuaTimeLimit:                    utils.Int64FromTF(d.LuaTimeLimit),
		ReplBacklogSizePercent:          utils.Int64FromTF(d.ReplBacklogSizePercent),
		ClusterRequireFullCoverage:      utils.BoolFromTF(d.ClusterRequireFullCoverage),
		ClusterAllowReadsWhenDown:       utils.BoolFromTF(d.ClusterAllowReadsWhenDown),
		ClusterAllowPubsubshardWhenDown: utils.BoolFromTF(d.ClusterAllowPubsubshardWhenDown),
		LfuDecayTime:                    utils.Int64FromTF(d.LfuDecayTime),
		LfuLogFactor:                    utils.Int64FromTF(d.LfuLogFactor),
		TurnBeforeSwitchover:            utils.BoolFromTF(d.TurnBeforeSwitchover),
		AllowDataLoss:                   utils.BoolFromTF(d.AllowDataLoss),
		UseLuajit:                       utils.BoolFromTF(d.UseLuajit),
		IoThreadsAllowed:                utils.BoolFromTF(d.IoThreadsAllowed),
		ZsetMaxListpackEntries:          utils.Int64FromTF(d.ZsetMaxListpackEntries),
	}

	if utils.IsPresent(d.ClientOutputBufferLimitNormal) {
		expandedNormal, err := expandLimit(d.ClientOutputBufferLimitNormal.ValueString())
		if err != nil {
			return nil, err
		}
		if len(expandedNormal) != 0 {
			normalLimit := &config.RedisConfig_ClientOutputBufferLimit{
				HardLimit:   expandedNormal[0],
				SoftLimit:   expandedNormal[1],
				SoftSeconds: expandedNormal[2],
			}
			c.SetClientOutputBufferLimitNormal(normalLimit)
		}
	}
	if utils.IsPresent(d.ClientOutputBufferLimitPubsub) {
		expandedPubsub, err := expandLimit(d.ClientOutputBufferLimitPubsub.ValueString())
		if err != nil {
			return nil, err
		}
		if len(expandedPubsub) != 0 {
			pubsubLimit := &config.RedisConfig_ClientOutputBufferLimit{
				HardLimit:   expandedPubsub[0],
				SoftLimit:   expandedPubsub[1],
				SoftSeconds: expandedPubsub[2],
			}
			c.SetClientOutputBufferLimitPubsub(pubsubLimit)
		}
	}

	if utils.IsPresent(d.MaxmemoryPolicy) {
		mp, err := parseRedisMaxmemoryPolicy(d.MaxmemoryPolicy.ValueString())
		if err != nil {
			return nil, err
		}
		c.MaxmemoryPolicy = mp
	}

	return &c, nil
}

func expandLimit(limit string) ([]*wrappers.Int64Value, error) {
	if limit == "" {
		return nil, nil
	}
	vals := strings.Split(limit, " ")
	if len(vals) != 3 {
		return nil, fmt.Errorf("%q should be space-separated 3-values string", limit)
	}
	var res []*wrappers.Int64Value
	for _, val := range vals {
		parsed, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return nil, err
		}
		res = append(res, &wrappers.Int64Value{Value: parsed})
	}
	return res, nil
}
