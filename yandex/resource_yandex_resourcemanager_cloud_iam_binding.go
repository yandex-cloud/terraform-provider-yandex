package yandex

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceYandexResourceManagerCloudIAMBinding() *schema.Resource {
	return resourceIamBinding(
		IamCloudSchema,
		newCloudIamUpdater,
		WithTimeout(
			&schema.ResourceTimeout{
				Default: schema.DefaultTimeout(yandexResourceManagerCloudDefaultTimeout),
			}),
		WithImporter(
			&schema.ResourceImporter{
				StateContext: iamBindingImport(cloudIDParseFunc),
			}),
	)
}
