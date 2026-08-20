package yandex

import (
	"reflect"
	"testing"
)

func TestPreserveStringSliceOrder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		preferred []string
		actual    []string
		want      []string
	}{
		{
			name:      "same values in different order",
			preferred: []string{"yc.ydb.topics.manage", "yc.ydb.tables.manage", "yc.postbox.send"},
			actual:    []string{"yc.postbox.send", "yc.ydb.tables.manage", "yc.ydb.topics.manage"},
			want:      []string{"yc.ydb.topics.manage", "yc.ydb.tables.manage", "yc.postbox.send"},
		},
		{
			name:      "API values changed",
			preferred: []string{"first", "second"},
			actual:    []string{"first", "third"},
			want:      []string{"first", "third"},
		},
		{
			name:      "duplicate counts differ",
			preferred: []string{"first", "first"},
			actual:    []string{"first", "second"},
			want:      []string{"first", "second"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := preserveStringSliceOrder(tt.preferred, tt.actual); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("preserveStringSliceOrder() = %v, want %v", got, tt.want)
			}
		})
	}
}
