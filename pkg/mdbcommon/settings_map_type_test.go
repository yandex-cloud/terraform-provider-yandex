package mdbcommon

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	config "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1/config"
)

type MockAttrInfoProvider struct{}

var settingsEnumNames = map[string]map[int32]string{
	"default_transaction_isolation":    config.PostgresqlConfig14_TransactionIsolation_name,
	"auto_explain_log_format":          config.PostgresqlConfig14_AutoExplainLogFormat_name,
	"shared_preload_libraries.element": config.PostgresqlConfig14_SharedPreloadLibraries_name,
}

var settingsEnumValues = map[string]map[string]int32{
	"default_transaction_isolation":    config.PostgresqlConfig14_TransactionIsolation_value,
	"auto_explain_log_format":          config.PostgresqlConfig14_AutoExplainLogFormat_value,
	"shared_preload_libraries.element": config.PostgresqlConfig14_SharedPreloadLibraries_value,
}

var listAttributes = map[string]struct{}{
	"shared_preload_libraries": {},
}

func (p *MockAttrInfoProvider) GetSettingsEnumNames() map[string]map[int32]string {
	return settingsEnumNames
}

func (p *MockAttrInfoProvider) GetSettingsEnumValues() map[string]map[string]int32 {
	return settingsEnumValues
}

func (p *MockAttrInfoProvider) GetSetAttributes() map[string]struct{} {
	return listAttributes
}

var mockProvider = MockAttrInfoProvider{}

func TestYandexProvider_MDBSettingsMapCreate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		testname      string
		reqVal        map[string]attr.Value
		expectedVal   SettingsMapValue
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
				"auto_explain_log_format":       types.Int64Value(2),
			},
			expectedVal: SettingsMapValue{
				MapValue: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"search_path":                   types.StringValue("value1"),
						"max_connections":               types.StringValue("5"),
						"row_security":                  types.StringValue("true"),
						"bgwriter_lru_multiplier":       types.StringValue("1.10"),
						"shared_preload_libraries":      types.StringValue("SHARED_PRELOAD_LIBRARIES_AUTO_EXPLAIN"),
						"default_transaction_isolation": types.StringValue("TRANSACTION_ISOLATION_READ_UNCOMMITTED"),
						"auto_explain_log_format":       types.StringValue("AUTO_EXPLAIN_LOG_FORMAT_XML"),
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
			expectedVal: SettingsMapValue{
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
		conf, diags := NewSettingsMapValue(c.reqVal, &mockProvider)
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

func TestYandexProvider_MDBSettingsMapGetPrimitiveMap(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	cases := []struct {
		testname      string
		reqVal        SettingsMapValue
		expectedVal   map[string]attr.Value
		expectedError bool
	}{
		{
			testname: "CheckBaseAttributes",
			reqVal: SettingsMapValue{
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
				p: &mockProvider,
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
			reqVal: SettingsMapValue{
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
				p: &mockProvider,
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
			reqVal: SettingsMapValue{
				MapValue: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"shared_preload_libraries": types.StringValue("SHARED_PRELOAD_LIBRARIES_AUTO_EXPLAIN"),
					},
				),
				p: &mockProvider,
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
