package validators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_clickhouse_cluster_v2/models"
)

type ShardsHostsConsistencyValidator struct{}

func (v ShardsHostsConsistencyValidator) Description(_ context.Context) string {
	return "Every uncoordinator host shard_name must be in the shards block, and every shard must have at least one host."
}

func (v ShardsHostsConsistencyValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v ShardsHostsConsistencyValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config models.ClusterResource
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.Shards.IsNull() || config.Shards.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("shards"),
			"Shards not defined",
			"The cluster must have at least 1 shard.",
		)
		return
	}
	if config.HostSpecs.IsNull() || config.HostSpecs.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("hosts"),
			"Hosts not defined",
			"The cluster must have at least 1 host.",
		)
		return
	}

	shards := map[string]models.Shard{}
	resp.Diagnostics.Append(config.Shards.ElementsAs(ctx, &shards, true)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hosts := map[string]models.Host{}
	resp.Diagnostics.Append(config.HostSpecs.ElementsAs(ctx, &hosts, true)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the set of shard names referenced by hosts.
	referencedShards := map[string]struct{}{}
	for hostName, host := range hosts {
		shardName := host.GetShard()
		if shardName == "zk" {
			// ZOOKEEPER/KEEPER hosts do not belong to a shard.
			continue
		}
		if shardName == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("hosts").AtMapKey(hostName).AtName("shard_name"),
				"shard_name is required",
				fmt.Sprintf("Host %q is a CLICKHOUSE host and must have shard_name set.", hostName),
			)
			continue
		}
		referencedShards[shardName] = struct{}{}
		if _, ok := shards[shardName]; !ok {
			resp.Diagnostics.AddAttributeError(
				path.Root("hosts").AtMapKey(hostName).AtName("shard_name"),
				"Shard not defined",
				fmt.Sprintf("Host %q references shard %q which is not present in the shards block. "+
					"Add an entry for %q to the shards block.", hostName, shardName, shardName),
			)
		}
	}

	// Every shard must have at least one host.
	for shardName := range shards {
		if _, ok := referencedShards[shardName]; !ok {
			resp.Diagnostics.AddAttributeError(
				path.Root("shards").AtMapKey(shardName),
				"Shard has no hosts",
				fmt.Sprintf("Shard %q has no hosts assigned to it. "+
					"Add at least one host with shard_name = %q, or remove the shard from the shards block.", shardName, shardName),
			)
		}
	}
}
