package cdn_rule

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/cdn/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/cdn_resource"
)

// expandOptions converts Terraform options to API ResourceOptions
// Reuses cdn_resource.ExpandCDNResourceOptions since cdn_rule uses the same options structure
func expandOptions(ctx context.Context, model *CDNRuleModel, diags *diag.Diagnostics) *cdn.ResourceOptions {
	if model.Options.IsNull() || len(model.Options.Elements()) == 0 {
		tflog.Debug(ctx, "No options specified for CDN rule")
		return nil
	}

	var optionsModels []cdn_resource.CDNOptionsModel
	diags.Append(model.Options.ElementsAs(ctx, &optionsModels, false)...)
	if diags.HasError() || len(optionsModels) == 0 {
		return nil
	}

	tflog.Debug(ctx, "Expanding CDN rule options using cdn_resource.ExpandCDNResourceOptions")

	// Delegate to cdn_resource.ExpandCDNResourceOptions - same structure, same logic
	return cdn_resource.ExpandCDNResourceOptions(ctx, optionsModels, diags)
}
