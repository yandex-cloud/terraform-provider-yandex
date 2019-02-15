package yandex

import "github.com/hashicorp/terraform/helper/schema"

func resourceYandexResourceManagerFolderIAMBinding() *schema.Resource {
	return resourceIamBindingWithImport(IamFolderSchema, newFolderIamUpdater, folderIDParseFunc)
}
