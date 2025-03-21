package yandex

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceYandexCMCertificateIAMMember() *schema.Resource {
	return resourceIamMember(
		IamCMCertificateSchema,
		newCMCertificateIamUpdater,
		WithTimeout(
			&schema.ResourceTimeout{
				Default: schema.DefaultTimeout(yandexIAMCMDefaultTimeout),
			}),
		WithImporter(
			&schema.ResourceImporter{
				StateContext: iamMemberImport(CMCertificateIDParseFunc),
			}),
		WithDescription("Allows creation and management of a single member for a single binding within the IAM policy for an existing Certificate.\n\n~> Roles controlled by yandex_cm_certificate_iam_binding should not be assigned using yandex_cm_certificate_iam_member.\n"),
	)
}
