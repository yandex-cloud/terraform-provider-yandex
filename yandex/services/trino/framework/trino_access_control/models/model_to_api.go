package models

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/trino/v1"
)

func ToAPI(ctx context.Context, model *AccessControlModel) (*trino.AccessControlConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	if model == nil {
		return nil, nil
	}
	cfg := &trino.AccessControlConfig{}

	if !model.Catalogs.IsNull() && !model.Catalogs.IsUnknown() && len(model.Catalogs.Elements()) > 0 {
		var catalogObjs []types.Object
		diags.Append(model.Catalogs.ElementsAs(ctx, &catalogObjs, false)...)
		for _, ruleObj := range catalogObjs {
			rule, dd := catalogRuleModelToAPI(ctx, ruleObj)
			diags.Append(dd...)
			cfg.Catalogs = append(cfg.Catalogs, rule)
		}
	}

	if !model.Schemas.IsNull() && !model.Schemas.IsUnknown() && len(model.Schemas.Elements()) > 0 {
		var schemaObjs []types.Object
		diags.Append(model.Schemas.ElementsAs(ctx, &schemaObjs, false)...)
		for _, ruleObj := range schemaObjs {
			rule, dd := schemaRuleModelToAPI(ctx, ruleObj)
			diags.Append(dd...)
			cfg.Schemas = append(cfg.Schemas, rule)
		}
	}

	if !model.Functions.IsNull() && !model.Functions.IsUnknown() && len(model.Functions.Elements()) > 0 {
		var functionObjs []types.Object
		diags.Append(model.Functions.ElementsAs(ctx, &functionObjs, false)...)
		for _, ruleObj := range functionObjs {
			rule, dd := functionRuleModelToAPI(ctx, ruleObj)
			diags.Append(dd...)
			cfg.Functions = append(cfg.Functions, rule)
		}
	}

	if !model.Procedures.IsNull() && !model.Procedures.IsUnknown() && len(model.Procedures.Elements()) > 0 {
		var procedureObjs []types.Object
		diags.Append(model.Procedures.ElementsAs(ctx, &procedureObjs, false)...)
		for _, ruleObj := range procedureObjs {
			rule, dd := procedureRuleModelToAPI(ctx, ruleObj)
			diags.Append(dd...)
			cfg.Procedures = append(cfg.Procedures, rule)
		}
	}

	if !model.Tables.IsNull() && !model.Tables.IsUnknown() && len(model.Tables.Elements()) > 0 {
		var tableObjs []types.Object
		diags.Append(model.Tables.ElementsAs(ctx, &tableObjs, false)...)
		for _, ruleObj := range tableObjs {
			rule, dd := tableRuleModelToAPI(ctx, ruleObj)
			diags.Append(dd...)
			cfg.Tables = append(cfg.Tables, rule)
		}
	}

	if !model.Queries.IsNull() && !model.Queries.IsUnknown() && len(model.Queries.Elements()) > 0 {
		var queryObjs []types.Object
		diags.Append(model.Queries.ElementsAs(ctx, &queryObjs, false)...)
		for _, ruleObj := range queryObjs {
			rule, dd := queryRuleModelToAPI(ctx, ruleObj)
			diags.Append(dd...)
			cfg.Queries = append(cfg.Queries, rule)
		}
	}

	if !model.SystemSessionProperties.IsNull() && !model.SystemSessionProperties.IsUnknown() && len(model.SystemSessionProperties.Elements()) > 0 {
		var sysPropObjs []types.Object
		diags.Append(model.SystemSessionProperties.ElementsAs(ctx, &sysPropObjs, false)...)
		for _, ruleObj := range sysPropObjs {
			rule, dd := systemSessionPropertyRuleModelToAPI(ctx, ruleObj)
			diags.Append(dd...)
			cfg.SystemSessionProperties = append(cfg.SystemSessionProperties, rule)
		}
	}

	if !model.CatalogSessionProperties.IsNull() && !model.CatalogSessionProperties.IsUnknown() && len(model.CatalogSessionProperties.Elements()) > 0 {
		var catPropObjs []types.Object
		diags.Append(model.CatalogSessionProperties.ElementsAs(ctx, &catPropObjs, false)...)
		for _, ruleObj := range catPropObjs {
			rule, dd := catalogSessionPropertyRuleModelToAPI(ctx, ruleObj)
			diags.Append(dd...)
			cfg.CatalogSessionProperties = append(cfg.CatalogSessionProperties, rule)
		}
	}

	return cfg, diags
}

