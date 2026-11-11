package pluginsdk

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
)

func generateMapSchemaDiffSuppressFuncV2(fieldsInfo *sdkcommon.ObjectFieldsInfo) func(k, old, new string, d *schema.ResourceData) bool {
	return func(k, old, new string, d *schema.ResourceData) bool {
		ph := strings.Split(k, ".")
		field := ph[len(ph)-1]

		if fieldsInfo == nil {
			return new == old
		}

		cvf := fieldsInfo.CompareValueFunc(field)
		if cvf != nil {
			return cvf(old, new)
		}

		return new == old
	}
}
