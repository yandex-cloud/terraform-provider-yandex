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
		WithDescription("Allows creation and management of a single binding within IAM policy for an existing Yandex KMS Asymmetric Signature Key.\n\n~> Roles controlled by `yandex_kms_asymmetric_signature_key_iam_binding` should not be assigned using `yandex_kms_asymmetric_signature_key_iam_member`.\n\n~> When you delete `yandex_kms_asymmetric_signature_key_iam_binding` resource, the roles can be deleted from other users within the folder as well. Be careful!\n"),
	)
}
