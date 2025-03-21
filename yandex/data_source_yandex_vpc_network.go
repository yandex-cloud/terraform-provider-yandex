package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-sdk/sdkresolvers"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

func dataSourceYandexVPCNetwork() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about a Yandex VPC network. For more information, see [Yandex Cloud VPC](https://yandex.cloud/docs/vpc/concepts/index).\n\nThis data source is used to define [VPC Networks](https://yandex.cloud/docs/vpc/concepts/network) that can be used by other resources.\n\n~> One of `network_id` or `name` should be specified.\n",

		Read: dataSourceYandexVPCNetworkRead,
		Schema: map[string]*schema.Schema{
			"network_id": {
				Type:        schema.TypeString,
				Description: "ID of the network.",
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
			"subnet_ids": {
				Type:        schema.TypeList,
				Description: common.ResourceDescriptions["subnet_ids"],
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"default_security_group_id": {
				Type:        schema.TypeString,
				Description: resourceYandexVPCNetwork().Schema["default_security_group_id"].Description,
				Computed:    true,
			},
			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
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
