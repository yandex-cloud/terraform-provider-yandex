package yandex

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceYandexKMSAsymmetricSignatureKeyIAMBinding() *schema.Resource {
	return resourceIamBinding(
		IamKMSAsymmetricSignatureKeySchema,
		newKMSAsymmetricSignatureKeyIamUpdater,
		WithTimeout(
			&schema.ResourceTimeout{
				Default: schema.DefaultTimeout(yandexIAMKMSDefaultTimeout),
			}),
		WithImporter(
			&schema.ResourceImporter{
				StateContext: iamBindingImport(kmsAsymmetricSignatureKeyIDParseFunc),
			}),
	)
}
