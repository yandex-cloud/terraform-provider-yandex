package yandex

import "github.com/hashicorp/terraform/helper/schema"

func resourceYandexResourceManagerFolderIAMMember() *schema.Resource {
	return ResourceIamMemberWithImport(IamFolderSchema, NewFolderIamUpdater, FolderIDParseFunc)
}
