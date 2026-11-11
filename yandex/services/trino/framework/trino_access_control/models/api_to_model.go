package models

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/trino/v1"
)

func FromAPI(ctx context.Context, clusterID string, accessControl *trino.AccessControlConfig) (AccessControlModel, diag.Diagnostics) {
	model := AccessControlModel{ClusterId: types.StringValue(clusterID)}
	if accessControl == nil {
		return model, nil
	}
	var diags diag.Diagnostics
	if len(accessControl.Catalogs) > 0 {
		catalogObjs := make([]attr.Value, 0, len(accessControl.Catalogs))
		for _, rule := range accessControl.Catalogs {
			v, dd := catalogRuleToModel(ctx, rule)
			diags.Append(dd...)
			catalogObjs = append(catalogObjs, v)
		}
		model.Catalogs, _ = types.ListValue(CatalogRuleT, catalogObjs)
	}
	if len(accessControl.Schemas) > 0 {
		schemaObjs := make([]attr.Value, 0, len(accessControl.Schemas))
		for _, rule := range accessControl.Schemas {
			v, dd := schemaRuleToModel(ctx, rule)
			diags.Append(dd...)
			schemaObjs = append(schemaObjs, v)
		}
		model.Schemas, _ = types.ListValue(SchemaRuleT, schemaObjs)
	}
	if len(accessControl.Functions) > 0 {
		functionObjs := make([]attr.Value, 0, len(accessControl.Functions))
		for _, rule := range accessControl.Functions {
			v, dd := functionRuleToModel(ctx, rule)
			diags.Append(dd...)
			functionObjs = append(functionObjs, v)
		}
		model.Functions, _ = types.ListValue(FunctionRuleT, functionObjs)
	}
	if len(accessControl.Procedures) > 0 {
		procedureObjs := make([]attr.Value, 0, len(accessControl.Procedures))
		for _, rule := range accessControl.Procedures {
			v, dd := procedureRuleToModel(ctx, rule)
			diags.Append(dd...)
			procedureObjs = append(procedureObjs, v)
		}
		model.Procedures, _ = types.ListValue(ProcedureRuleT, procedureObjs)
	}
	if len(accessControl.Tables) > 0 {
		tableObjs := make([]attr.Value, 0, len(accessControl.Tables))
		for _, rule := range accessControl.Tables {
			v, dd := tableRuleToModel(ctx, rule)
			diags.Append(dd...)
			tableObjs = append(tableObjs, v)
		}
		model.Tables, _ = types.ListValue(TableRuleT, tableObjs)
	}
	if len(accessControl.Queries) > 0 {
		queryObjs := make([]attr.Value, 0, len(accessControl.Queries))
		for _, rule := range accessControl.Queries {
			v, dd := queryRuleToModel(ctx, rule)
			diags.Append(dd...)
			queryObjs = append(queryObjs, v)
		}
		model.Queries, _ = types.ListValue(QueryRuleT, queryObjs)
	}
	if len(accessControl.SystemSessionProperties) > 0 {
		sysPropObjs := make([]attr.Value, 0, len(accessControl.SystemSessionProperties))
		for _, rule := range accessControl.SystemSessionProperties {
			v, dd := systemSessionPropertyRuleToModel(ctx, rule)
			diags.Append(dd...)
			sysPropObjs = append(sysPropObjs, v)
		}
		model.SystemSessionProperties, _ = types.ListValue(SystemSessionPropertyRuleT, sysPropObjs)
	}
	if len(accessControl.CatalogSessionProperties) > 0 {
		catPropObjs := make([]attr.Value, 0, len(accessControl.CatalogSessionProperties))
		for _, rule := range accessControl.CatalogSessionProperties {
			v, dd := catalogSessionPropertyRuleToModel(ctx, rule)
			diags.Append(dd...)
			catPropObjs = append(catPropObjs, v)
		}
		model.CatalogSessionProperties, _ = types.ListValue(CatalogSessionPropertyRuleT, catPropObjs)
	}
	return model, diags
}

