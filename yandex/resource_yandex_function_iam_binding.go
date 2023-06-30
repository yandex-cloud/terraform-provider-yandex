package yandex

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceYandexFunctionIAMBinding() *schema.Resource {
	return resourceIamBinding(
		IamFunctionSchema,
		newFunctionIamUpdater,
		WithTimeout(
			&schema.ResourceTimeout{
				Default: schema.DefaultTimeout(yandexIAMFunctionDefaultTimeout),
			}),
		WithImporter(
			&schema.ResourceImporter{
				StateContext: iamBindingImport(functionIDParseFunc),
			}),
	)
}
