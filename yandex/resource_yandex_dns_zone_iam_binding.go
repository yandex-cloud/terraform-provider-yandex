package yandex

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceYandexDnsZoneIAMBinding() *schema.Resource {
	return resourceIamBinding(
		IamDnsZoneSchema,
		newDnsZoneIamUpdater,
		WithTimeout(
			&schema.ResourceTimeout{
				Default: schema.DefaultTimeout(yandexIAMDnsZoneDefaultTimeout),
			}),
		WithImporter(
			&schema.ResourceImporter{
				StateContext: iamBindingImport(dnsZoneIDParseFunc),
			}),
	)
}
