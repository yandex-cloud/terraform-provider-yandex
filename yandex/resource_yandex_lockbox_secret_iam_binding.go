package yandex

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceYandexLockboxSecretIAMBinding() *schema.Resource {
	return resourceIamBindingWithImport(IamLockboxSecretSchema, newLockboxSecretIamUpdater, LockboxSecretIDParseFunc)
}
