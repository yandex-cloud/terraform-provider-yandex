package yandex

import (
	"reflect"
	"testing"

	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1/instancegroup"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
)

func TestParseInstanceGroupNetworkSettingsType(t *testing.T) {
	cases := []struct {
		name   string
		nsType string
		parsed instancegroup.NetworkSettings_Type
	}{
		{
			name:   "soft",
			nsType: "SOFTWARE_ACCELERATED",
			parsed: instancegroup.NetworkSettings_SOFTWARE_ACCELERATED,
		},
		{
			name:   "base",
			nsType: "STANDARD",
			parsed: instancegroup.NetworkSettings_STANDARD,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parseInstanceGroupNetworkSettingsType(tc.nsType)
			if err != nil {
				t.Fatalf("bad: %#v", err)
			}
			if !reflect.DeepEqual(result, tc.parsed) {
				t.Fatalf("Got:\n\n%#v\n\nExpected:\n\n%#v\n", result, tc.parsed)
			}
		})
	}
}

func TestParseInstanceGroupMetadataOptions(t *testing.T) {
	cases := []struct {
		name     string
		moStatus int
		parsed   instancegroup.MetadataOption
	}{
		{
			name:     "enabled",
			moStatus: 1,
			parsed:   instancegroup.MetadataOption_ENABLED,
		},
		{
			name:     "disabled",
			moStatus: 2,
			parsed:   instancegroup.MetadataOption_DISABLED,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parseInstanceGroupMetadataOption(tc.moStatus)
			if err != nil {
				t.Fatalf("bad: %#v", err)
			}
			if !reflect.DeepEqual(result, tc.parsed) {
				t.Fatalf("Got:\n\n%#v\n\nExpected:\n\n%#v\n", result, tc.parsed)
			}
		})
	}
}

func TestFlattenInstanceGroupVariable(t *testing.T) {
	cases := []struct {
		name     string
		v        []*instancegroup.Variable
		expected map[string]string
	}{
		{
			name:     "test1",
			v:        append(make([]*instancegroup.Variable, 0), &instancegroup.Variable{Key: "test_key", Value: "test_value"}),
			expected: map[string]string{"test_key": "test_value"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := flattenInstanceGroupVariable(tc.v)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Fatalf("Got:\n\n%#v\n\nExpected:\n\n%#v\n", result, tc.expected)
			}
		})
	}
}

func TestExpandInstanceGroupVariables(t *testing.T) {
	cases := []struct {
		name     string
		v        interface{}
		expected *instancegroup.Variable
	}{
		{
			name:     "soft",
			v:        map[string]interface{}{"test_key": "test_value"},
			expected: &instancegroup.Variable{Key: "test_key", Value: "test_value"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := expandInstanceGroupVariables(tc.v)
			if err != nil {
				t.Fatalf("bad: %#v", err)
			}
			if !reflect.DeepEqual(result[0], tc.expected) {
				t.Fatalf("Got:\n\n%#v\n\nExpected:\n\n%#v\n", result, tc.expected)
			}
		})
	}
}

func TestFlattenInstanceGroupNetworkSettings(t *testing.T) {
	cases := []struct {
		name     string
		ns       *instancegroup.NetworkSettings
		expected []map[string]interface{}
	}{
		{
			name:     "soft",
			ns:       &instancegroup.NetworkSettings{Type: instancegroup.NetworkSettings_SOFTWARE_ACCELERATED},
			expected: []map[string]interface{}{{"type": "SOFTWARE_ACCELERATED"}},
		},
		{
			name:     "base",
			ns:       &instancegroup.NetworkSettings{Type: instancegroup.NetworkSettings_STANDARD},
			expected: []map[string]interface{}{{"type": "STANDARD"}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := flattenInstanceGroupNetworkSettings(tc.ns)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Fatalf("Got:\n\n%#v\n\nExpected:\n\n%#v\n", result, tc.expected)
			}
		})
	}
}

