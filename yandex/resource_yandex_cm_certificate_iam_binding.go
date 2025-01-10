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
	)
}
