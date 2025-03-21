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
		WithDescription("Allows creation and management of a single member for a single binding within the IAM policy for an existing Yandex KMS Symmetric Key.\n\n~> Roles controlled by `yandex_kms_symmetric_key_iam_binding` should not be assigned using `yandex_kms_symmetric_key_iam_member`.\n"),
	)
}
