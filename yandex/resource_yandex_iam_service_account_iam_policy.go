package yandex

import "github.com/hashicorp/terraform/helper/schema"

func resourceYandexIAMServiceAccountIAMPolicy() *schema.Resource {
	return ResourceIamPolicyWithImport(IamServiceAccountSchema, NewServiceAccountIamUpdater, ServiceAccountIDParseFunc)
}
