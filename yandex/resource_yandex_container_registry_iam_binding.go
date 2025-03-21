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
		WithDescription("Allows creation and management of a single binding within IAM policy for an existing Yandex Container Registry."),
	)
}
