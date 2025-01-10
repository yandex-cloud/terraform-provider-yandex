package yandex

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceYandexKMSSymmetricKeyIAMMember() *schema.Resource {
	return resourceIamMember(
		IamKMSSymmetricKeySchema,
		newKMSSymmetricKeyIamUpdater,
		WithTimeout(
			&schema.ResourceTimeout{
				Default: schema.DefaultTimeout(yandexIAMKMSDefaultTimeout),
			}),
		WithImporter(
			&schema.ResourceImporter{
				StateContext: iamMemberImport(kmsSymmetricKeyIDParseFunc),
			}),
	)
}
