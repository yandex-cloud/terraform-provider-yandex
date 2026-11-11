package models

import "github.com/hashicorp/terraform-plugin-framework/diag"

// Updates model fields only if there is semantic difference.
// See [EqualSemantically] function for details.
func (a *AccessControlModel) ApplyChanges(other AccessControlModel) diag.Diagnostics {
	var diags diag.Diagnostics
	if a == nil {
		diags.AddError("Provider internal error", "Access control should be non-nil when applying changes")
		return diags
	}

	if !listsEqual(a.Catalogs, other.Catalogs) {
		a.Catalogs = other.Catalogs
	}

	if !listsEqual(a.Schemas, other.Schemas) {
		a.Schemas = other.Schemas
	}

	if !listsEqual(a.Tables, other.Tables) {
		a.Tables = other.Tables
	}

	if !listsEqual(a.Functions, other.Functions) {
		a.Functions = other.Functions
	}

	if !listsEqual(a.Procedures, other.Procedures) {
		a.Procedures = other.Procedures
	}

	if !listsEqual(a.Queries, other.Queries) {
		a.Queries = other.Queries
	}

	if !listsEqual(a.SystemSessionProperties, other.SystemSessionProperties) {
		a.SystemSessionProperties = other.SystemSessionProperties
	}

	if !listsEqual(a.CatalogSessionProperties, other.CatalogSessionProperties) {
		a.CatalogSessionProperties = other.CatalogSessionProperties
	}

	return diags
}
