package cdn_origin_group

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/cdn/v1"
)

// expandOriginParams converts OriginModel to API OriginParams
func expandOriginParams(ctx context.Context, origin *OriginModel) *cdn.OriginParams {
	if origin == nil {
		return nil
	}

	params := &cdn.OriginParams{
		Source:  origin.Source.ValueString(),
		Enabled: origin.Enabled.ValueBool(),
		Backup:  origin.Backup.ValueBool(),
	}

	tflog.Debug(ctx, "Expanded origin params", map[string]interface{}{
		"source":  params.Source,
		"enabled": params.Enabled,
		"backup":  params.Backup,
	})

	return params
}

// expandOrigins converts a set of OriginModel to API OriginParams slice
func expandOrigins(ctx context.Context, originsSet []OriginModel, diags *diag.Diagnostics) []*cdn.OriginParams {
	if len(originsSet) == 0 {
		return nil
	}

	origins := make([]*cdn.OriginParams, 0, len(originsSet))
	for i := range originsSet {
		origins = append(origins, expandOriginParams(ctx, &originsSet[i]))
	}

	tflog.Debug(ctx, "Expanded origins", map[string]interface{}{
		"count": len(origins),
	})

	return origins
}
