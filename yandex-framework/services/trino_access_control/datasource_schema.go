package trino_access_control

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/trino_access_control/models"
)

// Same as [AccessControlResourceSchema] but all attributes except cluster_id are Computed and have no validators.
func AccessControlDatasourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cluster_id": schema.StringAttribute{
				Required:            true,
				Description:         "ID of the Trino cluster. Provided by the client when the Access Control is created.",
				MarkdownDescription: "ID of the Trino cluster. Provided by the client when the Access Control is created.",
			},
			"catalogs": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"users":   usersAttributeDatasourceSchema(),
						"groups":  groupsAttributeDatasourceSchema(),
						"catalog": catalogAccessRuleMatcherDatasourceSchema(),
						"permission": datasourceEnumAttribute("Permission granted by the rule.",
							string(models.CatalogPermissionNone),
							string(models.CatalogPermissionReadOnly),
							string(models.CatalogPermissionAll)),
						"description": descriptionAttributeDatasourceSchema(),
					},
				},
				Description:         "Catalog access control rules.",
				MarkdownDescription: "Catalog access control rules.",
			},
			"schemas": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"users":   usersAttributeDatasourceSchema(),
						"groups":  groupsAttributeDatasourceSchema(),
						"catalog": catalogAccessRuleMatcherDatasourceSchema(),
						"schema":  nameMatcherDatasourceSchema("Schema", "schemas"),
						"owner": datasourceEnumAttribute("Ownership granted by the rule.",
							string(models.SchemaOwnerNo),
							string(models.SchemaOwnerYes)),
						"description": descriptionAttributeDatasourceSchema(),
					},
				},
				Description:         "Schema access control rules.",
				MarkdownDescription: "Schema access control rules.",
			},
			"functions": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"users":    usersAttributeDatasourceSchema(),
						"groups":   groupsAttributeDatasourceSchema(),
						"catalog":  catalogAccessRuleMatcherDatasourceSchema(),
						"schema":   nameMatcherDatasourceSchema("Schema", "Schemas"),
						"function": nameMatcherDatasourceSchema("Function", "functions"),
						"privileges": privilegesListDatasourceSchema(
							string(models.FunctionPrivilegeExecute),
							string(models.FunctionPrivilegeGrantExecute),
							string(models.FunctionPrivilegeOwnership),
						),
						"description": descriptionAttributeDatasourceSchema(),
					},
				},
				Description:         "Function access control rules.",
				MarkdownDescription: "Function access control rules.",
			},
			"procedures": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"users":       usersAttributeDatasourceSchema(),
						"groups":      groupsAttributeDatasourceSchema(),
						"catalog":     catalogAccessRuleMatcherDatasourceSchema(),
						"schema":      nameMatcherDatasourceSchema("Schema", "Schemas"),
						"procedure":   nameMatcherDatasourceSchema("Procedure", "procedures"),
						"privileges":  privilegesListDatasourceSchema(string(models.ProcedurePrivilegeExecute)),
						"description": descriptionAttributeDatasourceSchema(),
					},
				},
				Description:         "Procedure access control rules.",
				MarkdownDescription: "Procedure access control rules.",
			},
			"tables": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"users":   usersAttributeDatasourceSchema(),
						"groups":  groupsAttributeDatasourceSchema(),
						"catalog": catalogAccessRuleMatcherDatasourceSchema(),
						"schema":  nameMatcherDatasourceSchema("Schema", "Schemas"),
						"table":   nameMatcherDatasourceSchema("Table", "tables"),
						"privileges": privilegesListDatasourceSchema(
							string(models.TablePrivilegeSelect),
							string(models.TablePrivilegeInsert),
							string(models.TablePrivilegeDelete),
							string(models.TablePrivilegeUpdate),
							string(models.TablePrivilegeOwnership),
							string(models.TablePrivilegeGrantSelect),
						),
						"columns": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Computed:            true,
										Description:         "Column name.",
										MarkdownDescription: "Column name.",
									},
									"access": datasourceEnumAttribute("Column access mode.", string(models.ColumnAccessModeNone), string(models.ColumnAccessModeAll)),
									"mask": schema.StringAttribute{
										Computed:            true,
										Description:         "SQL expression mask to evaluate instead of original column values.",
										MarkdownDescription: "SQL expression mask to evaluate instead of original column values.",
									},
								},
							},
							Description:         "Column rules.",
							MarkdownDescription: "Column rules.",
						},
						"filter": schema.StringAttribute{
							Computed:            true,
							Description:         "Boolean SQL expression to filter table rows for particular user.",
							MarkdownDescription: "Boolean SQL expression to filter table rows for particular user.",
						},
						"description": descriptionAttributeDatasourceSchema(),
					},
				},
				Description:         "Table access control rules.",
				MarkdownDescription: "Table access control rules.",
			},
			"system_session_properties": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"users":    usersAttributeDatasourceSchema(),
						"groups":   groupsAttributeDatasourceSchema(),
						"property": nameMatcherDatasourceSchema("Property", "properties"),
						"allow": datasourceEnumAttribute("Whether the rule allows setting the property.",
							string(models.PropertyAllowNo),
							string(models.PropertyAllowYes)),
						"description": descriptionAttributeDatasourceSchema(),
					},
				},
				Description:         "System session property access control rules.",
				MarkdownDescription: "System session property access control rules.",
			},
			"catalog_session_properties": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"users":    usersAttributeDatasourceSchema(),
						"groups":   groupsAttributeDatasourceSchema(),
						"catalog":  catalogAccessRuleMatcherDatasourceSchema(),
						"property": nameMatcherDatasourceSchema("Property", "properties"),
						"allow": datasourceEnumAttribute("Whether the rule allows setting the property.",
							string(models.PropertyAllowNo),
							string(models.PropertyAllowYes)),
						"description": descriptionAttributeDatasourceSchema(),
					},
				},
				Description:         "Catalog session property access control rules.",
				MarkdownDescription: "Catalog session property access control rules.",
			},
			"queries": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"users":  usersAttributeDatasourceSchema(),
						"groups": groupsAttributeDatasourceSchema(),
						"query_owners": schema.ListAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							Description:         "Owners of queries the rule is applied to.",
							MarkdownDescription: "Owners of queries the rule is applied to.",
						},
						"privileges": privilegesListDatasourceSchema(
							string(models.QueryPrivilegeView),
							string(models.QueryPrivilegeExecute),
							string(models.QueryPrivilegeKill)),
						"description": descriptionAttributeDatasourceSchema(),
					},
				},
				Description:         "Query access control rules.",
				MarkdownDescription: "Query access control rules.",
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Read: true,
			}),
		},
		Description:         "Access control configuration for Trino cluster.",
		MarkdownDescription: "Access control configuration for Trino cluster.",
	}
}

