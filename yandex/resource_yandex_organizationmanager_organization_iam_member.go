package yandex

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceYandexOrganizationManagerOrganizationIAMMember() *schema.Resource {
	return resourceIamMember(
		IamOrganizationSchema,
		newOrganizationIamUpdater,
		WithTimeout(
			&schema.ResourceTimeout{
				Default: schema.DefaultTimeout(yandexOrganizationManagerOrganizationDefaultTimeout),
			}),
		WithImporter(
			&schema.ResourceImporter{
				StateContext: iamMemberImport(organizationIDParseFunc),
			}),
		WithDescription("Allows creation and management of a single member for a single binding within the IAM policy for an existing Yandex Organization Manager organization.\n\n~> Roles controlled by `yandex_organizationmanager_organization_iam_binding` should not be assigned using `yandex_organizationmanager_organization_iam_member`.\n\n~> When you delete `yandex_organizationmanager_organization_iam_binding` resource, the roles can be deleted from other users within the organization as well. Be careful!\n"),
	)
}
