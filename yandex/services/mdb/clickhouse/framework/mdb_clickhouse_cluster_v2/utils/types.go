package utils

import "github.com/hashicorp/terraform-plugin-framework/types"

func SetsAreEqual(set1, set2 types.Set) bool {
	if set1.IsUnknown() || set2.IsUnknown() {
		return false
	}

	// if one of sets is null and the other is empty then we assume that they are equal
	if len(set1.Elements()) == 0 && len(set2.Elements()) == 0 {
		return true
	}

	if !set1.IsNull() && set1.Equal(set2) {
		return true
	}

	return false
}
