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
		WithDescription("Allows creation and management of a single binding within IAM policy for an existing Yandex Container Repository. For more information, see [the official documentation](https://yandex.cloud/docs/container-registry/concepts/repository)."),
	)
}
