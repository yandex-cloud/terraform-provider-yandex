package yandex

import "github.com/hashicorp/terraform-plugin-sdk/helper/schema"

func resourceYandexContainerRegistryIAMBinding() *schema.Resource {
	return resourceIamBindingWithImport(IamContainerRegistrySchema, newContainerRegistryIamUpdater, containerRegistryIDParseFunc)
}
