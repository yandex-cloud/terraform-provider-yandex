package datalens

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestToInt64(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		input    interface{}
		expected int64
	}{
		{"float64", float64(42), 42},
		{"float64_zero", float64(0), 0},
		{"float64_negative", float64(-10), -10},
		{"float64_large", float64(1 << 52), 1 << 52},
		{"float64_fractional_returns_zero", float64(10.5), 0},
		{"float64_negative_fractional_returns_zero", float64(-10.9), 0},
		{"int64", int64(99), 99},
		{"int64_negative", int64(-5), -5},
		{"int", int(7), 7},
		{"string_returns_zero", "not a number", 0},
		{"nil_returns_zero", nil, 0},
		{"bool_returns_zero", true, 0},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ToInt64(c.input)
			if got != c.expected {
				t.Errorf("ToInt64(%v) = %d, want %d", c.input, got, c.expected)
			}
		})
	}
}

func TestToFloat64(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		input    interface{}
		expected float64
	}{
		{"float64", float64(3.14), 3.14},
		{"float64_zero", float64(0), 0},
		{"float64_negative", float64(-2.5), -2.5},
		{"int64", int64(99), 99.0},
		{"int64_negative", int64(-5), -5.0},
		{"int", int(7), 7.0},
		{"string_returns_zero", "not a number", 0},
		{"nil_returns_zero", nil, 0},
		{"bool_returns_zero", true, 0},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ToFloat64(c.input)
			if got != c.expected {
				t.Errorf("ToFloat64(%v) = %f, want %f", c.input, got, c.expected)
			}
		})
	}
}

func TestStringOrNull(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		resp     map[string]interface{}
		key      string
		expected types.String
	}{
		{
			name:     "present_string",
			resp:     map[string]interface{}{"foo": "bar"},
			key:      "foo",
			expected: types.StringValue("bar"),
		},
		{
			name:     "empty_string",
			resp:     map[string]interface{}{"foo": ""},
			key:      "foo",
			expected: types.StringValue(""),
		},
		{
			name:     "missing_key",
			resp:     map[string]interface{}{},
			key:      "foo",
			expected: types.StringNull(),
		},
		{
			name:     "nil_value",
			resp:     map[string]interface{}{"foo": nil},
			key:      "foo",
			expected: types.StringNull(),
		},
		{
			name:     "non_string_int",
			resp:     map[string]interface{}{"foo": 123},
			key:      "foo",
			expected: types.StringNull(),
		},
		{
			name:     "non_string_bool",
			resp:     map[string]interface{}{"foo": true},
			key:      "foo",
			expected: types.StringNull(),
		},
		{
			name:     "different_key",
			resp:     map[string]interface{}{"bar": "baz"},
			key:      "foo",
			expected: types.StringNull(),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := StringOrNull(c.resp, c.key)
			if !got.Equal(c.expected) {
				t.Errorf("StringOrNull(%v, %q) = %v, want %v", c.resp, c.key, got, c.expected)
			}
		})
	}
}
