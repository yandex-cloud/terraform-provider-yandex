package mdb_greenplum_resource_group

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/greenplum/v1"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type ResourceGroup struct {
	Id        types.String `tfsdk:"id"`
	ClusterID types.String `tfsdk:"cluster_id"`
	Name      types.String `tfsdk:"name"`

	IsUserDefined types.Bool `tfsdk:"is_user_defined"`

	Concurrency       types.Int64    `tfsdk:"concurrency"`
	CpuRateLimit      types.Int64    `tfsdk:"cpu_rate_limit"`
	MemoryLimit       types.Int64    `tfsdk:"memory_limit"`
	MemorySharedQuota types.Int64    `tfsdk:"memory_shared_quota"`
	MemorySpillRatio  types.Int64    `tfsdk:"memory_spill_ratio"`
	Timeouts          timeouts.Value `tfsdk:"timeouts"`
}

func resourceGroupToState(resourceGroup *greenplum.ResourceGroup, state *ResourceGroup) {
	state.Name = types.StringValue(resourceGroup.Name)

	state.IsUserDefined = types.BoolValue(resourceGroup.IsUserDefined.GetValue())

	if resourceGroup.GetConcurrency() != nil {
		state.Concurrency = types.Int64Value(resourceGroup.Concurrency.GetValue())
	}
	if resourceGroup.GetCpuRateLimit() != nil {
		state.CpuRateLimit = types.Int64Value(resourceGroup.CpuRateLimit.GetValue())
	}
	if resourceGroup.GetMemoryLimit() != nil {
		state.MemoryLimit = types.Int64Value(resourceGroup.MemoryLimit.GetValue())
	}
	if resourceGroup.GetMemorySharedQuota() != nil {
		state.MemorySharedQuota = types.Int64Value(resourceGroup.MemorySharedQuota.GetValue())
	}
	if resourceGroup.GetMemorySpillRatio() != nil {
		state.MemorySpillRatio = types.Int64Value(resourceGroup.MemorySpillRatio.GetValue())
	}
}

func resourceGroupFromState(ctx context.Context, state *ResourceGroup) *greenplum.ResourceGroup {
	rg := &greenplum.ResourceGroup{
		Name: state.Name.ValueString(),
	}

	if !state.IsUserDefined.IsUnknown() && !state.IsUserDefined.IsNull() {
		rg.IsUserDefined = wrapperspb.Bool(state.IsUserDefined.ValueBool())
	}

	if !state.Concurrency.IsUnknown() && !state.Concurrency.IsNull() {
		rg.Concurrency = wrapperspb.Int64(state.Concurrency.ValueInt64())
	}
	if !state.CpuRateLimit.IsUnknown() && !state.CpuRateLimit.IsNull() {
		rg.CpuRateLimit = wrapperspb.Int64(state.CpuRateLimit.ValueInt64())
	}
	if !state.MemoryLimit.IsUnknown() && !state.MemoryLimit.IsNull() {
		rg.MemoryLimit = wrapperspb.Int64(state.MemoryLimit.ValueInt64())
	}
	if !state.MemorySharedQuota.IsUnknown() && !state.MemorySharedQuota.IsNull() {
		rg.MemorySharedQuota = wrapperspb.Int64(state.MemorySharedQuota.ValueInt64())
	}
	if !state.MemorySpillRatio.IsUnknown() && !state.MemorySpillRatio.IsNull() {
		rg.MemorySpillRatio = wrapperspb.Int64(state.MemorySpillRatio.ValueInt64())
	}

	return rg
}
