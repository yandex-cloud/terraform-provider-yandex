package yq_ydb_connection

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/yqcommon"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"github.com/ydb-platform/ydb-go-genproto/draft/protos/Ydb_FederatedQuery"
)

type ydbConnectionStrategy struct {
}

func (r *ydbConnectionStrategy) ExpandSetting(ctx context.Context, config *provider_config.Config, plan *tfsdk.Plan, diagnostics *diag.Diagnostics) *Ydb_FederatedQuery.ConnectionSetting {
	var model ydbConnectionModel
	diagnostics.Append(plan.Get(ctx, &model)...)
	if diagnostics.HasError() {
		return nil
	}

	serviceAccountID := model.ServiceAccountID.ValueString()
	databaseID := model.DatabaseID.ValueString()

	auth := yqcommon.ParseServiceIDToIAMAuth(serviceAccountID)
	return &Ydb_FederatedQuery.ConnectionSetting{
		Connection: &Ydb_FederatedQuery.ConnectionSetting_YdbDatabase{
			YdbDatabase: &Ydb_FederatedQuery.YdbDatabase{
				DatabaseId: databaseID,
				Auth:       auth,
			},
		},
	}
}

func (r *ydbConnectionStrategy) PackToState(ctx context.Context, setting *Ydb_FederatedQuery.ConnectionSetting, state *tfsdk.State, diagnostics *diag.Diagnostics) {
	var model ydbConnectionModel
	ydb := setting.GetYdbDatabase()
	if ydb == nil {
		diagnostics.AddError("unexpected null YDB content setting from server", "")
		return
	}
	model.DatabaseID = types.StringValue(ydb.GetDatabaseId())
	serviceAccountId, err := yqcommon.IAMAuthToString(ydb.GetAuth())
	if err != nil {
		diagnostics.AddError("Failed to extract auth info from connection", err.Error())
		return
	}
	model.ServiceAccountID = types.StringValue(serviceAccountId)

	diagnostics.Append(state.Set(ctx, &model)...)
}

func newYdbConnectionStrategy() yqcommon.ConnectionStrategy {
	return &ydbConnectionStrategy{}
}

func newYdbConnectionResourceSchema() map[string]schema.Attribute {
	return yqcommon.NewConnectionResourceSchema(yqcommon.AttributeDatabaseID)
}

func NewResource() resource.Resource {
	return yqcommon.NewBaseConnectionResource(
		newYdbConnectionResourceSchema(),
		newYdbConnectionStrategy(),
		"_yq_ydb_connection",
		"Manages YDB connection in Yandex Query service. For more information, see [the official documentation](https://yandex.cloud/docs/query/concepts/glossary#connection).\n\n")
}
