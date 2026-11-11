package yandex_container_registry

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func moveStateFromV0(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	var priorStateData yandexContainerRegistryModel
	resp.Diagnostics.Append(req.State.Get(ctx, &priorStateData)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, priorStateData)...)
}
