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

func accessToObject(ctx context.Context, cfg *opensearch.Access, state *Config) (types.Object, diag.Diagnostics) {
	hasState := state != nil && !state.Access.IsNull() && !state.Access.IsUnknown()
	// The API may omit access when both flags are false. Preserve whether the
	// block was configured, since Terraform distinguishes an object from null.
	if !hasState && !cfg.GetDataTransfer() && !cfg.GetServerless() {
		return types.ObjectNull(accessAttrTypes), nil
	}

	access := Access{
		DataTransfer: types.BoolValue(cfg.GetDataTransfer()),
		Serverless:   types.BoolValue(cfg.GetServerless()),
	}
	if hasState {
		prior, diags := ParseAccess(ctx, state)
		if diags.HasError() {
			return types.ObjectUnknown(accessAttrTypes), diags
		}
		// Omitted optional flags also need to stay null when disabled. Always
		// retain actual API changes, including a previously enabled flag reset.
		if prior.DataTransfer.IsNull() && !cfg.GetDataTransfer() {
			access.DataTransfer = prior.DataTransfer
		}
		if prior.Serverless.IsNull() && !cfg.GetServerless() {
			access.Serverless = prior.Serverless
		}
	}
	return types.ObjectValueFrom(ctx, accessAttrTypes, access)
}

func ParseAccess(ctx context.Context, state *Config) (*Access, diag.Diagnostics) {
	res := &Access{}
	diags := state.Access.As(ctx, res, datasize.DefaultOpts)
	if diags.HasError() {
		return nil, diags
	}

	return res, diag.Diagnostics{}
}
