package mdb_clickhouse_cluster_v2

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
	clickhouseConfig "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1/config"
	clickhousesdk "github.com/yandex-cloud/go-sdk/services/mdb/clickhouse/v1"
	ycsdk "github.com/yandex-cloud/go-sdk/v2"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/retry"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	defaultMDBPageSize               = 1000
	defaultConvertTablesToReplicated = true
	redactedClickHousePassword       = "[REDACTED]"
)

var clickhouseApi = ClickHouseAPI{}

type ClickHouseAPI struct{}

type ClickHouseOpts struct {
	CopySchema               bool
	CoordinatorResources     *clickhouse.Resources
	HasCoordinator           bool
	PlanShardSpecByShardName map[string]*clickhouse.ShardConfigSpec
}

func redactClickHouseCreateClusterRequest(req *clickhouse.CreateClusterRequest) *clickhouse.CreateClusterRequest {
	if req == nil {
		return nil
	}
	redacted := proto.Clone(req).(*clickhouse.CreateClusterRequest)
	redactClickHouseConfigSpecPasswords(redacted.ConfigSpec)
	for _, userSpec := range redacted.UserSpecs {
		if userSpec.GetPassword() != "" {
			userSpec.Password = redactedClickHousePassword
		}
	}
	return redacted
}

func redactClickHouseUpdateClusterRequest(req *clickhouse.UpdateClusterRequest) *clickhouse.UpdateClusterRequest {
	if req == nil {
		return nil
	}
	redacted := proto.Clone(req).(*clickhouse.UpdateClusterRequest)
	redactClickHouseConfigSpecPasswords(redacted.ConfigSpec)
	return redacted
}

func redactClickHouseRestoreClusterRequest(req *clickhouse.RestoreClusterRequest) *clickhouse.RestoreClusterRequest {
	if req == nil {
		return nil
	}
	redacted := proto.Clone(req).(*clickhouse.RestoreClusterRequest)
	redactClickHouseConfigSpecPasswords(redacted.ConfigSpec)
	return redacted
}

func redactClickHouseConfigSpecPasswords(configSpec *clickhouse.ConfigSpec) {
	if configSpec == nil {
		return
	}
	if configSpec.GetAdminPassword() != "" {
		configSpec.SetAdminPassword(redactedClickHousePassword)
	}

	clickhouseConfig := configSpec.GetClickhouse().GetConfig()
	if clickhouseConfig == nil {
		return
	}
	if kafka := clickhouseConfig.GetKafka(); kafka != nil && kafka.GetSaslPassword() != "" {
		kafka.SaslPassword = redactedClickHousePassword
	}
	for _, kafkaTopic := range clickhouseConfig.GetKafkaTopics() {
		if settings := kafkaTopic.GetSettings(); settings != nil && settings.GetSaslPassword() != "" {
			settings.SaslPassword = redactedClickHousePassword
		}
	}
	if rabbitmq := clickhouseConfig.GetRabbitmq(); rabbitmq != nil && rabbitmq.GetPassword() != "" {
		rabbitmq.Password = redactedClickHousePassword
	}
}

// Cluster

func (c *ClickHouseAPI) GetCluster(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid string) *clickhouse.Cluster {
	tflog.Debug(ctx, "Reading ClickHouse Cluster", map[string]any{"cluster_id": cid})

	cluster, err := clickhousesdk.NewClusterClient(sdk).Get(ctx, &clickhouse.GetClusterRequest{
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

	op, err := clickhousesdk.NewClusterClient(sdk).Delete(ctx, &clickhouse.DeleteClusterRequest{
		ClusterId: cid,
	})

	if err != nil {
		diags.AddError(
			"Failed to delete resource",
			fmt.Sprintf("Error while requesting API to delete ClickHouse cluster %q: %s", cid, err.Error()),
		)
		return
	}

	if _, err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to delete resource",
			fmt.Sprintf("Error while waiting for operation %q to delete ClickHouse cluster %q: %s", op.ID(), cid, err.Error()),
		)
	}
}

