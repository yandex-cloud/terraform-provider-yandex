package validate

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_opensearch_cluster/model"
)

var _ validator.Int64 = (*diskSizeLimitGreaterEffectiveDisk)(nil)

// diskSizeLimitGreaterEffectiveDisk validates that the configured autoscaling limit is greater
// than the effective baseline disk size taken from sibling resources.disk_size /
// resources.disk_size_gb. The input attribute value is interpreted as bytes when inGigabytes is
// false and as GiB when inGigabytes is true.
type diskSizeLimitGreaterEffectiveDisk struct {
	inGigabytes bool
}

// DiskSizeLimitGreaterThanEffectiveDisk validates that disk_size_limit (bytes) is greater than
// the effective baseline disk size from resources.disk_size or resources.disk_size_gb.
func DiskSizeLimitGreaterThanEffectiveDisk() validator.Int64 {
	return &diskSizeLimitGreaterEffectiveDisk{inGigabytes: false}
}

// DiskSizeGbLimitGreaterThanEffectiveDisk validates that disk_size_gb_limit (GiB) is greater
// than the effective baseline disk size from resources.disk_size or resources.disk_size_gb.
func DiskSizeGbLimitGreaterThanEffectiveDisk() validator.Int64 {
	return &diskSizeLimitGreaterEffectiveDisk{inGigabytes: true}
}

func (v *diskSizeLimitGreaterEffectiveDisk) Description(_ context.Context) string {
	return "If set, must be greater than the node group initial disk size (from resources.disk_size or resources.disk_size_gb)."
}

func (v *diskSizeLimitGreaterEffectiveDisk) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v *diskSizeLimitGreaterEffectiveDisk) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	limit := req.ConfigValue.ValueInt64()
	if v.inGigabytes {
		limit = datasize.ToBytes(limit)
	}

	expressions := req.PathExpression.MergeExpressions(
		path.MatchRelative().AtParent().AtParent().AtName("resources"),
	)

	for _, expression := range expressions {
		matchedPaths, diags := req.Config.PathMatches(ctx, expression)
		resp.Diagnostics.Append(diags...)
		if diags.HasError() {
			continue
		}

		for _, matchedPath := range matchedPaths {
			var matchedPathValue attr.Value
			diags := req.Config.GetAttribute(ctx, matchedPath, &matchedPathValue)
			resp.Diagnostics.Append(diags...)
			if diags.HasError() {
				continue
			}

			if matchedPathValue.IsNull() || matchedPathValue.IsUnknown() {
				continue
			}

			obj, ok := matchedPathValue.(types.Object)
			if !ok {
				continue
			}

			var nr model.NodeResource
			asDiags := obj.As(ctx, &nr, basetypes.ObjectAsOptions{
				UnhandledNullAsEmpty:    false,
				UnhandledUnknownAsEmpty: false,
			})
			resp.Diagnostics.Append(asDiags...)
			if asDiags.HasError() {
				continue
			}

			baseline, d := model.EffectiveDiskSizeBytes(&nr)
			if d.HasError() {
				continue
			}

			if baseline >= limit {
				resp.Diagnostics.AddAttributeError(
					req.Path,
					"Invalid Attribute Value",
					fmt.Sprintf("%s must be greater than the initial disk size (%d bytes).", req.Path, baseline),
				)
				return
			}
		}
	}
}
