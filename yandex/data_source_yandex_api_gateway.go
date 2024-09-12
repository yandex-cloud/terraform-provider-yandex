package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/apigateway/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

func dataSourceYandexApiGateway() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexApiGatewayRead,

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"api_gateway_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"folder_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"labels": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"user_domains": {
				Type:       schema.TypeSet,
				Computed:   true,
				Elem:       &schema.Schema{Type: schema.TypeString},
				Set:        schema.HashString,
				Deprecated: fieldDeprecatedForAnother("user_domains", "custom_domains"),
			},

			"custom_domains": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"domain_id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"fqdn": {
							Type:     schema.TypeString,
							Required: true,
						},
						"certificate_id": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},

			"domain": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"log_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"connectivity": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"network_id": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},

			"variables": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"canary": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"weight": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 99),
						},
						"variables": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
					},
				},
			},

			"log_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"disabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"log_group_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"folder_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"min_level": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"execution_timeout": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func dataSourceYandexApiGatewayRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	err := checkOneOf(d, "api_gateway_id", "name")
	if err != nil {
		return err
	}

	apiGatewayID := d.Get("api_gateway_id").(string)
	_, tgNameOk := d.GetOk("name")

	if tgNameOk {
		apiGatewayID, err = resolveObjectID(ctx, config, d, sdkresolvers.APIGatewayResolver)
		if err != nil {
			return fmt.Errorf("failed to resolve data source Yandex Cloud API Gateway by name: %v", err)
		}
	}

	req := apigateway.GetApiGatewayRequest{
		ApiGatewayId: apiGatewayID,
	}

	apiGateway, err := config.sdk.Serverless().APIGateway().ApiGateway().Get(ctx, &req)
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Yandex Cloud API Gateway %q", d.Id()))
	}

	d.SetId(apiGateway.Id)
	d.Set("api_gateway_id", apiGateway.Id)
	return flattenYandexApiGateway(d, apiGateway, true)
}