func catalogRuleToModel(ctx context.Context, rule *trino.CatalogAccessRule) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	if rule == nil {
		return types.ObjectNull(CatalogRuleT.AttrTypes), diags
	}
	var dd diag.Diagnostics
	model := CatalogRule{}
	model.Catalog, dd = catalogMatcherToModel(ctx, rule.Catalog)
	diags.Append(dd...)
	model.Users, dd = types.ListValueFrom(ctx, types.StringType, rule.Users)
	diags.Append(dd...)
	model.Groups, dd = types.ListValueFrom(ctx, types.StringType, rule.Groups)
	diags.Append(dd...)
	model.Permission, dd = catalogPermissionToModel(rule.Permission)
	diags.Append(dd...)
	model.Description = stringToModel(rule.Description)

	obj, dd := types.ObjectValueFrom(ctx, CatalogRuleT.AttrTypes, model)
	diags.Append(dd...)
	return obj, diags
}

func catalogMatcherToModel(ctx context.Context, matcher *trino.CatalogAccessRuleMatcher) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	if matcher == nil {
		return types.ObjectNull(CatalogMatcherT.AttrTypes), diags
	}
	ids, dd := types.ListValueFrom(ctx, types.StringType, matcher.GetIds().GetAny())
	diags.Append(dd...)
	model := CatalogMatcherModel{
		IDs:        ids,
		NameRegexp: stringToModel(matcher.GetNameRegexp()),
	}
	obj, dd := types.ObjectValueFrom(ctx, CatalogMatcherT.AttrTypes, model)
	diags.Append(dd...)
	return obj, diags
}

func catalogPermissionToModel(p trino.CatalogAccessRule_Permission) (types.String, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch p {
	case trino.CatalogAccessRule_NONE:
		return types.StringValue(string(CatalogPermissionNone)), nil
	case trino.CatalogAccessRule_ALL:
		return types.StringValue(string(CatalogPermissionAll)), nil
	case trino.CatalogAccessRule_READ_ONLY:
		return types.StringValue(string(CatalogPermissionReadOnly)), nil
	default:
		diags.AddError("Invalid catalog permission", fmt.Sprintf("Unknown catalog permission %v", p))
		return types.StringUnknown(), diags
	}
}

func schemaRuleToModel(ctx context.Context, rule *trino.SchemaAccessRule) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	if rule == nil {
		return types.ObjectNull(SchemaRuleT.AttrTypes), diags
	}
	var dd diag.Diagnostics
	var d diag.Diagnostic
	model := SchemaRule{}
	model.Catalog, dd = catalogMatcherToModel(ctx, rule.Catalog)
	diags.Append(dd...)
	model.Schema, dd = schemaMatcherToModel(ctx, rule.Schema)
	diags.Append(dd...)
	model.Users, dd = types.ListValueFrom(ctx, types.StringType, rule.Users)
	diags.Append(dd...)
	model.Groups, dd = types.ListValueFrom(ctx, types.StringType, rule.Groups)
	diags.Append(dd...)
	model.Owner, d = schemaOwnerToModel(rule.Owner)
	diags.Append(d)
	model.Description = stringToModel(rule.Description)

	obj, dd := types.ObjectValueFrom(ctx, SchemaRuleT.AttrTypes, model)
	diags.Append(dd...)
	return obj, diags
}

func schemaMatcherToModel(ctx context.Context, matcher *trino.SchemaAccessRuleMatcher) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	if matcher == nil {
		return types.ObjectNull(NameMatcherT.AttrTypes), diags
	}
	names, dd := types.ListValueFrom(ctx, types.StringType, matcher.GetNames().GetAny())
	diags.Append(dd...)
	model := NameMatcherModel{
		Names:      names,
		NameRegexp: stringToModel(matcher.GetNameRegexp()),
	}
	obj, dd := types.ObjectValueFrom(ctx, NameMatcherT.AttrTypes, model)
	diags.Append(dd...)
	return obj, diags
}