func (c *ClickHouseAPI) CreateCluster(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, req *clickhouse.CreateClusterRequest) string {
	tflog.Debug(ctx, "Creating ClickHouse Cluster", map[string]any{"request": redactClickHouseCreateClusterRequest(req)})

	op, err := clickhousesdk.NewClusterClient(sdk).Create(ctx, req)
	if err != nil {
		diags.AddError(
			"Failed to create resource",
			fmt.Sprintf("Error while requesting API to create ClickHouse cluster: %s", err.Error()),
		)
		return ""
	}

	md := op.Metadata()

	if _, err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to create resource",
			fmt.Sprintf("Error while waiting for operation %q to create ClickHouse cluster: %s", op.ID(), err.Error()),
		)
		return ""
	}

	return md.ClusterId
}

func (c *ClickHouseAPI) UpdateCluster(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, req *clickhouse.UpdateClusterRequest) {
	tflog.Debug(ctx, "Updating ClickHouse Cluster", map[string]any{"request": redactClickHouseUpdateClusterRequest(req)})

	if req == nil || len(req.UpdateMask.Paths) == 0 {
		return
	}

	op, err := clickhousesdk.NewClusterClient(sdk).Update(ctx, req)
	if err != nil {
		diag.AddError(
			"Failed to update resource",
			fmt.Sprintf("Error while requesting API to update ClickHouse cluster: %s", err.Error()),
		)
		return
	}

	if _, err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to update resource",
			fmt.Sprintf("Error while waiting for operation %q to update ClickHouse cluster: %s", op.ID(), err.Error()),
		)
		return
	}
}

func (c *ClickHouseAPI) MoveCluster(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, req *clickhouse.MoveClusterRequest) {
	tflog.Debug(ctx, "Moving ClickHouse Cluster", map[string]any{"request": req})

	if req == nil {
		return
	}

	op, err := clickhousesdk.NewClusterClient(sdk).Move(ctx, req)
	if err != nil {
		diag.AddError(
			"Failed to move cluster",
			fmt.Sprintf("Error while requesting API to move ClickHouse cluster: %s", err.Error()),
		)
		return
	}

	if _, err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to move cluster",
			fmt.Sprintf("Error while waiting for operation %q to move ClickHouse cluster: %s", op.ID(), err.Error()),
		)
		return
	}
}

// Hosts

func (c *ClickHouseAPI) ListHosts(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string) []*clickhouse.Host {
	hosts := []*clickhouse.Host{}
	pageToken := ""

	for {
		resp, err := clickhousesdk.NewClusterClient(sdk).ListHosts(ctx, &clickhouse.ListClusterHostsRequest{
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
		addCoordinator(ctx, sdk, diags, cid, specs, opts.CoordinatorResources)
	} else {
		createHosts(ctx, sdk, diags, cid, specs, opts.CopySchema)
	}
}

func addCoordinator(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string, specs []*clickhouse.HostSpec, resources *clickhouse.Resources) {
	request := &clickhouse.AddClusterZookeeperRequest{
		ClusterId:                 cid,
		ConvertTablesToReplicated: wrapperspb.Bool(defaultConvertTablesToReplicated),
		Resources:                 resources,
		HostSpecs:                 specs,
	}

	tflog.Debug(ctx, "Creating ClickHouse coordinator", map[string]any{"request": request})

	op, err := clickhousesdk.NewClusterClient(sdk).AddZookeeper(ctx, request)
	if err != nil {
		diags.AddError(
			"Failed to create resource",
			fmt.Sprintf("Error while requesting API to create ClickHouse coordinator: %s", err.Error()),
		)
		return
	}

	if _, err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to create resource",
			fmt.Sprintf("Error while waiting for operation %q to create ClickHouse coordinator: %s", op.ID(), err.Error()),
		)
		return
	}
}

