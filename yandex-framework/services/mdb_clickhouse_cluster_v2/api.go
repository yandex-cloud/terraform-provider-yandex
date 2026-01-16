package mdb_clickhouse_cluster_v2

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/retry"
)

const (
	defaultMDBPageSize = 1000
)

var clickhouseApi = ClickHouseAPI{}

type ClickHouseAPI struct{}

type ClickHouseOpts struct {
	CopySchema            bool
	HasCoordinator        bool
	MapShardNameShardSpec map[string]*clickhouse.ShardConfigSpec
}

// Cluster

func (c *ClickHouseAPI) GetCluster(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string) *clickhouse.Cluster {
	tflog.Debug(ctx, "Reading ClickHouse Cluster", map[string]any{"cluster_id": cid})

	cluster, err := sdk.MDB().Clickhouse().Cluster().Get(ctx, &clickhouse.GetClusterRequest{
		ClusterId: cid,
	})

	if err != nil {
		diag.AddError(
			"Failed to read resource",
			fmt.Sprintf("Error while requesting API to read ClickHouse cluster %q: %s", cid, err.Error()),
		)
		return nil
	}

	return cluster
}

func (c *ClickHouseAPI) DeleteCluster(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string) {
	tflog.Debug(ctx, "Deleting ClickHouse Cluster", map[string]any{"cluster_id": cid})

	op, err := sdk.WrapOperation(sdk.MDB().Clickhouse().Cluster().Delete(ctx, &clickhouse.DeleteClusterRequest{
		ClusterId: cid,
	}))

	if err != nil {
		diags.AddError(
			"Failed to delete resource",
			fmt.Sprintf("Error while requesting API to delete ClickHouse cluster %q: %s", cid, err.Error()),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to delete resource",
			fmt.Sprintf("Error while waiting for operation %q to delete ClickHouse cluster %q: %s", op.Id(), cid, err.Error()),
		)
	}
}

func (c *ClickHouseAPI) CreateCluster(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, req *clickhouse.CreateClusterRequest) string {
	tflog.Debug(ctx, "Creating ClickHouse Cluster", map[string]any{"request": req})

	op, err := sdk.WrapOperation(sdk.MDB().Clickhouse().Cluster().Create(ctx, req))
	if err != nil {
		diags.AddError(
			"Failed to create resource",
			fmt.Sprintf("Error while requesting API to create ClickHouse cluster: %s", err.Error()),
		)
		return ""
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		diags.AddError(
			"Failed to create resource",
			fmt.Sprintf("Error while unmarshaling for operation %q API response metadata: %s", op.Id(), err.Error()),
		)
		return ""
	}

	md, ok := protoMetadata.(*clickhouse.CreateClusterMetadata)
	if !ok {
		diags.AddError(
			"Failed to create resource",
			fmt.Sprintf("Error while unmarshaling for operation %q API response metadata", op.Id()),
		)
		return ""
	}

	if err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to create resource",
			fmt.Sprintf("Error while waiting for operation %q to create ClickHouse cluster: %s", op.Id(), err.Error()),
		)
		return ""
	}

	return md.ClusterId
}

func (c *ClickHouseAPI) UpdateCluster(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, req *clickhouse.UpdateClusterRequest) {
	tflog.Debug(ctx, "Updating ClickHouse Cluster", map[string]any{"request": req})

	if req == nil || len(req.UpdateMask.Paths) == 0 {
		return
	}

	op, err := sdk.WrapOperation(sdk.MDB().Clickhouse().Cluster().Update(ctx, req))
	if err != nil {
		diag.AddError(
			"Failed to update resource",
			fmt.Sprintf("Error while requesting API to update ClickHouse cluster: %s", err.Error()),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to update resource",
			fmt.Sprintf("Error while waiting for operation %q to update ClickHouse cluster: %s", op.Id(), err.Error()),
		)
		return
	}
}

func (c *ClickHouseAPI) MoveCluster(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, req *clickhouse.MoveClusterRequest) {
	tflog.Debug(ctx, "Moving ClickHouse Cluster", map[string]any{"request": req})

	if req == nil {
		return
	}

	op, err := sdk.WrapOperation(sdk.MDB().Clickhouse().Cluster().Move(ctx, req))
	if err != nil {
		diag.AddError(
			"Failed to move cluster",
			fmt.Sprintf("Error while requesting API to move ClickHouse cluster: %s", err.Error()),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to move cluster",
			fmt.Sprintf("Error while waiting for operation %q to move ClickHouse cluster: %s", op.Id(), err.Error()),
		)
		return
	}
}

// Hosts

