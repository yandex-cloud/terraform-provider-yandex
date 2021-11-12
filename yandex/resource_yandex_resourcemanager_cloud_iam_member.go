package yandex

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceYandexResourceManagerCloudIAMMember() *schema.Resource {
	return resourceIamMemberWithImport(IamCloudSchema, newCloudIamUpdater, cloudIDParseFunc)
}
