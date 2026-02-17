package datalens

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// JSON response helpers.
//
// When unmarshaling JSON into map[string]interface{}, Go's encoding/json
// decodes all numbers as float64 (JSON has no integer type). These helpers
// provide safe conversions for use with DataLens API responses.

// StringOrNull extracts a string value from a response map, returning types.StringNull()
// if the key is missing or the value is nil.
func StringOrNull(resp map[string]interface{}, key string) types.String {
	v, ok := resp[key]
	if !ok || v == nil {
		return types.StringNull()
	}
	if s, ok := v.(string); ok {
		return types.StringValue(s)
	}
	return types.StringNull()
}

// ToInt64 converts a numeric interface{} (typically float64 from JSON) to int64.
// Returns 0 if the value is not a recognized numeric type or has a fractional part.
func ToInt64(v interface{}) int64 {
	switch n := v.(type) {
	case float64:
		if n != float64(int64(n)) {
			return 0
		}
		return int64(n)
	case int64:
		return n
	case int:
		return int64(n)
	default:
		return 0
	}
}

// ToFloat64 converts a numeric interface{} (typically float64 from JSON) to float64.
// Returns 0 if the value is not a recognized numeric type.
func ToFloat64(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int64:
		return float64(n)
	case int:
		return float64(n)
	default:
		return 0
	}
}
