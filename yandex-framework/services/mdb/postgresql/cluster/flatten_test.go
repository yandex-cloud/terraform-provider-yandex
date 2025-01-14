package cluster

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
)

func TestYandexProvider_MDBPostgresClusterConfigAccessFlattener(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	expectedAccessAttrs := map[string]attr.Type{
		"data_lens":     types.BoolType,
		"data_transfer": types.BoolType,
		"serverless":    types.BoolType,
		"web_sql":       types.BoolType,
	}

	cases := []struct {
		testname    string
		reqVal      *postgresql.Access
		expectedVal types.Object
	}{
		{
			testname: "CheckAllAttributes",
			reqVal: &postgresql.Access{
				WebSql:   true,
				DataLens: true,
			},
			expectedVal: types.ObjectValueMust(
				expectedAccessAttrs, map[string]attr.Value{
					"data_lens":     types.BoolValue(true),
					"data_transfer": types.BoolValue(false),
					"serverless":    types.BoolValue(false),
					"web_sql":       types.BoolValue(true),
				},
			),
		},
		{
			testname:    "CheckNullObject",
			reqVal:      nil,
			expectedVal: types.ObjectNull(expectedAccessAttrs),
		},
	}

	for _, c := range cases {
		diags := diag.Diagnostics{}
		access := flattenAccess(ctx, c.reqVal, &diags)
		if diags.HasError() {
			t.Errorf(
				"Unexpected flatten diagnostics status %s test: errors: %v",
				c.testname,
				diags.Errors(),
			)
			continue
		}

		if !c.expectedVal.Equal(access) {
			t.Errorf(
				"Unexpected flatten result value %s test: expected %s, actual %s",
				c.testname,
				c.expectedVal,
				access,
			)
		}
	}
}
