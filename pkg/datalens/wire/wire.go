// Package wire converts Terraform Plugin Framework models to/from
// map[string]interface{} payloads for the DataLens REST/JSON-RPC API.
//
// The same Go struct serves as both the Terraform model (via `tfsdk:"..."` tags)
// and the wire DTO (via `wire:"..."` tags). Marshal omits Null/Unknown attr.Values,
// nil pointers, and empty slices/maps (proto3-style). Unmarshal leaves missing
// keys as Null.
package wire

import (
	"context"
	"fmt"
	"math/big"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// fieldTag holds the parsed `wire:"name[,opt1,opt2,...]"` tag.
//
// Recognized options:
//
//	nullIfEmpty — unmarshal: treat zero-value scalar (empty string, 0, false) as Null.
//	alwaysEmit  — marshal: emit an empty container (`[]` for slices, `{}` for
//	              maps and nil/empty *struct/ObjectValue) instead of omitting
//	              the field. Use for API-required fields the server rejects
//	              when missing.
type fieldTag struct {
	name        string
	skip        bool
	nullIfEmpty bool
	alwaysEmit  bool
}

func parseTag(t string) fieldTag {
	if t == "" {
		return fieldTag{skip: true}
	}
	parts := strings.Split(t, ",")
	ft := fieldTag{name: parts[0]}
	if ft.name == "-" {
		ft.skip = true
		return ft
	}
	for _, opt := range parts[1:] {
		switch opt {
		case "nullIfEmpty":
			ft.nullIfEmpty = true
		case "alwaysEmit":
			ft.alwaysEmit = true
		}
	}
	if ft.name == "" {
		ft.skip = true
	}
	return ft
}

// Marshal converts a Terraform-Framework struct (typically a *resourceModel)
// into a map suitable for use as a JSON-RPC request body. Field names come
// from the `wire:"..."` tag. Untagged fields and `wire:"-"` are skipped.
//
// Omission rules:
//   - attr.Value with IsNull() or IsUnknown() → omitted
//   - nil pointer → omitted
//   - empty slice/map → omitted
//   - empty nested struct (no marshalled fields) → omitted
func Marshal(v any) (map[string]any, error) {
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil, nil
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil, fmt.Errorf("wire.Marshal: expected struct, got %s", rv.Kind())
	}
	out, _, err := marshalStruct(rv)
	if err != nil {
		return nil, err
	}
	if out == nil {
		out = map[string]any{}
	}
	return out, nil
}

// Initializer is implemented by struct types that need to pre-populate
// `types.Map`/`types.List` fields with the right ElementType before wire
// fills them from the response. Required for slice-of-struct element types
// that wire allocates from scratch via reflect.MakeSlice — those allocations
// produce zero-value attr.Values whose ElementType is nil, which wire can't
// recover without a hint.
type Initializer interface {
	Init()
}

// Unmarshal copies values from raw into the Terraform-Framework struct pointed
// to by v. Fields whose `wire:"..."` key is absent in raw are left untouched
// (so a multi-pass call sequence can fill the same model from several scopes).
// After the pass, any attr.Value field that is still Unknown is normalized to
// Null — Unknown must not survive into state, and the API is authoritative for
// whichever fields it returned.
func Unmarshal(raw map[string]any, v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("wire.Unmarshal: expected pointer to struct, got %T", v)
	}
	if err := unmarshalStruct(raw, rv.Elem()); err != nil {
		return err
	}
	normalizeUnknown(rv.Elem())
	return nil
}

// normalizeUnknown walks a struct value and replaces any Unknown attr.Value
// scalars with their Null equivalent. List/Set/Map element types are reused.
// Recurses into nested structs / *structs / []structs.
func normalizeUnknown(rv reflect.Value) {
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return
	}
	for i := 0; i < rv.NumField(); i++ {
		fv := rv.Field(i)
		if fv.CanInterface() {
			if av, ok := fv.Interface().(attr.Value); ok {
				if av.IsUnknown() {
					_ = setAttrNullLike(fv)
				}
				continue
			}
		}
		switch fv.Kind() {
		case reflect.Ptr, reflect.Struct:
			normalizeUnknown(fv)
		case reflect.Slice:
			for j := 0; j < fv.Len(); j++ {
				normalizeUnknown(fv.Index(j))
			}
		}
	}
}

