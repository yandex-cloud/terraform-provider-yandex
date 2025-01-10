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
	)
}