func catalogRuleModelToAPI(ctx context.Context, ruleObj types.Object) (*trino.CatalogAccessRule, diag.Diagnostics) {
	var diags diag.Diagnostics
	if ruleObj.IsNull() || ruleObj.IsUnknown() {
		return nil, diags
	}

	var model CatalogRule
	diags.Append(ruleObj.As(ctx, &model, baseOptions)...)
	if diags.HasError() {
		return nil, diags
	}

	catalog, dd := catalogMatcherModelToAPI(ctx, model.Catalog)
	diags.Append(dd...)
	users, dd := stringListModelToAPI(ctx, model.Users)
	diags.Append(dd...)
	groups, dd := stringListModelToAPI(ctx, model.Groups)
	diags.Append(dd...)
	permission, dd := catalogPermissionStringToAPI(model.Permission.ValueString())
	diags.Append(dd...)
	description := model.Description.ValueString()

	return &trino.CatalogAccessRule{
		Catalog:     catalog,
		Users:       users,
		Groups:      groups,
		Permission:  permission,
		Description: description,
	}, diags
}

func catalogMatcherModelToAPI(ctx context.Context, matcherObj types.Object) (*trino.CatalogAccessRuleMatcher, diag.Diagnostics) {
	var diags diag.Diagnostics
	if matcherObj.IsNull() || matcherObj.IsUnknown() {
		return nil, diags
	}

	var model CatalogMatcherModel
	diags.Append(matcherObj.As(ctx, &model, baseOptions)...)
	if diags.HasError() {
		return nil, diags
	}

	matcher := &trino.CatalogAccessRuleMatcher{}
	switch {
	case len(model.IDs.Elements()) > 0:
		IDs := make([]string, 0, len(model.IDs.Elements()))
		diags.Append(model.IDs.ElementsAs(ctx, &IDs, false)...)
		matcher.MatchBy = &trino.CatalogAccessRuleMatcher_Ids{Ids: &trino.CatalogAccessRuleMatcher_CatalogIds{Any: IDs}}
	case model.NameRegexp.ValueString() != "":
		matcher.MatchBy = &trino.CatalogAccessRuleMatcher_NameRegexp{NameRegexp: model.NameRegexp.ValueString()}
	}
	return matcher, diags
}

func stringListModelToAPI(ctx context.Context, v types.List) ([]string, diag.Diagnostics) {
	s := make([]string, 0, len(v.Elements()))
	diags := v.ElementsAs(ctx, &s, false)
	return s, diags
}

func catalogPermissionStringToAPI(p string) (trino.CatalogAccessRule_Permission, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch CatalogPermission(p) {
	case CatalogPermissionAll:
		return trino.CatalogAccessRule_ALL, diags
	case CatalogPermissionNone:
		return trino.CatalogAccessRule_NONE, diags
	case CatalogPermissionReadOnly:
		return trino.CatalogAccessRule_READ_ONLY, diags
	default:
		diags.AddError("Invalid attribute.", fmt.Sprintf("Unknown catalog permission %q", p))
		return trino.CatalogAccessRule_PERMISSION_UNSPECIFIED, diags
	}
}

func schemaRuleModelToAPI(ctx context.Context, ruleObj types.Object) (*trino.SchemaAccessRule, diag.Diagnostics) {
	var diags diag.Diagnostics
	if ruleObj.IsNull() || ruleObj.IsUnknown() {
		return nil, diags
	}

	var model SchemaRule
	diags.Append(ruleObj.As(ctx, &model, baseOptions)...)
	if diags.HasError() {
		return nil, diags
	}

	catalog, dd := catalogMatcherModelToAPI(ctx, model.Catalog)
	diags.Append(dd...)
	schema, dd := schemaMatcherModelToAPI(ctx, model.Schema)
	diags.Append(dd...)
	users, dd := stringListModelToAPI(ctx, model.Users)
	diags.Append(dd...)
	groups, dd := stringListModelToAPI(ctx, model.Groups)
	diags.Append(dd...)
	owner, dd := schemaOwnerStringToAPI(model.Owner.ValueString())
	diags.Append(dd...)
	description := model.Description.ValueString()

	return &trino.SchemaAccessRule{
		Catalog:     catalog,
		Schema:      schema,
		Users:       users,
		Groups:      groups,
		Owner:       owner,
		Description: description,
	}, diags
}

