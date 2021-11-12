package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

func dataSourceYandexVPCNetwork() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexVPCNetworkRead,
		Schema: map[string]*schema.Schema{
			"network_id": {
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
			"subnet_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"default_security_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceYandexVPCNetworkRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := config.Context()

	err := checkOneOf(d, "network_id", "name")
	if err != nil {
		return err
	}

	networkID := d.Get("network_id").(string)
	_, networkNameOk := d.GetOk("name")

	if networkNameOk {
		networkID, err = resolveObjectID(ctx, config, d, sdkresolvers.NetworkResolver)
		if err != nil {
			return fmt.Errorf("failed to resolve data source network by name: %v", err)
		}
	}

	d.SetId(networkID)

	if err := d.Set("network_id", networkID); err != nil {
		return err
	}

	return yandexVPCNetworkRead(d, meta, networkID)
}
