package yandex

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceYandexContainerRegistryIAMBinding() *schema.Resource {
	return resourceIamBinding(
		IamContainerRegistrySchema,
		newContainerRegistryIamUpdater,
		WithTimeout(
			&schema.ResourceTimeout{
				Default: schema.DefaultTimeout(yandexIAMContainerRegistryDefaultTimeout),
			}),
		WithImporter(
			&schema.ResourceImporter{
				StateContext: iamBindingImport(containerRegistryIDParseFunc),
			}),
	)
}
