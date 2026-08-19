package datalens_dashboard

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// dashboardDataSourceSchema mirrors the resource schema with everything Computed.
func dashboardDataSourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Retrieves information about a DataLens dashboard.",
		Attributes: map[string]schema.Attribute{
			"id":              schema.StringAttribute{Required: true, MarkdownDescription: "The ID of the dashboard."},
			"organization_id": schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "The organization ID."},

			"entry": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"name":         schema.StringAttribute{Computed: true},
					"workbook_id":  schema.StringAttribute{Computed: true},
					"created_at":   schema.StringAttribute{Computed: true},
					"updated_at":   schema.StringAttribute{Computed: true},
					"revision_id":  schema.StringAttribute{Computed: true},
					"saved_id":     schema.StringAttribute{Computed: true},
					"published_id": schema.StringAttribute{Computed: true},
					"annotation": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"description": schema.StringAttribute{Computed: true},
						},
					},
					"meta": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"title":  schema.StringAttribute{Computed: true},
							"locale": schema.StringAttribute{Computed: true},
						},
					},
					"data": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"counter":             schema.Int64Attribute{Computed: true},
							"salt":                schema.StringAttribute{Computed: true},
							"scheme_version":      schema.Int64Attribute{Computed: true},
							"access_description":  schema.StringAttribute{Computed: true},
							"support_description": schema.StringAttribute{Computed: true},
							"settings":            dsSettingsAttribute(),
							"tabs":                dsTabsAttribute(),
						},
					},
				},
			},
		},
	}
}

func dsSettingsAttribute() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Computed: true,
		Attributes: map[string]schema.Attribute{
			"autoupdate_interval":      schema.Int64Attribute{Computed: true},
			"max_concurrent_requests":  schema.Int64Attribute{Computed: true},
			"silent_loading":           schema.BoolAttribute{Computed: true},
			"dependent_selectors":      schema.BoolAttribute{Computed: true},
			"expand_toc":               schema.BoolAttribute{Computed: true},
			"hide_dash_title":          schema.BoolAttribute{Computed: true},
			"hide_tabs":                schema.BoolAttribute{Computed: true},
			"load_only_visible_charts": schema.BoolAttribute{Computed: true},
			"load_priority":            schema.StringAttribute{Computed: true},
			"global_params":            schema.MapAttribute{Computed: true, ElementType: types.StringType},
		},
	}
}

func dsTabsAttribute() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		Computed: true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"id":    schema.StringAttribute{Computed: true},
				"title": schema.StringAttribute{Computed: true},
				"items": dsTabItemsAttribute(),
				"layout": schema.ListNestedAttribute{
					Computed: true,
					NestedObject: schema.NestedAttributeObject{
						Attributes: map[string]schema.Attribute{
							"i":      schema.StringAttribute{Computed: true},
							"x":      schema.Int64Attribute{Computed: true},
							"y":      schema.Int64Attribute{Computed: true},
							"w":      schema.Int64Attribute{Computed: true},
							"h":      schema.Int64Attribute{Computed: true},
							"parent": schema.StringAttribute{Computed: true},
						},
					},
				},
				"connections": schema.ListNestedAttribute{
					Computed: true,
					NestedObject: schema.NestedAttributeObject{
						Attributes: map[string]schema.Attribute{
							"from": schema.StringAttribute{Computed: true},
							"to":   schema.StringAttribute{Computed: true},
							"kind": schema.StringAttribute{Computed: true},
						},
					},
				},
				"aliases": schema.MapAttribute{
					Computed:    true,
					ElementType: types.ListType{ElemType: types.ListType{ElemType: types.StringType}},
				},
			},
		},
	}
}

func dsTabItemsAttribute() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		Computed: true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"id":        schema.StringAttribute{Computed: true},
				"namespace": schema.StringAttribute{Computed: true},
				"widget": schema.SingleNestedAttribute{
					Computed: true,
					Attributes: map[string]schema.Attribute{
						"hide_title": schema.BoolAttribute{Computed: true},
						"tabs": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"id":          schema.StringAttribute{Computed: true},
									"title":       schema.StringAttribute{Computed: true},
									"description": schema.StringAttribute{Computed: true},
									"chart_id":    schema.StringAttribute{Computed: true},
									"is_default":  schema.BoolAttribute{Computed: true},
									"auto_height": schema.BoolAttribute{Computed: true},
									"params":      schema.MapAttribute{Computed: true, ElementType: types.StringType},
								},
							},
						},
					},
				},
				"group_control": schema.SingleNestedAttribute{
					Computed: true,
					Attributes: map[string]schema.Attribute{
						"auto_height":  schema.BoolAttribute{Computed: true},
						"button_apply": schema.BoolAttribute{Computed: true},
						"button_reset": schema.BoolAttribute{Computed: true},
						"group": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"id":             schema.StringAttribute{Computed: true},
									"namespace":      schema.StringAttribute{Computed: true},
									"placement_mode": schema.StringAttribute{Computed: true},
									"defaults":       schema.MapAttribute{Computed: true, ElementType: types.ListType{ElemType: types.StringType}},
									"source": schema.SingleNestedAttribute{
										Computed: true,
										Attributes: map[string]schema.Attribute{
											"acceptable_values": schema.ListNestedAttribute{
												Computed: true,
												NestedObject: schema.NestedAttributeObject{
													Attributes: map[string]schema.Attribute{
														"title": schema.StringAttribute{Computed: true},
														"value": schema.StringAttribute{Computed: true},
													},
												},
											},
											"accent_type":     schema.StringAttribute{Computed: true},
											"default_value":   schema.ListAttribute{Computed: true, ElementType: types.StringType},
											"element_type":    schema.StringAttribute{Computed: true},
											"field_name":      schema.StringAttribute{Computed: true},
											"hint":            schema.StringAttribute{Computed: true},
											"multiselectable": schema.BoolAttribute{Computed: true},
											"required":        schema.BoolAttribute{Computed: true},
											"show_hint":       schema.BoolAttribute{Computed: true},
											"title":           schema.StringAttribute{Computed: true},
										},
									},
								},
							},
						},
					},
				},
				"text": schema.SingleNestedAttribute{
					Computed: true,
					Attributes: map[string]schema.Attribute{
						"text": schema.StringAttribute{Computed: true},
					},
				},
				"title": schema.SingleNestedAttribute{
					Computed: true,
					Attributes: map[string]schema.Attribute{
						"text":        schema.StringAttribute{Computed: true},
						"size":        schema.StringAttribute{Computed: true},
						"show_in_toc": schema.BoolAttribute{Computed: true},
					},
				},
				"image": schema.SingleNestedAttribute{
					Computed: true,
					Attributes: map[string]schema.Attribute{
						"src":         schema.StringAttribute{Computed: true},
						"alt_text":    schema.StringAttribute{Computed: true},
						"description": schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}
