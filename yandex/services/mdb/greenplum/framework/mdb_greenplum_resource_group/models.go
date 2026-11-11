package mdb_greenplum_resource_group

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/greenplum/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/converter"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
)

type ResourceGroup struct {
	Id        types.String `tfsdk:"id"`
	ClusterID types.String `tfsdk:"cluster_id"`
	Name      types.String `tfsdk:"name"`

	IsUserDefined types.Bool `tfsdk:"is_user_defined"`

	Concurrency       types.Int64 `tfsdk:"concurrency"`
	CpuRateLimit      types.Int64 `tfsdk:"cpu_rate_limit"`
	MemoryLimit       types.Int64 `tfsdk:"memory_limit"`
	MemorySharedQuota types.Int64 `tfsdk:"memory_shared_quota"`
	MemorySpillRatio  types.Int64 `tfsdk:"memory_spill_ratio"`

	CpuMaxPercent types.Int64 `tfsdk:"cpu_max_percent"`
	CpuWeight     types.Int64 `tfsdk:"cpu_weight"`
	MemoryQuota   types.Int64 `tfsdk:"memory_quota"`
	MinCost       types.Int64 `tfsdk:"min_cost"`

	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func resourceGroupToState(ctx context.Context, resourceGroup *greenplum.ResourceGroup, state *ResourceGroup, diags *diag.Diagnostics) {
	state.Name = types.StringValue(resourceGroup.Name)

	state.IsUserDefined = types.BoolValue(resourceGroup.IsUserDefined.GetValue())

	state.Concurrency = mdbcommon.FlattenInt64Wrapper(ctx, resourceGroup.GetConcurrency(), diags)
	state.CpuRateLimit = mdbcommon.FlattenInt64Wrapper(ctx, resourceGroup.GetCpuRateLimit(), diags)
	state.MemoryLimit = mdbcommon.FlattenInt64Wrapper(ctx, resourceGroup.GetMemoryLimit(), diags)
	state.MemorySharedQuota = mdbcommon.FlattenInt64Wrapper(ctx, resourceGroup.GetMemorySharedQuota(), diags)
	state.MemorySpillRatio = mdbcommon.FlattenInt64Wrapper(ctx, resourceGroup.GetMemorySpillRatio(), diags)
	state.CpuMaxPercent = mdbcommon.FlattenInt64Wrapper(ctx, resourceGroup.GetCpuMaxPercent(), diags)
	state.CpuWeight = mdbcommon.FlattenInt64Wrapper(ctx, resourceGroup.GetCpuWeight(), diags)
	state.MemoryQuota = mdbcommon.FlattenInt64Wrapper(ctx, resourceGroup.GetMemoryQuota(), diags)
	state.MinCost = mdbcommon.FlattenInt64Wrapper(ctx, resourceGroup.GetMinCost(), diags)
}

func resourceGroupFromState(state *ResourceGroup) *greenplum.ResourceGroup {
	return &greenplum.ResourceGroup{
		Name:              state.Name.ValueString(),
		IsUserDefined:     converter.WrappedBool(state.IsUserDefined),
		Concurrency:       converter.WrappedInt64(state.Concurrency),
		CpuRateLimit:      converter.WrappedInt64(state.CpuRateLimit),
		MemoryLimit:       converter.WrappedInt64(state.MemoryLimit),
		MemorySharedQuota: converter.WrappedInt64(state.MemorySharedQuota),
		MemorySpillRatio:  converter.WrappedInt64(state.MemorySpillRatio),
		CpuMaxPercent:     converter.WrappedInt64(state.CpuMaxPercent),
		CpuWeight:         converter.WrappedInt64(state.CpuWeight),
		MemoryQuota:       converter.WrappedInt64(state.MemoryQuota),
		MinCost:           converter.WrappedInt64(state.MinCost),
	}
}