func schemaOwnerToModel(owner trino.SchemaAccessRule_Owner) (types.String, diag.Diagnostic) {
	switch owner {
	case trino.SchemaAccessRule_NO:
		return types.StringValue(string(SchemaOwnerNo)), nil
	case trino.SchemaAccessRule_YES:
		return types.StringValue(string(SchemaOwnerYes)), nil
	default:
		return types.StringUnknown(), diag.NewErrorDiagnostic("Invalid schema owner", fmt.Sprintf("Unknown schema owner %v", owner))
	}
}

func functionRuleToModel(ctx context.Context, rule *trino.FunctionAccessRule) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	if rule == nil {
		return types.ObjectNull(FunctionRuleT.AttrTypes), diags
	}
	var dd diag.Diagnostics
	model := FunctionRule{}
	model.Catalog, dd = catalogMatcherToModel(ctx, rule.Catalog)
	diags.Append(dd...)
	model.Schema, dd = schemaMatcherToModel(ctx, rule.Schema)
	diags.Append(dd...)
	model.Function, dd = functionMatcherToModel(ctx, rule.Function)
	diags.Append(dd...)
	model.Users, dd = types.ListValueFrom(ctx, types.StringType, rule.Users)
	diags.Append(dd...)
	model.Groups, dd = types.ListValueFrom(ctx, types.StringType, rule.Groups)
	diags.Append(dd...)
	model.Privileges, dd = functionPrivilegesToModel(ctx, rule.Privileges)
	diags.Append(dd...)
	model.Description = stringToModel(rule.Description)

	obj, dd := types.ObjectValueFrom(ctx, FunctionRuleT.AttrTypes, model)
	diags.Append(dd...)
	return obj, diags
}

func functionMatcherToModel(ctx context.Context, matcher *trino.FunctionAccessRuleMatcher) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	if matcher == nil {
		return types.ObjectNull(NameMatcherT.AttrTypes), diags
	}
	names, dd := types.ListValueFrom(ctx, types.StringType, matcher.GetNames().GetAny())
	diags.Append(dd...)
	model := NameMatcherModel{
		Names:      names,
		NameRegexp: stringToModel(matcher.GetNameRegexp()),
	}
	obj, dd := types.ObjectValueFrom(ctx, NameMatcherT.AttrTypes, model)
	diags.Append(dd...)
	return obj, diags
}

func functionPrivilegesToModel(ctx context.Context, privileges []trino.FunctionAccessRule_Privilege) (types.List, diag.Diagnostics) {
	if len(privileges) == 0 {
		return types.ListNull(types.StringType), nil
	}
	var diags diag.Diagnostics
	vals := make([]attr.Value, 0, len(privileges))
	for _, p := range privileges {
		switch p {
		case trino.FunctionAccessRule_EXECUTE:
			vals = append(vals, types.StringValue(string(FunctionPrivilegeExecute)))
		case trino.FunctionAccessRule_GRANT_EXECUTE:
			vals = append(vals, types.StringValue(string(FunctionPrivilegeGrantExecute)))
		case trino.FunctionAccessRule_OWNERSHIP:
			vals = append(vals, types.StringValue(string(FunctionPrivilegeOwnership)))
		default:
			diags.AddError("Invalid function privilege", fmt.Sprintf("Unknown function privilege %v", p))
		}
	}
	privs, dd := types.ListValueFrom(ctx, types.StringType, vals)
	diags.Append(dd...)
	return privs, diags
}

func procedureRuleToModel(ctx context.Context, rule *trino.ProcedureAccessRule) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	if rule == nil {
		return types.ObjectNull(ProcedureRuleT.AttrTypes), diags
	}
	var dd diag.Diagnostics
	model := ProcedureRule{}
	model.Catalog, dd = catalogMatcherToModel(ctx, rule.Catalog)
	diags.Append(dd...)
	model.Schema, dd = schemaMatcherToModel(ctx, rule.Schema)
	diags.Append(dd...)
	model.Procedure, dd = procedureMatcherToModel(ctx, rule.Procedure)
	diags.Append(dd...)
	model.Users, dd = types.ListValueFrom(ctx, types.StringType, rule.Users)
	diags.Append(dd...)
	model.Groups, dd = types.ListValueFrom(ctx, types.StringType, rule.Groups)
	diags.Append(dd...)
	model.Privileges, dd = procedurePrivilegesToModel(ctx, rule.Privileges)
	diags.Append(dd...)
	model.Description = stringToModel(rule.Description)

	obj, dd := types.ObjectValueFrom(ctx, ProcedureRuleT.AttrTypes, model)
	diags.Append(dd...)
	return obj, diags
}

