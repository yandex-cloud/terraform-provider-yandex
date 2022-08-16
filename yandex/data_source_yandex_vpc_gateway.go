package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

func dataSourceYandexVPCGateway() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexVPCGatewayRead,
		Schema: map[string]*schema.Schema{
			"gateway_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"folder_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
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
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"shared_egress_gateway": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
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
