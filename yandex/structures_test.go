package yandex

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform/helper/schema"
	"google.golang.org/grpc"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
)

type DiskClientGetter struct {
}

func (r *DiskClientGetter) Get(ctx context.Context, in *compute.GetDiskRequest, opts ...grpc.CallOption) (*compute.Disk, error) {
	return &compute.Disk{
		Id:          "",
		FolderId:    "",
		CreatedAt:   nil,
		Name:        "mock-disk-name",
		Description: "mock-disk-description",
		TypeId:      "network-hdd",
		ZoneId:      "",
		Size:        4 * (1 << 30),
		ProductIds:  nil,
	}, nil
}

func TestExpandLabels(t *testing.T) {
	cases := []struct {
		name     string
		labels   interface{}
		expected map[string]string
	}{
		{
			name: "two tags",
			labels: map[string]interface{}{
				"my_key":       "my_value",
				"my_other_key": "my_other_value",
			},
			expected: map[string]string{
				"my_key":       "my_value",
				"my_other_key": "my_other_value",
			},
		},
		{
			name:     "labels is nil",
			labels:   nil,
			expected: map[string]string{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := expandLabels(tc.labels)
			if err != nil {
				t.Fatalf("bad: %#v", err)
			}
			if !reflect.DeepEqual(result, tc.expected) {
				t.Fatalf("Got:\n\n%#v\n\nExpected:\n\n%#v\n", result, tc.expected)
			}
		})
	}
}

func TestExpandProductIds(t *testing.T) {
	cases := []struct {
		name       string
		productIds *schema.Set
		expected   []string
	}{
		{
			name: "two product ids",
			productIds: schema.NewSet(schema.HashString, []interface{}{
				"super-product",
				"very-good",
			}),
			expected: []string{
				"super-product",
				"very-good",
			},
		},
		{
			name:       "empty product ids",
			productIds: schema.NewSet(schema.HashString, []interface{}{}),
			expected:   []string{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := expandProductIds(tc.productIds)
			if err != nil {
				t.Fatalf("bad: %#v", err)
			}
			if !reflect.DeepEqual(result, tc.expected) {
				t.Fatalf("Got:\n\n%#v\n\nExpected:\n\n%#v\n", result, tc.expected)
			}
		})
	}
}

func TestFlattenInstanceResources(t *testing.T) {
	cases := []struct {
		name      string
		resources *compute.Resources
		expected  []map[string]interface{}
	}{
		{
			name: "cores 1 fraction 100 memory 5gb",
			resources: &compute.Resources{
				Cores:        1,
				CoreFraction: 100,
				Memory:       5 * (1 << 30),
			},
			expected: []map[string]interface{}{
				{
					"cores":         1,
					"core_fraction": 100,
					"memory":        5,
				},
			},
		},
		{
			name: "cores 8 fraction 5 memory 16gb",
			resources: &compute.Resources{
				Cores:        8,
				CoreFraction: 5,
				Memory:       16 * (1 << 30),
			},
			expected: []map[string]interface{}{
				{
					"cores":         8,
					"core_fraction": 5,
					"memory":        16,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := flattenInstanceResources(&compute.Instance{Resources: tc.resources})
			if err != nil {
				t.Fatalf("bad: %#v", err)
			}
			if !reflect.DeepEqual(result, tc.expected) {
				t.Fatalf("Got:\n\n%#v\n\nExpected:\n\n%#v\n", result, tc.expected)
			}
		})
	}
}

func TestFlattenInstanceBootDisk(t *testing.T) {
	cases := []struct {
		name     string
		bootDisk *compute.AttachedDisk
		expected []map[string]interface{}
	}{
		{
			name: "boot disk with diskID",
			bootDisk: &compute.AttachedDisk{
				Mode:       compute.AttachedDisk_READ_WRITE,
				DeviceName: "test-device-name",
				AutoDelete: false,
				DiskId:     "saeque9k",
			},
			expected: []map[string]interface{}{
				{
					"device_name": "test-device-name",
					"auto_delete": false,
					"disk_id":     "saeque9k",
					"mode":        "READ_WRITE",
					"initialize_params": []map[string]interface{}{
						{"snapshot_id": "",
							"name":        "mock-disk-name",
							"description": "mock-disk-description",
							"size":        4,
							"type_id":     "network-hdd",
							"image_id":    "",
						},
					},
				},
			},
		},
	}

	reducedDiskClient := &DiskClientGetter{}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := flattenInstanceBootDisk(&compute.Instance{BootDisk: tc.bootDisk}, reducedDiskClient)

			if err != nil {
				t.Fatalf("bad: %#v", err)
			}
			if !reflect.DeepEqual(result, tc.expected) {
				t.Fatalf("Got:\n\n%#v\n\nExpected:\n\n%#v\n", result, tc.expected)
			}
		})
	}
}
