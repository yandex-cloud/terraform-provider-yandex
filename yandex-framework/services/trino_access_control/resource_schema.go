package trino_access_control

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/trino_access_control/models"
)

func AccessControlResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cluster_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description:         "ID of the Trino cluster. Provided by the client when the Access Control is created.",
				MarkdownDescription: "ID of the Trino cluster. Provided by the client when the Access Control is created.",
			},
			"catalogs": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Validators: []validator.Object{
						// rule cannot be null
						objectvalidator.IsRequired(),
					},
					Attributes: map[string]schema.Attribute{
						"users":   usersAttributeSchema(),
						"groups":  groupsAttributeSchema(),
						"catalog": catalogAccessRuleMatcherSchema(),
						"permission": requiredEnumAttribute("Permission granted by the rule.",
							string(models.CatalogPermissionNone),
							string(models.CatalogPermissionReadOnly),
							string(models.CatalogPermissionAll)),
						"description": descriptionAttributeSchema(),
					},
				},
				Description:         "Catalog level access control rules.",
				MarkdownDescription: "Catalog level access control rules.",
			},
			"schemas": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Validators: []validator.Object{
						// rule cannot be null
						objectvalidator.IsRequired(),
					},
					Attributes: map[string]schema.Attribute{
						"users":   usersAttributeSchema(),
						"groups":  groupsAttributeSchema(),
						"catalog": catalogAccessRuleMatcherSchema(),
						"schema":  nameMatcherSchema("Schema", "schemas"),
						"owner": requiredEnumAttribute("Ownership granted by the rule.",
							string(models.SchemaOwnerNo),
							string(models.SchemaOwnerYes)),
						"description": descriptionAttributeSchema(),
					},
				},
				Description:         "Schema level access control rules.",
				MarkdownDescription: "Schema level access control rules.",
			},
			"functions": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Validators: []validator.Object{
						// rule cannot be null
						objectvalidator.IsRequired(),
					},
					Attributes: map[string]schema.Attribute{
						"users":    usersAttributeSchema(),
						"groups":   groupsAttributeSchema(),
						"catalog":  catalogAccessRuleMatcherSchema(),
						"schema":   nameMatcherSchema("Schema", "Schemas"),
						"function": nameMatcherSchema("Function", "functions"),
						"privileges": privilegesListSchema(
							string(models.FunctionPrivilegeExecute),
							string(models.FunctionPrivilegeGrantExecute),
							string(models.FunctionPrivilegeOwnership),
						),
						"description": descriptionAttributeSchema(),
					},
				},
				Description:         "Function level access control rules.",
				MarkdownDescription: "Function level access control rules.",
			},
			"procedures": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Validators: []validator.Object{
						// rule cannot be null
						objectvalidator.IsRequired(),
					},
					Attributes: map[string]schema.Attribute{
						"users":       usersAttributeSchema(),
						"groups":      groupsAttributeSchema(),
						"catalog":     catalogAccessRuleMatcherSchema(),
						"schema":      nameMatcherSchema("Schema", "Schemas"),
						"procedure":   nameMatcherSchema("Procedure", "procedures"),
						"privileges":  privilegesListSchema(string(models.ProcedurePrivilegeExecute)),
						"description": descriptionAttributeSchema(),
					},
				},
				Description:         "Procedure level access control rules.",
				MarkdownDescription: "Procedure level access control rules.",
			},
			"tables": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Validators: []validator.Object{
						// rule cannot be null
						objectvalidator.IsRequired(),
					},
					Attributes: map[string]schema.Attribute{
						"users":   usersAttributeSchema(),
						"groups":  groupsAttributeSchema(),
						"catalog": catalogAccessRuleMatcherSchema(),
						"schema":  nameMatcherSchema("Schema", "Schemas"),
						"table":   nameMatcherSchema("Table", "tables"),
						"privileges": privilegesListSchema(
							string(models.TablePrivilegeSelect),
							string(models.TablePrivilegeInsert),
							string(models.TablePrivilegeDelete),
							string(models.TablePrivilegeUpdate),
							string(models.TablePrivilegeOwnership),
							string(models.TablePrivilegeGrantSelect),
						),
						"columns": schema.ListNestedAttribute{
							Optional: true,
							NestedObject: schema.NestedAttributeObject{
								Validators: []validator.Object{
									// column cannot be null
									objectvalidator.IsRequired(),
								},
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthAtMost(63),
										},
										Description:         "Column name.",
										MarkdownDescription: "Column name.",
									},
									"access": requiredEnumAttribute("Column access mode.", string(models.ColumnAccessModeNone), string(models.ColumnAccessModeAll)),
									"mask": schema.StringAttribute{
										Optional: true,
										Validators: []validator.String{
											stringvalidator.LengthAtMost(128),
										},
										Description:         "SQL expression mask to evaluate instead of original column values.",
										MarkdownDescription: "SQL expression mask to evaluate instead of original column values.",
									},
								},
							},
							Description:         "Column rules.",
							MarkdownDescription: "Column rules.",
						},
						"filter": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								stringvalidator.LengthAtMost(128),
							},
							Description:         "Boolean SQL expression to filter table rows for particular user.",
							MarkdownDescription: "Boolean SQL expression to filter table rows for particular user.",
						},
						"description": descriptionAttributeSchema(),
					},
				},
				Description:         "Table level access control rules.",
				MarkdownDescription: "Table level access control rules.",
			},
			"queries": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Validators: []validator.Object{
						// rule cannot be null
						objectvalidator.IsRequired(),
					},
					Attributes: map[string]schema.Attribute{
						"users":  usersAttributeSchema(),
						"groups": groupsAttributeSchema(),
						"query_owners": schema.ListAttribute{
							ElementType: types.StringType,
							Optional:    true,
							Validators: []validator.List{
								listvalidator.UniqueValues(),
							},
							Description:         "Owners of queries the rule is applied to.",
							MarkdownDescription: "Owners of queries the rule is applied to.",
						},
						"privileges": privilegesListSchema(
							string(models.QueryPrivilegeView),
							string(models.QueryPrivilegeExecute),
							string(models.QueryPrivilegeKill)),
						"description": descriptionAttributeSchema(),
					},
				},
				Description:         "Query level access control rules.",
				MarkdownDescription: "Query level access control rules.",
			},
			"system_session_properties": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Validators: []validator.Object{
						// rule cannot be null
						objectvalidator.IsRequired(),
					},
					Attributes: map[string]schema.Attribute{
						"users":    usersAttributeSchema(),
						"groups":   groupsAttributeSchema(),
						"property": nameMatcherSchema("Property", "properties"),
						"allow": requiredEnumAttribute("Whether the rule allows setting the property.",
							string(models.PropertyAllowNo),
							string(models.PropertyAllowYes)),
						"description": descriptionAttributeSchema(),
					},
				},
				Description:         "System session property access control rules.",
				MarkdownDescription: "System session property access control rules.",
			},
			"catalog_session_properties": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Validators: []validator.Object{
						// rule cannot be null
						objectvalidator.IsRequired(),
					},
					Attributes: map[string]schema.Attribute{
						"users":    usersAttributeSchema(),
						"groups":   groupsAttributeSchema(),
						"catalog":  catalogAccessRuleMatcherSchema(),
						"property": nameMatcherSchema("Property", "properties"),
						"allow": requiredEnumAttribute("Whether the rule allows setting the property.",
							string(models.PropertyAllowNo),
							string(models.PropertyAllowYes)),
						"description": descriptionAttributeSchema(),
					},
				},
				Description:         "Catalog session property access control rules.",
				MarkdownDescription: "Catalog session property access control rules.",
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": schema.SingleNestedBlock{
				CustomType: timeouts.Type{},
			},
		},
		Description:         "Access control configuration for Managed Trino cluster.",
		MarkdownDescription: "Access control configuration for Managed Trino cluster.",
	}
}

