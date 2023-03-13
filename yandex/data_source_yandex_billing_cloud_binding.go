package yandex

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func dataSourceYandexBillingCloudBinding() *schema.Resource {
	const serviceInstanceType = "cloud"
	const idFieldName = "cloud_id"

	return &schema.Resource{
		ReadContext: dataSourceYandexBillingServiceInstanceBindingRead(serviceInstanceType, idFieldName),

		Schema: map[string]*schema.Schema{
			"billing_account_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			idFieldName: {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}
