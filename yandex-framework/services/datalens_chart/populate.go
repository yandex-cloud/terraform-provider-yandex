package datalens_chart

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datalens/wire"
)

// leafName returns the last `/`-separated segment of a DataLens entry key.
// DataLens echoes `key` but not `name`, so we derive the latter for reads
// that start from an id-only state (Import, datasource).
func leafName(key string) string {
	if i := strings.LastIndex(key, "/"); i >= 0 && i+1 < len(key) {
		return key[i+1:]
	}
	return key
}

// chartVariantType returns the variant discriminator ("wizard"/"ql") inferred
// from which sub-block on chartDataModel is set. Returns "" if neither is set.
func chartVariantType(plan *chartModel) string {
	if plan.Data == nil {
		return ""
	}
	switch {
	case plan.Data.Wizard != nil:
		return "wizard"
	case plan.Data.Ql != nil:
		return "ql"
	}
	return ""
}

// marshalChart builds the wire body. Variant blocks (Wizard/Ql) are tagged
// `wire:"-"`, so we serialize whichever one is set and merge it flat into
// `body["data"]` next to the common fields — that's how DataLens stores it.
//
// Type and data.version are inferred from the variant when not set explicitly:
// type from sub-block presence, version from a per-variant default
// (`15` for wizard, `7` for ql).
func marshalChart(plan *chartModel) (map[string]any, error) {
	if plan.Type.IsNull() || plan.Type.IsUnknown() {
		if t := chartVariantType(plan); t != "" {
			plan.Type = types.StringValue(t)
		}
	}
	if plan.Data != nil && (plan.Data.Version.IsNull() || plan.Data.Version.IsUnknown()) {
		switch plan.Type.ValueString() {
		case "wizard":
			plan.Data.Version = types.StringValue("15")
		case "ql":
			plan.Data.Version = types.StringValue("7")
		}
	}

	body, err := wire.Marshal(plan)
	if err != nil {
		return nil, err
	}
	if plan.Data == nil {
		return body, nil
	}
	data, _ := body["data"].(map[string]any)
	if data == nil {
		return body, nil
	}
	var variant any
	switch plan.Type.ValueString() {
	case "wizard":
		variant = plan.Data.Wizard
	case "ql":
		variant = plan.Data.Ql
	}
	if variant == nil {
		return body, nil
	}
	vb, err := wire.Marshal(variant)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", plan.Type.ValueString(), err)
	}
	for k, v := range vb {
		data[k] = v
	}
	return body, nil
}

// unmarshalChartResponse fills the typed chart model from a getXxxChart
// response. Common fields are populated by a single wire.Unmarshal; the
// variant struct is then filled from the same flat `data` map by a second
// wire.Unmarshal dispatched on `data.type`.
//
// The DataLens API echoes `key` (the entry path) but not `name`, so we
// pre-inject `name = leafName(key)` into the response before unmarshal —
// no need to keep `key` on the model.
func unmarshalChartResponse(model *chartModel, resp map[string]interface{}) error {
	if _, ok := resp["name"].(string); !ok {
		if k, ok := resp["key"].(string); ok && k != "" {
			resp["name"] = leafName(k)
		}
	}

	if err := wire.Unmarshal(resp, model); err != nil {
		return fmt.Errorf("chart: %w", err)
	}
	data, _ := resp["data"].(map[string]interface{})
	if data != nil {
		// Discover discriminator (model.Type may be empty on Import).
		if model.Type.IsNull() || model.Type.ValueString() == "" {
			if t, ok := data["type"].(string); ok {
				model.Type = types.StringValue(t)
			}
		}
		if model.Data == nil {
			model.Data = &chartDataModel{}
		}
		switch model.Type.ValueString() {
		case "wizard":
			model.Data.Wizard = &chartWizardModel{}
			if err := wire.Unmarshal(data, model.Data.Wizard); err != nil {
				return fmt.Errorf("wizard: %w", err)
			}
		case "ql":
			model.Data.Ql = &chartQLModel{}
			if err := wire.Unmarshal(data, model.Data.Ql); err != nil {
				return fmt.Errorf("ql: %w", err)
			}
		}
	}

	return nil
}
