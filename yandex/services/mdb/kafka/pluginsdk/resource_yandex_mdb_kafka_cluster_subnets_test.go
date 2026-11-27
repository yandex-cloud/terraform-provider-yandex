package pluginsdk_test

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/kafka/v1"
	sdk "github.com/yandex-cloud/go-sdk/v2"
	"github.com/yandex-cloud/go-sdk/v2/credentials"
	"github.com/yandex-cloud/go-sdk/v2/pkg/endpoints"
	"github.com/yandex-cloud/go-sdk/v2/pkg/options"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	svc_kafka "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/kafka/pluginsdk"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestKafkaClusterSubnetIDsRefresh(t *testing.T) {
	for _, tc := range []struct {
		name   string
		before []string
		want   []interface{}
	}{
		{"import", nil, []interface{}{"subnet-a", "subnet-d", "subnet-e"}},
		{"preserves_configured_order", []string{"subnet-e", "subnet-a", "subnet-d"}, []interface{}{"subnet-e", "subnet-a", "subnet-d"}},
		{"external_migration", []string{"subnet-d", "subnet-b", "subnet-a"}, []interface{}{"subnet-d", "subnet-a", "subnet-e"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := svc_kafka.ResourceYandexMDBKafkaCluster()
			state, diags := r.RefreshWithoutUpgrade(context.Background(), kafkaSubnetState(tc.before), kafkaSubnetTestConfig(t))
			require.False(t, diags.HasError(), "%v", diags)
			require.Equal(t, tc.want, r.Data(state).Get("subnet_ids"))
			require.Equal(t, "5", state.Attributes["host.#"], "all host pages must be read")
		})
	}
}

func TestKafkaClusterDataSourceSubnetIDs(t *testing.T) {
	r := svc_kafka.DataSourceYandexMDBKafkaCluster()
	d := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{"cluster_id": "test-cluster"})
	require.NoError(t, r.Read(d, kafkaSubnetTestConfig(t)))
	require.Equal(t, []interface{}{"subnet-a", "subnet-d", "subnet-e"}, d.Get("subnet_ids"))
}

func TestKafkaClusterSubnetIDsAutoSelectionRefreshDiff(t *testing.T) {
	resource := svc_kafka.ResourceYandexMDBKafkaCluster()
	planResource := &schema.Resource{Schema: map[string]*schema.Schema{
		"subnet_ids": resource.Schema["subnet_ids"],
	}}
	for _, tc := range []struct {
		name   string
		before []string
		config cty.Value
	}{
		{"omitted", nil, cty.NullVal(cty.List(cty.String))},
		{"explicit_empty", []string{}, cty.ListValEmpty(cty.String)},
		{"explicit_empty_after_import", nil, cty.ListValEmpty(cty.String)},
		{"switch_to_auto_selection", []string{"subnet-a", "subnet-d", "subnet-e"}, cty.ListValEmpty(cty.String)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state, diags := resource.RefreshWithoutUpgrade(context.Background(), kafkaSubnetState(tc.before), kafkaSubnetTestConfig(t))
			require.False(t, diags.HasError(), "%v", diags)
			require.Equal(t, []interface{}{"subnet-a", "subnet-d", "subnet-e"}, resource.Data(state).Get("subnet_ids"))

			// Plan with the production subnet schema after the actual resource Read.
			// Other cluster fields are excluded to isolate subnet changes.
			// The SDK's PlanResourceChange passes the raw configuration via state.
			state.RawConfig = cty.ObjectVal(map[string]cty.Value{"subnet_ids": tc.config})
			config := terraform.NewResourceConfigShimmed(state.RawConfig, planResource.CoreConfigSchema())
			diff, err := planResource.Diff(context.Background(), state, config, nil)
			require.NoError(t, err)
			require.True(t, diff == nil || diff.Empty(), "automatic subnet selection must not plan an update: %#v", diff)
		})
	}
}

