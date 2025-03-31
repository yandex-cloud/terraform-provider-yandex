package mdb_postgresql_cluster_v2

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestYandexProvider_MDBPostgresClusterPGSettingsMapCreate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		testname      string
		reqVal        map[string]attr.Value
		expectedVal   PgSettingsMapValue
		expectedError bool
	}{
		{
			testname: "CheckSeveralAttributes",
			reqVal: map[string]attr.Value{
				"search_path":                   types.StringValue("value1"),
				"max_connections":               types.Int64Value(5),
				"row_security":                  types.BoolValue(true),
				"bgwriter_lru_multiplier":       types.Float64Value(1.1),
				"shared_preload_libraries":      types.TupleValueMust([]attr.Type{types.Int64Type}, []attr.Value{types.Int64Value(1)}),
				"default_transaction_isolation": types.Int64Value(1),
			},
			expectedVal: PgSettingsMapValue{
				MapValue: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"search_path":                   types.StringValue("value1"),
						"max_connections":               types.StringValue("5"),
						"row_security":                  types.StringValue("true"),
						"bgwriter_lru_multiplier":       types.StringValue("1.10"),
						"shared_preload_libraries":      types.StringValue("SHARED_PRELOAD_LIBRARIES_AUTO_EXPLAIN"),
						"default_transaction_isolation": types.StringValue("TRANSACTION_ISOLATION_READ_UNCOMMITTED"),
					},
				),
			},
		},

		{
			testname: "CheckZeroAttributes",
			reqVal: map[string]attr.Value{
				"search_path":                   types.StringValue(""),
				"max_connections":               types.Int64Null(),
				"row_security":                  types.BoolNull(),
				"bgwriter_lru_multiplier":       types.Float64Null(),
				"shared_preload_libraries":      types.TupleNull([]attr.Type{}),
				"default_transaction_isolation": types.Int64Value(0),
			},
			expectedVal: PgSettingsMapValue{
				MapValue: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"search_path":                   types.StringValue(""),
						"max_connections":               types.StringNull(),
						"row_security":                  types.StringNull(),
						"bgwriter_lru_multiplier":       types.StringNull(),
						"shared_preload_libraries":      types.StringNull(),
						"default_transaction_isolation": types.StringValue("TRANSACTION_ISOLATION_UNSPECIFIED"),
					},
				),
			},
		},
	}

	for _, c := range cases {
		conf, diags := NewPgSettingsMapValue(c.reqVal)
		if diags.HasError() != c.expectedError {
			if !c.expectedError {
				t.Errorf(
					"Unexpected create diagnostics status %s test: errors: %v",
					c.testname,
					diags.Errors(),
				)
			} else {
				t.Errorf(
					"Unexpected create diagnostics status %s test: expected error, actual not",
					c.testname,
				)
			}

			continue
		}

		if !c.expectedVal.Equal(conf) {
			t.Errorf(
				"Unexpected create result value %s test: expected %s, actual %s",
				c.testname,
				c.expectedVal.String(),
				conf.String(),
			)
		}
	}
}

func TestYandexProvider_MDBPostgresClusterPGSettingsMapGetPrimitiveMap(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cases := []struct {
		testname      string
		reqVal        PgSettingsMapValue
		expectedVal   map[string]attr.Value
		expectedError bool
	}{
		{
			testname: "CheckBaseAttributes",
			reqVal: PgSettingsMapValue{
				MapValue: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"search_path":              types.StringValue("value1"),
						"max_connections":          types.StringValue("100"),
						"row_security":             types.StringValue("true"),
						"bgwriter_lru_multiplier":  types.StringValue("1.10"),
						"shared_preload_libraries": types.StringValue("SHARED_PRELOAD_LIBRARIES_AUTO_EXPLAIN,SHARED_PRELOAD_LIBRARIES_PG_CRON"),
					},
				),
			},
			expectedVal: map[string]attr.Value{
				"search_path":              types.StringValue("value1"),
				"max_connections":          types.Int64Value(100),
				"row_security":             types.BoolValue(true),
				"bgwriter_lru_multiplier":  types.Float64Value(1.1),
				"shared_preload_libraries": types.TupleValueMust([]attr.Type{types.Int64Type, types.Int64Type}, []attr.Value{types.Int64Value(1), types.Int64Value(5)}),
			},
		},
		{
			testname: "CheckZeroAttributes",
			reqVal: PgSettingsMapValue{
				MapValue: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"search_path":              types.StringValue(""),
						"max_connections":          types.StringNull(),
						"row_security":             types.StringNull(),
						"bgwriter_lru_multiplier":  types.StringNull(),
						"shared_preload_libraries": types.StringNull(),
					},
				),
			},
			expectedVal: map[string]attr.Value{
				"search_path":              types.StringValue(""),
				"max_connections":          types.StringNull(),
				"row_security":             types.StringNull(),
				"bgwriter_lru_multiplier":  types.StringNull(),
				"shared_preload_libraries": types.StringNull(),
			},
		},
		{
			testname: "CheckOneElementSPLAttribute",
			reqVal: PgSettingsMapValue{
				MapValue: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"shared_preload_libraries": types.StringValue("SHARED_PRELOAD_LIBRARIES_AUTO_EXPLAIN"),
					},
				),
			},
			expectedVal: map[string]attr.Value{
				"shared_preload_libraries": types.TupleValueMust([]attr.Type{types.Int64Type}, []attr.Value{types.Int64Value(1)}),
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
