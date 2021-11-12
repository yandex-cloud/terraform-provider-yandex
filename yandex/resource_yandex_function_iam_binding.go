package yandex

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceYandexFunctionIAMBinding() *schema.Resource {
	return resourceIamBindingWithImport(IamFunctionSchema, newFunctionIamUpdater, functionIDParseFunc)
}
