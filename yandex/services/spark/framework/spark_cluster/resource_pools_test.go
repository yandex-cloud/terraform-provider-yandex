package spark_cluster

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/spark/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
)

func TestExtractPoolsDecodesDriverWithoutPreemptible(t *testing.T) {
	ctx := context.Background()
	driver := types.ObjectValueMust(DriverPoolAttrTypes, map[string]attr.Value{
		"resource_preset_id": types.StringValue("c4-m16"),
		"size":               types.Int64Value(1),
		"min_size":           types.Int64Null(),
		"max_size":           types.Int64Null(),
	})
	executor := types.ObjectValueMust(ExecutorPoolAttrTypes, map[string]attr.Value{
		"resource_preset_id": types.StringValue("c4-m16"),
		"size":               types.Int64Null(),
		"min_size":           types.Int64Value(1),
		"max_size":           types.Int64Value(2),
		"preemptible":        types.BoolValue(true),
	})
	resourcePools := types.ObjectValueMust(ResourcePoosAttrTypes, map[string]attr.Value{
		"driver":   driver,
		"executor": executor,
	})

	var diags diag.Diagnostics
	driverPool, executorPool := extractPools(ctx, &ClusterModel{
		Config: ConfigValue{ResourcePools: resourcePools},
	}, &diags)
	if diags.HasError() {
		t.Fatalf("extractPools returned diagnostics: %v", diags.Errors())
	}
	if !driverPool.ResourcePresetId.Equal(types.StringValue("c4-m16")) || !driverPool.Size.Equal(types.Int64Value(1)) {
		t.Fatalf("unexpected driver pool: %#v", driverPool)
	}
	if !executorPool.Preemptible.Equal(types.BoolValue(true)) {
		t.Fatalf("unexpected executor preemptible value: %#v", executorPool.Preemptible)
	}
}

func TestResourcePoolsFromAPIIncludesExecutorPreemptible(t *testing.T) {
	ctx := context.Background()
	resourcePools, diags := resourcePoolsFromAPI(ctx, &spark.ResourcePools{
		Driver: &spark.ResourcePool{
			ResourcePresetId: "c4-m16",
			ScalePolicy: &spark.ScalePolicy{ScaleType: &spark.ScalePolicy_FixedScale_{
				FixedScale: &spark.ScalePolicy_FixedScale{Size: 1},
			}},
		},
		Executor: &spark.ResourcePool{
			ResourcePresetId: "c4-m16",
			Preemptible:      true,
			ScalePolicy: &spark.ScalePolicy{ScaleType: &spark.ScalePolicy_AutoScale_{
				AutoScale: &spark.ScalePolicy_AutoScale{MinSize: 1, MaxSize: 2},
			}},
		},
	})
	if diags.HasError() {
		t.Fatalf("resourcePoolsFromAPI returned diagnostics: %v", diags.Errors())
	}

	var pools ResourcePools
	diags.Append(resourcePools.As(ctx, &pools, datasize.DefaultOpts)...)
	var driver DriverResourcePool
	diags.Append(pools.Driver.As(ctx, &driver, datasize.DefaultOpts)...)
	var executor ResourcePool
	diags.Append(pools.Executor.As(ctx, &executor, datasize.DefaultOpts)...)
	if diags.HasError() {
		t.Fatalf("decoded resource pools returned diagnostics: %v", diags.Errors())
	}
	if !driver.Size.Equal(types.Int64Value(1)) {
		t.Fatalf("unexpected driver size: %#v", driver.Size)
	}
	if !executor.Preemptible.Equal(types.BoolValue(true)) {
		t.Fatalf("unexpected executor preemptible value: %#v", executor.Preemptible)
	}
}

func TestBuildUpdateClusterRequestUpdatesOnlyExecutorPreemptible(t *testing.T) {
	ctx := context.Background()
	state := testClusterModel(ctx, false)
	plan := testClusterModel(ctx, true)

	request, diags := BuildUpdateClusterRequest(ctx, state, plan)
	if diags.HasError() {
		t.Fatalf("BuildUpdateClusterRequest returned diagnostics: %v", diags.Errors())
	}

	expectedMask := []string{"config_spec.resource_pools.executor.preemptible"}
	if !reflect.DeepEqual(request.UpdateMask.Paths, expectedMask) {
		t.Fatalf("unexpected update mask: %#v", request.UpdateMask.Paths)
	}
	if !request.ConfigSpec.ResourcePools.Executor.Preemptible {
		t.Fatal("executor preemptible value was not propagated to update request")
	}
}

func testClusterModel(ctx context.Context, preemptible bool) *ClusterModel {
	resourcePools := types.ObjectValueMust(ResourcePoosAttrTypes, map[string]attr.Value{
		"driver": types.ObjectValueMust(DriverPoolAttrTypes, map[string]attr.Value{
			"resource_preset_id": types.StringValue("c4-m16"),
			"size":               types.Int64Value(1),
			"min_size":           types.Int64Null(),
			"max_size":           types.Int64Null(),
		}),
		"executor": types.ObjectValueMust(ExecutorPoolAttrTypes, map[string]attr.Value{
			"resource_preset_id": types.StringValue("c4-m16"),
			"size":               types.Int64Null(),
			"min_size":           types.Int64Value(1),
			"max_size":           types.Int64Value(2),
			"preemptible":        types.BoolValue(preemptible),
		}),
	})
	dependencies := types.ObjectValueMust(DependenciesAttrTypes, map[string]attr.Value{
		"pip_packages": types.SetValueMust(types.StringType, []attr.Value{}),
		"deb_packages": types.SetValueMust(types.StringType, []attr.Value{}),
	})
	historyServer := types.ObjectValueMust(HistoryServerAttrTypes, map[string]attr.Value{
		"enabled": types.BoolValue(true),
	})
	metastore := types.ObjectValueMust(MetastoreAttrTypes, map[string]attr.Value{
		"cluster_id": types.StringValue(""),
	})

	return &ClusterModel{
		Id:                 types.StringValue("cluster-id"),
		Name:               types.StringValue("cluster-name"),
		Description:        types.StringValue(""),
		Labels:             types.MapValueMust(types.StringType, map[string]attr.Value{}),
		DeletionProtection: types.BoolValue(false),
		ServiceAccountId:   types.StringValue(""),
		Config: NewConfigValueMust(ConfigValue{}.AttributeTypes(ctx), map[string]attr.Value{
			"resource_pools": resourcePools,
			"dependencies":   dependencies,
			"history_server": historyServer,
			"metastore":      metastore,
			"spark_version":  types.StringValue(""),
		}),
		Network: NewNetworkValueMust(NetworkValue{}.AttributeTypes(ctx), map[string]attr.Value{
			"security_group_ids": types.SetValueMust(types.StringType, []attr.Value{}),
			"subnet_ids":         types.SetValueMust(types.StringType, []attr.Value{}),
		}),
		Logging:           NewLoggingValueNull(),
		MaintenanceWindow: NewMaintenanceWindowValueNull(),
	}
}
