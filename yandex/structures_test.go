package yandex

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"
	ltagent "github.com/yandex-cloud/go-genproto/yandex/cloud/loadtesting/api/v1/agent"
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
		KmsKey: &compute.KMSKey{
			KeyId:     "mock-key-id",
			VersionId: "mock-key-version",
		},
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
			name: "two routes with gateway",
			v: schema.NewSet(resourceYandexVPCRouteTableHash, []interface{}{
				map[string]interface{}{
					"destination_prefix": "192.0.2.0/24",
					"next_hop_address":   "192.0.2.1",
				},
				map[string]interface{}{
					"destination_prefix": "0.0.0.0/0",
					"gateway_id":         "gateway-id",
				},
			},
			),
			expected: []*vpc.StaticRoute{
				{
					Destination: &vpc.StaticRoute_DestinationPrefix{DestinationPrefix: "0.0.0.0/0"},
					NextHop:     &vpc.StaticRoute_GatewayId{GatewayId: "gateway-id"},
				},
				{
					Destination: &vpc.StaticRoute_DestinationPrefix{DestinationPrefix: "192.0.2.0/24"},
					NextHop:     &vpc.StaticRoute_NextHopAddress{NextHopAddress: "192.0.2.1"},
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
		{
			name: "no next hop",
			v: schema.NewSet(resourceYandexVPCRouteTableHash, []interface{}{
				map[string]interface{}{
					"destination_prefix": "192.0.2.0/24",
				},
			},
			),
			shouldFail: true,
		},
		{
			name: "too many next hops",
			v: schema.NewSet(resourceYandexVPCRouteTableHash, []interface{}{
				map[string]interface{}{
					"destination_prefix": "192.0.2.0/24",
					"next_hop_address":   "192.0.2.1",
					"gateway_id":         "gateway-id",
				},
			},
			),
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
							"kms_key_id":  "mock-key-id",
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

func TestExpandHostAffinityRuleSpec(t *testing.T) {
	tests := []struct {
		name string
		data []interface{}
		spec []*compute.PlacementPolicy_HostAffinityRule
	}{
		{
			name: "empty rule set",
			data: []interface{}{},
			spec: []*compute.PlacementPolicy_HostAffinityRule{},
		},
		{
			name: "rule with host ID",
			data: []interface{}{
				map[string]interface{}{
					"key": "yc.hostId",
					"op":  "IN",
					"values": []interface{}{
						"host-id",
					},
				},
			},
			spec: []*compute.PlacementPolicy_HostAffinityRule{
				{
					Key:    "yc.hostId",
					Op:     compute.PlacementPolicy_HostAffinityRule_IN,
					Values: []string{"host-id"},
				},
			},
		},
		{
			name: "rule with host group ID",
			data: []interface{}{
				map[string]interface{}{
					"key": "yc.hostGroupId",
					"op":  "IN",
					"values": []interface{}{
						"host-group-id-1",
						"host-group-id-2",
					},
				},
			},
			spec: []*compute.PlacementPolicy_HostAffinityRule{
				{
					Key:    "yc.hostGroupId",
					Op:     compute.PlacementPolicy_HostAffinityRule_IN,
					Values: []string{"host-group-id-1", "host-group-id-2"},
				},
			},
		},
		{
			name: "rules with host and group ID",
			data: []interface{}{
				map[string]interface{}{
					"key": "yc.hostId",
					"op":  "IN",
					"values": []interface{}{
						"host-id",
					},
				},
				map[string]interface{}{
					"key": "yc.hostGroupId",
					"op":  "IN",
					"values": []interface{}{
						"host-group-id-1",
						"host-group-id-2",
					},
				},
			},
			spec: []*compute.PlacementPolicy_HostAffinityRule{
				{
					Key:    "yc.hostId",
					Op:     compute.PlacementPolicy_HostAffinityRule_IN,
					Values: []string{"host-id"},
				},
				{
					Key:    "yc.hostGroupId",
					Op:     compute.PlacementPolicy_HostAffinityRule_IN,
					Values: []string{"host-group-id-1", "host-group-id-2"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := expandHostAffinityRulesSpec(tt.data)
			if !reflect.DeepEqual(tt.spec, spec) {
				t.Errorf("%v not equals to %v", tt.spec, spec)
			}
		})
	}
}

func TestExpandLocalDiskSpecs(t *testing.T) {
	tests := []struct {
		name string
		data []interface{}
		spec []*compute.AttachedLocalDiskSpec
	}{
		{
			name: "empty specs by default",
			data: nil,
			spec: nil,
		},
		{
			name: "one local disk",
			data: []interface{}{
				map[string]interface{}{
					"size_bytes": 1,
				},
			},
			spec: []*compute.AttachedLocalDiskSpec{
				{
					Size: 1,
				},
			},
		},
		{
			name: "two local disk",
			data: []interface{}{
				map[string]interface{}{
					"size_bytes": 100,
				},
				map[string]interface{}{
					"size_bytes": 200,
				},
			},
			spec: []*compute.AttachedLocalDiskSpec{
				{
					Size: 100,
				},
				{
					Size: 200,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := expandLocalDiskSpecs(tt.data)
			if !reflect.DeepEqual(tt.spec, spec) {
				t.Errorf("%v is not equal to %v", tt.spec, spec)
			}
		})
	}
}

func TestFlattenLocalDiskLocalDisks(t *testing.T) {
	tests := []struct {
		name     string
		instance *compute.Instance
		expected []interface{}
	}{
		{
			name:     "no local disks",
			instance: &compute.Instance{},
			expected: nil,
		},
		{
			name: "one local disk",
			instance: &compute.Instance{
				LocalDisks: []*compute.AttachedLocalDisk{
					{
						Size:       1,
						DeviceName: "nvme-disk-0",
					},
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"size_bytes":  1,
					"device_name": "nvme-disk-0",
				},
			},
		},
		{
			name: "two local disks",
			instance: &compute.Instance{
				LocalDisks: []*compute.AttachedLocalDisk{
					{
						Size:       100,
						DeviceName: "nvme-disk-0",
					},
					{
						Size:       200,
						DeviceName: "nvme-disk-1",
					},
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"size_bytes":  100,
					"device_name": "nvme-disk-0",
				},
				map[string]interface{}{
					"size_bytes":  200,
					"device_name": "nvme-disk-1",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expected := flattenLocalDisks(tt.instance)
			if !reflect.DeepEqual(tt.expected, expected) {
				t.Errorf("%#v is not equal to %#v", tt.expected, expected)
			}
		})
	}
}

func TestConvertFQDN(t *testing.T) {
	testdata := map[string]string{
		"123.auto.internal":                 "",
		"breathtaking.ru-central1.internal": "breathtaking",
		"hello.world":                       "hello.world",
		"breathtaking":                      "breathtaking.",
	}

	for fqdn, hostname := range testdata {
		t.Run("fqdn "+fqdn, func(t *testing.T) {
			h, _ := parseHostnameFromFQDN(fqdn)
			if h != hostname {
				t.Errorf("%s is not equal to %s", h, hostname)
			}
		})
	}
}

// test expandContainerRepositoryLifecyclePolicyRules
// test flattenContainerRepositoryLifecyclePolicyRules

func TestFlattenContainerRepositoryLifecyclePolicyRule(t *testing.T) {
	t.Parallel()

	t.Run("FlattenContainerRepositoryLifecyclePolicyRule", func(t *testing.T) {
		test := struct {
			rule     *containerregistry.LifecycleRule
			expected interface{}
		}{
			rule: &containerregistry.LifecycleRule{
				Description:  "test description",
				ExpirePeriod: durationpb.New(24 * time.Hour),
				TagRegexp:    ".*",
				Untagged:     true,
				RetainedTop:  int64(5),
			},
			expected: map[string]interface{}{
				"description":   "test description",
				"expire_period": "24h0m0s",
				"tag_regexp":    ".*",
				"untagged":      true,
				"retained_top":  int64(5),
			},
		}

		got := flattenContainerRepositoryLifecyclePolicyRule(test.rule)
		if !reflect.DeepEqual(test.expected, got) {
			t.Errorf("%#v is not equal to %#v", test.expected, got)
		}
	})
}

func TestFlattenSnapshotScheduleSchedulePolicy(t *testing.T) {
	t.Parallel()

	t.Run("flattenSnapshotScheduleSchedulePolicy", func(t *testing.T) {
		test := struct {
			policy   *compute.SchedulePolicy
			expected []map[string]interface{}
		}{
			policy: &compute.SchedulePolicy{
				Expression: "* * * * *",
			},
			expected: []map[string]interface{}{
				{
					"expression": "* * * * *", "start_at": "",
				},
			},
		}

		got, err := flattenSnapshotScheduleSchedulePolicy(test.policy)
		if err != nil {
			t.Fatalf("Invalid schedule policy: %v", err)
		}
		if !reflect.DeepEqual(test.expected, got) {
			t.Errorf("%#v is not equal to %#v", test.expected, got)
		}
	})
}

func TestFlattenSnapshotScheduleSnapshotSpec(t *testing.T) {
	t.Parallel()

	t.Run("flattenSnapshotScheduleSnapshotSpec", func(t *testing.T) {
		test := struct {
			spec     *compute.SnapshotSpec
			expected []map[string]interface{}
		}{
			spec: &compute.SnapshotSpec{
				Description: "test description",
				Labels: map[string]string{
					"foo": "bar",
				},
			},
			expected: []map[string]interface{}{
				{
					"description": "test description",
					"labels": map[string]string{
						"foo": "bar",
					},
				},
			},
		}

		got, err := flattenSnapshotScheduleSnapshotSpec(test.spec)
		if err != nil {
			t.Fatalf("Invalid spec: %v", err)
		}
		if !reflect.DeepEqual(test.expected, got) {
			t.Errorf("%#v is not equal to %#v", test.expected, got)
		}
	})
}

func TestFlattenLoadtestingAgentLogSettingsParams(t *testing.T) {
	cases := []struct {
		name  string
		agent *ltagent.Agent
		want  []map[string]interface{}
	}{
		{
			name: "log group id specified",
			agent: &ltagent.Agent{
				LogSettings: &ltagent.LogSettings{
					CloudLogGroupId: "abcloggroupid",
				},
			},
			want: []map[string]interface{}{
				{
					"log_group_id": "abcloggroupid",
				},
			},
		},
		{
			name: "log group id not specified",
			agent: &ltagent.Agent{
				LogSettings: &ltagent.LogSettings{},
			},
			want: []map[string]interface{}{
				{
					"log_group_id": "",
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := flattenLoadtestingAgentLogSettingsParams(tc.agent)
			if err != nil {
				t.Errorf("flattenLoadtestingAgentLogSettingsParams() error = %v", err)
				return
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("flattenLoadtestingAgentLogSettingsParams()\ngot\n%v\nwant\n%v\n", got, tc.want)
				return
			}
		})
	}
}

func TestFlattenLoadtestingComputeInstanceResources(t *testing.T) {
	cases := []struct {
		name      string
		resources *compute.Resources
		want      []map[string]interface{}
	}{
		{
			name: "cores 1 fraction 100 memory 5 gb",
			resources: &compute.Resources{
				Cores:        1,
				CoreFraction: 100,
				Memory:       5 * (1 << 30),
			},
			want: []map[string]interface{}{
				{
					"cores":         1,
					"core_fraction": 100,
					"memory":        5.0,
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
			want: []map[string]interface{}{
				{
					"cores":         8,
					"core_fraction": 5,
					"memory":        16.0,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := flattenLoadtestingComputeInstanceResources(&compute.Instance{Resources: tc.resources})
			if err != nil {
				t.Errorf("flattenLoadtestingComputeInstanceResources() error = %v", err)
				return
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("flattenLoadtestingComputeInstanceResources()\ngot\n%v\nwant\n%v\n", got, tc.want)
				return
			}
		})
	}
}

func TestFlattenLoadtestingComputeInstanceBootDisk(t *testing.T) {
	cases := []struct {
		name     string
		bootDisk *compute.AttachedDisk
		want     []map[string]interface{}
	}{
		{
			name: "boot disk with diskID",
			bootDisk: &compute.AttachedDisk{
				Mode:       compute.AttachedDisk_READ_WRITE,
				DeviceName: "test-device-name",
				AutoDelete: false,
				DiskId:     "saeque9k",
			},
			want: []map[string]interface{}{
				{
					"device_name": "test-device-name",
					"auto_delete": false,
					"disk_id":     "saeque9k",
					"initialize_params": []map[string]interface{}{
						{
							"name":        "mock-disk-name",
							"description": "mock-disk-description",
							"size":        4,
							"block_size":  0,
							"type":        "network-hdd",
						},
					},
				},
			},
		},
	}

	reducedDiskClient := &DiskClientGetter{}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := flattenLoadtestingComputeInstanceBootDisk(context.Background(), &compute.Instance{BootDisk: tc.bootDisk}, reducedDiskClient)

			if err != nil {
				t.Errorf("flattenLoadtestingComputeInstanceBootDisk() error = %v", err)
				return
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("flattenLoadtestingComputeInstanceBootDisk()\ngot\n%v\nwant\n%v\n", got, tc.want)
				return
			}
		})
	}
}

func TestFlattenLoadtestingComputeInstanceNetworkInterfaces(t *testing.T) {
	tests := []struct {
		name     string
		instance *compute.Instance
		want     []map[string]interface{}
	}{
		{
			name: "no nics defined",
			instance: &compute.Instance{
				NetworkInterfaces: []*compute.NetworkInterface{},
			},
			want: []map[string]interface{}{},
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
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := flattenLoadtestingComputeInstanceNetworkInterfaces(tt.instance)
			if err != nil {
				t.Errorf("flattenLoadtestingComputeInstanceNetworkInterfaces() error = %v", err)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("flattenLoadtestingComputeInstanceNetworkInterfaces()\ngot\n%v\nwant\n%v\n", got, tt.want)
			}
		})
	}
}
