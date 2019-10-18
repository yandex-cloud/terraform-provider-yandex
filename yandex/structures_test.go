package yandex

import (
	"context"
	"reflect"
	"testing"

	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"google.golang.org/grpc"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1/instancegroup"
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
			name: "cores 1 fraction 100 memory 5 gb 0 gpus",
			resources: &compute.Resources{
				Cores:        1,
				CoreFraction: 100,
				Memory:       5 * (1 << 30),
				Gpus:         0,
			},
			expected: []map[string]interface{}{
				{
					"cores":         1,
					"core_fraction": 100,
					"memory":        5.0,
					"gpus":          0,
				},
			},
		},
		{
			name: "cores 8 fraction 5 memory 16 gb 0 gpus",
			resources: &compute.Resources{
				Cores:        8,
				CoreFraction: 5,
				Memory:       16 * (1 << 30),
				Gpus:         0,
			},
			expected: []map[string]interface{}{
				{
					"cores":         8,
					"core_fraction": 5,
					"memory":        16.0,
					"gpus":          0,
				},
			},
		},
		{
			name: "cores 2 fraction 20 memory 0.5 gb 0 gpus",
			resources: &compute.Resources{
				Cores:        2,
				CoreFraction: 20,
				Memory:       (1 << 30) / 2,
				Gpus:         0,
			},
			expected: []map[string]interface{}{
				{
					"cores":         2,
					"core_fraction": 20,
					"memory":        0.5,
					"gpus":          0,
				},
			},
		},
		{
			name: "cores 8 fraction 100 memory 96 gb 2 gpus",
			resources: &compute.Resources{
				Cores:        8,
				CoreFraction: 100,
				Memory:       96 * (1 << 30),
				Gpus:         2,
			},
			expected: []map[string]interface{}{
				{
					"cores":         8,
					"core_fraction": 100,
					"memory":        96.0,
					"gpus":          2,
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
							"type":        "network-hdd",
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
			result, err := flattenInstanceBootDisk(context.Background(), &compute.Instance{BootDisk: tc.bootDisk}, reducedDiskClient)

			if err != nil {
				t.Fatalf("bad: %#v", err)
			}
			if !reflect.DeepEqual(result, tc.expected) {
				t.Fatalf("Got:\n\n%#v\n\nExpected:\n\n%#v\n", result, tc.expected)
			}
		})
	}
}

func TestFlattenInstanceNetworkInterfaces(t *testing.T) {
	tests := []struct {
		name       string
		instance   *compute.Instance
		want       []map[string]interface{}
		externalIP string
		internalIP string
		wantErr    bool
	}{
		{
			name: "no nics defined",
			instance: &compute.Instance{
				NetworkInterfaces: []*compute.NetworkInterface{},
			},
			want:       []map[string]interface{}{},
			externalIP: "",
			internalIP: "",
			wantErr:    false,
		},
		{
			name: "one nic with internal address",
			instance: &compute.Instance{
				NetworkInterfaces: []*compute.NetworkInterface{
					{
						Index: "1",
						PrimaryV4Address: &compute.PrimaryAddress{
							Address: "192.168.19.16",
						},
						SubnetId:   "some-subnet-id",
						MacAddress: "aa-bb-cc-dd-ee-ff",
					},
				},
			},
			want: []map[string]interface{}{
				{
					"index":       1,
					"mac_address": "aa-bb-cc-dd-ee-ff",
					"subnet_id":   "some-subnet-id",
					"ip_address":  "192.168.19.16",
					"nat":         false,
				},
			},
			externalIP: "",
			internalIP: "192.168.19.16",
			wantErr:    false,
		},
		{
			name: "one nic with internal and external address",
			instance: &compute.Instance{
				NetworkInterfaces: []*compute.NetworkInterface{
					{
						Index: "1",
						PrimaryV4Address: &compute.PrimaryAddress{
							Address: "192.168.19.86",
							OneToOneNat: &compute.OneToOneNat{
								Address:   "92.68.12.34",
								IpVersion: compute.IpVersion_IPV4,
							},
						},
						SubnetId:   "some-subnet-id",
						MacAddress: "aa-bb-cc-dd-ee-ff",
					},
				},
			},
			want: []map[string]interface{}{
				{
					"index":          1,
					"mac_address":    "aa-bb-cc-dd-ee-ff",
					"subnet_id":      "some-subnet-id",
					"ip_address":     "192.168.19.86",
					"nat":            true,
					"nat_ip_address": "92.68.12.34",
					"nat_ip_version": "IPV4",
				},
			},
			externalIP: "92.68.12.34",
			internalIP: "192.168.19.86",
			wantErr:    false,
		},
		{
			name: "one nic with ipv6 address",
			instance: &compute.Instance{
				NetworkInterfaces: []*compute.NetworkInterface{
					{
						Index: "1",
						PrimaryV6Address: &compute.PrimaryAddress{
							Address: "2001:db8::370:7348",
						},
						SubnetId:   "some-subnet-id",
						MacAddress: "aa-bb-cc-dd-ee-ff",
					},
				},
			},
			want: []map[string]interface{}{
				{
					"index":        1,
					"mac_address":  "aa-bb-cc-dd-ee-ff",
					"subnet_id":    "some-subnet-id",
					"ipv6":         true,
					"ipv6_address": "2001:db8::370:7348",
				},
			},
			externalIP: "2001:db8::370:7348",
			internalIP: "",
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nics, externalIP, internalIP, err := flattenInstanceNetworkInterfaces(tt.instance)
			if (err != nil) != tt.wantErr {
				t.Errorf("flattenInstanceNetworkInterfaces() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(nics, tt.want) {
				t.Errorf("flattenInstanceNetworkInterfaces() nics = %v, want %v", nics, tt.want)
			}
			if externalIP != tt.externalIP {
				t.Errorf("flattenInstanceNetworkInterfaces() externalIP = %v, want %v", externalIP, tt.externalIP)
			}
			if internalIP != tt.internalIP {
				t.Errorf("flattenInstanceNetworkInterfaces() internalIP = %v, want %v", internalIP, tt.internalIP)
			}
		})
	}
}

