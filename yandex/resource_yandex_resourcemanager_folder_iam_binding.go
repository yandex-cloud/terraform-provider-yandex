package yandex

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceYandexResourceManagerFolderIAMBinding() *schema.Resource {
	return resourceIamBinding(
		IamFolderSchema,
		newFolderIamUpdater,
		WithTimeout(
			&schema.ResourceTimeout{
				Default: schema.DefaultTimeout(yandexResourceManagerFolderDefaultTimeout),
			}),
		WithImporter(
			&schema.ResourceImporter{
				StateContext: iamBindingImport(folderIDParseFunc),
			}),
		WithDescription("Allows creation and management of a single binding within IAM policy for an existing Yandex Resource Manager folder.\n\n~> This resource *must not* be used in conjunction with `yandex_resourcemanager_folder_iam_policy` or they will conflict over what your policy should be.\n\n~> When you delete `yandex_resourcemanager_folder_iam_binding` resource, the roles can be deleted from other users within the folder as well. Be careful!\n"),
	)
}