func (c *ClickHouseAPI) MigrateToKeeper(
	ctx context.Context,
	sdk *ycsdk.SDK,
	diags *diag.Diagnostics,
	request *clickhouse.MigrateClusterToKeeperRequest,
) {
	hosts, err := clickhousesdk.NewClusterClient(sdk).ListHosts(ctx, &clickhouse.ListClusterHostsRequest{
		ClusterId: request.GetClusterId(),
		PageSize:  defaultMDBPageSize,
	})
	if err != nil {
		diags.AddError(
			"Failed to inspect ClickHouse coordinator hosts",
			fmt.Sprintf("Error while requesting API to list hosts of ClickHouse cluster %q before migration to Keeper: %s", request.GetClusterId(), err.Error()),
		)
		return
	}

	liveCoordinatorTypes := getAPICoordinatorHostTypes(hosts.GetHosts())
	if liveCoordinatorTypes.hasKeeper && !liveCoordinatorTypes.hasZooKeeper {
		tflog.Debug(ctx, "ClickHouse cluster is already migrated to Keeper", map[string]any{"cluster_id": request.GetClusterId()})
		return
	}
	if liveCoordinatorTypes.hasKeeper && liveCoordinatorTypes.hasZooKeeper {
		diags.AddError(
			"Unable to migrate ClickHouse cluster to Keeper",
			fmt.Sprintf("ClickHouse cluster %q currently contains both ZooKeeper and Keeper hosts. Wait for the active migration to finish and retry Terraform apply.", request.GetClusterId()),
		)
		return
	}
	if !liveCoordinatorTypes.hasZooKeeper {
		diags.AddError(
			"Unable to migrate ClickHouse cluster to Keeper",
			fmt.Sprintf("ClickHouse cluster %q has no dedicated ZooKeeper hosts to migrate.", request.GetClusterId()),
		)
		return
	}

	tflog.Debug(ctx, "Migrating ClickHouse cluster to Keeper", map[string]any{"request": request})
	op, err := clickhousesdk.NewClusterClient(sdk).MigrateToKeeper(ctx, request)
	if err != nil {
		diags.AddError(
			"Failed to migrate ClickHouse cluster to Keeper",
			fmt.Sprintf("Error while requesting API to migrate ClickHouse cluster %q to Keeper: %s", request.GetClusterId(), err.Error()),
		)
		return
	}

	if _, err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to migrate ClickHouse cluster to Keeper",
			fmt.Sprintf("Error while waiting for operation to migrate ClickHouse cluster %q to Keeper: %s", request.GetClusterId(), err.Error()),
		)
	}
}

func createHosts(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string, specs []*clickhouse.HostSpec, copySchema bool) {
	request := &clickhouse.AddClusterHostsRequest{
		ClusterId:  cid,
		HostSpecs:  specs,
		CopySchema: &wrappers.BoolValue{Value: copySchema},
	}

	tflog.Debug(ctx, "Creating ClickHouse hosts", map[string]any{"request": request})

	op, err := clickhousesdk.NewClusterClient(sdk).AddHosts(ctx, request)
	if err != nil {
		diags.AddError(
			"Failed to create hosts",
			fmt.Sprintf("Error while requesting API to create hosts ClickHouse cluster %q: %s", cid, err.Error()),
		)
		return
	}

	if _, err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to create hosts",
			fmt.Sprintf("Error while waiting for operation %q to create host ClickHouse cluster %q: %s", op.ID(), cid, err.Error()),
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
		op, err := retry.ConflictingOperationV2(ctx, sdk, func() (*clickhousesdk.ClusterUpdateHostsOperation, error) {
			tflog.Debug(ctx, "Sending ClickHouse cluster update host request", map[string]any{"request": request})
			return clickhousesdk.NewClusterClient(sdk).UpdateHosts(ctx, request)
		})
		if err != nil {
			diags.AddError(
				"Failed to update hosts",
				fmt.Sprintf("Error while requesting API to update host ClickHouse cluster %q: %s", cid, err.Error()),
			)
			return
		}

		if _, err = op.Wait(ctx); err != nil {
			diags.AddError(
				"Failed to update hosts",
				fmt.Sprintf("Error while waiting for operation %q to update host ClickHouse cluster %q: %s", op.ID(), cid, err.Error()),
			)
			return
		}
	}
}

