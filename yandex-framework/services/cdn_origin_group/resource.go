package cdn_origin_group

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/cdn/v1"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	yandexCDNOriginGroupDefaultTimeout = 2 * time.Minute
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.Resource                = &cdnOriginGroupResource{}
	_ resource.ResourceWithConfigure   = &cdnOriginGroupResource{}
	_ resource.ResourceWithImportState = &cdnOriginGroupResource{}
)

type cdnOriginGroupResource struct {
	providerConfig *provider_config.Config
}

// NewResource creates a new CDN origin group resource
func NewResource() resource.Resource {
	return &cdnOriginGroupResource{}
}

func (r *cdnOriginGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cdn_origin_group"
}

func (r *cdnOriginGroupResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = CDNOriginGroupSchema(ctx)
}

func (r *cdnOriginGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *cdnOriginGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CDNOriginGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := plan.Timeouts.Create(ctx, yandexCDNOriginGroupDefaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	// Get folder ID
	folderID := r.getFolderID(&plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Determine provider type
	providerType := "ourcdn"
	if !plan.ProviderType.IsNull() && plan.ProviderType.ValueString() != "" {
		providerType = plan.ProviderType.ValueString()
	}

	// Extract origins
	var origins []OriginModel
	resp.Diagnostics.Append(plan.Origins.ElementsAs(ctx, &origins, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create request
	createReq := &cdn.CreateOriginGroupRequest{
		FolderId:     folderID,
		Name:         plan.Name.ValueString(),
		ProviderType: providerType,
		UseNext:      &wrappers.BoolValue{Value: plan.UseNext.ValueBool()},
		Origins:      expandOrigins(ctx, origins, &resp.Diagnostics),
	}

	tflog.Debug(ctx, "Creating CDN origin group", map[string]interface{}{
		"name":          createReq.Name,
		"folder_id":     createReq.FolderId,
		"provider_type": createReq.ProviderType,
	})

	op, err := r.providerConfig.SDK.WrapOperation(r.providerConfig.SDK.CDN().OriginGroup().Create(ctx, createReq))
	if err != nil {
		resp.Diagnostics.AddError("Failed to create CDN origin group", err.Error())
		return
	}

	// Wait for operation
	if err := op.Wait(ctx); err != nil {
		resp.Diagnostics.AddError("Error waiting for CDN origin group creation", err.Error())
		return
	}

	// Get metadata
	md, err := op.Metadata()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get operation metadata", err.Error())
		return
	}

	metadata, ok := md.(*cdn.CreateOriginGroupMetadata)
	if !ok {
		resp.Diagnostics.AddError("Unexpected metadata type", "Expected *cdn.CreateOriginGroupMetadata")
		return
	}

	plan.ID = types.StringValue(strconv.FormatInt(metadata.OriginGroupId, 10))

	// Read the created resource to populate computed fields
	if !r.readResourceToState(ctx, &plan, &resp.Diagnostics) {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *cdnOriginGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CDNOriginGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !r.readResourceToState(ctx, &state, &resp.Diagnostics) {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *cdnOriginGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CDNOriginGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateTimeout, diags := plan.Timeouts.Update(ctx, yandexCDNOriginGroupDefaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	// Get folder ID
	folderID := r.getFolderID(&plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse origin group ID
	groupID, err := strconv.ParseInt(plan.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid origin group ID", err.Error())
		return
	}

	// Extract origins
	var origins []OriginModel
	resp.Diagnostics.Append(plan.Origins.ElementsAs(ctx, &origins, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update request
	updateReq := &cdn.UpdateOriginGroupRequest{
		FolderId:      folderID,
		OriginGroupId: groupID,
		GroupName:     &wrappers.StringValue{Value: plan.Name.ValueString()},
		UseNext:       &wrappers.BoolValue{Value: plan.UseNext.ValueBool()},
		Origins:       expandOrigins(ctx, origins, &resp.Diagnostics),
	}

	tflog.Debug(ctx, "Updating CDN origin group", map[string]interface{}{
		"id":   plan.ID.ValueString(),
		"name": updateReq.GroupName.Value,
	})

	op, err := r.providerConfig.SDK.WrapOperation(r.providerConfig.SDK.CDN().OriginGroup().Update(ctx, updateReq))
	if err != nil {
		resp.Diagnostics.AddError("Failed to update CDN origin group", err.Error())
		return
	}

	// Wait for operation
	if err := op.Wait(ctx); err != nil {
		resp.Diagnostics.AddError("Error waiting for CDN origin group update", err.Error())
		return
	}

	// Read updated resource
	if !r.readResourceToState(ctx, &plan, &resp.Diagnostics) {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *cdnOriginGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CDNOriginGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := state.Timeouts.Delete(ctx, yandexCDNOriginGroupDefaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	// Get folder ID
	folderID := r.getFolderID(&state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse origin group ID
	groupID, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid origin group ID", err.Error())
		return
	}

	deleteReq := &cdn.DeleteOriginGroupRequest{
		FolderId:      folderID,
		OriginGroupId: groupID,
	}

	tflog.Debug(ctx, "Deleting CDN origin group", map[string]interface{}{
		"id": state.ID.ValueString(),
	})

	op, err := r.providerConfig.SDK.WrapOperation(r.providerConfig.SDK.CDN().OriginGroup().Delete(ctx, deleteReq))
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			tflog.Info(ctx, "CDN origin group already deleted", map[string]interface{}{
				"id": state.ID.ValueString(),
			})
			return
		}
		resp.Diagnostics.AddError("Failed to delete CDN origin group", err.Error())
		return
	}

	// Wait for operation
	if err := op.Wait(ctx); err != nil {
		resp.Diagnostics.AddError("Error waiting for CDN origin group deletion", err.Error())
		return
	}

	tflog.Info(ctx, "CDN origin group deleted", map[string]interface{}{
		"id": state.ID.ValueString(),
	})
}

func (r *cdnOriginGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Validate ID format (must be a valid int64)
	_, err := strconv.ParseInt(req.ID, 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected origin group ID to be a number, got: %s", req.ID),
		)
		return
	}

	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// getFolderID returns folder ID from model or provider config
func (r *cdnOriginGroupResource) getFolderID(model *CDNOriginGroupModel, diags *diag.Diagnostics) string {
	if !model.FolderID.IsNull() && model.FolderID.ValueString() != "" {
		return model.FolderID.ValueString()
	}
	if r.providerConfig.ProviderState.FolderID.ValueString() != "" {
		return r.providerConfig.ProviderState.FolderID.ValueString()
	}
	diags.AddError("folder_id is required", "Please set folder_id in this resource or at provider level")
	return ""
}

// readResourceToState reads the origin group from API and updates the state
// Returns false if the resource was not found (should be removed from state)
func (r *cdnOriginGroupResource) readResourceToState(ctx context.Context, state *CDNOriginGroupModel, diags *diag.Diagnostics) bool {
	// Get folder ID
	folderID := r.getFolderID(state, diags)
	if diags.HasError() {
		return false
	}

	// Parse origin group ID
	groupID, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		diags.AddError("Invalid origin group ID", err.Error())
		return false
	}

	getReq := &cdn.GetOriginGroupRequest{
		FolderId:      folderID,
		OriginGroupId: groupID,
	}

	tflog.Debug(ctx, "Reading CDN origin group", map[string]interface{}{
		"id": state.ID.ValueString(),
	})

	originGroup, err := r.providerConfig.SDK.CDN().OriginGroup().Get(ctx, getReq)
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			tflog.Info(ctx, "CDN origin group not found", map[string]interface{}{
				"id": state.ID.ValueString(),
			})
			return false
		}
		diags.AddError("Failed to read CDN origin group", err.Error())
		return false
	}

	flattenCDNOriginGroup(ctx, state, originGroup, diags)
	return true
}
