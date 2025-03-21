package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

func dataSourceYandexVPCRouteTable() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about a Yandex VPC route table. For more information, see [Yandex Cloud VPC](https://yandex.cloud/docs/vpc/concepts).\n\nThis data source is used to define [VPC Route Table](https://yandex.cloud/docs/vpc/concepts/) that can be used by other resources.\n\n~> One of `route_table_id` or `name` should be specified.\n",

		Read: dataSourceYandexVPCRouteTableRead,
		Schema: map[string]*schema.Schema{
			"route_table_id": {
				Type:        schema.TypeString,
				Description: "Route table ID.",
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
			"network_id": {
				Type:        schema.TypeString,
				Description: resourceYandexVPCRouteTable().Schema["network_id"].Description,
				Computed:    true,
			},
			"labels": {
				Type:        schema.TypeMap,
				Description: common.ResourceDescriptions["labels"],
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},
			"static_route": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"destination_prefix": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"next_hop_address": {
							Type:     schema.TypeString,
							Computed: true,
							Optional: true,
						},
						"gateway_id": {
							Type:     schema.TypeString,
							Computed: true,
							Optional: true,
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

func dataSourceYandexVPCRouteTableRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := config.Context()

	err := checkOneOf(d, "route_table_id", "name")
	if err != nil {
		return err
	}

	routeTableID := d.Get("route_table_id").(string)
	_, routeTableNameOk := d.GetOk("name")

	if routeTableNameOk {
		routeTableID, err = resolveObjectID(ctx, config, d, sdkresolvers.RouteTableResolver)
		if err != nil {
			return fmt.Errorf("failed to resolve data source route table by name: %v", err)
		}
	}

	routeTable, err := config.sdk.VPC().RouteTable().Get(ctx, &vpc.GetRouteTableRequest{
		RouteTableId: routeTableID,
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("route table with ID %q", routeTableID))
	}

	d.Set("route_table_id", routeTable.Id)
	d.Set("name", routeTable.Name)
	d.Set("description", routeTable.Description)
	d.Set("folder_id", routeTable.FolderId)
	d.Set("created_at", getTimestamp(routeTable.CreatedAt))
	d.Set("network_id", routeTable.NetworkId)
	if err := d.Set("labels", routeTable.Labels); err != nil {
		return err
	}

	staticRoutes := flattenStaticRoutes(routeTable)
	if err := d.Set("static_route", staticRoutes); err != nil {
		return err
	}

	d.SetId(routeTable.Id)

	return nil
}