// ---- marshal -----------------------------------------------------------------

func marshalStruct(rv reflect.Value) (map[string]any, bool, error) {
	out := map[string]any{}
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		ft := parseTag(f.Tag.Get("wire"))
		if ft.skip {
			continue
		}
		val, omit, err := marshalField(rv.Field(i))
		if err != nil {
			return nil, false, fmt.Errorf("%s: %w", ft.name, err)
		}
		if omit {
			if !ft.alwaysEmit {
				continue
			}
			val = emptyContainerFor(rv.Field(i))
		}
		out[ft.name] = val
	}
	if len(out) == 0 {
		return nil, true, nil
	}
	return out, false, nil
}

// emptyContainerFor returns the empty wire-shape ([]any{} or map[string]any{})
// matching the field's declared type, used by the `,alwaysEmit` tag option to
// emit an empty container instead of omitting the field when its value is
// empty/null/nil.
func emptyContainerFor(rv reflect.Value) any {
	if rv.CanInterface() {
		if av, ok := rv.Interface().(attr.Value); ok {
			switch av.(type) {
			case basetypes.ListValue, basetypes.SetValue:
				return []any{}
			case basetypes.MapValue, basetypes.ObjectValue:
				return map[string]any{}
			}
			if _, ok := av.(basetypes.ObjectValuable); ok {
				return map[string]any{}
			}
			return nil
		}
	}
	switch rv.Kind() {
	case reflect.Slice:
		return []any{}
	case reflect.Map, reflect.Ptr, reflect.Struct:
		return map[string]any{}
	}
	return nil
}

func marshalField(rv reflect.Value) (any, bool, error) {
	if rv.CanInterface() {
		if av, ok := rv.Interface().(attr.Value); ok {
			return marshalAttr(av)
		}
	}
	switch rv.Kind() {
	case reflect.Ptr:
		if rv.IsNil() {
			return nil, true, nil
		}
		return marshalField(rv.Elem())
	case reflect.Struct:
		m, omit, err := marshalStruct(rv)
		if err != nil {
			return nil, false, err
		}
		return m, omit, nil
	case reflect.Slice:
		if rv.Len() == 0 {
			return nil, true, nil
		}
		out := make([]any, 0, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			v, omit, err := marshalField(rv.Index(i))
			if err != nil {
				return nil, false, fmt.Errorf("[%d]: %w", i, err)
			}
			if omit {
				continue
			}
			out = append(out, v)
		}
		if len(out) == 0 {
			return nil, true, nil
		}
		return out, false, nil
	case reflect.Map:
		if rv.IsNil() {
			return nil, true, nil
		}
		out := make(map[string]any, rv.Len())
		iter := rv.MapRange()
		for iter.Next() {
			v, omit, err := marshalField(iter.Value())
			if err != nil {
				return nil, false, fmt.Errorf("[%q]: %w", iter.Key().String(), err)
			}
			if omit {
				continue
			}
			out[iter.Key().String()] = v
		}
		return out, false, nil
	case reflect.String:
		return rv.String(), false, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int(), false, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rv.Uint(), false, nil
	case reflect.Float32, reflect.Float64:
		return rv.Float(), false, nil
	case reflect.Bool:
		return rv.Bool(), false, nil
	}
	return nil, false, fmt.Errorf("unsupported field kind %s", rv.Kind())
}

