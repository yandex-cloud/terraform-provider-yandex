package validators

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_clickhouse_cluster_v2/models"
)

type ClickhouseShardConflictValidator struct {
	AttrName string
}

func (v ClickhouseShardConflictValidator) Description(_ context.Context) string {
	return fmt.Sprintf(`"clickhouse.%[1]s" and "shards[*].%[1]s" cannot both be set`, v.AttrName)
}

func (v ClickhouseShardConflictValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v ClickhouseShardConflictValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config models.Cluster
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Resolve clickhouse.<AttrName>
	if config.ClickHouse.IsNull() || config.ClickHouse.IsUnknown() {
		return
	}
	var ch models.Clickhouse
	resp.Diagnostics.Append(config.ClickHouse.As(ctx, &ch, datasize.DefaultOpts)...)
	if resp.Diagnostics.HasError() {
		return
	}
	chAttr := clickhouseAttrByName(ch, v.AttrName)
	if chAttr.IsNull() || chAttr.IsUnknown() {
		return // clickhouse.<attr> not set — no conflict possible
	}

	// Check if any shard has <AttrName> set
	if config.Shards.IsNull() || config.Shards.IsUnknown() {
		return
	}
	var shards map[string]models.Shard
	resp.Diagnostics.Append(config.Shards.ElementsAs(ctx, &shards, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	for shardName, shard := range shards {
		shardAttr := shardAttrByName(shard, v.AttrName)
		if !shardAttr.IsNull() && !shardAttr.IsUnknown() {
			resp.Diagnostics.AddAttributeError(
				path.Root("clickhouse").AtName(v.AttrName),
				"Invalid Attribute Combination",
				fmt.Sprintf(
					`"clickhouse.%s" and "shards[*].%s" cannot both be set. Shard %q has %s defined.`,
					v.AttrName, v.AttrName, shardName, v.AttrName,
				),
			)
			return
		}
	}
}

func clickhouseAttrByName(ch models.Clickhouse, name string) types.Object {
	switch name {
	case "resources":
		return ch.Resources
	case "disk_size_autoscaling":
		return ch.DiskSizeAutoscaling
	default:
		panic("unknown attr: " + name)
	}
}

func shardAttrByName(s models.Shard, name string) types.Object {
	switch name {
	case "resources":
		return s.Resources
	case "disk_size_autoscaling":
		return s.DiskSizeAutoscaling
	default:
		panic("unknown attr: " + name)
	}
}
