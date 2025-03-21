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
		WithDescription("Allows creation and management of a single member for a single binding within the IAM policy for an existing Yandex KMS Asymmetric Signature Key.\n\n~> Roles controlled by `yandex_kms_asymmetric_signature_key_iam_binding` should not be assigned using `yandex_kms_asymmetric_signature_key_iam_member`.\n"),
	)
}
