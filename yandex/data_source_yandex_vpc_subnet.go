package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
)

func dataSourceYandexVPCSubnet() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexVPCSubnetRead,
		Schema: map[string]*schema.Schema{
			"subnet_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"folder_id": {
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
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceYandexVPCSubnetRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := context.Background()
	var subnet *vpc.Subnet

	subnetID := d.Get("subnet_id").(string)
	subnet, err := config.sdk.VPC().Subnet().Get(ctx, &vpc.GetSubnetRequest{
		SubnetId: subnetID,
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("subnet with ID %q", subnetID))
	}

	createdAt, err := getTimestamp(subnet.CreatedAt)
	if err != nil {
		return err
	}

	d.Set("name", subnet.Name)
	d.Set("description", subnet.Description)
	d.Set("folder_id", subnet.FolderId)
	d.Set("created_at", createdAt)
	d.Set("network_id", subnet.NetworkId)
	d.Set("zone", subnet.ZoneId)
	if err := d.Set("labels", subnet.Labels); err != nil {
		return err
	}
	if err := d.Set("v4_cidr_blocks", subnet.V4CidrBlocks); err != nil {
		return err
	}
	if err := d.Set("v6_cidr_blocks", subnet.V6CidrBlocks); err != nil {
		return err
	}
	d.SetId(subnet.Id)

	return nil
}
