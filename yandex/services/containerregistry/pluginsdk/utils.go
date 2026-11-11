package pluginsdk

import (
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
)

func CheckEveryOf(d *schema.ResourceData, keys ...string) error {
	for _, key := range keys {
		if key == "" {
			return errors.New("empty keys not allowed")
		}

		_, ok := d.GetOk(key)

		if !ok {
			return fmt.Errorf("provide every of: %v", sdkcommon.GetJoinedKeys(keys))
		}
	}

	return nil
}
