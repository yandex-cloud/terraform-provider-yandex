package cdn_origin_group

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/cdn/v1"
)

// flattenOrigins converts API Origins to OriginModel slice
func flattenOrigins(ctx context.Context, origins []*cdn.Origin, diags *diag.Diagnostics) types.Set {
	if len(origins) == 0 {
		return types.SetNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"source":          types.StringType,
				"origin_group_id": types.Int64Type,
				"enabled":         types.BoolType,
				"backup":          types.BoolType,
			},
		})
	}

	originModels := make([]OriginModel, 0, len(origins))
	for _, origin := range origins {
		model := OriginModel{
			Source:        types.StringValue(origin.Source),
			OriginGroupID: types.Int64Value(origin.OriginGroupId),
			Enabled:       types.BoolValue(origin.Enabled),
			Backup:        types.BoolValue(origin.Backup),
		}
		originModels = append(originModels, model)
	}

	setVal, d := types.SetValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"source":          types.StringType,
			"origin_group_id": types.Int64Type,
			"enabled":         types.BoolType,
			"backup":          types.BoolType,
		},
	}, originModels)
	diags.Append(d...)

	tflog.Debug(ctx, "Flattened origins", map[string]interface{}{
		"count": len(originModels),
	})

	return setVal
}

// flattenCDNOriginGroup converts API OriginGroup to CDNOriginGroupModel
func flattenCDNOriginGroup(ctx context.Context, state *CDNOriginGroupModel, originGroup *cdn.OriginGroup, diags *diag.Diagnostics) {
	state.ID = types.StringValue(strconv.FormatInt(originGroup.Id, 10))
	state.FolderID = types.StringValue(originGroup.FolderId)
	state.Name = types.StringValue(originGroup.Name)
	state.ProviderType = types.StringValue(originGroup.ProviderType)
	state.UseNext = types.BoolValue(originGroup.UseNext)
	state.Origins = flattenOrigins(ctx, originGroup.Origins, diags)

	tflog.Debug(ctx, "Flattened CDN origin group", map[string]interface{}{
		"id":   state.ID.ValueString(),
		"name": state.Name.ValueString(),
	})
}
