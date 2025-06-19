package yqcommon

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
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
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb"
)

var (
	_ resource.Resource                = &baseBindingResource{}
	_ resource.ResourceWithConfigure   = &baseBindingResource{}
	_ resource.ResourceWithImportState = &baseBindingResource{}
)

type bindingBaseModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	ConnectionID types.String `tfsdk:"connection_id"`
}

type ColumnModel struct {
	Name    types.String `tfsdk:"name"`
	Type    types.String `tfsdk:"type"`
	NotNull types.Bool   `tfsdk:"not_null"`
}

var ColumnModelType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"name":     types.StringType,
		"type":     types.StringType,
		"not_null": types.BoolType,
	},
}

type baseBindingResource struct {
	providerConfig            *provider_config.Config
	attributes                map[string]schema.Attribute
	blocks                    map[string]schema.Block
	strategy                  BindingStrategy
	metadataSuffix            string
	schemaMarkdownDescription string
}

func NewBaseBindingResource(attributes map[string]schema.Attribute, blocks map[string]schema.Block, strategy BindingStrategy, metadataSuffix string, schemaMarkdownDescription string) resource.Resource {
	return &baseBindingResource{
		attributes:                attributes,
		blocks:                    blocks,
		strategy:                  strategy,
		metadataSuffix:            metadataSuffix,
		schemaMarkdownDescription: schemaMarkdownDescription,
	}
}

func (r *baseBindingResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + r.metadataSuffix
}

func (r *baseBindingResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	tflog.Debug(ctx, "Initializing YQ binding schema")
	resp.Schema = schema.Schema{
		MarkdownDescription: r.schemaMarkdownDescription,
		Attributes:          r.attributes,
		Blocks:              r.blocks,
	}
}

func (r *baseBindingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(AttributeID), req, resp)
}

func unpackBaseBindingModel(ctx context.Context, plan *tfsdk.Plan, baseModel *bindingBaseModel, diagnostics *diag.Diagnostics) bool {
	diagnostics.Append(plan.GetAttribute(ctx, path.Root(AttributeID), &baseModel.ID)...)
	if diagnostics.HasError() {
		return false
	}

	diagnostics.Append(plan.GetAttribute(ctx, path.Root(AttributeName), &baseModel.Name)...)
	if diagnostics.HasError() {
		return false
	}

	diagnostics.Append(plan.GetAttribute(ctx, path.Root(AttributeConnectionID), &baseModel.ConnectionID)...)
	if diagnostics.HasError() {
		return false
	}

	diagnostics.Append(plan.GetAttribute(ctx, path.Root(AttributeDescription), &baseModel.Description)...)
	return !diagnostics.HasError()
}

func (r *baseBindingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var baseModel bindingBaseModel
	if !unpackBaseBindingModel(ctx, &req.Plan, &baseModel, &resp.Diagnostics) {
		return
	}

	setting := r.strategy.ExpandSetting(ctx, &req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	request := &Ydb_FederatedQuery.CreateBindingRequest{
		Content: &Ydb_FederatedQuery.BindingContent{
			Name:         baseModel.Name.ValueString(),
			Description:  baseModel.Description.ValueString(),
			ConnectionId: baseModel.ConnectionID.ValueString(),
			Setting:      setting,
			Acl: &Ydb_FederatedQuery.Acl{
				Visibility: Ydb_FederatedQuery.Acl_SCOPE,
			},
		},
	}

	res, err := r.GetClient().CreateBinding(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create binding", err.Error())
		return
	}

	r.ReadToStateById(ctx, res.BindingId, &resp.State, &resp.Diagnostics)
}

func (r *baseBindingResource) ReadToStateById(ctx context.Context, bindingId string, state *tfsdk.State, diagnostics *diag.Diagnostics) {
	tflog.Debug(ctx, fmt.Sprintf("Reading binding %v", bindingId))

	request := &Ydb_FederatedQuery.DescribeBindingRequest{
		BindingId: bindingId,
	}

	bindingRes, err := r.GetClient().DescribeBinding(ctx, request)
	if err != nil {
		diagnostics.AddError("Failed to describe binding", err.Error())
		return
	}

	if bindingRes == nil {
		diagnostics.AddError("Failed to describe binding, empty response", "")
		return
	}

	binding := bindingRes.GetBinding()
	if binding == nil {
		diagnostics.AddError("unexpected null binding from server", "")
		return
	}

	content := binding.GetContent()
	if content == nil {
		diagnostics.AddError("unexpected null binding content from server", "")
		return
	}

	setting := content.GetSetting()
	if setting == nil {
		diagnostics.AddError("unexpected null binding content setting from server", "")
		return
	}

	r.strategy.PackToState(ctx, setting, state, diagnostics)

	diagnostics.Append(state.SetAttribute(ctx, path.Root(AttributeID), types.StringValue(binding.GetMeta().GetId()))...)
	diagnostics.Append(state.SetAttribute(ctx, path.Root(AttributeName), types.StringValue(content.GetName()))...)
	diagnostics.Append(state.SetAttribute(ctx, path.Root(AttributeDescription), types.StringValue(content.GetDescription()))...)
	diagnostics.Append(state.SetAttribute(ctx, path.Root(AttributeConnectionID), types.StringValue(content.GetConnectionId()))...)
}

func (r *baseBindingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var bindingId types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root(AttributeID), &bindingId)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.ReadToStateById(ctx, bindingId.ValueString(), &resp.State, &resp.Diagnostics)
}

