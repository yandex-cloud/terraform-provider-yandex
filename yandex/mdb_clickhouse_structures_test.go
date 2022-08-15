package yandex

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/protobuf/types/known/wrapperspb"

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
			name: "simple case add host",
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
			name: "add host and zk",
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
						ZoneId:    "ru-central1-c",
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
					{ZoneId: "ru-central1-c", Type: clickhouse.Host_CLICKHOUSE, SubnetId: "sub-c", AssignPublicIp: true, ShardName: "shard3"},
				},
			},
			toDelete: map[string][]string{"shard2": {"host_b"}},
			toAdd: map[string][]*clickhouse.HostSpec{
				"shard3": {{ZoneId: "ru-central1-c", Type: clickhouse.Host_CLICKHOUSE, SubnetId: "sub-c", AssignPublicIp: true, ShardName: "shard3"}}},
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
		ZoneId:    "ru-central1-c",
		ShardName: "zk",
		SubnetId:  "subnet-c",
	},
}
