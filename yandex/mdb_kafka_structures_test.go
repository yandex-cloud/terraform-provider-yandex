package yandex

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
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
