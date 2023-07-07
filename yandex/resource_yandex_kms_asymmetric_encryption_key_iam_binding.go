package yandex

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceYandexKMSAsymmetricEncryptionKeyIAMBinding() *schema.Resource {
	return resourceIamBinding(
		IamKMSAsymmetricEncryptionKeySchema,
		newKMSAsymmetricEncryptionKeyIamUpdater,
		WithTimeout(
			&schema.ResourceTimeout{
				Default: schema.DefaultTimeout(yandexIAMKMSDefaultTimeout),
			}),
		WithImporter(
			&schema.ResourceImporter{
				StateContext: iamBindingImport(kmsAsymmetricEncryptionKeyIDParseFunc),
			}),
	)
}
