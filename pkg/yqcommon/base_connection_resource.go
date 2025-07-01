package yqcommon

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/yqsdk"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"github.com/ydb-platform/ydb-go-genproto/draft/protos/Ydb_FederatedQuery"
)

var (
	_ resource.Resource                = &baseConnectionResource{}
	_ resource.ResourceWithConfigure   = &baseConnectionResource{}
	_ resource.ResourceWithImportState = &baseConnectionResource{}
)

type connectionBaseModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

type baseConnectionResource struct {
	providerConfig            *provider_config.Config
	attributes                map[string]schema.Attribute
	strategy                  ConnectionStrategy
	metadataSuffix            string
	schemaMarkdownDescription string
}

func (r *baseConnectionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + r.metadataSuffix
}

func NewBaseConnectionResource(attributes map[string]schema.Attribute, strategy ConnectionStrategy, metadataSuffix string, schemaMarkdownDescription string) resource.Resource {
	return &baseConnectionResource{
		attributes:                attributes,
		strategy:                  strategy,
		metadataSuffix:            metadataSuffix,
		schemaMarkdownDescription: schemaMarkdownDescription,
	}
}

func (r *baseConnectionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	tflog.Debug(ctx, "Initializing YQ connection schema")
	resp.Schema = schema.Schema{
		MarkdownDescription: r.schemaMarkdownDescription,
		Attributes:          r.attributes,
	}
}

func (r *baseConnectionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(AttributeID), req, resp)
}

func unpackBaseModel(ctx context.Context, plan *tfsdk.Plan, baseModel *connectionBaseModel, diagnostics *diag.Diagnostics) bool {
	diagnostics.Append(plan.GetAttribute(ctx, path.Root(AttributeID), &baseModel.ID)...)
	if diagnostics.HasError() {
		return false
	}

	diagnostics.Append(plan.GetAttribute(ctx, path.Root(AttributeName), &baseModel.Name)...)
	if diagnostics.HasError() {
		return false
	}

	diagnostics.Append(plan.GetAttribute(ctx, path.Root(AttributeDescription), &baseModel.Description)...)
	return !diagnostics.HasError()
}

func (r *baseConnectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var baseModel connectionBaseModel
	if !unpackBaseModel(ctx, &req.Plan, &baseModel, &resp.Diagnostics) {
		return
	}

	setting := r.strategy.ExpandSetting(ctx, r.providerConfig, &req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	request := &Ydb_FederatedQuery.CreateConnectionRequest{
		Content: &Ydb_FederatedQuery.ConnectionContent{
			Name:        baseModel.Name.ValueString(),
			Description: baseModel.Description.ValueString(),
			Setting:     setting,
			Acl: &Ydb_FederatedQuery.Acl{
				Visibility: Ydb_FederatedQuery.Acl_SCOPE,
			},
		},
	}

	res, err := r.GetClient().CreateConnection(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create connection", err.Error())
		return
	}

	r.ReadToStateById(ctx, res.ConnectionId, &resp.State, &resp.Diagnostics)
}

func (r *baseConnectionResource) ReadToStateById(ctx context.Context, connectionId string, state *tfsdk.State, diagnostics *diag.Diagnostics) {
	tflog.Debug(ctx, fmt.Sprintf("Reading connection %v", connectionId))

	request := &Ydb_FederatedQuery.DescribeConnectionRequest{
		ConnectionId: connectionId,
	}

	connectionRes, err := r.GetClient().DescribeConnection(ctx, request)
	if err != nil {
		diagnostics.AddError("Failed to describe connection", err.Error())
		return
	}

	if connectionRes == nil {
		diagnostics.AddError("Failed to describe connection, empty response", "")
		return
	}

	connection := connectionRes.GetConnection()
	if connection == nil {
		diagnostics.AddError("unexpected null connection from server", "")
		return
	}

	content := connection.GetContent()
	if content == nil {
		diagnostics.AddError("unexpected null connection content from server", "")
		return
	}

	setting := content.GetSetting()
	if setting == nil {
		diagnostics.AddError("unexpected null connection content setting from server", "")
		return
	}

	r.strategy.PackToState(ctx, setting, state, diagnostics)

	diagnostics.Append(state.SetAttribute(ctx, path.Root(AttributeID), types.StringValue(connection.GetMeta().GetId()))...)
	diagnostics.Append(state.SetAttribute(ctx, path.Root(AttributeName), types.StringValue(content.GetName()))...)
	diagnostics.Append(state.SetAttribute(ctx, path.Root(AttributeDescription), types.StringValue(content.GetDescription()))...)
}

func (r *baseConnectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var connectionId types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root(AttributeID), &connectionId)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.ReadToStateById(ctx, connectionId.ValueString(), &resp.State, &resp.Diagnostics)
}

func (r *baseConnectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var baseModel connectionBaseModel
	if !unpackBaseModel(ctx, &req.Plan, &baseModel, &resp.Diagnostics) {
		return
	}

	setting := r.strategy.ExpandSetting(ctx, r.providerConfig, &req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || setting == nil {
		return
	}

	connectionId := baseModel.ID.ValueString()
	request := &Ydb_FederatedQuery.ModifyConnectionRequest{
		ConnectionId: connectionId,
		Content: &Ydb_FederatedQuery.ConnectionContent{
			Name:        baseModel.Name.ValueString(),
			Description: baseModel.Description.ValueString(),
			Setting:     setting,
			Acl: &Ydb_FederatedQuery.Acl{
				Visibility: Ydb_FederatedQuery.Acl_SCOPE,
			},
		},
	}

	if err := r.GetClient().ModifyConnection(ctx, request); err != nil {
		resp.Diagnostics.AddError("Failed to modify connection", err.Error())
		return
	}

	r.ReadToStateById(ctx, connectionId, &resp.State, &resp.Diagnostics)
}

func (r *baseConnectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var connectionId types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root(AttributeID), &connectionId)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request := &Ydb_FederatedQuery.DeleteConnectionRequest{
		ConnectionId: connectionId.ValueString(),
	}

	err := r.GetClient().DeleteConnection(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete connection", err.Error())
		return
	}
}

func (r *baseConnectionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *baseConnectionResource) GetClient() yqsdk.YQClient {
	return r.providerConfig.YqSdk.Client()
}