func usersAttributeSchema() schema.ListAttribute {
	return schema.ListAttribute{
		ElementType: types.StringType,
		Optional:    true,
		Validators: []validator.List{
			listvalidator.UniqueValues(),
		},
		Description:         "IAM user IDs the rule is applied to.",
		MarkdownDescription: "IAM user IDs the rule is applied to.",
	}
}

func groupsAttributeSchema() schema.ListAttribute {
	return schema.ListAttribute{
		ElementType: types.StringType,
		Optional:    true,
		Validators: []validator.List{
			listvalidator.UniqueValues(),
		},
		Description:         "IAM group IDs the rule is applied to.",
		MarkdownDescription: "IAM group IDs the rule is applied to.",
	}
}

func descriptionAttributeSchema() schema.StringAttribute {
	return schema.StringAttribute{
		Optional: true,
		Validators: []validator.String{
			stringvalidator.LengthAtMost(128),
		},
		Description:         "Rule description.",
		MarkdownDescription: "Rule description.",
	}
}

func requiredEnumAttribute(attrDescription string, values ...string) schema.StringAttribute {
	return schema.StringAttribute{
		Required: true,
		Validators: []validator.String{
			stringvalidator.OneOf(values...),
		},
		Description:         fmt.Sprintf("%s Valid values: %s", attrDescription, strings.Join(values, ", ")),
		MarkdownDescription: fmt.Sprintf("%s Valid values: `%s`", attrDescription, strings.Join(values, "`, `")),
	}
}

