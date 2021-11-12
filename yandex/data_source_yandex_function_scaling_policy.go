package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceYandexFunctionScalingPolicy() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexFunctionScalingPolicyRead,

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"function_id": {
				Type:     schema.TypeString,
				Required: true,
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
