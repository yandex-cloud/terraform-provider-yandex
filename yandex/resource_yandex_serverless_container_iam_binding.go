package yandex

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceYandexServerlessContainerIAMBinding() *schema.Resource {
	return resourceIamBindingWithImport(IamServerlessContainerSchema, newServerlessContainerIamUpdater, serverlessContainerIDParseFunc)
}
