package datalens_dashboard

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	"github.com/yandex-cloud/terraform-provider-yandex/common/defaultschema"
)

func ResourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages a DataLens dashboard resource. The payload mirrors the DataLens API: " +
			"`entry { annotation, meta, data }` carries the dashboard content. " +
			"For more information, see [the official documentation](https://yandex.cloud/ru/docs/datalens/operations/api-start). " +
			"DataLens dashboard endpoints are marked as Experimental in the upstream API.",
		Attributes: map[string]schema.Attribute{
			"id": defaultschema.Id(),
			"organization_id": schema.StringAttribute{
				MarkdownDescription: "The organization ID for the DataLens instance. " +
					"If not specified, the provider-level `organization_id` is used.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},

			"entry": entryAttribute(),
		},
	}
}

func entryAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Dashboard `entry` block. Carries the typed dashboard payload " +
			"(`annotation { description }`, `meta { title, locale }`, `data { ... }`) " +
			"plus the entry-level identifiers (name, workbook_id, key) and the " +
			"computed timestamps and revision IDs.",
		Required: true,
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: common.ResourceDescriptions["name"] +
					" Changing this attribute forces recreation of the resource.",
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"workbook_id": schema.StringAttribute{
				MarkdownDescription: "The workbook ID where the dashboard will be created. " +
					"Changing this attribute forces recreation of the resource.",
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created_at":   defaultschema.CreatedAt(),
			"updated_at":   schema.StringAttribute{Computed: true, MarkdownDescription: "Last update timestamp."},
			"revision_id":  schema.StringAttribute{Computed: true, MarkdownDescription: "Current revision ID."},
			"saved_id":     schema.StringAttribute{Computed: true, MarkdownDescription: "Saved revision ID."},
			"published_id": schema.StringAttribute{Computed: true, MarkdownDescription: "Published revision ID."},

			"annotation": schema.SingleNestedAttribute{
				MarkdownDescription: "Annotation block (description).",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"description": schema.StringAttribute{
						MarkdownDescription: common.ResourceDescriptions["description"],
						Optional:            true,
					},
				},
			},
			"meta": schema.SingleNestedAttribute{
				MarkdownDescription: "Meta block (title, locale).",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"title":  schema.StringAttribute{Optional: true, MarkdownDescription: "Dashboard meta title."},
					"locale": schema.StringAttribute{Optional: true, MarkdownDescription: "Dashboard meta locale."},
				},
			},
			"data": dataAttribute(),
		},
	}
}

func dataAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Dashboard `data` payload (counter, salt, settings, tabs).",
		Required:            true,
		Attributes: map[string]schema.Attribute{
			"counter": schema.Int64Attribute{
				MarkdownDescription: "Internal counter used by DataLens for revisioning. Defaults to `1`.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(1),
			},
			"salt": schema.StringAttribute{
				MarkdownDescription: "Random salt string used by DataLens for revisioning.",
				Required:            true,
			},
			"scheme_version": schema.Int64Attribute{
				MarkdownDescription: "Dashboard data scheme version. Currently fixed at `8`.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(8),
			},
			"access_description":  schema.StringAttribute{Optional: true, MarkdownDescription: "Access policy description (markdown)."},
			"support_description": schema.StringAttribute{Optional: true, MarkdownDescription: "Support contact information (markdown)."},

			"settings": settingsAttribute(),
			"tabs":     tabsAttribute(),
		},
	}
}

func settingsAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Dashboard runtime settings (auto-update, concurrency, navigation behavior).",
		Required:            true,
		Attributes: map[string]schema.Attribute{
			"autoupdate_interval":      schema.Int64Attribute{Optional: true, MarkdownDescription: "Auto-update interval in seconds (>= 30) or null to disable."},
			"max_concurrent_requests":  schema.Int64Attribute{Optional: true, MarkdownDescription: "Maximum concurrent chart load requests (>= 1) or null for default."},
			"silent_loading": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether to suppress loading spinners.",
			},
			"dependent_selectors": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether selector controls cascade their values.",
			},
			"expand_toc": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether the table of contents starts expanded.",
			},
			"hide_dash_title":          schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Whether to hide the dashboard title."},
			"hide_tabs":                schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Whether to hide the tab bar."},
			"load_only_visible_charts": schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Lazy-load charts as the user scrolls."},
			"load_priority":            schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Load priority strategy (`charts` or `selectors`)."},
			"global_params":            schema.MapAttribute{Optional: true, ElementType: types.StringType, MarkdownDescription: "Dashboard-wide default parameter values applied to all controls."},
		},
	}
}

