package yq_yds_binding

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/yqcommon"
	"github.com/ydb-platform/ydb-go-genproto/draft/protos/Ydb_FederatedQuery"
)

type ydsBindingStrategy struct {
}

func (r *ydsBindingStrategy) ExpandSetting(ctx context.Context, plan *tfsdk.Plan, diagnostics *diag.Diagnostics) *Ydb_FederatedQuery.BindingSetting {
	var model ydsBindingModel
	diagnostics.Append(plan.Get(ctx, &model)...)
	if diagnostics.HasError() {
		return nil
	}

	formatSetting := make(map[string]string)
	diagnostics.Append(model.FormatSetting.ElementsAs(ctx, &formatSetting, false)...)
	if diagnostics.HasError() {
		return nil
	}

	schema := yqcommon.ParseSchema(ctx, &model.Column, diagnostics)

	return &Ydb_FederatedQuery.BindingSetting{
		Binding: &Ydb_FederatedQuery.BindingSetting_DataStreams{
			DataStreams: &Ydb_FederatedQuery.DataStreamsBinding{
				Format:        model.Format.ValueString(),
				Compression:   model.Compression.ValueString(),
				StreamName:    model.Stream.ValueString(),
				Schema:        schema,
				FormatSetting: formatSetting,
			},
		},
	}
}

func (r *ydsBindingStrategy) PackToState(ctx context.Context, setting *Ydb_FederatedQuery.BindingSetting, state *tfsdk.State, diagnostics *diag.Diagnostics) {
	var model ydsBindingModel
	ydsSetting := setting.GetDataStreams()
	if ydsSetting == nil {
		diagnostics.AddError("unexpected null settings", "")
		return
	}

	model.Stream = types.StringValue(ydsSetting.GetStreamName())
	model.Format = types.StringValue(ydsSetting.GetFormat())
	model.Compression = types.StringValue(ydsSetting.GetCompression())

	formatSetting, diag := types.MapValueFrom(ctx, types.StringType, ydsSetting.GetFormatSetting())
	if diag.HasError() {
		diagnostics.Append(diag...)
		return
	}
	model.FormatSetting = formatSetting

	schema := ydsSetting.GetSchema()
	model.Column = yqcommon.FlattenSchema(ctx, schema, diagnostics)
	if diagnostics.HasError() {
		return
	}

	diagnostics.Append(state.Set(ctx, &model)...)
}

func newYdsBindingStrategy() yqcommon.BindingStrategy {
	return &ydsBindingStrategy{}
}

func newYdsBindingResourceSchema() (map[string]schema.Attribute, map[string]schema.Block) {
	return yqcommon.NewBindingResourceSchema(yqcommon.AttributeStream)
}

func NewResource() resource.Resource {
	attributes, blocks := newYdsBindingResourceSchema()
	return yqcommon.NewBaseBindingResource(
		attributes,
		blocks,
		newYdsBindingStrategy(),
		"_yq_yds_binding",
		"Manages Yandex DataStreams binding in Yandex Query service. For more information, see [the official documentation](https://yandex.cloud/docs/query/concepts/glossary#Binding).\n\n")
}
