package yandex

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceYandexOrganizationManagerOrganizationIAMMember() *schema.Resource {
	return resourceIamMemberWithImport(IamOrganizationSchema, newOrganizationIamUpdater, organizationIDParseFunc)
}
