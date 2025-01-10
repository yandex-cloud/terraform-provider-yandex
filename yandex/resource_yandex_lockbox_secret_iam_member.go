package yandex

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceYandexLockboxSecretIAMMember() *schema.Resource {
	return resourceIamMember(
		IamLockboxSecretSchema,
		newLockboxSecretIamUpdater,
		WithTimeout(
			&schema.ResourceTimeout{
				Default: schema.DefaultTimeout(yandexIAMLockboxDefaultTimeout),
			}),
		WithImporter(
			&schema.ResourceImporter{
				StateContext: iamMemberImport(LockboxSecretIDParseFunc),
			}),
	)
}
