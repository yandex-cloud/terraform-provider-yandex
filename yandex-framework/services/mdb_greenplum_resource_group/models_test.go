package mdb_greenplum_resource_group

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/greenplum/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type resourceGroupCase struct {
	name        string
	model       ResourceGroup
	api         *greenplum.ResourceGroup
	updatePaths []string
}

func resourceGroupCases() []resourceGroupCase {
	return []resourceGroupCase{
		{
			name: "Greenplum",
			model: ResourceGroup{
				Name:              types.StringValue("greenplum-group"),
				IsUserDefined:     types.BoolValue(true),
				Concurrency:       types.Int64Value(0),
				CpuRateLimit:      types.Int64Value(10),
				MemoryLimit:       types.Int64Value(0),
				MemorySharedQuota: types.Int64Value(30),
				MemorySpillRatio:  types.Int64Value(40),
				CpuMaxPercent:     types.Int64Null(),
				CpuWeight:         types.Int64Null(),
				MemoryQuota:       types.Int64Null(),
				MinCost:           types.Int64Null(),
			},
			api: &greenplum.ResourceGroup{
				Name:              "greenplum-group",
				IsUserDefined:     wrapperspb.Bool(true),
				Concurrency:       wrapperspb.Int64(0),
				CpuRateLimit:      wrapperspb.Int64(10),
				MemoryLimit:       wrapperspb.Int64(0),
				MemorySharedQuota: wrapperspb.Int64(30),
				MemorySpillRatio:  wrapperspb.Int64(40),
			},
			updatePaths: []string{
				"resource_group.concurrency",
				"resource_group.cpu_rate_limit",
				"resource_group.memory_limit",
				"resource_group.memory_shared_quota",
				"resource_group.memory_spill_ratio",
			},
		},
		{
			name: "Cloudberry",
			model: ResourceGroup{
				Name:              types.StringValue("cloudberry-group"),
				IsUserDefined:     types.BoolValue(true),
				Concurrency:       types.Int64Value(8),
				CpuRateLimit:      types.Int64Null(),
				MemoryLimit:       types.Int64Null(),
				MemorySharedQuota: types.Int64Null(),
				MemorySpillRatio:  types.Int64Null(),
				CpuMaxPercent:     types.Int64Value(-1),
				CpuWeight:         types.Int64Value(200),
				MemoryQuota:       types.Int64Value(1024),
				MinCost:           types.Int64Value(25),
			},
			api: &greenplum.ResourceGroup{
				Name:          "cloudberry-group",
				IsUserDefined: wrapperspb.Bool(true),
				Concurrency:   wrapperspb.Int64(8),
				CpuMaxPercent: wrapperspb.Int64(-1),
				CpuWeight:     wrapperspb.Int64(200),
				MemoryQuota:   wrapperspb.Int64(1024),
				MinCost:       wrapperspb.Int64(25),
			},
			updatePaths: []string{
				"resource_group.concurrency",
				"resource_group.cpu_max_percent",
				"resource_group.cpu_weight",
				"resource_group.memory_quota",
				"resource_group.min_cost",
			},
		},
		{
			name: "no optional fields",
			model: ResourceGroup{
				Name:              types.StringValue("empty-group"),
				IsUserDefined:     types.BoolValue(false),
				Concurrency:       types.Int64Null(),
				CpuRateLimit:      types.Int64Null(),
				MemoryLimit:       types.Int64Null(),
				MemorySharedQuota: types.Int64Null(),
				MemorySpillRatio:  types.Int64Null(),
				CpuMaxPercent:     types.Int64Null(),
				CpuWeight:         types.Int64Null(),
				MemoryQuota:       types.Int64Null(),
				MinCost:           types.Int64Null(),
			},
			api: &greenplum.ResourceGroup{
				Name:          "empty-group",
				IsUserDefined: wrapperspb.Bool(false),
			},
		},
	}
}

func TestResourceGroupConversion(t *testing.T) {
	for _, tc := range resourceGroupCases() {
		t.Run(tc.name, func(t *testing.T) {
			gotAPI := resourceGroupFromState(&tc.model)
			if !proto.Equal(gotAPI, tc.api) {
				t.Fatalf("API model = %v, want %v", gotAPI, tc.api)
			}

			state := ResourceGroup{
				Name:              types.StringValue("stale"),
				IsUserDefined:     types.BoolValue(false),
				Concurrency:       types.Int64Value(99),
				CpuRateLimit:      types.Int64Value(99),
				MemoryLimit:       types.Int64Value(99),
				MemorySharedQuota: types.Int64Value(99),
				MemorySpillRatio:  types.Int64Value(99),
				CpuMaxPercent:     types.Int64Value(99),
				CpuWeight:         types.Int64Value(99),
				MemoryQuota:       types.Int64Value(99),
				MinCost:           types.Int64Value(99),
			}
			var diags diag.Diagnostics
			resourceGroupToState(context.Background(), tc.api, &state, &diags)
			if diags.HasError() {
				t.Fatalf("flatten diagnostics: %v", diags)
			}
			if !reflect.DeepEqual(state, tc.model) {
				t.Fatalf("state = %#v, want %#v", state, tc.model)
			}
		})
	}
}

func TestResourceGroupUpdatePaths(t *testing.T) {
	for _, tc := range resourceGroupCases() {
		t.Run(tc.name, func(t *testing.T) {
			state := &greenplum.ResourceGroup{IsUserDefined: tc.api.IsUserDefined}
			for _, test := range []struct {
				name  string
				plan  *greenplum.ResourceGroup
				state *greenplum.ResourceGroup
			}{
				{name: "add", plan: tc.api, state: state},
				{name: "clear", plan: state, state: tc.api},
				{name: "unchanged", plan: tc.api, state: tc.api},
			} {
				t.Run(test.name, func(t *testing.T) {
					paths, diags := getUpdatePaths(test.plan, test.state)
					want := tc.updatePaths
					if test.name == "unchanged" {
						want = nil
					}
					if diags.HasError() || !reflect.DeepEqual(paths, want) {
						t.Fatalf("paths = %v, want %v; diagnostics = %v", paths, want, diags)
					}
				})
			}
		})
	}
}
