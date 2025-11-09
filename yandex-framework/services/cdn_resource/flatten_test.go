package cdn_resource

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/cdn/v1"
)

func TestFlattenOriginProtocol(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		testname       string
		apiProtocol    cdn.OriginProtocol
		expectedValue  string
		expectedIsNull bool
		expectedError  bool
	}{
		{
			testname:       "HTTP protocol",
			apiProtocol:    cdn.OriginProtocol_HTTP,
			expectedValue:  "http",
			expectedIsNull: false,
			expectedError:  false,
		},
		{
			testname:       "HTTPS protocol",
			apiProtocol:    cdn.OriginProtocol_HTTPS,
			expectedValue:  "https",
			expectedIsNull: false,
			expectedError:  false,
		},
		{
			testname:       "MATCH protocol",
			apiProtocol:    cdn.OriginProtocol_MATCH,
			expectedValue:  "match",
			expectedIsNull: false,
			expectedError:  false,
		},
		{
			testname:       "UNSPECIFIED protocol",
			apiProtocol:    cdn.OriginProtocol_ORIGIN_PROTOCOL_UNSPECIFIED,
			expectedValue:  "",
			expectedIsNull: true,
			expectedError:  true,
		},
		{
			testname:       "Invalid protocol value 999",
			apiProtocol:    cdn.OriginProtocol(999),
			expectedValue:  "",
			expectedIsNull: true,
			expectedError:  true,
		},
	}

	for _, c := range cases {
		t.Run(c.testname, func(t *testing.T) {
			diags := diag.Diagnostics{}
			result := flattenOriginProtocol(ctx, c.apiProtocol, &diags)

			if diags.HasError() != c.expectedError {
				t.Errorf(
					"Unexpected error status for %s: expected hasError=%v, got hasError=%v, diagnostics=%v",
					c.testname,
					c.expectedError,
					diags.HasError(),
					diags.Errors(),
				)
			}

			if result.IsNull() != c.expectedIsNull {
				t.Errorf(
					"Unexpected null status for %s: expected isNull=%v, got isNull=%v",
					c.testname,
					c.expectedIsNull,
					result.IsNull(),
				)
			}

			if !c.expectedIsNull && result.ValueString() != c.expectedValue {
				t.Errorf(
					"Unexpected value for %s: expected %s, got %s",
					c.testname,
					c.expectedValue,
					result.ValueString(),
				)
			}
		})
	}
}
