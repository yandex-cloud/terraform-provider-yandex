package yandex

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceYandexYDBDatabaseIAMBinding() *schema.Resource {
	return resourceIamBinding(
		IamYDBDatabaseSchema,
		newYDBDatabaseIamUpdater,
		WithTimeout(
			&schema.ResourceTimeout{
				Default: schema.DefaultTimeout(yandexIAMYDBDefaultTimeout),
			}),
		WithImporter(
			&schema.ResourceImporter{
				StateContext: iamBindingImport(ydbDatabaseIDParseFunc),
			}),
	)
}
