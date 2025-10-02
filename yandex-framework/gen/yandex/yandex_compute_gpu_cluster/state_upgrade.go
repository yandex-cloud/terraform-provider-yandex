package yandex_compute_gpu_cluster

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func moveStateFromV0(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	var priorStateData yandexComputeGpuClusterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &priorStateData)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, priorStateData)...)
}