func schemaMatcherModelToAPI(ctx context.Context, matcherObj types.Object) (*trino.SchemaAccessRuleMatcher, diag.Diagnostics) {
	var diags diag.Diagnostics
	if matcherObj.IsNull() || matcherObj.IsUnknown() {
		return nil, diags
	}

	var model NameMatcherModel
	diags.Append(matcherObj.As(ctx, &model, baseOptions)...)
	if diags.HasError() {
		return nil, diags
	}

	matcher := &trino.SchemaAccessRuleMatcher{}
	switch {
	case len(model.Names.Elements()) > 0:
		names := make([]string, 0, len(model.Names.Elements()))
		diags.Append(model.Names.ElementsAs(ctx, &names, false)...)
		matcher.MatchBy = &trino.SchemaAccessRuleMatcher_Names{Names: &trino.SchemaAccessRuleMatcher_SchemaNames{Any: names}}
	case model.NameRegexp.ValueString() != "":
		matcher.MatchBy = &trino.SchemaAccessRuleMatcher_NameRegexp{NameRegexp: model.NameRegexp.ValueString()}
	}
	return matcher, diags
}

func schemaOwnerStringToAPI(owner string) (trino.SchemaAccessRule_Owner, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch SchemaOwner(owner) {
	case SchemaOwnerNo:
		return trino.SchemaAccessRule_NO, diags
	case SchemaOwnerYes:
		return trino.SchemaAccessRule_YES, diags
	default:
		diags.AddError("Invalid attribute.", fmt.Sprintf("Unknown schema owner %q", owner))
		return trino.SchemaAccessRule_OWNER_UNSPECIFIED, diags
	}
}

func functionRuleModelToAPI(ctx context.Context, ruleObj types.Object) (*trino.FunctionAccessRule, diag.Diagnostics) {
	var diags diag.Diagnostics
	if ruleObj.IsNull() || ruleObj.IsUnknown() {
		return nil, diags
	}

	var model FunctionRule
	diags.Append(ruleObj.As(ctx, &model, baseOptions)...)
	if diags.HasError() {
		return nil, diags
	}

	catalog, dd := catalogMatcherModelToAPI(ctx, model.Catalog)
	diags.Append(dd...)
	schema, dd := schemaMatcherModelToAPI(ctx, model.Schema)
	diags.Append(dd...)
	function, dd := functionMatcherModelToAPI(ctx, model.Function)
	diags.Append(dd...)
	users, dd := stringListModelToAPI(ctx, model.Users)
	diags.Append(dd...)
	groups, dd := stringListModelToAPI(ctx, model.Groups)
	diags.Append(dd...)
	privileges, dd := functionPrivilegesModelToAPI(ctx, model.Privileges)
	diags.Append(dd...)
	description := model.Description.ValueString()

	return &trino.FunctionAccessRule{
		Catalog:     catalog,
		Schema:      schema,
		Function:    function,
		Users:       users,
		Groups:      groups,
		Privileges:  privileges,
		Description: description,
	}, diags
}

func functionMatcherModelToAPI(ctx context.Context, matcherObj types.Object) (*trino.FunctionAccessRuleMatcher, diag.Diagnostics) {
	var diags diag.Diagnostics
	if matcherObj.IsNull() || matcherObj.IsUnknown() {
		return nil, diags
	}

	var model NameMatcherModel
	diags.Append(matcherObj.As(ctx, &model, baseOptions)...)
	if diags.HasError() {
		return nil, diags
	}

	matcher := &trino.FunctionAccessRuleMatcher{}
	switch {
	case len(model.Names.Elements()) > 0:
		names := make([]string, 0, len(model.Names.Elements()))
		diags.Append(model.Names.ElementsAs(ctx, &names, false)...)
		matcher.MatchBy = &trino.FunctionAccessRuleMatcher_Names{Names: &trino.FunctionAccessRuleMatcher_FunctionNames{Any: names}}
	case model.NameRegexp.ValueString() != "":
		matcher.MatchBy = &trino.FunctionAccessRuleMatcher_NameRegexp{NameRegexp: model.NameRegexp.ValueString()}
	}
	return matcher, diags
}

