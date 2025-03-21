package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-sdk/sdkresolvers"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

func dataSourceYandexVPCGateway() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about a Yandex VPC gateway. For more information, see [Yandex Cloud VPC](https://yandex.cloud/docs/vpc/concepts).\n\nThis data source is used to define [VPC Gateways](https://yandex.cloud/docs/vpc/concepts/gateways) that can be used by other resources.\n\n~> One of `gateway_id` or `name` should be specified.\n",

		Read: dataSourceYandexVPCGatewayRead,
		Schema: map[string]*schema.Schema{
			"gateway_id": {
				Type:        schema.TypeString,
				Description: "ID of the VPC Gateway.",
				Optional:    true,
				Computed:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Optional:    true,
				Computed:    true,
			},
			"folder_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["folder_id"],
				Optional:    true,
				Computed:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["description"],
				Computed:    true,
			},
			"labels": {
				Type:        schema.TypeMap,
				Description: common.ResourceDescriptions["labels"],
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},
			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},
			"shared_egress_gateway": {
				Type:        schema.TypeList,
				Description: resourceYandexVPCGateway().Schema["shared_egress_gateway"].Description,
				MaxItems:    1,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{},
				},
			},
		},
	}
}

func dataSourceYandexVPCGatewayRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := config.Context()

	err := checkOneOf(d, "gateway_id", "name")
	if err != nil {
		return err
	}

	gatewayID := d.Get("gateway_id").(string)
	_, gatewayNameOk := d.GetOk("name")

	if gatewayNameOk {
		gatewayID, err = resolveObjectID(ctx, config, d, sdkresolvers.GatewayResolver)
		if err != nil {
			return fmt.Errorf("failed to resolve data source gateway by name: %v", err)
		}
	}

	d.SetId(gatewayID)

	if err := d.Set("gateway_id", gatewayID); err != nil {
		return err
	}

	return yandexVPCGatewayRead(d, meta, gatewayID)
}
