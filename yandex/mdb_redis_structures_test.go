package yandex

import (
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/redis/v1"
	"google.golang.org/genproto/protobuf/field_mask"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetSentinelHosts(t *testing.T) {
	diskTypeId := ""
	publicIPFlags := []*bool{nil}
	replicaPriorities := []*int{nil}
	expected := `
  host {
  	zone      = "ru-central1-c"
	subnet_id = "${yandex_vpc_subnet.foo.id}"
	
	
  }
`

	actual := getSentinelHosts(diskTypeId, publicIPFlags, replicaPriorities)
	require.Equal(t, expected, actual)
}

func TestRedisHostsDiff(t *testing.T) {
	cases := []struct {
		sharded          bool
		name             string
		currHosts        []*redis.Host
		targetHosts      []*redis.HostSpec
		expectedError    string
		expectedToDelete map[string][]string
		expectedtoUpdate map[string][]*HostUpdateInfo
		expectedToAdd    map[string][]*redis.HostSpec
	}{
		{
			name: "0 add, 0 update, 0 delete",
			currHosts: []*redis.Host{
				{
					Name:            "fqdn1",
					ShardName:       "shard1",
					SubnetId:        "subnet1",
					ReplicaPriority: &wrappers.Int64Value{Value: 100},
					AssignPublicIp:  false,
				},
			},
			targetHosts: []*redis.HostSpec{
				{
					ShardName:       "shard1",
					SubnetId:        "subnet1",
					ReplicaPriority: &wrappers.Int64Value{Value: 100},
					AssignPublicIp:  false,
				},
			},
			expectedToAdd:    map[string][]*redis.HostSpec{},
			expectedtoUpdate: map[string][]*HostUpdateInfo{},
			expectedToDelete: map[string][]string{},
		},
		{
			name: "0 add, 1 update (ip), 0 delete",
			currHosts: []*redis.Host{
				{
					Name:            "fqdn1",
					ShardName:       "shard1",
					SubnetId:        "subnet1",
					ReplicaPriority: &wrappers.Int64Value{Value: 100},
					AssignPublicIp:  false,
				},
			},
			targetHosts: []*redis.HostSpec{
				{
					ShardName:       "shard1",
					SubnetId:        "subnet1",
					ReplicaPriority: &wrappers.Int64Value{Value: 100},
					AssignPublicIp:  true,
				},
			},
			expectedToAdd: map[string][]*redis.HostSpec{},
			expectedtoUpdate: map[string][]*HostUpdateInfo{
				"shard1": {
					{
						HostName:        "fqdn1",
						AssignPublicIp:  true,
						ReplicaPriority: &wrappers.Int64Value{Value: 100},
						UpdateMask: &field_mask.FieldMask{
							Paths: []string{"assign_public_ip"},
						},
					},
				},
			},
			expectedToDelete: map[string][]string{},
		},
		{
			name: "0 add, 1 update (priority), 0 delete",
			currHosts: []*redis.Host{
				{
					Name:            "fqdn1",
					ShardName:       "shard1",
					SubnetId:        "subnet1",
					ReplicaPriority: &wrappers.Int64Value{Value: 100},
					AssignPublicIp:  false,
				},
			},
			targetHosts: []*redis.HostSpec{
				{
					ShardName:       "shard1",
					SubnetId:        "subnet1",
					ReplicaPriority: &wrappers.Int64Value{Value: 99},
					AssignPublicIp:  false,
				},
			},
			expectedToAdd: map[string][]*redis.HostSpec{},
			expectedtoUpdate: map[string][]*HostUpdateInfo{
				"shard1": {
					{
						HostName:        "fqdn1",
						AssignPublicIp:  false,
						ReplicaPriority: &wrappers.Int64Value{Value: 99},
						UpdateMask: &field_mask.FieldMask{
							Paths: []string{"replica_priority"},
						},
					},
				},
			},
			expectedToDelete: map[string][]string{},
		},
		{
			name:    "0 add, 1 update (ip), 0 delete - works in sharded",
			sharded: true,
			currHosts: []*redis.Host{
				{
					Name:            "fqdn1",
					ShardName:       "shard1",
					SubnetId:        "subnet1",
					ReplicaPriority: &wrappers.Int64Value{Value: 100},
					AssignPublicIp:  false,
				},
			},
			targetHosts: []*redis.HostSpec{
				{
					ShardName:       "shard1",
					SubnetId:        "subnet1",
					ReplicaPriority: &wrappers.Int64Value{Value: 100},
					AssignPublicIp:  true,
				},
			},
			expectedToAdd: map[string][]*redis.HostSpec{},
			expectedtoUpdate: map[string][]*HostUpdateInfo{
				"shard1": {
					{
						HostName:        "fqdn1",
						AssignPublicIp:  true,
						ReplicaPriority: &wrappers.Int64Value{Value: 100},
						UpdateMask: &field_mask.FieldMask{
							Paths: []string{"assign_public_ip"},
						},
					},
				},
			},
			expectedToDelete: map[string][]string{},
		},
		{
			name:    "0 add, 1 update (priority), 0 delete - fails in sharded",
			sharded: true,
			currHosts: []*redis.Host{
				{
					Name:            "fqdn1",
					ShardName:       "shard1",
					SubnetId:        "subnet1",
					ReplicaPriority: &wrappers.Int64Value{Value: 100},
					AssignPublicIp:  false,
				},
			},
			targetHosts: []*redis.HostSpec{
				{
					ShardName:       "shard1",
					SubnetId:        "subnet1",
					ReplicaPriority: &wrappers.Int64Value{Value: 99},
					AssignPublicIp:  false,
				},
			},
			expectedError: "modifying replica priority in hosts of sharded clusters is not supported: fqdn1",
		},
		{
			name: "0 add, 1 update (ip and priority), 0 delete",
			currHosts: []*redis.Host{
				{
					Name:            "fqdn1",
					ShardName:       "shard1",
					SubnetId:        "subnet1",
					ReplicaPriority: &wrappers.Int64Value{Value: 100},
					AssignPublicIp:  false,
				},
			},
			targetHosts: []*redis.HostSpec{
				{
					ShardName:       "shard1",
					SubnetId:        "subnet1",
					ReplicaPriority: &wrappers.Int64Value{Value: 99},
					AssignPublicIp:  true,
				},
			},
			expectedToAdd: map[string][]*redis.HostSpec{},
			expectedtoUpdate: map[string][]*HostUpdateInfo{
				"shard1": {
					{
						HostName:        "fqdn1",
						AssignPublicIp:  true,
						ReplicaPriority: &wrappers.Int64Value{Value: 99},
						UpdateMask: &field_mask.FieldMask{
							Paths: []string{"replica_priority", "assign_public_ip"},
						},
					},
				},
			},
			expectedToDelete: map[string][]string{},
		},
		{
			name: "1 add, 0 update, 0 delete",
			currHosts: []*redis.Host{
				{
					Name:            "fqdn1",
					ShardName:       "shard1",
					SubnetId:        "subnet1",
					ReplicaPriority: &wrappers.Int64Value{Value: 100},
					AssignPublicIp:  false,
				},
			},
			targetHosts: []*redis.HostSpec{
				{
					ShardName:       "shard1",
					SubnetId:        "subnet1",
					ReplicaPriority: &wrappers.Int64Value{Value: 100},
					AssignPublicIp:  false,
				},
				{
					ShardName:       "shard1",
					SubnetId:        "subnet2",
					ReplicaPriority: &wrappers.Int64Value{Value: 100},
					AssignPublicIp:  false,
				},
			},
			expectedToAdd: map[string][]*redis.HostSpec{
				"shard1": {
					{
						ShardName:       "shard1",
						SubnetId:        "subnet2",
						AssignPublicIp:  false,
						ReplicaPriority: &wrappers.Int64Value{Value: 100},
						ZoneId:          "",
					},
				},
			},
			expectedtoUpdate: map[string][]*HostUpdateInfo{},
			expectedToDelete: map[string][]string{},
		},
		{
			name: "1 add, 1 update (priority), 0 delete",
			currHosts: []*redis.Host{
				{
					Name:            "fqdn1",
					ShardName:       "shard1",
					SubnetId:        "subnet1",
					ReplicaPriority: &wrappers.Int64Value{Value: 100},
					AssignPublicIp:  false,
				},
			},
			targetHosts: []*redis.HostSpec{
				{
					ShardName:       "shard1",
					SubnetId:        "subnet1",
					ReplicaPriority: &wrappers.Int64Value{Value: 99},
					AssignPublicIp:  false,
				},
				{
					ShardName:       "shard1",
					SubnetId:        "subnet2",
					ReplicaPriority: &wrappers.Int64Value{Value: 101},
					AssignPublicIp:  false,
				},
			},
			expectedToAdd: map[string][]*redis.HostSpec{
				"shard1": {
					{
						ShardName:       "shard1",
						SubnetId:        "subnet2",
						AssignPublicIp:  false,
						ReplicaPriority: &wrappers.Int64Value{Value: 101},
						ZoneId:          "",
					},
				},
			},
			expectedtoUpdate: map[string][]*HostUpdateInfo{
				"shard1": {
					{
						HostName:        "fqdn1",
						AssignPublicIp:  false,
						ReplicaPriority: &wrappers.Int64Value{Value: 99},
						UpdateMask: &field_mask.FieldMask{
							Paths: []string{"replica_priority"},
						},
					},
				},
			},
			expectedToDelete: map[string][]string{},
		},
		{
			name: "1 add, 1 update (ip), 0 delete",
			currHosts: []*redis.Host{
				{
					Name:            "fqdn1",
					ShardName:       "shard1",
					SubnetId:        "subnet1",
					ReplicaPriority: &wrappers.Int64Value{Value: 100},
					AssignPublicIp:  false,
				},
			},
			targetHosts: []*redis.HostSpec{
				{
					ShardName:       "shard1",
					SubnetId:        "subnet1",
					ReplicaPriority: &wrappers.Int64Value{Value: 100},
					AssignPublicIp:  true,
				},
				{
					ShardName:       "shard1",
					SubnetId:        "subnet2",
					ReplicaPriority: &wrappers.Int64Value{Value: 100},
					AssignPublicIp:  true,
				},
			},
			expectedToAdd: map[string][]*redis.HostSpec{
				"shard1": {
					{
						ShardName:       "shard1",
						SubnetId:        "subnet2",
						AssignPublicIp:  true,
						ReplicaPriority: &wrappers.Int64Value{Value: 100},
						ZoneId:          "",
					},
				},
			},
			expectedtoUpdate: map[string][]*HostUpdateInfo{
				"shard1": {
					{
						HostName:        "fqdn1",
						AssignPublicIp:  true,
						ReplicaPriority: &wrappers.Int64Value{Value: 100},
						UpdateMask: &field_mask.FieldMask{
							Paths: []string{"assign_public_ip"},
						},
					},
				},
			},
			expectedToDelete: map[string][]string{},
		},
		{
			name: "1 add, 1 update (ip and priority), 0 delete",
			currHosts: []*redis.Host{
				{
					Name:            "fqdn1",
					ShardName:       "shard1",
					SubnetId:        "subnet1",
					ReplicaPriority: &wrappers.Int64Value{Value: 100},
					AssignPublicIp:  false,
				},
			},
			targetHosts: []*redis.HostSpec{
				{
					ShardName:       "shard1",
					SubnetId:        "subnet1",
					ReplicaPriority: &wrappers.Int64Value{Value: 99},
					AssignPublicIp:  true,
				},
				{
					ShardName:       "shard1",
					SubnetId:        "subnet2",
					ReplicaPriority: &wrappers.Int64Value{Value: 100},
					AssignPublicIp:  false,
				},
			},
			expectedToAdd: map[string][]*redis.HostSpec{
				"shard1": {
					{
						ShardName:       "shard1",
						SubnetId:        "subnet2",
						AssignPublicIp:  false,
						ReplicaPriority: &wrappers.Int64Value{Value: 100},
						ZoneId:          "",
					},
				},
			},
			expectedtoUpdate: map[string][]*HostUpdateInfo{
				"shard1": {
					{
						HostName:        "fqdn1",
						AssignPublicIp:  true,
						ReplicaPriority: &wrappers.Int64Value{Value: 99},
						UpdateMask: &field_mask.FieldMask{
							Paths: []string{"replica_priority", "assign_public_ip"},
						},
					},
				},
			},
			expectedToDelete: map[string][]string{},
		},
		{
			name: "1 add, 0 update, 1 delete",
			currHosts: []*redis.Host{
				{
					Name:            "fqdn1",
					ShardName:       "shard1",
					SubnetId:        "subnet1",
					ReplicaPriority: &wrappers.Int64Value{Value: 100},
					AssignPublicIp:  false,
				},
			},
			targetHosts: []*redis.HostSpec{
				{
					ShardName:       "shard1",
					SubnetId:        "subnet2",
					ReplicaPriority: &wrappers.Int64Value{Value: 100},
					AssignPublicIp:  false,
				},
			},
			expectedToAdd: map[string][]*redis.HostSpec{
				"shard1": {
					{
						ShardName:       "shard1",
						SubnetId:        "subnet2",
						AssignPublicIp:  false,
						ReplicaPriority: &wrappers.Int64Value{Value: 100},
						ZoneId:          "",
					},
				},
			},
			expectedtoUpdate: map[string][]*HostUpdateInfo{},
			expectedToDelete: map[string][]string{
				"shard1": {
					"fqdn1",
				},
			},
		},
		{
			name: "1 add, 1 update (ip and priority), 1 delete",
			currHosts: []*redis.Host{
				{
					Name:            "fqdn1",
					ShardName:       "shard1",
					SubnetId:        "subnet1",
					ReplicaPriority: &wrappers.Int64Value{Value: 100},
					AssignPublicIp:  false,
				},
				{
					Name:            "fqdn2",
					ShardName:       "shard1",
					SubnetId:        "subnet1",
					ReplicaPriority: &wrappers.Int64Value{Value: 100},
					AssignPublicIp:  false,
				},
			},
			targetHosts: []*redis.HostSpec{
				{
					ShardName:       "shard1",
					SubnetId:        "subnet1",
					ReplicaPriority: &wrappers.Int64Value{Value: 99},
					AssignPublicIp:  true,
				},
				{
					ShardName:       "shard1",
					SubnetId:        "subnet2",
					ReplicaPriority: &wrappers.Int64Value{Value: 100},
					AssignPublicIp:  false,
				},
			},
			expectedToAdd: map[string][]*redis.HostSpec{
				"shard1": {
					{
						ShardName:       "shard1",
						SubnetId:        "subnet2",
						AssignPublicIp:  false,
						ReplicaPriority: &wrappers.Int64Value{Value: 100},
						ZoneId:          "",
					},
				},
			},
			expectedtoUpdate: map[string][]*HostUpdateInfo{
				"shard1": {
					{
						HostName:        "fqdn1",
						AssignPublicIp:  true,
						ReplicaPriority: &wrappers.Int64Value{Value: 99},
						UpdateMask: &field_mask.FieldMask{
							Paths: []string{"replica_priority", "assign_public_ip"},
						},
					},
				},
			},
			expectedToDelete: map[string][]string{
				"shard1": {
					"fqdn2",
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actualToDelete, actualToUpdate, actualToAdd, err := redisHostsDiff(tc.sharded, tc.currHosts, tc.targetHosts)
			if tc.expectedError == "" {
				require.Nil(t, err)
			} else {
				require.NotNil(t, err)
				require.Equal(t, err.Error(), tc.expectedError)
			}
			require.Equal(t, tc.expectedToAdd, actualToAdd, "unexpected ADD")
			require.Equal(t, tc.expectedtoUpdate, actualToUpdate, "unexpected UPDATE")
			require.Equal(t, tc.expectedToDelete, actualToDelete, "unexpected DELETE")
		})
	}
}

func TestSortRedisHostsNonsharded(t *testing.T) {
	h1 := &redis.Host{
		Name:            "fqdn1",
		ZoneId:          "zone1",
		SubnetId:        "subnet1",
		ShardName:       "shard1",
		ReplicaPriority: &wrappers.Int64Value{Value: 100},
		AssignPublicIp:  false,
	}
	h2 := &redis.Host{
		Name:            "fqdn2",
		ZoneId:          "zone1",
		SubnetId:        "subnet1",
		ShardName:       "shard1",
		ReplicaPriority: &wrappers.Int64Value{Value: 101},
		AssignPublicIp:  true,
	}
	h3 := &redis.Host{
		Name:            "fqdn3",
		ZoneId:          "zone1",
		SubnetId:        "subnet1",
		ShardName:       "shard1",
		ReplicaPriority: &wrappers.Int64Value{Value: 100},
		AssignPublicIp:  true,
	}
	specs := []*redis.HostSpec{
		{
			ZoneId:          "zone1",
			SubnetId:        "subnet1",
			ShardName:       "shard1",
			ReplicaPriority: &wrappers.Int64Value{Value: 100},
			AssignPublicIp:  false,
		},
		{
			ZoneId:          "zone1",
			SubnetId:        "subnet1",
			ShardName:       "shard1",
			ReplicaPriority: &wrappers.Int64Value{Value: 101},
			AssignPublicIp:  true,
		},
		{
			ZoneId:          "zone1",
			SubnetId:        "subnet1",
			ShardName:       "shard1",
			ReplicaPriority: &wrappers.Int64Value{Value: 100},
			AssignPublicIp:  true,
		},
	}
	expectedHosts := []*redis.Host{h1, h2, h3}

	cases := []struct {
		name          string
		hosts         []*redis.Host
		specs         []*redis.HostSpec
		expectedHosts []*redis.Host
	}{
		{
			name:          "same order",
			hosts:         []*redis.Host{h1, h2, h3},
			specs:         specs,
			expectedHosts: expectedHosts,
		},
		{
			name:          "mixed order (1 3 2)",
			hosts:         []*redis.Host{h1, h3, h2},
			specs:         specs,
			expectedHosts: expectedHosts,
		},
		{
			name:          "mixed order (2 3 1)",
			hosts:         []*redis.Host{h2, h3, h1},
			specs:         specs,
			expectedHosts: expectedHosts,
		},

		{
			name:          "mixed order (2 1 3)",
			hosts:         []*redis.Host{h2, h1, h3},
			specs:         specs,
			expectedHosts: expectedHosts,
		},
		{
			name:          "mixed order (3 2 1)",
			hosts:         []*redis.Host{h3, h2, h1},
			specs:         specs,
			expectedHosts: expectedHosts,
		},
		{
			name:          "mixed order (3 1 2)",
			hosts:         []*redis.Host{h3, h1, h2},
			specs:         specs,
			expectedHosts: expectedHosts,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sortRedisHosts(false, tc.hosts, tc.specs)
			require.Equal(t, tc.expectedHosts, tc.hosts)
		})
	}
}

func TestSortRedisHostsSharded(t *testing.T) {
	h1 := &redis.Host{
		Name:            "fqdn1",
		ZoneId:          "zone1",
		SubnetId:        "subnet1",
		ShardName:       "shard1",
		ReplicaPriority: &wrappers.Int64Value{Value: 100},
		AssignPublicIp:  false,
	}
	h2 := &redis.Host{
		Name:            "fqdn2",
		ZoneId:          "zone1",
		SubnetId:        "subnet1",
		ShardName:       "shard2",
		ReplicaPriority: &wrappers.Int64Value{Value: 101},
		AssignPublicIp:  true,
	}
	h3 := &redis.Host{
		Name:            "fqdn3",
		ZoneId:          "zone1",
		SubnetId:        "subnet1",
		ShardName:       "shard3",
		ReplicaPriority: &wrappers.Int64Value{Value: 100},
		AssignPublicIp:  true,
	}
	specs := []*redis.HostSpec{
		{
			ZoneId:          "zone1",
			SubnetId:        "subnet1",
			ShardName:       "shard1",
			ReplicaPriority: &wrappers.Int64Value{Value: 100},
			AssignPublicIp:  false,
		},
		{
			ZoneId:          "zone1",
			SubnetId:        "subnet1",
			ShardName:       "shard2",
			ReplicaPriority: &wrappers.Int64Value{Value: 101},
			AssignPublicIp:  true,
		},
		{
			ZoneId:          "zone1",
			SubnetId:        "subnet1",
			ShardName:       "shard3",
			ReplicaPriority: &wrappers.Int64Value{Value: 100},
			AssignPublicIp:  true,
		},
	}
	expectedHosts := []*redis.Host{h1, h2, h3}

	cases := []struct {
		name          string
		hosts         []*redis.Host
		specs         []*redis.HostSpec
		expectedHosts []*redis.Host
	}{
		{
			name:          "same order",
			hosts:         []*redis.Host{h1, h2, h3},
			specs:         specs,
			expectedHosts: expectedHosts,
		},
		{
			name:          "mixed order (1 3 2)",
			hosts:         []*redis.Host{h1, h3, h2},
			specs:         specs,
			expectedHosts: expectedHosts,
		},
		{
			name:          "mixed order (2 3 1)",
			hosts:         []*redis.Host{h2, h3, h1},
			specs:         specs,
			expectedHosts: expectedHosts,
		},

		{
			name:          "mixed order (2 1 3)",
			hosts:         []*redis.Host{h2, h1, h3},
			specs:         specs,
			expectedHosts: expectedHosts,
		},
		{
			name:          "mixed order (3 2 1)",
			hosts:         []*redis.Host{h3, h2, h1},
			specs:         specs,
			expectedHosts: expectedHosts,
		},
		{
			name:          "mixed order (3 1 2)",
			hosts:         []*redis.Host{h3, h1, h2},
			specs:         specs,
			expectedHosts: expectedHosts,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sortRedisHosts(true, tc.hosts, tc.specs)
			require.Equal(t, tc.expectedHosts, tc.hosts)
		})
	}
}