func marshalAttr(av attr.Value) (any, bool, error) {
	if av.IsNull() || av.IsUnknown() {
		return nil, true, nil
	}
	switch v := av.(type) {
	case basetypes.StringValue:
		return v.ValueString(), false, nil
	case basetypes.Int64Value:
		return v.ValueInt64(), false, nil
	case basetypes.Float64Value:
		return v.ValueFloat64(), false, nil
	case basetypes.BoolValue:
		return v.ValueBool(), false, nil
	case basetypes.NumberValue:
		f, _ := v.ValueBigFloat().Float64()
		return f, false, nil
	case basetypes.ListValue:
		return marshalAttrSlice(v.Elements())
	case basetypes.SetValue:
		return marshalAttrSlice(v.Elements())
	case basetypes.MapValue:
		return marshalAttrMap(v.Elements())
	case basetypes.ObjectValue:
		return marshalAttrMap(v.Attributes())
	}
	// Custom ObjectValuable types (codegen-generated wrappers around ObjectValue).
	if ov, ok := av.(basetypes.ObjectValuable); ok {
		obj, diags := ov.ToObjectValue(context.Background())
		if diags.HasError() {
			return nil, false, fmt.Errorf("ObjectValuable: %s", diags)
		}
		return marshalAttrMap(obj.Attributes())
	}
	return nil, false, fmt.Errorf("unsupported attr.Value type %T", av)
}

func marshalAttrSlice(elems []attr.Value) (any, bool, error) {
	if len(elems) == 0 {
		return nil, true, nil
	}
	out := make([]any, 0, len(elems))
	for _, e := range elems {
		v, omit, err := marshalAttr(e)
		if err != nil {
			return nil, false, err
		}
		if omit {
			out = append(out, nil)
			continue
		}
		out = append(out, v)
	}
	return out, false, nil
}

func marshalAttrMap(elems map[string]attr.Value) (any, bool, error) {
	if len(elems) == 0 {
		return nil, true, nil
	}
	out := make(map[string]any, len(elems))
	for k, e := range elems {
		v, omit, err := marshalAttr(e)
		if err != nil {
			return nil, false, err
		}
		if omit {
			continue
		}
		out[k] = v
	}
	if len(out) == 0 {
		return nil, true, nil
	}
	return out, false, nil
}

// ---- unmarshal ---------------------------------------------------------------

func unmarshalStruct(raw map[string]any, rv reflect.Value) error {
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		ft := parseTag(f.Tag.Get("wire"))
		if ft.skip {
			continue
		}
		val, ok := raw[ft.name]
		if !ok {
			continue
		}
		if ft.nullIfEmpty && isZeroScalar(val) {
			val = nil
		}
		if err := unmarshalField(val, rv.Field(i)); err != nil {
			return fmt.Errorf("%s: %w", ft.name, err)
		}
	}
	return nil
}

func isZeroScalar(v any) bool {
	switch x := v.(type) {
	case string:
		return x == ""
	case float64:
		return x == 0
	case float32:
		return x == 0
	case int:
		return x == 0
	case int64:
		return x == 0
	case bool:
		return !x
	}
	return false
}