func privilegesListSchema(values ...string) schema.ListAttribute {
	return schema.ListAttribute{
		ElementType: types.StringType,
		Optional:    true,
		Validators: []validator.List{
			listvalidator.UniqueValues(),
			listvalidator.ValueStringsAre(stringvalidator.OneOf(values...)),
		},
		Description:         fmt.Sprintf("Privileges granted by the rule. Valid values: %s.", strings.Join(values, ", ")),
		MarkdownDescription: fmt.Sprintf("Privileges granted by the rule. Valid values: `%s`.", strings.Join(values, "`, `")),
	}
}

func catalogAccessRuleMatcherSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Attributes: map[string]schema.Attribute{
			"name_regexp": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("ids")),
				},
				Description:         "Catalog name regexp the rule is applied to.",
				MarkdownDescription: "Catalog name regexp the rule is applied to.",
			},
			"ids": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Validators: []validator.List{
					listvalidator.UniqueValues(),
					listvalidator.SizeAtLeast(1),
					listvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("name_regexp")),
				},
				Description:         "Catalog IDs rule is applied to.",
				MarkdownDescription: "Catalog IDs rule is applied to.",
			},
		},
		Optional:            true,
		Description:         "Catalog matcher specifying what catalogs the rule is applied to. Exactly one of name_regexp, ids attributes should be set.",
		MarkdownDescription: "Catalog matcher specifying what catalogs the rule is applied to. Exactly one of `name_regexp`, `ids` attributes should be set.",
	}
}

func nameMatcherSchema(pascalName, pluralName string) schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Attributes: map[string]schema.Attribute{
			"name_regexp": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("names")),
				},
				Description:         fmt.Sprintf("%s name regexp the rule is applied to.", pascalName),
				MarkdownDescription: fmt.Sprintf("%s name regexp the rule is applied to.", pascalName),
			},
			"names": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Validators: []validator.List{
					listvalidator.UniqueValues(),
					listvalidator.SizeAtLeast(1),
					listvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("name_regexp")),
				},
				Description:         fmt.Sprintf("%s names rule is applied to.", pascalName),
				MarkdownDescription: fmt.Sprintf("%s names rule is applied to.", pascalName),
			},
		},
		Optional:            true,
		Description:         fmt.Sprintf("Matcher specifying what %s the rule is applied to. Exactly one of name_regexp, names attributes should be set.", pluralName),
		MarkdownDescription: fmt.Sprintf("Matcher specifying what %s the rule is applied to. Exactly one of `name_regexp`, `names` attributes should be set.", pluralName),
	}
}
