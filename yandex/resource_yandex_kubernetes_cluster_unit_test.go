package yandex

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/k8s/v1"
)

func TestGetIPAllocationPolicy(t *testing.T) {
	tests := []struct {
		name     string
		raw      map[string]interface{}
		expected *k8s.IPAllocationPolicy
	}{
		{
			name:     "empty values",
			raw:      map[string]interface{}{},
			expected: &k8s.IPAllocationPolicy{},
		},
		{
			name: "ipv4 singular only",
			raw: map[string]interface{}{
				"cluster_ipv4_range": "10.20.0.0/16",
			},
			expected: &k8s.IPAllocationPolicy{
				ClusterIpv4CidrBlock: "10.20.0.0/16",
			},
		},
		{
			name: "ipv6 singular only",
			raw: map[string]interface{}{
				"cluster_ipv6_range": "fc00::/96",
			},
			expected: &k8s.IPAllocationPolicy{
				ClusterIpv6CidrBlock: "fc00::/96",
			},
		},
		{
			name: "ipv4 list only",
			raw: map[string]interface{}{
				"cluster_ipv4_ranges": []interface{}{"10.20.0.0/16", "10.22.0.0/16"},
			},
			expected: &k8s.IPAllocationPolicy{
				ClusterIpv4CidrBlocks: []string{"10.20.0.0/16", "10.22.0.0/16"},
			},
		},
		{
			name: "ipv6 list only",
			raw: map[string]interface{}{
				"cluster_ipv6_ranges": []interface{}{"fc00::/96", "fc02::/96"},
			},
			expected: &k8s.IPAllocationPolicy{
				ClusterIpv6CidrBlocks: []string{"fc00::/96", "fc02::/96"},
			},
		},
		{
			name: "ipv4 list and ipv6 singular",
			raw: map[string]interface{}{
				"cluster_ipv4_ranges": []interface{}{"10.20.0.0/16", "10.22.0.0/16"},
				"cluster_ipv6_range":  "fc00::/96",
			},
			expected: &k8s.IPAllocationPolicy{
				ClusterIpv4CidrBlocks: []string{"10.20.0.0/16", "10.22.0.0/16"},
				ClusterIpv6CidrBlock:  "fc00::/96",
			},
		},
		{
			name: "ipv6 list and ipv4 singular",
			raw: map[string]interface{}{
				"cluster_ipv4_range":  "10.20.0.0/16",
				"cluster_ipv6_ranges": []interface{}{"fc00::/96", "fc02::/96"},
			},
			expected: &k8s.IPAllocationPolicy{
				ClusterIpv4CidrBlock:  "10.20.0.0/16",
				ClusterIpv6CidrBlocks: []string{"fc00::/96", "fc02::/96"},
			},
		},
		{
			name: "both lists",
			raw: map[string]interface{}{
				"cluster_ipv4_ranges": []interface{}{"10.20.0.0/16", "10.22.0.0/16"},
				"cluster_ipv6_ranges": []interface{}{"fc00::/96", "fc02::/96"},
			},
			expected: &k8s.IPAllocationPolicy{
				ClusterIpv4CidrBlocks: []string{"10.20.0.0/16", "10.22.0.0/16"},
				ClusterIpv6CidrBlocks: []string{"fc00::/96", "fc02::/96"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resourceData := schema.TestResourceDataRaw(t, resourceYandexKubernetesCluster().Schema, test.raw)
			actual := getIPAllocationPolicy(resourceData)
			test.expected.NodeIpv4CidrMaskSize = 24
			require.Equal(t, test.expected, actual)
		})
	}
}

func TestFlattenKubernetesClusterIPAllocation(t *testing.T) {
	tests := []struct {
		name           string
		policy         *k8s.IPAllocationPolicy
		wantIPv4Range  string
		wantIPv6Range  string
		wantIPv4Ranges []interface{}
		wantIPv6Ranges []interface{}
	}{
		{
			name:           "empty policy",
			policy:         &k8s.IPAllocationPolicy{},
			wantIPv4Ranges: []interface{}{},
			wantIPv6Ranges: []interface{}{},
		},
		{
			name: "singular only",
			policy: &k8s.IPAllocationPolicy{
				ClusterIpv4CidrBlock: "10.20.0.0/16",
				ClusterIpv6CidrBlock: "fc00::/96",
			},
			wantIPv4Range:  "10.20.0.0/16",
			wantIPv6Range:  "fc00::/96",
			wantIPv4Ranges: []interface{}{},
			wantIPv6Ranges: []interface{}{},
		},
		{
			name: "api returned blocks and legacy first cidr",
			policy: &k8s.IPAllocationPolicy{
				ClusterIpv4CidrBlock:  "10.20.0.0/16",
				ClusterIpv6CidrBlock:  "fc00::/96",
				ClusterIpv4CidrBlocks: []string{"10.20.0.0/16", "10.22.0.0/16"},
				ClusterIpv6CidrBlocks: []string{"fc00::/96", "fc02::/96"},
			},
			wantIPv4Range:  "10.20.0.0/16",
			wantIPv6Range:  "fc00::/96",
			wantIPv4Ranges: []interface{}{"10.20.0.0/16", "10.22.0.0/16"},
			wantIPv6Ranges: []interface{}{"fc00::/96", "fc02::/96"},
		},
		{
			name: "api returned blocks only",
			policy: &k8s.IPAllocationPolicy{
				ClusterIpv4CidrBlocks: []string{"10.20.0.0/16", "10.22.0.0/16"},
				ClusterIpv6CidrBlocks: []string{"fc00::/96", "fc02::/96"},
			},
			wantIPv4Ranges: []interface{}{"10.20.0.0/16", "10.22.0.0/16"},
			wantIPv6Ranges: []interface{}{"fc00::/96", "fc02::/96"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, resourceYandexKubernetesCluster().Schema, map[string]interface{}{})
			err := flattenKubernetesClusterAttributes(testK8SClusterWithIPPolicy(test.policy), d, true)
			require.NoError(t, err)

			assert.Equal(t, test.wantIPv4Range, d.Get("cluster_ipv4_range"))
			assert.Equal(t, test.wantIPv6Range, d.Get("cluster_ipv6_range"))
			assert.Equal(t, test.wantIPv4Ranges, d.Get("cluster_ipv4_ranges"))
			assert.Equal(t, test.wantIPv6Ranges, d.Get("cluster_ipv6_ranges"))
		})
	}
}

