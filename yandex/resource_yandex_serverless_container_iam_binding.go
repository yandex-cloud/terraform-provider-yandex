package yandex

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceYandexServerlessContainerIAMBinding() *schema.Resource {
	return resourceIamBinding(
		IamServerlessContainerSchema,
		newServerlessContainerIamUpdater,
		WithTimeout(
			&schema.ResourceTimeout{
				Default: schema.DefaultTimeout(yandexIAMServerlessContainerDefaultTimeout),
			}),
		WithImporter(
			&schema.ResourceImporter{
				StateContext: iamBindingImport(serverlessContainerIDParseFunc),
			}),
	)
}
