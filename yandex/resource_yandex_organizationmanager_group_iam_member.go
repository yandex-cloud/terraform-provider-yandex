package yandex

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceYandexOrganizationManagerGroupIAMMember() *schema.Resource {
	return resourceIamMemberWithImport(IamGroupSchema, newGroupIamUpdater, groupIDParseFunc)
}
