package datalens_dataset

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func dataSourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Retrieves information about a DataLens dataset.",
		Attributes: map[string]schema.Attribute{
			"id":              schema.StringAttribute{Required: true, MarkdownDescription: "The ID of the dataset."},
			"organization_id": schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "The organization ID."},

			"workbook_id": schema.StringAttribute{Computed: true, MarkdownDescription: "The workbook ID."},
			"dir_path":    schema.StringAttribute{Computed: true, MarkdownDescription: "The directory path."},
			"name":        schema.StringAttribute{Computed: true, MarkdownDescription: "Dataset name."},
			"created_via": schema.StringAttribute{Computed: true, MarkdownDescription: "Origin of the dataset entry."},
			"preview":     schema.BoolAttribute{Computed: true, MarkdownDescription: "Preview-on-create flag."},
			"is_favorite": schema.BoolAttribute{Computed: true, MarkdownDescription: "Favorite flag."},

			"dataset": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Dataset content.",
				Attributes: map[string]schema.Attribute{
					"description":             schema.StringAttribute{Computed: true, MarkdownDescription: "Dataset description."},
					"load_preview_by_default": schema.BoolAttribute{Computed: true, MarkdownDescription: "Auto-load preview."},
					"template_enabled":        schema.BoolAttribute{Computed: true, MarkdownDescription: "Enable templating."},
					"data_export_forbidden":   schema.BoolAttribute{Computed: true, MarkdownDescription: "Forbid raw data export."},
					"schema_update_enabled":   schema.BoolAttribute{Computed: true, MarkdownDescription: "Allow schema updates."},
					"preview_enabled":         schema.BoolAttribute{Computed: true, MarkdownDescription: "Enable preview functionality."},

					"avatar_relations":   datasourceAvatarRelations(),
					"source_avatars":     datasourceSourceAvatars(),
					"sources":            datasourceSources(),
					"result_schema":      datasourceResultSchema(),
					"obligatory_filters": datasourceObligatoryFilters(),
					"rls2":               datasourceRls2(),
					"cache_invalidation_source": schema.SingleNestedAttribute{
						Computed:            true,
						MarkdownDescription: "Cache invalidation source.",
						Attributes: map[string]schema.Attribute{
							"mode":    schema.StringAttribute{Computed: true},
							"field":   schema.StringAttribute{Computed: true},
							"sql":     schema.StringAttribute{Computed: true},
							"filters": schema.ListAttribute{Computed: true, ElementType: types.StringType},
						},
					},
				},
			},
		},
	}
}

func datasourceAvatarRelations() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		Computed: true, MarkdownDescription: "Join relationships.",
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"id":              schema.StringAttribute{Computed: true},
				"left_avatar_id":  schema.StringAttribute{Computed: true},
				"right_avatar_id": schema.StringAttribute{Computed: true},
				"join_type":       schema.StringAttribute{Computed: true},
				"managed_by":      schema.StringAttribute{Computed: true},
				"required":        schema.BoolAttribute{Computed: true},
				"virtual":         schema.BoolAttribute{Computed: true},
				"conditions": schema.ListNestedAttribute{
					Computed: true,
					NestedObject: schema.NestedAttributeObject{
						Attributes: map[string]schema.Attribute{
							"type":     schema.StringAttribute{Computed: true},
							"operator": schema.StringAttribute{Computed: true},
							"left":     datasourceJoinPart(),
							"right":    datasourceJoinPart(),
						},
					},
				},
			},
		},
	}
}

func datasourceJoinPart() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Computed: true,
		Attributes: map[string]schema.Attribute{
			"calc_mode": schema.StringAttribute{Computed: true},
			"source":    schema.StringAttribute{Computed: true},
			"field_id":  schema.StringAttribute{Computed: true},
			"formula":   schema.StringAttribute{Computed: true},
		},
	}
}

func datasourceSourceAvatars() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		Computed: true, MarkdownDescription: "Source avatars.",
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"id":         schema.StringAttribute{Computed: true},
				"source_id":  schema.StringAttribute{Computed: true},
				"title":      schema.StringAttribute{Computed: true},
				"is_root":    schema.BoolAttribute{Computed: true},
				"managed_by": schema.StringAttribute{Computed: true},
				"valid":      schema.BoolAttribute{Computed: true},
				"virtual":    schema.BoolAttribute{Computed: true},
			},
		},
	}
}

