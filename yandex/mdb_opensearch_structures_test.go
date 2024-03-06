package yandex

import (
	"testing"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/opensearch/v1"
)

func Test_modifyConfig(t *testing.T) {
	t.Parallel()

	type args struct {
		oldConfig *opensearch.ConfigCreateSpec
		newConfig *opensearch.ConfigCreateSpec
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "case negative: empty old config should return false",
			args: args{
				oldConfig: nil,
				newConfig: nil,
			},
			want: false,
		},
		{
			name: "case negative: nil oldConfig.OpensearchSpec should return false",
			args: args{
				oldConfig: &opensearch.ConfigCreateSpec{
					Version:        "v0.1.2",
					OpensearchSpec: nil, // here just for clarification
				},
				newConfig: &opensearch.ConfigCreateSpec{
					Version:        "v0.1.2",
					OpensearchSpec: &opensearch.OpenSearchCreateSpec{},
				},
			},
			want: false,
		},
		{
			name: "case negative: nil newConfig.OpensearchSpec should return false",
			args: args{
				oldConfig: &opensearch.ConfigCreateSpec{
					Version:        "v0.1.2",
					OpensearchSpec: &opensearch.OpenSearchCreateSpec{},
				},
				newConfig: &opensearch.ConfigCreateSpec{
					Version:        "v0.1.2",
					OpensearchSpec: nil, // here just for clarification
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := modifyConfig(tt.args.oldConfig, tt.args.newConfig); got != tt.want {
				t.Errorf("modifyConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
