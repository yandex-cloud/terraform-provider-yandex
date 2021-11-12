package yandex

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceYandexIAMServiceAccountIAMPolicy() *schema.Resource {
	return resourceIamPolicyWithImport(IamServiceAccountSchema, newServiceAccountIamUpdater, serviceAccountIDParseFunc)
}
