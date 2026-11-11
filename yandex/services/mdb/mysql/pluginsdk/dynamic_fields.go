package pluginsdk

import (
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
)

func defaultStringCompare(fieldsInfo *sdkcommon.ObjectFieldsInfo, old, new, fieldname string) bool {
	return old == new
}
