package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Returns whether two instances of [AccessControlModel] are semantically equal.
// Semantically, empty string equals null string field and empty container (list, set, map, etc.)
// equals null container.
func EqualSemantically(a1 AccessControlModel, a2 AccessControlModel) bool {
	return a1.ClusterId.Equal(a2.ClusterId) &&
		listsEqual(a1.Catalogs, a2.Catalogs) &&
		listsEqual(a1.Schemas, a2.Schemas) &&
		listsEqual(a1.Tables, a2.Tables) &&
		listsEqual(a1.Functions, a2.Functions) &&
		listsEqual(a1.Procedures, a2.Procedures) &&
		listsEqual(a1.Queries, a2.Queries) &&
		listsEqual(a1.SystemSessionProperties, a2.SystemSessionProperties) &&
		listsEqual(a1.CatalogSessionProperties, a2.CatalogSessionProperties)
}

func listsEqual(l1, l2 types.List) bool {
	if l1.Equal(l2) {
		return true
	}
	// if one of lists is null and the other is empty then we assume that they are equal
	if len(l1.Elements()) == 0 && len(l2.Elements()) == 0 {
		return true
	}
	return false
}
