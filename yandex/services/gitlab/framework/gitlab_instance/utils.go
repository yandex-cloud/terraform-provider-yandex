package gitlab_instance

import "github.com/hashicorp/terraform-plugin-framework/types"

func instanceIDLogField(cid string) map[string]interface{} {
	return map[string]interface{}{
		"instance_id": cid,
	}
}

func mapsAreEqual(map1, map2 types.Map) bool {
	if map1.Equal(map2) {
		return true
	}
	if len(map1.Elements()) == 0 && len(map2.Elements()) == 0 {
		return true
	}
	return false
}

func stringsAreEqual(str1, str2 types.String) bool {
	if str1.Equal(str2) {
		return true
	}
	if str1.ValueString() == "" && str2.ValueString() == "" {
		return true
	}
	return false
}

func int64AreEqual(int1, int2 types.Int64) bool {
	if int1.Equal(int2) {
		return true
	}
	if int1.ValueInt64() == 0 && int2.ValueInt64() == 0 {
		return true
	}
	return false
}