func (c *ClickHouseAPI) DeleteHosts(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string, fqdns []string) {
	if len(fqdns) == 0 {
		return
	}

	op, err :=
		clickhousesdk.NewClusterClient(sdk).DeleteHosts(ctx, &clickhouse.DeleteClusterHostsRequest{
			ClusterId: cid,
			HostNames: fqdns,
		})
	if err != nil {
		diags.AddError(
			"Failed to delete hosts",
			fmt.Sprintf("Error while requesting API to delete hosts ClickHouse cluster %q: %s", cid, err.Error()),
		)
		return
	}

	if _, err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to delete hosts",
			fmt.Sprintf("Error while waiting for operation %q to delete hosts ClickHouse cluster %q: %s", op.ID(), cid, err.Error()),
		)
		return
	}
}

// Shards

func (c *ClickHouseAPI) GetShard(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, cid, shardName string) *clickhouse.Shard {
	tflog.Debug(ctx, "Reading ClickHouse shard", map[string]any{"cluster_id": cid, "shard_name": shardName})

	cluster, err := clickhousesdk.NewClusterClient(sdk).GetShard(ctx, &clickhouse.GetClusterShardRequest{
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

	if shardSpec, ok := opts.PlanShardSpecByShardName[shardName]; ok {
		request.ConfigSpec = shardSpec
	}

	op, err :=
		clickhousesdk.NewClusterClient(sdk).AddShard(ctx, request)
	if err != nil {
		diags.AddError(
			"Failed to create shard",
			fmt.Sprintf("Error while requesting API to create shard ClickHouse cluster %q: %s", cid, err.Error()),
		)
		return
	}

	if _, err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to create shard",
			fmt.Sprintf("Error while waiting for operation %q to create shard ClickHouse cluster %q: %s", op.ID(), cid, err.Error()),
		)
		return
	}
}

func (c *ClickHouseAPI) CreateShards(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string, hostSpecsByShardName map[string][]*clickhouse.HostSpec, opts ClickHouseOpts) {
	if len(hostSpecsByShardName) == 0 {
		return
	}

	var hostSpecs []*clickhouse.HostSpec
	var shardSpecs []*clickhouse.ShardSpec
	for shardName, shardHostSpecs := range hostSpecsByShardName {
		shardSpec := &clickhouse.ShardSpec{
			Name: shardName,
		}
		if shardConfigSpec, ok := opts.PlanShardSpecByShardName[shardName]; ok {
			shardSpec.ConfigSpec = shardConfigSpec
		}

		hostSpecs = append(hostSpecs, shardHostSpecs...)
		shardSpecs = append(shardSpecs, shardSpec)
	}

	request := &clickhouse.AddClusterShardsRequest{
		ClusterId:  cid,
		ShardSpecs: shardSpecs,
		HostSpecs:  hostSpecs,
		CopySchema: &wrappers.BoolValue{Value: opts.CopySchema},
	}

	op, err :=
		clickhousesdk.NewClusterClient(sdk).AddShards(ctx, request)
	if err != nil {
		diags.AddError(
			"Failed to create shards",
			fmt.Sprintf("Error while requesting API to create shards ClickHouse cluster %q: %s", cid, err.Error()),
		)
		return
	}

	if _, err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to create shards",
			fmt.Sprintf("Error while waiting for operation %q to create shards ClickHouse cluster %q: %s", op.ID(), cid, err.Error()),
		)
		return
	}
}

func (c *ClickHouseAPI) UpdateShard(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, req *clickhouse.UpdateClusterShardRequest) {
	tflog.Debug(ctx, "Updating ClickHouse shard", map[string]any{"request": req})

	op, err := clickhousesdk.NewClusterClient(sdk).UpdateShard(ctx, req)
	if err != nil {
		diag.AddError(
			"Failed to update resource",
			fmt.Sprintf("Error while requesting API to update ClickHouse shard: %s", err.Error()),
		)
		return
	}

	if _, err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to update resource",
			fmt.Sprintf("Error while waiting for operation %q to update ClickHouse shard: %s", op.ID(), err.Error()),
		)
		return
	}
}

