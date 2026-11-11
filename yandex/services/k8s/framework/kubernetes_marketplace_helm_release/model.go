package kubernetes_marketplace_helm_release

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	marketplace "github.com/yandex-cloud/go-genproto/yandex/cloud/k8s/marketplace/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/timestamp"
)

// helmReleaseResourceModel maps the resource schema data.
type helmReleaseResourceModel struct {
	ID               types.String   `tfsdk:"id"`
	ClusterID        types.String   `tfsdk:"cluster_id"`
	Name             types.String   `tfsdk:"name"`
	Namespace        types.String   `tfsdk:"namespace"`
	ProductID        types.String   `tfsdk:"product_id"`
	ProductName      types.String   `tfsdk:"product_name"`
	ProductVersionID types.String   `tfsdk:"product_version"`
	Status           types.String   `tfsdk:"status"`
	CreatedAt        types.String   `tfsdk:"created_at"`
	UserValues       types.Map      `tfsdk:"user_values"`
	Timeouts         timeouts.Value `tfsdk:"timeouts"`
}

func helmReleaseToModel(ctx context.Context, helmRelease *marketplace.HelmRelease, model *helmReleaseResourceModel) diag.Diagnostics {
	tflog.Debug(ctx, fmt.Sprintf("helmReleaseToState: Helm Release state: %+v", model))
	tflog.Debug(ctx, fmt.Sprintf("helmReleaseToState: Received Helm Release data: %+v", helmRelease))

	model.CreatedAt = types.StringValue(timestamp.Get(helmRelease.GetCreatedAt()))
	model.ClusterID = types.StringValue(helmRelease.GetClusterId())
	model.Name = types.StringValue(helmRelease.GetAppName())
	model.Namespace = types.StringValue(helmRelease.GetAppNamespace())
	model.ProductID = types.StringValue(helmRelease.GetProductId())
	model.ProductVersionID = types.StringValue(helmRelease.GetProductVersion())
	model.Status = types.StringValue(helmRelease.GetStatus().String())

	// Do not override model.ProductName if it is null and server empty string,
	// otherwise we will get "Provider produced inconsistent result after
	// apply". Same principle applies to other attributes that can be null or
	// empty.
	newProductName := types.StringValue(helmRelease.GetProductName())
	if stringsAreEqual(newProductName, model.ProductName) {
		model.ProductName = newProductName
	}

	return nil
}

func stringsAreEqual(str1, str2 types.String) bool {
	if str1.Equal(str2) {
		return true
	}
	// if one of strings is null and the other is empty then we assume that they are equal
	if str1.ValueString() == "" && str2.ValueString() == "" {
		return true
	}
	return false
}