func unmarshalField(val any, rv reflect.Value) error {
	if rv.CanInterface() {
		if _, ok := rv.Interface().(attr.Value); ok {
			return unmarshalAttr(val, rv)
		}
	}
	switch rv.Kind() {
	case reflect.Ptr:
		if val == nil {
			return nil
		}
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
			if init, ok := rv.Interface().(Initializer); ok {
				init.Init()
			}
		}
		return unmarshalField(val, rv.Elem())
	case reflect.Struct:
		m, ok := val.(map[string]any)
		if !ok {
			return fmt.Errorf("expected map for struct %s, got %T", rv.Type(), val)
		}
		return unmarshalStruct(m, rv)
	case reflect.Slice:
		list, ok := val.([]any)
		if !ok {
			return fmt.Errorf("expected slice for %s, got %T", rv.Type(), val)
		}
		// Empty list from the API is symmetric with Marshal omitting an empty
		// slice: leave the field as the plan/state set it (nil or empty) so
		// `null` plans don't see a `[]` state diff.
		if len(list) == 0 {
			return nil
		}
		existing := rv
		slice := reflect.MakeSlice(rv.Type(), len(list), len(list))
		for i, item := range list {
			elem := slice.Index(i)
			// Carry over plan-side initialization for elements that exist in the
			// previous slice — preserves nested empty-vs-nil semantics on
			// child slices and the ElementType on typed Map/List fields.
			if i < existing.Len() {
				elem.Set(existing.Index(i))
			} else if elem.CanAddr() {
				if init, ok := elem.Addr().Interface().(Initializer); ok {
					init.Init()
				}
			}
			if err := unmarshalField(item, elem); err != nil {
				return fmt.Errorf("[%d]: %w", i, err)
			}
		}
		rv.Set(slice)
		return nil
	case reflect.Map:
		raw, ok := val.(map[string]any)
		if !ok {
			return fmt.Errorf("expected map for %s, got %T", rv.Type(), val)
		}
		// Empty map from the API: preserve plan-side nil so a `null` plan
		// doesn't see a `{}` state diff. Mirrors the slice branch above.
		if len(raw) == 0 {
			return nil
		}
		out := reflect.MakeMapWithSize(rv.Type(), len(raw))
		elemType := rv.Type().Elem()
		for k, v := range raw {
			elem := reflect.New(elemType).Elem()
			if err := unmarshalField(v, elem); err != nil {
				return fmt.Errorf("[%q]: %w", k, err)
			}
			out.SetMapIndex(reflect.ValueOf(k), elem)
		}
		rv.Set(out)
		return nil
	case reflect.String:
		s, ok := toString(val)
		if !ok {
			return fmt.Errorf("expected string, got %T", val)
		}
		rv.SetString(s)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, ok := toInt64(val)
		if !ok {
			return fmt.Errorf("expected int, got %T", val)
		}
		rv.SetInt(i)
		return nil
	case reflect.Float32, reflect.Float64:
		f, ok := toFloat64(val)
		if !ok {
			return fmt.Errorf("expected float, got %T", val)
		}
		rv.SetFloat(f)
		return nil
	case reflect.Bool:
		b, ok := val.(bool)
		if !ok {
			return fmt.Errorf("expected bool, got %T", val)
		}
		rv.SetBool(b)
		return nil
	}
	return fmt.Errorf("unsupported field kind %s", rv.Kind())
}

func unmarshalAttr(val any, rv reflect.Value) error {
	if val == nil {
		return setAttrNullLike(rv)
	}
	cur := rv.Interface().(attr.Value)
	switch cur.(type) {
	case basetypes.StringValue:
		s, ok := toString(val)
		if !ok {
			return fmt.Errorf("expected string, got %T", val)
		}
		rv.Set(reflect.ValueOf(types.StringValue(s)))
	case basetypes.Int64Value:
		i, ok := toInt64(val)
		if !ok {
			return fmt.Errorf("expected int, got %T", val)
		}
		rv.Set(reflect.ValueOf(types.Int64Value(i)))
	case basetypes.Float64Value:
		f, ok := toFloat64(val)
		if !ok {
			return fmt.Errorf("expected number, got %T", val)
		}
		rv.Set(reflect.ValueOf(types.Float64Value(f)))
	case basetypes.BoolValue:
		b, ok := val.(bool)
		if !ok {
			return fmt.Errorf("expected bool, got %T", val)
		}
		rv.Set(reflect.ValueOf(types.BoolValue(b)))
	case basetypes.ListValue:
		// Symmetric with empty-slice marshal omission: leave plan-side value
		// alone when the API returns an empty list.
		if list, ok := val.([]any); ok && len(list) == 0 {
			return nil
		}
		lv, err := buildListValue(cur.(basetypes.ListValue).ElementType(context.Background()), val)
		if err != nil {
			return err
		}
		rv.Set(reflect.ValueOf(lv))
	case basetypes.SetValue:
		if list, ok := val.([]any); ok && len(list) == 0 {
			return nil
		}
		lv, err := buildSetValue(cur.(basetypes.SetValue).ElementType(context.Background()), val)
		if err != nil {
			return err
		}
		rv.Set(reflect.ValueOf(lv))
	case basetypes.MapValue:
		if m, ok := val.(map[string]any); ok && len(m) == 0 {
			return nil
		}
		mv, err := buildMapValue(cur.(basetypes.MapValue).ElementType(context.Background()), val)
		if err != nil {
			return err
		}
		rv.Set(reflect.ValueOf(mv))
	default:
		// Custom ObjectValuable types (codegen-generated wrappers).
		if ov, ok := cur.(basetypes.ObjectValuable); ok {
			objType, ok := ov.Type(context.Background()).(basetypes.ObjectTypable)
			if !ok {
				return fmt.Errorf("ObjectValuable type %T does not implement ObjectTypable", ov)
			}
			attrTypes := objType.(interface {
				AttributeTypes() map[string]attr.Type
			}).AttributeTypes()
			obj, err := buildObjectValue(attrTypes, val)
			if err != nil {
				return err
			}
			converted, diags := objType.ValueFromObject(context.Background(), obj)
			if diags.HasError() {
				return fmt.Errorf("ObjectValuable conversion: %s", diags)
			}
			rv.Set(reflect.ValueOf(converted))
			return nil
		}
		return fmt.Errorf("unsupported attr.Value type %T", cur)
	}
	return nil
}

