package mdb_redis_cluster_v2

import (
	"context"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/redis/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
	utils "github.com/yandex-cloud/terraform-provider-yandex/pkg/wrappers"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

func redisClusterCreatePassword(config *Config, passwordWo types.String) string {
	if !passwordWo.IsNull() && !passwordWo.IsUnknown() {
		return passwordWo.ValueString()
	}
	return config.Password.ValueString()
}

func prepareCreateRedisRequest(ctx context.Context, meta *provider_config.Config, diagnostics *diag.Diagnostics, plan *Cluster, passwordWo types.String, hostSpecs []*redis.HostSpec) *redis.CreateClusterRequest {
	planConfig := plan.Config
	var diags diag.Diagnostics

	var labels map[string]string
	diagnostics.Append(plan.Labels.ElementsAs(ctx, &labels, false)...)
	folderID, d := validate.FolderID(plan.FolderID, &meta.ProviderState)
	diagnostics.Append(d)

	e := plan.Environment
	env := mdbcommon.ExpandEnvironment[redis.Cluster_Environment](ctx, e, diagnostics)

	conf, err := expandRedisConfig(planConfig)
	if err != nil {
		diagnostics.AddError(
			"Wrong attribute value",
			err.Error(),
		)
		return nil
	}
	conf.Password = redisClusterCreatePassword(planConfig, passwordWo)

	resources := mdbcommon.ExpandResources[redis.Resources](ctx, plan.Resources, diagnostics)

	autoscaling, diags := expandAutoscaling(ctx, plan.DiskSizeAutoscaling)
	diagnostics.Append(diags...)

	modules, _, diags := expandModules(ctx, plan.Modules)
	diagnostics.Append(diags...)

	access, diags := expandAccess(ctx, plan.Access)
	diagnostics.Append(diags...)

	backupWindow := mdbcommon.ExpandBackupWindow(ctx, planConfig.BackupWindowStart, diagnostics)

	configSpec := &redis.ConfigSpec{
		Version:                planConfig.Version.ValueString(),
		Resources:              resources,
		BackupWindowStart:      backupWindow,
		Access:                 access,
		Redis:                  conf,
		DiskSizeAutoscaling:    autoscaling,
		Modules:                modules,
		BackupRetainPeriodDays: utils.Int64FromTF(planConfig.BackupRetainPeriodDays),
	}

	securityGroupIds := mdbcommon.ExpandSecurityGroupIds(ctx, plan.SecurityGroupIDs, &diags)

	networkID, d := validate.NetworkId(plan.NetworkID, &meta.ProviderState)
	diagnostics.Append(d)

	persistenceMode, err := parsePersistenceMode(plan.PersistenceMode.ValueString())
	if err != nil {
		diagnostics.AddError(
			"Wrong attribute value",
			err.Error(),
		)
	}

	maintenanceWindow := mdbcommon.ExpandClusterMaintenanceWindow[
		redis.MaintenanceWindow,
		redis.WeeklyMaintenanceWindow,
		redis.AnytimeMaintenanceWindow,
		redis.WeeklyMaintenanceWindow_WeekDay,
	](ctx, plan.MaintenanceWindow, diagnostics)

	req := redis.CreateClusterRequest{
		FolderId:            folderID,
		Name:                plan.Name.ValueString(),
		Description:         plan.Description.ValueString(),
		Labels:              labels,
		Environment:         env,
		ConfigSpec:          configSpec,
		HostSpecs:           hostSpecs,
		NetworkId:           networkID,
		Sharded:             plan.Sharded.ValueBool(),
		SecurityGroupIds:    securityGroupIds,
		TlsEnabled:          &wrappers.BoolValue{Value: plan.TlsEnabled.ValueBool()},
		DeletionProtection:  plan.DeletionProtection.ValueBool(),
		PersistenceMode:     persistenceMode,
		AnnounceHostnames:   plan.AnnounceHostnames.ValueBool(),
		MaintenanceWindow:   maintenanceWindow,
		AuthSentinel:        plan.AuthSentinel.ValueBool(),
		DiskEncryptionKeyId: mdbcommon.ExpandStringWrapper(ctx, plan.DiskEncryptionKeyId, diagnostics),
	}
	return &req
}