func TestFlattenInstanceGroupManagedInstanceNetworkInterfaces(t *testing.T) {
	tests := []struct {
		name       string
		instance   *instancegroup.ManagedInstance
		want       []map[string]interface{}
		externalIP string
		internalIP string
		wantErr    bool
	}{
		{
			name: "no nics defined",
			instance: &instancegroup.ManagedInstance{
				NetworkInterfaces: []*instancegroup.NetworkInterface{},
			},
			want:       []map[string]interface{}{},
			externalIP: "",
			internalIP: "",
			wantErr:    false,
		},
		{
			name: "one nic with internal address",
			instance: &instancegroup.ManagedInstance{
				NetworkInterfaces: []*instancegroup.NetworkInterface{
					{
						Index: "1",
						PrimaryV4Address: &instancegroup.PrimaryAddress{
							Address: "192.168.19.16",
						},
						SubnetId:   "some-subnet-id",
						MacAddress: "aa-bb-cc-dd-ee-ff",
					},
				},
			},
			want: []map[string]interface{}{
				{
					"index":       1,
					"mac_address": "aa-bb-cc-dd-ee-ff",
					"subnet_id":   "some-subnet-id",
					"ip_address":  "192.168.19.16",
					"nat":         false,
				},
			},
			externalIP: "",
			internalIP: "192.168.19.16",
			wantErr:    false,
		},
		{
			name: "one nic with internal and external address",
			instance: &instancegroup.ManagedInstance{
				NetworkInterfaces: []*instancegroup.NetworkInterface{
					{
						Index: "1",
						PrimaryV4Address: &instancegroup.PrimaryAddress{
							Address: "192.168.19.86",
							OneToOneNat: &instancegroup.OneToOneNat{
								Address:   "92.68.12.34",
								IpVersion: instancegroup.IpVersion_IPV4,
							},
						},
						SubnetId:   "some-subnet-id",
						MacAddress: "aa-bb-cc-dd-ee-ff",
					},
				},
			},
			want: []map[string]interface{}{
				{
					"index":          1,
					"mac_address":    "aa-bb-cc-dd-ee-ff",
					"subnet_id":      "some-subnet-id",
					"ip_address":     "192.168.19.86",
					"nat":            true,
					"nat_ip_address": "92.68.12.34",
					"nat_ip_version": "IPV4",
				},
			},
			externalIP: "92.68.12.34",
			internalIP: "192.168.19.86",
			wantErr:    false,
		},
		{
			name: "one nic with ipv6 address",
			instance: &instancegroup.ManagedInstance{
				NetworkInterfaces: []*instancegroup.NetworkInterface{
					{
						Index: "1",
						PrimaryV6Address: &instancegroup.PrimaryAddress{
							Address: "2001:db8::370:7348",
						},
						SubnetId:   "some-subnet-id",
						MacAddress: "aa-bb-cc-dd-ee-ff",
					},
				},
			},
			want: []map[string]interface{}{
				{
					"index":        1,
					"mac_address":  "aa-bb-cc-dd-ee-ff",
					"subnet_id":    "some-subnet-id",
					"ipv6":         true,
					"ipv6_address": "2001:db8::370:7348",
				},
			},
			externalIP: "2001:db8::370:7348",
			internalIP: "",
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nics, externalIP, internalIP, err := flattenInstanceGroupManagedInstanceNetworkInterfaces(tt.instance)
			if (err != nil) != tt.wantErr {
				t.Errorf("flattenInstanceGroupManagedInstanceNetworkInterfaces() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(nics, tt.want) {
				t.Errorf("flattenInstanceGroupManagedInstanceNetworkInterfaces() nics = %v, want %v", nics, tt.want)
			}
			if externalIP != tt.externalIP {
				t.Errorf("flattenInstanceGroupManagedInstanceNetworkInterfaces() externalIP = %v, want %v", externalIP, tt.externalIP)
			}
			if internalIP != tt.internalIP {
				t.Errorf("flattenInstanceGroupManagedInstanceNetworkInterfaces() internalIP = %v, want %v", internalIP, tt.internalIP)
			}
		})
	}
}

