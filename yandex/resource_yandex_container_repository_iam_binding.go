package yandex

import "github.com/hashicorp/terraform-plugin-sdk/helper/schema"

func resourceYandexContainerRepositoryIAMBinding() *schema.Resource {
	return resourceIamBindingWithImport(IamContainerRepositorySchema, newContainerRepositoryIamUpdater, containerRepositoryIDParseFunc)
}
