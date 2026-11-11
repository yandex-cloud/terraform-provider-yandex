package pluginsdk

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
)

var IamFolderSchema = map[string]*schema.Schema{
	"folder_id": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: "The ID of the folder to attach a policy to.",
	},
}

func newFolderIamUpdater(d *schema.ResourceData, config *providerconfig.Config) (sdkcommon.ResourceIamUpdater, error) {
	return NewFolderIamUpdaterFromFolderID(d.Get("folder_id").(string), config)
}

func folderIDParseFunc(d *schema.ResourceData, _ *providerconfig.Config) error {
	d.Set("folder_id", d.Id())
	return nil
}