func TestKafkaClusterSubnetIDsDiff(t *testing.T) {
	// Exercise SDK planning with the production attribute schema, independently
	// of unrelated cluster attributes and their defaults.
	r := &schema.Resource{Schema: map[string]*schema.Schema{
		"subnet_ids": svc_kafka.ResourceYandexMDBKafkaCluster().Schema["subnet_ids"],
	}}
	ids := func(values ...string) cty.Value {
		result := make([]cty.Value, len(values))
		for i, value := range values {
			result[i] = cty.StringVal(value)
		}
		if len(result) == 0 {
			return cty.ListValEmpty(cty.String)
		}
		return cty.ListVal(result)
	}
	for _, tc := range []struct {
		name     string
		before   []string
		config   cty.Value
		wantDiff bool
	}{
		{"same_order", []string{"a", "d", "e"}, ids("a", "d", "e"), false},
		{"different_order_after_import", []string{"a", "d", "e"}, ids("e", "a", "d"), false},
		{"omitted", []string{"a", "d", "e"}, cty.NullVal(cty.List(cty.String)), false},
		{"replacement", []string{"a", "b", "d"}, ids("a", "d", "e"), true},
		{"addition", []string{"a", "d"}, ids("a", "d", "e"), true},
		{"removal", []string{"a", "d", "e"}, ids("a", "d"), true},
		{"explicit_empty", []string{"a", "d", "e"}, ids(), false},
		{"unknown_list", []string{"a", "d", "e"}, cty.UnknownVal(cty.List(cty.String)), true},
		{"unknown_element", []string{"a", "d", "e"}, cty.ListVal([]cty.Value{cty.StringVal("a"), cty.StringVal("d"), cty.UnknownVal(cty.String)}), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state := kafkaSubnetState(tc.before)
			state.RawConfig = cty.ObjectVal(map[string]cty.Value{"subnet_ids": tc.config})
			config := terraform.NewResourceConfigShimmed(state.RawConfig, r.CoreConfigSchema())
			diff, err := r.Diff(context.Background(), state, config, nil)
			require.NoError(t, err)
			require.Equal(t, tc.wantDiff, diff != nil && !diff.Empty(), "%#v", diff)
		})
	}
}

func kafkaSubnetState(ids []string) *terraform.InstanceState {
	state := &terraform.InstanceState{ID: "test-cluster", Attributes: map[string]string{"id": "test-cluster"}}
	if ids != nil {
		state.Attributes["subnet_ids.#"] = strconv.Itoa(len(ids))
		for i, id := range ids {
			state.Attributes[fmt.Sprintf("subnet_ids.%d", i)] = id
		}
	}
	return state
}

type kafkaSubnetClusterAPI struct {
	kafka.UnimplementedClusterServiceServer
}

func (*kafkaSubnetClusterAPI) Get(_ context.Context, req *kafka.GetClusterRequest) (*kafka.Cluster, error) {
	if req.ClusterId != "test-cluster" {
		return nil, fmt.Errorf("unexpected cluster ID: %q", req.ClusterId)
	}
	return &kafka.Cluster{
		Id: "test-cluster", Name: "test-cluster", FolderId: "test-folder", NetworkId: "test-network",
		Environment: kafka.Cluster_PRODUCTION,
		Config: &kafka.ConfigSpec{
			Version: "4.1", ZoneId: []string{"ru-central1-a", "ru-central1-d", "ru-central1-e"}, BrokersCount: wrapperspb.Int64(1),
			Kafka: &kafka.ConfigSpec_Kafka{Resources: &kafka.Resources{
				ResourcePresetId: "s2.micro", DiskTypeId: "network-ssd", DiskSize: 16 << 30,
			}},
		},
	}, nil
}

