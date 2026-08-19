package wire

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type scalars struct {
	Name        types.String  `tfsdk:"name"        wire:"name"`
	Count       types.Int64   `tfsdk:"count"       wire:"count"`
	Ratio       types.Float64 `tfsdk:"ratio"       wire:"ratio"`
	Enabled     types.Bool    `tfsdk:"enabled"     wire:"enabled"`
	Description types.String  `tfsdk:"description" wire:"description"`
	Internal    types.String  `tfsdk:"internal"    wire:"-"`
}

func TestMarshal_ScalarsOmitNullAndUnknown(t *testing.T) {
	in := &scalars{
		Name:        types.StringValue("dash"),
		Count:       types.Int64Value(1),
		Ratio:       types.Float64Value(0.5),
		Enabled:     types.BoolValue(false), // false is a real value, not zero-omit
		Description: types.StringNull(),     // omitted
		Internal:    types.StringValue("x"), // wire:"-" -> omitted
	}
	got, err := Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := map[string]any{
		"name":    "dash",
		"count":   int64(1),
		"ratio":   0.5,
		"enabled": false,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Marshal mismatch:\n got: %#v\nwant: %#v", got, want)
	}
}

func TestUnmarshal_ScalarsMissingKeyStaysNull(t *testing.T) {
	out := &scalars{
		// pre-populated to ensure types are known
		Name:        types.StringNull(),
		Count:       types.Int64Null(),
		Ratio:       types.Float64Null(),
		Enabled:     types.BoolNull(),
		Description: types.StringNull(),
	}
	raw := map[string]any{
		"name":    "x",
		"count":   float64(7), // JSON often surfaces as float64
		"ratio":   0.25,
		"enabled": true,
	}
	if err := Unmarshal(raw, out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.Name.ValueString() != "x" {
		t.Errorf("Name: %v", out.Name)
	}
	if out.Count.ValueInt64() != 7 {
		t.Errorf("Count: %v", out.Count)
	}
	if out.Ratio.ValueFloat64() != 0.25 {
		t.Errorf("Ratio: %v", out.Ratio)
	}
	if !out.Enabled.ValueBool() {
		t.Errorf("Enabled: %v", out.Enabled)
	}
	if !out.Description.IsNull() {
		t.Errorf("Description should stay null, got %v", out.Description)
	}
}

type nested struct {
	Title types.String `tfsdk:"title" wire:"title"`
}

type withNested struct {
	Outer types.String `tfsdk:"outer" wire:"outer"`
	Inner *nested      `tfsdk:"inner" wire:"inner"`
}

func TestMarshal_NilPointerOmitted(t *testing.T) {
	in := &withNested{Outer: types.StringValue("o"), Inner: nil}
	got, _ := Marshal(in)
	want := map[string]any{"outer": "o"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestMarshal_EmptyNestedStructOmitted(t *testing.T) {
	in := &withNested{Outer: types.StringValue("o"), Inner: &nested{Title: types.StringNull()}}
	got, _ := Marshal(in)
	want := map[string]any{"outer": "o"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestRoundTrip_Nested(t *testing.T) {
	in := &withNested{
		Outer: types.StringValue("o"),
		Inner: &nested{Title: types.StringValue("t")},
	}
	raw, err := Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	out := &withNested{
		Outer: types.StringNull(),
		Inner: &nested{Title: types.StringNull()},
	}
	if err := Unmarshal(raw, out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if out.Outer.ValueString() != "o" {
		t.Errorf("Outer: %v", out.Outer)
	}
	if out.Inner == nil || out.Inner.Title.ValueString() != "t" {
		t.Errorf("Inner: %+v", out.Inner)
	}
}

type tab struct {
	Id    types.String `tfsdk:"id"    wire:"id"`
	Title types.String `tfsdk:"title" wire:"title"`
}

type withList struct {
	Tabs []tab `tfsdk:"tabs" wire:"tabs"`
}

func TestMarshal_EmptyListOmitted(t *testing.T) {
	got, _ := Marshal(&withList{Tabs: nil})
	if len(got) != 0 {
		t.Errorf("expected empty map, got %#v", got)
	}
	got, _ = Marshal(&withList{Tabs: []tab{}})
	if len(got) != 0 {
		t.Errorf("expected empty map, got %#v", got)
	}
}

func TestRoundTrip_ListOfStructs(t *testing.T) {
	in := &withList{Tabs: []tab{
		{Id: types.StringValue("a"), Title: types.StringValue("A")},
		{Id: types.StringValue("b"), Title: types.StringNull()},
	}}
	raw, err := Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	wantTabs := []any{
		map[string]any{"id": "a", "title": "A"},
		map[string]any{"id": "b"}, // null title omitted
	}
	if !reflect.DeepEqual(raw["tabs"], wantTabs) {
		t.Errorf("got tabs %#v, want %#v", raw["tabs"], wantTabs)
	}

	out := &withList{}
	if err := Unmarshal(raw, out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if len(out.Tabs) != 2 {
		t.Fatalf("Tabs len: %d", len(out.Tabs))
	}
	if out.Tabs[0].Id.ValueString() != "a" || out.Tabs[0].Title.ValueString() != "A" {
		t.Errorf("Tabs[0]: %+v", out.Tabs[0])
	}
	if out.Tabs[1].Id.ValueString() != "b" || !out.Tabs[1].Title.IsNull() {
		t.Errorf("Tabs[1]: %+v", out.Tabs[1])
	}
}

type withTypedList struct {
	Names types.List `tfsdk:"names" wire:"names"`
	Tags  types.Map  `tfsdk:"tags"  wire:"tags"`
}

func TestRoundTrip_TypedListAndMap(t *testing.T) {
	in := &withTypedList{
		Names: mustList(t, types.StringType, []string{"a", "b"}),
		Tags:  mustMap(t, types.StringType, map[string]string{"k": "v"}),
	}
	raw, err := Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if !reflect.DeepEqual(raw["names"], []any{"a", "b"}) {
		t.Errorf("names: %#v", raw["names"])
	}
	if !reflect.DeepEqual(raw["tags"], map[string]any{"k": "v"}) {
		t.Errorf("tags: %#v", raw["tags"])
	}

	// Unmarshal needs the existing value to know the element type.
	out := &withTypedList{
		Names: types.ListNull(types.StringType),
		Tags:  types.MapNull(types.StringType),
	}
	if err := Unmarshal(raw, out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got := out.Names.Elements(); len(got) != 2 ||
		got[0].(types.String).ValueString() != "a" ||
		got[1].(types.String).ValueString() != "b" {
		t.Errorf("names: %v", got)
	}
	if got := out.Tags.Elements(); len(got) != 1 || got["k"].(types.String).ValueString() != "v" {
		t.Errorf("tags: %v", got)
	}
}

type withAliases struct {
	// Map(List(List(String))) — used by dashboard tab aliases.
	Aliases types.Map `tfsdk:"aliases" wire:"aliases"`
}

func TestRoundTrip_NestedListInMap(t *testing.T) {
	innerType := types.ListType{ElemType: types.ListType{ElemType: types.StringType}}
	outer, diags := types.ListValue(types.ListType{ElemType: types.StringType},
		[]attr.Value{mustList(t, types.StringType, []string{"x", "y"})})
	if diags.HasError() {
		t.Fatalf("inner list: %v", diags)
	}
	mv, diags := types.MapValue(innerType, map[string]attr.Value{"default": outer})
	if diags.HasError() {
		t.Fatalf("aliases map: %v", diags)
	}
	in := &withAliases{Aliases: mv}

	raw, err := Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	got := raw["aliases"].(map[string]any)["default"].([]any)
	if !reflect.DeepEqual(got, []any{[]any{"x", "y"}}) {
		t.Errorf("aliases: %#v", got)
	}

	out := &withAliases{Aliases: types.MapNull(innerType)}
	if err := Unmarshal(raw, out); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if l := len(out.Aliases.Elements()); l != 1 {
		t.Errorf("aliases len: %d", l)
	}
}

// ---- helpers ---------------------------------------------------------------

func mustList(t *testing.T, et attr.Type, in any) types.List {
	t.Helper()
	var elems []attr.Value
	switch v := in.(type) {
	case []string:
		for _, s := range v {
			elems = append(elems, types.StringValue(s))
		}
	default:
		t.Fatalf("mustList: unsupported %T", in)
	}
	lv, diags := types.ListValue(et, elems)
	if diags.HasError() {
		t.Fatalf("ListValue: %v", diags)
	}
	return lv
}

func mustMap(t *testing.T, et attr.Type, in map[string]string) types.Map {
	t.Helper()
	elems := make(map[string]attr.Value, len(in))
	for k, v := range in {
		elems[k] = types.StringValue(v)
	}
	mv, diags := types.MapValue(et, elems)
	if diags.HasError() {
		t.Fatalf("MapValue: %v", diags)
	}
	return mv
}
