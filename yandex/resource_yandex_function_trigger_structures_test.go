package yandex

import (
	"testing"

	"github.com/golang/protobuf/ptypes/duration"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"
	triggers "github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/triggers/v1"
)

func TestExpandBatchSettings(t *testing.T) {
	tests := []struct {
		name     string
		raw      map[string]interface{}
		prefix   string
		expected *triggers.BatchSettings
		wantErr  bool
	}{
		{
			name:     "block absent",
			raw:      map[string]interface{}{},
			prefix:   "iot.0",
			expected: nil,
		},
		{

			name: "only batch_cutoff set",
			raw: map[string]interface{}{
				"iot": []interface{}{map[string]interface{}{
					"batch_cutoff": "30",
				}},
			},
			prefix: "iot.0",
			expected: &triggers.BatchSettings{
				Cutoff: &duration.Duration{Seconds: 30},
			},
		},
		{
			name: "both batch_cutoff and batch_size set",
			raw: map[string]interface{}{
				"iot": []interface{}{map[string]interface{}{
					"batch_cutoff": "20",
					"batch_size":   "5",
				}},
			},
			prefix: "iot.0",
			expected: &triggers.BatchSettings{
				Cutoff: &duration.Duration{Seconds: 20},
				Size:   5,
			},
		},
		{
			name: "invalid batch_cutoff",
			raw: map[string]interface{}{
				"iot": []interface{}{map[string]interface{}{
					"batch_cutoff": "not-a-number",
				}},
			},
			prefix:  "iot.0",
			wantErr: true,
		},
		{
			name: "invalid batch_size",
			raw: map[string]interface{}{
				"iot": []interface{}{map[string]interface{}{
					"batch_cutoff": "10",
					"batch_size":   "not-a-number",
				}},
			},
			prefix:  "iot.0",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, resourceYandexFunctionTrigger().Schema, tc.raw)
			actual, err := expandBatchSettings(d, tc.prefix)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expected, actual)
		})
	}
}