func usersAttributeDatasourceSchema() schema.ListAttribute {
	return schema.ListAttribute{
		ElementType:         types.StringType,
		Computed:            true,
		Description:         "IAM user IDs the rule is applied to.",
		MarkdownDescription: "IAM user IDs the rule is applied to.",
	}
}

func groupsAttributeDatasourceSchema() schema.ListAttribute {
	return schema.ListAttribute{
		ElementType:         types.StringType,
		Computed:            true,
		Description:         "IAM group IDs the rule is applied to.",
		MarkdownDescription: "IAM group IDs the rule is applied to.",
	}
}

func descriptionAttributeDatasourceSchema() schema.StringAttribute {
	return schema.StringAttribute{
		Computed:            true,
		Description:         "Rule description.",
		MarkdownDescription: "Rule description.",
	}
}

func datasourceEnumAttribute(attrDescription string, values ...string) schema.StringAttribute {
	return schema.StringAttribute{
		Computed:            true,
		Description:         fmt.Sprintf("%s Valid values: %s", attrDescription, strings.Join(values, ", ")),
		MarkdownDescription: fmt.Sprintf("%s Valid values: `%s`", attrDescription, strings.Join(values, "`, `")),
	}
}

func privilegesListDatasourceSchema(values ...string) schema.ListAttribute {
	return schema.ListAttribute{
		ElementType:         types.StringType,
		Computed:            true,
		Description:         fmt.Sprintf("Privileges granted by the rule. Valid values: %s.", strings.Join(values, ", ")),
		MarkdownDescription: fmt.Sprintf("Privileges granted by the rule. Valid values: `%s`.", strings.Join(values, "`, `")),
	}
}

// Helper function for catalog access rule matcher
func catalogAccessRuleMatcherDatasourceSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Attributes: map[string]schema.Attribute{
			"name_regexp": schema.StringAttribute{
				Computed:            true,
				Description:         "Catalog name regexp the rule is applied to.",
				MarkdownDescription: "Catalog name regexp the rule is applied to.",
			},
			"ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				Description:         "Catalog IDs rule is applied to.",
				MarkdownDescription: "Catalog IDs rule is applied to.",
			},
		},
		Computed:            true,
		Description:         "Catalog matcher specifying what catalogs the rule is applied to.",
		MarkdownDescription: "Catalog matcher specifying what catalogs the rule is applied to.",
	}
}

func nameMatcherDatasourceSchema(pascalName, pluralName string) schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Attributes: map[string]schema.Attribute{
			"name_regexp": schema.StringAttribute{
				Computed:            true,
				Description:         fmt.Sprintf("%s name regexp the rule is applied to.", pascalName),
				MarkdownDescription: fmt.Sprintf("%s name regexp the rule is applied to.", pascalName),
			},
			"names": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				Description:         fmt.Sprintf("%s names rule is applied to.", pascalName),
				MarkdownDescription: fmt.Sprintf("%s names rule is applied to.", pascalName),
			},
		},
		Computed:            true,
		Description:         fmt.Sprintf("Matcher specifying what %s the rule is applied to.", pluralName),
		MarkdownDescription: fmt.Sprintf("Matcher specifying what %s the rule is applied to.", pluralName),
	}
}
