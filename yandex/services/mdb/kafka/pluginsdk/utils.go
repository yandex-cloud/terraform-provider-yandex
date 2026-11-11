package pluginsdk

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func convertStringArrToSchemaSet(sslice []string) *schema.Set {
	result := schema.NewSet(schema.HashString, nil)
	for _, sliceElement := range sslice {
		result.Add(sliceElement)
	}
	return result
}
