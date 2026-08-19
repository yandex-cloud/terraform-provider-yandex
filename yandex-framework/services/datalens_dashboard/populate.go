package datalens_dashboard

import (
	"fmt"
	"strings"

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

// itemTypeAndVariant returns the discriminator and variant struct pointer
// for a tab item, looked up by which sub-block is set. Returns "", nil when
// no variant is populated.
func itemTypeAndVariant(item *dashboardTabItemModel) (string, any) {
	switch {
	case item.Widget != nil:
		return "widget", item.Widget
	case item.GroupCtl != nil:
		return "group_control", item.GroupCtl
	case item.Text != nil:
		return "text", item.Text
	case item.Title != nil:
		return "title", item.Title
	case item.Image != nil:
		return "image", item.Image
	}
	return "", nil
}

// inflateItemVariant allocates the variant struct matching the API-supplied
// type and fills it from the item's flat `data` map via wire.Unmarshal.
// Other variant pointers are cleared so a server-side type change can't
// leave stale residue from a previous state.
func inflateItemVariant(item *dashboardTabItemModel, itemType string, data map[string]interface{}) error {
	item.Widget = nil
	item.GroupCtl = nil
	item.Text = nil
	item.Title = nil
	item.Image = nil
	switch itemType {
	case "widget":
		item.Widget = &dashboardWidgetItemModel{}
		return wire.Unmarshal(data, item.Widget)
	case "group_control":
		item.GroupCtl = &dashboardGroupCtlItemModel{}
		return wire.Unmarshal(data, item.GroupCtl)
	case "text":
		item.Text = &dashboardTextItemModel{}
		return wire.Unmarshal(data, item.Text)
	case "title":
		item.Title = &dashboardTitleItemModel{}
		return wire.Unmarshal(data, item.Title)
	case "image":
		item.Image = &dashboardImageItemModel{}
		return wire.Unmarshal(data, item.Image)
	}
	return nil
}

// marshalDashboard builds the wire body. Variant blocks (Widget/GroupCtl/
// Text/Title/Image) are tagged `wire:"-"`; for each tab item we marshal the
// populated variant, inject its discriminator into `item["type"]`, and put
// the payload under `item["data"]` — the shape DataLens expects. Empty
// containers DataLens validates (`tabs[].items|layout|connections|aliases`)
// are emitted via wire `,alwaysEmit` — see models.go.
func marshalDashboard(plan *dashboardModel) (map[string]any, error) {
	if plan.Entry == nil {
		return nil, fmt.Errorf("`entry` block is required")
	}
	body, err := wire.Marshal(plan)
	if err != nil {
		return nil, err
	}
	if plan.Entry.Data == nil {
		return body, nil
	}

	entry, _ := body["entry"].(map[string]any)
	if entry == nil {
		return body, nil
	}
	data, _ := entry["data"].(map[string]any)
	if data == nil {
		return body, nil
	}
	tabsRaw, _ := data["tabs"].([]any)
	for ti := range plan.Entry.Data.Tabs {
		if ti >= len(tabsRaw) {
			break
		}
		tabMap, _ := tabsRaw[ti].(map[string]any)
		if tabMap == nil {
			continue
		}
		itemsRaw, _ := tabMap["items"].([]any)
		for ii := range plan.Entry.Data.Tabs[ti].Items {
			if ii >= len(itemsRaw) {
				break
			}
			itemMap, _ := itemsRaw[ii].(map[string]any)
			if itemMap == nil {
				continue
			}
			itemType, variant := itemTypeAndVariant(&plan.Entry.Data.Tabs[ti].Items[ii])
			if variant == nil {
				itemMap["data"] = map[string]any{}
				continue
			}
			itemMap["type"] = itemType
			vb, err := wire.Marshal(variant)
			if err != nil {
				return nil, fmt.Errorf("tabs[%d].items[%d]: %w", ti, ii, err)
			}
			itemMap["data"] = vb
		}
	}
	return body, nil
}

// unmarshalDashboardResponse fills the typed model from a getDashboard
// response. Common fields are populated by a single wire.Unmarshal; per-item
// variant structs are then filled from each item's `data` map by a second
// wire.Unmarshal dispatched on `item.type`.
//
// We pre-process the raw response to bridge a few API quirks:
//   - `entry.entryId` → top-level `id` (where Terraform import expects it).
//   - `entry.name` is not echoed; derive it from `entry.key`.
//   - `entry.annotation/meta` echo all-null payloads even when unconfigured;
//     drop them so config-null matches state-null.
func unmarshalDashboardResponse(model *dashboardModel, resp map[string]interface{}) error {
	if entry, ok := resp["entry"].(map[string]interface{}); ok {
		if v, ok := entry["entryId"].(string); ok && v != "" {
			resp["id"] = v
		}
		if _, ok := entry["name"].(string); !ok {
			if k, ok := entry["key"].(string); ok && k != "" {
				entry["name"] = leafName(k)
			}
		}
		if a, ok := entry["annotation"].(map[string]interface{}); ok {
			if a["description"] == nil {
				delete(entry, "annotation")
			}
		}
		if m, ok := entry["meta"].(map[string]interface{}); ok {
			if m["title"] == nil && m["locale"] == nil {
				delete(entry, "meta")
			}
		}
	}

	if err := wire.Unmarshal(resp, model); err != nil {
		return fmt.Errorf("dashboard: %w", err)
	}

	// Dispatch each tab item's variant from its raw `data` map.
	entry, _ := resp["entry"].(map[string]interface{})
	if model.Entry == nil || model.Entry.Data == nil || entry == nil {
		return nil
	}
	rawData, _ := entry["data"].(map[string]interface{})
	if rawData == nil {
		return nil
	}
	rawTabs, _ := rawData["tabs"].([]interface{})
	for ti := range model.Entry.Data.Tabs {
		if ti >= len(rawTabs) {
			break
		}
		tabRaw, _ := rawTabs[ti].(map[string]interface{})
		if tabRaw == nil {
			continue
		}
		rawItems, _ := tabRaw["items"].([]interface{})
		for ii := range model.Entry.Data.Tabs[ti].Items {
			if ii >= len(rawItems) {
				break
			}
			itemRaw, _ := rawItems[ii].(map[string]interface{})
			if itemRaw == nil {
				continue
			}
			itemType, _ := itemRaw["type"].(string)
			itemData, _ := itemRaw["data"].(map[string]interface{})
			if itemData == nil {
				continue
			}
			if err := inflateItemVariant(&model.Entry.Data.Tabs[ti].Items[ii], itemType, itemData); err != nil {
				return fmt.Errorf("tabs[%d].items[%d]: %w", ti, ii, err)
			}
		}
	}
	return nil
}
