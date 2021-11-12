package yandex

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceYandexIAMServiceAccountIAMMember() *schema.Resource {
	return resourceIamMemberWithImport(IamServiceAccountSchema, newServiceAccountIamUpdater, serviceAccountIDParseFunc)
}
