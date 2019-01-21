package yandex

import "github.com/hashicorp/terraform/helper/schema"

func resourceYandexResourceManagerCloudIAMMember() *schema.Resource {
	return ResourceIamMemberWithImport(IamCloudSchema, NewCloudIamUpdater, CloudIDParseFunc)
}
