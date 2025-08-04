package mdb_redis_cluster_v2

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/redis/v1"
	config "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/redis/v1/config"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
	utils "github.com/yandex-cloud/terraform-provider-yandex/pkg/wrappers"
	"google.golang.org/genproto/protobuf/field_mask"
)

func updateRedisClusterParams(ctx context.Context, sdk *ycsdk.SDK, diagnostics *diag.Diagnostics, plan, state *Cluster) {
	var diags diag.Diagnostics
	req := &redis.UpdateClusterRequest{
		ClusterId: state.ID.ValueString(),
		UpdateMask: &field_mask.FieldMask{
			Paths: []string{},
		},
	}
	if !plan.Name.Equal(state.Name) {
		req.Name = plan.Name.ValueString()
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "name")
	}

	if !plan.PersistenceMode.Equal(state.PersistenceMode) {
		mode, err := parsePersistenceMode(plan.PersistenceMode.ValueString())
		if err != nil {
			diagnostics.AddError(
				"Wrong attribute value",
				err.Error(),
			)
		}

		req.PersistenceMode = mode
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "persistence_mode")
	}

	if !plan.AnnounceHostnames.Equal(state.AnnounceHostnames) {
		req.AnnounceHostnames = plan.AnnounceHostnames.ValueBool()

		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "announce_hostnames")

	}

	if !plan.AuthSentinel.Equal(state.AuthSentinel) {
		req.AuthSentinel = plan.AuthSentinel.ValueBool()

		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "auth_sentinel")
	}

	if !plan.Labels.Equal(state.Labels) {
		var labels map[string]string
		diagnostics.Append(plan.Labels.ElementsAs(ctx, &labels, false)...)
		req.Labels = labels
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "labels")

	}

	if !plan.Description.Equal(state.Description) {
		req.Description = plan.Description.ValueString()
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "description")

	}

	if req.ConfigSpec == nil {
		req.ConfigSpec = &redis.ConfigSpec{}
	}

	if !plan.Resources.Equal(state.Resources) {
		req.ConfigSpec.Resources = mdbcommon.ExpandResources[redis.Resources](ctx, plan.Resources, diagnostics)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "config_spec.resources")
	}

	if !plan.DiskSizeAutoscaling.Equal(state.DiskSizeAutoscaling) {
		req.ConfigSpec.DiskSizeAutoscaling, diags = expandAutoscaling(ctx, plan.DiskSizeAutoscaling)
		diagnostics.Append(diags...)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "config_spec.disk_size_autoscaling")
	}

	mask := plan.Config.EvalUpdateMask(state.Config)
	if msk := mask; len(msk) > 0 {
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, mask...)
		conf, err := expandRedisConfig(plan.Config)
		if err != nil {
			diagnostics.AddError(
				"Wrong attribute value",
				err.Error(),
			)
		}

		req.ConfigSpec.Version = plan.Config.Version.ValueString()
		req.ConfigSpec.Redis = conf
	}
	if !plan.Config.BackupWindowStart.Equal(state.Config.BackupWindowStart) {
		req.ConfigSpec.BackupWindowStart = mdbcommon.ExpandBackupWindow(ctx, plan.Config.BackupWindowStart, diagnostics)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "config_spec.backup_window_start")
	}
	if !plan.Config.BackupRetainPeriodDays.Equal(state.Config.BackupRetainPeriodDays) {
		req.ConfigSpec.BackupRetainPeriodDays = utils.Int64FromTF(plan.Config.BackupRetainPeriodDays)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "config_spec.backup_retain_period_days")
	}

	if !plan.SecurityGroupIDs.Equal(state.SecurityGroupIDs) {
		var securityGroupIds []string
		diagnostics.Append(plan.SecurityGroupIDs.ElementsAs(ctx, &securityGroupIds, false)...)

		req.SecurityGroupIds = securityGroupIds
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "security_group_ids")

	}

	if !plan.DeletionProtection.Equal(state.DeletionProtection) {
		req.DeletionProtection = plan.DeletionProtection.ValueBool()
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "deletion_protection")
	}

	if !plan.MaintenanceWindow.Equal(state.MaintenanceWindow) {
		req.MaintenanceWindow = mdbcommon.ExpandClusterMaintenanceWindow[
			redis.MaintenanceWindow,
			redis.WeeklyMaintenanceWindow,
			redis.AnytimeMaintenanceWindow,
			redis.WeeklyMaintenanceWindow_WeekDay,
		](ctx, plan.MaintenanceWindow, diagnostics)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "maintenance_window")

	}

	if !plan.Access.Equal(state.Access) {
		req.ConfigSpec.Access, diags = expandAccess(ctx, plan.Access)
		diagnostics.Append(diags...)
		req.UpdateMask.Paths = append(
			req.UpdateMask.Paths,
			"config_spec.access.web_sql",
			"config_spec.access.data_lens",
			"config_spec.access.data_transfer",
			"config_spec.access.serverless",
		)
	}

	if diagnostics.HasError() {
		return
	}
	password := plan.Config.Password.ValueString()

	tflog.Debug(ctx, "Update Params", map[string]interface{}{
		"update_mask": req.UpdateMask.Paths,
	})

	if len(req.UpdateMask.Paths) == 0 && password == "" {
		return
	}

	if len(req.UpdateMask.Paths) != 0 {
		redisAPI.UpdateCluster(ctx, sdk, diagnostics, req)
		if diagnostics.HasError() {
			return
		}
	}

	if password != "" && !plan.Config.Password.Equal(state.Config.Password) {
		reqPasswordUpdate := &redis.UpdateClusterRequest{
			ClusterId: state.ID.ValueString(),
			ConfigSpec: &redis.ConfigSpec{
				Redis: &config.RedisConfig{Password: password},
			},
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{"config_spec.redis.password"},
			},
		}
		redisAPI.UpdateCluster(ctx, sdk, diagnostics, reqPasswordUpdate)
		if diagnostics.HasError() {
			return
		}
	}
}

