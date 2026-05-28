package mdbcommon

import (
	"strconv"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// RawConfigProvider is implemented by both *schema.ResourceData and
// *schema.ResourceDiff: it exposes the user's raw configuration as a cty.Value.
type RawConfigProvider interface {
	GetRawConfig() cty.Value
}

var (
	_ RawConfigProvider = (*schema.ResourceData)(nil)
	_ RawConfigProvider = (*schema.ResourceDiff)(nil)
)

// LookupRawConfigPath returns the cty.Value at a dot-separated path inside the resource's
// raw configuration. Numeric segments index lists/sets, like ResourceData/ResourceDiff keys.
// ok == false means the value is not present or not known: any intermediate or final
// null/unknown, missing attribute, or out-of-range index.
func LookupRawConfigPath(d RawConfigProvider, path string) (cty.Value, bool) {
	rawCfg := d.GetRawConfig()
	if rawCfg.IsNull() || !rawCfg.IsKnown() {
		return cty.NilVal, false
	}

	v, ok := lookupCtyPath(rawCfg, stringPathToCtyPath(path))
	if !ok || v.IsNull() || !v.IsKnown() {
		return cty.NilVal, false
	}
	return v, true
}

// stringPathToCtyPath converts a dot-separated path into a cty.Path; numeric segments
// become list/set indices. An empty input yields an empty path (= the root value).
func stringPathToCtyPath(path string) cty.Path {
	if path == "" {
		return nil
	}
	parts := strings.Split(path, ".")
	out := make(cty.Path, 0, len(parts))
	for _, part := range parts {
		if idx, err := strconv.Atoi(part); err == nil {
			out = append(out, cty.IndexStep{Key: cty.NumberIntVal(int64(idx))})
		} else {
			out = append(out, cty.GetAttrStep{Name: part})
		}
	}
	return out
}

// lookupCtyPath walks v along path; (NilVal, false) on any null/unknown or missing step.
func lookupCtyPath(v cty.Value, path cty.Path) (cty.Value, bool) {
	current := v
	for _, step := range path {
		if current.IsNull() || !current.IsKnown() {
			return cty.NilVal, false
		}
		switch s := step.(type) {
		case cty.GetAttrStep:
			if !current.Type().IsObjectType() || !current.Type().HasAttribute(s.Name) {
				return cty.NilVal, false
			}
			current = current.GetAttr(s.Name)
		case cty.IndexStep:
			t := current.Type()
			if !t.IsListType() && !t.IsTupleType() && !t.IsSetType() {
				return cty.NilVal, false
			}
			if current.LengthInt() == 0 {
				return cty.NilVal, false
			}
			idx, _ := s.Key.AsBigFloat().Int64()
			if idx < 0 || idx >= int64(current.LengthInt()) {
				return cty.NilVal, false
			}
			current = current.Index(s.Key)
		default:
			return cty.NilVal, false
		}
	}
	return current, true
}
