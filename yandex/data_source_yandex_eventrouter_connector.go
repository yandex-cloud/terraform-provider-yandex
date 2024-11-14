package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/eventrouter/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

func dataSourceYandexEventrouterConnector() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceYandexEventrouterConnectorRead,

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the connector",
			},

			"connector_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID of the connector",
			},

			"bus_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the bus that the connector belongs to",
			},

			"folder_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the folder that the connector resides in",
			},

			"cloud_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the cloud that the connector resides in",
			},

			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation timestamp",
			},

			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of the connector",
			},

			"labels": {
				Type:        schema.TypeMap,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Description: "Connector labels",
			},

			eventrouterSourceTypeYds: {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Data Stream source of the connector.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"database": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Stream database. Example: /ru-central1/aoegtvhtp8ob********/cc8004q4lbo6********",
						},
						"stream_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Stream name, absolute or relative",
						},
						"consumer": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Consumer name",
						},
						"service_account_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Service account which has read permission on the stream",
						},
					},
				},
			},

			eventrouterSourceTypeYmq: {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Message Queue source of the connector.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"queue_arn": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Required field. Queue ARN. Example: yrn:yc:ymq:ru-central1:aoe***:test",
						},
						"service_account_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Service account which has read access to the queue",
						},
						"visibility_timeout": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Queue visibility timeout override",
						},
						"batch_size": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Batch size for polling",
						},
						"polling_timeout": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Queue polling timeout",
						},
					},
				},
			},

			"deletion_protection": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Deletion protection",
			},
		},
	}
}

func dataSourceYandexEventrouterConnectorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	err := checkOneOf(d, "connector_id", "name")
	if err != nil {
		return diag.FromErr(err)
	}

	connectorID := d.Get("connector_id").(string)
	_, tgNameOk := d.GetOk("name")

	if tgNameOk {
		connectorID, err = resolveObjectID(ctx, config, d, sdkresolvers.EventrouterConnectorResolver)
		if err != nil {
			return diag.Errorf("failed to resolve data source Event Router connector by name: %v", err)
		}
	}

	req := eventrouter.GetConnectorRequest{
		ConnectorId: connectorID,
	}

	connector, err := config.sdk.Serverless().Eventrouter().Connector().Get(ctx, &req)
	if err != nil {
		return diag.FromErr(handleNotFoundError(err, d, fmt.Sprintf("Event Router connector %q", d.Id())))
	}

	d.SetId(connector.Id)
	d.Set("connector_id", connector.Id)
	flattenYandexEventrouterConnector(d, connector)
	return nil
}
