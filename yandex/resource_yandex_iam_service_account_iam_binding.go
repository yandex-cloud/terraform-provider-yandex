package yandex

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceYandexIAMServiceAccountIAMBinding() *schema.Resource {
	return resourceIamBindingWithImport(IamServiceAccountSchema, newServiceAccountIamUpdater, serviceAccountIDParseFunc)
}
