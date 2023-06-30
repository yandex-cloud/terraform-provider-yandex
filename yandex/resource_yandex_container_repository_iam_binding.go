package yandex

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceYandexContainerRepositoryIAMBinding() *schema.Resource {
	return resourceIamBinding(
		IamContainerRepositorySchema,
		newContainerRepositoryIamUpdater,
		WithTimeout(
			&schema.ResourceTimeout{
				Default: schema.DefaultTimeout(yandexIAMContainerRepositoryDefaultTimeout),
			}),
		WithImporter(
			&schema.ResourceImporter{
				StateContext: iamBindingImport(containerRepositoryIDParseFunc),
			}),
	)
}
