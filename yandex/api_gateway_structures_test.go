package yandex

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/apigateway/v1"
	"testing"
)

func TestExpandApiGatewayConnectivity(t *testing.T) {
	networkId := acctest.RandomWithPrefix("tf-api-gateway-connectivity")
	tests := []struct {
		name     string
		raw      map[string]interface{}
		expected *apigateway.Connectivity
	}{
		{
			name:     "nil",
			raw:      nil,
			expected: nil,
		},
		{
			name:     "empty",
			raw:      map[string]interface{}{},
			expected: nil,
		},
		{
			name: "filled connectivity",
			raw: map[string]interface{}{
				"connectivity": []interface{}{map[string]interface{}{
					"network_id": networkId,
				}},
			},
			expected: &apigateway.Connectivity{
				NetworkId: networkId,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resourceData := schema.TestResourceDataRaw(t, resourceYandexApiGateway().Schema, test.raw)
			actual := expandApiGatewayConnectivity(resourceData)
			require.Equal(t, test.expected, actual)
		})
	}
}

func TestFlattenApiGatewayConnectivity(t *testing.T) {
	networkId := acctest.RandomWithPrefix("tf-api-gateway-connectivity")
	tests := []struct {
		name     string
		spec     *apigateway.Connectivity
		expected []interface{}
	}{
		{
			name:     "nil",
			spec:     nil,
			expected: nil,
		},
		{
			name:     "empty connectivity",
			spec:     &apigateway.Connectivity{},
			expected: nil,
		},
		{
			name:     "empty network_id",
			spec:     &apigateway.Connectivity{NetworkId: ""},
			expected: nil,
		},
		{
			name:     "filled connectivity",
			spec:     &apigateway.Connectivity{NetworkId: networkId},
			expected: []interface{}{map[string]interface{}{"network_id": networkId}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := flattenApiGatewayConnectivity(test.spec)
			require.Equal(t, test.expected, actual)
		})
	}
}
