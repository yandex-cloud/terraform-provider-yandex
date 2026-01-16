package utils

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func ExpandEnum(keyName string, value string, enumValues map[string]int32, diags *diag.Diagnostics) *int32 {
	var unspecifiedValue int32
	if value == "" {
		return &unspecifiedValue
	}

	if val, ok := enumValues[value]; ok {
		return &val
	} else {
		diags.AddError(
			"Failed in conversion enum",
			fmt.Sprintf("value for '%s' must be one of %s, not `%s`",
				keyName, getJoinedKeys(getEnumValueMapKeys(enumValues)), value),
		)
		return nil
	}
}

func getJoinedKeys(keys []string) string {
	return "`" + strings.Join(keys, "`, `") + "`"
}

func getEnumValueMapKeys(m map[string]int32) []string {
	return getEnumValueMapKeysExt(m, false)
}

func getEnumValueMapKeysExt(m map[string]int32, skipDefault bool) []string {
	keys := make([]string, 0, len(m))
	for k, v := range m {
		if v == 0 && skipDefault {
			continue
		}

		keys = append(keys, k)
	}
	return keys
}
