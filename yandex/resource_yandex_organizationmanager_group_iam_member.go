package yandex

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceYandexOrganizationManagerGroupIAMMember() *schema.Resource {
	return resourceIamMember(
		IamGroupSchema,
		newGroupIamUpdater,
		WithTimeout(
			&schema.ResourceTimeout{
				Default: schema.DefaultTimeout(yandexOrganizationManagerGroupDefaultTimeout),
			}),
		WithImporter(
			&schema.ResourceImporter{
				StateContext: iamMemberImport(groupIDParseFunc),
			}),
	)
}