func TestFlattenInstanceGroupMetadataOptions(t *testing.T) {
	cases := []struct {
		name     string
		mo       *instancegroup.MetadataOptions
		expected []map[string]interface{}
	}{
		{
			name: "all_enabled",
			mo: &instancegroup.MetadataOptions{
				GceHttpEndpoint:   instancegroup.MetadataOption_ENABLED,
				AwsV1HttpEndpoint: instancegroup.MetadataOption_ENABLED,
				GceHttpToken:      instancegroup.MetadataOption_ENABLED,
				AwsV1HttpToken:    instancegroup.MetadataOption_ENABLED,
			},
			expected: []map[string]interface{}{{
				"gce_http_endpoint":    1,
				"aws_v1_http_endpoint": 1,
				"gce_http_token":       1,
				"aws_v1_http_token":    1,
			}},
		},
		{
			name: "all_disabled",
			mo: &instancegroup.MetadataOptions{
				GceHttpEndpoint:   instancegroup.MetadataOption_DISABLED,
				AwsV1HttpEndpoint: instancegroup.MetadataOption_DISABLED,
				GceHttpToken:      instancegroup.MetadataOption_DISABLED,
				AwsV1HttpToken:    instancegroup.MetadataOption_DISABLED,
			},
			expected: []map[string]interface{}{{
				"gce_http_endpoint":    2,
				"aws_v1_http_endpoint": 2,
				"gce_http_token":       2,
				"aws_v1_http_token":    2,
			}},
		},
		{
			name: "fifty_fifty",
			mo: &instancegroup.MetadataOptions{
				GceHttpEndpoint:   instancegroup.MetadataOption_ENABLED,
				AwsV1HttpEndpoint: instancegroup.MetadataOption_DISABLED,
				GceHttpToken:      instancegroup.MetadataOption_DISABLED,
				AwsV1HttpToken:    instancegroup.MetadataOption_ENABLED,
			},
			expected: []map[string]interface{}{{
				"gce_http_endpoint":    1,
				"aws_v1_http_endpoint": 2,
				"gce_http_token":       2,
				"aws_v1_http_token":    1,
			}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := flattenInstanceGroupMetadataOptions(tc.mo)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Fatalf("Got:\n\n%#v\n\nExpected:\n\n%#v\n", result, tc.expected)
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
					"ipv4":        true,
					"ipv6":        false,
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
					"ipv4":           true,
					"ipv6":           false,
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
					"ipv4":         false,
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
				Name: "boot-disk-name",
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
				"name": "boot-disk-name",
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
			res, _, err := flattenInstanceGroupHealthChecks(&instancegroup.InstanceGroup{HealthChecksSpec: tt.spec})

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
						AutoScaleType:       instancegroup.ScalePolicy_AutoScale_REGIONAL,
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
							"auto_scale_type":      "REGIONAL",
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
						AutoScaleType:         instancegroup.ScalePolicy_AutoScale_ZONAL,
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
							"auto_scale_type":        "ZONAL",
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
		{
			name: "auto scale with custom rules",
			spec: &instancegroup.ScalePolicy{
				ScaleType: &instancegroup.ScalePolicy_AutoScale_{
					AutoScale: &instancegroup.ScalePolicy_AutoScale{
						AutoScaleType:         instancegroup.ScalePolicy_AutoScale_ZONAL,
						MinZoneSize:           1,
						MaxSize:               2,
						MeasurementDuration:   &duration.Duration{Seconds: 10},
						WarmupDuration:        &duration.Duration{Seconds: 20},
						StabilizationDuration: &duration.Duration{Seconds: 30},
						InitialSize:           3,
						CustomRules: []*instancegroup.ScalePolicy_CustomRule{
							{
								RuleType:   instancegroup.ScalePolicy_CustomRule_UTILIZATION,
								MetricType: instancegroup.ScalePolicy_CustomRule_GAUGE,
								MetricName: "metric1",
								Target:     20.5,
								Labels:     map[string]string{},
							},
							{
								RuleType:   instancegroup.ScalePolicy_CustomRule_WORKLOAD,
								MetricType: instancegroup.ScalePolicy_CustomRule_COUNTER,
								MetricName: "metric2",
								Target:     25,
								Labels:     map[string]string{"label1": "value1", "label2": "value2"},
							},
						},
					},
				},
			},
			expected: []map[string]interface{}{
				{
					"auto_scale": []map[string]interface{}{
						{
							"auto_scale_type":        "ZONAL",
							"min_zone_size":          1,
							"max_size":               2,
							"initial_size":           3,
							"measurement_duration":   10,
							"warmup_duration":        20,
							"stabilization_duration": 30,
							"custom_rule": []map[string]interface{}{
								{
									"rule_type":   "UTILIZATION",
									"metric_type": "GAUGE",
									"metric_name": "metric1",
									"target":      20.5,
									"labels":      map[string]string{},
								},
								{
									"rule_type":   "WORKLOAD",
									"metric_type": "COUNTER",
									"metric_name": "metric2",
									"target":      25.,
									"labels":      map[string]string{"label1": "value1", "label2": "value2"},
								},
							},
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
					"instance_tag":   "",
					"fqdn":           "fqdn1",
					"name":           "name1",
					"status_message": "status_message1",
					"zone_id":        "zone1",
					"network_interface": []map[string]interface{}{
						{
							"index":        1,
							"mac_address":  "aa-bb-cc-dd-ee-ff",
							"subnet_id":    "some-subnet-id",
							"ipv4":         false,
							"ipv6":         true,
							"ipv6_address": "2001:db8::370:7348",
						},
					},
					"status_changed_at": "1970-01-06T18:53:20Z",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := flattenInstanceGroupManagedInstances(tt.spec)

			if err != nil {
				t.Errorf("%v", err)
			}
			if !reflect.DeepEqual(res, tt.expected) {
				t.Errorf("flattenInstances() got = %v, want %v", res, tt.expected)
			}
		})
	}
}

func TestFlattenRules(t *testing.T) {
	tests := []struct {
		name     string
		spec     []*vpc.SecurityGroupRule
		expected *schema.Set
	}{
		{
			name: "2 rules",
			spec: []*vpc.SecurityGroupRule{
				{
					Id:          "21",
					Description: "desc1",
					Labels: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
					Direction: 1,
					Ports: &vpc.PortRange{
						FromPort: 22,
						ToPort:   23,
					},
					ProtocolName:   "TCP",
					ProtocolNumber: 6,
					Target: &vpc.SecurityGroupRule_CidrBlocks{
						CidrBlocks: &vpc.CidrBlocks{
							V4CidrBlocks: []string{"10.0.0.0/24"},
						},
					},
				},
				{
					Id:          "22",
					Description: "desc2",
					Labels: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
					Direction: 2,
					Ports: &vpc.PortRange{
						FromPort: 25,
						ToPort:   25,
					},
					ProtocolName:   "ANY",
					ProtocolNumber: -1,
					Target: &vpc.SecurityGroupRule_CidrBlocks{
						CidrBlocks: &vpc.CidrBlocks{
							V4CidrBlocks: []string{"10.0.3.0/24"},
						},
					},
				},
				{
					Id:          "23",
					Description: "desc3",
					Labels: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
					Direction: 2,
					Ports: &vpc.PortRange{
						FromPort: 1,
						ToPort:   65535,
					},
					ProtocolName: "ICMP",
					Target: &vpc.SecurityGroupRule_CidrBlocks{
						CidrBlocks: &vpc.CidrBlocks{
							V4CidrBlocks: []string{"10.0.0.0/24", "10.0.1.0/24"},
						},
					},
				},
				{
					Id:        "24",
					Direction: 2,
					Ports: &vpc.PortRange{
						FromPort: 80,
						ToPort:   80,
					},
					ProtocolName: "TCP",
					Target: &vpc.SecurityGroupRule_PredefinedTarget{
						PredefinedTarget: "self_security_group",
					},
				},
				{
					Id:           "25",
					Direction:    2,
					ProtocolName: "TCP",
					Target: &vpc.SecurityGroupRule_SecurityGroupId{
						SecurityGroupId: "some_sg_id",
					},
				},
			},
			expected: schema.NewSet(resourceYandexVPCSecurityGroupRuleHash, []interface{}{
				map[string]interface{}{
					"id":          "22",
					"description": "desc2",
					"labels": map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
					"v4_cidr_blocks":    []interface{}{"10.0.3.0/24"},
					"protocol":          "ANY",
					"port":              int64(25),
					"from_port":         int64(-1),
					"to_port":           int64(-1),
					"security_group_id": "",
					"predefined_target": "",
				},
				map[string]interface{}{
					"id":          "23",
					"description": "desc3",
					"labels": map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
					"v4_cidr_blocks":    []interface{}{"10.0.0.0/24", "10.0.1.0/24"},
					"protocol":          "ICMP",
					"port":              int64(-1),
					"from_port":         int64(1),
					"to_port":           int64(65535),
					"security_group_id": "",
					"predefined_target": "",
				},
				map[string]interface{}{
					"id":                "24",
					"description":       "",
					"labels":            map[string]string{},
					"protocol":          "TCP",
					"port":              int64(80),
					"from_port":         int64(-1),
					"to_port":           int64(-1),
					"security_group_id": "",
					"predefined_target": "self_security_group",
				},
				map[string]interface{}{
					"id":                "24",
					"description":       "",
					"labels":            map[string]string{},
					"protocol":          "TCP",
					"port":              int64(-1),
					"from_port":         int64(-1),
					"to_port":           int64(-1),
					"security_group_id": "some_sg_id",
					"predefined_target": "",
				},
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, egress := flattenSecurityGroupRulesSpec(tt.spec)

			if egress.Difference(tt.expected).Len() > 0 {
				t.Errorf("flattenInstances() got = %v, want %v", egress.List(), tt.expected.List())
			}
		})
	}
}

func TestExpandNetworkSettings(t *testing.T) {
	cases := []struct {
		name     string
		ns       interface{}
		expected *instancegroup.NetworkSettings
	}{
		{
			name:     "soft",
			ns:       "SOFTWARE_ACCELERATED",
			expected: &instancegroup.NetworkSettings{Type: instancegroup.NetworkSettings_SOFTWARE_ACCELERATED},
		},
		{
			name:     "base",
			ns:       "STANDARD",
			expected: &instancegroup.NetworkSettings{Type: instancegroup.NetworkSettings_STANDARD},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := expandInstanceGroupNetworkSettings(tc.ns)
			if err != nil {
				t.Fatalf("bad: %#v", err)
			}
			if !reflect.DeepEqual(result, tc.expected) {
				t.Fatalf("Got:\n\n%#v\n\nExpected:\n\n%#v\n", result, tc.expected)
			}
		})
	}
}

func TestExpandMetadataOptions(t *testing.T) {
	cases := []struct {
		name     string
		mo       map[string]interface{}
		expected *instancegroup.MetadataOptions
	}{
		{
			name: "all_enabled",
			mo: map[string]interface{}{
				"gce_http_endpoint":    1,
				"aws_v1_http_endpoint": 1,
				"gce_http_token":       1,
				"aws_v1_http_token":    1,
			},
			expected: &instancegroup.MetadataOptions{
				GceHttpEndpoint:   instancegroup.MetadataOption_ENABLED,
				AwsV1HttpEndpoint: instancegroup.MetadataOption_ENABLED,
				GceHttpToken:      instancegroup.MetadataOption_ENABLED,
				AwsV1HttpToken:    instancegroup.MetadataOption_ENABLED,
			},
		},
		{
			name: "all_disabled",
			mo: map[string]interface{}{
				"gce_http_endpoint":    2,
				"aws_v1_http_endpoint": 2,
				"gce_http_token":       2,
				"aws_v1_http_token":    2,
			},
			expected: &instancegroup.MetadataOptions{
				GceHttpEndpoint:   instancegroup.MetadataOption_DISABLED,
				AwsV1HttpEndpoint: instancegroup.MetadataOption_DISABLED,
				GceHttpToken:      instancegroup.MetadataOption_DISABLED,
				AwsV1HttpToken:    instancegroup.MetadataOption_DISABLED,
			},
		},
		{
			name: "fifty_fifty",
			mo: map[string]interface{}{
				"gce_http_endpoint":    1,
				"aws_v1_http_endpoint": 2,
				"gce_http_token":       2,
				"aws_v1_http_token":    1,
			},
			expected: &instancegroup.MetadataOptions{
				GceHttpEndpoint:   instancegroup.MetadataOption_ENABLED,
				AwsV1HttpEndpoint: instancegroup.MetadataOption_DISABLED,
				GceHttpToken:      instancegroup.MetadataOption_DISABLED,
				AwsV1HttpToken:    instancegroup.MetadataOption_ENABLED,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := expandInstanceGroupMetadataOptions(tc.mo)
			if err != nil {
				t.Fatalf("bad: %#v", err)
			}
			if !reflect.DeepEqual(result, tc.expected) {
				t.Fatalf("Got:\n\n%#v\n\nExpected:\n\n%#v\n", result, tc.expected)
			}
		})
	}
}

func TestExpandInstanceGroupNetworkInterface(t *testing.T) {
	cases := []struct {
		name     string
		data     map[string]interface{}
		expected *instancegroup.NetworkInterfaceSpec
	}{
		{
			name: "no_dns_record",
			data: map[string]interface{}{
				"network_id": "net_id",
				"subnet_ids": schema.NewSet(schema.HashString, []interface{}{
					"subnet1",
					"subnet2",
				}),
				"ipv4": true,
				"nat":  true,
			},
			expected: &instancegroup.NetworkInterfaceSpec{
				NetworkId: "net_id",
				SubnetIds: []string{"subnet1", "subnet2"},
				PrimaryV4AddressSpec: &instancegroup.PrimaryAddressSpec{
					OneToOneNatSpec: &instancegroup.OneToOneNatSpec{IpVersion: instancegroup.IpVersion_IPV4},
					DnsRecordSpecs:  nil,
				},
				PrimaryV6AddressSpec: nil,
				SecurityGroupIds:     nil,
			},
		},
		{
			name: "ip_address",
			data: map[string]interface{}{
				"network_id": "net_id",
				"subnet_ids": schema.NewSet(schema.HashString, []interface{}{
					"subnet1",
					"subnet2",
				}),
				"ip_address": "10.0.0.1",
				"nat":        true,
			},
			expected: &instancegroup.NetworkInterfaceSpec{
				NetworkId: "net_id",
				SubnetIds: []string{"subnet1", "subnet2"},
				PrimaryV4AddressSpec: &instancegroup.PrimaryAddressSpec{
					Address:         "10.0.0.1",
					OneToOneNatSpec: &instancegroup.OneToOneNatSpec{IpVersion: instancegroup.IpVersion_IPV4},
					DnsRecordSpecs:  nil,
				},
				PrimaryV6AddressSpec: nil,
				SecurityGroupIds:     nil,
			},
		},
		{
			name: "dns_record",
			data: map[string]interface{}{
				"network_id": "net_id",
				"subnet_ids": schema.NewSet(schema.HashString, []interface{}{
					"subnet1",
					"subnet2",
				}),
				"ipv4": true,
				"nat":  false,
				"dns_record": []interface{}{
					map[string]interface{}{
						"fqdn": "fqdn",
					},
					map[string]interface{}{
						"fqdn":        "fqdn2",
						"dns_zone_id": "zone2",
						"ttl":         200,
						"ptr":         true,
					},
				},
			},
			expected: &instancegroup.NetworkInterfaceSpec{
				NetworkId: "net_id",
				SubnetIds: []string{"subnet1", "subnet2"},
				PrimaryV4AddressSpec: &instancegroup.PrimaryAddressSpec{
					DnsRecordSpecs: []*instancegroup.DnsRecordSpec{
						{Fqdn: "fqdn", DnsZoneId: "", Ttl: 0, Ptr: false},
						{Fqdn: "fqdn2", DnsZoneId: "zone2", Ttl: 200, Ptr: true},
					},
				},
				PrimaryV6AddressSpec: nil,
				SecurityGroupIds:     nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := expandInstanceGroupNetworkInterfaceSpec(tc.data)
			if err != nil {
				t.Fatalf(err.Error())
			}
			if !reflect.DeepEqual(result, tc.expected) {
				t.Fatalf("Got:\n\n%#v\n\nExpected:\n\n%#v\n", result, tc.expected)
			}
		})
	}
}

func TestFlattenInstanceGroupPlacementPolicy(t *testing.T) {
	tests := []struct {
		name     string
		spec     *instancegroup.PlacementPolicy
		expected []map[string]interface{}
	}{
		{
			name: "normal",
			spec: &instancegroup.PlacementPolicy{
				PlacementGroupId: "123",
			},
			expected: []map[string]interface{}{
				{
					"placement_group_id": "123",
				},
			},
		},
		{
			name:     "nil",
			spec:     nil,
			expected: nil,
		},
		{
			name: "old nil",
			spec: &instancegroup.PlacementPolicy{
				PlacementGroupId: "",
			},
			expected: []map[string]interface{}{
				{
					"placement_group_id": "",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := flattenInstanceGroupPlacementPolicy(tt.spec)

			if err != nil {
				t.Errorf("%v", err)
			}
			if !reflect.DeepEqual(res, tt.expected) {
				t.Errorf("flattenInstanceGroupScalePolicy() got = %v, want %v", res, tt.expected)
			}
		})
	}
}
