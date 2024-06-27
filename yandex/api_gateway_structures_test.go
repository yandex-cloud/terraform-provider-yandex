package yandex

import (
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/stretchr/testify/require"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/apigateway/v1"
	"google.golang.org/protobuf/types/known/durationpb"
)

const (
	executionTimeoutKey = "execution_timeout"
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

func TestExpandApiGatewayVariables(t *testing.T) {
	tests := []struct {
		name     string
		raw      map[string]interface{}
		expected map[string]*apigateway.VariableInput
	}{
		{
			name:     "nil",
			raw:      nil,
			expected: nil,
		},
		{
			name: "empty",
			raw: map[string]interface{}{
				"variables": map[string]interface{}{},
			},
			expected: nil,
		},
		{
			name: "filled variables",
			raw: map[string]interface{}{
				"variables": map[string]interface{}{
					"variable1": "foo",
					"variable2": "1",
					"variable3": "2.3",
					"variable4": "true",
				},
			},
			expected: map[string]*apigateway.VariableInput{
				"variable1": {VariableValue: &apigateway.VariableInput_StringValue{StringValue: "foo"}},
				"variable2": {VariableValue: &apigateway.VariableInput_IntValue{IntValue: 1}},
				"variable3": {VariableValue: &apigateway.VariableInput_DoubleValue{DoubleValue: 2.3}},
				"variable4": {VariableValue: &apigateway.VariableInput_BoolValue{BoolValue: true}},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resourceData := schema.TestResourceDataRaw(t, resourceYandexApiGateway().Schema, test.raw)
			actual := expandApiGatewayVariables(resourceData)
			require.Equal(t, test.expected, actual)
		})
	}
}

func TestFlattenApiGatewayVariables(t *testing.T) {
	tests := []struct {
		name      string
		variables map[string]*apigateway.VariableInput
		expected  map[string]interface{}
	}{
		{
			name:      "nil",
			variables: nil,
			expected:  nil,
		},
		{
			name:      "empty variables",
			variables: map[string]*apigateway.VariableInput{},
			expected:  nil,
		},
		{
			name: "filled variables",
			variables: map[string]*apigateway.VariableInput{
				"variable1": {VariableValue: &apigateway.VariableInput_StringValue{StringValue: "foo"}},
				"variable2": {VariableValue: &apigateway.VariableInput_IntValue{IntValue: 1}},
				"variable3": {VariableValue: &apigateway.VariableInput_DoubleValue{DoubleValue: 2.3}},
				"variable4": {VariableValue: &apigateway.VariableInput_BoolValue{BoolValue: true}},
			},
			expected: map[string]interface{}{
				"variable1": "foo",
				"variable2": "1",
				"variable3": "2.3",
				"variable4": "true",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := flattenApiGatewayVariables(test.variables)
			require.Equal(t, test.expected, actual)
		})
	}
}

func TestExpandApiGatewayCanary(t *testing.T) {
	tests := []struct {
		name     string
		raw      map[string]interface{}
		expected *apigateway.Canary
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
			name: "filled canary",
			raw: map[string]interface{}{
				"canary": []interface{}{map[string]interface{}{
					"weight": 2,
					"variables": map[string]interface{}{
						"variable1": "foo",
						"variable2": "1",
						"variable3": "2.3",
						"variable4": "true",
					},
				}},
			},
			expected: &apigateway.Canary{
				Weight: 2,
				Variables: map[string]*apigateway.VariableInput{
					"variable1": {VariableValue: &apigateway.VariableInput_StringValue{StringValue: "foo"}},
					"variable2": {VariableValue: &apigateway.VariableInput_IntValue{IntValue: 1}},
					"variable3": {VariableValue: &apigateway.VariableInput_DoubleValue{DoubleValue: 2.3}},
					"variable4": {VariableValue: &apigateway.VariableInput_BoolValue{BoolValue: true}},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resourceData := schema.TestResourceDataRaw(t, resourceYandexApiGateway().Schema, test.raw)
			actual := expandApiGatewayCanary(resourceData)
			require.Equal(t, test.expected, actual)
		})
	}
}

func TestFlattenApiGatewayCanary(t *testing.T) {
	tests := []struct {
		name     string
		spec     *apigateway.Canary
		expected []interface{}
	}{
		{
			name:     "nil",
			spec:     nil,
			expected: nil,
		},
		{
			name:     "empty canary",
			spec:     &apigateway.Canary{},
			expected: nil,
		},
		{
			name: "filled canary",
			spec: &apigateway.Canary{
				Weight: 2,
				Variables: map[string]*apigateway.VariableInput{
					"variable1": {VariableValue: &apigateway.VariableInput_StringValue{StringValue: "foo"}},
					"variable2": {VariableValue: &apigateway.VariableInput_IntValue{IntValue: 1}},
					"variable3": {VariableValue: &apigateway.VariableInput_DoubleValue{DoubleValue: 2.3}},
					"variable4": {VariableValue: &apigateway.VariableInput_BoolValue{BoolValue: true}},
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"weight": 2,
					"variables": map[string]interface{}{
						"variable1": "foo",
						"variable2": "1",
						"variable3": "2.3",
						"variable4": "true",
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := flattenApiGatewayCanary(test.spec)
			if test.expected != nil {
				require.Equal(t, len(test.expected), len(actual))
				require.Equal(t, (test.expected[0].(map[string]interface{}))["variables"], (actual[0].(map[string]interface{}))["variables"])
				return
			}
			require.Equal(t, test.expected, actual)
		})
	}
}

func TestExpandApiGatewayExecutionTimeout(t *testing.T) {
	for _, test := range []struct {
		name          string
		raw           map[string]interface{}
		expected      *durationpb.Duration
		expectedError error
	}{
		{
			name:          "nil",
			raw:           nil,
			expected:      nil,
			expectedError: nil,
		},
		{
			name:          "empty",
			raw:           map[string]interface{}{},
			expected:      nil,
			expectedError: nil,
		},
		{
			name: "valid",
			raw: map[string]interface{}{
				executionTimeoutKey: "200",
			},
			expected:      durationpb.New(200 * time.Second),
			expectedError: nil,
		},
		{
			name: "invalid",
			raw: map[string]interface{}{
				executionTimeoutKey: "invalid duration",
			},
			expected: nil,
			expectedError: &strconv.NumError{
				Func: "ParseInt",
				Num:  "invalid duration",
				Err:  errors.New("invalid syntax"),
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			resourceData := schema.TestResourceDataRaw(t, resourceYandexApiGateway().Schema, test.raw)
			actual, err := expandApiGatewayExecutionTimeout(resourceData)
			require.Equal(t, test.expected, actual)
			require.Equal(t, test.expectedError, err)
		})
	}
}

func TestFlattenApiGatewayExecutionTimeout(t *testing.T) {
	for _, test := range []struct {
		name     string
		duration *durationpb.Duration
		expected string
	}{
		{
			name:     "valid",
			duration: durationpb.New(100 * time.Second),
			expected: "100",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.expected, strconv.FormatInt(test.duration.Seconds, 10))
		})
	}
}