func (c *ClickHouseAPI) DeleteShard(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string, shardName string) {
	op, err :=
		clickhousesdk.NewClusterClient(sdk).DeleteShard(ctx, &clickhouse.DeleteClusterShardRequest{
			ClusterId: cid,
			ShardName: shardName,
		})
	if err != nil {
		diags.AddError(
			"Failed to delete shard",
			fmt.Sprintf("Error while requesting API to delete shard ClickHouse cluster %q: %s", cid, err.Error()),
		)
		return
	}

	if _, err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to delete shard",
			fmt.Sprintf("Error while waiting for operation %q to delete shard ClickHouse cluster %q: %s", op.ID(), cid, err.Error()),
		)
		return
	}
}

func (c *ClickHouseAPI) DeleteShards(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string, shardNames []string) {
	op, err :=
		clickhousesdk.NewClusterClient(sdk).DeleteShards(ctx, &clickhouse.DeleteClusterShardsRequest{
			ClusterId:  cid,
			ShardNames: shardNames,
		})
	if err != nil {
		diags.AddError(
			"Failed to delete shards",
			fmt.Sprintf("Error while requesting API to delete shards ClickHouse cluster %q: %s", cid, err.Error()),
		)
		return
	}

	if _, err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to delete shards",
			fmt.Sprintf("Error while waiting for operation %q to delete shards ClickHouse cluster %q: %s", op.ID(), cid, err.Error()),
		)
		return
	}
}

func (c *ClickHouseAPI) ListShards(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string) []*clickhouse.Shard {
	shards := []*clickhouse.Shard{}
	pageToken := ""

	for {
		resp, err := clickhousesdk.NewClusterClient(sdk).ListShards(ctx, &clickhouse.ListClusterShardsRequest{
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

	op, err := clickhousesdk.NewFormatSchemaClient(sdk).Create(ctx, req)
	if err != nil {
		diags.AddError(
			"Failed to create resource",
			fmt.Sprintf("Error while requesting API to create ClickHouse format schema: %s", err.Error()),
		)
		return
	}

	if _, err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to create resource",
			fmt.Sprintf("Error while waiting for operation %q to create ClickHouse format schema: %s", op.ID(), err.Error()),
		)
		return
	}
}

func (c *ClickHouseAPI) UpdateFormatSchema(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, req *clickhouse.UpdateFormatSchemaRequest) {
	tflog.Debug(ctx, "Updating ClickHouse format schema", map[string]any{"request": req})

	op, err := clickhousesdk.NewFormatSchemaClient(sdk).Update(ctx, req)
	if err != nil {
		diags.AddError(
			"Failed to update resource",
			fmt.Sprintf("Error while requesting API to update ClickHouse format schema: %s", err.Error()),
		)
		return
	}

	if _, err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to update resource",
			fmt.Sprintf("Error while waiting for operation %q to update ClickHouse format schema: %s", op.ID(), err.Error()),
		)
	}
}

func (c *ClickHouseAPI) ListFormatSchemas(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string) []*clickhouse.FormatSchema {
	var schemas []*clickhouse.FormatSchema
	pageToken := ""

	for {
		resp, err := clickhousesdk.NewFormatSchemaClient(sdk).List(ctx, &clickhouse.ListFormatSchemasRequest{
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
			return schemas
		}
		pageToken = resp.NextPageToken
	}
}

func (c *ClickHouseAPI) DeleteFormatSchema(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid, name string) {
	tflog.Debug(ctx, "Deleting ClickHouse format schema", map[string]any{"name": name})

	op, err := clickhousesdk.NewFormatSchemaClient(sdk).Delete(ctx, &clickhouse.DeleteFormatSchemaRequest{
		ClusterId:        cid,
		FormatSchemaName: name,
	})
	if err != nil {
		diags.AddError(
			"Failed to delete resource",
			fmt.Sprintf("Error while requesting API to delete ClickHouse format schema %q: %s", cid, err.Error()),
		)
		return
	}

	if _, err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to delete resource",
			fmt.Sprintf("Error while waiting for operation %q to delete ClickHouse format schema %q: %s", op.ID(), cid, err.Error()),
		)
	}
}

