package customplanmodifiers

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func MarkCacheSizePlanModifier() planmodifier.Int64 {
	return &markCacheSizeModifier{}
}

type markCacheSizeModifier struct{}

func (m *markCacheSizeModifier) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	if !req.ConfigValue.IsNull() && !req.ConfigValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}

	if m.shardsChanged(ctx, req, resp) || m.clickhouseResourcesChanged(ctx, req, resp) {
		resp.PlanValue = types.Int64Unknown()
	}
}

func (m *markCacheSizeModifier) shardsChanged(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) bool {
	var configShards types.Map
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("shards"), &configShards)...)
	if resp.Diagnostics.HasError() || configShards.IsNull() || configShards.IsUnknown() {
		return false
	}

	var stateShards types.Map
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("shards"), &stateShards)...)
	if resp.Diagnostics.HasError() {
		return false
	}

	configElems := configShards.Elements()
	stateElems := stateShards.Elements()
	if len(configElems) != len(stateElems) {
		return true
	}
	for k := range configElems {
		if _, ok := stateElems[k]; !ok {
			return true
		}
	}

	for k := range configElems {
		cfgPreset := m.extractShardPreset(configElems[k])
		if cfgPreset.IsNull() || cfgPreset.IsUnknown() {
			continue
		}
		stPreset := m.extractShardPreset(stateElems[k])
		if !cfgPreset.Equal(stPreset) {
			return true
		}
	}
	return false
}

func (m *markCacheSizeModifier) clickhouseResourcesChanged(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) bool {
	var configResources types.Object
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("clickhouse").AtName("resources"), &configResources)...)
	if resp.Diagnostics.HasError() {
		return false
	}

	var stateResources types.Object
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("clickhouse").AtName("resources"), &stateResources)...)
	if resp.Diagnostics.HasError() {
		return false
	}

	configPreset := m.extractResourcePreset(configResources)
	if configPreset.IsNull() || configPreset.IsUnknown() {
		return false
	}
	statePreset := m.extractResourcePreset(stateResources)
	return !configPreset.Equal(statePreset)
}

func (m *markCacheSizeModifier) extractShardPreset(v attr.Value) types.String {
	obj, ok := v.(types.Object)
	if !ok || obj.IsNull() || obj.IsUnknown() {
		return types.StringNull()
	}
	resources, ok := obj.Attributes()["resources"].(types.Object)
	if !ok || resources.IsNull() || resources.IsUnknown() {
		return types.StringNull()
	}
	preset, ok := resources.Attributes()["resource_preset_id"].(types.String)
	if !ok {
		return types.StringNull()
	}
	return preset
}

func (m *markCacheSizeModifier) extractResourcePreset(resources types.Object) types.String {
	if resources.IsNull() || resources.IsUnknown() {
		return types.StringNull()
	}
	preset, ok := resources.Attributes()["resource_preset_id"].(types.String)
	if !ok {
		return types.StringNull()
	}
	return preset
}

func (m *markCacheSizeModifier) Description(context.Context) string {
	return "Marks mark_cache_size as unknown when shard set or clickhouse resources change, since the backend recalculates it based on host RAM."
}

func (m *markCacheSizeModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}
