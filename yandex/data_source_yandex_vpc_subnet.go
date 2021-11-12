package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

func dataSourceYandexVPCSubnet() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexVPCSubnetRead,
		Schema: map[string]*schema.Schema{
			"subnet_id": {
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
				Computed: true,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"network_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"route_table_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"v4_cidr_blocks": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"v6_cidr_blocks": {
				Type:     schema.TypeList,
				Computed: true,
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
				Type:     schema.TypeString,
				Computed: true,
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
