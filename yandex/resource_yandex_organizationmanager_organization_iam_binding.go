package yandex

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceYandexOrganizationManagerOrganizationIAMBinding() *schema.Resource {
	return resourceIamBinding(
		IamOrganizationSchema,
		newOrganizationIamUpdater,
		WithTimeout(
			&schema.ResourceTimeout{
				Default: schema.DefaultTimeout(yandexOrganizationManagerOrganizationDefaultTimeout),
			}),
		WithImporter(
			&schema.ResourceImporter{
				StateContext: iamBindingImport(organizationIDParseFunc),
			}),
		WithDescription("Allows creation and management of a single binding within IAM policy for an existing Yandex Cloud Organization Manager organization."),
	)
}
