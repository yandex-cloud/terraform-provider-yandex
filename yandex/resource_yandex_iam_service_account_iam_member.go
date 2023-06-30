package yandex

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceYandexIAMServiceAccountIAMMember() *schema.Resource {
	return resourceIamMember(
		IamServiceAccountSchema,
		newServiceAccountIamUpdater,
		WithTimeout(
			&schema.ResourceTimeout{
				Default: schema.DefaultTimeout(yandexIAMServiceAccountDefaultTimeout),
			}),
		WithImporter(
			&schema.ResourceImporter{
				StateContext: iamMemberImport(serviceAccountIDParseFunc),
			}),
	)
}
