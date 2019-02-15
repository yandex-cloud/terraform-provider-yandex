package yandex

import "github.com/hashicorp/terraform/helper/schema"

func resourceYandexResourceManagerCloudIAMMember() *schema.Resource {
	return resourceIamMemberWithImport(IamCloudSchema, newCloudIamUpdater, cloudIDParseFunc)
}
