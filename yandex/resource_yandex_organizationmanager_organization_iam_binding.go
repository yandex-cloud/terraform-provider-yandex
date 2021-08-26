package yandex

import "github.com/hashicorp/terraform-plugin-sdk/helper/schema"

func resourceYandexOrganizationManagerOrganizationIAMBinding() *schema.Resource {
	return resourceIamBindingWithImport(IamOrganizationSchema, newOrganizationIamUpdater, organizationIDParseFunc)
}
