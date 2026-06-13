package yandex

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/kafka/v1"
	"sort"
	"testing"
)

func Test_parseSetToStringArray(t *testing.T) {
	type args struct {
		set interface{}
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "set is nil -> return nil slice",
			args: args{
				set: nil,
			},
			want: nil,
		},
		{
			name: "set is empty -> return empty slice",
			args: args{
				set: schema.NewSet(schema.HashString, []interface{}{}),
			},
			want: []string{},
		},
		{
			name: "correct scenario",
			args: args{
				set: schema.NewSet(schema.HashString, []interface{}{
					"xyz",
					"abcabc",
					"babba",
				}),
			},
			want: []string{"xyz", "abcabc", "babba"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseSetToStringArray(tt.args.set)
			sort.Strings(result)
			sort.Strings(tt.want)
			assert.Equalf(t, tt.want, result, "parseSetToStringArray(%v)", tt.args.set)
		})
	}
}

func Test_parseKafkaPermissionAllowHosts(t *testing.T) {
	type args struct {
		allowHosts interface{}
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "allowHosts is nil -> return nil slice",
			args: args{
				allowHosts: nil,
			},
			want: nil,
		},
		{
			name: "allowHosts is empty -> return nil slice",
			args: args{
				allowHosts: schema.NewSet(schema.HashString, []interface{}{}),
			},
			want: nil,
		},
		{
			name: "correct scenario",
			args: args{
				allowHosts: schema.NewSet(schema.HashString, []interface{}{
					"host2",
					"host1",
					"host3",
				}),
			},
			want: []string{"host1", "host2", "host3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseKafkaPermissionAllowHosts(tt.args.allowHosts)
			sort.Strings(result)
			sort.Strings(tt.want)
			assert.Equalf(t, tt.want, result, "parseKafkaPermissionAllowHosts(%v)", tt.args.allowHosts)
		})
	}
}

func Test_flattenKafkaClusterSubnets(t *testing.T) {
	tests := []struct {
		name  string
		hosts []*kafka.Host
		want  []string
	}{
		{
			name:  "nil hosts",
			hosts: nil,
			want:  []string{},
		},
		{
			name: "single host",
			hosts: []*kafka.Host{
				{SubnetId: "subnet-a"},
			},
			want: []string{"subnet-a"},
		},
		{
			name: "deduplicates and preserves first-seen order",
			hosts: []*kafka.Host{
				{SubnetId: "subnet-b"},
				{SubnetId: "subnet-a"},
				{SubnetId: "subnet-b"},
				{SubnetId: "subnet-c"},
				{SubnetId: "subnet-a"},
			},
			want: []string{"subnet-b", "subnet-a", "subnet-c"},
		},
		{
			name: "skips empty subnet ids",
			hosts: []*kafka.Host{
				{SubnetId: ""},
				{SubnetId: "subnet-a"},
				{SubnetId: ""},
			},
			want: []string{"subnet-a"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := flattenKafkaClusterSubnets(tt.hosts)
			assert.Equalf(t, tt.want, got, "flattenKafkaClusterSubnets(%v)", tt.hosts)
		})
	}
}

func Test_orderKafkaSubnetsByState(t *testing.T) {
	tests := []struct {
		name          string
		stateSubnets  []string
		actualSubnets []string
		want          []string
	}{
		{
			name:          "empty state keeps actual order",
			stateSubnets:  nil,
			actualSubnets: []string{"subnet-a", "subnet-b"},
			want:          []string{"subnet-a", "subnet-b"},
		},
		{
			name:          "preserves state order when the set is unchanged",
			stateSubnets:  []string{"subnet-c", "subnet-a", "subnet-b"},
			actualSubnets: []string{"subnet-a", "subnet-b", "subnet-c"},
			want:          []string{"subnet-c", "subnet-a", "subnet-b"},
		},
		{
			name:          "drops subnets no longer present",
			stateSubnets:  []string{"subnet-a", "subnet-gone", "subnet-b"},
			actualSubnets: []string{"subnet-b", "subnet-a"},
			want:          []string{"subnet-a", "subnet-b"},
		},
		{
			name:          "appends new subnets after the known ones",
			stateSubnets:  []string{"subnet-b", "subnet-a"},
			actualSubnets: []string{"subnet-a", "subnet-b", "subnet-new"},
			want:          []string{"subnet-b", "subnet-a", "subnet-new"},
		},
		{
			name:          "deduplicates repeated state entries",
			stateSubnets:  []string{"subnet-a", "subnet-a", "subnet-b"},
			actualSubnets: []string{"subnet-a", "subnet-b"},
			want:          []string{"subnet-a", "subnet-b"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := orderKafkaSubnetsByState(tt.stateSubnets, tt.actualSubnets)
			assert.Equalf(t, tt.want, got, "orderKafkaSubnetsByState(%v, %v)", tt.stateSubnets, tt.actualSubnets)
		})
	}
}
