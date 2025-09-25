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

	if len(a.Catalogs) == len(other.Catalogs) {
		for i := range a.Catalogs {
			if !catalogsEqual(a.Catalogs[i], other.Catalogs[i]) {
				a.Catalogs[i] = other.Catalogs[i]
			}
		}
	} else {
		a.Catalogs = other.Catalogs
	}

	if len(a.Schemas) == len(other.Schemas) {
		for i := range a.Schemas {
			if !schemasEqual(a.Schemas[i], other.Schemas[i]) {
				a.Schemas[i] = other.Schemas[i]
			}
		}
	} else {
		a.Schemas = other.Schemas
	}

	if len(a.Tables) == len(other.Tables) {
		for i := range a.Tables {
			if !tablesEqual(a.Tables[i], other.Tables[i]) {
				a.Tables[i] = other.Tables[i]
			}
		}
	} else {
		a.Tables = other.Tables
	}

	if len(a.Functions) == len(other.Functions) {
		for i := range a.Functions {
			if !functionsEqual(a.Functions[i], other.Functions[i]) {
				a.Functions[i] = other.Functions[i]
			}
		}
	} else {
		a.Functions = other.Functions
	}

	if len(a.Procedures) == len(other.Procedures) {
		for i := range a.Procedures {
			if !proceduresEqual(a.Procedures[i], other.Procedures[i]) {
				a.Procedures[i] = other.Procedures[i]
			}
		}
	} else {
		a.Procedures = other.Procedures
	}

	if len(a.Queries) == len(other.Queries) {
		for i := range a.Queries {
			if !queriesEqual(a.Queries[i], other.Queries[i]) {
				a.Queries[i] = other.Queries[i]
			}
		}
	} else {
		a.Queries = other.Queries
	}

	if len(a.SystemSessionProperties) == len(other.SystemSessionProperties) {
		for i := range a.SystemSessionProperties {
			if !systemSessionPropertiesEqual(a.SystemSessionProperties[i], other.SystemSessionProperties[i]) {
				a.SystemSessionProperties[i] = other.SystemSessionProperties[i]
			}
		}
	} else {
		a.SystemSessionProperties = other.SystemSessionProperties
	}

	if len(a.CatalogSessionProperties) == len(other.CatalogSessionProperties) {
		for i := range a.CatalogSessionProperties {
			if !catalogSessionPropertiesEqual(a.CatalogSessionProperties[i], other.CatalogSessionProperties[i]) {
				a.CatalogSessionProperties[i] = other.CatalogSessionProperties[i]
			}
		}
	} else {
		a.CatalogSessionProperties = other.CatalogSessionProperties
	}

	return diags
}
