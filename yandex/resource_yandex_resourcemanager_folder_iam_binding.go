package yandex

import "github.com/hashicorp/terraform/helper/schema"

func resourceYandexResourceManagerFolderIAMBinding() *schema.Resource {
	return ResourceIamBindingWithImport(IamFolderSchema, NewFolderIamUpdater, FolderIDParseFunc)
}
