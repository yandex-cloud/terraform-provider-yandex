package cdn_rule

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/cdn/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/cdn_resource"
)

// flattenOptions converts API options to Terraform options
// Reuses cdn_resource.FlattenCDNResourceOptions since cdn_rule uses the same options structure
func flattenOptions(ctx context.Context, options *cdn.ResourceOptions, diags *diag.Diagnostics) types.List {
	tflog.Debug(ctx, "Flattening CDN rule options using cdn_resource.FlattenCDNResourceOptions")

	// Delegate to cdn_resource flatten function - options structure is identical
	// Pass untyped null for planOptions since cdn_rule doesn't need disabled block preservation
	// (that feature is specific to cdn_resource where caches are enabled by default)
	return cdn_resource.FlattenCDNResourceOptions(ctx, options, types.List{}, diags)
}
