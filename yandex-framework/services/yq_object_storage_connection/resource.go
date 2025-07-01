package yq_object_storage_connection

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

type objectStorageConnectionStrategy struct {
}

func (r *objectStorageConnectionStrategy) ExpandSetting(ctx context.Context, config *provider_config.Config, plan *tfsdk.Plan, diagnostics *diag.Diagnostics) *Ydb_FederatedQuery.ConnectionSetting {
	var model objectStorageConnectionModel
	diagnostics.Append(plan.Get(ctx, &model)...)
	if diagnostics.HasError() {
		return nil
	}

	serviceAccountID := model.ServiceAccountID.ValueString()
	bucket := model.Bucket.ValueString()

	auth := yqcommon.ParseServiceIDToIAMAuth(serviceAccountID)
	return &Ydb_FederatedQuery.ConnectionSetting{
		Connection: &Ydb_FederatedQuery.ConnectionSetting_ObjectStorage{
			ObjectStorage: &Ydb_FederatedQuery.ObjectStorageConnection{
				Bucket: bucket,
				Auth:   auth,
			},
		},
	}
}

func (r *objectStorageConnectionStrategy) PackToState(ctx context.Context, setting *Ydb_FederatedQuery.ConnectionSetting, state *tfsdk.State, diagnostics *diag.Diagnostics) {
	var model objectStorageConnectionModel
	objectStorage := setting.GetObjectStorage()
	if objectStorage == nil {
		diagnostics.AddError("unexpected null ObjectStorage content setting from server", "")
		return
	}
	model.Bucket = types.StringValue(objectStorage.GetBucket())
	serviceAccountId, err := yqcommon.IAMAuthToString(objectStorage.GetAuth())
	if err != nil {
		diagnostics.AddError("Failed to extract auth info from connection", err.Error())
		return
	}
	model.ServiceAccountID = types.StringValue(serviceAccountId)

	diagnostics.Append(state.Set(ctx, &model)...)
}

func newObjectStorageConnectionStrategy() yqcommon.ConnectionStrategy {
	return &objectStorageConnectionStrategy{}
}

func newObjectStorageConnectionResourceSchema() map[string]schema.Attribute {
	return yqcommon.NewConnectionResourceSchema(yqcommon.AttributeBucket)
}

func NewResource() resource.Resource {
	return yqcommon.NewBaseConnectionResource(
		newObjectStorageConnectionResourceSchema(),
		newObjectStorageConnectionStrategy(),
		"_yq_object_storage_connection",
		"Manages Object Storage connection in Yandex Query service. For more information, see [the official documentation](https://yandex.cloud/docs/query/concepts/glossary#connection).\n\n")
}
