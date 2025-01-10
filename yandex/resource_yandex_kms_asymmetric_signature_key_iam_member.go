package yandex

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceYandexKMSAsymmetricSignatureKeyIAMMember() *schema.Resource {
	return resourceIamMember(
		IamKMSAsymmetricSignatureKeySchema,
		newKMSAsymmetricSignatureKeyIamUpdater,
		WithTimeout(
			&schema.ResourceTimeout{
				Default: schema.DefaultTimeout(yandexIAMKMSDefaultTimeout),
			}),
		WithImporter(
			&schema.ResourceImporter{
				StateContext: iamMemberImport(kmsAsymmetricSignatureKeyIDParseFunc),
			}),
	)
}
