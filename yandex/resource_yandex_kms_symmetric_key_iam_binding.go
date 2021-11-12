package yandex

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceYandexKMSSymmetricKeyIAMBinding() *schema.Resource {
	return resourceIamBindingWithImport(IamKMSSymmetricKeySchema, newKMSSymmetricKeyIamUpdater, kmsSymmetricKeyIDParseFunc)
}
