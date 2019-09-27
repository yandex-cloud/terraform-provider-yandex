package yandex

import "github.com/hashicorp/terraform-plugin-sdk/helper/schema"

func resourceYandexResourceManagerFolderIAMMember() *schema.Resource {
	return resourceIamMemberWithImport(IamFolderSchema, newFolderIamUpdater, folderIDParseFunc)
}
