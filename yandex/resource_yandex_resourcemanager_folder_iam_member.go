package yandex

import "github.com/hashicorp/terraform/helper/schema"

func resourceYandexResourceManagerFolderIAMMember() *schema.Resource {
	return resourceIamMemberWithImport(IamFolderSchema, newFolderIamUpdater, folderIDParseFunc)
}
