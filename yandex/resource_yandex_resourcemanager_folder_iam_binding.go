package yandex

import "github.com/hashicorp/terraform-plugin-sdk/helper/schema"

func resourceYandexResourceManagerFolderIAMBinding() *schema.Resource {
	return resourceIamBindingWithImport(IamFolderSchema, newFolderIamUpdater, folderIDParseFunc)
}
