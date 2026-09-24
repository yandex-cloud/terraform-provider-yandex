package yandex_cloudregistry_registry

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	cloudregistry "github.com/yandex-cloud/go-genproto/yandex/cloud/cloudregistry/v1"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

func afterReadRegistry(
	ctx context.Context,
	providerConfig *provider_config.Config,
	res *cloudregistry.Registry,
	model *yandexCloudregistryRegistryModel,
	prior *yandexCloudregistryRegistryModel,
) diag.Diagnostics {
	var diags diag.Diagnostics

	if prior == nil || prior.Properties.IsNull() || prior.Properties.IsUnknown() {
		return diags
	}

	priorProperties := make(map[string]types.String, len(prior.Properties.Elements()))
	diags.Append(prior.Properties.ElementsAs(ctx, &priorProperties, false)...)
	if diags.HasError() {
		return diags
	}

	restored := make(map[string]attr.Value, len(model.Properties.Elements())+len(priorProperties))
	if !model.Properties.IsNull() && !model.Properties.IsUnknown() {
		for k, v := range model.Properties.Elements() {
			restored[k] = v
		}
	}

	responseProperties := res.GetProperties()
	for k, v := range priorProperties {
		if _, ok := responseProperties[k]; ok {
			continue
		}
		tflog.Debug(ctx, "registry property missing from GetRegistry response, restoring it from prior state: "+k)
		restored[k] = v
	}

	if len(restored) == 0 {
		return diags
	}

	properties, d := types.MapValue(types.StringType, restored)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	model.Properties = properties

	return diags
}