func procedureMatcherToModel(ctx context.Context, matcher *trino.ProcedureAccessRuleMatcher) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	if matcher == nil {
		return types.ObjectNull(NameMatcherT.AttrTypes), diags
	}
	names, dd := types.ListValueFrom(ctx, types.StringType, matcher.GetNames().GetAny())
	diags.Append(dd...)
	model := NameMatcherModel{
		Names:      names,
		NameRegexp: stringToModel(matcher.GetNameRegexp()),
	}
	obj, dd := types.ObjectValueFrom(ctx, NameMatcherT.AttrTypes, model)
	diags.Append(dd...)
	return obj, diags
}

func procedurePrivilegesToModel(ctx context.Context, privileges []trino.ProcedureAccessRule_Privilege) (types.List, diag.Diagnostics) {
	if len(privileges) == 0 {
		return types.ListNull(types.StringType), nil
	}
	var diags diag.Diagnostics
	vals := make([]attr.Value, 0, len(privileges))
	for _, p := range privileges {
		switch p {
		case trino.ProcedureAccessRule_EXECUTE:
			vals = append(vals, types.StringValue(string(ProcedurePrivilegeExecute)))
		default:
			diags.AddError("Invalid procedure privilege", fmt.Sprintf("Unknown procedure privilege %v", p))
		}
	}
	privs, dd := types.ListValueFrom(ctx, types.StringType, vals)
	diags.Append(dd...)
	return privs, diags
}

func systemSessionPropertyRuleToModel(ctx context.Context, rule *trino.SystemSessionPropertyAccessRule) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	if rule == nil {
		return types.ObjectNull(SystemSessionPropertyRuleT.AttrTypes), diags
	}
	var dd diag.Diagnostics
	var d diag.Diagnostic
	model := SystemSessionPropertyRule{}
	model.Property, dd = propertyMatcherToModel(ctx, rule.Property)
	diags.Append(dd...)
	model.Users, dd = types.ListValueFrom(ctx, types.StringType, rule.Users)
	diags.Append(dd...)
	model.Groups, dd = types.ListValueFrom(ctx, types.StringType, rule.Groups)
	diags.Append(dd...)
	model.Allow, d = systemPropertyAllowToModel(rule.Allow)
	diags.Append(d)
	model.Description = stringToModel(rule.Description)

	obj, dd := types.ObjectValueFrom(ctx, SystemSessionPropertyRuleT.AttrTypes, model)
	diags.Append(dd...)
	return obj, diags
}

func catalogSessionPropertyRuleToModel(ctx context.Context, rule *trino.CatalogSessionPropertyAccessRule) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	if rule == nil {
		return types.ObjectNull(CatalogSessionPropertyRuleT.AttrTypes), diags
	}
	var dd diag.Diagnostics
	var d diag.Diagnostic
	model := CatalogSessionPropertyRule{}
	model.Catalog, dd = catalogMatcherToModel(ctx, rule.Catalog)
	diags.Append(dd...)
	model.Property, dd = propertyMatcherToModel(ctx, rule.Property)
	diags.Append(dd...)
	model.Users, dd = types.ListValueFrom(ctx, types.StringType, rule.Users)
	diags.Append(dd...)
	model.Groups, dd = types.ListValueFrom(ctx, types.StringType, rule.Groups)
	diags.Append(dd...)
	model.Allow, d = catalogPropertyAllowToModel(rule.Allow)
	diags.Append(d)
	model.Description = stringToModel(rule.Description)

	obj, dd := types.ObjectValueFrom(ctx, CatalogSessionPropertyRuleT.AttrTypes, model)
	diags.Append(dd...)
	return obj, diags
}

