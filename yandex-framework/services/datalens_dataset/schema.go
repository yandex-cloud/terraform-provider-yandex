package datalens_dataset

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	"github.com/yandex-cloud/terraform-provider-yandex/common/defaultschema"
)

const datasetMarkdownDescription = "Manages a DataLens dataset resource. Datasets define the schema, sources, joins, " +
	"calculated fields, obligatory filters and row-level security on top of a `yandex_datalens_connection`. " +
	"For more information, see [the official documentation](https://yandex.cloud/ru/docs/datalens/operations/api-start). " +
	"\n\n" +
	"Note: this resource exposes the well-documented portion of the DataLens dataset schema " +
	"(`avatar_relations`, `source_avatars`, `sources`, `result_schema`, `obligatory_filters`, `rls2` and flags). " +
	"Less common API fields (component_errors, source-type-specific parameters beyond `table_name`/`schema_name`/`db_name`/`subsql`) " +
	"may be added in follow-up releases."

func ResourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		MarkdownDescription: datasetMarkdownDescription,
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
			"workbook_id": schema.StringAttribute{
				MarkdownDescription: "The workbook ID where the dataset will be created. " +
					"Either `workbook_id` or `dir_path` must be specified. " +
					"Changing this attribute forces recreation of the resource.",
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"dir_path": schema.StringAttribute{
				MarkdownDescription: "The directory path where the dataset entry will be stored. " +
					"Either `workbook_id` or `dir_path` must be specified. " +
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
			"created_via": schema.StringAttribute{
				MarkdownDescription: "Origin of the dataset entry. One of `user`, `script`, `api`. " +
					"Changing this attribute forces recreation of the resource.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"preview": schema.BoolAttribute{
				MarkdownDescription: "Enable preview mode at create time.",
				Optional:            true,
			},
			"is_favorite": schema.BoolAttribute{
				MarkdownDescription: "Whether the dataset is marked as favorite for the calling user.",
				Computed:            true,
			},

			"dataset": schema.SingleNestedAttribute{
				MarkdownDescription: "Dataset content (`DatasetContentInternal`). Sent in full on every create/update; " +
					"DataLens always replaces the previous revision wholesale.",
				Required:   true,
				Attributes: datasetContentAttributes(),
			},
		},
	}
}

func datasetContentAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"description": schema.StringAttribute{
			MarkdownDescription: common.ResourceDescriptions["description"],
			Optional:            true,
		},
		"load_preview_by_default": schema.BoolAttribute{
			MarkdownDescription: "Auto-load preview when the dataset is opened.",
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
		},
		"template_enabled": schema.BoolAttribute{
			MarkdownDescription: "Enable templating.",
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
		},
		"data_export_forbidden": schema.BoolAttribute{
			MarkdownDescription: "Forbid raw data export.",
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
		},
		"schema_update_enabled": schema.BoolAttribute{
			MarkdownDescription: "Allow schema updates.",
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(true),
		},
		"preview_enabled": schema.BoolAttribute{
			MarkdownDescription: "Enable preview functionality.",
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(true),
		},

		"avatar_relations":   avatarRelationsAttribute(),
		"source_avatars":     sourceAvatarsAttribute(),
		"sources":            sourcesAttribute(),
		"result_schema":      resultSchemaAttribute(),
		"obligatory_filters": obligatoryFiltersAttribute(),
		"rls2":               rls2Attribute(),
		"cache_invalidation_source": schema.SingleNestedAttribute{
			MarkdownDescription: "Cache invalidation source configuration. Controls how DataLens detects when " +
				"cached query results should be considered stale.",
			Optional: true,
			Computed: true,
			// API echoes `{mode:"off"}` server-side default for unconfigured caches.
			// Declaring it as the schema Default makes plan-time TransformDefaults
			// fill the same value (so plan == state), and instructs the framework
			// to skip MarkComputedNilsAsUnknown for this attribute (and avoid the
			// known cascade-Unknown bug, see plugin-framework#898).
			Default: objectdefault.StaticValue(types.ObjectValueMust(
				cacheInvalidationSourceAttrTypes(),
				map[string]attr.Value{
					"mode":    types.StringValue("off"),
					"field":   types.StringNull(),
					"sql":     types.StringNull(),
					"filters": types.ListNull(types.StringType),
				},
			)),
			Attributes: map[string]schema.Attribute{
				"mode": schema.StringAttribute{
					MarkdownDescription: "Invalidation mode. Typical values: `off`, `automatic`, `field`, `sql`.",
					Required:            true,
				},
				"field": schema.StringAttribute{
					MarkdownDescription: "Field to monitor for cache invalidation (used with `mode = field`).",
					Optional:            true,
				},
				"sql": schema.StringAttribute{
					MarkdownDescription: "Validation query (used with `mode = sql`).",
					Optional:            true,
				},
				"filters": schema.ListAttribute{
					MarkdownDescription: "Additional filter expressions applied to the validation query.",
					Optional:            true,
					ElementType:         types.StringType,
				},
			},
		},
	}
}

