package yandex

import "github.com/hashicorp/terraform/helper/schema"

func resourceYandexResourceManagerCloudIAMBinding() *schema.Resource {
	return ResourceIamBindingWithImport(IamCloudSchema, NewCloudIamUpdater, CloudIDParseFunc)
}
