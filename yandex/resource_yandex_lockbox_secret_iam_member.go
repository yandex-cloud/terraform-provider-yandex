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
		WithDescription("Allows creation and management of a single member for a single binding within the IAM policy for an existing Yandex Lockbox Secret.\n\n~> Roles controlled by `yandex_lockbox_secret_iam_binding` should not be assigned using `yandex_lockbox_secret_iam_member`.\n"),
	)
}