func avatarRelationsAttribute() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "Join relationships between source avatars.",
		Optional:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"id":              schema.StringAttribute{Required: true, MarkdownDescription: "Relation identifier."},
				"left_avatar_id":  schema.StringAttribute{Required: true, MarkdownDescription: "Left source avatar ID."},
				"right_avatar_id": schema.StringAttribute{Required: true, MarkdownDescription: "Right source avatar ID."},
				"join_type": schema.StringAttribute{
					Required:            true,
					MarkdownDescription: "Join type. One of `inner`, `left`, `right`, `full`.",
				},
				"managed_by": schema.StringAttribute{
					Optional:            true,
					Computed:            true,
					Default:             stringdefault.StaticString("user"),
					MarkdownDescription: "Manager of the relation. One of `user`, `feature`, `component`.",
				},
				"required": schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Whether the join is mandatory."},
				"virtual":  schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Virtual relation flag."},
				"conditions": schema.ListNestedAttribute{
					Required:            true,
					MarkdownDescription: "Join conditions.",
					NestedObject: schema.NestedAttributeObject{
						Attributes: map[string]schema.Attribute{
							"type": schema.StringAttribute{
								Required:            true,
								MarkdownDescription: "Condition type. Currently `binary`.",
							},
							"operator": schema.StringAttribute{
								Required:            true,
								MarkdownDescription: "Comparison operator. One of `eq`, `ne`, `gt`, `gte`, `lt`, `lte`.",
							},
							"left":  joinPartAttribute(true),
							"right": joinPartAttribute(true),
						},
					},
				},
			},
		},
	}
}

func joinPartAttribute(required bool) schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		MarkdownDescription: "Join operand. `calc_mode` selects which one of `source`, `field_id`, `formula` is meaningful.",
		Required:            required,
		Attributes: map[string]schema.Attribute{
			"calc_mode": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Operand mode. One of `direct`, `formula`, `result_field`.",
			},
			"source":   schema.StringAttribute{Optional: true, MarkdownDescription: "Column name (used when `calc_mode = direct`)."},
			"field_id": schema.StringAttribute{Optional: true, MarkdownDescription: "Result field GUID (used when `calc_mode = result_field`)."},
			"formula":  schema.StringAttribute{Optional: true, MarkdownDescription: "Inline formula (used when `calc_mode = formula`)."},
		},
	}
}

func sourceAvatarsAttribute() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "Source avatars (instances of sources used as join nodes).",
		Optional:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"id":        schema.StringAttribute{Required: true, MarkdownDescription: "Avatar identifier."},
				"source_id": schema.StringAttribute{Required: true, MarkdownDescription: "Backing source ID."},
				"title":     schema.StringAttribute{Required: true, MarkdownDescription: "Display title."},
				"is_root":   schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Whether this avatar is the root."},
				"managed_by": schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("user"), MarkdownDescription: "Manager. One of `user`, `feature`, `component`."},
				"valid":   schema.BoolAttribute{Computed: true, MarkdownDescription: "Validation status."},
				"virtual": schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Virtual avatar flag."},
			},
		},
	}
}

