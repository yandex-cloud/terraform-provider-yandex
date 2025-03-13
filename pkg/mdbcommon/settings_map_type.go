package mdbcommon

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
)

type SettingsAttributeInfoProvider interface {
	GetSettingsEnumNames() map[string]map[int32]string
	GetSettingsEnumValues() map[string]map[string]int32
	GetSetAttributes() map[string]struct{}
}

var _ basetypes.MapTypable = SettingsMapType{}

// SettingsMapType type is based on the example in the terraform plugin framerwork documentation
// https://developer.hashicorp.com/terraform/plugin/framework/handling-data/types/custom
//
// Type is add-on to the string map for the mapping to primitive fields (Numbers, Bool, String)
// Used to configure Postgresql settings.
type SettingsMapType struct {
	p SettingsAttributeInfoProvider
	basetypes.MapType
}

func (t SettingsMapType) String() string {
	return "SettingsMapType"
}

// Equal compare MsSettingsMapType with provided type
func (t SettingsMapType) Equal(o attr.Type) bool {
	other, ok := o.(SettingsMapType)
	if !ok {
		return false
	}

	return t.MapType.Equal(other.MapType)
}

// ValueFromMap used to get MsSettingsMapType from a map value
func (t SettingsMapType) ValueFromMap(ctx context.Context, in basetypes.MapValue) (basetypes.MapValuable, diag.Diagnostics) {
	value := SettingsMapValue{
		MapValue: in,
		p:        t.p,
	}

	return value, nil
}

// ValueFromTerraform is a basic implementation of getting MsSettingsMapType value from terraform value
//
// From example: https://developer.hashicorp.com/terraform/plugin/framework/handling-data/types/custom
func (t SettingsMapType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {

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
		return nil, fmt.Errorf("unexpected error converting MapValue to MapValuable: %v", diags)
	}

	return dValuable, nil
}

func (t SettingsMapType) ValueType(ctx context.Context) attr.Value {
	return SettingsMapValue{
		p:        t.p,
		MapValue: types.MapNull(types.StringType),
	}
}

type SettingsMapValue struct {
	basetypes.MapValue
	p SettingsAttributeInfoProvider
}

func (t SettingsMapValue) Type(ctx context.Context) attr.Type {
	return SettingsMapType{
		MapType: types.MapType{
			ElemType: types.StringType,
		},
		p: t.p,
	}
}

// Compare map values and elements values inside
func (v SettingsMapValue) Equal(o attr.Value) bool {
	other, ok := o.(SettingsMapValue)

	if !ok {
		return false
	}

	return v.MapValue.Equal(other.MapValue)
}

// convertFromStringValue is necessary for converting string types to primitives.
func (v SettingsMapValue) convertFromStringValue(ctx context.Context, a string, val types.String) (attr.Value, diag.Diagnostic) {

	s := val.ValueString()

	if _, ok := v.p.GetSetAttributes()[a]; ok {
		els := strings.Split(s, ",")
		tupleElems := make([]attr.Value, len(els))
		for idx, elem := range els {
			attrVal, d := v.convertFromStringValue(ctx, a+".element", types.StringValue(elem))
			if d != nil {
				return nil, d
			}
			tupleElems[idx] = attrVal
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

	if nameEnum, ok := v.p.GetSettingsEnumValues()[a]; ok {
		if num, ok := nameEnum[s]; ok {
			return types.Int64Value(int64(num)), nil
		}

		return types.StringNull(), diag.NewErrorDiagnostic("Enum conversion error", fmt.Sprintf("Attribute %s has a unknown value %v", a, val))
	}

	if attrVal, err := strconv.ParseInt(s, 10, 64); err == nil {
		return types.Int64Value(attrVal), nil
	}

	if attrVal, err := strconv.ParseFloat(s, 64); err == nil {
		return types.Float64Value(attrVal), nil
	}

	if attrVal, err := strconv.ParseBool(s); err == nil {
		return types.BoolValue(attrVal), nil
	}

	return val, nil
}

// PrimitiveElements is necessary to get primitive values map from a MsSettingsMapValue
func (v SettingsMapValue) PrimitiveElements(ctx context.Context, diags *diag.Diagnostics) map[string]attr.Value {

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

		attrVal, d := v.convertFromStringValue(ctx, attr, val.(types.String))
		if d != nil {
			diags.Append(d)
			continue
		}
		newMap[attr] = attrVal
	}

	return newMap
}

func NewSettingsMapNull() SettingsMapValue {
	return SettingsMapValue{MapValue: types.MapNull(types.StringType)}
}

func NewSettingsMapUnknown() SettingsMapValue {
	return SettingsMapValue{MapValue: types.MapUnknown(types.StringType)}
}

// convertToStringValue converts attr.Value to a string
//
// attr.Value can be enum that converted to a enum value string
func (v SettingsMapValue) convertToStringValue(ctx context.Context, attr string, val attr.Value) (types.String, diag.Diagnostic) {
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
			s, d := v.convertToStringValue(ctx, attr+".element", el)
			strTuple[i] = s.ValueString()
			if d != nil {
				return types.StringNull(), d
			}
		}

		return types.StringValue(strings.Join(strTuple, ",")), nil
	}

	if enumValues, ok := v.p.GetSettingsEnumNames()[attr]; ok {
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

// NewMsSettingsMapValue creates MsSettingsMapValue from map[string]attr.Value, where attr.Value is a primitive value
func NewSettingsMapValue(elements map[string]attr.Value, p SettingsAttributeInfoProvider) (SettingsMapValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	v := SettingsMapValue{
		p: p,
	}

	strMap := make(map[string]attr.Value)
	for attr, val := range elements {
		s, d := v.convertToStringValue(context.Background(), attr, val)
		strMap[attr] = s
		if d != nil {
			diags.Append(d)
		}
	}

	mv, d := types.MapValue(types.StringType, strMap)
	diags.Append(d...)
	v.MapValue = mv

	return v, diags
}

func NewSettingsMapValueMust(elements map[string]attr.Value, p SettingsAttributeInfoProvider) SettingsMapValue {
	v, d := NewSettingsMapValue(elements, p)
	if d.HasError() {
		panic(d)
	}
	return v
}

func NewSettingsMapType(p SettingsAttributeInfoProvider) SettingsMapType {
	return SettingsMapType{
		p: p,
		MapType: basetypes.MapType{
			ElemType: types.StringType,
		},
	}
}