func functionPrivilegesModelToAPI(ctx context.Context, privilegesList types.List) ([]trino.FunctionAccessRule_Privilege, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(privilegesList.Elements()) == 0 {
		return nil, diags
	}

	privileges := make([]string, 0, len(privilegesList.Elements()))
	diags.Append(privilegesList.ElementsAs(ctx, &privileges, false)...)

	result := make([]trino.FunctionAccessRule_Privilege, 0, len(privileges))
	for _, p := range privileges {
		switch FunctionPrivilege(p) {
		case FunctionPrivilegeExecute:
			result = append(result, trino.FunctionAccessRule_EXECUTE)
		case FunctionPrivilegeGrantExecute:
			result = append(result, trino.FunctionAccessRule_GRANT_EXECUTE)
		case FunctionPrivilegeOwnership:
			result = append(result, trino.FunctionAccessRule_OWNERSHIP)
		default:
			diags.AddError("Invalid attribute.", fmt.Sprintf("Unknown function privilege %q", p))
		}
	}
	return result, diags
}

func procedureRuleModelToAPI(ctx context.Context, ruleObj types.Object) (*trino.ProcedureAccessRule, diag.Diagnostics) {
	var diags diag.Diagnostics
	if ruleObj.IsNull() || ruleObj.IsUnknown() {
		return nil, diags
	}

	var model ProcedureRule
	diags.Append(ruleObj.As(ctx, &model, baseOptions)...)
	if diags.HasError() {
		return nil, diags
	}

	catalog, dd := catalogMatcherModelToAPI(ctx, model.Catalog)
	diags.Append(dd...)
	schema, dd := schemaMatcherModelToAPI(ctx, model.Schema)
	diags.Append(dd...)
	procedure, dd := procedureMatcherModelToAPI(ctx, model.Procedure)
	diags.Append(dd...)
	users, dd := stringListModelToAPI(ctx, model.Users)
	diags.Append(dd...)
	groups, dd := stringListModelToAPI(ctx, model.Groups)
	diags.Append(dd...)
	privileges, dd := procedurePrivilegesModelToAPI(ctx, model.Privileges)
	diags.Append(dd...)
	description := model.Description.ValueString()

	return &trino.ProcedureAccessRule{
		Catalog:     catalog,
		Schema:      schema,
		Procedure:   procedure,
		Users:       users,
		Groups:      groups,
		Privileges:  privileges,
		Description: description,
	}, diags
}

func procedureMatcherModelToAPI(ctx context.Context, matcherObj types.Object) (*trino.ProcedureAccessRuleMatcher, diag.Diagnostics) {
	var diags diag.Diagnostics
	if matcherObj.IsNull() || matcherObj.IsUnknown() {
		return nil, diags
	}

	var model NameMatcherModel
	diags.Append(matcherObj.As(ctx, &model, baseOptions)...)
	if diags.HasError() {
		return nil, diags
	}

	matcher := &trino.ProcedureAccessRuleMatcher{}
	switch {
	case len(model.Names.Elements()) > 0:
		names := make([]string, 0, len(model.Names.Elements()))
		diags.Append(model.Names.ElementsAs(ctx, &names, false)...)
		matcher.MatchBy = &trino.ProcedureAccessRuleMatcher_Names{Names: &trino.ProcedureAccessRuleMatcher_ProcedureNames{Any: names}}
	case model.NameRegexp.ValueString() != "":
		matcher.MatchBy = &trino.ProcedureAccessRuleMatcher_NameRegexp{NameRegexp: model.NameRegexp.ValueString()}
	}
	return matcher, diags
}

