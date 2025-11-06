package models

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type CatalogPermission string

const (
	CatalogPermissionNone     CatalogPermission = "NONE"
	CatalogPermissionAll      CatalogPermission = "ALL"
	CatalogPermissionReadOnly CatalogPermission = "READ_ONLY"
)

type SchemaOwner string

const (
	SchemaOwnerNo  SchemaOwner = "NO"
	SchemaOwnerYes SchemaOwner = "YES"
)

type FunctionPrivilege string

const (
	FunctionPrivilegeExecute      FunctionPrivilege = "EXECUTE"
	FunctionPrivilegeGrantExecute FunctionPrivilege = "GRANT_EXECUTE"
	FunctionPrivilegeOwnership    FunctionPrivilege = "OWNERSHIP"
)

type ProcedurePrivilege string

const (
	ProcedurePrivilegeExecute ProcedurePrivilege = "EXECUTE"
)

type PropertyAllow string

const (
	PropertyAllowNo  PropertyAllow = "NO"
	PropertyAllowYes PropertyAllow = "YES"
)

type QueryPrivilege string

const (
	QueryPrivilegeView    QueryPrivilege = "VIEW"
	QueryPrivilegeExecute QueryPrivilege = "EXECUTE"
	QueryPrivilegeKill    QueryPrivilege = "KILL"
)

type TablePrivilege string

const (
	TablePrivilegeSelect      TablePrivilege = "SELECT"
	TablePrivilegeInsert      TablePrivilege = "INSERT"
	TablePrivilegeDelete      TablePrivilege = "DELETE"
	TablePrivilegeUpdate      TablePrivilege = "UPDATE"
	TablePrivilegeOwnership   TablePrivilege = "OWNERSHIP"
	TablePrivilegeGrantSelect TablePrivilege = "GRANT_SELECT"
)

type ColumnAccessMode string

const (
	ColumnAccessModeNone ColumnAccessMode = "NONE"
	ColumnAccessModeAll  ColumnAccessMode = "ALL"
)

type AccessControlModel struct {
	ClusterId                types.String   `tfsdk:"cluster_id"`
	Catalogs                 types.List     `tfsdk:"catalogs"`
	Schemas                  types.List     `tfsdk:"schemas"`
	Tables                   types.List     `tfsdk:"tables"`
	Functions                types.List     `tfsdk:"functions"`
	Procedures               types.List     `tfsdk:"procedures"`
	Queries                  types.List     `tfsdk:"queries"`
	SystemSessionProperties  types.List     `tfsdk:"system_session_properties"`
	CatalogSessionProperties types.List     `tfsdk:"catalog_session_properties"`
	Timeouts                 timeouts.Value `tfsdk:"timeouts"`
}

func (a AccessControlModel) Validate() diag.Diagnostics {
	var diags diag.Diagnostics
	if a.hasUnknownFields() {
		// Validate later.
		return diags
	}
	if a.hasNoRules() {
		diags.Append(diag.NewErrorDiagnostic("Access control is invalid.", "At least one rule should be specified."))
	}
	return diags
}

func (a AccessControlModel) hasUnknownFields() bool {
	return a.Catalogs.IsUnknown() ||
		a.Schemas.IsUnknown() ||
		a.Tables.IsUnknown() ||
		a.Functions.IsUnknown() ||
		a.Procedures.IsUnknown() ||
		a.Queries.IsUnknown() ||
		a.SystemSessionProperties.IsUnknown() ||
		a.CatalogSessionProperties.IsUnknown()
}

func (a AccessControlModel) hasNoRules() bool {
	return len(a.Catalogs.Elements()) == 0 &&
		len(a.Schemas.Elements()) == 0 &&
		len(a.Tables.Elements()) == 0 &&
		len(a.Functions.Elements()) == 0 &&
		len(a.Procedures.Elements()) == 0 &&
		len(a.Queries.Elements()) == 0 &&
		len(a.SystemSessionProperties.Elements()) == 0 &&
		len(a.CatalogSessionProperties.Elements()) == 0
}

type CatalogRule struct {
	Catalog     types.Object `tfsdk:"catalog"`
	Users       types.List   `tfsdk:"users"`
	Groups      types.List   `tfsdk:"groups"`
	Permission  types.String `tfsdk:"permission"`
	Description types.String `tfsdk:"description"`
}

type SchemaRule struct {
	Catalog     types.Object `tfsdk:"catalog"`
	Schema      types.Object `tfsdk:"schema"`
	Users       types.List   `tfsdk:"users"`
	Groups      types.List   `tfsdk:"groups"`
	Owner       types.String `tfsdk:"owner"`
	Description types.String `tfsdk:"description"`
}

type TableRule struct {
	Catalog     types.Object `tfsdk:"catalog"`
	Schema      types.Object `tfsdk:"schema"`
	Table       types.Object `tfsdk:"table"`
	Users       types.List   `tfsdk:"users"`
	Groups      types.List   `tfsdk:"groups"`
	Privileges  types.List   `tfsdk:"privileges"`
	Columns     types.List   `tfsdk:"columns"`
	Filter      types.String `tfsdk:"filter"`
	Description types.String `tfsdk:"description"`
}