func tabsAttribute() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "Dashboard tabs. At least one tab is required.",
		Required:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"id":    schema.StringAttribute{Required: true, MarkdownDescription: "Tab identifier (referenced by item ids in `layout`)."},
				"title": schema.StringAttribute{Required: true, MarkdownDescription: "Tab display title."},
				"items": tabItemsAttribute(),
				"layout": schema.ListNestedAttribute{
					Optional:            true,
					MarkdownDescription: "Layout positions for items on the tab grid.",
					NestedObject: schema.NestedAttributeObject{
						Attributes: map[string]schema.Attribute{
							"i":      schema.StringAttribute{Required: true, MarkdownDescription: "References item id."},
							"x":      schema.Int64Attribute{Required: true, MarkdownDescription: "Grid X position."},
							"y":      schema.Int64Attribute{Required: true, MarkdownDescription: "Grid Y position."},
							"w":      schema.Int64Attribute{Required: true, MarkdownDescription: "Grid width."},
							"h":      schema.Int64Attribute{Required: true, MarkdownDescription: "Grid height."},
							"parent": schema.StringAttribute{Optional: true, MarkdownDescription: "Optional parent item id (for nested layout)."},
						},
					},
				},
				"connections": schema.ListNestedAttribute{
					Optional:            true,
					MarkdownDescription: "Per-tab item connections (control ↔ widget links).",
					NestedObject: schema.NestedAttributeObject{
						Attributes: map[string]schema.Attribute{
							"from": schema.StringAttribute{Required: true},
							"to":   schema.StringAttribute{Required: true},
							"kind": schema.StringAttribute{Required: true, MarkdownDescription: "Connection kind. Typically `ignore`."},
						},
					},
				},
				"aliases": schema.MapAttribute{
					Optional:            true,
					Computed:            true,
					ElementType:         types.ListType{ElemType: types.ListType{ElemType: types.StringType}},
					MarkdownDescription: "Field aliases grouped by namespace (e.g. `{ default = [[\"a\",\"b\"]] }`).",
					// API echoes an empty map for tabs without aliases; declare
					// the same Default so plan-time matches state.
					Default: mapdefault.StaticValue(types.MapValueMust(
						types.ListType{ElemType: types.ListType{ElemType: types.StringType}},
						map[string]attr.Value{},
					)),
				},
			},
		},
	}
}

func tabItemsAttribute() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		Required:            true,
		MarkdownDescription: "Items placed on the tab. Exactly one of `widget`/`group_control`/`text`/`title`/`image` selects the item kind.",
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"id": schema.StringAttribute{Required: true, MarkdownDescription: "Item identifier."},
				"namespace": schema.StringAttribute{
					Optional:            true,
					Computed:            true,
					MarkdownDescription: "Item namespace. Defaults to `default`.",
				},

				"widget":        widgetItemAttribute(),
				"group_control": groupControlItemAttribute(),
				"text":          textItemAttribute(),
				"title":         titleItemAttribute(),
				"image":         imageItemAttribute(),
			},
		},
	}
}

func widgetItemAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Widget item — wraps one or more chart references with shared layout.",
		Optional:            true,
		Attributes: map[string]schema.Attribute{
			"hide_title": schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Hide the widget header."},
			"tabs": schema.ListNestedAttribute{
				Required:            true,
				MarkdownDescription: "Chart references (multi-tab widget shows multiple charts in one frame).",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":          schema.StringAttribute{Required: true},
						"title":       schema.StringAttribute{Required: true},
						"description": schema.StringAttribute{Optional: true, Computed: true},
						"chart_id":    schema.StringAttribute{Required: true, MarkdownDescription: "Chart entry ID to embed."},
						"is_default":  schema.BoolAttribute{Optional: true, Computed: true},
						"auto_height": schema.BoolAttribute{Optional: true, Computed: true},
						"params":      schema.MapAttribute{Optional: true, Computed: true, ElementType: types.StringType},
					},
				},
			},
		},
	}
}

func groupControlItemAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Group of dashboard controls (selectors).",
		Optional:            true,
		Attributes: map[string]schema.Attribute{
			"auto_height":  schema.BoolAttribute{Optional: true, Computed: true},
			"button_apply": schema.BoolAttribute{Optional: true, Computed: true},
			"button_reset": schema.BoolAttribute{Optional: true, Computed: true},
			"group": schema.ListNestedAttribute{
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":             schema.StringAttribute{Required: true},
						"namespace":      schema.StringAttribute{Optional: true, Computed: true},
						"placement_mode": schema.StringAttribute{Optional: true, Computed: true},
						"defaults":       schema.MapAttribute{Optional: true, Computed: true, ElementType: types.ListType{ElemType: types.StringType}},
						"source": schema.SingleNestedAttribute{
							Required: true,
							Attributes: map[string]schema.Attribute{
								"acceptable_values": schema.ListNestedAttribute{
									Optional: true,
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"title": schema.StringAttribute{Required: true},
											"value": schema.StringAttribute{Required: true},
										},
									},
								},
								"accent_type":      schema.StringAttribute{Optional: true, Computed: true},
								"default_value":    schema.ListAttribute{Optional: true, ElementType: types.StringType},
								"element_type":     schema.StringAttribute{Required: true, MarkdownDescription: "Control type (e.g. `select`, `input`, `date`)."},
								"field_name":       schema.StringAttribute{Required: true, MarkdownDescription: "Underlying parameter name."},
								"hint":             schema.StringAttribute{Optional: true, Computed: true},
								"multiselectable":  schema.BoolAttribute{Optional: true, Computed: true},
								"required":         schema.BoolAttribute{Optional: true, Computed: true},
								"show_hint":        schema.BoolAttribute{Optional: true, Computed: true},
								"title":            schema.StringAttribute{Required: true, MarkdownDescription: "Visible label of the control."},
							},
						},
					},
				},
			},
		},
	}
}

func textItemAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Markdown text item.",
		Optional:            true,
		Attributes: map[string]schema.Attribute{
			"text": schema.StringAttribute{Required: true, MarkdownDescription: "Markdown body."},
		},
	}
}

func titleItemAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Heading/title item.",
		Optional:            true,
		Attributes: map[string]schema.Attribute{
			"text":        schema.StringAttribute{Required: true},
			"size":        schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Heading size (`l`, `m`, `s`)."},
			"show_in_toc": schema.BoolAttribute{Optional: true, Computed: true},
		},
	}
}

func imageItemAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Image item.",
		Optional:            true,
		Attributes: map[string]schema.Attribute{
			"src":         schema.StringAttribute{Required: true, MarkdownDescription: "Image URL."},
			"alt_text":    schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Accessibility alternative text."},
			"description": schema.StringAttribute{Optional: true, Computed: true},
		},
	}
}
