package yandex

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceYandexResourceManagerCloudIAMBinding() *schema.Resource {
	return resourceIamBindingWithImport(IamCloudSchema, newCloudIamUpdater, cloudIDParseFunc)
}
