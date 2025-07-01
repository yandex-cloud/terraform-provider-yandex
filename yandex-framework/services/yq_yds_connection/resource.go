package yq_yds_connection

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

type ydsConnectionStrategy struct {
}

func (r *ydsConnectionStrategy) ExpandSetting(ctx context.Context, config *provider_config.Config, plan *tfsdk.Plan, diagnostics *diag.Diagnostics) *Ydb_FederatedQuery.ConnectionSetting {
	var model ydsConnectionModel
	diagnostics.Append(plan.Get(ctx, &model)...)
	if diagnostics.HasError() {
		return nil
	}

	serviceAccountID := model.ServiceAccountID.ValueString()
	databaseID := model.DatabaseID.ValueString()
	sharedReading := model.SharedReading.ValueBool()

	auth := yqcommon.ParseServiceIDToIAMAuth(serviceAccountID)
	return &Ydb_FederatedQuery.ConnectionSetting{
		Connection: &Ydb_FederatedQuery.ConnectionSetting_DataStreams{
			DataStreams: &Ydb_FederatedQuery.DataStreams{
				DatabaseId:    databaseID,
				SharedReading: sharedReading,
				Auth:          auth,
			},
		},
	}
}

func (r *ydsConnectionStrategy) PackToState(ctx context.Context, setting *Ydb_FederatedQuery.ConnectionSetting, state *tfsdk.State, diagnostics *diag.Diagnostics) {
	var model ydsConnectionModel
	yds := setting.GetDataStreams()
	if yds == nil {
		diagnostics.AddError("unexpected null YDS content setting from server", "")
		return
	}
	model.DatabaseID = types.StringValue(yds.GetDatabaseId())
	model.SharedReading = types.BoolValue(yds.GetSharedReading())
	serviceAccountId, err := yqcommon.IAMAuthToString(yds.GetAuth())
	if err != nil {
		diagnostics.AddError("Failed to extract auth info from connection", err.Error())
		return
	}
	model.ServiceAccountID = types.StringValue(serviceAccountId)

	diagnostics.Append(state.Set(ctx, &model)...)
}

func newYdsConnectionStrategy() yqcommon.ConnectionStrategy {
	return &ydsConnectionStrategy{}
}

func newYdsConnectionResourceSchema() map[string]schema.Attribute {
	return yqcommon.NewConnectionResourceSchema(yqcommon.AttributeDatabaseID, yqcommon.AttributeSharedReading)
}

func NewResource() resource.Resource {
	return yqcommon.NewBaseConnectionResource(
		newYdsConnectionResourceSchema(),
		newYdsConnectionStrategy(),
		"_yq_yds_connection",
		"Manages Yandex DataStreams connection in Yandex Query service. For more information, see [the official documentation](https://yandex.cloud/docs/query/concepts/glossary#connection).\n\n")
}