func (*kafkaSubnetClusterAPI) ListHosts(_ context.Context, req *kafka.ListClusterHostsRequest) (*kafka.ListClusterHostsResponse, error) {
	if req.ClusterId != "test-cluster" {
		return nil, fmt.Errorf("unexpected cluster ID: %q", req.ClusterId)
	}
	// Host-name order deliberately differs from subnet order; controllers share
	// subnets with brokers, and a host without a subnet must not add an empty ID.
	switch req.PageToken {
	case "":
		return &kafka.ListClusterHostsResponse{Hosts: []*kafka.Host{
			{Name: "host-a", SubnetId: "subnet-e", Role: kafka.Host_KAFKA},
			{Name: "host-b", SubnetId: "subnet-a", Role: kafka.Host_KAFKA},
		}, NextPageToken: "page-2"}, nil
	case "page-2":
		return &kafka.ListClusterHostsResponse{Hosts: []*kafka.Host{
			{Name: "host-c", SubnetId: "subnet-a", Role: kafka.Host_KRAFT},
			{Name: "host-d", SubnetId: "subnet-d", Role: kafka.Host_KAFKA},
			{Name: "host-e", Role: kafka.Host_ZOOKEEPER},
		}}, nil
	default:
		return nil, fmt.Errorf("unexpected page token: %q", req.PageToken)
	}
}

type kafkaSubnetTopicAPI struct {
	kafka.UnimplementedTopicServiceServer
}

func (*kafkaSubnetTopicAPI) List(context.Context, *kafka.ListTopicsRequest) (*kafka.ListTopicsResponse, error) {
	return &kafka.ListTopicsResponse{}, nil
}

type kafkaSubnetUserAPI struct {
	kafka.UnimplementedUserServiceServer
}

func (*kafkaSubnetUserAPI) List(context.Context, *kafka.ListUsersRequest) (*kafka.ListUsersResponse, error) {
	return &kafka.ListUsersResponse{}, nil
}

func kafkaSubnetTestConfig(t *testing.T) *providerconfig.Config {
	t.Helper()
	return kafkaClusterTestConfig(t, &kafkaSubnetClusterAPI{})
}

func kafkaClusterTestConfig(t *testing.T, cluster kafka.ClusterServiceServer) *providerconfig.Config {
	t.Helper()
	listener := &kafkaSubnetListener{incoming: make(chan net.Conn), closed: make(chan struct{})}
	server := grpc.NewServer()
	kafka.RegisterClusterServiceServer(server, cluster)
	kafka.RegisterTopicServiceServer(server, &kafkaSubnetTopicAPI{})
	kafka.RegisterUserServiceServer(server, &kafkaSubnetUserAPI{})
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(func() { server.Stop(); _ = listener.Close() })
	client, err := sdk.Build(context.Background(),
		options.WithCredentials(credentials.NoAuthentication()),
		options.WithEndpointsResolver(endpoints.NewSingleEndpointResolver("passthrough:///kafka-subnet-test")),
		options.WithPlaintext(),
		options.WithCustomDialOptions(grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return listener.DialContext(ctx)
		})),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Shutdown(context.Background()) })
	return &providerconfig.Config{SDK: client}
}

// An in-memory listener keeps these tests independent of network access.
type kafkaSubnetListener struct {
	incoming chan net.Conn
	closed   chan struct{}
	once     sync.Once
}

func (l *kafkaSubnetListener) Accept() (net.Conn, error) {
	select {
	case conn := <-l.incoming:
		return conn, nil
	case <-l.closed:
		return nil, net.ErrClosed
	}
}

func (l *kafkaSubnetListener) Close() error   { l.once.Do(func() { close(l.closed) }); return nil }
func (l *kafkaSubnetListener) Addr() net.Addr { return &net.TCPAddr{} }

func (l *kafkaSubnetListener) DialContext(ctx context.Context) (net.Conn, error) {
	client, server := net.Pipe()
	select {
	case l.incoming <- server:
		return client, nil
	case <-ctx.Done():
		_ = client.Close()
		_ = server.Close()
		return nil, ctx.Err()
	case <-l.closed:
		_ = client.Close()
		_ = server.Close()
		return nil, net.ErrClosed
	}
}
