package pluginsdk

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func expandStringSet(v interface{}) []string {
	if v == nil {
		return nil
	}
	var result []string
	set := v.(*schema.Set)
	for _, val := range set.List() {
		result = append(result, val.(string))
	}
	return result
}
