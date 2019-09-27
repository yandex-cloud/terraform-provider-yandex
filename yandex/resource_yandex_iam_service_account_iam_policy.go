package yandex

import "github.com/hashicorp/terraform-plugin-sdk/helper/schema"

func resourceYandexIAMServiceAccountIAMPolicy() *schema.Resource {
	return resourceIamPolicyWithImport(IamServiceAccountSchema, newServiceAccountIamUpdater, serviceAccountIDParseFunc)
}