func testK8SClusterWithIPPolicy(policy *k8s.IPAllocationPolicy) *k8s.Cluster {
	return &k8s.Cluster{
		Id:                 "cluster-id",
		IpAllocationPolicy: policy,
		Master: &k8s.Master{
			Version: "1.33",
			MaintenancePolicy: &k8s.MasterMaintenancePolicy{
				AutoUpgrade: true,
			},
			MasterType: &k8s.Master_ZonalMaster{
				ZonalMaster: &k8s.ZonalMaster{
					ZoneId: "ru-central1-a",
				},
			},
			VersionInfo: &k8s.VersionInfo{
				CurrentVersion: "1.33",
			},
		},
	}
}

func TestKubernetesClusterCIDRAttributesConflict(t *testing.T) {
	resourceSchema := resourceYandexKubernetesCluster().Schema

	assert.Equal(t, []string{"cluster_ipv4_ranges"}, resourceSchema["cluster_ipv4_range"].ConflictsWith)
	assert.Equal(t, []string{"cluster_ipv4_range"}, resourceSchema["cluster_ipv4_ranges"].ConflictsWith)
	assert.Equal(t, []string{"cluster_ipv6_ranges"}, resourceSchema["cluster_ipv6_range"].ConflictsWith)
	assert.Equal(t, []string{"cluster_ipv6_range"}, resourceSchema["cluster_ipv6_ranges"].ConflictsWith)
	assert.Contains(t, resourceSchema["cluster_ipv4_range"].Deprecated, "cluster_ipv4_ranges")
	assert.Contains(t, resourceSchema["cluster_ipv6_range"].Deprecated, "cluster_ipv6_ranges")
	assert.Empty(t, resourceSchema["cluster_ipv4_ranges"].Deprecated)
	assert.Empty(t, resourceSchema["cluster_ipv6_ranges"].Deprecated)

	dataSource := dataSourceYandexKubernetesCluster()
	require.NoError(t, dataSource.InternalValidate(nil, false))

	for _, name := range []string{"cluster_ipv4_range", "cluster_ipv4_ranges", "cluster_ipv6_range", "cluster_ipv6_ranges"} {
		assert.Empty(t, dataSource.Schema[name].ConflictsWith)
	}
}

func TestKubernetesClusterCIDRListUpdateFieldsMap(t *testing.T) {
	assert.Equal(t, "ip_allocation_policy.cluster_ipv4_cidr_blocks", updateKubernetesClusterFieldsMap["cluster_ipv4_ranges"])
	assert.Equal(t, "ip_allocation_policy.cluster_ipv6_cidr_blocks", updateKubernetesClusterFieldsMap["cluster_ipv6_ranges"])
	_, hasIPv4Singular := updateKubernetesClusterFieldsMap["cluster_ipv4_range"]
	_, hasIPv6Singular := updateKubernetesClusterFieldsMap["cluster_ipv6_range"]
	assert.False(t, hasIPv4Singular)
	assert.False(t, hasIPv6Singular)
}
