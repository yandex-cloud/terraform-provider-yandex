package yandex

import "github.com/hashicorp/terraform-plugin-sdk/helper/schema"

func resourceYandexResourceManagerFolderIAMPolicy() *schema.Resource {
	return resourceIamPolicyWithImport(IamFolderSchema, newFolderIamUpdater, folderIDParseFunc)
}