func setAttrNullLike(rv reflect.Value) error {
	cur := rv.Interface().(attr.Value)
	switch cur.(type) {
	case basetypes.StringValue:
		rv.Set(reflect.ValueOf(types.StringNull()))
	case basetypes.Int64Value:
		rv.Set(reflect.ValueOf(types.Int64Null()))
	case basetypes.Float64Value:
		rv.Set(reflect.ValueOf(types.Float64Null()))
	case basetypes.BoolValue:
		rv.Set(reflect.ValueOf(types.BoolNull()))
	case basetypes.ListValue:
		rv.Set(reflect.ValueOf(types.ListNull(cur.(basetypes.ListValue).ElementType(context.Background()))))
	case basetypes.SetValue:
		rv.Set(reflect.ValueOf(types.SetNull(cur.(basetypes.SetValue).ElementType(context.Background()))))
	case basetypes.MapValue:
		rv.Set(reflect.ValueOf(types.MapNull(cur.(basetypes.MapValue).ElementType(context.Background()))))
	default:
		// Custom ObjectValuable types: zero value is the Null state (state field
		// defaults to attr.ValueStateNull).
		if _, ok := cur.(basetypes.ObjectValuable); ok {
			rv.Set(reflect.Zero(rv.Type()))
		}
	}
	return nil
}

// ---- list/set/map element builders ------------------------------------------

func buildListValue(elemType attr.Type, val any) (basetypes.ListValue, error) {
	list, ok := val.([]any)
	if !ok {
		return basetypes.ListValue{}, fmt.Errorf("expected slice, got %T", val)
	}
	elems := make([]attr.Value, 0, len(list))
	for i, item := range list {
		ev, err := buildAttrValue(elemType, item)
		if err != nil {
			return basetypes.ListValue{}, fmt.Errorf("[%d]: %w", i, err)
		}
		elems = append(elems, ev)
	}
	lv, diags := types.ListValue(elemType, elems)
	if diags.HasError() {
		return basetypes.ListValue{}, fmt.Errorf("list value: %s", diags)
	}
	return lv, nil
}

func buildSetValue(elemType attr.Type, val any) (basetypes.SetValue, error) {
	list, ok := val.([]any)
	if !ok {
		return basetypes.SetValue{}, fmt.Errorf("expected slice, got %T", val)
	}
	elems := make([]attr.Value, 0, len(list))
	for i, item := range list {
		ev, err := buildAttrValue(elemType, item)
		if err != nil {
			return basetypes.SetValue{}, fmt.Errorf("[%d]: %w", i, err)
		}
		elems = append(elems, ev)
	}
	sv, diags := types.SetValue(elemType, elems)
	if diags.HasError() {
		return basetypes.SetValue{}, fmt.Errorf("set value: %s", diags)
	}
	return sv, nil
}