// ML models

func (c *ClickHouseAPI) CreateMlModel(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, req *clickhouse.CreateMlModelRequest) {
	tflog.Debug(ctx, "Creating ClickHouse ML model", map[string]any{"request": req})

	op, err := clickhousesdk.NewMlModelClient(sdk).Create(ctx, req)
	if err != nil {
		diags.AddError(
			"Failed to create resource",
			fmt.Sprintf("Error while requesting API to create ClickHouse ML model: %s", err.Error()),
		)
		return
	}

	if _, err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to create resource",
			fmt.Sprintf("Error while waiting for operation %q to create ClickHouse ML model: %s", op.ID(), err.Error()),
		)
	}
}

func (c *ClickHouseAPI) UpdateMlModel(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, req *clickhouse.UpdateMlModelRequest) {
	tflog.Debug(ctx, "Updating ClickHouse ML model", map[string]any{"request": req})

	op, err := clickhousesdk.NewMlModelClient(sdk).Update(ctx, req)
	if err != nil {
		diags.AddError(
			"Failed to update resource",
			fmt.Sprintf("Error while requesting API to update ClickHouse ML model: %s", err.Error()),
		)
		return
	}

	if _, err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to update resource",
			fmt.Sprintf("Error while waiting for operation %q to update ClickHouse ML model: %s", op.ID(), err.Error()),
		)
	}
}

func (c *ClickHouseAPI) ListMlModels(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string) []*clickhouse.MlModel {
	var models []*clickhouse.MlModel
	pageToken := ""

	for {
		resp, err := clickhousesdk.NewMlModelClient(sdk).List(ctx, &clickhouse.ListMlModelsRequest{
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
			return models
		}
		pageToken = resp.NextPageToken
	}
}

func (c *ClickHouseAPI) DeleteMlModel(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid, name string) {
	tflog.Debug(ctx, "Deleting ClickHouse ML model", map[string]any{"name": name})

	op, err := clickhousesdk.NewMlModelClient(sdk).Delete(ctx, &clickhouse.DeleteMlModelRequest{
		ClusterId:   cid,
		MlModelName: name,
	})
	if err != nil {
		diags.AddError(
			"Failed to delete resource",
			fmt.Sprintf("Error while requesting API to delete ClickHouse ML model %q: %s", cid, err.Error()),
		)
		return
	}

	if _, err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to delete resource",
			fmt.Sprintf("Error while waiting for operation %q to delete ClickHouse ML model %q: %s", op.ID(), cid, err.Error()),
		)
	}
}

// Shard groups

func (c *ClickHouseAPI) CreateShardGroup(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, req *clickhouse.CreateClusterShardGroupRequest) {
	tflog.Debug(ctx, "Creating ClickHouse shard group", map[string]any{"request": req})

	op, err := clickhousesdk.NewClusterClient(sdk).CreateShardGroup(ctx, req)
	if err != nil {
		diags.AddError(
			"Failed to create resource",
			fmt.Sprintf("Error while requesting API to create ClickHouse shard group: %s", err.Error()),
		)
		return
	}

	if _, err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to create resource",
			fmt.Sprintf("Error while waiting for operation %q to create ClickHouse shard group: %s", op.ID(), err.Error()),
		)
	}
}

