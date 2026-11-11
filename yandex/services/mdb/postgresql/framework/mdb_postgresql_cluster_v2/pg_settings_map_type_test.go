package mdb_postgresql_cluster_v2

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	mdbcommon "github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
)

func TestYandexProvider_PGSettingsMapGetPrimitiveMap(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	settings, _ := mdbcommon.NewSettingsMapValue(
		map[string]attr.Value{
			"max_connections":          types.StringValue("100"),
			"shared_preload_libraries": types.StringValue("SHARED_PRELOAD_LIBRARIES_AUTO_EXPLAIN,SHARED_PRELOAD_LIBRARIES_PG_CRON,SHARED_PRELOAD_LIBRARIES_ANON"),
		},
		pgAttrProvider,
	)

	cases := []struct {
		testname      string
		reqVal        mdbcommon.SettingsMapValue
		expectedVal   map[string]attr.Value
		expectedError bool
	}{
		{
			testname: "CheckBaseAttributes",
			reqVal:   settings,
			expectedVal: map[string]attr.Value{
				"max_connections":          types.Int64Value(100),
				"shared_preload_libraries": types.TupleValueMust([]attr.Type{types.Int64Type, types.Int64Type, types.Int64Type}, []attr.Value{types.Int64Value(1), types.Int64Value(5), types.Int64Value(9)}),
			},
		},
	}

	for _, c := range cases {
		diags := diag.Diagnostics{}

		m := c.reqVal.PrimitiveElements(ctx, &diags)
		if diags.HasError() != c.expectedError {
			if !c.expectedError {
				t.Errorf(
					"Unexpected diagnostics status %s test: errors: %v",
					c.testname,
					diags.Errors(),
				)
			} else {
				t.Errorf(
					"Unexpected diagnostics status %s test: expected error, actual not",
					c.testname,
				)
			}

			continue
		}

		if !reflect.DeepEqual(m, c.expectedVal) {
			t.Errorf(
				"Unexpected result value %s test: expected %v,\n actual %s",
				c.testname,
				c.expectedVal,
				m,
			)
		}
	}
}
