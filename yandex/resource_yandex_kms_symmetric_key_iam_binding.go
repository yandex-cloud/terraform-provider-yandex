package yandex

import "github.com/hashicorp/terraform-plugin-sdk/helper/schema"

func resourceYandexKMSSymmetricKeyIAMBinding() *schema.Resource {
	return resourceIamBindingWithImport(IamKMSSymmetricKeySchema, newKMSSymmetricKeyIamUpdater, kmsSymmetricKeyIDParseFunc)
}
