package pluginsdk

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
)

func ResourceYandexResourceManagerFolderIAMPolicy() *schema.Resource {
	return sdkcommon.ResourceIamPolicy(
		IamFolderSchema,
		newFolderIamUpdater, sdkcommon.WithTimeout(
			&schema.ResourceTimeout{
				Default: schema.DefaultTimeout(YandexResourceManagerFolderDefaultTimeout),
			}), sdkcommon.WithImporter(
			&schema.ResourceImporter{
				StateContext: sdkcommon.IamPolicyImport(folderIDParseFunc),
			}), sdkcommon.WithDescription("Allows creation and management of the IAM policy for an existing Yandex Resource Manager folder."),
	)
}