func (c *ClickHouseAPI) ListHosts(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string) []*clickhouse.Host {
	hosts := []*clickhouse.Host{}
	pageToken := ""

	for {
		resp, err := sdk.MDB().Clickhouse().Cluster().ListHosts(ctx, &clickhouse.ListClusterHostsRequest{
			ClusterId: cid,
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			diags.AddError(
				"Failed to read resource",
				fmt.Sprintf("Error while requesting API to read hosts of cluster ClickHouse '%s': %s", cid, err.Error()),
			)
			return nil
		}

		hosts = append(hosts, resp.Hosts...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}
	return hosts
}

func (c *ClickHouseAPI) CreateHosts(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string, specs []*clickhouse.HostSpec, opts ClickHouseOpts) {
	if len(specs) == 0 {
		return
	}

	hostType := specs[0].Type

	if (hostType == clickhouse.Host_ZOOKEEPER || hostType == clickhouse.Host_KEEPER) && !opts.HasCoordinator {
		addCoordinator(ctx, sdk, diags, cid, specs)
	} else {
		createHosts(ctx, sdk, diags, cid, specs, opts.CopySchema)
	}
}

func addCoordinator(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string, specs []*clickhouse.HostSpec) {
	request := &clickhouse.AddClusterZookeeperRequest{
		ClusterId: cid,
		HostSpecs: specs,
	}

	tflog.Debug(ctx, "Creating ClickHouse coordinator", map[string]any{"request": request})

	op, err := sdk.WrapOperation(sdk.MDB().Clickhouse().Cluster().AddZookeeper(ctx, request))
	if err != nil {
		diags.AddError(
			"Failed to create resource",
			fmt.Sprintf("Error while requesting API to create ClickHouse coordinator: %s", err.Error()),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to create resource",
			fmt.Sprintf("Error while waiting for operation %q to create ClickHouse coordinator: %s", op.Id(), err.Error()),
		)
		return
	}
}

func createHosts(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string, specs []*clickhouse.HostSpec, copySchema bool) {
	request := &clickhouse.AddClusterHostsRequest{
		ClusterId:  cid,
		HostSpecs:  specs,
		CopySchema: &wrappers.BoolValue{Value: copySchema},
	}

	tflog.Debug(ctx, "Creating ClickHouse hosts", map[string]any{"request": request})

	op, err := sdk.WrapOperation(sdk.MDB().Clickhouse().Cluster().AddHosts(ctx, request))
	if err != nil {
		diags.AddError(
			"Failed to create hosts",
			fmt.Sprintf("Error while requesting API to create hosts ClickHouse cluster %q: %s", cid, err.Error()),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to create hosts",
			fmt.Sprintf("Error while waiting for operation %q to create host ClickHouse cluster %q: %s", op.Id(), cid, err.Error()),
		)
		return
	}
}

func (c *ClickHouseAPI) UpdateHosts(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string, specs []*clickhouse.UpdateHostSpec) {
	for _, spec := range specs {
		request := &clickhouse.UpdateClusterHostsRequest{
			ClusterId: cid,
			UpdateHostSpecs: []*clickhouse.UpdateHostSpec{
				spec,
			},
		}
		op, err := retry.ConflictingOperation(ctx, sdk, func() (*operation.Operation, error) {
			tflog.Debug(ctx, "Sending ClickHouse cluster update host request", map[string]any{"request": request})
			return sdk.MDB().Clickhouse().Cluster().UpdateHosts(ctx, request)
		})
		if err != nil {
			diags.AddError(
				"Failed to update hosts",
				fmt.Sprintf("Error while requesting API to update host ClickHouse cluster %q: %s", cid, err.Error()),
			)
			return
		}

		if err = op.Wait(ctx); err != nil {
			diags.AddError(
				"Failed to update hosts",
				fmt.Sprintf("Error while waiting for operation %q to update host ClickHouse cluster %q: %s", op.Id(), cid, err.Error()),
			)
			return
		}
	}
}

func (c *ClickHouseAPI) DeleteHosts(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string, fqdns []string) {
	if len(fqdns) == 0 {
		return
	}

	op, err := sdk.WrapOperation(
		sdk.MDB().Clickhouse().Cluster().DeleteHosts(ctx, &clickhouse.DeleteClusterHostsRequest{
			ClusterId: cid,
			HostNames: fqdns,
		}),
	)
	if err != nil {
		diags.AddError(
			"Failed to delete hosts",
			fmt.Sprintf("Error while requesting API to delete hosts ClickHouse cluster %q: %s", cid, err.Error()),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to delete hosts",
			fmt.Sprintf("Error while waiting for operation %q to delete hosts ClickHouse cluster %q: %s", op.Id(), cid, err.Error()),
		)
		return
	}
}

// Shards

func (c *ClickHouseAPI) GetShard(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid, shardName string) *clickhouse.Shard {
	tflog.Debug(ctx, "Reading ClickHouse shard", map[string]any{"cluster_id": cid, "shard_name": shardName})

	cluster, err := sdk.MDB().Clickhouse().Cluster().GetShard(ctx, &clickhouse.GetClusterShardRequest{
		ClusterId: cid,
		ShardName: shardName,
	})

	if err != nil {
		diag.AddError(
			"Failed to read resource",
			fmt.Sprintf("Error while requesting API to read ClickHouse shard %q: %s", cid, err.Error()),
		)
		return nil
	}

	return cluster
}

func (c *ClickHouseAPI) CreateShard(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid, shardName string, hostSpecs []*clickhouse.HostSpec, opts ClickHouseOpts) {
	if len(hostSpecs) == 0 {
		return
	}

	request := &clickhouse.AddClusterShardRequest{
		ClusterId:  cid,
		ShardName:  shardName,
		HostSpecs:  hostSpecs,
		CopySchema: &wrappers.BoolValue{Value: opts.CopySchema},
	}

	if shardSpec, ok := opts.MapShardNameShardSpec[shardName]; ok {
		request.ConfigSpec = shardSpec
	}

	op, err := sdk.WrapOperation(
		sdk.MDB().Clickhouse().Cluster().AddShard(ctx, request),
	)
	if err != nil {
		diags.AddError(
			"Failed to create shard",
			fmt.Sprintf("Error while requesting API to create shard ClickHouse cluster %q: %s", cid, err.Error()),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to create hosts",
			fmt.Sprintf("Error while waiting for operation %q to create shard ClickHouse cluster %q: %s", op.Id(), cid, err.Error()),
		)
		return
	}
}

func (c *ClickHouseAPI) UpdateShard(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, req *clickhouse.UpdateClusterShardRequest) {
	tflog.Debug(ctx, "Updating ClickHouse shard", map[string]any{"request": req})

	op, err := sdk.WrapOperation(sdk.MDB().Clickhouse().Cluster().UpdateShard(ctx, req))
	if err != nil {
		diag.AddError(
			"Failed to update resource",
			fmt.Sprintf("Error while requesting API to update ClickHouse shard: %s", err.Error()),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to update resource",
			fmt.Sprintf("Error while waiting for operation %q to update ClickHouse shard: %s", op.Id(), err.Error()),
		)
		return
	}
}

func (c *ClickHouseAPI) DeleteShard(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string, shardName string) {
	op, err := sdk.WrapOperation(
		sdk.MDB().Clickhouse().Cluster().DeleteShard(ctx, &clickhouse.DeleteClusterShardRequest{
			ClusterId: cid,
			ShardName: shardName,
		}),
	)
	if err != nil {
		diags.AddError(
			"Failed to delete shard",
			fmt.Sprintf("Error while requesting API to delete shard ClickHouse cluster %q: %s", cid, err.Error()),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to delete shard",
			fmt.Sprintf("Error while waiting for operation %q to delete shard ClickHouse cluster %q: %s", op.Id(), cid, err.Error()),
		)
		return
	}
}

func (c *ClickHouseAPI) ListShards(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string) []*clickhouse.Shard {
	shards := []*clickhouse.Shard{}
	pageToken := ""

	for {
		resp, err := sdk.MDB().Clickhouse().Cluster().ListShards(ctx, &clickhouse.ListClusterShardsRequest{
			ClusterId: cid,
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			diags.AddError(
				"Failed to read resource",
				fmt.Sprintf("Error while requesting API to read shards of cluster ClickHouse '%s': %s", cid, err.Error()),
			)
			return nil
		}

		shards = append(shards, resp.Shards...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}
	return shards
}

// Format schemas

func (c *ClickHouseAPI) CreateFormatSchema(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, req *clickhouse.CreateFormatSchemaRequest) {
	tflog.Debug(ctx, "Creating ClickHouse format schema", map[string]any{"request": req})

	op, err := sdk.WrapOperation(sdk.MDB().Clickhouse().FormatSchema().Create(ctx, req))
	if err != nil {
		diags.AddError(
			"Failed to create resource",
			fmt.Sprintf("Error while requesting API to create ClickHouse format schema: %s", err.Error()),
		)
		return
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		diags.AddError(
			"Failed to create resource",
			fmt.Sprintf("Error while unmarshaling for operation %q API response metadata: %s", op.Id(), err.Error()),
		)
		return
	}

	_, ok := protoMetadata.(*clickhouse.CreateFormatSchemaMetadata)
	if !ok {
		diags.AddError(
			"Failed to create resource",
			fmt.Sprintf("Error while unmarshaling for operation %q API response metadata", op.Id()),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to create resource",
			fmt.Sprintf("Error while waiting for operation %q to create ClickHouse format schema: %s", op.Id(), err.Error()),
		)
		return
	}
}

func (c *ClickHouseAPI) UpdateFormatSchema(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, req *clickhouse.UpdateFormatSchemaRequest) {
	tflog.Debug(ctx, "Updating ClickHouse format schema", map[string]any{"request": req})

	op, err := sdk.WrapOperation(sdk.MDB().Clickhouse().FormatSchema().Update(ctx, req))
	if err != nil {
		diag.AddError(
			"Failed to update resource",
			fmt.Sprintf("Error while requesting API to update ClickHouse format schema: %s", err.Error()),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to update resource",
			fmt.Sprintf("Error while waiting for operation %q to update ClickHouse format schema: %s", op.Id(), err.Error()),
		)
		return
	}
}

func (c *ClickHouseAPI) ListFormatSchemas(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string) []*clickhouse.FormatSchema {
	schemas := []*clickhouse.FormatSchema{}
	pageToken := ""

	for {
		resp, err := sdk.MDB().Clickhouse().FormatSchema().List(ctx, &clickhouse.ListFormatSchemasRequest{
			ClusterId: cid,
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			diags.AddError(
				"Failed to read resource",
				fmt.Sprintf("Error while requesting API to read format schemas of cluster ClickHouse '%s': %s", cid, err.Error()),
			)
			return nil
		}

		schemas = append(schemas, resp.FormatSchemas...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}
	return schemas
}

func (c *ClickHouseAPI) DeleteFormatSchema(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid, name string) {
	tflog.Debug(ctx, "Deleting ClickHouse format schema", map[string]any{"name": name})

	op, err := sdk.WrapOperation(sdk.MDB().Clickhouse().FormatSchema().Delete(ctx, &clickhouse.DeleteFormatSchemaRequest{
		ClusterId:        cid,
		FormatSchemaName: name,
	}))

	if err != nil {
		diags.AddError(
			"Failed to delete resource",
			fmt.Sprintf("Error while requesting API to delete ClickHouse format schema %q: %s", cid, err.Error()),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to delete resource",
			fmt.Sprintf("Error while waiting for operation %q to delete ClickHouse format schema %q: %s", op.Id(), cid, err.Error()),
		)
		return
	}
}

// ML models

func (c *ClickHouseAPI) CreateMlModel(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, req *clickhouse.CreateMlModelRequest) {
	tflog.Debug(ctx, "Creating ClickHouse ML model", map[string]any{"request": req})

	op, err := sdk.WrapOperation(sdk.MDB().Clickhouse().MlModel().Create(ctx, req))
	if err != nil {
		diags.AddError(
			"Failed to create resource",
			fmt.Sprintf("Error while requesting API to create ClickHouse ML model: %s", err.Error()),
		)
		return
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		diags.AddError(
			"Failed to create resource",
			fmt.Sprintf("Error while unmarshaling for operation %q API response metadata: %s", op.Id(), err.Error()),
		)
		return
	}

	_, ok := protoMetadata.(*clickhouse.CreateMlModelMetadata)
	if !ok {
		diags.AddError(
			"Failed to create resource",
			fmt.Sprintf("Error while unmarshaling for operation %q API response metadata", op.Id()),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to create resource",
			fmt.Sprintf("Error while waiting for operation %q to create ClickHouse ML model: %s", op.Id(), err.Error()),
		)
		return
	}
}

func (c *ClickHouseAPI) UpdateMlModel(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, req *clickhouse.UpdateMlModelRequest) {
	tflog.Debug(ctx, "Updating ClickHouse ML model", map[string]any{"request": req})

	op, err := sdk.WrapOperation(sdk.MDB().Clickhouse().MlModel().Update(ctx, req))
	if err != nil {
		diag.AddError(
			"Failed to update resource",
			fmt.Sprintf("Error while requesting API to update ClickHouse ML model: %s", err.Error()),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to update resource",
			fmt.Sprintf("Error while waiting for operation %q to update ClickHouse ML model: %s", op.Id(), err.Error()),
		)
		return
	}
}

func (c *ClickHouseAPI) ListMlModels(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string) []*clickhouse.MlModel {
	models := []*clickhouse.MlModel{}
	pageToken := ""

	for {
		resp, err := sdk.MDB().Clickhouse().MlModel().List(ctx, &clickhouse.ListMlModelsRequest{
			ClusterId: cid,
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			diags.AddError(
				"Failed to read resource",
				fmt.Sprintf("Error while requesting API to read ML models of cluster ClickHouse '%s': %s", cid, err.Error()),
			)
			return nil
		}

		models = append(models, resp.MlModels...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}
	return models
}

func (c *ClickHouseAPI) DeleteMlModel(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid, name string) {
	tflog.Debug(ctx, "Deleting ClickHouse ML model", map[string]any{"name": name})

	op, err := sdk.WrapOperation(sdk.MDB().Clickhouse().MlModel().Delete(ctx, &clickhouse.DeleteMlModelRequest{
		ClusterId:   cid,
		MlModelName: name,
	}))

	if err != nil {
		diags.AddError(
			"Failed to delete resource",
			fmt.Sprintf("Error while requesting API to delete ClickHouse ML model %q: %s", cid, err.Error()),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to delete resource",
			fmt.Sprintf("Error while waiting for operation %q to delete ClickHouse ML model %q: %s", op.Id(), cid, err.Error()),
		)
		return
	}
}

// Shard groups

func (c *ClickHouseAPI) CreateShardGroup(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, req *clickhouse.CreateClusterShardGroupRequest) {
	tflog.Debug(ctx, "Creating ClickHouse shard group", map[string]any{"request": req})

	op, err := sdk.WrapOperation(sdk.MDB().Clickhouse().Cluster().CreateShardGroup(ctx, req))
	if err != nil {
		diags.AddError(
			"Failed to create resource",
			fmt.Sprintf("Error while requesting API to create ClickHouse shard group: %s", err.Error()),
		)
		return
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		diags.AddError(
			"Failed to create resource",
			fmt.Sprintf("Error while unmarshaling for operation %q API response metadata: %s", op.Id(), err.Error()),
		)
		return
	}

	_, ok := protoMetadata.(*clickhouse.CreateClusterShardGroupMetadata)
	if !ok {
		diags.AddError(
			"Failed to create resource",
			fmt.Sprintf("Error while unmarshaling for operation %q API response metadata", op.Id()),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to create resource",
			fmt.Sprintf("Error while waiting for operation %q to create ClickHouse shard group: %s", op.Id(), err.Error()),
		)
		return
	}
}

func (c *ClickHouseAPI) UpdateShardGroup(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, req *clickhouse.UpdateClusterShardGroupRequest) {
	tflog.Debug(ctx, "Updating ClickHouse shard group", map[string]any{"request": req})

	op, err := sdk.WrapOperation(sdk.MDB().Clickhouse().Cluster().UpdateShardGroup(ctx, req))
	if err != nil {
		diag.AddError(
			"Failed to update resource",
			fmt.Sprintf("Error while requesting API to update ClickHouse shard group: %s", err.Error()),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to update resource",
			fmt.Sprintf("Error while waiting for operation %q to update ClickHouse shard group: %s", op.Id(), err.Error()),
		)
		return
	}
}

func (c *ClickHouseAPI) ListShardGroups(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string) []*clickhouse.ShardGroup {
	groups := []*clickhouse.ShardGroup{}
	pageToken := ""

	for {
		resp, err := sdk.MDB().Clickhouse().Cluster().ListShardGroups(ctx, &clickhouse.ListClusterShardGroupsRequest{
			ClusterId: cid,
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			diags.AddError(
				"Failed to read resource",
				fmt.Sprintf("Error while requesting API to read shard groups of cluster ClickHouse '%s': %s", cid, err.Error()),
			)
			return nil
		}

		groups = append(groups, resp.ShardGroups...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}
	return groups
}

func (c *ClickHouseAPI) DeleteShardGroup(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid, name string) {
	tflog.Debug(ctx, "Deleting ClickHouse shard group", map[string]any{"name": name})

	op, err := sdk.WrapOperation(sdk.MDB().Clickhouse().Cluster().DeleteShardGroup(ctx, &clickhouse.DeleteClusterShardGroupRequest{
		ClusterId:      cid,
		ShardGroupName: name,
	}))

	if err != nil {
		diags.AddError(
			"Failed to delete resource",
			fmt.Sprintf("Error while requesting API to delete ClickHouse shard group %q: %s", cid, err.Error()),
		)
		return
	}

	if err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to delete resource",
			fmt.Sprintf("Error while waiting for operation %q to delete ClickHouse shard group %q: %s", op.Id(), cid, err.Error()),
		)
		return
	}
}
