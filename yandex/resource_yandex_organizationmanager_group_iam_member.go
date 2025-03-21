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
		WithDescription("Allows creation and management of a single member for a single binding within the IAM policy for an existing Yandex Organization Manager Group."),
	)
}
