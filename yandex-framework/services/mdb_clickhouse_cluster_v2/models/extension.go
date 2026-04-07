package models

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
)

type Extension struct {
	Name    types.String `tfsdk:"name"`
	Version types.String `tfsdk:"version"`
}

var ExtensionAttrTypes = map[string]attr.Type{
	"name":    types.StringType,
	"version": types.StringType,
}

func FlattenListExtensions(ctx context.Context, extensions []*clickhouse.ClusterExtension, diags *diag.Diagnostics) types.Set {
	if extensions == nil {
		return types.SetNull(types.ObjectType{AttrTypes: ExtensionAttrTypes})
	}

	tfExtensions := make([]types.Object, len(extensions))
	for i, r := range extensions {
		tfExtensions[i] = FlattenExtension(ctx, r, diags)
	}
	set, d := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: ExtensionAttrTypes}, tfExtensions)

	diags.Append(d...)

	return set
}

func FlattenExtension(ctx context.Context, extension *clickhouse.ClusterExtension, diags *diag.Diagnostics) types.Object {
	if extension == nil {
		return types.ObjectNull(ExtensionAttrTypes)
	}

	obj, d := types.ObjectValueFrom(
		ctx, ExtensionAttrTypes, Extension{
			Name:    types.StringValue(extension.Name),
			Version: types.StringValue(extension.Version),
		},
	)
	diags.Append(d...)
	return obj
}

func ExpandListExtensions(ctx context.Context, extensions types.Set, diags *diag.Diagnostics) []*clickhouse.ExtensionSpec {
	emptyList := []*clickhouse.ExtensionSpec{}

	if extensions.IsNull() || extensions.IsUnknown() {
		return emptyList
	}

	result := make([]*clickhouse.ExtensionSpec, 0, len(extensions.Elements()))
	modelExtensions := make([]Extension, 0, len(extensions.Elements()))
	diags.Append(extensions.ElementsAs(ctx, &modelExtensions, false)...)
	if diags.HasError() {
		return emptyList
	}

	for _, e := range modelExtensions {
		result = append(result, &clickhouse.ExtensionSpec{
			Name:    e.Name.ValueString(),
			Version: e.Version.ValueString(),
		})
	}
	return result
}
