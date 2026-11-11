package pluginsdk

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
)

func getCloudID(d *schema.ResourceData, config *providerconfig.Config) (string, error) {
	res, ok := d.GetOk("cloud_id")
	if !ok {
		if config.CloudID != "" {
			return config.CloudID, nil
		}
		return "", fmt.Errorf("cannot determine cloud_id: please set 'cloud_id' key in this resource or at provider level")
	}
	return res.(string), nil
}
