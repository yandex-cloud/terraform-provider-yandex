package sdkcommon

import (
	"bytes"
	"fmt"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/hashcode"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func StorageBucketS3SetFunc(keys ...string) schema.SchemaSetFunc {
	return func(v interface{}) int {
		var buf bytes.Buffer
		m, ok := v.(map[string]interface{})

		if !ok {
			return 0
		}

		for _, key := range keys {
			if v, ok := m[key]; ok {
				value := fmt.Sprintf("%v", v)
				buf.WriteString(value + "-")
			}
		}

		return hashcode.String(buf.String())
	}
}
