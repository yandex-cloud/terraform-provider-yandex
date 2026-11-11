package pluginsdk_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"
	svc_network_load_balancer "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/loadbalancer/pluginsdk"
)

func TestExpandLBListenerSpecValidation(t *testing.T) {
	cases := []struct {
		name          string
		listener      map[string]interface{}
		expectedError string
	}{
		{
			name: "two specs",
			listener: map[string]interface{}{
				"external_address_spec": schema.NewSet(svc_network_load_balancer.ResourceLBNetworkLoadBalancerExternalAddressHash, []interface{}{
					map[string]interface{}{"address": "10.0.0.1"},
				}),
				"internal_address_spec": schema.NewSet(svc_network_load_balancer.ResourceLBNetworkLoadBalancerInternalAddressHash, []interface{}{
					map[string]interface{}{"subnet_id": "subnet"},
				}),
			},
			expectedError: "use one of 'external_address_spec' or 'internal_address_spec', not both",
		},
		{
			name: "only external_address_spec",
			listener: map[string]interface{}{
				"external_address_spec": schema.NewSet(svc_network_load_balancer.ResourceLBNetworkLoadBalancerExternalAddressHash, []interface{}{
					map[string]interface{}{"address": "10.0.0.1"},
				}),
			},
			expectedError: "",
		},
		{
			name: "only internal_address_spec",
			listener: map[string]interface{}{
				"internal_address_spec": schema.NewSet(svc_network_load_balancer.ResourceLBNetworkLoadBalancerInternalAddressHash, []interface{}{
					map[string]interface{}{"subnet_id": "subnet"},
				}),
			},
			expectedError: "",
		},
		{
			name:          "without specs",
			listener:      map[string]interface{}{},
			expectedError: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc_network_load_balancer.ExpandLBListenerSpec(tc.listener)
			if tc.expectedError == "" {
				require.Nil(t, err)
			} else {
				require.NotNil(t, err)
				require.Equal(t, err.Error(), tc.expectedError)
			}
		})
	}
}
