package cdn_resource

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/cdn/v1"
)

func TestExpandOriginProtocol(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		testname      string
		protocolValue string
		expectedProto cdn.OriginProtocol
		expectedError bool
	}{
		{
			testname:      "HTTP protocol",
			protocolValue: "http",
			expectedProto: cdn.OriginProtocol_HTTP,
			expectedError: false,
		},
		{
			testname:      "HTTPS protocol",
			protocolValue: "https",
			expectedProto: cdn.OriginProtocol_HTTPS,
			expectedError: false,
		},
		{
			testname:      "MATCH protocol",
			protocolValue: "match",
			expectedProto: cdn.OriginProtocol_MATCH,
			expectedError: false,
		},
		{
			testname:      "Invalid protocol",
			protocolValue: "htps",
			expectedProto: cdn.OriginProtocol_ORIGIN_PROTOCOL_UNSPECIFIED,
			expectedError: true,
		},
		{
			testname:      "Empty protocol",
			protocolValue: "",
			expectedProto: cdn.OriginProtocol_ORIGIN_PROTOCOL_UNSPECIFIED,
			expectedError: true,
		},
		{
			testname:      "Unknown protocol",
			protocolValue: "ftp",
			expectedProto: cdn.OriginProtocol_ORIGIN_PROTOCOL_UNSPECIFIED,
			expectedError: true,
		},
	}

	for _, c := range cases {
		t.Run(c.testname, func(t *testing.T) {
			diags := diag.Diagnostics{}
			result := expandOriginProtocol(ctx, c.protocolValue, &diags)

			if diags.HasError() != c.expectedError {
				t.Errorf(
					"Unexpected error status for %s: expected hasError=%v, got hasError=%v, diagnostics=%v",
					c.testname,
					c.expectedError,
					diags.HasError(),
					diags.Errors(),
				)
			}

			if !c.expectedError && result != c.expectedProto {
				t.Errorf(
					"Unexpected protocol for %s: expected %v, got %v",
					c.testname,
					c.expectedProto,
					result,
				)
			}
		})
	}
}
