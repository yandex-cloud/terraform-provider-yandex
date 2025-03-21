package yandex

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceYandexCMCertificateIAMBinding() *schema.Resource {
	return resourceIamBinding(
		IamCMCertificateSchema,
		newCMCertificateIamUpdater,
		WithTimeout(
			&schema.ResourceTimeout{
				Default: schema.DefaultTimeout(yandexIAMCMDefaultTimeout),
			}),
		WithImporter(
			&schema.ResourceImporter{
				StateContext: iamBindingImport(CMCertificateIDParseFunc),
			}),
		WithDescription("Allows creation and management of a single binding within IAM policy for an existing Certificate.\n\n~> Roles controlled by `yandex_cm_certificate_iam_binding` should not be assigned using `yandex_cm_certificate_iam_member`.\n\n~> When you delete `yandex_cm_certificate_iam_binding` resource, the roles can be deleted from other users within the folder as well. Be careful!\n\n"),
	)
}