type FunctionRule struct {
	Catalog     types.Object `tfsdk:"catalog"`
	Schema      types.Object `tfsdk:"schema"`
	Function    types.Object `tfsdk:"function"`
	Users       types.List   `tfsdk:"users"`
	Groups      types.List   `tfsdk:"groups"`
	Privileges  types.List   `tfsdk:"privileges"`
	Description types.String `tfsdk:"description"`
}

type ProcedureRule struct {
	Catalog     types.Object `tfsdk:"catalog"`
	Schema      types.Object `tfsdk:"schema"`
	Procedure   types.Object `tfsdk:"procedure"`
	Users       types.List   `tfsdk:"users"`
	Groups      types.List   `tfsdk:"groups"`
	Privileges  types.List   `tfsdk:"privileges"`
	Description types.String `tfsdk:"description"`
}

type SystemSessionPropertyRule struct {
	Property    types.Object `tfsdk:"property"`
	Users       types.List   `tfsdk:"users"`
	Groups      types.List   `tfsdk:"groups"`
	Allow       types.String `tfsdk:"allow"`
	Description types.String `tfsdk:"description"`
}

type CatalogSessionPropertyRule struct {
	Catalog     types.Object `tfsdk:"catalog"`
	Property    types.Object `tfsdk:"property"`
	Users       types.List   `tfsdk:"users"`
	Groups      types.List   `tfsdk:"groups"`
	Allow       types.String `tfsdk:"allow"`
	Description types.String `tfsdk:"description"`
}

type QueryRule struct {
	Users       types.List   `tfsdk:"users"`
	Groups      types.List   `tfsdk:"groups"`
	QueryOwners types.List   `tfsdk:"query_owners"`
	Privileges  types.List   `tfsdk:"privileges"`
	Description types.String `tfsdk:"description"`
}

type ColumnRule struct {
	Name   types.String `tfsdk:"name"`
	Access types.String `tfsdk:"access"`
	Mask   types.String `tfsdk:"mask"`
}

type CatalogMatcherModel struct {
	NameRegexp types.String `tfsdk:"name_regexp"`
	IDs        types.List   `tfsdk:"ids"`
}

type NameMatcherModel struct {
	NameRegexp types.String `tfsdk:"name_regexp"`
	Names      types.List   `tfsdk:"names"`
}

// Base options for conversion
var baseOptions = basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: false}

// ObjectType definitions
var CatalogMatcherT = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"name_regexp": types.StringType,
		"ids":         types.ListType{ElemType: types.StringType},
	},
}

var NameMatcherT = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"name_regexp": types.StringType,
		"names":       types.ListType{ElemType: types.StringType},
	},
}

var ColumnRuleT = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"name":   types.StringType,
		"access": types.StringType,
		"mask":   types.StringType,
	},
}

var CatalogRuleT = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"catalog":     CatalogMatcherT,
		"users":       types.ListType{ElemType: types.StringType},
		"groups":      types.ListType{ElemType: types.StringType},
		"permission":  types.StringType,
		"description": types.StringType,
	},
}

var SchemaRuleT = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"catalog":     CatalogMatcherT,
		"schema":      NameMatcherT,
		"users":       types.ListType{ElemType: types.StringType},
		"groups":      types.ListType{ElemType: types.StringType},
		"owner":       types.StringType,
		"description": types.StringType,
	},
}

var TableRuleT = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"catalog":     CatalogMatcherT,
		"schema":      NameMatcherT,
		"table":       NameMatcherT,
		"users":       types.ListType{ElemType: types.StringType},
		"groups":      types.ListType{ElemType: types.StringType},
		"privileges":  types.ListType{ElemType: types.StringType},
		"columns":     types.ListType{ElemType: ColumnRuleT},
		"filter":      types.StringType,
		"description": types.StringType,
	},
}

var FunctionRuleT = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"catalog":     CatalogMatcherT,
		"schema":      NameMatcherT,
		"function":    NameMatcherT,
		"users":       types.ListType{ElemType: types.StringType},
		"groups":      types.ListType{ElemType: types.StringType},
		"privileges":  types.ListType{ElemType: types.StringType},
		"description": types.StringType,
	},
}

var ProcedureRuleT = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"catalog":     CatalogMatcherT,
		"schema":      NameMatcherT,
		"procedure":   NameMatcherT,
		"users":       types.ListType{ElemType: types.StringType},
		"groups":      types.ListType{ElemType: types.StringType},
		"privileges":  types.ListType{ElemType: types.StringType},
		"description": types.StringType,
	},
}

var QueryRuleT = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"users":        types.ListType{ElemType: types.StringType},
		"groups":       types.ListType{ElemType: types.StringType},
		"query_owners": types.ListType{ElemType: types.StringType},
		"privileges":   types.ListType{ElemType: types.StringType},
		"description":  types.StringType,
	},
}

var SystemSessionPropertyRuleT = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"property":    NameMatcherT,
		"users":       types.ListType{ElemType: types.StringType},
		"groups":      types.ListType{ElemType: types.StringType},
		"allow":       types.StringType,
		"description": types.StringType,
	},
}

var CatalogSessionPropertyRuleT = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"catalog":     CatalogMatcherT,
		"property":    NameMatcherT,
		"users":       types.ListType{ElemType: types.StringType},
		"groups":      types.ListType{ElemType: types.StringType},
		"allow":       types.StringType,
		"description": types.StringType,
	},
}