func procedurePrivilegesModelToAPI(ctx context.Context, privilegesList types.List) ([]trino.ProcedureAccessRule_Privilege, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(privilegesList.Elements()) == 0 {
		return nil, diags
	}

	privileges := make([]string, 0, len(privilegesList.Elements()))
	diags.Append(privilegesList.ElementsAs(ctx, &privileges, false)...)

	result := make([]trino.ProcedureAccessRule_Privilege, 0, len(privileges))
	for _, p := range privileges {
		switch ProcedurePrivilege(p) {
		case ProcedurePrivilegeExecute:
			result = append(result, trino.ProcedureAccessRule_EXECUTE)
		default:
			diags.AddError("Invalid attribute.", fmt.Sprintf("Unknown procedure privilege %q", p))
		}
	}
	return result, diags
}

func systemSessionPropertyRuleModelToAPI(ctx context.Context, ruleObj types.Object) (*trino.SystemSessionPropertyAccessRule, diag.Diagnostics) {
	var diags diag.Diagnostics
	if ruleObj.IsNull() || ruleObj.IsUnknown() {
		return nil, diags
	}

	var model SystemSessionPropertyRule
	diags.Append(ruleObj.As(ctx, &model, baseOptions)...)
	if diags.HasError() {
		return nil, diags
	}

	property, dd := propertyMatcherModelToAPI(ctx, model.Property)
	diags.Append(dd...)
	users, dd := stringListModelToAPI(ctx, model.Users)
	diags.Append(dd...)
	groups, dd := stringListModelToAPI(ctx, model.Groups)
	diags.Append(dd...)
	allow, dd := systemPropertyAllowStringToAPI(model.Allow.ValueString())
	diags.Append(dd...)
	description := model.Description.ValueString()

	return &trino.SystemSessionPropertyAccessRule{
		Property:    property,
		Users:       users,
		Groups:      groups,
		Allow:       allow,
		Description: description,
	}, diags
}

func catalogSessionPropertyRuleModelToAPI(ctx context.Context, ruleObj types.Object) (*trino.CatalogSessionPropertyAccessRule, diag.Diagnostics) {
	var diags diag.Diagnostics
	if ruleObj.IsNull() || ruleObj.IsUnknown() {
		return nil, diags
	}

	var model CatalogSessionPropertyRule
	diags.Append(ruleObj.As(ctx, &model, baseOptions)...)
	if diags.HasError() {
		return nil, diags
	}

	catalog, dd := catalogMatcherModelToAPI(ctx, model.Catalog)
	diags.Append(dd...)
	property, dd := propertyMatcherModelToAPI(ctx, model.Property)
	diags.Append(dd...)
	users, dd := stringListModelToAPI(ctx, model.Users)
	diags.Append(dd...)
	groups, dd := stringListModelToAPI(ctx, model.Groups)
	diags.Append(dd...)
	allow, dd := catalogPropertyAllowStringToAPI(model.Allow.ValueString())
	diags.Append(dd...)
	description := model.Description.ValueString()

	return &trino.CatalogSessionPropertyAccessRule{
		Catalog:     catalog,
		Property:    property,
		Users:       users,
		Groups:      groups,
		Allow:       allow,
		Description: description,
	}, diags
}

func propertyMatcherModelToAPI(ctx context.Context, matcherObj types.Object) (*trino.PropertyAccessRuleMatcher, diag.Diagnostics) {
	var diags diag.Diagnostics
	if matcherObj.IsNull() || matcherObj.IsUnknown() {
		return nil, diags
	}

	var model NameMatcherModel
	diags.Append(matcherObj.As(ctx, &model, baseOptions)...)
	if diags.HasError() {
		return nil, diags
	}

	matcher := &trino.PropertyAccessRuleMatcher{}
	switch {
	case len(model.Names.Elements()) > 0:
		names := make([]string, 0, len(model.Names.Elements()))
		diags.Append(model.Names.ElementsAs(ctx, &names, false)...)
		matcher.MatchBy = &trino.PropertyAccessRuleMatcher_Names{Names: &trino.PropertyAccessRuleMatcher_PropertyNames{Any: names}}
	case model.NameRegexp.ValueString() != "":
		matcher.MatchBy = &trino.PropertyAccessRuleMatcher_NameRegexp{NameRegexp: model.NameRegexp.ValueString()}
	}
	return matcher, diags
}

