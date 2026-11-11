package pluginsdk_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/stretchr/testify/require"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/containers/v1"
	svc_serverless_containers "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/serverless/containers/pluginsdk"
)

func TestExpandServerlessContainerConnectivity(t *testing.T) {
	networkId := acctest.RandomWithPrefix("tf-serverless-container-connectivity")
	tests := []struct {
		name     string
		raw      map[string]interface{}
		expected *containers.Connectivity
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
			expected: &containers.Connectivity{
				NetworkId: networkId,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resourceData := schema.TestResourceDataRaw(t, svc_serverless_containers.ResourceYandexServerlessContainer().Schema, test.raw)
			actual := svc_serverless_containers.ExpandServerlessContainerConnectivity(resourceData)
			require.Equal(t, test.expected, actual)
		})
	}
}

func TestFlattenServerlessContainerConnectivity(t *testing.T) {
	networkId := acctest.RandomWithPrefix("tf-serverless-container-connectivity")
	tests := []struct {
		name     string
		spec     *containers.Connectivity
		expected []interface{}
	}{
		{
			name:     "nil",
			spec:     nil,
			expected: nil,
		},
		{
			name:     "empty connectivity",
			spec:     &containers.Connectivity{},
			expected: nil,
		},
		{
			name:     "empty network_id",
			spec:     &containers.Connectivity{NetworkId: ""},
			expected: nil,
		},
		{
			name:     "filled connectivity",
			spec:     &containers.Connectivity{NetworkId: networkId},
			expected: []interface{}{map[string]interface{}{"network_id": networkId}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := svc_serverless_containers.FlattenServerlessContainerConnectivity(test.spec)
			require.Equal(t, test.expected, actual)
		})
	}
}
