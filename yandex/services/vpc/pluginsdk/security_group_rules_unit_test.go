package pluginsdk_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
	svc_vpc "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/vpc/pluginsdk"
)

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
			expected: schema.NewSet(svc_vpc.ResourceYandexVPCSecurityGroupRuleHash, []interface{}{
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
			_, egress := svc_vpc.FlattenSecurityGroupRulesSpec(tt.spec)

			if egress.Difference(tt.expected).Len() > 0 {
				t.Errorf("flattenInstances() got = %v, want %v", egress.List(), tt.expected.List())
			}
		})
	}
}
