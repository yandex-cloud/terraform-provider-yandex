package yandex

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceYandexContainerRegistryIAMBinding() *schema.Resource {
	return resourceIamBindingWithImport(IamContainerRegistrySchema, newContainerRegistryIamUpdater, containerRegistryIDParseFunc)
}