func propertyMatcherToModel(ctx context.Context, matcher *trino.PropertyAccessRuleMatcher) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	if matcher == nil {
		return types.ObjectNull(NameMatcherT.AttrTypes), diags
	}
	names, dd := types.ListValueFrom(ctx, types.StringType, matcher.GetNames().GetAny())
	diags.Append(dd...)
	model := NameMatcherModel{
		Names:      names,
		NameRegexp: stringToModel(matcher.GetNameRegexp()),
	}
	obj, dd := types.ObjectValueFrom(ctx, NameMatcherT.AttrTypes, model)
	diags.Append(dd...)
	return obj, diags
}

func systemPropertyAllowToModel(allow trino.SystemSessionPropertyAccessRule_Allow) (types.String, diag.Diagnostic) {
	switch allow {
	case trino.SystemSessionPropertyAccessRule_NO:
		return types.StringValue(string(PropertyAllowNo)), nil
	case trino.SystemSessionPropertyAccessRule_YES:
		return types.StringValue(string(PropertyAllowYes)), nil
	default:
		return types.StringUnknown(), diag.NewErrorDiagnostic("Invalid property allow", fmt.Sprintf("Unknown property allow %v", allow))
	}
}

func queryRuleToModel(ctx context.Context, rule *trino.QueryAccessRule) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	if rule == nil {
		return types.ObjectNull(QueryRuleT.AttrTypes), diags
	}
	var dd diag.Diagnostics
	model := QueryRule{}
	model.Users, dd = types.ListValueFrom(ctx, types.StringType, rule.Users)
	diags.Append(dd...)
	model.Groups, dd = types.ListValueFrom(ctx, types.StringType, rule.Groups)
	diags.Append(dd...)
	model.QueryOwners, dd = types.ListValueFrom(ctx, types.StringType, rule.QueryOwners)
	diags.Append(dd...)
	model.Privileges, dd = queryPrivilegesToModel(ctx, rule.Privileges)
	diags.Append(dd...)
	model.Description = stringToModel(rule.Description)

	obj, dd := types.ObjectValueFrom(ctx, QueryRuleT.AttrTypes, model)
	diags.Append(dd...)
	return obj, diags
}

func queryPrivilegesToModel(ctx context.Context, privileges []trino.QueryAccessRule_Privilege) (types.List, diag.Diagnostics) {
	if len(privileges) == 0 {
		return types.ListNull(types.StringType), nil
	}
	var diags diag.Diagnostics
	vals := make([]attr.Value, 0, len(privileges))
	for _, p := range privileges {
		switch p {
		case trino.QueryAccessRule_VIEW:
			vals = append(vals, types.StringValue(string(QueryPrivilegeView)))
		case trino.QueryAccessRule_EXECUTE:
			vals = append(vals, types.StringValue(string(QueryPrivilegeExecute)))
		case trino.QueryAccessRule_KILL:
			vals = append(vals, types.StringValue(string(QueryPrivilegeKill)))
		default:
			diags.AddError("Invalid query privilege", fmt.Sprintf("Unknown query privilege %v", p))
		}
	}
	privs, dd := types.ListValueFrom(ctx, types.StringType, vals)
	diags.Append(dd...)
	return privs, diags
}

func tableRuleToModel(ctx context.Context, rule *trino.TableAccessRule) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	if rule == nil {
		return types.ObjectNull(TableRuleT.AttrTypes), diags
	}
	var dd diag.Diagnostics
	model := TableRule{}
	model.Catalog, dd = catalogMatcherToModel(ctx, rule.Catalog)
	diags.Append(dd...)
	model.Schema, dd = schemaMatcherToModel(ctx, rule.Schema)
	diags.Append(dd...)
	model.Table, dd = tableMatcherToModel(ctx, rule.Table)
	diags.Append(dd...)
	model.Users, dd = types.ListValueFrom(ctx, types.StringType, rule.Users)
	diags.Append(dd...)
	model.Groups, dd = types.ListValueFrom(ctx, types.StringType, rule.Groups)
	diags.Append(dd...)
	model.Privileges, dd = tablePrivilegesToModel(ctx, rule.Privileges)
	diags.Append(dd...)
	model.Columns, dd = columnRulesToModel(ctx, rule.Columns)
	diags.Append(dd...)
	model.Filter = stringToModel(rule.Filter)
	model.Description = stringToModel(rule.Description)

	obj, dd := types.ObjectValueFrom(ctx, TableRuleT.AttrTypes, model)
	diags.Append(dd...)
	return obj, diags
}