func systemPropertyAllowStringToAPI(allow string) (trino.SystemSessionPropertyAccessRule_Allow, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch PropertyAllow(allow) {
	case PropertyAllowNo:
		return trino.SystemSessionPropertyAccessRule_NO, diags
	case PropertyAllowYes:
		return trino.SystemSessionPropertyAccessRule_YES, diags
	default:
		diags.AddError("Invalid attribute.", fmt.Sprintf("Unknown system property allow %q", allow))
		return trino.SystemSessionPropertyAccessRule_ALLOW_UNSPECIFIED, diags
	}
}

func queryRuleModelToAPI(ctx context.Context, ruleObj types.Object) (*trino.QueryAccessRule, diag.Diagnostics) {
	var diags diag.Diagnostics
	if ruleObj.IsNull() || ruleObj.IsUnknown() {
		return nil, diags
	}

	var model QueryRule
	diags.Append(ruleObj.As(ctx, &model, baseOptions)...)
	if diags.HasError() {
		return nil, diags
	}

	users, dd := stringListModelToAPI(ctx, model.Users)
	diags.Append(dd...)
	groups, dd := stringListModelToAPI(ctx, model.Groups)
	diags.Append(dd...)
	queryOwners, dd := stringListModelToAPI(ctx, model.QueryOwners)
	diags.Append(dd...)
	privileges, dd := queryPrivilegesModelToAPI(ctx, model.Privileges)
	diags.Append(dd...)
	description := model.Description.ValueString()

	return &trino.QueryAccessRule{
		Users:       users,
		Groups:      groups,
		QueryOwners: queryOwners,
		Privileges:  privileges,
		Description: description,
	}, diags
}

func queryPrivilegesModelToAPI(ctx context.Context, privilegesList types.List) ([]trino.QueryAccessRule_Privilege, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(privilegesList.Elements()) == 0 {
		return nil, diags
	}

	privileges := make([]string, 0, len(privilegesList.Elements()))
	diags.Append(privilegesList.ElementsAs(ctx, &privileges, false)...)

	result := make([]trino.QueryAccessRule_Privilege, 0, len(privileges))
	for _, p := range privileges {
		switch QueryPrivilege(p) {
		case QueryPrivilegeView:
			result = append(result, trino.QueryAccessRule_VIEW)
		case QueryPrivilegeExecute:
			result = append(result, trino.QueryAccessRule_EXECUTE)
		case QueryPrivilegeKill:
			result = append(result, trino.QueryAccessRule_KILL)
		default:
			diags.AddError("Invalid attribute.", fmt.Sprintf("Unknown query privilege %q", p))
		}
	}
	return result, diags
}

func tableRuleModelToAPI(ctx context.Context, ruleObj types.Object) (*trino.TableAccessRule, diag.Diagnostics) {
	var diags diag.Diagnostics
	if ruleObj.IsNull() || ruleObj.IsUnknown() {
		return nil, diags
	}

	var model TableRule
	diags.Append(ruleObj.As(ctx, &model, baseOptions)...)
	if diags.HasError() {
		return nil, diags
	}

	catalog, dd := catalogMatcherModelToAPI(ctx, model.Catalog)
	diags.Append(dd...)
	schema, dd := schemaMatcherModelToAPI(ctx, model.Schema)
	diags.Append(dd...)
	table, dd := tableMatcherModelToAPI(ctx, model.Table)
	diags.Append(dd...)
	users, dd := stringListModelToAPI(ctx, model.Users)
	diags.Append(dd...)
	groups, dd := stringListModelToAPI(ctx, model.Groups)
	diags.Append(dd...)
	privileges, dd := tablePrivilegesModelToAPI(ctx, model.Privileges)
	diags.Append(dd...)
	columns, dd := columnRulesModelToAPI(ctx, model.Columns)
	diags.Append(dd...)
	filter := model.Filter.ValueString()
	description := model.Description.ValueString()

	return &trino.TableAccessRule{
		Catalog:     catalog,
		Schema:      schema,
		Table:       table,
		Users:       users,
		Groups:      groups,
		Privileges:  privileges,
		Columns:     columns,
		Filter:      filter,
		Description: description,
	}, diags
}

