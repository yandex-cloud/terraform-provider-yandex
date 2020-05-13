package yandex

import "github.com/hashicorp/terraform-plugin-sdk/helper/schema"

func resourceYandexFunctionIAMBinding() *schema.Resource {
	return resourceIamBindingWithImport(IamFunctionSchema, newFunctionIamUpdater, functionIDParseFunc)
}
