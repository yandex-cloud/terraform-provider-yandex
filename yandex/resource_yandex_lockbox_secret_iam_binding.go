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
		WithDescription("Allows creation and management of a single binding within IAM policy for an existing Yandex Lockbox Secret.\n\n~> Roles controlled by `yandex_lockbox_secret_iam_binding` should not be assigned using `yandex_lockbox_secret_iam_member`.\n\n~> When you delete `yandex_lockbox_secret_iam_binding` resource, the roles can be deleted from other users within the folder as well. Be careful!\n"),
	)
}
