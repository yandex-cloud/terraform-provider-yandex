package model

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/opensearch/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
)

type Access struct {
	DataTransfer types.Bool `tfsdk:"data_transfer"`
	Serverless   types.Bool `tfsdk:"serverless"`
}

var accessAttrTypes = map[string]attr.Type{
	"data_transfer": types.BoolType,
	"serverless":    types.BoolType,
}

func accessToObject(ctx context.Context, cfg *opensearch.Access) (types.Object, diag.Diagnostics) {
	if cfg == nil {
		return types.ObjectNull(accessAttrTypes), nil
	}

	return types.ObjectValueFrom(ctx, accessAttrTypes, Access{
		DataTransfer: types.BoolValue(cfg.GetDataTransfer()),
		Serverless:   types.BoolValue(cfg.GetServerless()),
	})
}

func ParseAccess(ctx context.Context, state *Config) (*Access, diag.Diagnostics) {
	res := &Access{}
	diags := state.Access.As(ctx, res, datasize.DefaultOpts)
	if diags.HasError() {
		return nil, diags
	}

	return res, diag.Diagnostics{}
}
