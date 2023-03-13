package yandex

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

const billingCloudServiceInstanceBindingType = "cloud"
const billingCloudIdBindingFieldName = "cloud_id"

func resourceYandexBillingCloudBinding() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceYandexBillingServiceInstanceBindingCreate(billingCloudServiceInstanceBindingType, billingCloudIdBindingFieldName),
		ReadContext:   resourceYandexBillingServiceInstanceBindingRead(billingCloudServiceInstanceBindingType, billingCloudIdBindingFieldName),
		UpdateContext: resourceYandexBillingServiceInstanceBindingUpdate(billingCloudServiceInstanceBindingType, billingCloudIdBindingFieldName),
		DeleteContext: resourceYandexBillingServiceInstanceBindingDelete(billingCloudServiceInstanceBindingType, billingCloudIdBindingFieldName),

		Schema: map[string]*schema.Schema{
			"billing_account_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			billingCloudIdBindingFieldName: {
				Type:     schema.TypeString,
				Required: true,
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}
