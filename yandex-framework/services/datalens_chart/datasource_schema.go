package datalens_chart

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// chartDataSourceSchema mirrors the resource schema (`annotation { description }`,
// `data { ... }` blocks) with everything Computed.
func chartDataSourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Retrieves information about a DataLens chart.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Required: true, MarkdownDescription: "The ID of the chart."},
			"type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The chart type. One of `wizard`, `ql`.",
				Validators:          []validator.String{stringvalidator.OneOf("wizard", "ql")},
			},
			"organization_id": schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "The organization ID."},
			"workbook_id":     schema.StringAttribute{Computed: true, MarkdownDescription: "The workbook ID."},
			"name":            schema.StringAttribute{Computed: true, MarkdownDescription: "The chart name."},
			"created_at":      schema.StringAttribute{Computed: true, MarkdownDescription: "Creation timestamp."},
			"updated_at":      schema.StringAttribute{Computed: true, MarkdownDescription: "Last update timestamp."},
			"revision_id":     schema.StringAttribute{Computed: true, MarkdownDescription: "Current revision ID."},
			"saved_id":        schema.StringAttribute{Computed: true, MarkdownDescription: "Saved revision ID."},
			"published_id":    schema.StringAttribute{Computed: true, MarkdownDescription: "Published revision ID."},

			"annotation": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"description": schema.StringAttribute{Computed: true},
				},
			},
			"data": dsDataAttribute(),
		},
	}
}

func dsDataAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Computed: true,
		Attributes: map[string]schema.Attribute{
			"version":          schema.StringAttribute{Computed: true},
			"visualization":    dsVisualizationAttribute(),
			"extra_settings":   dsExtraSettingsAttribute(),
			"colors":           dsFieldRefList(),
			"labels":           dsFieldRefList(),
			"shapes":           dsFieldRefList(),
			"tooltips":         dsFieldRefList(),
			"filters":          dsFieldRefList(),
			"sort":             dsFieldRefList(),
			"hierarchies":      dsFieldRefList(),
			"segments":         dsFieldRefList(),
			"updates":          dsFieldRefList(),
			"links":            dsLinksAttribute(),
			"wizard":           dsWizardAttribute(),
			"ql":               dsQLAttribute(),
			"colors_config":    schema.MapAttribute{Computed: true, ElementType: types.StringType},
			"shapes_config":    schema.MapAttribute{Computed: true, ElementType: types.StringType},
			"geopoints_config": schema.MapAttribute{Computed: true, ElementType: types.StringType},
		},
	}
}

func dsFieldRefList() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		Computed: true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"id":                schema.StringAttribute{Computed: true},
				"guid":              schema.StringAttribute{Computed: true},
				"title":             schema.StringAttribute{Computed: true},
				"dataset_id":        schema.StringAttribute{Computed: true},
				"type":              schema.StringAttribute{Computed: true},
				"data_type":         schema.StringAttribute{Computed: true},
				"initial_data_type": schema.StringAttribute{Computed: true},
				"cast":              schema.StringAttribute{Computed: true},
				"calc_mode":         schema.StringAttribute{Computed: true},
				"aggregation":       schema.StringAttribute{Computed: true},
				"source":            schema.StringAttribute{Computed: true},
				"formula":           schema.StringAttribute{Computed: true},
				"guid_formula":      schema.StringAttribute{Computed: true},
				"description":       schema.StringAttribute{Computed: true},
				"hidden":            schema.BoolAttribute{Computed: true},
				"managed_by":        schema.StringAttribute{Computed: true},
				"avatar_id":         schema.StringAttribute{Computed: true},
				"ui_settings":       schema.StringAttribute{Computed: true},
			},
		},
	}
}

func dsVisualizationAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Computed: true,
		Attributes: map[string]schema.Attribute{
			"id":   schema.StringAttribute{Computed: true},
			"type": schema.StringAttribute{Computed: true},
			"placeholders": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":       schema.StringAttribute{Computed: true},
						"type":     schema.StringAttribute{Computed: true},
						"title":    schema.StringAttribute{Computed: true},
						"required": schema.BoolAttribute{Computed: true},
						"capacity": schema.Int64Attribute{Computed: true},
						"items":    dsFieldRefList(),
						"settings": schema.MapAttribute{Computed: true, ElementType: types.StringType},
					},
				},
			},
		},
	}
}

func dsExtraSettingsAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Computed: true,
		Attributes: map[string]schema.Attribute{
			"title":                schema.StringAttribute{Computed: true},
			"title_mode":           schema.StringAttribute{Computed: true},
			"indicator_title_mode": schema.StringAttribute{Computed: true},
			"legend_mode":          schema.StringAttribute{Computed: true},
			"pivot_inline_sort":    schema.StringAttribute{Computed: true},
			"stacking":             schema.StringAttribute{Computed: true},
			"tooltip_sum":          schema.StringAttribute{Computed: true},
			"feed":                 schema.StringAttribute{Computed: true},
			"pagination":           schema.StringAttribute{Computed: true},
			"limit":                schema.Int64Attribute{Computed: true},
			"navigator_settings": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"is_navigator_available": schema.BoolAttribute{Computed: true},
					"selected_lines":         schema.ListAttribute{Computed: true, ElementType: types.StringType},
				},
			},
		},
	}
}

func dsLinksAttribute() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		Computed: true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"id": schema.StringAttribute{Computed: true},
				"fields": schema.ListNestedAttribute{
					Computed: true,
					NestedObject: schema.NestedAttributeObject{
						Attributes: map[string]schema.Attribute{
							"dataset_id": schema.StringAttribute{Computed: true},
							"field":      schema.MapAttribute{Computed: true, ElementType: types.StringType},
						},
					},
				},
			},
		},
	}
}

func dsWizardAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Computed: true,
		Attributes: map[string]schema.Attribute{
			"datasets_ids": schema.ListAttribute{Computed: true, ElementType: types.StringType},
			"datasets_partial_fields": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"fields": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"guid":      schema.StringAttribute{Computed: true},
									"title":     schema.StringAttribute{Computed: true},
									"calc_mode": schema.StringAttribute{Computed: true},
								},
							},
						},
					},
				},
			},
			"convert": schema.BoolAttribute{Computed: true},
		},
	}
}

func dsQLAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Computed: true,
		Attributes: map[string]schema.Attribute{
			"chart_type": schema.StringAttribute{Computed: true},
			"connection": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"entry_id": schema.StringAttribute{Computed: true},
					"type":     schema.StringAttribute{Computed: true},
				},
			},
			"query_value": schema.StringAttribute{Computed: true},
			"queries": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"value":  schema.StringAttribute{Computed: true},
						"hidden": schema.BoolAttribute{Computed: true},
						"params": dsQLParamsList(),
					},
				},
			},
			"params": dsQLParamsList(),
			"order":  schema.StringAttribute{Computed: true},
		},
	}
}

func dsQLParamsList() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		Computed: true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"name":          schema.StringAttribute{Computed: true},
				"type":          schema.StringAttribute{Computed: true},
				"default_value": schema.StringAttribute{Computed: true},
			},
		},
	}
}
