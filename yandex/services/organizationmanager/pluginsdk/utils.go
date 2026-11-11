package pluginsdk

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
)

func getOrganizationID(d *schema.ResourceData, config *providerconfig.Config) (string, error) {
	res, ok := d.GetOk("organization_id")
	if !ok {
		if config.OrganizationID != "" {
			return config.OrganizationID, nil
		}
		return "", fmt.Errorf("cannot determine organization_id: please set 'organization_id' key in this resource or at provider level")
	}
	return res.(string), nil
}
