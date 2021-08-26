package yandex

import "github.com/hashicorp/terraform-plugin-sdk/helper/schema"

func resourceYandexOrganizationManagerOrganizationIAMMember() *schema.Resource {
	return resourceIamMemberWithImport(IamOrganizationSchema, newOrganizationIamUpdater, organizationIDParseFunc)
}
