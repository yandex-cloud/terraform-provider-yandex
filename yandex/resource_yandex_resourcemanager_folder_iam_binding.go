package yandex

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceYandexResourceManagerFolderIAMBinding() *schema.Resource {
	return resourceIamBindingWithImport(IamFolderSchema, newFolderIamUpdater, folderIDParseFunc)
}
