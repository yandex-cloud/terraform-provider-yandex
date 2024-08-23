package yandex

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
)

func dataSourceYandexIamServiceAgent() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceYandexIamServiceAgentRead,
		Schema: map[string]*schema.Schema{
			"cloud_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"service_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"microservice_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func dataSourceYandexIamServiceAgentRead(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	cloudID := d.Get("cloud_id").(string)
	serviceID := d.Get("service_id").(string)
	microserviceID := d.Get("microservice_id").(string)

	serviceAgent, err := config.sdk.IAM().ServiceControl().ResolveAgent(context,
		&iam.ResolveServiceAgentRequest{
			ServiceId:      serviceID,
			MicroserviceId: microserviceID,
			Resource:       &iam.Resource{Type: "resource-manager.cloud", Id: cloudID},
		})

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(serviceAgent.GetServiceAccountId())
	d.Set("service_account_id", serviceAgent.GetServiceAccountId())
	d.Set("service_id", serviceAgent.GetServiceId())
	d.Set("microservice_id", serviceAgent.GetMicroserviceId())

	return nil
}