func sourcesAttribute() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "Underlying data sources (table or subselect references on top of a connection).",
		Optional:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"id":            schema.StringAttribute{Required: true, MarkdownDescription: "Source identifier."},
				"title":         schema.StringAttribute{Required: true, MarkdownDescription: "Display title."},
				"source_type":   schema.StringAttribute{Required: true, MarkdownDescription: "Source type (e.g. `CH_TABLE`, `PG_SUBSELECT`, `YDB_TABLE`, ...)."},
				"connection_id": schema.StringAttribute{Required: true, MarkdownDescription: "Backing `yandex_datalens_connection` ID."},
				"managed_by":    schema.StringAttribute{Optional: true, Computed: true, Default: stringdefault.StaticString("user"), MarkdownDescription: "Manager."},
				"valid":         schema.BoolAttribute{Computed: true, MarkdownDescription: "Validation status."},
				"ref_source_id": schema.StringAttribute{Optional: true, MarkdownDescription: "Reference source ID."},
				"is_ref":        schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Whether this source is a reference."},
				"parameters": schema.SingleNestedAttribute{
					MarkdownDescription: "Common source parameters. Less common per-source-type fields can be added in follow-up releases.",
					Optional:            true,
					Attributes: map[string]schema.Attribute{
						"table_name":  schema.StringAttribute{Optional: true, MarkdownDescription: "Source table name."},
						"schema_name": schema.StringAttribute{Optional: true, MarkdownDescription: "Source schema name."},
						"db_name":     schema.StringAttribute{Optional: true, MarkdownDescription: "Source database name."},
						"subsql":      schema.StringAttribute{Optional: true, MarkdownDescription: "Subquery SQL (used by `*_SUBSELECT` source types)."},
					},
				},
				"raw_schema": schema.ListNestedAttribute{
					MarkdownDescription: "Raw schema returned by the underlying source.",
					Optional:            true,
					NestedObject: schema.NestedAttributeObject{
						Attributes: map[string]schema.Attribute{
							"name": schema.StringAttribute{Required: true, MarkdownDescription: "Column name."},
							"title": schema.StringAttribute{Required: true, MarkdownDescription: "Column title (display name)."},
							"user_type": schema.StringAttribute{Required: true, MarkdownDescription: "DataLens user type (`string`, `integer`, ...)."},
							"description": schema.StringAttribute{
								Optional:            true,
								Computed:            true,
								Default:             stringdefault.StaticString(""),
								MarkdownDescription: "Column description.",
							},
							"nullable": schema.BoolAttribute{
								Optional:            true,
								Computed:            true,
								Default:             booldefault.StaticBool(true),
								MarkdownDescription: "Whether the column is nullable.",
							},
							"has_auto_aggregation": schema.BoolAttribute{
								Optional:            true,
								Computed:            true,
								Default:             booldefault.StaticBool(false),
								MarkdownDescription: "Whether the column auto-aggregates.",
							},
							"lock_aggregation": schema.BoolAttribute{
								Optional:            true,
								Computed:            true,
								Default:             booldefault.StaticBool(false),
								MarkdownDescription: "Whether aggregation is locked for this column.",
							},
							"native_type": schema.SingleNestedAttribute{
								MarkdownDescription: "Native type metadata.",
								Optional:            true,
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{Required: true, MarkdownDescription: "Native type name."},
									"nullable": schema.BoolAttribute{
										Optional:            true,
										Computed:            true,
										Default:             booldefault.StaticBool(true),
										MarkdownDescription: "Whether the column is nullable.",
									},
									"native_type_class_name": schema.StringAttribute{
										Optional:            true,
										Computed:            true,
										Default:             stringdefault.StaticString("common_native_type"),
										MarkdownDescription: "Native type class. Defaults to `common_native_type`.",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resultSchemaAttribute() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "Result schema fields (DIMENSIONs and MEASUREs exposed by the dataset).",
		Optional:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"guid":  schema.StringAttribute{Required: true, MarkdownDescription: "Field GUID."},
				"title": schema.StringAttribute{Required: true, MarkdownDescription: "Field display title."},
				"source": schema.StringAttribute{Optional: true, Computed: true,
					MarkdownDescription: "Source column name (used when `calc_mode = direct`)."},
				"data_type": schema.StringAttribute{
					Required:            true,
					MarkdownDescription: "DataLens data type.",
				},
				"cast": schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Cast type."},
				"type": schema.StringAttribute{
					Required:            true,
					MarkdownDescription: "Field role. One of `DIMENSION`, `MEASURE`.",
				},
				"aggregation": schema.StringAttribute{Optional: true, Computed: true,
					MarkdownDescription: "Aggregation function."},
				"calc_mode": schema.StringAttribute{Optional: true, Computed: true,
					MarkdownDescription: "Calculation mode. One of `direct`, `formula`, `parameter`."},
				"formula":      schema.StringAttribute{Optional: true, MarkdownDescription: "Formula (used when `calc_mode = formula`)."},
				"guid_formula": schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Formula with GUID-substituted field references."},
				"default_value": schema.StringAttribute{Optional: true, MarkdownDescription: "Default value (parameters)."},
				"description":  schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Field description."},
				"hidden":       schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Hide field from UI."},
				"managed_by":   schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Manager."},
				"valid":        schema.BoolAttribute{Computed: true, MarkdownDescription: "Validation status."},
				"avatar_id":    schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Source avatar this field belongs to."},
				"has_auto_aggregation": schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Auto-aggregation flag."},
				"lock_aggregation":     schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Whether aggregation is locked."},
				"autoaggregated":       schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Auto-aggregated flag."},
				"aggregation_locked":   schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Aggregation locked flag."},
				"value_constraint": schema.SingleNestedAttribute{
					MarkdownDescription: "Constraints on the field value.",
					Optional:            true,
					Attributes: map[string]schema.Attribute{
						"type":    schema.StringAttribute{Required: true, MarkdownDescription: "Constraint type."},
						"pattern": schema.StringAttribute{Optional: true, MarkdownDescription: "Pattern (for pattern-based constraints)."},
					},
				},
			},
		},
	}
}

func obligatoryFiltersAttribute() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "Obligatory filters applied whenever the dataset is queried.",
		Optional:            true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"id":          schema.StringAttribute{Required: true, MarkdownDescription: "Filter identifier."},
				"field_guid":  schema.StringAttribute{Required: true, MarkdownDescription: "Target field GUID."},
				"managed_by":  schema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Manager."},
				"valid":       schema.BoolAttribute{Computed: true, MarkdownDescription: "Validation status."},
				"default_filters": schema.ListNestedAttribute{
					MarkdownDescription: "Default filter conditions.",
					Required:            true,
					NestedObject: schema.NestedAttributeObject{
						Attributes: map[string]schema.Attribute{
							"operation": schema.StringAttribute{Required: true, MarkdownDescription: "Filter operation (e.g. `EQ`, `IN`, `BETWEEN`, `CONTAINS`)."},
							"values":    schema.ListAttribute{Required: true, ElementType: types.StringType, MarkdownDescription: "Filter values."},
							"disabled":  schema.BoolAttribute{Optional: true, Computed: true, MarkdownDescription: "Whether this default is currently disabled."},
						},
					},
				},
			},
		},
	}
}

func rls2Attribute() schema.MapAttribute {
	return schema.MapAttribute{
		MarkdownDescription: "Row-level security entries (RLS v2). Keyed by `field_guid` per the DataLens API: " +
			"`{ \"<field_guid>\": [ { subject = { subject_id, subject_type }, allowed_value, pattern_type } ] }`.",
		Optional:    true,
		ElementType: rls2EntryListType(),
	}
}

func rls2EntryListType() types.ListType {
	return types.ListType{ElemType: rls2EntryObjectType()}
}

func rls2EntryObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"allowed_value": types.StringType,
		"pattern_type":  types.StringType,
		"subject":       rls2SubjectObjectType(),
	}}
}

func rls2SubjectObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"subject_id":   types.StringType,
		"subject_type": types.StringType,
	}}
}

func cacheInvalidationSourceAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"mode":    types.StringType,
		"field":   types.StringType,
		"sql":     types.StringType,
		"filters": types.ListType{ElemType: types.StringType},
	}
}
