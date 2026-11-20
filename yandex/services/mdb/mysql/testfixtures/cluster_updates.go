// Package testfixtures provides offline MySQL API fixtures for resource tests.
package testfixtures

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"
	msconfig "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1/config"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	ycsdk "github.com/yandex-cloud/go-sdk/v2"
	"github.com/yandex-cloud/go-sdk/v2/credentials"
	"github.com/yandex-cloud/go-sdk/v2/pkg/endpoints"
	"github.com/yandex-cloud/go-sdk/v2/pkg/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// ClusterUpdates records real SDK calls, including operation polling, without
// contacting a cloud endpoint. Unexpected calls fail the resource operation.
type ClusterUpdates struct {
	Requests       []*mysql.UpdateClusterRequest
	Events         []string
	FailOperation  string
	Cluster        *mysql.Cluster
	Hosts          []*mysql.Host
	UpgradedConfig *mysql.ClusterConfig
	GetError       error
	pending        *mysql.Cluster
}

func (s *ClusterUpdates) SDK(t testing.TB) *ycsdk.SDK {
	t.Helper()
	sdk, err := ycsdk.Build(context.Background(),
		options.WithCredentials(credentials.NoAuthentication()),
		options.WithEndpointsResolver(endpoints.NewSingleEndpointResolver("unused.invalid:443")),
		options.WithPlaintext(),
		options.WithCustomDialOptions(grpc.WithUnaryInterceptor(func(ctx context.Context, method string, req, reply interface{}, _ *grpc.ClientConn, _ grpc.UnaryInvoker, _ ...grpc.CallOption) error {
			var response proto.Message
			switch method {
			case "/yandex.cloud.mdb.mysql.v1.ClusterService/Update":
				request := req.(*mysql.UpdateClusterRequest)
				s.Requests = append(s.Requests, proto.Clone(request).(*mysql.UpdateClusterRequest))
				id := "settings"
				if slices.Contains(request.GetUpdateMask().GetPaths(), "config_spec.version") {
					id = "version"
					if len(request.UpdateMask.Paths) != 1 || request.UpdateMask.Paths[0] != "config_spec.version" || request.ConfigSpec.MysqlConfig != nil {
						return status.Error(codes.InvalidArgument, "version upgrade must be a separate request")
					}
				}
				s.Events = append(s.Events, "update:"+id)
				if id == "settings" {
					var err error
					s.pending, err = s.applySettings(request)
					if err != nil {
						return err
					}
				}
				metadata, err := anypb.New(&mysql.UpdateClusterMetadata{ClusterId: request.ClusterId})
				require.NoError(t, err)
				response = &operation.Operation{Id: id, Metadata: metadata}
			case "/yandex.cloud.operation.OperationService/Get":
				id := req.(*operation.GetOperationRequest).OperationId
				s.Events = append(s.Events, "wait:"+id)
				op := &operation.Operation{Id: id, Done: true}
				if s.FailOperation == id {
					op.Result = &operation.Operation_Error{Error: status.New(codes.FailedPrecondition, "test operation failure").Proto()}
				} else {
					if id == "version" && s.UpgradedConfig != nil {
						s.Cluster.Config = proto.Clone(s.UpgradedConfig).(*mysql.ClusterConfig)
					}
					if id == "settings" {
						s.Cluster = s.pending
					}
					value, err := anypb.New(s.Cluster)
					require.NoError(t, err)
					op.Result = &operation.Operation_Response{Response: value}
				}
				response = op
			case "/yandex.cloud.mdb.mysql.v1.ClusterService/Get":
				s.Events = append(s.Events, "get")
				if s.GetError != nil {
					return s.GetError
				}
				response = s.Cluster
			case "/yandex.cloud.mdb.mysql.v1.ClusterService/ListHosts":
				response = &mysql.ListClusterHostsResponse{Hosts: s.Hosts}
			default:
				return fmt.Errorf("unexpected MySQL test RPC: %s", method)
			}
			proto.Merge(reply.(proto.Message), response)
			return nil
		})),
	)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, sdk.Shutdown(context.Background())) })
	return sdk
}

// Model the settings/description updates used by these tests. The request mask
// is applied to a copy, and the result becomes visible only when polling succeeds.
func (s *ClusterUpdates) applySettings(request *mysql.UpdateClusterRequest) (*mysql.Cluster, error) {
	if !request.UpdateMask.IsValid(request) {
		return nil, status.Error(codes.InvalidArgument, "invalid update mask")
	}
	next := proto.Clone(s.Cluster).(*mysql.Cluster)
	for _, path := range request.UpdateMask.Paths {
		if path == "description" {
			next.Description = request.Description
			continue
		}
		name, ok := strings.CutPrefix(path, "config_spec.mysql_config_9_7.")
		if !ok {
			return nil, fmt.Errorf("unexpected test update path: %s", path)
		}
		if next.GetConfig().GetMysqlConfig_9_7() == nil {
			return nil, status.Error(codes.FailedPrecondition, "settings update before upgrade completed")
		}
		set := next.Config.GetMysqlConfig_9_7()
		if set.UserConfig == nil {
			set.UserConfig = &msconfig.MysqlConfig9_7{}
		}
		source := proto.Clone(request.ConfigSpec.GetMysqlConfig_9_7()).(*msconfig.MysqlConfig9_7).ProtoReflect()
		target := set.UserConfig.ProtoReflect()
		field := source.Descriptor().Fields().ByName(protoreflect.Name(name))
		if source.Has(field) {
			target.Set(field, source.Get(field))
		} else {
			target.Clear(field)
		}
	}
	if proto.Equal(next, s.Cluster) {
		return nil, status.Error(codes.FailedPrecondition, "no changes detected")
	}
	return next, nil
}

