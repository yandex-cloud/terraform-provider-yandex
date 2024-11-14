package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/eventrouter/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

func dataSourceYandexEventrouterBus() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceYandexEventrouterBusRead,

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the bus",
			},

			"bus_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID of the bus",
			},

			"folder_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the folder that the bus belongs to",
			},

			"cloud_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the cloud that the bus resides in",
			},

			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation timestamp",
			},

			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of the bus",
			},

			"labels": {
				Type:        schema.TypeMap,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Description: "Bus labels",
			},

			"deletion_protection": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Deletion protection",
			},
		},
	}
}

func dataSourceYandexEventrouterBusRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	err := checkOneOf(d, "bus_id", "name")
	if err != nil {
		return diag.FromErr(err)
	}

	busID := d.Get("bus_id").(string)
	_, tgNameOk := d.GetOk("name")

	if tgNameOk {
		busID, err = resolveObjectID(ctx, config, d, sdkresolvers.EventrouterBusResolver)
		if err != nil {
			return diag.Errorf("failed to resolve data source Event Router bus by name: %v", err)
		}
	}

	req := eventrouter.GetBusRequest{
		BusId: busID,
	}

	bus, err := config.sdk.Serverless().Eventrouter().Bus().Get(ctx, &req)
	if err != nil {
		return diag.FromErr(handleNotFoundError(err, d, fmt.Sprintf("Event Router bus %q", d.Id())))
	}

	d.SetId(bus.Id)
	d.Set("bus_id", bus.Id)
	return diag.FromErr(flattenYandexEventrouterBus(d, bus))
}
