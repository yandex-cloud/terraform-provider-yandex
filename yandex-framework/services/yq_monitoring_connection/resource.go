package yq_monitoring_connection

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/resourcemanager/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/yqcommon"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"github.com/ydb-platform/ydb-go-genproto/draft/protos/Ydb_FederatedQuery"
)

type monitoringConnectionStrategy struct {
}

func (r *monitoringConnectionStrategy) ExpandSetting(ctx context.Context, config *provider_config.Config, plan *tfsdk.Plan, diagnostics *diag.Diagnostics) *Ydb_FederatedQuery.ConnectionSetting {
	var model monitoringConnectionModel
	diagnostics.Append(plan.Get(ctx, &model)...)
	if diagnostics.HasError() {
		return nil
	}

	serviceAccountID := model.ServiceAccountID.ValueString()
	folderID, d := validate.FolderID(model.FolderID, &config.ProviderState)
	diagnostics.Append(d)
	if diagnostics.HasError() {
		return nil
	}

	cloudID := model.CloudID.ValueString()
	if len(cloudID) == 0 {
		folder, err := config.SDK.ResourceManager().Folder().Get(ctx, &resourcemanager.GetFolderRequest{
			FolderId: folderID,
		})
		if err != nil {
			diagnostics.AddError("Failed to extract auth info from connection", err.Error())
			return nil
		}
		cloudID = folder.CloudId
	}

	auth := yqcommon.ParseServiceIDToIAMAuth(serviceAccountID)
	return &Ydb_FederatedQuery.ConnectionSetting{
		Connection: &Ydb_FederatedQuery.ConnectionSetting_Monitoring{
			Monitoring: &Ydb_FederatedQuery.Monitoring{
				Project: cloudID,
				Cluster: folderID,
				Auth:    auth,
			},
		},
	}
}

func (r *monitoringConnectionStrategy) PackToState(ctx context.Context, setting *Ydb_FederatedQuery.ConnectionSetting, state *tfsdk.State, diagnostics *diag.Diagnostics) {
	var model monitoringConnectionModel
	monitoring := setting.GetMonitoring()
	if monitoring == nil {
		diagnostics.AddError("unexpected null Monitoring content setting from server", "")
		return
	}
	model.CloudID = types.StringValue(monitoring.GetProject())
	model.FolderID = types.StringValue(monitoring.GetCluster())
	serviceAccountId, err := yqcommon.IAMAuthToString(monitoring.GetAuth())
	if err != nil {
		diagnostics.AddError("Failed to extract auth info from connection", err.Error())
		return
	}
	model.ServiceAccountID = types.StringValue(serviceAccountId)

	diagnostics.Append(state.Set(ctx, &model)...)
}

func newMonitoringConnectionStrategy() yqcommon.ConnectionStrategy {
	return &monitoringConnectionStrategy{}
}

func newMonitoringConnectionResourceSchema() map[string]schema.Attribute {
	return yqcommon.NewConnectionResourceSchema(yqcommon.AttributeCloudID, yqcommon.AttributeFolderID)
}

func NewResource() resource.Resource {
	return yqcommon.NewBaseConnectionResource(
		newMonitoringConnectionResourceSchema(),
		newMonitoringConnectionStrategy(),
		"_yq_monitoring_connection",
		"Manages Monitoring connection in Yandex Query service. For more information, see [the official documentation](https://yandex.cloud/docs/query/concepts/glossary#connection).\n\n")
}
