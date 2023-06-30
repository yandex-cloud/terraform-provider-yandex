package yandex

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceYandexIAMServiceAccountIAMBinding() *schema.Resource {
	return resourceIamBinding(
		IamServiceAccountSchema,
		newServiceAccountIamUpdater,
		WithTimeout(
			&schema.ResourceTimeout{
				Default: schema.DefaultTimeout(yandexIAMServiceAccountDefaultTimeout),
			}),
		WithImporter(
			&schema.ResourceImporter{
				StateContext: iamBindingImport(serviceAccountIDParseFunc),
			}),
	)
}
