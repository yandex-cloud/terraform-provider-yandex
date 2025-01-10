package yandex

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceYandexKMSAsymmetricEncryptionKeyIAMMember() *schema.Resource {
	return resourceIamMember(
		IamKMSAsymmetricEncryptionKeySchema,
		newKMSAsymmetricEncryptionKeyIamUpdater,
		WithTimeout(
			&schema.ResourceTimeout{
				Default: schema.DefaultTimeout(yandexIAMKMSDefaultTimeout),
			}),
		WithImporter(
			&schema.ResourceImporter{
				StateContext: iamMemberImport(kmsAsymmetricEncryptionKeyIDParseFunc),
			}),
	)
}
