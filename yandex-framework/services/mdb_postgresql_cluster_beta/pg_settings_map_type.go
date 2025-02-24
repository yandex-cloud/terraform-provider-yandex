package mdb_postgresql_cluster_beta

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	config "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1/config"
)

var pgSettingsEnumNames = map[string]map[int32]string{
	"wal_level":                     config.PostgresqlConfig13_WalLevel_name,
	"synchronous_commit":            config.PostgresqlConfig13_SynchronousCommit_name,
	"constraint_exclusion":          config.PostgresqlConfig13_ConstraintExclusion_name,
	"force_parallel_mode":           config.PostgresqlConfig13_ForceParallelMode_name,
	"client_min_messages":           config.PostgresqlConfig13_LogLevel_name,
	"log_min_messages":              config.PostgresqlConfig13_LogLevel_name,
	"log_min_error_statement":       config.PostgresqlConfig13_LogLevel_name,
	"log_error_verbosity":           config.PostgresqlConfig13_LogErrorVerbosity_name,
	"log_statement":                 config.PostgresqlConfig13_LogStatement_name,
	"default_transaction_isolation": config.PostgresqlConfig13_TransactionIsolation_name,
	"bytea_output":                  config.PostgresqlConfig13_ByteaOutput_name,
	"xmlbinary":                     config.PostgresqlConfig13_XmlBinary_name,
	"xmloption":                     config.PostgresqlConfig13_XmlOption_name,
	"backslash_quote":               config.PostgresqlConfig13_BackslashQuote_name,
	"plan_cache_mode":               config.PostgresqlConfig13_PlanCacheMode_name,
	"pg_hint_plan_debug_print":      config.PostgresqlConfig13_PgHintPlanDebugPrint_name,
	"pg_hint_plan_message_level":    config.PostgresqlConfig13_LogLevel_name,
	"shared_preload_libraries":      config.PostgresqlConfig13_SharedPreloadLibraries_name,
	"password_encryption":           config.PostgresqlConfig13_PasswordEncryption_name,
}

var pgSettingsEnumValues = map[string]map[string]int32{
	"wal_level":                        config.PostgresqlConfig13_WalLevel_value,
	"synchronous_commit":               config.PostgresqlConfig13_SynchronousCommit_value,
	"constraint_exclusion":             config.PostgresqlConfig13_ConstraintExclusion_value,
	"force_parallel_mode":              config.PostgresqlConfig13_ForceParallelMode_value,
	"client_min_messages":              config.PostgresqlConfig13_LogLevel_value,
	"log_min_messages":                 config.PostgresqlConfig13_LogLevel_value,
	"log_min_error_statement":          config.PostgresqlConfig13_LogLevel_value,
	"log_error_verbosity":              config.PostgresqlConfig13_LogErrorVerbosity_value,
	"log_statement":                    config.PostgresqlConfig13_LogStatement_value,
	"default_transaction_isolation":    config.PostgresqlConfig13_TransactionIsolation_value,
	"bytea_output":                     config.PostgresqlConfig13_ByteaOutput_value,
	"xmlbinary":                        config.PostgresqlConfig13_XmlBinary_value,
	"xmloption":                        config.PostgresqlConfig13_XmlOption_value,
	"backslash_quote":                  config.PostgresqlConfig13_BackslashQuote_value,
	"plan_cache_mode":                  config.PostgresqlConfig13_PlanCacheMode_value,
	"pg_hint_plan_debug_print":         config.PostgresqlConfig13_PgHintPlanDebugPrint_value,
	"pg_hint_plan_message_level":       config.PostgresqlConfig13_LogLevel_value,
	"shared_preload_libraries.element": config.PostgresqlConfig13_SharedPreloadLibraries_value,
	"password_encryption":              config.PostgresqlConfig13_PasswordEncryption_value,
}

var listAttributes = map[string]struct{}{
	"shared_preload_libraries": {},
}

var _ basetypes.MapTypable = PgSettingsMapType{}

