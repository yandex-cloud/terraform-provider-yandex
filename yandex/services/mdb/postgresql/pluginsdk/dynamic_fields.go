package pluginsdk

import (
	"log"
	"slices"

	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
)

func stringOfEnumSliceCompareWithDefault(fieldsInfo *sdkcommon.ObjectFieldsInfo, old, new, fieldname string) bool {
	oldIntValues, err := fieldsInfo.StringToIntSlice(fieldname, old)
	if err != nil {
		log.Printf("[ERROR] Cannot convert enum values to int slice. Fieldname: %s, value: %s", fieldname, new)
		return false
	}
	newIntValues, err := fieldsInfo.StringToIntSlice(fieldname, new)
	if err != nil {
		log.Printf("[ERROR] Cannot convert enum values to int slice. Fieldname: %s, value: %s", fieldname, new)
		return false
	}
	return slices.Equal(oldIntValues, newIntValues)
}
