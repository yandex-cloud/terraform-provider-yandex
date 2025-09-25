package models

import (
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Returns whether two instances of [AccessControlModel] are semantically equal.
// Semantically, empty string equals null string field and empty container (list, set, map, etc.)
// equals null container.
func EqualSemantically(a1 AccessControlModel, a2 AccessControlModel) bool {
	return a1.ClusterId.Equal(a2.ClusterId) &&
		slices.EqualFunc(a1.Catalogs, a2.Catalogs, catalogsEqual) &&
		slices.EqualFunc(a1.Schemas, a2.Schemas, schemasEqual) &&
		slices.EqualFunc(a1.Tables, a2.Tables, tablesEqual) &&
		slices.EqualFunc(a1.Functions, a2.Functions, functionsEqual) &&
		slices.EqualFunc(a1.Procedures, a2.Procedures, proceduresEqual) &&
		slices.EqualFunc(a1.Queries, a2.Queries, queriesEqual) &&
		slices.EqualFunc(a1.SystemSessionProperties, a2.SystemSessionProperties, systemSessionPropertiesEqual) &&
		slices.EqualFunc(a1.CatalogSessionProperties, a2.CatalogSessionProperties, catalogSessionPropertiesEqual)
}

func catalogsEqual(c1, c2 *CatalogRule) bool {
	if c1 == nil || c2 == nil {
		return c1 == c2
	}
	return stringsEqual(c1.Description, c2.Description) &&
		stringsEqual(c1.Permission, c2.Permission) &&
		catalogMatchersEqual(c1.Catalog, c2.Catalog) &&
		listsEqual(c1.Users, c2.Users) &&
		listsEqual(c1.Groups, c2.Groups)
}

func catalogMatchersEqual(m1, m2 *CatalogMatcherModel) bool {
	if m1 == nil || m2 == nil {
		return m1 == m2
	}
	return stringsEqual(m1.NameRegexp, m2.NameRegexp) && listsEqual(m1.IDs, m2.IDs)
}

func schemasEqual(s1, s2 *SchemaRule) bool {
	if s1 == nil || s2 == nil {
		return s1 == s2
	}
	return stringsEqual(s1.Description, s2.Description) &&
		stringsEqual(s1.Owner, s2.Owner) &&
		catalogMatchersEqual(s1.Catalog, s2.Catalog) &&
		nameMatchersEqual(s1.Schema, s2.Schema) &&
		listsEqual(s1.Users, s2.Users) &&
		listsEqual(s1.Groups, s2.Groups)
}

func functionsEqual(f1, f2 *FunctionRule) bool {
	if f1 == nil || f2 == nil {
		return f1 == f2
	}
	return stringsEqual(f1.Description, f2.Description) &&
		catalogMatchersEqual(f1.Catalog, f2.Catalog) &&
		nameMatchersEqual(f1.Schema, f2.Schema) &&
		nameMatchersEqual(f1.Function, f2.Function) &&
		listsEqual(f1.Users, f2.Users) &&
		listsEqual(f1.Groups, f2.Groups) &&
		listsEqual(f1.Privileges, f2.Privileges)
}

func proceduresEqual(p1, p2 *ProcedureRule) bool {
	if p1 == nil || p2 == nil {
		return p1 == p2
	}
	return stringsEqual(p1.Description, p2.Description) &&
		catalogMatchersEqual(p1.Catalog, p2.Catalog) &&
		nameMatchersEqual(p1.Schema, p2.Schema) &&
		nameMatchersEqual(p1.Procedure, p2.Procedure) &&
		listsEqual(p1.Users, p2.Users) &&
		listsEqual(p1.Groups, p2.Groups) &&
		listsEqual(p1.Privileges, p2.Privileges)
}

func tablesEqual(t1, t2 *TableRule) bool {
	if t1 == nil || t2 == nil {
		return t1 == t2
	}
	return stringsEqual(t1.Description, t2.Description) &&
		stringsEqual(t1.Filter, t2.Filter) &&
		catalogMatchersEqual(t1.Catalog, t2.Catalog) &&
		nameMatchersEqual(t1.Schema, t2.Schema) &&
		nameMatchersEqual(t1.Table, t2.Table) &&
		listsEqual(t1.Users, t2.Users) &&
		listsEqual(t1.Groups, t2.Groups) &&
		listsEqual(t1.Privileges, t2.Privileges) &&
		columnRulesEqual(t1.Columns, t2.Columns)
}

func columnRulesEqual(c1, c2 []*ColumnRule) bool {
	if len(c1) != len(c2) {
		return false
	}
	for i := range c1 {
		if !columnRuleEqual(c1[i], c2[i]) {
			return false
		}
	}
	return true
}

func columnRuleEqual(c1, c2 *ColumnRule) bool {
	if c1 == nil || c2 == nil {
		return c1 == c2
	}
	return stringsEqual(c1.Name, c2.Name) &&
		stringsEqual(c1.Access, c2.Access) &&
		stringsEqual(c1.Mask, c2.Mask)
}

func systemSessionPropertiesEqual(s1, s2 *SystemSessionPropertyRule) bool {
	if s1 == nil || s2 == nil {
		return s1 == s2
	}
	return stringsEqual(s1.Description, s2.Description) &&
		stringsEqual(s1.Allow, s2.Allow) &&
		nameMatchersEqual(s1.Property, s2.Property) &&
		listsEqual(s1.Users, s2.Users) &&
		listsEqual(s1.Groups, s2.Groups)
}

func catalogSessionPropertiesEqual(c1, c2 *CatalogSessionPropertyRule) bool {
	if c1 == nil || c2 == nil {
		return c1 == c2
	}
	return stringsEqual(c1.Description, c2.Description) &&
		stringsEqual(c1.Allow, c2.Allow) &&
		catalogMatchersEqual(c1.Catalog, c2.Catalog) &&
		nameMatchersEqual(c1.Property, c2.Property) &&
		listsEqual(c1.Users, c2.Users) &&
		listsEqual(c1.Groups, c2.Groups)
}

func queriesEqual(q1, q2 *QueryRule) bool {
	if q1 == nil || q2 == nil {
		return q1 == q2
	}
	return stringsEqual(q1.Description, q2.Description) &&
		listsEqual(q1.Users, q2.Users) &&
		listsEqual(q1.Groups, q2.Groups) &&
		listsEqual(q1.QueryOwners, q2.QueryOwners) &&
		listsEqual(q1.Privileges, q2.Privileges)
}

func nameMatchersEqual(m1, m2 *NameMatcherModel) bool {
	if m1 == nil || m2 == nil {
		return m1 == m2
	}
	return stringsEqual(m1.NameRegexp, m2.NameRegexp) && listsEqual(m1.Names, m2.Names)
}

func stringsEqual(s1, s2 types.String) bool {
	if s1.Equal(s2) {
		return true
	}
	return s1.ValueString() == "" && s2.ValueString() == ""
}

func listsEqual(l1, l2 types.List) bool {
	if l1.Equal(l2) {
		return true
	}
	return len(l1.Elements()) == 0 && len(l2.Elements()) == 0
}
