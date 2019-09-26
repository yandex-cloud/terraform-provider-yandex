package yandex

import "github.com/hashicorp/terraform-plugin-sdk/helper/schema"

func resourceYandexResourceManagerCloudIAMBinding() *schema.Resource {
	return resourceIamBindingWithImport(IamCloudSchema, newCloudIamUpdater, cloudIDParseFunc)
}
