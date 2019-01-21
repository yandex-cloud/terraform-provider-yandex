package yandex

import "github.com/hashicorp/terraform/helper/schema"

func resourceYandexIAMServiceAccountIAMBinding() *schema.Resource {
	return ResourceIamBindingWithImport(IamServiceAccountSchema, NewServiceAccountIamUpdater, ServiceAccountIDParseFunc)
}
