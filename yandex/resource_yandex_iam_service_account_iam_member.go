package yandex

import "github.com/hashicorp/terraform/helper/schema"

func resourceYandexIAMServiceAccountIAMMember() *schema.Resource {
	return ResourceIamMemberWithImport(IamServiceAccountSchema, NewServiceAccountIamUpdater, ServiceAccountIDParseFunc)
}
