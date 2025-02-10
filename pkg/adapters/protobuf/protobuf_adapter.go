package protobuf_adapter

import (
	"context"
	"fmt"
	"maps"
	"math/big"

	"reflect"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func NewProtobufMapDataAdapter() *ProtobufMapDataAdapter {
	return &ProtobufMapDataAdapter{}
}

type ProtobufMapDataAdapter struct{}

const protoTag = "protobuf"

var wrapperTypes = []reflect.Type{
	reflect.TypeOf(&wrapperspb.BoolValue{}),
	reflect.TypeOf(&wrapperspb.BytesValue{}),
	reflect.TypeOf(&wrapperspb.DoubleValue{}),
	reflect.TypeOf(&wrapperspb.FloatValue{}),
	reflect.TypeOf(&wrapperspb.Int32Value{}),
	reflect.TypeOf(&wrapperspb.Int64Value{}),
	reflect.TypeOf(&wrapperspb.StringValue{}),
	reflect.TypeOf(&wrapperspb.UInt32Value{}),
	reflect.TypeOf(&wrapperspb.UInt64Value{}),
}

var wrapperNulls = []attr.Value{
	types.BoolNull(),
	types.StringNull(),
	types.NumberNull(),
	types.NumberNull(),
	types.NumberNull(),
	types.NumberNull(),
	types.StringNull(),
	types.NumberNull(),
	types.NumberNull(),
}

var primitiveTypes = []reflect.Kind{
	reflect.Bool,
	reflect.Int,
	reflect.Int8,
	reflect.Int16,
	reflect.Int32,
	reflect.Int64,
	reflect.Uint,
	reflect.Uint8,
	reflect.Uint16,
	reflect.Uint32,
	reflect.Uint64,
	reflect.Float32,
	reflect.Float64,
	reflect.String,
}

var intTypes = []reflect.Kind{
	reflect.Int,
	reflect.Int8,
	reflect.Int16,
	reflect.Int32,
	reflect.Int64,
}

func (b *ProtobufMapDataAdapter) findTag(f reflect.StructField, tag, attrTagName string) (string, bool) {
	tagF := f.Tag.Get(tag)
	if tagF == "" {
		return "", false
	}

	tags := strings.Split(tagF, ",")
	for _, tag := range tags {
		kvs := strings.SplitN(tag, "=", 2)
		if kvs[0] != attrTagName {
			continue
		}
		val := ""
		if len(kvs) == 2 {
			val = kvs[1]
		}
		return val, true
	}
	return "", false
}

/*-------------------------Fill------------------------------------------*/

// Fill a target struct with provided attributes map
// target must be a pointer to struct
//
// Returns err if all attributes is not mapped to a struct
// Mapping fields by protobuf tag
// Mapping attribute type must be compatible with a target field type
func (f *ProtobufMapDataAdapter) Fill(ctx context.Context, target any, attributes map[string]attr.Value, diags *diag.Diagnostics) {
	unhandledAttrs := maps.Clone(attributes)
	f.fill(ctx, target, unhandledAttrs, diags)
}

func (f *ProtobufMapDataAdapter) fill(ctx context.Context, target any, attributes map[string]attr.Value, diags *diag.Diagnostics) {

	targetReflectValPtr := reflect.ValueOf(target)
	targetReflectTypePtr := reflect.TypeOf(target)

	if targetReflectTypePtr.Kind() != reflect.Ptr || targetReflectTypePtr.Elem().Kind() != reflect.Struct {
		diags.AddError("Error protobuf filler", "Target must be a pointer to struct")
		return
	}

	targetType := targetReflectTypePtr.Elem()
	targetReflectVal := targetReflectValPtr.Elem()

	for i := 0; i < targetReflectVal.NumField(); i++ {
		field := targetType.Field(i)
		if field.Tag.Get(protoTag) == "" {
			continue
		}

		// If pointer to struct
		if field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct && !slices.Contains(wrapperTypes, field.Type) {
			targetNestedField := targetReflectVal.Field(i)
			if targetNestedField.IsNil() {
				targetNestedField.Set(reflect.New(field.Type.Elem()))
			}
			f.fill(ctx, targetNestedField.Interface(), attributes, diags)

			if diags.HasError() {
				return
			}
			continue
		}

		// If struct
		if field.Type.Kind() == reflect.Struct {
			targetNestedField := targetReflectVal.Field(i).Addr().Interface()
			f.fill(ctx, targetNestedField, attributes, diags)
			continue
		}

		fieldName, ok := f.findTag(field, protoTag, "name")
		if !ok {
			continue
		}

		attrVal, ok := attributes[fieldName]
		delete(attributes, fieldName)

		if !ok || attrVal.IsNull() || attrVal.IsUnknown() {
			continue
		}

		setVal := f.mapAttributeToType(ctx, field.Type, attrVal, diags)
		if diags.HasError() {
			diags.AddError("Error protobuf filler", fmt.Sprintf("Attribute is not mapped for field %s", fieldName))
			return
		}

		targetReflectVal.Field(i).Set(setVal)

	}

	for key := range attributes {
		diags.AddError("Error protobuf filler", fmt.Sprintf("Attribute %s is not mapped", key))
	}
}

func (f *ProtobufMapDataAdapter) mapAttributeToType(ctx context.Context, t reflect.Type, attribute attr.Value, diags *diag.Diagnostics) reflect.Value {

	if t.Kind() == reflect.Ptr && slices.Contains(wrapperTypes, t) {
		return f.mapToWrapper(ctx, t, attribute, diags)
	}

	if t.Implements(reflect.TypeOf((*protoreflect.Enum)(nil)).Elem()) {
		return f.mapToEnum(ctx, t, attribute, diags)
	}

	if slices.Contains(primitiveTypes, t.Kind()) {
		return f.mapToPrimitive(ctx, t, attribute, diags)
	}

	if t.Kind() == reflect.Slice {
		return f.mapToSlice(ctx, t, attribute, diags)
	}

	diags.AddError("Error protobuf filler", fmt.Sprintf("%s type is not supported for mapping", t.Name()))
	return reflect.Value{}
}

func (f *ProtobufMapDataAdapter) mapToEnum(ctx context.Context, t reflect.Type, attribute attr.Value, diags *diag.Diagnostics) reflect.Value {
	if !slices.Contains(intTypes, t.Kind()) {
		diags.AddError("Error protobuf filler", fmt.Sprintf("Type %v is not enum: %v", t.Name(), t.Kind()))
		return reflect.Value{}
	}

	enumValue := reflect.New(t).Elem()

	switch attribute.Type(ctx) {
	case types.Int64Type:
		v := attribute.(types.Int64).ValueInt64()
		enumValue.SetInt(v)
	case types.NumberType:
		v := attribute.(types.Number).ValueBigFloat()
		if !v.IsInt() {
			diags.AddError("Error protobuf filler", fmt.Sprintf("Attribute for %s must be integer", t.Name()))
			return reflect.Value{}
		}
		vi, _ := v.Int64()
		enumValue.SetInt(vi)
	default:
		diags.AddError("Error protobuf filler", fmt.Sprintf("Attribute for %s must be integer", t.Name()))
		return reflect.Value{}
	}

	if !enumValue.IsValid() {
		diags.AddError("Error protobuf filler", fmt.Sprintf("Invalid value for enum: %d", enumValue.Int()))
		return reflect.Value{}
	}

	return enumValue
}

func (b *ProtobufMapDataAdapter) mapToWrapper(ctx context.Context, t reflect.Type, attribute attr.Value, diags *diag.Diagnostics) reflect.Value {
	switch t {
	// Bool
	case wrapperTypes[0]:
		v, ok := attribute.(types.Bool)
		if !ok {
			diags.AddError("Error protobuf filler", fmt.Sprintf("Attribute for %s must be bool", t.Name()))
			return reflect.Value{}
		}
		return reflect.ValueOf(wrapperspb.Bool(v.ValueBool()))
	// Double
	case wrapperTypes[3]:
		var attrVal float64
		switch attribute.Type(ctx) {
		case types.Float64Type:
			attrVal = attribute.(types.Float64).ValueFloat64()
		case types.NumberType:
			attrVal, _ = attribute.(types.Number).ValueBigFloat().Float64()
		default:
			diags.AddError("Error protobuf filler", fmt.Sprintf("Attribute for %s must be double", t.Name()))
			return reflect.Value{}
		}
		return reflect.ValueOf(wrapperspb.Double(attrVal))
	// Float
	case wrapperTypes[3]:
		var attrVal float32
		switch attribute.Type(ctx) {
		case types.Float64Type:
			attrVal = float32(attribute.(types.Float64).ValueFloat64())
		case types.NumberType:
			attrVal, _ = attribute.(types.Number).ValueBigFloat().Float32()
		default:
			diags.AddError("Error protobuf filler", fmt.Sprintf("Attribute for %s must be float", t.Name()))
			return reflect.Value{}
		}
		return reflect.ValueOf(wrapperspb.Float(attrVal))
	// Int64Type
	case wrapperTypes[4], wrapperTypes[5]:
		var attrVal int64
		switch attribute.Type(ctx) {
		case types.Int64Type:
			attrVal = attribute.(types.Int64).ValueInt64()
		case types.NumberType:
			v := attribute.(types.Number).ValueBigFloat()
			if !v.IsInt() {
				diags.AddError("Error protobuf filler", fmt.Sprintf("Attribute for %s must be integer", t.Name()))
				return reflect.Value{}
			}
			vi, _ := v.Int64()
			attrVal = vi
		default:
			diags.AddError("Error protobuf filler", fmt.Sprintf("Attribute for %s must be integer", t.Name()))
			return reflect.Value{}
		}
		if t == wrapperTypes[4] {
			return reflect.ValueOf(wrapperspb.Int32(int32(attrVal)))
		}
		return reflect.ValueOf(wrapperspb.Int64(attrVal))
	// String
	case wrapperTypes[6]:
		v, ok := attribute.(types.String)
		if !ok {
			diags.AddError("Error protobuf filler", fmt.Sprintf("Attribute for %s must be string", t.Name()))
			return reflect.Value{}
		}
		return reflect.ValueOf(wrapperspb.String(v.ValueString()))
	default:
		diags.AddError("Error protobuf filler", fmt.Sprintf("%s type is not supported for wrapped types", t.Name()))
		return reflect.Value{}
	}
}

func (b *ProtobufMapDataAdapter) mapToPrimitive(ctx context.Context, t reflect.Type, attribute attr.Value, diags *diag.Diagnostics) reflect.Value {
	switch t.Kind() {
	case reflect.Bool:
		v, ok := attribute.(types.Bool)
		if !ok {
			diags.AddError("Error protobuf filler", fmt.Sprintf("Attribute for %s must be bool", t.Name()))
			return reflect.Value{}
		}
		return reflect.ValueOf(v.ValueBool())
	case reflect.Float64:
		var attrVal float64
		switch attribute.Type(ctx) {
		case types.Float64Type:
			attrVal = attribute.(types.Float64).ValueFloat64()
		case types.NumberType:
			attrVal, _ = attribute.(types.Number).ValueBigFloat().Float64()
		default:
			diags.AddError("Error protobuf filler", fmt.Sprintf("Attribute for %s must be float", t.Name()))
			return reflect.Value{}
		}
		return reflect.ValueOf(attrVal)
	case reflect.Int64, reflect.Int32:
		var attrVal int64
		switch attribute.Type(ctx) {
		case types.Int64Type:
			attrVal = attribute.(types.Int64).ValueInt64()
		case types.NumberType:
			v := attribute.(types.Number).ValueBigFloat()
			if !v.IsInt() {
				diags.AddError("Error protobuf filler", fmt.Sprintf("Attribute for %s must be integer", t.Name()))
				return reflect.Value{}
			}
			vi, _ := v.Int64()
			attrVal = vi
		default:
			diags.AddError("Error protobuf filler", fmt.Sprintf("Attribute for %s must be integer", t.Name()))
			return reflect.Value{}
		}
		if t.Kind() == reflect.Int32 {
			return reflect.ValueOf(int32(attrVal))
		}
		return reflect.ValueOf(attrVal)
	case reflect.String:
		v, ok := attribute.(types.String)
		if !ok {
			diags.AddError("Error protobuf filler", fmt.Sprintf("Attribute for %s must be string", t.Name()))
			return reflect.Value{}
		}
		return reflect.ValueOf(v.ValueString())
	}

	return reflect.Value{}
}

func (b *ProtobufMapDataAdapter) mapToSlice(ctx context.Context, t reflect.Type, attribute attr.Value, diags *diag.Diagnostics) reflect.Value {
	if t.Kind() != reflect.Slice {
		diags.AddError("Error protobuf filler", fmt.Sprintf("Attribute for %s must be a iterable collection", t.Name()))
		return reflect.Value{}
	}

	var attrElems []attr.Value
	var attrCnt int

	type Collection interface {
		Elements() []attr.Value
	}

	collAttr, ok := attribute.(Collection)
	if !ok {
		diags.AddError("Error protobuf filler", fmt.Sprintf("Attribute for %s must be collection", t.Name()))
		return reflect.Value{}
	}
	attrElems = collAttr.Elements()
	attrCnt = len(attrElems)

	slice := reflect.MakeSlice(t, attrCnt, attrCnt)
	for i := 0; i < attrCnt; i++ {
		iVal := b.mapAttributeToType(ctx, t.Elem(), attrElems[i], diags)
		if diags.HasError() {
			diags.AddError("Error protobuf filler", fmt.Sprintf("Attribute collection elements is not map for %s", t.Name()))
			return reflect.Value{}
		}
		slice.Index(i).Set(iVal)
	}
	return slice
}

/*-------------------------------Extract-------------------------------*/

// Extract attr.Value from a provided struct by tag
// src must be a struct or a pointer to a struct
//
// For any float/int value returns types.NumberValue
// For any nil returns NullValue
// For any primitive value returns explicit Value
// For slices returns types.TupleValue
func (b *ProtobufMapDataAdapter) Extract(ctx context.Context, src any, diags *diag.Diagnostics) map[string]attr.Value {

	srcType := reflect.TypeOf(src)
	srcVal := reflect.ValueOf(src)

	attributes := make(map[string]attr.Value)

	if srcType.Kind() == reflect.Ptr && srcType.Elem().Kind() == reflect.Struct {
		if srcVal.IsNil() {
			return nil
		}
		return b.Extract(ctx, srcVal.Elem().Interface(), diags)
	}

	if srcType.Kind() != reflect.Struct {
		diags.AddError("Error protobuf filler", "Source must be a struct or a pointer to a struct")
		return nil
	}

	for i := 0; i < srcType.NumField(); i++ {
		field := srcType.Field(i)
		fieldName, ok := b.findTag(field, protoTag, "name")
		if !ok {
			continue
		}

		if field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct &&
			!slices.Contains(wrapperTypes, field.Type) || field.Type.Kind() == reflect.Struct {

			extendedAttributes := b.Extract(ctx, srcVal.Field(i).Interface(), diags)
			fmt.Println(extendedAttributes)
			if diags.HasError() {
				return nil
			}
			for k, v := range extendedAttributes {
				attributes[k] = v
			}
			continue
		}

		setVal := b.getAttributeFromReflectValue(ctx, field.Type, srcVal.Field(i), diags)
		if diags.HasError() {
			return nil
		}

		if setVal != nil {
			attributes[fieldName] = setVal
		}

	}

	return attributes
}

func (b *ProtobufMapDataAdapter) getAttributeFromReflectValue(ctx context.Context, srcType reflect.Type, srcVal reflect.Value, diags *diag.Diagnostics) attr.Value {

	if slices.Contains(wrapperTypes, srcType) {
		if srcVal.IsNil() {
			return wrapperNulls[slices.Index(wrapperTypes, srcType)]
		}
		srcTypeField, ok := srcType.Elem().FieldByName("Value")
		if !ok {
			diags.AddError("Error protobuf extract", fmt.Sprintf("Attribute it's not Value in wrapper type %s", srcType.Name()))
			return nil
		}

		srcType = srcTypeField.Type
		srcVal = srcVal.Elem().FieldByName("Value")
	}

	if srcType.Kind() == reflect.Slice {
		if srcVal.IsNil() {
			return types.TupleNull([]attr.Type{})
		}

		srcType = srcType.Elem()
		var attrElemType attr.Type
		switch srcType.Kind() {
		case reflect.Bool:
			attrElemType = types.BoolType
		case reflect.Int, reflect.Int32, reflect.Int64, reflect.Float64, reflect.Float32:
			attrElemType = types.NumberType
		case reflect.String:
			attrElemType = types.StringType
		}

		elements := make([]attr.Value, srcVal.Len())
		elementsType := make([]attr.Type, srcVal.Len())
		for i := 0; i < srcVal.Len(); i++ {
			attrElemVal := b.getAttributeFromReflectValue(ctx, srcType, srcVal.Index(i), diags)
			if diags.HasError() {
				return nil
			}
			elements[i] = attrElemVal
			elementsType[i] = attrElemType
		}

		res, d := types.TupleValue(elementsType, elements)
		if d.HasError() {
			diags.Append(d...)
			return nil
		}

		return res
	}

	switch srcType.Kind() {
	case reflect.Bool:
		return types.BoolValue(srcVal.Bool())
	case reflect.Int, reflect.Int32, reflect.Int64:
		return types.NumberValue(big.NewFloat(float64(srcVal.Int())))
	case reflect.String:
		return types.StringValue(srcVal.String())
	case reflect.Float64, reflect.Float32:
		return types.NumberValue(big.NewFloat(srcVal.Float()))
	default:
		diags.AddError("Error protobuf extractor", fmt.Sprintf("Unsupported type %s", srcType.Name()))
		return nil
	}
}
