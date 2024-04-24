package yandex

import (
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
)

func Test_clickHouseHostsDiff(t *testing.T) {
	type args struct {
		currHosts   []*clickhouse.Host
		targetHosts []*clickhouse.HostSpec
	}
	tests := []struct {
		name     string
		args     args
		toDelete map[string][]string
		toAdd    map[string][]*clickhouse.HostSpec
		toUpdate map[string]*clickhouse.UpdateHostSpec
	}{
		{
			name: "add shard to another subnet",
			args: args{
				currHostsSingleHost,
				targetHostsTwoHosts,
			},
			toDelete: map[string][]string{},
			toAdd: map[string][]*clickhouse.HostSpec{
				"shard1": {
					{
						ZoneId:    "ru-central1-b",
						Type:      clickhouse.Host_CLICKHOUSE,
						ShardName: "shard1",
						SubnetId:  "subnet-b",
					},
				},
			},
			toUpdate: map[string]*clickhouse.UpdateHostSpec{},
		},
		{
			name: "add 2 host (one shard) in 1 subnet and 3 zk in 1 subnet",
			args: args{
				nil,
				[]*clickhouse.HostSpec{
					{
						Type:      clickhouse.Host_CLICKHOUSE,
						ZoneId:    "ru-central1-a",
						ShardName: "shard1",
						SubnetId:  "subnet-a",
					},
					{
						Type:      clickhouse.Host_CLICKHOUSE,
						ZoneId:    "ru-central1-a",
						ShardName: "shard1",
						SubnetId:  "subnet-a",
					},

					{
						Type:      clickhouse.Host_ZOOKEEPER,
						ZoneId:    "ru-central1-a",
						ShardName: "zk",
						SubnetId:  "subnet-a",
					},

					{
						Type:      clickhouse.Host_ZOOKEEPER,
						ZoneId:    "ru-central1-a",
						ShardName: "zk",
						SubnetId:  "subnet-a",
					},

					{
						Type:      clickhouse.Host_ZOOKEEPER,
						ZoneId:    "ru-central1-a",
						ShardName: "zk",
						SubnetId:  "subnet-a",
					},
				},
			},
			toDelete: map[string][]string{},
			toAdd: map[string][]*clickhouse.HostSpec{
				"shard1": {
					{
						ZoneId:    "ru-central1-a",
						Type:      clickhouse.Host_CLICKHOUSE,
						ShardName: "shard1",
						SubnetId:  "subnet-a",
					},
					{
						ZoneId:    "ru-central1-a",
						Type:      clickhouse.Host_CLICKHOUSE,
						ShardName: "shard1",
						SubnetId:  "subnet-a",
					},
				},
				"zk": {
					{
						Type:      clickhouse.Host_ZOOKEEPER,
						ZoneId:    "ru-central1-a",
						ShardName: "zk",
						SubnetId:  "subnet-a",
					},

					{
						Type:      clickhouse.Host_ZOOKEEPER,
						ZoneId:    "ru-central1-a",
						ShardName: "zk",
						SubnetId:  "subnet-a",
					},

					{
						Type:      clickhouse.Host_ZOOKEEPER,
						ZoneId:    "ru-central1-a",
						ShardName: "zk",
						SubnetId:  "subnet-a",
					},
				},
			},
			toUpdate: map[string]*clickhouse.UpdateHostSpec{},
		},
		{
			name: "add 2 host (two shards in 1 subnet) and 3 zk in 1 subnet",
			args: args{
				nil,
				[]*clickhouse.HostSpec{
					{
						Type:      clickhouse.Host_CLICKHOUSE,
						ZoneId:    "ru-central1-a",
						ShardName: "shard1",
						SubnetId:  "subnet-a",
					},
					{
						Type:      clickhouse.Host_CLICKHOUSE,
						ZoneId:    "ru-central1-a",
						ShardName: "shard2",
						SubnetId:  "subnet-a",
					},
					{
						Type:      clickhouse.Host_ZOOKEEPER,
						ZoneId:    "ru-central1-a",
						ShardName: "zk",
						SubnetId:  "subnet-a",
					},

					{
						Type:      clickhouse.Host_ZOOKEEPER,
						ZoneId:    "ru-central1-a",
						ShardName: "zk",
						SubnetId:  "subnet-a",
					},

					{
						Type:      clickhouse.Host_ZOOKEEPER,
						ZoneId:    "ru-central1-a",
						ShardName: "zk",
						SubnetId:  "subnet-a",
					},
				},
			},
			toDelete: map[string][]string{},
			toAdd: map[string][]*clickhouse.HostSpec{
				"shard1": {
					{
						ZoneId:    "ru-central1-a",
						Type:      clickhouse.Host_CLICKHOUSE,
						ShardName: "shard1",
						SubnetId:  "subnet-a",
					},
				},
				"shard2": {
					{
						ZoneId:    "ru-central1-a",
						Type:      clickhouse.Host_CLICKHOUSE,
						ShardName: "shard2",
						SubnetId:  "subnet-a",
					},
				},
				"zk": {
					{
						Type:      clickhouse.Host_ZOOKEEPER,
						ZoneId:    "ru-central1-a",
						ShardName: "zk",
						SubnetId:  "subnet-a",
					},

					{
						Type:      clickhouse.Host_ZOOKEEPER,
						ZoneId:    "ru-central1-a",
						ShardName: "zk",
						SubnetId:  "subnet-a",
					},

					{
						Type:      clickhouse.Host_ZOOKEEPER,
						ZoneId:    "ru-central1-a",
						ShardName: "zk",
						SubnetId:  "subnet-a",
					},
				},
			},
			toUpdate: map[string]*clickhouse.UpdateHostSpec{},
		},
		{
			name: "add 2 host (one shard in different subnets) and 3 zk in 2 subnets",
			args: args{
				nil,
				targetHostsZooIn2Subnets,
			},
			toDelete: map[string][]string{},
			toAdd: map[string][]*clickhouse.HostSpec{
				"shard1": {
					{
						ZoneId:    "ru-central1-a",
						Type:      clickhouse.Host_CLICKHOUSE,
						ShardName: "shard1",
						SubnetId:  "subnet-a",
					},
					{
						ZoneId:    "ru-central1-b",
						Type:      clickhouse.Host_CLICKHOUSE,
						ShardName: "shard1",
						SubnetId:  "subnet-b",
					},
				},
				"zk": {
					{
						Type:      clickhouse.Host_ZOOKEEPER,
						ZoneId:    "ru-central1-a",
						ShardName: "zk",
						SubnetId:  "subnet-a",
					},

					{
						Type:      clickhouse.Host_ZOOKEEPER,
						ZoneId:    "ru-central1-b",
						ShardName: "zk",
						SubnetId:  "subnet-b",
					},

					{
						Type:      clickhouse.Host_ZOOKEEPER,
						ZoneId:    "ru-central1-a",
						ShardName: "zk",
						SubnetId:  "subnet-a",
					},
				},
			},
			toUpdate: map[string]*clickhouse.UpdateHostSpec{},
		},
		{
			name: "exists 2 hosts(2 shards in 1 subnets) and 3 zk in 1 subnets. move to another subnet",
			args: args{
				[]*clickhouse.Host{
					{
						Name:      "first_host",
						ZoneId:    "ru-central1-a",
						Type:      clickhouse.Host_CLICKHOUSE,
						ShardName: "shard1",
						SubnetId:  "subnet-a",
					},
					{
						Name:      "second_host",
						ZoneId:    "ru-central1-a",
						Type:      clickhouse.Host_CLICKHOUSE,
						ShardName: "shard2",
						SubnetId:  "subnet-a",
					},
					{
						Type:      clickhouse.Host_ZOOKEEPER,
						ZoneId:    "ru-central1-a",
						ShardName: "zk",
						SubnetId:  "subnet-a",
					},

					{
						Type:      clickhouse.Host_ZOOKEEPER,
						ZoneId:    "ru-central1-a",
						ShardName: "zk",
						SubnetId:  "subnet-a",
					},

					{
						Type:      clickhouse.Host_ZOOKEEPER,
						ZoneId:    "ru-central1-a",
						ShardName: "zk",
						SubnetId:  "subnet-a",
					},
				},
				[]*clickhouse.HostSpec{
					{
						Type:      clickhouse.Host_CLICKHOUSE,
						ZoneId:    "ru-central1-b",
						ShardName: "shard1",
						SubnetId:  "subnet-b",
					},
					{
						Type:      clickhouse.Host_CLICKHOUSE,
						ZoneId:    "ru-central1-b",
						ShardName: "shard2",
						SubnetId:  "subnet-b",
					},

					{
						Type:      clickhouse.Host_ZOOKEEPER,
						ZoneId:    "ru-central1-b",
						ShardName: "zk",
						SubnetId:  "subnet-b",
					},

					{
						Type:      clickhouse.Host_ZOOKEEPER,
						ZoneId:    "ru-central1-b",
						ShardName: "zk",
						SubnetId:  "subnet-b",
					},

					{
						Type:      clickhouse.Host_ZOOKEEPER,
						ZoneId:    "ru-central1-b",
						ShardName: "zk",
						SubnetId:  "subnet-b",
					},
				},
			},
			toDelete: map[string][]string{
				"shard1": {"first_host"},
				"shard2": {"second_host"},
				"zk":     {"", "", ""},
			},
			toAdd: map[string][]*clickhouse.HostSpec{
				"shard1": {
					{
						ZoneId:    "ru-central1-b",
						Type:      clickhouse.Host_CLICKHOUSE,
						ShardName: "shard1",
						SubnetId:  "subnet-b",
					},
				},
				"shard2": {
					{
						ZoneId:    "ru-central1-b",
						Type:      clickhouse.Host_CLICKHOUSE,
						ShardName: "shard2",
						SubnetId:  "subnet-b",
					},
				},
				"zk": {
					{
						Type:      clickhouse.Host_ZOOKEEPER,
						ZoneId:    "ru-central1-b",
						ShardName: "zk",
						SubnetId:  "subnet-b",
					},

					{
						Type:      clickhouse.Host_ZOOKEEPER,
						ZoneId:    "ru-central1-b",
						ShardName: "zk",
						SubnetId:  "subnet-b",
					},

					{
						Type:      clickhouse.Host_ZOOKEEPER,
						ZoneId:    "ru-central1-b",
						ShardName: "zk",
						SubnetId:  "subnet-b",
					},
				},
			},
			toUpdate: map[string]*clickhouse.UpdateHostSpec{},
		},
		{
			name: "exists 2 hosts(2 shards in 1 subnets) and 3 zk in 1 subnets. full change subnets and add hosts",
			args: args{
				[]*clickhouse.Host{
					{
						Name:      "host1",
						ZoneId:    "ru-central1-a",
						Type:      clickhouse.Host_CLICKHOUSE,
						ShardName: "shard1",
						SubnetId:  "subnet-a",
					},
					{
						Name:      "host2",
						ZoneId:    "ru-central1-a",
						Type:      clickhouse.Host_CLICKHOUSE,
						ShardName: "shard2",
						SubnetId:  "subnet-a",
					},
					{
						Type:      clickhouse.Host_ZOOKEEPER,
						ZoneId:    "ru-central1-a",
						ShardName: "zk",
						SubnetId:  "subnet-a",
					},

					{
						Type:      clickhouse.Host_ZOOKEEPER,
						ZoneId:    "ru-central1-a",
						ShardName: "zk",
						SubnetId:  "subnet-a",
					},

					{
						Type:      clickhouse.Host_ZOOKEEPER,
						ZoneId:    "ru-central1-a",
						ShardName: "zk",
						SubnetId:  "subnet-a",
					},
				},
				[]*clickhouse.HostSpec{
					{
						Type:      clickhouse.Host_CLICKHOUSE,
						ZoneId:    "ru-central1-a",
						ShardName: "shard1",
						SubnetId:  "subnet-a",
					},
					{
						Type:      clickhouse.Host_CLICKHOUSE,
						ZoneId:    "ru-central1-a",
						ShardName: "shard2",
						SubnetId:  "subnet-a",
					},
					{
						Type:      clickhouse.Host_CLICKHOUSE,
						ZoneId:    "ru-central1-b",
						ShardName: "shard1",
						SubnetId:  "subnet-b",
					},
					{
						Type:      clickhouse.Host_CLICKHOUSE,
						ZoneId:    "ru-central1-b",
						ShardName: "shard2",
						SubnetId:  "subnet-b",
					},

					{
						Type:      clickhouse.Host_ZOOKEEPER,
						ZoneId:    "ru-central1-a",
						ShardName: "zk",
						SubnetId:  "subnet-a",
					},

					{
						Type:      clickhouse.Host_ZOOKEEPER,
						ZoneId:    "ru-central1-b",
						ShardName: "zk",
						SubnetId:  "subnet-b",
					},

					{
						Type:      clickhouse.Host_ZOOKEEPER,
						ZoneId:    "ru-central1-d",
						ShardName: "zk",
						SubnetId:  "subnet-c",
					},
				},
			},
			toDelete: map[string][]string{
				"zk": {"", ""},
			},
			toAdd: map[string][]*clickhouse.HostSpec{
				"shard1": {
					{
						ZoneId:    "ru-central1-b",
						Type:      clickhouse.Host_CLICKHOUSE,
						ShardName: "shard1",
						SubnetId:  "subnet-b",
					},
				},
				"shard2": {
					{
						ZoneId:    "ru-central1-b",
						Type:      clickhouse.Host_CLICKHOUSE,
						ShardName: "shard2",
						SubnetId:  "subnet-b",
					},
				},
				"zk": {
					{
						Type:      clickhouse.Host_ZOOKEEPER,
						ZoneId:    "ru-central1-b",
						ShardName: "zk",
						SubnetId:  "subnet-b",
					},

					{
						Type:      clickhouse.Host_ZOOKEEPER,
						ZoneId:    "ru-central1-d",
						ShardName: "zk",
						SubnetId:  "subnet-c",
					},
				},
			},
			toUpdate: map[string]*clickhouse.UpdateHostSpec{},
		},
		{
			name: "exist 1 host, add 1 shard in another subnet and 3 zk in different subnet",
			args: args{
				currHostsSingleHost,
				targetHostsHA,
			},
			toDelete: map[string][]string{},
			toAdd: map[string][]*clickhouse.HostSpec{
				"shard1": {
					{
						ZoneId:    "ru-central1-b",
						Type:      clickhouse.Host_CLICKHOUSE,
						ShardName: "shard1",
						SubnetId:  "subnet-b",
					},
				},
				"zk": {
					{
						Type:      clickhouse.Host_ZOOKEEPER,
						ZoneId:    "ru-central1-a",
						ShardName: "zk",
						SubnetId:  "subnet-a",
					},

					{
						Type:      clickhouse.Host_ZOOKEEPER,
						ZoneId:    "ru-central1-b",
						ShardName: "zk",
						SubnetId:  "subnet-b",
					},

					{
						Type:      clickhouse.Host_ZOOKEEPER,
						ZoneId:    "ru-central1-d",
						ShardName: "zk",
						SubnetId:  "subnet-c",
					},
				},
			},
			toUpdate: map[string]*clickhouse.UpdateHostSpec{},
		},
		{
			name: "simple case remove host",
			args: args{
				currentHostsTwoHosts,
				targetHostsSingleHost,
			},
			toDelete: map[string][]string{"shard1": {"host_a"}},
			toAdd:    map[string][]*clickhouse.HostSpec{},
			toUpdate: map[string]*clickhouse.UpdateHostSpec{},
		},
		{
			name: "update assign_public_ip",
			args: args{
				currentHostsTwoHosts,
				targetHostsTwoHostsPublicIP,
			},
			toDelete: map[string][]string{},
			toAdd:    map[string][]*clickhouse.HostSpec{},
			toUpdate: map[string]*clickhouse.UpdateHostSpec{
				"host_a": {HostName: "host_a", AssignPublicIp: &wrapperspb.BoolValue{Value: true}, UpdateMask: &field_mask.FieldMask{Paths: []string{"assign_public_ip"}}},
				"host_b": {HostName: "host_b", AssignPublicIp: &wrapperspb.BoolValue{Value: true}, UpdateMask: &field_mask.FieldMask{Paths: []string{"assign_public_ip"}}},
			},
		},
		{
			name: "TestAccMDBClickHouseCluster_sharded",
			args: args{
				[]*clickhouse.Host{
					{Name: "host_a", ZoneId: "ru-central1-a", Type: clickhouse.Host_CLICKHOUSE, SubnetId: "sub-a", AssignPublicIp: false, ShardName: "shard1"},
					{Name: "host_b", ZoneId: "ru-central1-b", Type: clickhouse.Host_CLICKHOUSE, SubnetId: "sub-b", AssignPublicIp: false, ShardName: "shard2"},
				},
				[]*clickhouse.HostSpec{
					{ZoneId: "ru-central1-a", Type: clickhouse.Host_CLICKHOUSE, SubnetId: "sub-a", AssignPublicIp: true, ShardName: "shard1"},
					{ZoneId: "ru-central1-d", Type: clickhouse.Host_CLICKHOUSE, SubnetId: "sub-d", AssignPublicIp: true, ShardName: "shard3"},
				},
			},
			toDelete: map[string][]string{"shard2": {"host_b"}},
			toAdd: map[string][]*clickhouse.HostSpec{
				"shard3": {{ZoneId: "ru-central1-d", Type: clickhouse.Host_CLICKHOUSE, SubnetId: "sub-d", AssignPublicIp: true, ShardName: "shard3"}}},
			toUpdate: map[string]*clickhouse.UpdateHostSpec{
				"host_a": {HostName: "host_a", AssignPublicIp: &wrapperspb.BoolValue{Value: true}, UpdateMask: &field_mask.FieldMask{Paths: []string{"assign_public_ip"}}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotToDelete, gotToAdd, gotToUpdate := clickHouseHostsDiff(tt.args.currHosts, tt.args.targetHosts)
			assert.EqualValuesf(t, tt.toDelete, gotToDelete, "clickHouseDeleteHostsDiff(%v, %v)", tt.args.currHosts, tt.args.targetHosts)
			assert.EqualValuesf(t, tt.toAdd, gotToAdd, "clickHouseAddHostsDiff(%v, %v)", tt.args.currHosts, tt.args.targetHosts)
			assert.EqualValuesf(t, tt.toUpdate, gotToUpdate, "clickHouseUpdateHostsDiff(%v, %v)", tt.args.currHosts, tt.args.targetHosts)
		})
	}
}

var currHostsSingleHost = []*clickhouse.Host{
	{
		Name:      "host1",
		ZoneId:    "ru-central1-a",
		Type:      clickhouse.Host_CLICKHOUSE,
		ShardName: "shard1",
		SubnetId:  "subnet-a",
	},
}

var currentHostsTwoHosts = []*clickhouse.Host{
	{
		Name:      "host_a",
		ZoneId:    "ru-central1-a",
		Type:      clickhouse.Host_CLICKHOUSE,
		ShardName: "shard1",
		SubnetId:  "subnet-a",
	},
	{
		Name:      "host_b",
		ZoneId:    "ru-central1-b",
		Type:      clickhouse.Host_CLICKHOUSE,
		ShardName: "shard1",
		SubnetId:  "subnet-b",
	},
}

var targetHostsSingleHost = []*clickhouse.HostSpec{
	{
		ZoneId:    "ru-central1-b",
		Type:      clickhouse.Host_CLICKHOUSE,
		ShardName: "shard1",
		SubnetId:  "subnet-b",
	},
}

var targetHostsTwoHosts = []*clickhouse.HostSpec{
	{
		ZoneId:    "ru-central1-a",
		Type:      clickhouse.Host_CLICKHOUSE,
		ShardName: "shard1",
		SubnetId:  "subnet-a",
	},
	{
		ZoneId:    "ru-central1-b",
		Type:      clickhouse.Host_CLICKHOUSE,
		ShardName: "shard1",
		SubnetId:  "subnet-b",
	},
}

var targetHostsTwoHostsPublicIP = []*clickhouse.HostSpec{
	{
		ZoneId:         "ru-central1-a",
		Type:           clickhouse.Host_CLICKHOUSE,
		ShardName:      "shard1",
		SubnetId:       "subnet-a",
		AssignPublicIp: true,
	},
	{
		ZoneId:         "ru-central1-b",
		Type:           clickhouse.Host_CLICKHOUSE,
		ShardName:      "shard1",
		SubnetId:       "subnet-b",
		AssignPublicIp: true,
	},
}

var targetHostsHA = []*clickhouse.HostSpec{
	{
		Type:      clickhouse.Host_CLICKHOUSE,
		ZoneId:    "ru-central1-a",
		ShardName: "shard1",
		SubnetId:  "subnet-a",
	},
	{
		Type:      clickhouse.Host_CLICKHOUSE,
		ZoneId:    "ru-central1-b",
		ShardName: "shard1",
		SubnetId:  "subnet-b",
	},

	{
		Type:      clickhouse.Host_ZOOKEEPER,
		ZoneId:    "ru-central1-a",
		ShardName: "zk",
		SubnetId:  "subnet-a",
	},

	{
		Type:      clickhouse.Host_ZOOKEEPER,
		ZoneId:    "ru-central1-b",
		ShardName: "zk",
		SubnetId:  "subnet-b",
	},

	{
		Type:      clickhouse.Host_ZOOKEEPER,
		ZoneId:    "ru-central1-d",
		ShardName: "zk",
		SubnetId:  "subnet-c",
	},
}

var targetHostsZooIn2Subnets = []*clickhouse.HostSpec{
	{
		Type:      clickhouse.Host_CLICKHOUSE,
		ZoneId:    "ru-central1-a",
		ShardName: "shard1",
		SubnetId:  "subnet-a",
	},
	{
		Type:      clickhouse.Host_CLICKHOUSE,
		ZoneId:    "ru-central1-b",
		ShardName: "shard1",
		SubnetId:  "subnet-b",
	},

	{
		Type:      clickhouse.Host_ZOOKEEPER,
		ZoneId:    "ru-central1-a",
		ShardName: "zk",
		SubnetId:  "subnet-a",
	},

	{
		Type:      clickhouse.Host_ZOOKEEPER,
		ZoneId:    "ru-central1-b",
		ShardName: "zk",
		SubnetId:  "subnet-b",
	},

	{
		Type:      clickhouse.Host_ZOOKEEPER,
		ZoneId:    "ru-central1-a",
		ShardName: "zk",
		SubnetId:  "subnet-a",
	},
}
