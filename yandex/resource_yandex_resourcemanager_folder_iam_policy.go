package yandex

import "github.com/hashicorp/terraform/helper/schema"

func resourceYandexResourceManagerFolderIAMPolicy() *schema.Resource {
	return ResourceIamPolicyWithImport(IamFolderSchema, NewFolderIamUpdater, FolderIDParseFunc)
}
