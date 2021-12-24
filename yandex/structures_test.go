package yandex

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/grpc"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
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

func TestExpandStaticRoutes(t *testing.T) {
	cases := []struct {
		name       string
		v          interface{}
		expected   []*vpc.StaticRoute
		shouldFail bool
	}{
		{
			name: "two routes",
			v: schema.NewSet(resourceYandexVPCRouteTableHash, []interface{}{
				map[string]interface{}{
					"destination_prefix": "192.0.2.0/24",
					"next_hop_address":   "192.0.2.1",
				},
				map[string]interface{}{
					"destination_prefix": "198.51.100.0/24",
					"next_hop_address":   "198.51.100.1",
				},
			},
			),
			expected: []*vpc.StaticRoute{
				{
					Destination: &vpc.StaticRoute_DestinationPrefix{DestinationPrefix: "192.0.2.0/24"},
					NextHop:     &vpc.StaticRoute_NextHopAddress{NextHopAddress: "192.0.2.1"},
				},
				{
					Destination: &vpc.StaticRoute_DestinationPrefix{DestinationPrefix: "198.51.100.0/24"},
					NextHop:     &vpc.StaticRoute_NextHopAddress{NextHopAddress: "198.51.100.1"},
				},
			},
			shouldFail: false,
		},
		{
			name:       "missing",
			v:          nil,
			expected:   []*vpc.StaticRoute{},
			shouldFail: false,
		},
		{
			name:       "empty set element",
			v:          schema.NewSet(resourceYandexVPCRouteTableHash, []interface{}{map[string]interface{}{}}),
			shouldFail: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := expandStaticRoutes(tc.v)
			if err != nil && !tc.shouldFail {
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
							"block_size":  0,
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
					"ipv4":        true,
					"ipv6":        false,
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
					"ipv4":           true,
					"ipv6":           false,
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
					"ipv4":         false,
					"ipv6":         true,
					"ipv6_address": "2001:db8::370:7348",
				},
			},
			externalIP: "2001:db8::370:7348",
			internalIP: "",
			wantErr:    false,
		},
		{
			name: "one nic with security group ids",
			instance: &compute.Instance{
				NetworkInterfaces: []*compute.NetworkInterface{
					{
						Index: "1",
						PrimaryV4Address: &compute.PrimaryAddress{
							Address: "192.168.19.16",
						},
						SubnetId:         "some-subnet-id",
						MacAddress:       "aa-bb-cc-dd-ee-ff",
						SecurityGroupIds: []string{"test-sg-id1", "test-sg-id2"},
					},
				},
			},
			want: []map[string]interface{}{
				{
					"index":              1,
					"mac_address":        "aa-bb-cc-dd-ee-ff",
					"subnet_id":          "some-subnet-id",
					"ipv4":               true,
					"ipv6":               false,
					"ip_address":         "192.168.19.16",
					"nat":                false,
					"security_group_ids": append(make([]interface{}, 0), "test-sg-id1", "test-sg-id2"),
				},
			},
			externalIP: "",
			internalIP: "192.168.19.16",
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

func TestExpandComputePrimaryV4AddressSpec(t *testing.T) {
	tests := []struct {
		name string
		data map[string]interface{}
		spec *compute.PrimaryAddressSpec
	}{
		{
			name: "1",
			data: map[string]interface{}{
				"ipv4":       true,
				"ip_address": "10.0.0.1",
			},
			spec: &compute.PrimaryAddressSpec{
				Address:         "10.0.0.1",
				OneToOneNatSpec: nil,
				DnsRecordSpecs:  nil,
			},
		},
		{
			name: "1",
			data: map[string]interface{}{
				"ipv4":       true,
				"ip_address": "10.0.0.1",
				"dns_record": []interface{}{
					map[string]interface{}{
						"fqdn":        "a.example.com.",
						"dns_zone_id": "",
					},
					map[string]interface{}{
						"fqdn":        "b.example.com.",
						"dns_zone_id": "zone_id",
						"ttl":         3600,
						"ptr":         true,
					},
				},
			},
			spec: &compute.PrimaryAddressSpec{
				Address:         "10.0.0.1",
				OneToOneNatSpec: nil,
				DnsRecordSpecs: []*compute.DnsRecordSpec{
					{
						Fqdn:      "a.example.com.",
						DnsZoneId: "",
						Ttl:       0,
						Ptr:       false,
					},
					{
						Fqdn:      "b.example.com.",
						DnsZoneId: "zone_id",
						Ttl:       3600,
						Ptr:       true,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, err := expandPrimaryV4AddressSpec(tt.data)
			if err != nil {
				t.Error(err.Error())
			}
			if !reflect.DeepEqual(tt.spec, spec) {
				t.Errorf("%v not equals to %v", tt.spec, spec)
			}
		})
	}
}
