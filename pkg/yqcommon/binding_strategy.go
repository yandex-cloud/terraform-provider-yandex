package yqcommon

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/ydb-platform/ydb-go-genproto/draft/protos/Ydb_FederatedQuery"
)

type BindingStrategy interface {
	PackToState(ctx context.Context, setting *Ydb_FederatedQuery.BindingSetting, state *tfsdk.State, diagnostics *diag.Diagnostics)
	ExpandSetting(ctx context.Context, plan *tfsdk.Plan, diagnostics *diag.Diagnostics) *Ydb_FederatedQuery.BindingSetting
}
