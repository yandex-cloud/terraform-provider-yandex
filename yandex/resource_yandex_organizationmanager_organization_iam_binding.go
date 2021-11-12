package yandex

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceYandexOrganizationManagerOrganizationIAMBinding() *schema.Resource {
	return resourceIamBindingWithImport(IamOrganizationSchema, newOrganizationIamUpdater, organizationIDParseFunc)
}
