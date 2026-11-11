package pluginsdk

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
)

var IamCloudSchema = map[string]*schema.Schema{
	"cloud_id": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: "The ID of the cloud to attach the policy to.",
	},
}

func newCloudIamUpdater(d *schema.ResourceData, config *providerconfig.Config) (sdkcommon.ResourceIamUpdater, error) {
	return NewCloudIamUpdater(d.Get("cloud_id").(string), config), nil
}

func cloudIDParseFunc(d *schema.ResourceData, _ *providerconfig.Config) error {
	d.Set("cloud_id", d.Id())
	return nil
}