func datasourceSources() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		Computed: true, MarkdownDescription: "Underlying data sources.",
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"id":            schema.StringAttribute{Computed: true},
				"title":         schema.StringAttribute{Computed: true},
				"source_type":   schema.StringAttribute{Computed: true},
				"connection_id": schema.StringAttribute{Computed: true},
				"managed_by":    schema.StringAttribute{Computed: true},
				"valid":         schema.BoolAttribute{Computed: true},
				"ref_source_id": schema.StringAttribute{Computed: true},
				"is_ref":        schema.BoolAttribute{Computed: true},
				"parameters": schema.SingleNestedAttribute{
					Computed: true,
					Attributes: map[string]schema.Attribute{
						"table_name":  schema.StringAttribute{Computed: true},
						"schema_name": schema.StringAttribute{Computed: true},
						"db_name":     schema.StringAttribute{Computed: true},
						"subsql":      schema.StringAttribute{Computed: true},
					},
				},
				"raw_schema": schema.ListNestedAttribute{
					Computed: true,
					NestedObject: schema.NestedAttributeObject{
						Attributes: map[string]schema.Attribute{
							"name":                 schema.StringAttribute{Computed: true},
							"title":                schema.StringAttribute{Computed: true},
							"user_type":            schema.StringAttribute{Computed: true},
							"description":          schema.StringAttribute{Computed: true},
							"nullable":             schema.BoolAttribute{Computed: true},
							"has_auto_aggregation": schema.BoolAttribute{Computed: true},
							"lock_aggregation":     schema.BoolAttribute{Computed: true},
							"native_type": schema.SingleNestedAttribute{
								Computed: true,
								Attributes: map[string]schema.Attribute{
									"name":                   schema.StringAttribute{Computed: true},
									"nullable":               schema.BoolAttribute{Computed: true},
									"native_type_class_name": schema.StringAttribute{Computed: true},
								},
							},
						},
					},
				},
			},
		},
	}
}

func datasourceResultSchema() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		Computed: true, MarkdownDescription: "Result schema fields.",
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"guid":                 schema.StringAttribute{Computed: true},
				"title":                schema.StringAttribute{Computed: true},
				"source":               schema.StringAttribute{Computed: true},
				"data_type":            schema.StringAttribute{Computed: true},
				"cast":                 schema.StringAttribute{Computed: true},
				"type":                 schema.StringAttribute{Computed: true},
				"aggregation":          schema.StringAttribute{Computed: true},
				"calc_mode":            schema.StringAttribute{Computed: true},
				"formula":              schema.StringAttribute{Computed: true},
				"guid_formula":         schema.StringAttribute{Computed: true},
				"default_value":        schema.StringAttribute{Computed: true},
				"description":          schema.StringAttribute{Computed: true},
				"hidden":               schema.BoolAttribute{Computed: true},
				"managed_by":           schema.StringAttribute{Computed: true},
				"valid":                schema.BoolAttribute{Computed: true},
				"avatar_id":            schema.StringAttribute{Computed: true},
				"has_auto_aggregation": schema.BoolAttribute{Computed: true},
				"lock_aggregation":     schema.BoolAttribute{Computed: true},
				"autoaggregated":       schema.BoolAttribute{Computed: true},
				"aggregation_locked":   schema.BoolAttribute{Computed: true},
				"value_constraint": schema.SingleNestedAttribute{
					Computed: true,
					Attributes: map[string]schema.Attribute{
						"type":    schema.StringAttribute{Computed: true},
						"pattern": schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func datasourceObligatoryFilters() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		Computed: true, MarkdownDescription: "Obligatory filters.",
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"id":         schema.StringAttribute{Computed: true},
				"field_guid": schema.StringAttribute{Computed: true},
				"managed_by": schema.StringAttribute{Computed: true},
				"valid":      schema.BoolAttribute{Computed: true},
				"default_filters": schema.ListNestedAttribute{
					Computed: true,
					NestedObject: schema.NestedAttributeObject{
						Attributes: map[string]schema.Attribute{
							"operation": schema.StringAttribute{Computed: true},
							"values":    schema.ListAttribute{Computed: true, ElementType: types.StringType},
							"disabled":  schema.BoolAttribute{Computed: true},
						},
					},
				},
			},
		},
	}
}

func datasourceRls2() schema.MapAttribute {
	return schema.MapAttribute{
		Computed:            true,
		MarkdownDescription: "Row-level security entries (RLS v2). Keyed by `field_guid`.",
		ElementType:         rls2EntryListType(),
	}
}
