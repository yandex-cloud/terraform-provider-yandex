package mdb_sharded_postgresql_cluster

import (
	"context"
	"fmt"
	"maps"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/spqr/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	protobuf_adapter "github.com/yandex-cloud/terraform-provider-yandex/pkg/adapters/protobuf"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
)

func prepareConfigChange(ctx context.Context, plan, state *Config) (*spqr.ConfigSpec, []string, diag.Diagnostics) {
	var updateMaskPaths []string
	config := &spqr.ConfigSpec{SpqrSpec: &spqr.SpqrSpec{}}
	diags := diag.Diagnostics{}

	if !plan.Access.Equal(state.Access) {
		config.SetAccess(mdbcommon.ExpandAccess[*spqr.Access](ctx, plan.Access, &diags))
		updateMaskPaths = append(
			updateMaskPaths,
			"config_spec.access.web_sql",
			"config_spec.access.data_lens",
			"config_spec.access.data_transfer",
			"config_spec.access.serverless",
		)
	}

	if !plan.BackupWindowStart.Equal(state.BackupWindowStart) {
		config.SetBackupWindowStart(mdbcommon.ExpandBackupWindow(ctx, plan.BackupWindowStart, &diags))
		updateMaskPaths = append(updateMaskPaths, "config_spec.backup_window_start")
	}

	if !plan.BackupRetainPeriodDays.Equal(state.BackupRetainPeriodDays) {
		config.SetBackupRetainPeriodDays(expandBackupRetainPeriodDays(ctx, plan.BackupRetainPeriodDays, &diags))
		updateMaskPaths = append(updateMaskPaths, "config_spec.backup_retain_period_days")
	}

	a := protobuf_adapter.NewProtobufMapDataAdapter()

	// FIXME: a.Fill does not work on common settings like log_level or console_password
	if !plan.SPQRConfig.Common.Equal(state.SPQRConfig.Common) {
		logLevel := spqr.LogLevel(getAttrOrDefault(ctx, &diags, plan.SPQRConfig.Common, "log_level", types.Int64Value(0)).(types.Int64).ValueInt64())
		config.SpqrSpec.LogLevel = spqr.LogLevel(logLevel)
		config.SpqrSpec.ConsolePassword = getAttrOrDefault(ctx, &diags, plan.SPQRConfig.Common, "console_password", types.StringValue("")).(types.String).ValueString()
		updateMaskPaths = append(updateMaskPaths, appendSettingsToUpdateMask(ctx, &diags, "", &state.SPQRConfig.Common, &plan.SPQRConfig.Common)...)
	}

	if !plan.SPQRConfig.Balancer.Equal(state.SPQRConfig.Balancer) {
		config.SpqrSpec.Balancer = &spqr.BalancerSettings{}
		a.Fill(ctx, config.SpqrSpec.Balancer, plan.SPQRConfig.Balancer.PrimitiveElements(ctx, &diags), &diags)
		updateMaskPaths = append(updateMaskPaths, appendSettingsToUpdateMask(ctx, &diags, "balancer", &state.SPQRConfig.Balancer, &plan.SPQRConfig.Balancer)...)
	}

	var newResources *spqr.Resources
	var newRouterSettings *spqr.RouterSettings
	if r := plan.SPQRConfig.Router; r != nil {
		stateCfg := mdbcommon.NewSettingsMapNull()
		if state.SPQRConfig.Router != nil {
			stateCfg = state.SPQRConfig.Router.Config
		}
		if !r.Config.Equal(stateCfg) {
			newRouterSettings = &spqr.RouterSettings{}
			a.Fill(ctx, newRouterSettings, plan.SPQRConfig.Router.Config.PrimitiveElements(ctx, &diags), &diags)
			updateMaskPaths = append(updateMaskPaths, appendSettingsToUpdateMask(ctx, &diags, "router", &stateCfg, &r.Config)...)
		}
		if !state.SPQRConfig.Router.Resources.IsNull() && !r.Resources.Equal(state.SPQRConfig.Router.Resources) {
			newResources = mdbcommon.ExpandResources[spqr.Resources](ctx, r.Resources, &diags)
			updateMaskPaths = append(updateMaskPaths, "config_spec.spqr_spec.router.resources")
		}
	}
	if newResources != nil || newRouterSettings != nil {
		config.SpqrSpec.Router = &spqr.SpqrSpec_Router{
			Resources: newResources,
			Config:    newRouterSettings,
		}
	}

	newResources = nil
	var newCoordinatorSettings *spqr.CoordinatorSettings
	if c := plan.SPQRConfig.Coordinator; c != nil {
		stateCfg := mdbcommon.NewSettingsMapNull()
		if state.SPQRConfig.Coordinator != nil {
			stateCfg = state.SPQRConfig.Coordinator.Config
		}
		if !c.Config.Equal(stateCfg) {
			newCoordinatorSettings = &spqr.CoordinatorSettings{}
			a.Fill(ctx, newCoordinatorSettings, plan.SPQRConfig.Coordinator.Config.PrimitiveElements(ctx, &diags), &diags)
			updateMaskPaths = append(updateMaskPaths, appendSettingsToUpdateMask(ctx, &diags, "coordinator", &stateCfg, &c.Config)...)
		}
		if !state.SPQRConfig.Coordinator.Resources.IsNull() && !c.Resources.Equal(state.SPQRConfig.Coordinator.Resources) {
			newResources = mdbcommon.ExpandResources[spqr.Resources](ctx, c.Resources, &diags)
			updateMaskPaths = append(updateMaskPaths, "config_spec.spqr_spec.coordinator.resources")
		}
	}
	if newResources != nil || newCoordinatorSettings != nil {
		config.SpqrSpec.Coordinator = &spqr.SpqrSpec_Coordinator{
			Config:    newCoordinatorSettings,
			Resources: newResources,
		}
	}

	newResources = nil
	var newInfraRouterSettings *spqr.RouterSettings = nil
	var newInfraCoordinatorSettings *spqr.CoordinatorSettings = nil
	if i := plan.SPQRConfig.Infra; i != nil {
		stateRouterCfg := mdbcommon.NewSettingsMapNull()
		stateCoordCfg := mdbcommon.NewSettingsMapNull()
		if state.SPQRConfig.Infra != nil {
			stateRouterCfg = state.SPQRConfig.Infra.Router
			stateCoordCfg = state.SPQRConfig.Infra.Coordinator
		}
		if !i.Router.Equal(stateRouterCfg) {
			newInfraRouterSettings = &spqr.RouterSettings{}
			a.Fill(ctx, newInfraRouterSettings, plan.SPQRConfig.Infra.Router.PrimitiveElements(ctx, &diags), &diags)
			updateMaskPaths = append(updateMaskPaths, appendSettingsToUpdateMask(ctx, &diags, "infra.router", &stateRouterCfg, &i.Router)...)
		}
		if !i.Coordinator.Equal(stateCoordCfg) {
			newInfraCoordinatorSettings = &spqr.CoordinatorSettings{}
			a.Fill(ctx, newInfraCoordinatorSettings, plan.SPQRConfig.Infra.Coordinator.PrimitiveElements(ctx, &diags), &diags)
			updateMaskPaths = append(updateMaskPaths, appendSettingsToUpdateMask(ctx, &diags, "infra.coordinator", &stateCoordCfg, &i.Coordinator)...)
		}
		if state.SPQRConfig.Infra == nil || !state.SPQRConfig.Infra.Resources.IsNull() && !i.Resources.Equal(state.SPQRConfig.Infra.Resources) {
			newResources = mdbcommon.ExpandResources[spqr.Resources](ctx, i.Resources, &diags)
			updateMaskPaths = append(updateMaskPaths, "config_spec.spqr_spec.infra.resources")
		}
	}
	if newResources != nil || newInfraRouterSettings != nil || newInfraCoordinatorSettings != nil {
		config.SpqrSpec.Infra = &spqr.SpqrSpec_Infra{
			Resources:   newResources,
			Router:      newInfraRouterSettings,
			Coordinator: newInfraCoordinatorSettings,
		}
	}

	return config, updateMaskPaths, diags
}