func (r *baseBindingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var baseModel bindingBaseModel
	if !unpackBaseBindingModel(ctx, &req.Plan, &baseModel, &resp.Diagnostics) {
		return
	}

	setting := r.strategy.ExpandSetting(ctx, &req.Plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || setting == nil {
		return
	}

	bindingId := baseModel.ID.ValueString()
	request := &Ydb_FederatedQuery.ModifyBindingRequest{
		BindingId: bindingId,
		Content: &Ydb_FederatedQuery.BindingContent{
			Name:         baseModel.Name.ValueString(),
			Description:  baseModel.Description.ValueString(),
			ConnectionId: baseModel.ConnectionID.ValueString(),
			Setting:      setting,
			Acl: &Ydb_FederatedQuery.Acl{
				Visibility: Ydb_FederatedQuery.Acl_SCOPE,
			},
		},
	}

	if err := r.GetClient().ModifyBinding(ctx, request); err != nil {
		resp.Diagnostics.AddError("Failed to modify binding", err.Error())
		return
	}

	r.ReadToStateById(ctx, bindingId, &resp.State, &resp.Diagnostics)
}

func (r *baseBindingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var bindingId types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root(AttributeID), &bindingId)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request := &Ydb_FederatedQuery.DeleteBindingRequest{
		BindingId: bindingId.ValueString(),
	}

	err := r.GetClient().DeleteBinding(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete binding", err.Error())
		return
	}
}

func (r *baseBindingResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *baseBindingResource) GetClient() yqsdk.YQClient {
	return r.providerConfig.YqSdk.Client()
}

func formatTypeString(t *Ydb.Type, diagnostics *diag.Diagnostics) string {
	typeId := t.GetTypeId()
	switch typeId {
	case Ydb.Type_STRING:
		return "String"
	case Ydb.Type_BOOL:
		return "Bool"
	case Ydb.Type_INT8:
		return "Int8"
	case Ydb.Type_UINT8:
		return "Uint8"
	case Ydb.Type_INT16:
		return "Int16"
	case Ydb.Type_UINT16:
		return "Uint16"
	case Ydb.Type_INT32:
		return "Int32"
	case Ydb.Type_UINT32:
		return "Uin32"
	case Ydb.Type_INT64:
		return "Int64"
	case Ydb.Type_UINT64:
		return "Uint64"
	case Ydb.Type_FLOAT:
		return "Float"
	case Ydb.Type_DOUBLE:
		return "Double"
	case Ydb.Type_DATE:
		return "Date"
	case Ydb.Type_DATETIME:
		return "Datetime"
	case Ydb.Type_TIMESTAMP:
		return "Timestamp"
	case Ydb.Type_INTERVAL:
		return "Interval"
	case Ydb.Type_TZ_DATE:
		return "TzDate"
	case Ydb.Type_TZ_DATETIME:
		return "TzDatetime"
	case Ydb.Type_TZ_TIMESTAMP:
		return "TzTimestamp"
	case Ydb.Type_DATE32:
		return "Date32"
	case Ydb.Type_DATETIME64:
		return "Datetime64"
	case Ydb.Type_TIMESTAMP64:
		return "Timestamp64"
	case Ydb.Type_INTERVAL64:
		return "Interval64"
	case Ydb.Type_UTF8:
		return "Utf8"
	case Ydb.Type_YSON:
		return "Yson"
	case Ydb.Type_JSON:
		return "Json"
	case Ydb.Type_UUID:
		return "Uuid"
	case Ydb.Type_JSON_DOCUMENT:
		return "JsonDocument"
	case Ydb.Type_DYNUMBER:
		return "DyNumber"
	}

	diagnostics.AddError(fmt.Sprintf("unsupported type %v", typeId), "")
	return ""
}

func unwrapOptional(t *Ydb.Type) *Ydb.Type {
	for t.GetOptionalType() != nil {
		t = t.GetOptionalType().GetItem()
	}
	return t
}

func flattenColumn(column *Ydb.Column, diagnostics *diag.Diagnostics) ColumnModel {
	var result ColumnModel
	result.Name = types.StringValue(column.Name)

	result.NotNull = types.BoolValue(column.Type.GetOptionalType() == nil)
	columnType := formatTypeString(unwrapOptional(column.Type), diagnostics)
	if diagnostics.HasError() {
		return result
	}
	result.Type = types.StringValue(columnType)

	return result
}

