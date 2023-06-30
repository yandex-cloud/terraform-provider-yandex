package yandex

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceYandexIAMServiceAccountIAMPolicy() *schema.Resource {
	return resourceIamPolicy(
		IamServiceAccountSchema,
		newServiceAccountIamUpdater,
		WithTimeout(
			&schema.ResourceTimeout{
				Default: schema.DefaultTimeout(yandexIAMServiceAccountDefaultTimeout),
			}),
		WithImporter(
			&schema.ResourceImporter{
				StateContext: iamPolicyImport(serviceAccountIDParseFunc),
			}),
	)
}
