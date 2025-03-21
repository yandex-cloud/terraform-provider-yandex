package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceYandexFunctionScalingPolicy() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about a Yandex Cloud Function Scaling Policy. For more information about Yandex Cloud Functions, see [Yandex Cloud Functions](https://yandex.cloud/docs/functions/).\n\nThis data source is used to define [Yandex Cloud Function Scaling Policy](https://yandex.cloud/docs/functions/) that can be used by other resources.\n",

		Read: dataSourceYandexFunctionScalingPolicyRead,

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"function_id": {
				Type:        schema.TypeString,
				Description: "Yandex Cloud Function id used to define function.",
				Required:    true,
			},

			"policy": {
				Type:     schema.TypeList,
				Computed: true,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"tag": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"zone_requests_limit": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
						},
						"zone_instances_limit": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
						},
					},
				},
			},
		},
	}
}

func dataSourceYandexFunctionScalingPolicyRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := config.Context()
	functionID := d.Get("function_id").(string)

	policies, err := fetchFunctionScalingPolicies(ctx, config, functionID)
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Yandex Cloud Function %s Scaling Policy", functionID))
	}

	d.SetId(functionID)
	return flattenYandexFunctionScalingPolicy(d, policies)
}