func FlattenSchema(ctx context.Context, schema *Ydb_FederatedQuery.Schema, diagnostics *diag.Diagnostics) types.List {
	if schema == nil {
		return types.ListUnknown(ColumnModelType)
	}

	columns := make([]ColumnModel, 0, len(schema.Column))
	for _, column := range schema.Column {
		c := flattenColumn(column, diagnostics)
		if diagnostics.HasError() {
			return types.ListUnknown(ColumnModelType)
		}
		columns = append(columns, c)
	}

	column, diag := types.ListValueFrom(ctx, ColumnModelType, columns)
	if diag.HasError() {
		diagnostics.Append(diag...)
		return types.ListUnknown(ColumnModelType)
	}
	return column
}

func makePrimitiveType(typeId Ydb.Type_PrimitiveTypeId) *Ydb.Type {
	return &Ydb.Type{
		Type: &Ydb.Type_TypeId{
			TypeId: typeId,
		},
	}
}

func baseParseColumnType(t string, diagnostics *diag.Diagnostics) *Ydb.Type {
	switch t {
	case "String":
		return makePrimitiveType(Ydb.Type_STRING)
	case "Bool":
		return makePrimitiveType(Ydb.Type_BOOL)
	case "Int32":
		return makePrimitiveType(Ydb.Type_INT32)
	case "Uint32":
		return makePrimitiveType(Ydb.Type_UINT32)
	case "Int64":
		return makePrimitiveType(Ydb.Type_INT64)
	case "Uint64":
		return makePrimitiveType(Ydb.Type_UINT64)
	case "Float":
		return makePrimitiveType(Ydb.Type_FLOAT)
	case "Double":
		return makePrimitiveType(Ydb.Type_DOUBLE)
	case "Yson":
		return makePrimitiveType(Ydb.Type_YSON)
	case "Utf8":
		return makePrimitiveType(Ydb.Type_UTF8)
	case "Json":
		return makePrimitiveType(Ydb.Type_JSON)
	case "Date":
		return makePrimitiveType(Ydb.Type_DATE)
	case "Datetime":
		return makePrimitiveType(Ydb.Type_DATETIME)
	case "Timestamp":
		return makePrimitiveType(Ydb.Type_TIMESTAMP)
	case "Interval":
		return makePrimitiveType(Ydb.Type_INTERVAL)
	case "Int8":
		return makePrimitiveType(Ydb.Type_INT8)
	case "Uint8":
		return makePrimitiveType(Ydb.Type_UINT8)
	case "Int16":
		return makePrimitiveType(Ydb.Type_INT16)
	case "Uint16":
		return makePrimitiveType(Ydb.Type_UINT16)
	case "TzDate":
		return makePrimitiveType(Ydb.Type_TZ_DATE)
	case "TzDatetime":
		return makePrimitiveType(Ydb.Type_TZ_DATETIME)
	case "TzTimestamp":
		return makePrimitiveType(Ydb.Type_TZ_TIMESTAMP)
	case "Uuid":
		return makePrimitiveType(Ydb.Type_UUID)
	case "Date32":
		return makePrimitiveType(Ydb.Type_DATE32)
	case "Datetime64":
		return makePrimitiveType(Ydb.Type_DATETIME64)
	case "Timestamp64":
		return makePrimitiveType(Ydb.Type_TIMESTAMP64)
	case "Interval64":
		return makePrimitiveType(Ydb.Type_INTERVAL64)
	}

	diagnostics.AddError(fmt.Sprintf("unsupported type %s", t), "")
	return nil
}

func wrapWithOptional(t *Ydb.Type) *Ydb.Type {
	if t == nil {
		return nil
	}

	return &Ydb.Type{
		Type: &Ydb.Type_OptionalType{
			OptionalType: &Ydb.OptionalType{
				Item: t,
			},
		},
	}
}

func wrapWithOptionalIfNeeded(t *Ydb.Type) *Ydb.Type {
	if t.GetOptionalType() != nil {
		return t
	}

	return wrapWithOptional(t)
}

func parseColumnType(t string, notNull bool, diagnostics *diag.Diagnostics) *Ydb.Type {
	c := baseParseColumnType(t, diagnostics)
	if diagnostics.HasError() {
		return nil
	}

	if notNull {
		return c
	}
	return wrapWithOptionalIfNeeded(c)
}

func ParseSchema(ctx context.Context, column *types.List, diagnostics *diag.Diagnostics) *Ydb_FederatedQuery.Schema {
	columns := make([]ColumnModel, 0, len(column.Elements()))
	diagnostics.Append(column.ElementsAs(ctx, &columns, false)...)
	if diagnostics.HasError() {
		return nil
	}

	result := make([]*Ydb.Column, 0, len(columns))
	for _, c := range columns {
		t := parseColumnType(c.Type.ValueString(), c.NotNull.ValueBool(), diagnostics)
		if diagnostics.HasError() {
			return nil
		}

		result = append(result, &Ydb.Column{
			Name: c.Name.ValueString(),
			Type: t,
		})
	}

	return &Ydb_FederatedQuery.Schema{
		Column: result,
	}
}
