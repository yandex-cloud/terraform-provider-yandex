package sdkcommon

import (
	"fmt"

	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/hashcode"
)

func MysqlDatabaseHash(v interface{}) int {
	m := v.(map[string]interface{})

	if n, ok := m["name"]; ok {
		return hashcode.String(n.(string))
	}
	return 0
}

func StringToTimeValidateFunc(value interface{}, key string) (fields []string, errors []error) {
	if strTime, ok := value.(string); ok {
		_, err := mdbcommon.ParseStringToTime(strTime)
		if err != nil {
			errors = append(errors, err)
		}
	} else {
		errors = append(errors, fmt.Errorf("value %v is not string", value))
	}

	return fields, errors
}
