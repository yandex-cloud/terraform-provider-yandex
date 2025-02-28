package mdbcommon

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/objectid"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

// GetClusterIdForDatasource retrieves the cluster ID for a given datasource.
// It accepts either a direct cluster ID or resolves it using the name attribute.
func GetClusterIdForDatasource(ctx context.Context, providerConfig *provider_config.Config, config tfsdk.Config) (string, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	var clusterId types.String
	var name types.String
	diags.Append(config.GetAttribute(ctx, path.Root("cluster_id"), &clusterId)...)
	diags.Append(config.GetAttribute(ctx, path.Root("name"), &name)...)
	if diags.HasError() {
		return "", diags
	}

	if clusterId.ValueString() == "" && name.ValueString() == "" {
		diags.AddError(
			"At least one of cluster_id or name is required",
			"The cluster ID or Name must be specified in the configuration",
		)
		return "", diags
	}

	clusterIdStr := clusterId.ValueString()
	if clusterIdStr == "" {
		folderID, d := validate.FolderID(types.StringUnknown(), &providerConfig.ProviderState)
		if diags.Append(d); diags.HasError() {
			return "", diags
		}

		clusterIdStr, d = objectid.ResolveByNameAndFolderID(ctx, providerConfig.SDK, folderID, name.ValueString(), sdkresolvers.RedisClusterResolver)
		if diags.Append(d); diags.HasError() {
			return "", diags
		}
	}
	return clusterIdStr, nil
}
