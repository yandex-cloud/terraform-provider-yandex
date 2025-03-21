package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

func dataSourceYandexVPCSubnet() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about a Yandex VPC subnet. For more information, see [Yandex Cloud VPC](https://yandex.cloud/docs/vpc/concepts/index).\n\nThis data source is used to define [VPC Subnets](https://yandex.cloud/docs/vpc/concepts/network#subnet) that can be used by other resources.\n\n~> One of `subnet_id` or `name` should be specified.\n",

		Read: dataSourceYandexVPCSubnetRead,
		Schema: map[string]*schema.Schema{
			"subnet_id": {
				Type:        schema.TypeString,
				Description: "Subnet ID.",
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
				Computed:    true,
				Optional:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["description"],
				Computed:    true,
			},
			"network_id": {
				Type:        schema.TypeString,
				Description: resourceYandexVPCSubnet().Schema["network_id"].Description,
				Computed:    true,
			},
			"labels": {
				Type:        schema.TypeMap,
				Description: common.ResourceDescriptions["labels"],
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},
			"zone": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["zone"],
				Computed:    true,
			},
			"route_table_id": {
				Type:        schema.TypeString,
				Description: resourceYandexVPCSubnet().Schema["route_table_id"].Description,
				Computed:    true,
			},
			"v4_cidr_blocks": {
				Type:        schema.TypeList,
				Description: resourceYandexVPCSubnet().Schema["v4_cidr_blocks"].Description,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"v6_cidr_blocks": {
				Type:        schema.TypeList,
				Description: resourceYandexVPCSubnet().Schema["v6_cidr_blocks"].Description,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"dhcp_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"domain_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"domain_name_servers": {
							Type:     schema.TypeList,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Computed: true,
						},
						"ntp_servers": {
							Type:     schema.TypeList,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Computed: true,
						},
					},
				},
			},
			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},
		},
	}
}

func dataSourceYandexVPCSubnetRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := config.Context()

	err := checkOneOf(d, "subnet_id", "name")
	if err != nil {
		return err
	}

	subnetID := d.Get("subnet_id").(string)
	_, subnetNameOk := d.GetOk("name")

	if subnetNameOk {
		subnetID, err = resolveObjectID(ctx, config, d, sdkresolvers.SubnetResolver)
		if err != nil {
			return fmt.Errorf("failed to resolve data source subnet by name: %v", err)
		}
	}

	subnet, err := config.sdk.VPC().Subnet().Get(ctx, &vpc.GetSubnetRequest{
		SubnetId: subnetID,
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("subnet with ID %q", subnetID))
	}

	d.Set("subnet_id", subnet.Id)
	d.Set("name", subnet.Name)
	d.Set("description", subnet.Description)
	d.Set("folder_id", subnet.FolderId)
	d.Set("created_at", getTimestamp(subnet.CreatedAt))
	d.Set("network_id", subnet.NetworkId)
	d.Set("zone", subnet.ZoneId)
	d.Set("route_table_id", subnet.RouteTableId)
	if err := d.Set("labels", subnet.Labels); err != nil {
		return err
	}
	if err := d.Set("v4_cidr_blocks", subnet.V4CidrBlocks); err != nil {
		return err
	}
	if err := d.Set("v6_cidr_blocks", subnet.V6CidrBlocks); err != nil {
		return err
	}
	if err := d.Set("dhcp_options", flattenDhcpOptions(subnet.DhcpOptions)); err != nil {
		return err
	}
	d.SetId(subnet.Id)

	return nil
}
