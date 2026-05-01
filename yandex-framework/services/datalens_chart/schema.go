package datalens_chart

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	"github.com/yandex-cloud/terraform-provider-yandex/common/defaultschema"
)

func ResourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages a DataLens chart resource. Supports `wizard` and `ql` chart " +
			"types via the `type` discriminator. The chart payload mirrors the DataLens API " +
			"(`annotation { description }` and `data { ... }` blocks). " +
			"For more information, see [the official documentation](https://yandex.cloud/ru/docs/datalens/operations/api-start). " +
			"DataLens chart endpoints are marked as Experimental in the upstream API.",
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
			"type": schema.StringAttribute{
				MarkdownDescription: "Chart variant. One of `wizard`, `ql`. " +
					"Inferred from the presence of `data.wizard` or `data.ql` if not set.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"workbook_id": schema.StringAttribute{
				MarkdownDescription: "The workbook ID where the chart will be created. " +
					"Changing this attribute forces recreation of the resource.",
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: common.ResourceDescriptions["name"] +
					" Changing this attribute forces recreation of the resource.",
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			"annotation": annotationAttribute(),
			"data":       dataAttribute(),

			"created_at": defaultschema.CreatedAt(),
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "Last update timestamp.",
				Computed:            true,
			},
			"revision_id":  schema.StringAttribute{Computed: true, MarkdownDescription: "Current revision ID."},
			"saved_id":     schema.StringAttribute{Computed: true, MarkdownDescription: "Saved revision ID."},
			"published_id": schema.StringAttribute{Computed: true, MarkdownDescription: "Published revision ID."},
		},
	}
}

func annotationAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Chart annotation block. Carries human-readable metadata such as the description.",
		Optional:            true,
		// API does not echo annotation back on get*Chart, so a schema Default
		// would diverge between Apply state and Import state. Elision hook
		// in populate.go handles the symmetry instead.
		Attributes: map[string]schema.Attribute{
			"description": schema.StringAttribute{
				MarkdownDescription: common.ResourceDescriptions["description"],
				Optional:            true,
			},
		},
	}
}

func dataAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Chart `data` payload (renderer + placeholders + variant-specific config).",
		Required:            true,
		Attributes: map[string]schema.Attribute{
			"version": schema.StringAttribute{
				MarkdownDescription: "DataLens chart payload schema version. Defaults to `15` (wizard) / `7` (ql).",
				Optional:            true,
				Computed:            true,
			},

			"visualization":  visualizationAttribute(),
			"extra_settings": extraSettingsAttribute(),

			"colors":      fieldRefListAttribute("Color encoding fields."),
			"labels":      fieldRefListAttribute("Label fields."),
			"shapes":      fieldRefListAttribute("Shape encoding fields."),
			"tooltips":    fieldRefListAttribute("Tooltip fields."),
			"filters":     fieldRefListAttribute("Filter fields."),
			"sort":        fieldRefListAttribute("Sort fields."),
			"hierarchies": fieldRefListAttribute("Hierarchy field groupings."),
			"segments":    fieldRefListAttribute("Segment fields."),
			"updates":     fieldRefListAttribute("Live-update fields."),

			"links": linksAttribute(),

			"wizard": wizardAttribute(),
			"ql":     qlAttribute(),

			"colors_config": schema.MapAttribute{
				MarkdownDescription: "Server-required `colorsConfig` map. Defaults to empty.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Default:             mapdefault.StaticValue(types.MapValueMust(types.StringType, map[string]attr.Value{})),
			},
			"shapes_config": schema.MapAttribute{
				MarkdownDescription: "Server-required `shapesConfig` map. Defaults to empty.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Default:             mapdefault.StaticValue(types.MapValueMust(types.StringType, map[string]attr.Value{})),
			},
			"geopoints_config": schema.MapAttribute{
				MarkdownDescription: "Server-required `geopointsConfig` map. Defaults to empty.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				Default:             mapdefault.StaticValue(types.MapValueMust(types.StringType, map[string]attr.Value{})),
			},
		},
	}
}

func linksAttribute() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "Cross-dataset relations (used in multi-dataset wizard charts).",
		Optional:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"id": schema.StringAttribute{Required: true, MarkdownDescription: "Link identifier."},
				"fields": schema.ListNestedAttribute{
					Required:            true,
					MarkdownDescription: "Linked field per dataset.",
					NestedObject: schema.NestedAttributeObject{
						Attributes: map[string]schema.Attribute{
							"dataset_id": schema.StringAttribute{Required: true, MarkdownDescription: "Dataset ID."},
							"field": schema.MapAttribute{
								Optional:            true,
								ElementType:         types.StringType,
								MarkdownDescription: "Linked-field metadata as a free-form string map (typically `{guid, title, calc_mode, data_type}`).",
							},
						},
					},
				},
			},
		},
	}
}

func visualizationAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Chart rendering configuration. `id` selects the renderer.",
		Required:            true,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Visualization renderer (e.g. `line`, `flatTable`, `metric`, `bar`, `area`, `scatter`).",
			},
			"type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Higher-level visualization category (`table`, `metric`, `chart`, ...).",
			},
			"placeholders": schema.ListNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Placeholder slots for fields.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":       schema.StringAttribute{Required: true},
						"type":     schema.StringAttribute{Optional: true, Computed: true},
						"title":    schema.StringAttribute{Optional: true, Computed: true},
						"required": schema.BoolAttribute{Optional: true, Computed: true},
						"capacity": schema.Int64Attribute{Optional: true, Computed: true},
						"items":    fieldRefListAttribute("Fields placed into this placeholder."),
						"settings": schema.MapAttribute{Optional: true, Computed: true, ElementType: types.StringType},
					},
				},
			},
		},
	}
}

func extraSettingsAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Optional cosmetic and behavioral settings (title, legend, pagination, navigator, etc.).",
		Optional:            true,
		// API does not echo extra_settings back; elision hook in populate.go
		// keeps Apply and Import states symmetric.
		Attributes: map[string]schema.Attribute{
			"title":                schema.StringAttribute{Optional: true},
			"title_mode":           schema.StringAttribute{Optional: true},
			"indicator_title_mode": schema.StringAttribute{Optional: true},
			"legend_mode":          schema.StringAttribute{Optional: true},
			"pivot_inline_sort":    schema.StringAttribute{Optional: true},
			"stacking":             schema.StringAttribute{Optional: true},
			"tooltip_sum":          schema.StringAttribute{Optional: true},
			"feed":                 schema.StringAttribute{Optional: true},
			"pagination":           schema.StringAttribute{Optional: true},
			"limit":                schema.Int64Attribute{Optional: true},
			"navigator_settings": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"is_navigator_available": schema.BoolAttribute{Optional: true},
					"selected_lines":         schema.ListAttribute{Optional: true, ElementType: types.StringType},
				},
			},
		},
	}
}

func wizardAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Wizard-specific configuration. Required when `type = \"wizard\"`.",
		Optional:            true,
		Attributes: map[string]schema.Attribute{
			"datasets_ids": schema.ListAttribute{
				Required:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Dataset IDs the wizard chart pulls fields from.",
			},
			"datasets_partial_fields": schema.ListNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Per-dataset field references picked from the datasets. Outer index aligns with `datasets_ids`.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"fields": schema.ListNestedAttribute{
							Required: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"guid":      schema.StringAttribute{Required: true},
									"title":     schema.StringAttribute{Optional: true},
									"calc_mode": schema.StringAttribute{Optional: true},
								},
							},
						},
					},
				},
			},
			"convert": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether the chart was converted from a previous schema version.",
			},
		},
	}
}

func qlAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "QL-specific configuration. Required when `type = \"ql\"`.",
		Optional:            true,
		Attributes: map[string]schema.Attribute{
			"chart_type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("sql"),
				MarkdownDescription: "QL chart sub-type. Defaults to `sql`.",
			},
			"connection": schema.SingleNestedAttribute{
				Required:            true,
				MarkdownDescription: "Connection that the SQL query runs against.",
				Attributes: map[string]schema.Attribute{
					"entry_id": schema.StringAttribute{Required: true},
					"type":     schema.StringAttribute{Required: true},
				},
			},
			"query_value": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The raw SQL query text. Use `{{param_name}}` placeholders to reference `params`.",
			},
			"queries": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"value":  schema.StringAttribute{Required: true},
						"hidden": schema.BoolAttribute{Optional: true},
						"params": qlParamsListAttribute(),
					},
				},
			},
			"params": qlParamsListAttribute(),
			"order":  schema.StringAttribute{Optional: true, MarkdownDescription: "Optional row ordering directive."},
		},
	}
}

func qlParamsListAttribute() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		Optional: true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"name":          schema.StringAttribute{Required: true},
				"type":          schema.StringAttribute{Required: true},
				"default_value": schema.StringAttribute{Optional: true},
			},
		},
	}
}

func fieldRefListAttribute(desc string) schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: desc,
		Optional:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"id":                schema.StringAttribute{Optional: true, Computed: true},
				"guid":              schema.StringAttribute{Required: true},
				"title":             schema.StringAttribute{Required: true},
				"dataset_id":        schema.StringAttribute{Optional: true},
				"type":              schema.StringAttribute{Required: true},
				"data_type":         schema.StringAttribute{Required: true},
				"initial_data_type": schema.StringAttribute{Optional: true, Computed: true},
				"cast":              schema.StringAttribute{Optional: true, Computed: true},
				"calc_mode":         schema.StringAttribute{Optional: true, Computed: true},
				"aggregation":       schema.StringAttribute{Optional: true, Computed: true},
				"source":            schema.StringAttribute{Optional: true, Computed: true},
				"formula":           schema.StringAttribute{Optional: true, Computed: true},
				"guid_formula":      schema.StringAttribute{Optional: true, Computed: true},
				"description":       schema.StringAttribute{Optional: true, Computed: true},
				"hidden":            schema.BoolAttribute{Optional: true, Computed: true},
				"managed_by":        schema.StringAttribute{Optional: true, Computed: true},
				"avatar_id":         schema.StringAttribute{Optional: true, Computed: true},
				"ui_settings":       schema.StringAttribute{Optional: true, Computed: true},
			},
		},
	}
}

