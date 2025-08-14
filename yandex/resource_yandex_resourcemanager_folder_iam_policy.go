package yandex

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const yandexResourceManagerFolderDefaultTimeout = 30 * time.Minute

func resourceYandexResourceManagerFolderIAMPolicy() *schema.Resource {
	return resourceIamPolicy(
		IamFolderSchema,
		newFolderIamUpdater,
		WithTimeout(
			&schema.ResourceTimeout{
				Default: schema.DefaultTimeout(yandexResourceManagerFolderDefaultTimeout),
			}),
		WithImporter(
			&schema.ResourceImporter{
				StateContext: iamPolicyImport(folderIDParseFunc),
			}),
		WithDescription("Allows creation and management of the IAM policy for an existing Yandex Resource Manager folder."),
	)
}
