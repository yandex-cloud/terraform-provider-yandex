package yandex

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/apploadbalancer/v1"
)

func Test_flattenALBRateLimit(t *testing.T) {
	t.Parallel()

	testsTable := []struct {
		name           string
		rateLimit      *apploadbalancer.RateLimit
		expectedResult []map[string]interface{}
	}{
		{
			name: "nil rate limit",
		},
		{
			name:           "empty rate limit",
			rateLimit:      &apploadbalancer.RateLimit{},
			expectedResult: []map[string]interface{}{{}},
		},
		{
			name: "empty all requests limit",
			rateLimit: &apploadbalancer.RateLimit{
				AllRequests: &apploadbalancer.RateLimit_Limit{},
			},
			expectedResult: []map[string]interface{}{
				{
					allRequestsSchemaKey: []map[string]interface{}{{}},
				},
			},
		},
		{
			name: "all requests rps",
			rateLimit: &apploadbalancer.RateLimit{
				AllRequests: &apploadbalancer.RateLimit_Limit{
					Rate: &apploadbalancer.RateLimit_Limit_PerSecond{
						PerSecond: 15,
					},
				},
			},
			expectedResult: []map[string]interface{}{
				{
					allRequestsSchemaKey: []map[string]interface{}{
						{
							perSecondSchemaKey: 15,
						},
					},
				},
			},
		},
		{
			name: "all requests rpm",
			rateLimit: &apploadbalancer.RateLimit{
				AllRequests: &apploadbalancer.RateLimit_Limit{
					Rate: &apploadbalancer.RateLimit_Limit_PerMinute{
						PerMinute: 15,
					},
				},
			},
			expectedResult: []map[string]interface{}{
				{
					allRequestsSchemaKey: []map[string]interface{}{
						{
							perMinuteSchemaKey: 15,
						},
					},
				},
			},
		},
		{
			name: "empty requests per ip limit",
			rateLimit: &apploadbalancer.RateLimit{
				RequestsPerIp: &apploadbalancer.RateLimit_Limit{},
			},
			expectedResult: []map[string]interface{}{
				{
					requestsPerIPSchemaKey: []map[string]interface{}{{}},
				},
			},
		},
		{
			name: "requests per ip rps",
			rateLimit: &apploadbalancer.RateLimit{
				RequestsPerIp: &apploadbalancer.RateLimit_Limit{
					Rate: &apploadbalancer.RateLimit_Limit_PerSecond{
						PerSecond: 15,
					},
				},
			},
			expectedResult: []map[string]interface{}{
				{
					requestsPerIPSchemaKey: []map[string]interface{}{
						{
							perSecondSchemaKey: 15,
						},
					},
				},
			},
		},
		{
			name: "all requests rpm",
			rateLimit: &apploadbalancer.RateLimit{
				RequestsPerIp: &apploadbalancer.RateLimit_Limit{
					Rate: &apploadbalancer.RateLimit_Limit_PerMinute{
						PerMinute: 15,
					},
				},
			},
			expectedResult: []map[string]interface{}{
				{
					requestsPerIPSchemaKey: []map[string]interface{}{
						{
							perMinuteSchemaKey: 15,
						},
					},
				},
			},
		},
		{
			name: "all requests and requests per ip limits",
			rateLimit: &apploadbalancer.RateLimit{
				AllRequests: &apploadbalancer.RateLimit_Limit{
					Rate: &apploadbalancer.RateLimit_Limit_PerSecond{
						PerSecond: 10,
					},
				},
				RequestsPerIp: &apploadbalancer.RateLimit_Limit{
					Rate: &apploadbalancer.RateLimit_Limit_PerMinute{
						PerMinute: 15,
					},
				},
			},
			expectedResult: []map[string]interface{}{
				{
					allRequestsSchemaKey: []map[string]interface{}{
						{
							perSecondSchemaKey: 10,
						},
					},
					requestsPerIPSchemaKey: []map[string]interface{}{
						{
							perMinuteSchemaKey: 15,
						},
					},
				},
			},
		},
	}

	for _, testCase := range testsTable {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			actualResult := flattenALBRateLimit(testCase.rateLimit)

			assert.Equal(t, testCase.expectedResult, actualResult)
		})
	}
}
