package yqcommon

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/ydb-platform/ydb-go-genproto/draft/protos/Ydb_FederatedQuery"
)

type ConnectionStrategy interface {
	PackToState(ctx context.Context, setting *Ydb_FederatedQuery.ConnectionSetting, state *tfsdk.State, diagnostics *diag.Diagnostics)
	ExpandSetting(ctx context.Context, sdk *ycsdk.SDK, plan *tfsdk.Plan, diagnostics *diag.Diagnostics) *Ydb_FederatedQuery.ConnectionSetting
}
