package yandex

import (
	"reflect"
	"testing"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/ydb/v1"
)

func TestFlattenYDBLocation(t *testing.T) {
	tests := []struct {
		name     string
		spec     ydb.Database_DatabaseType
		expected []map[string]interface{}
	}{
		{
			name:     "dedicated",
			spec:     nil,
			expected: nil,
		},
		{
			name: "regional",
			spec: &ydb.Database_RegionalDatabase{
				RegionalDatabase: &ydb.RegionalDatabase{
					RegionId: "region_id",
				},
			},
			expected: []map[string]interface{}{
				{
					"region": []map[string]interface{}{
						{
							"id": "region_id",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := flattenYDBLocation(&ydb.Database{DatabaseType: tt.spec})

			if err != nil {
				t.Errorf("%v", err)
			}
			if !reflect.DeepEqual(res, tt.expected) {
				t.Errorf("flattenYDBLocation() got = %v, want %v", res, tt.expected)
			}
		})
	}
}

func TestFlattenYDBStorageConfig(t *testing.T) {
	tests := []struct {
		name     string
		spec     *ydb.StorageConfig
		expected []map[string]interface{}
	}{
		{
			name:     "empty",
			spec:     nil,
			expected: nil,
		},
		{
			name:     "zero elements",
			spec:     &ydb.StorageConfig{},
			expected: []map[string]interface{}{},
		},
		{
			name: "zero elements - 2",
			spec: &ydb.StorageConfig{
				StorageOptions: []*ydb.StorageOption{},
			},
			expected: []map[string]interface{}{},
		},
		{
			name: "ssd",
			spec: &ydb.StorageConfig{
				StorageOptions: []*ydb.StorageOption{
					{
						StorageTypeId: "ssd",
						GroupCount:    1,
					},
				},
			},
			expected: []map[string]interface{}{
				{
					"storage_type_id": "ssd",
					"group_count":     1,
				},
			},
		},
		{
			name: "2 groups",
			spec: &ydb.StorageConfig{
				StorageOptions: []*ydb.StorageOption{
					{
						StorageTypeId: "ssd",
						GroupCount:    1,
					},
					{
						StorageTypeId: "hdd",
						GroupCount:    3,
					},
				},
			},
			expected: []map[string]interface{}{
				{
					"storage_type_id": "ssd",
					"group_count":     1,
				},
				{
					"storage_type_id": "hdd",
					"group_count":     3,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := flattenYDBStorageConfig(tt.spec)

			if err != nil {
				t.Errorf("%v", err)
			}
			if !reflect.DeepEqual(res, tt.expected) {
				t.Errorf("flattenYDBStorageConfig() got = %v, want %v", res, tt.expected)
			}
		})
	}
}

func TestFlattenYDBScalePolicy(t *testing.T) {
	tests := []struct {
		name     string
		spec     *ydb.ScalePolicy
		expected []map[string]interface{}
	}{
		{
			name:     "empty",
			spec:     nil,
			expected: nil,
		},
		{
			name: "fixed scale",
			spec: &ydb.ScalePolicy{
				ScaleType: &ydb.ScalePolicy_FixedScale_{
					FixedScale: &ydb.ScalePolicy_FixedScale{Size: 3},
				},
			},
			expected: []map[string]interface{}{
				{
					"fixed_scale": []map[string]interface{}{
						{
							"size": 3,
						},
					},
				},
			},
		},
		{
			name: "fixed scale - 5",
			spec: &ydb.ScalePolicy{
				ScaleType: &ydb.ScalePolicy_FixedScale_{
					FixedScale: &ydb.ScalePolicy_FixedScale{Size: 5},
				},
			},
			expected: []map[string]interface{}{
				{
					"fixed_scale": []map[string]interface{}{
						{
							"size": 5,
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := flattenYDBScalePolicy(&ydb.Database{ScalePolicy: tt.spec})

			if err != nil {
				t.Errorf("%v", err)
			}
			if !reflect.DeepEqual(res, tt.expected) {
				t.Errorf("flattenYDBScalePolicy() got = %v, want %v", res, tt.expected)
			}
		})
	}
}
