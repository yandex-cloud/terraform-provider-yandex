package yandex

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

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
	)
}
