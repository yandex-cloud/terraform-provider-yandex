package mdb_clickhouse_cluster_v2

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_clickhouse_cluster_v2/models"
)

type coordinatorHostTypes struct {
	hasZooKeeper bool
	hasKeeper    bool
	hasUnknown   bool
}

func detectKeeperMigration(ctx context.Context, stateHosts, planHosts types.Map, diags *diag.Diagnostics) bool {
	stateTypes := getCoordinatorHostTypes(ctx, stateHosts, diags)
	planTypes := getCoordinatorHostTypes(ctx, planHosts, diags)
	if diags.HasError() || stateTypes.hasUnknown || planTypes.hasUnknown {
		return false
	}

	if stateTypes.hasZooKeeper && stateTypes.hasKeeper {
		diags.AddError(
			"Invalid coordinator state",
			"The ClickHouse cluster contains both ZooKeeper and Keeper hosts. Migration cannot be planned for a mixed coordinator state.",
		)
		return false
	}

	if planTypes.hasZooKeeper && planTypes.hasKeeper {
		diags.AddError(
			"Invalid coordinator host transition",
			"ZooKeeper and Keeper hosts cannot be configured at the same time. Change all coordinator hosts from ZOOKEEPER to KEEPER in one apply.",
		)
		return false
	}

	if stateTypes.hasKeeper && planTypes.hasZooKeeper {
		diags.AddError(
			"Unsupported coordinator host transition",
			"Migration from ClickHouse Keeper back to ZooKeeper is not supported.",
		)
		return false
	}

	return stateTypes.hasZooKeeper && planTypes.hasKeeper
}

func getCoordinatorHostTypes(ctx context.Context, hosts types.Map, diags *diag.Diagnostics) coordinatorHostTypes {
	result := coordinatorHostTypes{}
	if hosts.IsNull() || hosts.IsUnknown() {
		return result
	}

	var hostMap map[string]models.Host
	diags.Append(hosts.ElementsAs(ctx, &hostMap, false)...)
	if diags.HasError() {
		return result
	}

	for _, host := range hostMap {
		if host.Type.IsUnknown() {
			result.hasUnknown = true
			continue
		}

		switch host.Type.ValueString() {
		case clickhouse.Host_ZOOKEEPER.String():
			result.hasZooKeeper = true
		case clickhouse.Host_KEEPER.String():
			result.hasKeeper = true
		}
	}

	return result
}

func markKeeperFQDNsUnknown(ctx context.Context, hosts types.Map, diags *diag.Diagnostics) types.Map {
	if hosts.IsNull() || hosts.IsUnknown() {
		return hosts
	}

	var hostMap map[string]models.Host
	diags.Append(hosts.ElementsAs(ctx, &hostMap, false)...)
	if diags.HasError() {
		return hosts
	}

	for label, host := range hostMap {
		if host.Type.ValueString() == clickhouse.Host_KEEPER.String() {
			host.FQDN = types.StringUnknown()
			hostMap[label] = host
		}
	}

	result, d := types.MapValueFrom(ctx, models.HostType, hostMap)
	diags.Append(d...)
	if diags.HasError() {
		return hosts
	}

	return result
}

func prepareMigrateToKeeperRequest(
	ctx context.Context,
	clusterID string,
	keeperHosts types.Map,
	resources *clickhouse.Resources,
	allowDegradationToReadOnly bool,
	diags *diag.Diagnostics,
) *clickhouse.MigrateClusterToKeeperRequest {
	hostSpecs, d := mdbcommon.CreateClusterHosts(ctx, clickhouseHostService, keeperHosts)
	diags.Append(d...)
	if diags.HasError() {
		return nil
	}

	return &clickhouse.MigrateClusterToKeeperRequest{
		ClusterId:                  clusterID,
		Resources:                  resources,
		HostSpecs:                  hostSpecs,
		AllowDegradationToReadOnly: allowDegradationToReadOnly,
	}
}

func getAPICoordinatorHostTypes(hosts []*clickhouse.Host) coordinatorHostTypes {
	result := coordinatorHostTypes{}
	for _, host := range hosts {
		switch host.GetType() {
		case clickhouse.Host_ZOOKEEPER:
			result.hasZooKeeper = true
		case clickhouse.Host_KEEPER:
			result.hasKeeper = true
		}
	}

	return result
}
