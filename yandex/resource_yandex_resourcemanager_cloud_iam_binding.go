package yandex

import "github.com/hashicorp/terraform/helper/schema"

func resourceYandexResourceManagerCloudIAMBinding() *schema.Resource {
	return resourceIamBindingWithImport(IamCloudSchema, newCloudIamUpdater, cloudIDParseFunc)
}
