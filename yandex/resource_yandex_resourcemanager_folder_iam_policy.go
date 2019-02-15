package yandex

import "github.com/hashicorp/terraform/helper/schema"

func resourceYandexResourceManagerFolderIAMPolicy() *schema.Resource {
	return resourceIamPolicyWithImport(IamFolderSchema, newFolderIamUpdater, folderIDParseFunc)
}