func (c *ClickHouseAPI) UpdateShardGroup(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, req *clickhouse.UpdateClusterShardGroupRequest) {
	tflog.Debug(ctx, "Updating ClickHouse shard group", map[string]any{"request": req})

	op, err := clickhousesdk.NewClusterClient(sdk).UpdateShardGroup(ctx, req)
	if err != nil {
		diags.AddError(
			"Failed to update resource",
			fmt.Sprintf("Error while requesting API to update ClickHouse shard group: %s", err.Error()),
		)
		return
	}

	if _, err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to update resource",
			fmt.Sprintf("Error while waiting for operation %q to update ClickHouse shard group: %s", op.ID(), err.Error()),
		)
	}
}

func (c *ClickHouseAPI) ListShardGroups(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string) []*clickhouse.ShardGroup {
	var groups []*clickhouse.ShardGroup
	pageToken := ""

	for {
		resp, err := clickhousesdk.NewClusterClient(sdk).ListShardGroups(ctx, &clickhouse.ListClusterShardGroupsRequest{
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
			return groups
		}
		pageToken = resp.NextPageToken
	}
}

func (c *ClickHouseAPI) DeleteShardGroup(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid, name string) {
	tflog.Debug(ctx, "Deleting ClickHouse shard group", map[string]any{"name": name})

	op, err := clickhousesdk.NewClusterClient(sdk).DeleteShardGroup(ctx, &clickhouse.DeleteClusterShardGroupRequest{
		ClusterId:      cid,
		ShardGroupName: name,
	})
	if err != nil {
		diags.AddError(
			"Failed to delete resource",
			fmt.Sprintf("Error while requesting API to delete ClickHouse shard group %q: %s", cid, err.Error()),
		)
		return
	}

	if _, err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to delete resource",
			fmt.Sprintf("Error while waiting for operation %q to delete ClickHouse shard group %q: %s", op.ID(), cid, err.Error()),
		)
	}
}

// Extensions

func (c *ClickHouseAPI) ListExtensions(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string) []*clickhouse.ClusterExtension {
	var extensions []*clickhouse.ClusterExtension
	pageToken := ""

	for {
		resp, err := clickhousesdk.NewClusterExtensionClient(sdk).List(ctx, &clickhouse.ListClusterExtensionsRequest{
			ClusterId: cid,
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			diags.AddError(
				"Failed to read resource",
				fmt.Sprintf("Error while requesting API to read extensions of cluster ClickHouse '%s': %s", cid, err.Error()),
			)
			return nil
		}

		extensions = append(extensions, resp.Extensions...)
		if resp.NextPageToken == "" {
			return extensions
		}
		pageToken = resp.NextPageToken
	}
}

func (c *ClickHouseAPI) CreateExtension(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, req *clickhouse.CreateClusterExtensionRequest) {
	tflog.Debug(ctx, "Creating ClickHouse cluster extension", map[string]any{"cluster_id": req.ClusterId, "extension": req.ExtensionSpec.GetName()})

	op, err := clickhousesdk.NewClusterExtensionClient(sdk).Create(ctx, req)
	if err != nil {
		diags.AddError(
			"Failed to create resource",
			fmt.Sprintf("Error while requesting API to create extension %q of ClickHouse cluster '%s': %s", req.ExtensionSpec.GetName(), req.ClusterId, err.Error()),
		)
		return
	}

	if _, err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to create resource",
			fmt.Sprintf("Error while waiting for operation %q to create extension %q of ClickHouse cluster '%s': %s", op.ID(), req.ExtensionSpec.GetName(), req.ClusterId, err.Error()),
		)
	}
}

func (c *ClickHouseAPI) SetExtensions(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string, specs []*clickhouse.ExtensionSpec) {
	tflog.Debug(ctx, "Setting ClickHouse cluster extensions", map[string]any{"cluster_id": cid})

	op, err := clickhousesdk.NewClusterExtensionClient(sdk).SetExtensions(ctx, &clickhouse.SetClusterExtensionsRequest{
		ClusterId:      cid,
		ExtensionSpecs: specs,
	})
	if err != nil {
		diags.AddError(
			"Failed to update resource",
			fmt.Sprintf("Error while requesting API to set extensions of ClickHouse cluster '%s': %s", cid, err.Error()),
		)
		return
	}

	if _, err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to update resource",
			fmt.Sprintf("Error while waiting for operation %q to set extensions of ClickHouse cluster '%s': %s", op.ID(), cid, err.Error()),
		)
	}
}