// MySQLCluster returns a complete read response with no user settings.
func MySQLCluster() *mysql.Cluster {
	return &mysql.Cluster{
		Id: "test-id", Name: "test-cluster", FolderId: "test-folder", NetworkId: "test-network", Environment: mysql.Cluster_PRESTABLE,
		MaintenanceWindow: &mysql.MaintenanceWindow{Policy: &mysql.MaintenanceWindow_Anytime{Anytime: &mysql.AnytimeMaintenanceWindow{}}},
		Config: &mysql.ClusterConfig{
			Version:                "9.7",
			BackupRetainPeriodDays: wrapperspb.Int64(7),
			Resources:              &mysql.Resources{ResourcePresetId: "s1.micro", DiskTypeId: "network-ssd", DiskSize: 10 * 1024 * 1024 * 1024},
			MysqlConfig:            &mysql.ClusterConfig_MysqlConfig_9_7{MysqlConfig_9_7: &msconfig.MysqlConfigSet9_7{UserConfig: &msconfig.MysqlConfig9_7{}}},
		},
	}
}

// UpgradeCluster models the server-side 8.4 -> 9.7 migration independently of
// the provider's conversion code. Redo capacity is twice the old per-file size.
// An operation failure leaves the last successfully applied state intact.
func UpgradeCluster() *ClusterUpdates {
	cluster := MySQLCluster()
	cluster.Config.Version = "8.4"
	cluster.Config.MysqlConfig = &mysql.ClusterConfig_MysqlConfig_8_4{MysqlConfig_8_4: &msconfig.MysqlConfigSet8_4{UserConfig: &msconfig.MysqlConfig8_4{
		MaxConnections: wrapperspb.Int64(100), SlaveParallelWorkers: wrapperspb.Int64(4),
		RplSemiSyncMasterWaitForSlaveCount: wrapperspb.Int64(1), InnodbLogFileSize: wrapperspb.Int64(134217728),
	}}}
	upgraded := proto.Clone(cluster.Config).(*mysql.ClusterConfig)
	upgraded.Version = "9.7"
	upgraded.MysqlConfig = &mysql.ClusterConfig_MysqlConfig_9_7{MysqlConfig_9_7: &msconfig.MysqlConfigSet9_7{UserConfig: &msconfig.MysqlConfig9_7{
		MaxConnections: wrapperspb.Int64(100), ReplicaParallelWorkers: wrapperspb.Int64(4),
		RplSemiSyncSourceWaitForReplicaCount: wrapperspb.Int64(1), InnodbRedoLogCapacity: wrapperspb.Int64(268435456),
	}}}
	return &ClusterUpdates{Cluster: cluster, UpgradedConfig: upgraded}
}

// UpgradeWithDefaultRedoCapacity models the migration when innodb_log_file_size
// was never set on 8.4. The server adds the capacity to user_config on 9.7.
func UpgradeWithDefaultRedoCapacity(maxConnections *wrapperspb.Int64Value) *ClusterUpdates {
	api := UpgradeCluster()
	api.Cluster.Config.GetMysqlConfig_8_4().UserConfig = &msconfig.MysqlConfig8_4{MaxConnections: maxConnections}
	api.UpgradedConfig.GetMysqlConfig_9_7().UserConfig = &msconfig.MysqlConfig9_7{
		MaxConnections: maxConnections, InnodbRedoLogCapacity: wrapperspb.Int64(512 << 20),
	}
	return api
}

func UpgradeSettings(version string) map[string]string {
	if version == "8.4" {
		return map[string]string{
			"max_connections": "100", "slave_parallel_workers": "4",
			"rpl_semi_sync_master_wait_for_slave_count": "1", "innodb_log_file_size": "134217728",
		}
	}
	return map[string]string{
		"max_connections": "100", "replica_parallel_workers": "4",
		"rpl_semi_sync_source_wait_for_replica_count": "1", "innodb_redo_log_capacity": "268435456",
	}
}

// UpgradeCase describes a settings update after the server has migrated 8.4 to 9.7.
type UpgradeCase struct {
	Name         string
	Settings     map[string]string
	AlreadyReset bool
	Description  string
	GetError     error
	WantPaths    []string
}

func UpgradeCases() []UpgradeCase {
	unchanged := UpgradeSettings("9.7")
	changed := UpgradeSettings("9.7")
	changed["max_connections"] = "200"
	reset := map[string]string{"max_connections": "100"}
	return []UpgradeCase{
		{Name: "renames only", Settings: unchanged},
		{
			Name: "changed value", Settings: changed,
			WantPaths: []string{"config_spec.mysql_config_9_7.max_connections"},
		},
		{
			Name: "reset", Settings: reset,
			WantPaths: []string{
				"config_spec.mysql_config_9_7.replica_parallel_workers",
				"config_spec.mysql_config_9_7.rpl_semi_sync_source_wait_for_replica_count",
				"config_spec.mysql_config_9_7.innodb_redo_log_capacity",
			},
		},
		{Name: "already reset", Settings: reset, AlreadyReset: true},
		{
			Name: "description", Settings: unchanged, Description: "updated description",
			WantPaths: []string{"description"},
		},
		{Name: "read failure", Settings: unchanged, GetError: status.Error(codes.Unavailable, "test read failure")},
	}
}