// buildObjectValue constructs a basetypes.ObjectValue from a JSON-decoded map,
// using the schema-declared per-attribute types to recursively decode each
// value. Missing attributes are filled with their type's null equivalent.
func buildObjectValue(attrTypes map[string]attr.Type, val any) (basetypes.ObjectValue, error) {
	src, ok := val.(map[string]any)
	if !ok {
		return basetypes.ObjectValue{}, fmt.Errorf("expected object map, got %T", val)
	}
	elems := make(map[string]attr.Value, len(attrTypes))
	for name, t := range attrTypes {
		raw, present := src[name]
		if !present {
			raw = nil
		}
		ev, err := buildAttrValue(t, raw)
		if err != nil {
			return basetypes.ObjectValue{}, fmt.Errorf("[%q]: %w", name, err)
		}
		elems[name] = ev
	}
	ov, diags := types.ObjectValue(attrTypes, elems)
	if diags.HasError() {
		return basetypes.ObjectValue{}, fmt.Errorf("object value: %s", diags)
	}
	return ov, nil
}

func buildMapValue(elemType attr.Type, val any) (basetypes.MapValue, error) {
	src, ok := val.(map[string]any)
	if !ok {
		return basetypes.MapValue{}, fmt.Errorf("expected map, got %T", val)
	}
	elems := make(map[string]attr.Value, len(src))
	for k, item := range src {
		ev, err := buildAttrValue(elemType, item)
		if err != nil {
			return basetypes.MapValue{}, fmt.Errorf("[%q]: %w", k, err)
		}
		elems[k] = ev
	}
	mv, diags := types.MapValue(elemType, elems)
	if diags.HasError() {
		return basetypes.MapValue{}, fmt.Errorf("map value: %s", diags)
	}
	return mv, nil
}

func buildAttrValue(t attr.Type, val any) (attr.Value, error) {
	if val == nil {
		switch tt := t.(type) {
		case basetypes.StringType:
			return types.StringNull(), nil
		case basetypes.Int64Type:
			return types.Int64Null(), nil
		case basetypes.Float64Type:
			return types.Float64Null(), nil
		case basetypes.BoolType:
			return types.BoolNull(), nil
		case basetypes.ListType:
			return types.ListNull(tt.ElemType), nil
		case basetypes.SetType:
			return types.SetNull(tt.ElemType), nil
		case basetypes.MapType:
			return types.MapNull(tt.ElemType), nil
		case basetypes.ObjectType:
			return types.ObjectNull(tt.AttrTypes), nil
		}
		return nil, fmt.Errorf("nil value not supported for type %T", t)
	}
	switch tt := t.(type) {
	case basetypes.StringType:
		s, ok := toString(val)
		if !ok {
			return nil, fmt.Errorf("expected string, got %T", val)
		}
		return types.StringValue(s), nil
	case basetypes.Int64Type:
		i, ok := toInt64(val)
		if !ok {
			return nil, fmt.Errorf("expected int, got %T", val)
		}
		return types.Int64Value(i), nil
	case basetypes.Float64Type:
		f, ok := toFloat64(val)
		if !ok {
			return nil, fmt.Errorf("expected number, got %T", val)
		}
		return types.Float64Value(f), nil
	case basetypes.BoolType:
		b, ok := val.(bool)
		if !ok {
			return nil, fmt.Errorf("expected bool, got %T", val)
		}
		return types.BoolValue(b), nil
	case basetypes.ListType:
		return buildListValue(tt.ElemType, val)
	case basetypes.SetType:
		return buildSetValue(tt.ElemType, val)
	case basetypes.MapType:
		return buildMapValue(tt.ElemType, val)
	case basetypes.ObjectType:
		return buildObjectValue(tt.AttrTypes, val)
	}
	return nil, fmt.Errorf("unsupported element type %T", t)
}

// ---- value coercion helpers --------------------------------------------------

func toString(v any) (string, bool) {
	if s, ok := v.(string); ok {
		return s, true
	}
	return "", false
}

func toInt64(v any) (int64, bool) {
	switch x := v.(type) {
	case int:
		return int64(x), true
	case int32:
		return int64(x), true
	case int64:
		return x, true
	case float32:
		return int64(x), true
	case float64:
		return int64(x), true
	}
	return 0, false
}

func toFloat64(v any) (float64, bool) {
	switch x := v.(type) {
	case float32:
		return float64(x), true
	case float64:
		return x, true
	case int:
		return float64(x), true
	case int32:
		return float64(x), true
	case int64:
		return float64(x), true
	case *big.Float:
		f, _ := x.Float64()
		return f, true
	}
	return 0, false
}
