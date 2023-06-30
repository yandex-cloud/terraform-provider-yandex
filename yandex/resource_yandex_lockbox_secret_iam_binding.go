package yandex

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceYandexLockboxSecretIAMBinding() *schema.Resource {
	return resourceIamBinding(
		IamLockboxSecretSchema,
		newLockboxSecretIamUpdater,
		WithTimeout(
			&schema.ResourceTimeout{
				Default: schema.DefaultTimeout(yandexIAMLockboxDefaultTimeout),
			}),
		WithImporter(
			&schema.ResourceImporter{
				StateContext: iamBindingImport(LockboxSecretIDParseFunc),
			}),
	)
}
