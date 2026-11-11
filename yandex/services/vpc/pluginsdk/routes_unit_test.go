package pluginsdk_test

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
	svc_vpc "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/vpc/pluginsdk"
)

func TestExpandStaticRoutes(t *testing.T) {
	cases := []struct {
		name       string
		v          interface{}
		expected   []*vpc.StaticRoute
		shouldFail bool
	}{
		{
			name: "two routes",
			v: schema.NewSet(svc_vpc.ResourceYandexVPCRouteTableHash, []interface{}{
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
			v: schema.NewSet(svc_vpc.ResourceYandexVPCRouteTableHash, []interface{}{
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
			v:          schema.NewSet(svc_vpc.ResourceYandexVPCRouteTableHash, []interface{}{map[string]interface{}{}}),
			shouldFail: true,
		},
		{
			name: "no next hop",
			v: schema.NewSet(svc_vpc.ResourceYandexVPCRouteTableHash, []interface{}{
				map[string]interface{}{
					"destination_prefix": "192.0.2.0/24",
				},
			},
			),
			shouldFail: true,
		},
		{
			name: "too many next hops",
			v: schema.NewSet(svc_vpc.ResourceYandexVPCRouteTableHash, []interface{}{
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
			result, err := svc_vpc.ExpandStaticRoutes(tc.v)
			if err != nil && !tc.shouldFail {
				t.Fatalf("bad: %#v", err)
			}
			if !reflect.DeepEqual(result, tc.expected) {
				t.Fatalf("Got:\n\n%#v\n\nExpected:\n\n%#v\n", result, tc.expected)
			}
		})
	}
}