// PgSettingsMapType type is based on the example in the terraform plugin framerwork documentation
// https://developer.hashicorp.com/terraform/plugin/framework/handling-data/types/custom
//
// Type is add-on to the string map for the mapping to primitive fields (Numbers, Bool, String)
// Used to configure Postgresql settings.
type PgSettingsMapType struct {
	basetypes.MapType
}

func (t PgSettingsMapType) String() string {
	return "PgSettingsMapType"
}

// Equal compare PgSettingsMapType with provided type
func (t PgSettingsMapType) Equal(o attr.Type) bool {
	other, ok := o.(PgSettingsMapType)
	if !ok {
		return false
	}

	return t.MapType.Equal(other.MapType)
}

// ValueFromMap used to get PgSettingsMapType from a map value
func (t PgSettingsMapType) ValueFromMap(ctx context.Context, in basetypes.MapValue) (basetypes.MapValuable, diag.Diagnostics) {
	value := PgSettingsMapValue{
		MapValue: in,
	}

	return value, nil
}

// ValueFromTerraform is a basic implementation of getting PgSettingsMapType value from terraform value
//
// From example: https://developer.hashicorp.com/terraform/plugin/framework/handling-data/types/custom
func (t PgSettingsMapType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {

	attrValue, err := t.MapType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	dinVal, ok := attrValue.(basetypes.MapValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	dValuable, diags := t.ValueFromMap(ctx, dinVal)

	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting StringValue to StringValuable: %v", diags)
	}

	return dValuable, nil
}

func (t PgSettingsMapType) ValueType(ctx context.Context) attr.Value {
	return PgSettingsMapValue{}
}

type PgSettingsMapValue struct {
	basetypes.MapValue
}

func (t PgSettingsMapValue) Type(ctx context.Context) attr.Type {
	return PgSettingsMapType{
		MapType: types.MapType{
			ElemType: types.StringType,
		},
	}
}

// Compare map values and elements values inside
func (v PgSettingsMapValue) Equal(o attr.Value) bool {
	other, ok := o.(PgSettingsMapValue)

	if !ok {
		return false
	}

	return v.MapValue.Equal(other.MapValue)
}

// convertFromStringValue is necessary for converting string types to primitives.
func convertFromStringValue(ctx context.Context, a string, v types.String) (attr.Value, diag.Diagnostic) {

	s := v.ValueString()

	if _, ok := listAttributes[a]; ok {
		els := strings.Split(s, ",")
		tupleElems := make([]attr.Value, len(els))
		for idx, elem := range els {
			v, d := convertFromStringValue(ctx, a+".element", types.StringValue(elem))
			if d != nil {
				return nil, d
			}
			tupleElems[idx] = v
		}

		attrTypes := make([]attr.Type, len(tupleElems))
		for idx, elem := range tupleElems {
			attrTypes[idx] = elem.Type(ctx)
		}
		tupleVal, d := types.TupleValue(attrTypes, tupleElems)
		if d.HasError() {
			return nil, d[0]
		}
		return tupleVal, nil
	}

	if nameEnum, ok := pgSettingsEnumValues[a]; ok {
		if num, ok := nameEnum[s]; ok {
			return types.Int64Value(int64(num)), nil
		}

		return types.StringNull(), diag.NewErrorDiagnostic("Enum conversion error", fmt.Sprintf("Attribute %s has a unknown value %v", a, v))
	}

	if v, err := strconv.ParseInt(s, 10, 64); err == nil {
		return types.Int64Value(v), nil
	}

	if v, err := strconv.ParseFloat(s, 64); err == nil {
		return types.Float64Value(v), nil
	}

	if v, err := strconv.ParseBool(s); err == nil {
		return types.BoolValue(v), nil
	}

	return v, nil
}

// PrimitiveElements is necessary to get primitive values map from a PgSettingsMapValue
func (v PgSettingsMapValue) PrimitiveElements(ctx context.Context, diags *diag.Diagnostics) map[string]attr.Value {

	if v.IsNull() || v.IsUnknown() {
		return map[string]attr.Value{}
	}

	if v.ElementType(ctx) != types.StringType {
		diags.AddError("Error to convert string map to primitive map.", "Element type must be string is source map.")
	}

	newMap := make(map[string]attr.Value)
	for attr, val := range v.Elements() {
		if val.IsNull() || val.IsUnknown() {
			newMap[attr] = val
			continue
		}

		v, d := convertFromStringValue(ctx, attr, val.(types.String))
		if d != nil {
			diags.Append(d)
			continue
		}
		newMap[attr] = v
	}

	return newMap
}

func NewPgSettingsMapNull() PgSettingsMapValue {
	return PgSettingsMapValue{MapValue: types.MapNull(types.StringType)}
}

func NewPgSettingsMapUnknown() PgSettingsMapValue {
	return PgSettingsMapValue{MapValue: types.MapUnknown(types.StringType)}
}

// convertToStringValue converts attr.Value to a string
//
// attr.Value can be enum that converted to a enum value string
func convertToStringValue(ctx context.Context, attr string, val attr.Value) (types.String, diag.Diagnostic) {
	if val.IsNull() {
		return types.StringNull(), nil
	}

	if val.IsUnknown() {
		return types.StringUnknown(), nil
	}

	if valTuple, ok := val.(types.Tuple); ok {
		els := valTuple.Elements()
		strTuple := make([]string, len(els))
		for i, el := range els {
			var d diag.Diagnostic
			s, d := convertToStringValue(ctx, attr, el)
			strTuple[i] = s.ValueString()
			if d != nil {
				return types.StringNull(), d
			}
		}

		return types.StringValue(strings.Join(strTuple, ",")), nil
	}

	if enumValues, ok := pgSettingsEnumNames[attr]; ok {
		if valInt, ok := val.(types.Int64); ok {
			return types.StringValue(enumValues[int32(valInt.ValueInt64())]), nil
		}

		if valNum, ok := val.(types.Number); ok {
			if !valNum.ValueBigFloat().IsInt() {
				return types.StringNull(), diag.NewErrorDiagnostic("Error conversion enum", "Enum for attribute must be a integer")
			}
			i, _ := valNum.ValueBigFloat().Int64()
			return types.StringValue(enumValues[int32(i)]), nil

		}

		return types.StringNull(), diag.NewErrorDiagnostic("Error conversion enum", "Enum for attribute must be a integer")
	}

	if valInt, ok := val.(types.Int64); ok {
		return types.StringValue(strconv.Itoa(int(valInt.ValueInt64()))), nil
	}

	if valFloat, ok := val.(types.Float64); ok {
		return types.StringValue(
			strconv.FormatFloat(
				valFloat.ValueFloat64(), 'f', 2, 64,
			),
		), nil
	}

	if valNum, ok := val.(types.Number); ok {
		if valNum.ValueBigFloat().IsInt() {
			i, _ := valNum.ValueBigFloat().Int64()
			return types.StringValue(strconv.Itoa(int(i))), nil
		}

		f, _ := valNum.ValueBigFloat().Float64()
		return types.StringValue(strconv.FormatFloat(
			f, 'f', 2, 64,
		)), nil

	}

	if valBool, ok := val.(types.Bool); ok {
		return types.StringValue(strconv.FormatBool(valBool.ValueBool())), nil

	}

	if valString, ok := val.(types.String); ok {
		return valString, nil
	}

	return types.StringNull(), diag.NewErrorDiagnostic("Conversion error", fmt.Sprintf("Cannot convert value %s to string", val.String()))
}

// NewPgSettingsMapValue creates PgSettingsMapValue from map[string]attr.Value, where attr.Value is a primitive value
func NewPgSettingsMapValue(elements map[string]attr.Value) (PgSettingsMapValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	strMap := make(map[string]attr.Value)
	for attr, val := range elements {
		s, d := convertToStringValue(context.Background(), attr, val)
		strMap[attr] = s
		if d != nil {
			diags.Append(d)
		}
	}

	mv, d := types.MapValue(types.StringType, strMap)
	diags.Append(d...)
	return PgSettingsMapValue{MapValue: mv}, diags
}

func NewPgSettingsMapValueMust(elements map[string]attr.Value) PgSettingsMapValue {
	v, d := NewPgSettingsMapValue(elements)
	if d.HasError() {
		panic(d)
	}
	return v
}
