package models

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	ClusterId                types.String                  `tfsdk:"cluster_id"`
	Catalogs                 []*CatalogRule                `tfsdk:"catalogs"`
	Schemas                  []*SchemaRule                 `tfsdk:"schemas"`
	Tables                   []*TableRule                  `tfsdk:"tables"`
	Functions                []*FunctionRule               `tfsdk:"functions"`
	Procedures               []*ProcedureRule              `tfsdk:"procedures"`
	Queries                  []*QueryRule                  `tfsdk:"queries"`
	SystemSessionProperties  []*SystemSessionPropertyRule  `tfsdk:"system_session_properties"`
	CatalogSessionProperties []*CatalogSessionPropertyRule `tfsdk:"catalog_session_properties"`
	Timeouts                 timeouts.Value                `tfsdk:"timeouts"`
}

func (a AccessControlModel) Validate() diag.Diagnostics {
	var diags diag.Diagnostics
	if a.hasNoRules() {
		diags.Append(diag.NewErrorDiagnostic("Access control is invalid.", "At least one rule should be specified."))
	}
	return diags
}

func (a AccessControlModel) hasNoRules() bool {
	return len(a.Catalogs) == 0 &&
		len(a.Schemas) == 0 &&
		len(a.Tables) == 0 &&
		len(a.Functions) == 0 &&
		len(a.Procedures) == 0 &&
		len(a.Queries) == 0 &&
		len(a.SystemSessionProperties) == 0 &&
		len(a.CatalogSessionProperties) == 0
}

type CatalogRule struct {
	Catalog     *CatalogMatcherModel `tfsdk:"catalog"`
	Users       types.List           `tfsdk:"users"`
	Groups      types.List           `tfsdk:"groups"`
	Permission  types.String         `tfsdk:"permission"`
	Description types.String         `tfsdk:"description"`
}

type SchemaRule struct {
	Catalog     *CatalogMatcherModel `tfsdk:"catalog"`
	Schema      *NameMatcherModel    `tfsdk:"schema"`
	Users       types.List           `tfsdk:"users"`
	Groups      types.List           `tfsdk:"groups"`
	Owner       types.String         `tfsdk:"owner"`
	Description types.String         `tfsdk:"description"`
}

type TableRule struct {
	Catalog     *CatalogMatcherModel `tfsdk:"catalog"`
	Schema      *NameMatcherModel    `tfsdk:"schema"`
	Table       *NameMatcherModel    `tfsdk:"table"`
	Users       types.List           `tfsdk:"users"`
	Groups      types.List           `tfsdk:"groups"`
	Privileges  types.List           `tfsdk:"privileges"`
	Columns     []*ColumnRule        `tfsdk:"columns"`
	Filter      types.String         `tfsdk:"filter"`
	Description types.String         `tfsdk:"description"`
}

type FunctionRule struct {
	Catalog     *CatalogMatcherModel `tfsdk:"catalog"`
	Schema      *NameMatcherModel    `tfsdk:"schema"`
	Function    *NameMatcherModel    `tfsdk:"function"`
	Users       types.List           `tfsdk:"users"`
	Groups      types.List           `tfsdk:"groups"`
	Privileges  types.List           `tfsdk:"privileges"`
	Description types.String         `tfsdk:"description"`
}

type ProcedureRule struct {
	Catalog     *CatalogMatcherModel `tfsdk:"catalog"`
	Schema      *NameMatcherModel    `tfsdk:"schema"`
	Procedure   *NameMatcherModel    `tfsdk:"procedure"`
	Users       types.List           `tfsdk:"users"`
	Groups      types.List           `tfsdk:"groups"`
	Privileges  types.List           `tfsdk:"privileges"`
	Description types.String         `tfsdk:"description"`
}

type SystemSessionPropertyRule struct {
	Property    *NameMatcherModel `tfsdk:"property"`
	Users       types.List        `tfsdk:"users"`
	Groups      types.List        `tfsdk:"groups"`
	Allow       types.String      `tfsdk:"allow"`
	Description types.String      `tfsdk:"description"`
}

type CatalogSessionPropertyRule struct {
	Catalog     *CatalogMatcherModel `tfsdk:"catalog"`
	Property    *NameMatcherModel    `tfsdk:"property"`
	Users       types.List           `tfsdk:"users"`
	Groups      types.List           `tfsdk:"groups"`
	Allow       types.String         `tfsdk:"allow"`
	Description types.String         `tfsdk:"description"`
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
