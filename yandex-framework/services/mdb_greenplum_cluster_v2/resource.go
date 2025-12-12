package mdb_greenplum_cluster_v2

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/greenplum/v1"
	greenplumv1sdk "github.com/yandex-cloud/go-sdk/services/mdb/greenplum/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/converter"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
	utils "github.com/yandex-cloud/terraform-provider-yandex/pkg/wrappers"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

const (
	ClusterDefaultTimeout = 120 * time.Minute
	ClusterUpdateTimeout  = 120 * time.Minute
	ClusterExpandTimeout  = 24 * 60 * time.Minute
	ClusterExpandDuration = 7200 // in seconds
)

var _ resource.ResourceWithConfigure = (*clusterResource)(nil)
var _ resource.ResourceWithImportState = (*clusterResource)(nil)

type clusterResource struct {
	providerConfig *providerconfig.Config
}

func NewResource() resource.Resource {
	return &clusterResource{}
}

func (r *clusterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mdb_greenplum_cluster_v2"
}

func (r *clusterResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerConfig, ok := req.ProviderData.(*providerconfig.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *provider_config.Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.providerConfig = providerConfig
}

func (r *clusterResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = YandexMdbGreenplumClusterV2ResourceSchema(ctx)
}

func (r *clusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *clusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Load the current state of the resource
	var state yandexMdbGreenplumClusterV2Model
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update Resource State
	r.refreshResourceState(ctx, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *clusterResource) refreshResourceState(ctx context.Context, state *yandexMdbGreenplumClusterV2Model, respDiagnostics *diag.Diagnostics) {
	readTimeout, timeoutInitError := state.Timeouts.Read(ctx, providerconfig.DefaultTimeout)
	if timeoutInitError != nil {
		respDiagnostics.Append(timeoutInitError...)
		return
	}
	ctx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()

	cid := state.ID.ValueString()
	cluster, err := r.providerConfig.SDK.MDB().Greenplum().Cluster().Get(ctx, &greenplum.GetClusterRequest{
		ClusterId: cid,
	})

	if err != nil {
		respDiagnostics.AddError(
			"Failed to read resource",
			fmt.Sprintf("Error while requesting API to read Greenplum cluster %q: %s", cid, err.Error()),
		)
		return
	}

	*state = flattenYandexMdbGreenplumClusterV2(ctx, cluster, *state, state.Timeouts, respDiagnostics)
	if respDiagnostics.HasError() {
		return
	}
}

func (r *clusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan yandexMdbGreenplumClusterV2Model
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, timeoutInitError := plan.Timeouts.Create(ctx, ClusterDefaultTimeout)
	if timeoutInitError != nil {
		resp.Diagnostics.Append(timeoutInitError...)
		return
	}

	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	createReq := &greenplum.CreateClusterRequest{}
	createReq.SetFolderId(converter.GetFolderID(plan.FolderId.ValueString(), r.providerConfig, &diags))
	createReq.SetName(plan.Name.ValueString())
	createReq.SetDescription(plan.Description.ValueString())
	createReq.SetLabels(expandYandexMdbGreenplumClusterV2Labels(ctx, plan.Labels, &diags))
	createReq.SetEnvironment(greenplum.Cluster_Environment(greenplum.Cluster_Environment_value[plan.Environment.ValueString()]))
	createReq.SetConfig(expandYandexMdbGreenplumClusterV2Config(ctx, plan.Config, &diags))
	createReq.SetMasterConfig(expandYandexMdbGreenplumClusterV2MasterConfig(ctx, plan.MasterConfig, &diags))
	createReq.SetSegmentConfig(expandYandexMdbGreenplumClusterV2SegmentConfig(ctx, plan.SegmentConfig, &diags))
	createReq.SetMasterHostCount(plan.MasterHostCount.ValueInt64())
	createReq.SetSegmentInHost(plan.SegmentInHost.ValueInt64())
	createReq.SetSegmentHostCount(plan.SegmentHostCount.ValueInt64())
	createReq.SetUserName(plan.UserName.ValueString())
	createReq.SetUserPassword(plan.UserPassword.ValueString())
	createReq.SetNetworkId(plan.NetworkId.ValueString())
	createReq.SetSecurityGroupIds(expandYandexMdbGreenplumClusterV2SecurityGroupIds(ctx, plan.SecurityGroupIds, &diags))
	createReq.SetDeletionProtection(plan.DeletionProtection.ValueBool())
	createReq.SetHostGroupIds(expandYandexMdbGreenplumClusterV2HostGroupIds(ctx, plan.HostGroupIds, &diags))
	createReq.SetMaintenanceWindow(expandYandexMdbGreenplumClusterV2MaintenanceWindow(ctx, plan.MaintenanceWindow, &diags))
	createReq.SetConfigSpec(expandYandexMdbGreenplumClusterV2ClusterConfig_create(ctx, plan.ClusterConfig, &diags))
	createReq.SetCloudStorage(expandYandexMdbGreenplumClusterV2CloudStorage(ctx, plan.CloudStorage, &diags))
	createReq.SetMasterHostGroupIds(expandYandexMdbGreenplumClusterV2HostGroupIds(ctx, plan.MasterHostGroupIds, &diags))
	createReq.SetSegmentHostGroupIds(expandYandexMdbGreenplumClusterV2HostGroupIds(ctx, plan.SegmentHostGroupIds, &diags))
	createReq.SetServiceAccountId(plan.ServiceAccountId.ValueString())
	createReq.SetLogging(expandYandexMdbGreenplumClusterV2Logging(ctx, plan.Logging, &diags))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("Create cluster request: %s", validate.ProtoDump(createReq)))

	if utils.IsPresent(plan.Restore) {
		r.restoreCluster(ctx, diags, &plan, resp)
	} else {
		r.createCluster(ctx, createReq, &plan, resp)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	r.refreshResourceState(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *clusterResource) createCluster(
	ctx context.Context,
	createReq *greenplum.CreateClusterRequest,
	plan *yandexMdbGreenplumClusterV2Model,
	resp *resource.CreateResponse,
) {
	md := new(metadata.MD)
	op, err := greenplumv1sdk.NewClusterClient(r.providerConfig.SDKv2).Create(ctx, createReq, grpc.Header(md))
	if traceHeader := md.Get("x-server-trace-id"); len(traceHeader) > 0 {
		tflog.Debug(ctx, fmt.Sprintf("[DEBUG] Create cluster x-server-trace-id: %s", traceHeader[0]))
	}
	if traceHeader := md.Get("x-server-request-id"); len(traceHeader) > 0 {
		tflog.Debug(ctx, fmt.Sprintf("Create cluster x-server-request-id: %s", traceHeader[0]))
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Create resource",
			"Error while requesting API to create cluster:"+err.Error(),
		)
		return
	}

	createRes, err := op.Wait(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			fmt.Sprintf("An unexpected error occurred while waiting longrunning response. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"Error: %s", err),
		)
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("Create cluster response: %s", validate.ProtoDump(createRes)))

	plan.ID = types.StringValue(op.Metadata().ClusterId)
}

func prepareRestoreRequest(
	ctx context.Context,
	plan *yandexMdbGreenplumClusterV2Model,
	providerConfig *providerconfig.State,
) (*greenplum.RestoreClusterRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	var restoreConf Restore

	diags.Append(plan.Restore.As(ctx, &restoreConf, datasize.DefaultOpts)...)
	if diags.HasError() {
		return nil, diags
	}

	var timeBackup *timestamppb.Timestamp = nil

	if utils.IsPresent(restoreConf.Time) {
		time, err := mdbcommon.ParseStringToTime(restoreConf.Time.ValueString())
		if err != nil {

			diags.Append(
				diag.NewErrorDiagnostic(
					"Failed to create Greenplum cluster from backup",
					fmt.Sprintf(
						"Error while parsing restore time to create Greenplum Cluster from backup %v, value: %v error: %s",
						restoreConf.BackupId,
						restoreConf.Time,
						err.Error(),
					),
				),
			)
		}
		timeBackup = &timestamppb.Timestamp{
			Seconds: time.Unix(),
		}
	}

	var masterConf yandexMdbGreenplumClusterV2MasterConfigModel
	var segmentConf yandexMdbGreenplumClusterV2SegmentConfigModel
	diags.Append(plan.MasterConfig.As(ctx, &masterConf, datasize.DefaultOpts)...)
	if diags.HasError() {
		return nil, diags
	}
	diags.Append(plan.SegmentConfig.As(ctx, &segmentConf, datasize.DefaultOpts)...)
	if diags.HasError() {
		return nil, diags
	}

	folderID, d := validate.FolderID(plan.FolderId, providerConfig)
	diags.Append(d)

	gpConfig := expandYandexMdbGreenplumClusterV2Config(ctx, plan.Config, &diags)

	request := &greenplum.RestoreClusterRequest{
		BackupId:    restoreConf.BackupId.ValueString(),
		Time:        timeBackup,
		FolderId:    folderID,
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Labels:      mdbcommon.ExpandLabels(ctx, plan.Labels, &diags),
		Environment: mdbcommon.ExpandEnvironment[greenplum.Cluster_Environment](ctx, plan.Environment, &diags),
		Config: &greenplum.GreenplumRestoreConfig{
			BackupWindowStart: gpConfig.BackupWindowStart,
			Access:            gpConfig.Access,
			ZoneId:            gpConfig.ZoneId,
			SubnetId:          gpConfig.SubnetId,
			AssignPublicIp:    gpConfig.AssignPublicIp,
		},
		MasterResources:     mdbcommon.ExpandResources[greenplum.Resources](ctx, masterConf.Resources, &diags),
		SegmentResources:    mdbcommon.ExpandResources[greenplum.Resources](ctx, segmentConf.Resources, &diags),
		NetworkId:           plan.NetworkId.ValueString(),
		SecurityGroupIds:    mdbcommon.ExpandSecurityGroupIds(ctx, plan.SecurityGroupIds, &diags),
		DeletionProtection:  plan.DeletionProtection.ValueBool(),
		HostGroupIds:        expandYandexMdbGreenplumClusterV2HostGroupIds(ctx, plan.HostGroupIds, &diags),
		MaintenanceWindow:   expandYandexMdbGreenplumClusterV2MaintenanceWindow(ctx, plan.MaintenanceWindow, &diags),
		SegmentHostCount:    plan.SegmentHostCount.ValueInt64(),
		SegmentInHost:       plan.SegmentInHost.ValueInt64(),
		RestoreOnly:         expandYandexMdbGreenplumClusterV2RestoreOnly(ctx, restoreConf.RestoreOnly, &diags),
		MasterHostGroupIds:  expandYandexMdbGreenplumClusterV2HostGroupIds(ctx, plan.MasterHostGroupIds, &diags),
		SegmentHostGroupIds: expandYandexMdbGreenplumClusterV2HostGroupIds(ctx, plan.SegmentHostGroupIds, &diags),
		ServiceAccountId:    plan.GetServiceAccountId().ValueString(),
	}

	return request, diags
}

func (r *clusterResource) restoreCluster(
	ctx context.Context,
	diags diag.Diagnostics,
	plan *yandexMdbGreenplumClusterV2Model,
	resp *resource.CreateResponse,
) {
	restoreReq, diags := prepareRestoreRequest(ctx, plan, &r.providerConfig.ProviderState)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	md := new(metadata.MD)
	op, err := greenplumv1sdk.NewClusterClient(r.providerConfig.SDKv2).Restore(ctx, restoreReq, grpc.Header(md))
	if traceHeader := md.Get("x-server-trace-id"); len(traceHeader) > 0 {
		tflog.Debug(ctx, fmt.Sprintf("[DEBUG] Restore cluster x-server-trace-id: %s", traceHeader[0]))
	}
	if traceHeader := md.Get("x-server-request-id"); len(traceHeader) > 0 {
		tflog.Debug(ctx, fmt.Sprintf("Restore cluster x-server-request-id: %s", traceHeader[0]))
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Create resource",
			"Error while requesting API to create cluster:"+err.Error(),
		)
		return
	}

	restoreRes, err := op.Wait(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			fmt.Sprintf("An unexpected error occurred while waiting longrunning response. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"Error: %s", err),
		)
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("Restore cluster response: %s", validate.ProtoDump(restoreRes)))

	plan.ID = types.StringValue(op.Metadata().ClusterId)
}

func (r *clusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state yandexMdbGreenplumClusterV2Model
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, timeoutInitError := state.Timeouts.Delete(ctx, ClusterDefaultTimeout)
	if timeoutInitError != nil {
		resp.Diagnostics.Append(timeoutInitError...)
		return
	}

	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	reqApi := &greenplum.DeleteClusterRequest{}
	id := state.ID.ValueString()
	if !state.ID.IsUnknown() && !state.ID.IsNull() {
		id = state.ID.ValueString()
	}
	reqApi.SetClusterId(id)
	tflog.Debug(ctx, fmt.Sprintf("Delete cluster request: %s", validate.ProtoDump(reqApi)))

	md := new(metadata.MD)

	op, err := greenplumv1sdk.NewClusterClient(r.providerConfig.SDKv2).Delete(ctx, reqApi, grpc.Header(md))
	if traceHeader := md.Get("x-server-trace-id"); len(traceHeader) > 0 {
		tflog.Debug(ctx, fmt.Sprintf("Delete cluster x-server-trace-id: %s", traceHeader[0]))
	}
	if traceHeader := md.Get("x-server-request-id"); len(traceHeader) > 0 {
		tflog.Debug(ctx, fmt.Sprintf("Delete cluster x-server-request-id: %s", traceHeader[0]))
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Delete resource",
			"Error while requesting API to delete cluster:"+err.Error(),
		)
		return
	}
	deleteRes, err := op.Wait(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Delete Resource",
			fmt.Sprintf("An unexpected error occurred while waiting longrunning response. "+
				"Please retry the operation or report this issue to the provider developers.\n\n"+
				"Error: %s", err),
		)
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("[DEBUG] Delete cluster response: %s", validate.ProtoDump(deleteRes)))
}