func TestFlattenInstanceGroupInstanceTemplateResources(t *testing.T) {
	cases := []struct {
		name      string
		resources *instancegroup.ResourcesSpec
		expected  []map[string]interface{}
	}{
		{
			name: "cores 1 fraction 100 memory 5 gb 0 gpus",
			resources: &instancegroup.ResourcesSpec{
				Cores:        1,
				CoreFraction: 100,
				Memory:       5 * (1 << 30),
				Gpus:         0,
			},
			expected: []map[string]interface{}{
				{
					"cores":         1,
					"core_fraction": 100,
					"memory":        5.0,
					"gpus":          0,
				},
			},
		},
		{
			name: "cores 8 fraction 5 memory 16 gb 0 gpus",
			resources: &instancegroup.ResourcesSpec{
				Cores:        8,
				CoreFraction: 5,
				Memory:       16 * (1 << 30),
				Gpus:         0,
			},
			expected: []map[string]interface{}{
				{
					"cores":         8,
					"core_fraction": 5,
					"memory":        16.0,
					"gpus":          0,
				},
			},
		},
		{
			name: "cores 2 fraction 20 memory 0.5 gb 0 gpus",
			resources: &instancegroup.ResourcesSpec{
				Cores:        2,
				CoreFraction: 20,
				Memory:       (1 << 30) / 2,
				Gpus:         0,
			},
			expected: []map[string]interface{}{
				{
					"cores":         2,
					"core_fraction": 20,
					"memory":        0.5,
					"gpus":          0,
				},
			},
		},
		{
			name: "cores 8 fraction 100 memory 96 gb 2 gpus",
			resources: &instancegroup.ResourcesSpec{
				Cores:        8,
				CoreFraction: 100,
				Memory:       96 * (1 << 30),
				Gpus:         2,
			},
			expected: []map[string]interface{}{
				{
					"cores":         8,
					"core_fraction": 100,
					"memory":        96.0,
					"gpus":          2,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := flattenInstanceGroupInstanceTemplateResources(tc.resources)
			if err != nil {
				t.Fatalf("bad: %#v", err)
			}
			if !reflect.DeepEqual(result, tc.expected) {
				t.Fatalf("Got:\n\n%#v\n\nExpected:\n\n%#v\n", result, tc.expected)
			}
		})
	}
}

func TestFlattenInstanceGroupAttachedDisk(t *testing.T) {
	cases := []struct {
		name     string
		spec     *instancegroup.AttachedDiskSpec
		expected map[string]interface{}
	}{
		{
			name: "boot disk with diskID",
			spec: &instancegroup.AttachedDiskSpec{
				Mode:       instancegroup.AttachedDiskSpec_READ_WRITE,
				DeviceName: "test-device-name",
				DiskSpec: &instancegroup.AttachedDiskSpec_DiskSpec{
					Description: "mock-disk-description",
					TypeId:      "network-hdd",
					Size:        100 * (1 << 30),
					SourceOneof: &instancegroup.AttachedDiskSpec_DiskSpec_ImageId{
						ImageId: "imageId",
					},
				},
			},
			expected: map[string]interface{}{
				"device_name": "test-device-name",
				"mode":        "READ_WRITE",
				"initialize_params": []map[string]interface{}{
					{
						"description": "mock-disk-description",
						"size":        100,
						"type":        "network-hdd",
						"image_id":    "imageId",
						"snapshot_id": "",
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := flattenInstanceGroupAttachedDisk(tc.spec)

			if err != nil {
				t.Fatalf("bad: %#v", err)
			}
			if !reflect.DeepEqual(result, tc.expected) {
				t.Fatalf("Got:\n\n%#v\n\nExpected:\n\n%#v\n", result, tc.expected)
			}
		})
	}
}

func TestFlattenInstanceGroupHealthChecks(t *testing.T) {
	tests := []struct {
		name     string
		spec     *instancegroup.HealthChecksSpec
		expected []map[string]interface{}
	}{
		{
			name: "one tcp",
			spec: &instancegroup.HealthChecksSpec{
				HealthCheckSpecs: []*instancegroup.HealthCheckSpec{
					{
						Interval:           &duration.Duration{Seconds: 10},
						Timeout:            &duration.Duration{Seconds: 20},
						UnhealthyThreshold: 1,
						HealthyThreshold:   2,
						HealthCheckOptions: &instancegroup.HealthCheckSpec_TcpOptions_{
							TcpOptions: &instancegroup.HealthCheckSpec_TcpOptions{
								Port: 22,
							},
						},
					},
				},
			},
			expected: []map[string]interface{}{
				{
					"interval":            10,
					"timeout":             20,
					"unhealthy_threshold": 1,
					"healthy_threshold":   2,
					"tcp_options": []map[string]interface{}{
						{
							"port": 22,
						},
					},
				},
			},
		},
		{
			name: "tcp + http",
			spec: &instancegroup.HealthChecksSpec{
				HealthCheckSpecs: []*instancegroup.HealthCheckSpec{
					{
						Interval:           &duration.Duration{Seconds: 10},
						Timeout:            &duration.Duration{Seconds: 20},
						UnhealthyThreshold: 1,
						HealthyThreshold:   2,
						HealthCheckOptions: &instancegroup.HealthCheckSpec_TcpOptions_{
							TcpOptions: &instancegroup.HealthCheckSpec_TcpOptions{
								Port: 22,
							},
						},
					},
					{
						Interval:           &duration.Duration{Seconds: 10},
						Timeout:            &duration.Duration{Seconds: 20},
						UnhealthyThreshold: 1,
						HealthyThreshold:   2,
						HealthCheckOptions: &instancegroup.HealthCheckSpec_HttpOptions_{
							HttpOptions: &instancegroup.HealthCheckSpec_HttpOptions{
								Port: 8080,
								Path: "/",
							},
						},
					},
				},
			},
			expected: []map[string]interface{}{
				{
					"interval":            10,
					"timeout":             20,
					"unhealthy_threshold": 1,
					"healthy_threshold":   2,
					"tcp_options": []map[string]interface{}{
						{
							"port": 22,
						},
					},
				},
				{
					"interval":            10,
					"timeout":             20,
					"unhealthy_threshold": 1,
					"healthy_threshold":   2,
					"http_options": []map[string]interface{}{
						{
							"port": 8080,
							"path": "/",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := flattenInstanceGroupHealthChecks(&instancegroup.InstanceGroup{HealthChecksSpec: tt.spec})

			if err != nil {
				t.Errorf("%v", err)
			}
			if !reflect.DeepEqual(res, tt.expected) {
				t.Errorf("flattenInstanceGroupHealthChecks() got = %v, want %v", res, tt.expected)
			}
		})
	}
}

func TestFlattenInstanceGroupScalePolicy(t *testing.T) {
	tests := []struct {
		name     string
		spec     *instancegroup.ScalePolicy
		expected []map[string]interface{}
	}{
		{
			name: "fixed scale",
			spec: &instancegroup.ScalePolicy{
				ScaleType: &instancegroup.ScalePolicy_FixedScale_{
					FixedScale: &instancegroup.ScalePolicy_FixedScale{Size: 3},
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
			name: "auto scale",
			spec: &instancegroup.ScalePolicy{
				ScaleType: &instancegroup.ScalePolicy_AutoScale_{
					AutoScale: &instancegroup.ScalePolicy_AutoScale{
						MinZoneSize:         1,
						MaxSize:             2,
						MeasurementDuration: &duration.Duration{Seconds: 10},
						InitialSize:         3,
					},
				},
			},
			expected: []map[string]interface{}{
				{
					"auto_scale": []map[string]interface{}{
						{
							"min_zone_size":        1,
							"max_size":             2,
							"initial_size":         3,
							"measurement_duration": 10,
						},
					},
				},
			},
		},
		{
			name: "auto scale 2",
			spec: &instancegroup.ScalePolicy{
				ScaleType: &instancegroup.ScalePolicy_AutoScale_{
					AutoScale: &instancegroup.ScalePolicy_AutoScale{
						MinZoneSize:           1,
						MaxSize:               2,
						MeasurementDuration:   &duration.Duration{Seconds: 10},
						WarmupDuration:        &duration.Duration{Seconds: 20},
						StabilizationDuration: &duration.Duration{Seconds: 30},
						InitialSize:           3,
						CpuUtilizationRule:    &instancegroup.ScalePolicy_CpuUtilizationRule{UtilizationTarget: 80},
					},
				},
			},
			expected: []map[string]interface{}{
				{
					"auto_scale": []map[string]interface{}{
						{
							"min_zone_size":          1,
							"max_size":               2,
							"initial_size":           3,
							"measurement_duration":   10,
							"warmup_duration":        20,
							"stabilization_duration": 30,
							"cpu_utilization_target": 80.0,
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := flattenInstanceGroupScalePolicy(&instancegroup.InstanceGroup{ScalePolicy: tt.spec})

			if err != nil {
				t.Errorf("%v", err)
			}
			if !reflect.DeepEqual(res, tt.expected) {
				t.Errorf("flattenInstanceGroupScalePolicy() got = %v, want %v", res, tt.expected)
			}
		})
	}
}

func TestFlattenInstances(t *testing.T) {
	tests := []struct {
		name     string
		spec     []*instancegroup.ManagedInstance
		expected []map[string]interface{}
	}{
		{
			name: "fixed scale",
			spec: []*instancegroup.ManagedInstance{
				{
					Id:            "id1",
					Status:        instancegroup.ManagedInstance_RUNNING_ACTUAL,
					InstanceId:    "compute_id",
					Fqdn:          "fqdn1",
					Name:          "name1",
					StatusMessage: "status_message1",
					ZoneId:        "zone1",
					NetworkInterfaces: []*instancegroup.NetworkInterface{
						{
							Index: "1",
							PrimaryV6Address: &instancegroup.PrimaryAddress{
								Address: "2001:db8::370:7348",
							},
							SubnetId:   "some-subnet-id",
							MacAddress: "aa-bb-cc-dd-ee-ff",
						},
					},
					StatusChangedAt: &timestamp.Timestamp{Seconds: 500000},
				},
			},
			expected: []map[string]interface{}{
				{
					"status":         "RUNNING_ACTUAL",
					"instance_id":    "compute_id",
					"fqdn":           "fqdn1",
					"name":           "name1",
					"status_message": "status_message1",
					"zone_id":        "zone1",
					"network_interface": []map[string]interface{}{
						{
							"index":        1,
							"mac_address":  "aa-bb-cc-dd-ee-ff",
							"subnet_id":    "some-subnet-id",
							"ipv6":         true,
							"ipv6_address": "2001:db8::370:7348",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := flattenInstances(tt.spec)

			if err != nil {
				t.Errorf("%v", err)
			}
			if !reflect.DeepEqual(res, tt.expected) {
				t.Errorf("flattenInstances() got = %v, want %v", res, tt.expected)
			}
		})
	}
}