func updateHosts(ctx context.Context,
	sdk *ycsdk.SDK,
	diagnostics *diag.Diagnostics,
	utilsHostService *SPQRHostService,
	hostsApiService *ShardedPostgreSQLAPI,
	cid string,
	plan, state types.Map,
	cfg *Config,
) {
	entityIdToPlanHost := make(map[string]Host)
	diagnostics.Append(plan.ElementsAs(ctx, &entityIdToPlanHost, false)...)
	if diagnostics.HasError() {
		return
	}
	entityIdToApiHosts := make(map[string]Host)
	diagnostics.Append(state.ElementsAs(ctx, &entityIdToApiHosts, false)...)
	if diagnostics.HasError() {
		return
	}
	entityIdToApiHosts = mdbcommon.ModifyStateDependsPlan(utilsHostService, entityIdToPlanHost, entityIdToApiHosts)

	toCreate, toUpdate, toDelete, diags := mdbcommon.HostsDiff(utilsHostService, entityIdToPlanHost, entityIdToApiHosts)
	if diagnostics.Append(diags...); diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "host operations will be processed", map[string]interface{}{
		"created": len(toCreate),
		"updated": len(toUpdate),
		"deleted": len(toDelete),
	})

	resources := make(map[spqr.Host_Type]*spqr.Resources)
	if cfg.SPQRConfig.Router != nil {
		resources[spqr.Host_ROUTER] = mdbcommon.ExpandResources[spqr.Resources](ctx, cfg.SPQRConfig.Router.Resources, diagnostics)
	}
	if cfg.SPQRConfig.Coordinator != nil {
		resources[spqr.Host_COORDINATOR] = mdbcommon.ExpandResources[spqr.Resources](ctx, cfg.SPQRConfig.Coordinator.Resources, diagnostics)
	}
	if cfg.SPQRConfig.Infra != nil {
		resources[spqr.Host_INFRA] = mdbcommon.ExpandResources[spqr.Resources](ctx, cfg.SPQRConfig.Infra.Resources, diagnostics)
	}

	hostsApiService.CreateHostsWithSubclusterCheck(ctx, sdk, diagnostics, cid, toCreate, resources)
	if diagnostics.HasError() {
		return
	}

	hostsApiService.UpdateHosts(ctx, sdk, diagnostics, cid, toUpdate)
	if diagnostics.HasError() {
		return
	}

	hostsApiService.DeleteHosts(ctx, sdk, diagnostics, cid, toDelete)
	if diagnostics.HasError() {
		return
	}
}

func appendSettingsToUpdateMask(ctx context.Context, diags *diag.Diagnostics, component string, state, plan *mdbcommon.SettingsMapValue) []string {
	attrsState := state.PrimitiveElements(ctx, diags)
	attrsPlan := plan.PrimitiveElements(ctx, diags)
	updateMaskPaths := []string{}

	maps.Copy(attrsPlan, attrsState)
	for attr := range attrsPlan {
		path := fmt.Sprintf("config_spec.spqr_spec.%s.config.%s", component, attr)
		if component == "" {
			path = fmt.Sprintf("config_spec.spqr_spec.%s", attr)
		}
		updateMaskPaths = append(updateMaskPaths, path)
	}
	return updateMaskPaths
}

func getAttrOrDefault(ctx context.Context, diags *diag.Diagnostics, attrs mdbcommon.SettingsMapValue, key string, def attr.Value) attr.Value {
	v, ok := attrs.PrimitiveElements(ctx, diags)[key]
	if !ok {
		return def
	}
	return v
}
