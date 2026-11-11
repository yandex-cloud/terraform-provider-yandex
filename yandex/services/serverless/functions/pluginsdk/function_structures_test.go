package pluginsdk_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/stretchr/testify/require"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/functions/v1"
	svc_cloud_functions "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/serverless/functions/pluginsdk"
)

func TestExpandFunctionConnectivity(t *testing.T) {
	networkId := acctest.RandomWithPrefix("tf-function-connectivity")
	tests := []struct {
		name     string
		raw      map[string]interface{}
		expected *functions.Connectivity
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
			expected: &functions.Connectivity{
				NetworkId: networkId,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resourceData := schema.TestResourceDataRaw(t, svc_cloud_functions.ResourceYandexFunction().Schema, test.raw)
			actual := svc_cloud_functions.ExpandFunctionConnectivity(resourceData)
			require.Equal(t, test.expected, actual)
		})
	}
}

func TestFlattenFunctionConnectivity(t *testing.T) {
	networkId := acctest.RandomWithPrefix("tf-function-connectivity")
	tests := []struct {
		name     string
		spec     *functions.Connectivity
		expected []interface{}
	}{
		{
			name:     "nil",
			spec:     nil,
			expected: nil,
		},
		{
			name:     "empty connectivity",
			spec:     &functions.Connectivity{},
			expected: nil,
		},
		{
			name:     "empty network_id",
			spec:     &functions.Connectivity{NetworkId: ""},
			expected: nil,
		},
		{
			name:     "filled connectivity",
			spec:     &functions.Connectivity{NetworkId: networkId},
			expected: []interface{}{map[string]interface{}{"network_id": networkId}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := svc_cloud_functions.FlattenFunctionConnectivity(test.spec)
			require.Equal(t, test.expected, actual)
		})
	}
}