func (c Config) EvalUpdateMask(o *Config) []string {
	var updateMask []string
	if !c.Timeout.Equal(o.Timeout) {
		updateMask = append(updateMask, "timeout")
	}
	if !c.MaxmemoryPolicy.Equal(o.MaxmemoryPolicy) {
		updateMask = append(updateMask, "maxmemory_policy")
	}
	if !c.NotifyKeyspaceEvents.Equal(o.NotifyKeyspaceEvents) {
		updateMask = append(updateMask, "notify_keyspace_events")
	}
	if !c.SlowlogLogSlowerThan.Equal(o.SlowlogLogSlowerThan) {
		updateMask = append(updateMask, "slowlog_log_slower_than")
	}
	if !c.SlowlogMaxLen.Equal(o.SlowlogMaxLen) {
		updateMask = append(updateMask, "slowlog_max_len")
	}
	if !c.Databases.Equal(o.Databases) {
		updateMask = append(updateMask, "databases")
	}
	if !c.MaxmemoryPercent.Equal(o.MaxmemoryPercent) {
		updateMask = append(updateMask, "maxmemory_percent")
	}
	if !c.ClientOutputBufferLimitNormal.Equal(o.ClientOutputBufferLimitNormal) {
		updateMask = append(updateMask, "client_output_buffer_limit_normal")
	}
	if !c.ClientOutputBufferLimitPubsub.Equal(o.ClientOutputBufferLimitPubsub) {
		updateMask = append(updateMask, "client_output_buffer_limit_pubsub")
	}
	if !c.UseLuajit.Equal(o.UseLuajit) {
		updateMask = append(updateMask, "use_luajit")
	}
	if !c.IoThreadsAllowed.Equal(o.IoThreadsAllowed) {
		updateMask = append(updateMask, "io_threads_allowed")
	}
	if !c.LuaTimeLimit.Equal(o.LuaTimeLimit) {
		updateMask = append(updateMask, "lua_time_limit")
	}
	if !c.ReplBacklogSizePercent.Equal(o.ReplBacklogSizePercent) {
		updateMask = append(updateMask, "repl_backlog_size_percent")
	}
	if !c.ClusterRequireFullCoverage.Equal(o.ClusterRequireFullCoverage) {
		updateMask = append(updateMask, "cluster_require_full_coverage")
	}
	if !c.ClusterAllowReadsWhenDown.Equal(o.ClusterAllowReadsWhenDown) {
		updateMask = append(updateMask, "cluster_allow_reads_when_down")
	}
	if !c.ClusterAllowPubsubshardWhenDown.Equal(o.ClusterAllowPubsubshardWhenDown) {
		updateMask = append(updateMask, "cluster_allow_pubsubshard_when_down")
	}
	if !c.LfuDecayTime.Equal(o.LfuDecayTime) {
		updateMask = append(updateMask, "lfu_decay_time")
	}
	if !c.LfuLogFactor.Equal(o.LfuLogFactor) {
		updateMask = append(updateMask, "lfu_log_factor")
	}
	if !c.TurnBeforeSwitchover.Equal(o.TurnBeforeSwitchover) {
		updateMask = append(updateMask, "turn_before_switchover")
	}
	if !c.AllowDataLoss.Equal(o.AllowDataLoss) {
		updateMask = append(updateMask, "allow_data_loss")
	}
	if !c.BackupRetainPeriodDays.Equal(o.BackupRetainPeriodDays) {
		updateMask = append(updateMask, "backup_retain_period_days")
	}
	if !c.ZsetMaxListpackEntries.Equal(o.ZsetMaxListpackEntries) {
		updateMask = append(updateMask, "zset_max_listpack_entries")
	}

	for i := range updateMask {
		updateMask[i] = "config_spec.redis." + updateMask[i]
	}

	if !c.Version.Equal(o.Version) {
		updateMask = append(updateMask, "config_spec.version")
	}
	return updateMask
}
