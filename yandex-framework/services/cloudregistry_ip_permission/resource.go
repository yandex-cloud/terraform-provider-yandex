package cloudregistry_ip_permission

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/cloudregistry/v1"
	"github.com/yandex-cloud/go-sdk/operation"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

var _ resource.ResourceWithConfigure = (*yandexCloudregistryIPPermissionResource)(nil)
var _ resource.ResourceWithImportState = (*yandexCloudregistryIPPermissionResource)(nil)

type yandexCloudregistryIPPermissionResource struct {
	providerConfig *provider_config.Config
}

func NewResource() resource.Resource {
	return &yandexCloudregistryIPPermissionResource{}
}

func (r *yandexCloudregistryIPPermissionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "yandex_cloudregistry_registry_ip_permission"
}

func (r *yandexCloudregistryIPPermissionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerConfig, ok := req.ProviderData.(*provider_config.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *provider_config.Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.providerConfig = providerConfig
}

func (r *yandexCloudregistryIPPermissionResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = YandexCloudregistryIPPermissionResourceSchema(ctx)
}

func (r *yandexCloudregistryIPPermissionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("registry_id"), req, resp)
}

func (r *yandexCloudregistryIPPermissionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state yandexCloudregistryIPPermissionModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	readTimeout, timeoutInitError := state.Timeouts.Read(ctx, yandexCloudRegistryIPPermissionDefaultTimeout)
	if timeoutInitError != nil {
		resp.Diagnostics.Append(timeoutInitError...)
		return
	}

	ctx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()

	reqApi := &cloudregistry.ListIpPermissionsRequest{
		RegistryId: state.RegistryId.ValueString(),
	}

	tflog.Debug(ctx, fmt.Sprintf("Read IP permission request: %s", validate.ProtoDump(reqApi)))

	md := new(metadata.MD)
	res, err := r.providerConfig.SDK.CloudRegistry().Registry().ListIpPermissions(ctx, reqApi, grpc.Header(md))
	if traceHeader := md.Get("x-server-trace-id"); len(traceHeader) > 0 {
		tflog.Debug(ctx, fmt.Sprintf("Read IP permission x-server-trace-id: %s", traceHeader[0]))
	}
	if traceHeader := md.Get("x-server-request-id"); len(traceHeader) > 0 {
		tflog.Debug(ctx, fmt.Sprintf("Read IP permission x-server-request-id: %s", traceHeader[0]))
	}

	if err != nil {
		if validate.IsStatusWithCode(err, codes.NotFound) {
			resp.Diagnostics.AddWarning(
				"Failed to Read resource",
				"IP permission not found",
			)
		} else {
			resp.Diagnostics.AddError(
				"Failed to Read resource",
				"Failed to Read	Error while requesting API to get IP permission: "+err.Error(),
			)
		}
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Read IP permission response: %s", validate.ProtoDump(res)))

	if res == nil {
		resp.Diagnostics.AddError("Failed to read", "Resource not found")
		return
	}

	newState := flattenYandexCloudregistryIPPermission(ctx, res.GetPermissions(), state, state.RegistryId, state.Timeouts, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *yandexCloudregistryIPPermissionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan yandexCloudregistryIPPermissionModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, timeoutInitError := plan.Timeouts.Create(ctx, yandexCloudRegistryIPPermissionDefaultTimeout)
	if timeoutInitError != nil {
		resp.Diagnostics.Append(timeoutInitError...)
		return
	}

	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	ipPermissions := append(
		getCloudRegistryIPPermission(diags, ctx, plan.Pull, cloudregistry.IpPermission_PULL),
		getCloudRegistryIPPermission(diags, ctx, plan.Push, cloudregistry.IpPermission_PUSH)...)

	createReq := &cloudregistry.SetIpPermissionsRequest{
		RegistryId:    plan.RegistryId.ValueString(),
		IpPermissions: ipPermissions,
	}

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Create IP permission request: %s", validate.ProtoDump(createReq)))

	md := new(metadata.MD)

	op, err := r.providerConfig.SDK.WrapOperation(r.providerConfig.SDK.CloudRegistry().Registry().SetIpPermissions(ctx, createReq, grpc.Header(md)))
	if traceHeader := md.Get("x-server-trace-id"); len(traceHeader) > 0 {
		tflog.Debug(ctx, fmt.Sprintf("Create IP permission x-server-trace-id: %s", traceHeader[0]))
	}
	if traceHeader := md.Get("x-server-request-id"); len(traceHeader) > 0 {
		tflog.Debug(ctx, fmt.Sprintf("Create IP permission x-server-request-id: %s", traceHeader[0]))
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Create resource",
			"Error while requesting API to set IP permission: "+err.Error(),
		)
		return
	}
	if err = waitCloudRegistryIPPermissionOperation(ctx, op); err != nil {
		resp.Diagnostics.AddError(
			"Failed to Create resource",
			"Error while requesting API to set IP permission: "+err.Error(),
		)
		return
	}
	if op == nil {
		resp.Diagnostics.AddError(
			"Unable to Create Resource",
			fmt.Sprintf("Error waiting for operation: %s", err),
		)
		return
	}

	readReq := &cloudregistry.ListIpPermissionsRequest{
		RegistryId: plan.RegistryId.ValueString(),
	}

	tflog.Debug(ctx, fmt.Sprintf("Read IP permission request: %s", validate.ProtoDump(readReq)))

	md = new(metadata.MD)
	res, err := r.providerConfig.SDK.CloudRegistry().Registry().ListIpPermissions(ctx, readReq, grpc.Header(md))
	if traceHeader := md.Get("x-server-trace-id"); len(traceHeader) > 0 {
		tflog.Debug(ctx, fmt.Sprintf("Read IP permission x-server-trace-id: %s", traceHeader[0]))
	}
	if traceHeader := md.Get("x-server-request-id"); len(traceHeader) > 0 {
		tflog.Debug(ctx, fmt.Sprintf("Read IP permission x-server-request-id: %s", traceHeader[0]))
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Read after Create",
			"Error getting IP permission: "+err.Error(),
		)
		return
	}

	if res == nil {
		resp.Diagnostics.AddError("Failed to read", "Resource not found")
		return
	}

	newState := flattenYandexCloudregistryIPPermission(ctx, res.GetPermissions(), plan, plan.RegistryId, plan.Timeouts, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *yandexCloudregistryIPPermissionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state yandexCloudregistryIPPermissionModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, timeoutInitError := state.Timeouts.Delete(ctx, yandexCloudRegistryIPPermissionDefaultTimeout)
	if timeoutInitError != nil {
		resp.Diagnostics.Append(timeoutInitError...)
		return
	}

	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	deleteReq := &cloudregistry.SetIpPermissionsRequest{
		RegistryId:    state.RegistryId.ValueString(),
		IpPermissions: []*cloudregistry.IpPermission{},
	}

	tflog.Debug(ctx, fmt.Sprintf("Delete IP permission request: %s", validate.ProtoDump(deleteReq)))

	md := new(metadata.MD)
	op, err := r.providerConfig.SDK.WrapOperation(r.providerConfig.SDK.CloudRegistry().Registry().SetIpPermissions(ctx, deleteReq, grpc.Header(md)))
	if traceHeader := md.Get("x-server-trace-id"); len(traceHeader) > 0 {
		tflog.Debug(ctx, fmt.Sprintf("Delete IP permission x-server-trace-id: %s", traceHeader[0]))
	}
	if traceHeader := md.Get("x-server-request-id"); len(traceHeader) > 0 {
		tflog.Debug(ctx, fmt.Sprintf("Delete IP permission x-server-request-id: %s", traceHeader[0]))
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Delete resource",
			"Error while requesting API to delete IP permission: "+err.Error(),
		)
		return
	}

	if err := waitCloudRegistryIPPermissionOperation(ctx, op); err != nil {
		resp.Diagnostics.AddError(
			"Failed to Create resource",
			"Error while requesting API to set IP permission: "+err.Error(),
		)
		return
	}
	if op == nil {
		resp.Diagnostics.AddError(
			"Unable to Delete Resource",
			fmt.Sprintf("Error waiting for operation: %s", err),
		)
	}
}

func (r *yandexCloudregistryIPPermissionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state yandexCloudregistryIPPermissionModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateTimeout, timeoutInitError := plan.Timeouts.Update(ctx, yandexCloudRegistryIPPermissionDefaultTimeout)
	if timeoutInitError != nil {
		resp.Diagnostics.Append(timeoutInitError...)
		return
	}

	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	ipPermissions := append(
		getCloudRegistryIPPermission(diags, ctx, plan.Pull, cloudregistry.IpPermission_PULL),
		getCloudRegistryIPPermission(diags, ctx, plan.Push, cloudregistry.IpPermission_PUSH)...)

	updateReq := &cloudregistry.SetIpPermissionsRequest{
		RegistryId:    plan.RegistryId.ValueString(),
		IpPermissions: ipPermissions,
	}

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Update IP permission request: %s", validate.ProtoDump(updateReq)))

	md := new(metadata.MD)
	op, err := r.providerConfig.SDK.WrapOperation(r.providerConfig.SDK.CloudRegistry().Registry().SetIpPermissions(ctx, updateReq, grpc.Header(md)))
	if traceHeader := md.Get("x-server-trace-id"); len(traceHeader) > 0 {
		tflog.Debug(ctx, fmt.Sprintf("Update IP permission x-server-trace-id: %s", traceHeader[0]))
	}
	if traceHeader := md.Get("x-server-request-id"); len(traceHeader) > 0 {
		tflog.Debug(ctx, fmt.Sprintf("Update IP permission x-server-request-id: %s", traceHeader[0]))
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Update resource",
			"Error while requesting API to update IP permission: "+err.Error(),
		)
		return
	}
	if err := waitCloudRegistryIPPermissionOperation(ctx, op); err != nil {
		resp.Diagnostics.AddError(
			"Failed to Create resource",
			"Error while requesting API to set IP permission: "+err.Error(),
		)
		return
	}
	if op == nil {
		resp.Diagnostics.AddError(
			"Unable to Update Resource",
			fmt.Sprintf("Error waiting for operation: %s", err),
		)
		return
	}

	readReq := &cloudregistry.ListIpPermissionsRequest{
		RegistryId: plan.RegistryId.ValueString(),
	}

	tflog.Debug(ctx, fmt.Sprintf("Read IP permission request: %s", validate.ProtoDump(readReq)))

	md = new(metadata.MD)
	res, err := r.providerConfig.SDK.CloudRegistry().Registry().ListIpPermissions(ctx, readReq, grpc.Header(md))
	if traceHeader := md.Get("x-server-trace-id"); len(traceHeader) > 0 {
		tflog.Debug(ctx, fmt.Sprintf("Read IP permission x-server-trace-id: %s", traceHeader[0]))
	}
	if traceHeader := md.Get("x-server-request-id"); len(traceHeader) > 0 {
		tflog.Debug(ctx, fmt.Sprintf("Read IP permission x-server-request-id: %s", traceHeader[0]))
	}

	if err != nil {
		if validate.IsStatusWithCode(err, codes.NotFound) {
			resp.Diagnostics.AddWarning(
				"Failed to Read resource",
				"IP permission not found",
			)
		} else {
			resp.Diagnostics.AddError(
				"Failed to Read resource",
				"Failed to Read	Error while requesting API to get IP permission: "+err.Error(),
			)
		}
		return
	}

	if res == nil {
		resp.Diagnostics.AddError("Failed to read", "Resource not found")
		return
	}

	newState := flattenYandexCloudregistryIPPermission(ctx, res.GetPermissions(), plan, plan.RegistryId, plan.Timeouts, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func getCloudRegistryIPPermission(diags diag.Diagnostics, ctx context.Context, tfSet types.Set, action cloudregistry.IpPermission_Action) []*cloudregistry.IpPermission {
	var permissions []*cloudregistry.IpPermission

	expanded, diag := expandCloudRegistryIPPermission(ctx, tfSet, action)
	permissions = append(permissions, expanded...)
	diags.Append(diag...)

	return permissions
}

func expandCloudRegistryIPPermission(ctx context.Context, tfSet types.Set, action cloudregistry.IpPermission_Action) ([]*cloudregistry.IpPermission, diag.Diagnostics) {
	var diags diag.Diagnostics

	if tfSet.IsNull() || tfSet.IsUnknown() {
		return nil, diags
	}

	var addresses []string
	diags.Append(tfSet.ElementsAs(ctx, &addresses, false)...)
	if diags.HasError() {
		return nil, diags
	}

	permissions := make([]*cloudregistry.IpPermission, len(addresses))
	for i, addr := range addresses {
		permissions[i] = &cloudregistry.IpPermission{
			Ip:     addr,
			Action: action,
		}
	}

	return permissions, diags
}

func waitCloudRegistryIPPermissionOperation(ctx context.Context, operation *operation.Operation) error {
	if err := operation.Wait(ctx); err != nil {
		return fmt.Errorf("error while waiting operation to set ip permission: %s", err)
	}

	if _, err := operation.Response(); err != nil {
		return fmt.Errorf("failed to set ip permission: %s", err)
	}

	return nil
}