func (r *clusterResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() {
		return
	}
	var plan yandexMdbGreenplumClusterV2Model
	var state yandexMdbGreenplumClusterV2Model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Modifying plan for cluster", map[string]interface{}{"id": plan.ID.ValueString()})

	// remove changes in restore section
	plan.Restore = state.Restore

	resp.Diagnostics.Append(resp.Plan.Set(ctx, plan)...)
}

func (r *clusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state yandexMdbGreenplumClusterV2Model
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateTimeout, timeoutInitError := plan.Timeouts.Update(ctx, ClusterUpdateTimeout)
	if timeoutInitError != nil {
		resp.Diagnostics.Append(timeoutInitError...)
		return
	}

	var updatePaths []string

	var yandexMdbGreenplumClusterV2CloudStorageState, yandexMdbGreenplumClusterV2CloudStoragePlan yandexMdbGreenplumClusterV2CloudStorageModel
	resp.Diagnostics.Append(plan.CloudStorage.As(ctx, &yandexMdbGreenplumClusterV2CloudStoragePlan, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	resp.Diagnostics.Append(state.CloudStorage.As(ctx, &yandexMdbGreenplumClusterV2CloudStorageState, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !yandexMdbGreenplumClusterV2CloudStoragePlan.Enable.Equal(yandexMdbGreenplumClusterV2CloudStorageState.Enable) {
		updatePaths = append(updatePaths, "cloud_storage.enable")
	}

	var yandexMdbGreenplumClusterV2ClusterConfigState, yandexMdbGreenplumClusterV2ClusterConfigPlan yandexMdbGreenplumClusterV2ClusterConfigModel
	resp.Diagnostics.Append(plan.ClusterConfig.As(ctx, &yandexMdbGreenplumClusterV2ClusterConfigPlan, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	resp.Diagnostics.Append(state.ClusterConfig.As(ctx, &yandexMdbGreenplumClusterV2ClusterConfigState, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if resp.Diagnostics.HasError() {
		return
	}

	var yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesState, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesPlan yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesModel
	resp.Diagnostics.Append(yandexMdbGreenplumClusterV2ClusterConfigPlan.BackgroundActivities.As(ctx, &yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesPlan, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	resp.Diagnostics.Append(yandexMdbGreenplumClusterV2ClusterConfigState.BackgroundActivities.As(ctx, &yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesState, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if resp.Diagnostics.HasError() {
		return
	}

	var yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumState, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumPlan yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumModel
	resp.Diagnostics.Append(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesPlan.AnalyzeAndVacuum.As(ctx, &yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumPlan, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	resp.Diagnostics.Append(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesState.AnalyzeAndVacuum.As(ctx, &yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumState, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumPlan.AnalyzeTimeout.Equal(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumState.AnalyzeTimeout) {
		updatePaths = append(updatePaths, "config_spec.background_activities.analyze_and_vacuum.analyze_timeout")
	}

	var yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumStartState, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumStartPlan yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumStartModel
	resp.Diagnostics.Append(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumPlan.Start.As(ctx, &yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumStartPlan, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	resp.Diagnostics.Append(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumState.Start.As(ctx, &yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumStartState, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumStartPlan.Hours.Equal(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumStartState.Hours) {
		updatePaths = append(updatePaths, "config_spec.background_activities.analyze_and_vacuum.start.hours")
	}
	if !yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumStartPlan.Minutes.Equal(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumStartState.Minutes) {
		updatePaths = append(updatePaths, "config_spec.background_activities.analyze_and_vacuum.start.minutes")
	}
	if !yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumPlan.VacuumTimeout.Equal(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesAnalyzeAndVacuumState.VacuumTimeout) {
		updatePaths = append(updatePaths, "config_spec.background_activities.analyze_and_vacuum.vacuum_timeout")
	}

	var yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsState, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsPlan yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsModel
	resp.Diagnostics.Append(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesPlan.QueryKillerScripts.As(ctx, &yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsPlan, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	resp.Diagnostics.Append(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesState.QueryKillerScripts.As(ctx, &yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsState, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if resp.Diagnostics.HasError() {
		return
	}

	var yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleState, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdlePlan yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleModel
	resp.Diagnostics.Append(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsPlan.Idle.As(ctx, &yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdlePlan, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	resp.Diagnostics.Append(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsState.Idle.As(ctx, &yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleState, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdlePlan.Enable.Equal(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleState.Enable) {
		updatePaths = append(updatePaths, "config_spec.background_activities.query_killer_scripts.idle.enable")
	}
	if yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdlePlan.IgnoreUsers.IsNull() {
		yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdlePlan.IgnoreUsers = types.SetNull(types.StringType)
	}
	if yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleState.IgnoreUsers.IsNull() {
		yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleState.IgnoreUsers = types.SetNull(types.StringType)
	}
	if !yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdlePlan.IgnoreUsers.Equal(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleState.IgnoreUsers) {
		updatePaths = append(updatePaths, "config_spec.background_activities.query_killer_scripts.idle.ignore_users")
	}
	if !yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdlePlan.MaxAge.Equal(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleState.MaxAge) {
		updatePaths = append(updatePaths, "config_spec.background_activities.query_killer_scripts.idle.max_age")
	}

	var yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionState, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionPlan yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionModel
	resp.Diagnostics.Append(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsPlan.IdleInTransaction.As(ctx, &yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionPlan, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	resp.Diagnostics.Append(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsState.IdleInTransaction.As(ctx, &yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionState, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionPlan.Enable.Equal(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionState.Enable) {
		updatePaths = append(updatePaths, "config_spec.background_activities.query_killer_scripts.idle_in_transaction.enable")
	}
	if yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionPlan.IgnoreUsers.IsNull() {
		yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionPlan.IgnoreUsers = types.SetNull(types.StringType)
	}
	if yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionState.IgnoreUsers.IsNull() {
		yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionState.IgnoreUsers = types.SetNull(types.StringType)
	}
	if !yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionPlan.IgnoreUsers.Equal(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionState.IgnoreUsers) {
		updatePaths = append(updatePaths, "config_spec.background_activities.query_killer_scripts.idle_in_transaction.ignore_users")
	}
	if !yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionPlan.MaxAge.Equal(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsIdleInTransactionState.MaxAge) {
		updatePaths = append(updatePaths, "config_spec.background_activities.query_killer_scripts.idle_in_transaction.max_age")
	}

	var yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningState, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningPlan yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningModel
	resp.Diagnostics.Append(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsPlan.LongRunning.As(ctx, &yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningPlan, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	resp.Diagnostics.Append(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsState.LongRunning.As(ctx, &yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningState, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningPlan.Enable.Equal(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningState.Enable) {
		updatePaths = append(updatePaths, "config_spec.background_activities.query_killer_scripts.long_running.enable")
	}
	if yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningPlan.IgnoreUsers.IsNull() {
		yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningPlan.IgnoreUsers = types.SetNull(types.StringType)
	}
	if yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningState.IgnoreUsers.IsNull() {
		yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningState.IgnoreUsers = types.SetNull(types.StringType)
	}
	if !yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningPlan.IgnoreUsers.Equal(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningState.IgnoreUsers) {
		updatePaths = append(updatePaths, "config_spec.background_activities.query_killer_scripts.long_running.ignore_users")
	}
	if !yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningPlan.MaxAge.Equal(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesQueryKillerScriptsLongRunningState.MaxAge) {
		updatePaths = append(updatePaths, "config_spec.background_activities.query_killer_scripts.long_running.max_age")
	}

	var yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesState, yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesPlan yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesModel
	resp.Diagnostics.Append(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesPlan.TableSizes.As(ctx, &yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesPlan, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	resp.Diagnostics.Append(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesState.TableSizes.As(ctx, &yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesState, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if resp.Diagnostics.HasError() {
		return
	}

	if yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesPlan.Starts.IsNull() {
		yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesPlan.Starts = types.SetNull(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesBackgroundActivityStartAtStructModelType)
	}
	if yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesState.Starts.IsNull() {
		yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesState.Starts = types.SetNull(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesBackgroundActivityStartAtStructModelType)
	}
	if !yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesPlan.Starts.Equal(yandexMdbGreenplumClusterV2ClusterConfigBackgroundActivitiesTableSizesState.Starts) {
		updatePaths = append(updatePaths, "config_spec.background_activities.table_sizes.starts")
	}

	var yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State, yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Plan yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Model
	resp.Diagnostics.Append(yandexMdbGreenplumClusterV2ClusterConfigPlan.GreenplumConfig6.As(ctx, &yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Plan, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	resp.Diagnostics.Append(yandexMdbGreenplumClusterV2ClusterConfigState.GreenplumConfig6.As(ctx, &yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Plan.GpAddColumnInheritsTableSetting.Equal(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State.GpAddColumnInheritsTableSetting) {
		updatePaths = append(updatePaths, "config_spec.greenplum_config_6.gp_add_column_inherits_table_setting")
	}
	if !yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Plan.GpAutostatsMode.Equal(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State.GpAutostatsMode) {
		updatePaths = append(updatePaths, "config_spec.greenplum_config_6.gp_autostats_mode")
	}
	if !yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Plan.GpAutostatsOnChangeThreshold.Equal(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State.GpAutostatsOnChangeThreshold) {
		updatePaths = append(updatePaths, "config_spec.greenplum_config_6.gp_autostats_on_change_threshold")
	}
	if !yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Plan.GpCachedSegworkersThreshold.Equal(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State.GpCachedSegworkersThreshold) {
		updatePaths = append(updatePaths, "config_spec.greenplum_config_6.gp_cached_segworkers_threshold")
	}
	if !yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Plan.GpEnableGlobalDeadlockDetector.Equal(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State.GpEnableGlobalDeadlockDetector) {
		updatePaths = append(updatePaths, "config_spec.greenplum_config_6.gp_enable_global_deadlock_detector")
	}
	if !yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Plan.GpEnableZstdMemoryAccounting.Equal(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State.GpEnableZstdMemoryAccounting) {
		updatePaths = append(updatePaths, "config_spec.greenplum_config_6.gp_enable_zstd_memory_accounting")
	}
	if !yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Plan.GpGlobalDeadlockDetectorPeriod.Equal(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State.GpGlobalDeadlockDetectorPeriod) {
		updatePaths = append(updatePaths, "config_spec.greenplum_config_6.gp_global_deadlock_detector_period")
	}
	if !yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Plan.GpMaxPlanSize.Equal(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State.GpMaxPlanSize) {
		updatePaths = append(updatePaths, "config_spec.greenplum_config_6.gp_max_plan_size")
	}
	if !yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Plan.GpMaxSlices.Equal(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State.GpMaxSlices) {
		updatePaths = append(updatePaths, "config_spec.greenplum_config_6.gp_max_slices")
	}
	if !yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Plan.GpResourceGroupMemoryLimit.Equal(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State.GpResourceGroupMemoryLimit) {
		updatePaths = append(updatePaths, "config_spec.greenplum_config_6.gp_resource_group_memory_limit")
	}
	if !yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Plan.GpVmemProtectSegworkerCacheLimit.Equal(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State.GpVmemProtectSegworkerCacheLimit) {
		updatePaths = append(updatePaths, "config_spec.greenplum_config_6.gp_vmem_protect_segworker_cache_limit")
	}
	if !yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Plan.GpWorkfileCompression.Equal(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State.GpWorkfileCompression) {
		updatePaths = append(updatePaths, "config_spec.greenplum_config_6.gp_workfile_compression")
	}
	if !yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Plan.GpWorkfileLimitFilesPerQuery.Equal(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State.GpWorkfileLimitFilesPerQuery) {
		updatePaths = append(updatePaths, "config_spec.greenplum_config_6.gp_workfile_limit_files_per_query")
	}
	if !yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Plan.GpWorkfileLimitPerQuery.Equal(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State.GpWorkfileLimitPerQuery) {
		updatePaths = append(updatePaths, "config_spec.greenplum_config_6.gp_workfile_limit_per_query")
	}
	if !yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Plan.GpWorkfileLimitPerSegment.Equal(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State.GpWorkfileLimitPerSegment) {
		updatePaths = append(updatePaths, "config_spec.greenplum_config_6.gp_workfile_limit_per_segment")
	}
	if !yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Plan.IdleInTransactionSessionTimeout.Equal(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State.IdleInTransactionSessionTimeout) {
		updatePaths = append(updatePaths, "config_spec.greenplum_config_6.idle_in_transaction_session_timeout")
	}
	if !yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Plan.LockTimeout.Equal(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State.LockTimeout) {
		updatePaths = append(updatePaths, "config_spec.greenplum_config_6.lock_timeout")
	}
	if !yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Plan.LogStatement.Equal(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State.LogStatement) {
		updatePaths = append(updatePaths, "config_spec.greenplum_config_6.log_statement")
	}
	if !yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Plan.MaxConnections.Equal(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State.MaxConnections) {
		updatePaths = append(updatePaths, "config_spec.greenplum_config_6.max_connections")
	}
	if !yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Plan.MaxPreparedTransactions.Equal(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State.MaxPreparedTransactions) {
		updatePaths = append(updatePaths, "config_spec.greenplum_config_6.max_prepared_transactions")
	}
	if !yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Plan.MaxSlotWalKeepSize.Equal(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State.MaxSlotWalKeepSize) {
		updatePaths = append(updatePaths, "config_spec.greenplum_config_6.max_slot_wal_keep_size")
	}
	if !yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Plan.MaxStatementMem.Equal(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State.MaxStatementMem) {
		updatePaths = append(updatePaths, "config_spec.greenplum_config_6.max_statement_mem")
	}
	if !yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6Plan.RunawayDetectorActivationPercent.Equal(yandexMdbGreenplumClusterV2ClusterConfigGreenplumConfig6State.RunawayDetectorActivationPercent) {
		updatePaths = append(updatePaths, "config_spec.greenplum_config_6.runaway_detector_activation_percent")
	}

	var yandexMdbGreenplumClusterV2ClusterConfigPoolState, yandexMdbGreenplumClusterV2ClusterConfigPoolPlan yandexMdbGreenplumClusterV2ClusterConfigPoolModel
	resp.Diagnostics.Append(yandexMdbGreenplumClusterV2ClusterConfigPlan.Pool.As(ctx, &yandexMdbGreenplumClusterV2ClusterConfigPoolPlan, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	resp.Diagnostics.Append(yandexMdbGreenplumClusterV2ClusterConfigState.Pool.As(ctx, &yandexMdbGreenplumClusterV2ClusterConfigPoolState, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !yandexMdbGreenplumClusterV2ClusterConfigPoolPlan.ClientIdleTimeout.Equal(yandexMdbGreenplumClusterV2ClusterConfigPoolState.ClientIdleTimeout) {
		updatePaths = append(updatePaths, "config_spec.pool.client_idle_timeout")
	}
	if !yandexMdbGreenplumClusterV2ClusterConfigPoolPlan.IdleInTransactionTimeout.Equal(yandexMdbGreenplumClusterV2ClusterConfigPoolState.IdleInTransactionTimeout) {
		updatePaths = append(updatePaths, "config_spec.pool.idle_in_transaction_timeout")
	}
	if !yandexMdbGreenplumClusterV2ClusterConfigPoolPlan.Mode.Equal(yandexMdbGreenplumClusterV2ClusterConfigPoolState.Mode) {
		updatePaths = append(updatePaths, "config_spec.pool.mode")
	}
	if !yandexMdbGreenplumClusterV2ClusterConfigPoolPlan.Size.Equal(yandexMdbGreenplumClusterV2ClusterConfigPoolState.Size) {
		updatePaths = append(updatePaths, "config_spec.pool.size")
	}

	var yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState, yandexMdbGreenplumClusterV2ClusterConfigPxfConfigPlan yandexMdbGreenplumClusterV2ClusterConfigPxfConfigModel
	resp.Diagnostics.Append(yandexMdbGreenplumClusterV2ClusterConfigPlan.PxfConfig.As(ctx, &yandexMdbGreenplumClusterV2ClusterConfigPxfConfigPlan, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	resp.Diagnostics.Append(yandexMdbGreenplumClusterV2ClusterConfigState.PxfConfig.As(ctx, &yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !yandexMdbGreenplumClusterV2ClusterConfigPxfConfigPlan.ConnectionTimeout.Equal(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState.ConnectionTimeout) {
		updatePaths = append(updatePaths, "config_spec.pxf_config.connection_timeout")
	}
	if !yandexMdbGreenplumClusterV2ClusterConfigPxfConfigPlan.MaxThreads.Equal(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState.MaxThreads) {
		updatePaths = append(updatePaths, "config_spec.pxf_config.max_threads")
	}
	if !yandexMdbGreenplumClusterV2ClusterConfigPxfConfigPlan.PoolAllowCoreThreadTimeout.Equal(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState.PoolAllowCoreThreadTimeout) {
		updatePaths = append(updatePaths, "config_spec.pxf_config.pool_allow_core_thread_timeout")
	}
	if !yandexMdbGreenplumClusterV2ClusterConfigPxfConfigPlan.PoolCoreSize.Equal(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState.PoolCoreSize) {
		updatePaths = append(updatePaths, "config_spec.pxf_config.pool_core_size")
	}
	if !yandexMdbGreenplumClusterV2ClusterConfigPxfConfigPlan.PoolMaxSize.Equal(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState.PoolMaxSize) {
		updatePaths = append(updatePaths, "config_spec.pxf_config.pool_max_size")
	}
	if !yandexMdbGreenplumClusterV2ClusterConfigPxfConfigPlan.PoolQueueCapacity.Equal(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState.PoolQueueCapacity) {
		updatePaths = append(updatePaths, "config_spec.pxf_config.pool_queue_capacity")
	}
	if !yandexMdbGreenplumClusterV2ClusterConfigPxfConfigPlan.UploadTimeout.Equal(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState.UploadTimeout) {
		updatePaths = append(updatePaths, "config_spec.pxf_config.upload_timeout")
	}
	if !yandexMdbGreenplumClusterV2ClusterConfigPxfConfigPlan.Xms.Equal(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState.Xms) {
		updatePaths = append(updatePaths, "config_spec.pxf_config.xms")
	}
	if !yandexMdbGreenplumClusterV2ClusterConfigPxfConfigPlan.Xmx.Equal(yandexMdbGreenplumClusterV2ClusterConfigPxfConfigState.Xmx) {
		updatePaths = append(updatePaths, "config_spec.pxf_config.xmx")
	}

	var yandexMdbGreenplumClusterV2ConfigState, yandexMdbGreenplumClusterV2ConfigPlan yandexMdbGreenplumClusterV2ConfigModel
	resp.Diagnostics.Append(plan.Config.As(ctx, &yandexMdbGreenplumClusterV2ConfigPlan, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	resp.Diagnostics.Append(state.Config.As(ctx, &yandexMdbGreenplumClusterV2ConfigState, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if resp.Diagnostics.HasError() {
		return
	}

	var yandexMdbGreenplumClusterV2ConfigAccessState, yandexMdbGreenplumClusterV2ConfigAccessPlan yandexMdbGreenplumClusterV2ConfigAccessModel
	resp.Diagnostics.Append(yandexMdbGreenplumClusterV2ConfigPlan.Access.As(ctx, &yandexMdbGreenplumClusterV2ConfigAccessPlan, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	resp.Diagnostics.Append(yandexMdbGreenplumClusterV2ConfigState.Access.As(ctx, &yandexMdbGreenplumClusterV2ConfigAccessState, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !yandexMdbGreenplumClusterV2ConfigAccessPlan.DataLens.Equal(yandexMdbGreenplumClusterV2ConfigAccessState.DataLens) {
		updatePaths = append(updatePaths, "config.access.data_lens")
	}
	if !yandexMdbGreenplumClusterV2ConfigAccessPlan.DataTransfer.Equal(yandexMdbGreenplumClusterV2ConfigAccessState.DataTransfer) {
		updatePaths = append(updatePaths, "config.access.data_transfer")
	}
	if !yandexMdbGreenplumClusterV2ConfigAccessPlan.WebSql.Equal(yandexMdbGreenplumClusterV2ConfigAccessState.WebSql) {
		updatePaths = append(updatePaths, "config.access.web_sql")
	}
	if !yandexMdbGreenplumClusterV2ConfigAccessPlan.YandexQuery.Equal(yandexMdbGreenplumClusterV2ConfigAccessState.YandexQuery) {
		updatePaths = append(updatePaths, "config.access.yandex_query")
	}
	if !yandexMdbGreenplumClusterV2ConfigPlan.AssignPublicIp.Equal(yandexMdbGreenplumClusterV2ConfigState.AssignPublicIp) {
		updatePaths = append(updatePaths, "config.assign_public_ip")
	}
	if !yandexMdbGreenplumClusterV2ConfigPlan.BackupRetainPeriodDays.Equal(yandexMdbGreenplumClusterV2ConfigState.BackupRetainPeriodDays) {
		updatePaths = append(updatePaths, "config.backup_retain_period_days")
	}
	if !yandexMdbGreenplumClusterV2ConfigPlan.BackupWindowStart.Equal(yandexMdbGreenplumClusterV2ConfigState.BackupWindowStart) {
		updatePaths = append(updatePaths, "config.backup_window_start")
	}
	if !yandexMdbGreenplumClusterV2ConfigPlan.SubnetId.Equal(yandexMdbGreenplumClusterV2ConfigState.SubnetId) {
		updatePaths = append(updatePaths, "config.subnet_id")
	}
	if !yandexMdbGreenplumClusterV2ConfigPlan.Version.Equal(yandexMdbGreenplumClusterV2ConfigState.Version) {
		updatePaths = append(updatePaths, "config.version")
	}
	if !yandexMdbGreenplumClusterV2ConfigPlan.ZoneId.Equal(yandexMdbGreenplumClusterV2ConfigState.ZoneId) {
		updatePaths = append(updatePaths, "config.zone_id")
	}
	if !plan.DeletionProtection.Equal(state.DeletionProtection) {
		updatePaths = append(updatePaths, "deletion_protection")
	}
	if !plan.Description.Equal(state.Description) {
		updatePaths = append(updatePaths, "description")
	}
	if plan.Labels.IsNull() {
		plan.Labels = types.MapNull(types.StringType)
	}
	if state.Labels.IsNull() {
		state.Labels = types.MapNull(types.StringType)
	}
	if !plan.Labels.Equal(state.Labels) {
		updatePaths = append(updatePaths, "labels")
	}

	var yandexMdbGreenplumClusterV2LoggingState, yandexMdbGreenplumClusterV2LoggingPlan yandexMdbGreenplumClusterV2LoggingModel
	resp.Diagnostics.Append(plan.Logging.As(ctx, &yandexMdbGreenplumClusterV2LoggingPlan, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	resp.Diagnostics.Append(state.Logging.As(ctx, &yandexMdbGreenplumClusterV2LoggingState, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !yandexMdbGreenplumClusterV2LoggingPlan.CommandCenterEnabled.Equal(yandexMdbGreenplumClusterV2LoggingState.CommandCenterEnabled) {
		updatePaths = append(updatePaths, "logging.command_center_enabled")
	}
	if !yandexMdbGreenplumClusterV2LoggingPlan.Enabled.Equal(yandexMdbGreenplumClusterV2LoggingState.Enabled) {
		updatePaths = append(updatePaths, "logging.enabled")
	}
	if !yandexMdbGreenplumClusterV2LoggingPlan.FolderId.Equal(yandexMdbGreenplumClusterV2LoggingState.FolderId) {
		updatePaths = append(updatePaths, "logging.folder_id")
	}
	if !yandexMdbGreenplumClusterV2LoggingPlan.GreenplumEnabled.Equal(yandexMdbGreenplumClusterV2LoggingState.GreenplumEnabled) {
		updatePaths = append(updatePaths, "logging.greenplum_enabled")
	}
	if !yandexMdbGreenplumClusterV2LoggingPlan.LogGroupId.Equal(yandexMdbGreenplumClusterV2LoggingState.LogGroupId) {
		updatePaths = append(updatePaths, "logging.log_group_id")
	}
	if !yandexMdbGreenplumClusterV2LoggingPlan.PoolerEnabled.Equal(yandexMdbGreenplumClusterV2LoggingState.PoolerEnabled) {
		updatePaths = append(updatePaths, "logging.pooler_enabled")
	}

	var yandexMdbGreenplumClusterV2MaintenanceWindowState, yandexMdbGreenplumClusterV2MaintenanceWindowPlan yandexMdbGreenplumClusterV2MaintenanceWindowModel
	resp.Diagnostics.Append(plan.MaintenanceWindow.As(ctx, &yandexMdbGreenplumClusterV2MaintenanceWindowPlan, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	resp.Diagnostics.Append(state.MaintenanceWindow.As(ctx, &yandexMdbGreenplumClusterV2MaintenanceWindowState, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if resp.Diagnostics.HasError() {
		return
	}

	var yandexMdbGreenplumClusterV2MaintenanceWindowWeeklyMaintenanceWindowState, yandexMdbGreenplumClusterV2MaintenanceWindowWeeklyMaintenanceWindowPlan yandexMdbGreenplumClusterV2MaintenanceWindowWeeklyMaintenanceWindowModel
	resp.Diagnostics.Append(yandexMdbGreenplumClusterV2MaintenanceWindowPlan.WeeklyMaintenanceWindow.As(ctx, &yandexMdbGreenplumClusterV2MaintenanceWindowWeeklyMaintenanceWindowPlan, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	resp.Diagnostics.Append(yandexMdbGreenplumClusterV2MaintenanceWindowState.WeeklyMaintenanceWindow.As(ctx, &yandexMdbGreenplumClusterV2MaintenanceWindowWeeklyMaintenanceWindowState, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !yandexMdbGreenplumClusterV2MaintenanceWindowWeeklyMaintenanceWindowPlan.Day.Equal(yandexMdbGreenplumClusterV2MaintenanceWindowWeeklyMaintenanceWindowState.Day) {
		updatePaths = append(updatePaths, "maintenance_window.weekly_maintenance_window.day")
	}
	if !yandexMdbGreenplumClusterV2MaintenanceWindowWeeklyMaintenanceWindowPlan.Hour.Equal(yandexMdbGreenplumClusterV2MaintenanceWindowWeeklyMaintenanceWindowState.Hour) {
		updatePaths = append(updatePaths, "maintenance_window.weekly_maintenance_window.hour")
	}

	var yandexMdbGreenplumClusterV2MasterConfigState, yandexMdbGreenplumClusterV2MasterConfigPlan yandexMdbGreenplumClusterV2MasterConfigModel
	resp.Diagnostics.Append(plan.MasterConfig.As(ctx, &yandexMdbGreenplumClusterV2MasterConfigPlan, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	resp.Diagnostics.Append(state.MasterConfig.As(ctx, &yandexMdbGreenplumClusterV2MasterConfigState, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if resp.Diagnostics.HasError() {
		return
	}

	var yandexMdbGreenplumClusterV2MasterConfigResourcesState, yandexMdbGreenplumClusterV2MasterConfigResourcesPlan yandexMdbGreenplumClusterV2MasterConfigResourcesModel
	resp.Diagnostics.Append(yandexMdbGreenplumClusterV2MasterConfigPlan.Resources.As(ctx, &yandexMdbGreenplumClusterV2MasterConfigResourcesPlan, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	resp.Diagnostics.Append(yandexMdbGreenplumClusterV2MasterConfigState.Resources.As(ctx, &yandexMdbGreenplumClusterV2MasterConfigResourcesState, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !yandexMdbGreenplumClusterV2MasterConfigResourcesPlan.DiskSize.Equal(yandexMdbGreenplumClusterV2MasterConfigResourcesState.DiskSize) {
		updatePaths = append(updatePaths, "master_config.resources.disk_size")
	}
	if !yandexMdbGreenplumClusterV2MasterConfigResourcesPlan.DiskTypeId.Equal(yandexMdbGreenplumClusterV2MasterConfigResourcesState.DiskTypeId) {
		updatePaths = append(updatePaths, "master_config.resources.disk_type_id")
	}
	if !yandexMdbGreenplumClusterV2MasterConfigResourcesPlan.ResourcePresetId.Equal(yandexMdbGreenplumClusterV2MasterConfigResourcesState.ResourcePresetId) {
		updatePaths = append(updatePaths, "master_config.resources.resource_preset_id")
	}
	if !plan.Name.Equal(state.Name) {
		updatePaths = append(updatePaths, "name")
	}

	var yandexMdbGreenplumClusterV2SegmentConfigState, yandexMdbGreenplumClusterV2SegmentConfigPlan yandexMdbGreenplumClusterV2SegmentConfigModel
	resp.Diagnostics.Append(plan.SegmentConfig.As(ctx, &yandexMdbGreenplumClusterV2SegmentConfigPlan, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	resp.Diagnostics.Append(state.SegmentConfig.As(ctx, &yandexMdbGreenplumClusterV2SegmentConfigState, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if resp.Diagnostics.HasError() {
		return
	}

	var yandexMdbGreenplumClusterV2SegmentConfigResourcesState, yandexMdbGreenplumClusterV2SegmentConfigResourcesPlan yandexMdbGreenplumClusterV2SegmentConfigResourcesModel
	resp.Diagnostics.Append(yandexMdbGreenplumClusterV2SegmentConfigPlan.Resources.As(ctx, &yandexMdbGreenplumClusterV2SegmentConfigResourcesPlan, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	resp.Diagnostics.Append(yandexMdbGreenplumClusterV2SegmentConfigState.Resources.As(ctx, &yandexMdbGreenplumClusterV2SegmentConfigResourcesState, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !yandexMdbGreenplumClusterV2SegmentConfigResourcesPlan.DiskSize.Equal(yandexMdbGreenplumClusterV2SegmentConfigResourcesState.DiskSize) {
		updatePaths = append(updatePaths, "segment_config.resources.disk_size")
	}
	if !yandexMdbGreenplumClusterV2SegmentConfigResourcesPlan.DiskTypeId.Equal(yandexMdbGreenplumClusterV2SegmentConfigResourcesState.DiskTypeId) {
		updatePaths = append(updatePaths, "segment_config.resources.disk_type_id")
	}
	if !yandexMdbGreenplumClusterV2SegmentConfigResourcesPlan.ResourcePresetId.Equal(yandexMdbGreenplumClusterV2SegmentConfigResourcesState.ResourcePresetId) {
		updatePaths = append(updatePaths, "segment_config.resources.resource_preset_id")
	}
	if !plan.ServiceAccountId.Equal(state.ServiceAccountId) {
		updatePaths = append(updatePaths, "service_account_id")
	}
	if !plan.UserPassword.Equal(state.UserPassword) {
		updatePaths = append(updatePaths, "user_password")
	}

	// Prepare expand req
	var segHostCount int64
	var addSegCount int64

	if !plan.SegmentHostCount.Equal(state.SegmentHostCount) {
		segHostCount = plan.SegmentHostCount.ValueInt64() - state.SegmentHostCount.ValueInt64()
	}

	if !plan.SegmentInHost.Equal(state.SegmentInHost) {
		addSegCount = plan.SegmentInHost.ValueInt64() - state.SegmentInHost.ValueInt64()
	}

	id := plan.ID.ValueString()
	expandReq := &greenplum.ExpandRequest{
		ClusterId:               id,
		SegmentHostCount:        segHostCount,
		AddSegmentsPerHostCount: addSegCount,
		Duration:                ClusterExpandDuration,
	}

	if len(updatePaths) != 0 {
		ctx, cancel := context.WithTimeout(ctx, updateTimeout)
		defer cancel()

		updateReq := &greenplum.UpdateClusterRequest{}
		updateReq.SetClusterId(id)
		updateReq.SetDescription(plan.Description.ValueString())
		updateReq.SetLabels(expandYandexMdbGreenplumClusterV2Labels(ctx, plan.Labels, &diags))
		updateReq.SetName(plan.Name.ValueString())
		updateReq.SetConfig(expandYandexMdbGreenplumClusterV2Config(ctx, plan.Config, &diags))
		updateReq.SetMasterConfig(expandYandexMdbGreenplumClusterV2MasterConfig(ctx, plan.MasterConfig, &diags))
		updateReq.SetSegmentConfig(expandYandexMdbGreenplumClusterV2SegmentConfig(ctx, plan.SegmentConfig, &diags))
		updateReq.SetUserPassword(plan.UserPassword.ValueString())
		updateReq.SetNetworkId(plan.NetworkId.ValueString())
		updateReq.SetMaintenanceWindow(expandYandexMdbGreenplumClusterV2MaintenanceWindow(ctx, plan.MaintenanceWindow, &diags))
		updateReq.SetSecurityGroupIds(expandYandexMdbGreenplumClusterV2SecurityGroupIds(ctx, plan.SecurityGroupIds, &diags))
		updateReq.SetDeletionProtection(plan.DeletionProtection.ValueBool())
		updateReq.SetConfigSpec(expandYandexMdbGreenplumClusterV2ClusterConfig_update(ctx, plan.ClusterConfig, &diags))
		updateReq.SetCloudStorage(expandYandexMdbGreenplumClusterV2CloudStorage(ctx, plan.CloudStorage, &diags))
		updateReq.SetServiceAccountId(plan.ServiceAccountId.ValueString())
		updateReq.SetLogging(expandYandexMdbGreenplumClusterV2Logging(ctx, plan.Logging, &diags))
		updateReq.SetUpdateMask(&field_mask.FieldMask{Paths: updatePaths})

		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		tflog.Debug(ctx, fmt.Sprintf("Update cluster request: %s", validate.ProtoDump(updateReq)))

		md := new(metadata.MD)
		op, err := greenplumv1sdk.NewClusterClient(r.providerConfig.SDKv2).Update(ctx, updateReq, grpc.Header(md))
		if traceHeader := md.Get("x-server-trace-id"); len(traceHeader) > 0 {
			tflog.Debug(ctx, fmt.Sprintf("Update cluster x-server-trace-id: %s", traceHeader[0]))
		}
		if traceHeader := md.Get("x-server-request-id"); len(traceHeader) > 0 {
			tflog.Debug(ctx, fmt.Sprintf("Update cluster x-server-request-id: %s", traceHeader[0]))
		}
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to Update resource",
				"Error while requesting API to update cluster:"+err.Error(),
			)
			return
		}
		updateRes, err := op.Wait(ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Update Resource",
				fmt.Sprintf("An unexpected error occurred while waiting longrunning response. "+
					"Please retry the operation or report this issue to the provider developers.\n\n"+
					"Error: %s", err),
			)
			return
		}
		tflog.Debug(ctx, fmt.Sprintf("Update cluster response: %s", validate.ProtoDump(updateRes)))
	}
	if expandReq.SegmentHostCount != 0 || expandReq.AddSegmentsPerHostCount != 0 {
		ctx, cancel := context.WithTimeout(ctx, ClusterExpandTimeout)
		defer cancel()

		md := new(metadata.MD)
		op, err := greenplumv1sdk.NewClusterClient(r.providerConfig.SDKv2).Expand(ctx, expandReq, grpc.Header(md))
		if traceHeader := md.Get("x-server-trace-id"); len(traceHeader) > 0 {
			tflog.Debug(ctx, fmt.Sprintf("Expand cluster x-server-trace-id: %s", traceHeader[0]))
		}
		if traceHeader := md.Get("x-server-request-id"); len(traceHeader) > 0 {
			tflog.Debug(ctx, fmt.Sprintf("Expand cluster x-server-request-id: %s", traceHeader[0]))
		}
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to Update resource",
				"Error while requesting API to update cluster:"+err.Error(),
			)
			return
		}
		expandRes, err := op.Wait(ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to Update Resource",
				fmt.Sprintf("An unexpected error occurred while waiting longrunning response. "+
					"Please retry the operation or report this issue to the provider developers.\n\n"+
					"Error: %s", err),
			)
			return
		}
		tflog.Debug(ctx, fmt.Sprintf("Expand cluster response: %s", validate.ProtoDump(expandRes)))
	}

	r.refreshResourceState(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