func tableMatcherModelToAPI(ctx context.Context, matcherObj types.Object) (*trino.TableAccessRuleMatcher, diag.Diagnostics) {
	var diags diag.Diagnostics
	if matcherObj.IsNull() || matcherObj.IsUnknown() {
		return nil, diags
	}

	var model NameMatcherModel
	diags.Append(matcherObj.As(ctx, &model, baseOptions)...)
	if diags.HasError() {
		return nil, diags
	}

	matcher := &trino.TableAccessRuleMatcher{}
	switch {
	case len(model.Names.Elements()) > 0:
		names := make([]string, 0, len(model.Names.Elements()))
		diags.Append(model.Names.ElementsAs(ctx, &names, false)...)
		matcher.MatchBy = &trino.TableAccessRuleMatcher_Names{Names: &trino.TableAccessRuleMatcher_TableNames{Any: names}}
	case model.NameRegexp.ValueString() != "":
		matcher.MatchBy = &trino.TableAccessRuleMatcher_NameRegexp{NameRegexp: model.NameRegexp.ValueString()}
	}
	return matcher, diags
}

func tablePrivilegesModelToAPI(ctx context.Context, privilegesList types.List) ([]trino.TableAccessRule_Privilege, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(privilegesList.Elements()) == 0 {
		return nil, diags
	}

	privileges := make([]string, 0, len(privilegesList.Elements()))
	diags.Append(privilegesList.ElementsAs(ctx, &privileges, false)...)

	result := make([]trino.TableAccessRule_Privilege, 0, len(privileges))
	for _, p := range privileges {
		switch TablePrivilege(p) {
		case TablePrivilegeSelect:
			result = append(result, trino.TableAccessRule_SELECT)
		case TablePrivilegeInsert:
			result = append(result, trino.TableAccessRule_INSERT)
		case TablePrivilegeDelete:
			result = append(result, trino.TableAccessRule_DELETE)
		case TablePrivilegeUpdate:
			result = append(result, trino.TableAccessRule_UPDATE)
		case TablePrivilegeOwnership:
			result = append(result, trino.TableAccessRule_OWNERSHIP)
		case TablePrivilegeGrantSelect:
			result = append(result, trino.TableAccessRule_GRANT_SELECT)
		default:
			diags.AddError("Invalid attribute.", fmt.Sprintf("Unknown table privilege %q", p))
		}
	}
	return result, diags
}

func columnRulesModelToAPI(ctx context.Context, columnsList types.List) ([]*trino.TableAccessRule_Column, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(columnsList.Elements()) == 0 {
		return nil, diags
	}

	var columnObjs []types.Object
	diags.Append(columnsList.ElementsAs(ctx, &columnObjs, false)...)
	if diags.HasError() {
		return nil, diags
	}

	result := make([]*trino.TableAccessRule_Column, 0, len(columnObjs))
	for _, colObj := range columnObjs {
		if colObj.IsNull() || colObj.IsUnknown() {
			continue
		}

		var col ColumnRule
		diags.Append(colObj.As(ctx, &col, baseOptions)...)
		if diags.HasError() {
			continue
		}

		column := &trino.TableAccessRule_Column{
			Name: col.Name.ValueString(),
			Mask: col.Mask.ValueString(),
		}
		switch ColumnAccessMode(col.Access.ValueString()) {
		case ColumnAccessModeNone:
			column.Access = trino.TableAccessRule_Column_NONE
		case ColumnAccessModeAll:
			column.Access = trino.TableAccessRule_Column_ALL
		default:
			diags.AddError("Invalid attribute.", fmt.Sprintf("Unknown column access mode %q", col.Access.ValueString()))
		}
		result = append(result, column)
	}
	return result, diags
}

func catalogPropertyAllowStringToAPI(allow string) (trino.CatalogSessionPropertyAccessRule_Allow, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch PropertyAllow(allow) {
	case PropertyAllowNo:
		return trino.CatalogSessionPropertyAccessRule_NO, diags
	case PropertyAllowYes:
		return trino.CatalogSessionPropertyAccessRule_YES, diags
	default:
		diags.AddError("Invalid attribute.", fmt.Sprintf("Unknown catalog property allow %q", allow))
		return trino.CatalogSessionPropertyAccessRule_ALLOW_UNSPECIFIED, diags
	}
}
