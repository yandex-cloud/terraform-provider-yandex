package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
)

func dataSourceYandexVPCNetwork() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexVPCNetworkRead,
		Schema: map[string]*schema.Schema{
			"network_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
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
			"labels": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceYandexVPCNetworkRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := context.Background()
	var network *vpc.Network

	networkID := d.Get("network_id").(string)
	network, err := config.sdk.VPC().Network().Get(ctx, &vpc.GetNetworkRequest{
		NetworkId: networkID,
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("network with ID %q", networkID))
	}

	subnets, err := config.sdk.VPC().Network().ListSubnets(ctx, &vpc.ListNetworkSubnetsRequest{
		NetworkId: networkID,
	})

	if err != nil {
		return err
	}

	subnetIds := make([]string, len(subnets.Subnets))
	for i, subnet := range subnets.Subnets {
		subnetIds[i] = subnet.Id
	}

	createdAt, err := getTimestamp(network.CreatedAt)
	if err != nil {
		return err
	}

	d.Set("description", network.Description)
	d.Set("created_at", createdAt)
	d.Set("name", network.Name)
	d.Set("folder_id", network.FolderId)
	d.Set("labels", network.Labels)

	if err := d.Set("subnet_ids", subnetIds); err != nil {
		return err
	}

	d.SetId(network.Id)

	return nil
}
