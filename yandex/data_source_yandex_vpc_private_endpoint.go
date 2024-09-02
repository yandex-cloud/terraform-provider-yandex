package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

func dataSourceYandexVPCPrivateEndpoint() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexVPCPrivateEndpointRead,
		Schema: map[string]*schema.Schema{
			"private_endpoint_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
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
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"network_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"object_storage": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{},
				},
			},
			"endpoint_address": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subnet_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"address_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"dns_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"private_dns_records_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceYandexVPCPrivateEndpointRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := config.Context()

	err := checkOneOf(d, "private_endpoint_id", "name")
	if err != nil {
		return err
	}

	peID := d.Get("private_endpoint_id").(string)
	_, peNameOk := d.GetOk("name")

	if peNameOk {
		peID, err = resolveObjectID(ctx, config, d, sdkresolvers.PrivateEndpointResolver)
		if err != nil {
			return fmt.Errorf("failed to resolve data source private endpoint by name: %v", err)
		}
	}

	d.SetId(peID)

	if err := d.Set("private_endpoint_id", peID); err != nil {
		return err
	}

	return yandexVPCPrivateEndpointRead(d, meta, peID)
}
