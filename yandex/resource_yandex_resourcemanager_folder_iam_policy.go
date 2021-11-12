package yandex

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceYandexResourceManagerFolderIAMPolicy() *schema.Resource {
	return resourceIamPolicyWithImport(IamFolderSchema, newFolderIamUpdater, folderIDParseFunc)
}