// Restore

func (c *ClickHouseAPI) RestoreCluster(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, req *clickhouse.RestoreClusterRequest) string {
	tflog.Debug(ctx, "Restoring ClickHouse Cluster from backup", map[string]any{"request": redactClickHouseRestoreClusterRequest(req)})

	op, err := clickhousesdk.NewClusterClient(sdk).Restore(ctx, req)
	if err != nil {
		diags.AddError(
			"Failed to restore resource",
			fmt.Sprintf("Error while requesting API to restore ClickHouse cluster from backup: %s", err.Error()),
		)
		return ""
	}

	md := op.Metadata()
	if _, err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to restore resource",
			fmt.Sprintf("Error while waiting for operation %q to restore ClickHouse cluster from backup: %s", op.ID(), err.Error()),
		)
		return ""
	}
	return md.ClusterId
}

// External dictionaries

func (c *ClickHouseAPI) ListExternalDictionaries(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid string) []*clickhouseConfig.ClickhouseConfig_ExternalDictionary {
	var externalDictionaries []*clickhouseConfig.ClickhouseConfig_ExternalDictionary
	pageToken := ""

	for {
		resp, err := clickhousesdk.NewClusterClient(sdk).ListExternalDictionaries(ctx, &clickhouse.ListClusterExternalDictionariesRequest{
			ClusterId: cid,
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			diags.AddError(
				"Failed to read resource",
				fmt.Sprintf("Error while requesting API to list external dictionaries of ClickHouse cluster '%s': %s", cid, err.Error()),
			)
			return nil
		}

		externalDictionaries = append(externalDictionaries, resp.ExternalDictionaries...)

		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return externalDictionaries
}

func (c *ClickHouseAPI) CreateExternalDictionary(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, req *clickhouse.CreateClusterExternalDictionaryRequest) {
	tflog.Debug(ctx, "Creating ClickHouse external dictionary", map[string]any{"cluster_id": req.ClusterId, "name": req.ExternalDictionary.GetName()})

	op, err := clickhousesdk.NewClusterClient(sdk).CreateExternalDictionary(ctx, req)
	if err != nil {
		diags.AddError(
			"Failed to create resource",
			fmt.Sprintf("Error while requesting API to create external dictionary %q in ClickHouse cluster '%s': %s", req.ExternalDictionary.GetName(), req.ClusterId, err.Error()),
		)
		return
	}

	if _, err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to create resource",
			fmt.Sprintf("Error while waiting for operation %q to create external dictionary %q in ClickHouse cluster '%s': %s", op.ID(), req.ExternalDictionary.GetName(), req.ClusterId, err.Error()),
		)
	}
}

func (c *ClickHouseAPI) DeleteExternalDictionary(ctx context.Context, sdk *ycsdk.SDK, diags *diag.Diagnostics, cid, name string) {
	tflog.Debug(ctx, "Deleting ClickHouse external dictionary", map[string]any{"cluster_id": cid, "name": name})

	op, err := clickhousesdk.NewClusterClient(sdk).DeleteExternalDictionary(ctx, &clickhouse.DeleteClusterExternalDictionaryRequest{
		ClusterId:              cid,
		ExternalDictionaryName: name,
	})
	if err != nil {
		diags.AddError(
			"Failed to delete resource",
			fmt.Sprintf("Error while requesting API to delete external dictionary %q from ClickHouse cluster '%s': %s", name, cid, err.Error()),
		)
		return
	}

	if _, err = op.Wait(ctx); err != nil {
		diags.AddError(
			"Failed to delete resource",
			fmt.Sprintf("Error while waiting for operation %q to delete external dictionary %q from ClickHouse cluster '%s': %s", op.ID(), name, cid, err.Error()),
		)
	}
}
