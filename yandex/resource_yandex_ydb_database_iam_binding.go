package yandex

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceYandexYDBDatabaseIAMBinding() *schema.Resource {
	return resourceIamBindingWithImport(IamYDBDatabaseSchema, newYDBDatabaseIamUpdater, ydbDatabaseIDParseFunc)
}