func tableMatcherToModel(ctx context.Context, matcher *trino.TableAccessRuleMatcher) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	if matcher == nil {
		return types.ObjectNull(NameMatcherT.AttrTypes), diags
	}
	names, dd := types.ListValueFrom(ctx, types.StringType, matcher.GetNames().GetAny())
	diags.Append(dd...)
	model := NameMatcherModel{
		Names:      names,
		NameRegexp: stringToModel(matcher.GetNameRegexp()),
	}
	obj, dd := types.ObjectValueFrom(ctx, NameMatcherT.AttrTypes, model)
	diags.Append(dd...)
	return obj, diags
}

func tablePrivilegesToModel(ctx context.Context, privileges []trino.TableAccessRule_Privilege) (types.List, diag.Diagnostics) {
	if len(privileges) == 0 {
		return types.ListNull(types.StringType), nil
	}
	var diags diag.Diagnostics
	vals := make([]attr.Value, 0, len(privileges))
	for _, p := range privileges {
		switch p {
		case trino.TableAccessRule_SELECT:
			vals = append(vals, types.StringValue(string(TablePrivilegeSelect)))
		case trino.TableAccessRule_INSERT:
			vals = append(vals, types.StringValue(string(TablePrivilegeInsert)))
		case trino.TableAccessRule_DELETE:
			vals = append(vals, types.StringValue(string(TablePrivilegeDelete)))
		case trino.TableAccessRule_UPDATE:
			vals = append(vals, types.StringValue(string(TablePrivilegeUpdate)))
		case trino.TableAccessRule_OWNERSHIP:
			vals = append(vals, types.StringValue(string(TablePrivilegeOwnership)))
		case trino.TableAccessRule_GRANT_SELECT:
			vals = append(vals, types.StringValue(string(TablePrivilegeGrantSelect)))
		default:
			diags.AddError("Invalid table privilege", fmt.Sprintf("Unknown table privilege %v", p))
		}
	}
	privs, dd := types.ListValueFrom(ctx, types.StringType, vals)
	diags.Append(dd...)
	return privs, diags
}

func columnRulesToModel(ctx context.Context, columns []*trino.TableAccessRule_Column) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(columns) == 0 {
		return types.ListNull(ColumnRuleT), diags
	}

	columnObjs := make([]attr.Value, 0, len(columns))
	for _, col := range columns {
		columnRule := ColumnRule{
			Name: types.StringValue(col.Name),
			Mask: stringToModel(col.Mask),
		}
		switch col.Access {
		case trino.TableAccessRule_Column_NONE:
			columnRule.Access = types.StringValue(string(ColumnAccessModeNone))
		case trino.TableAccessRule_Column_ALL:
			columnRule.Access = types.StringValue(string(ColumnAccessModeAll))
		default:
			diags.AddError("Invalid column access mode", fmt.Sprintf("Unknown column access mode %v", col.Access))
		}
		obj, dd := types.ObjectValueFrom(ctx, ColumnRuleT.AttrTypes, columnRule)
		diags.Append(dd...)
		columnObjs = append(columnObjs, obj)
	}

	list, dd := types.ListValue(ColumnRuleT, columnObjs)
	diags.Append(dd...)
	return list, diags
}

func catalogPropertyAllowToModel(allow trino.CatalogSessionPropertyAccessRule_Allow) (types.String, diag.Diagnostic) {
	switch allow {
	case trino.CatalogSessionPropertyAccessRule_NO:
		return types.StringValue(string(PropertyAllowNo)), nil
	case trino.CatalogSessionPropertyAccessRule_YES:
		return types.StringValue(string(PropertyAllowYes)), nil
	default:
		return types.StringUnknown(), diag.NewErrorDiagnostic("Invalid catalog property allow", fmt.Sprintf("Unknown catalog property allow %v", allow))
	}
}

func stringToModel(s string) types.String {
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}
