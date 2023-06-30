package yandex

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceYandexKMSSymmetricKeyIAMBinding() *schema.Resource {
	return resourceIamBinding(
		IamKMSSymmetricKeySchema,
		newKMSSymmetricKeyIamUpdater,
		WithTimeout(
			&schema.ResourceTimeout{
				Default: schema.DefaultTimeout(yandexIAMKMSDefaultTimeout),
			}),
		WithImporter(
			&schema.ResourceImporter{
				StateContext: iamBindingImport(kmsSymmetricKeyIDParseFunc),
			}),
	)
}
