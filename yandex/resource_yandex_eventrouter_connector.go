package yandex

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/eventrouter/v1"
	"google.golang.org/genproto/protobuf/field_mask"
)

const (
	yandexEventrouterConnectorDefaultTimeout = 10 * time.Minute

	eventrouterSourceTypeYds = "yds"
	eventrouterSourceTypeYmq = "ymq"
)

var eventrouterSourceTypesList = []string{
	eventrouterSourceTypeYds,
	eventrouterSourceTypeYmq,
}

func resourceYandexEventrouterConnector() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceYandexEventrouterConnectorCreate,
		ReadContext:   resourceYandexEventrouterConnectorRead,
		UpdateContext: resourceYandexEventrouterConnectorUpdate,
		DeleteContext: resourceYandexEventrouterConnectorDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexEventrouterConnectorDefaultTimeout),
			Read:   schema.DefaultTimeout(yandexEventrouterConnectorDefaultTimeout),
			Update: schema.DefaultTimeout(yandexEventrouterConnectorDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexEventrouterConnectorDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the connector",
			},

			"bus_id": {
				Type:        schema.TypeString,
				Required:    true,
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
				Optional:    true,
				Description: "Description of the connector",
			},

			"labels": {
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Description: "Connector labels",
			},

			eventrouterSourceTypeYds: {
				Type:          schema.TypeList,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: yandexEventrouterSourceConflictingTypes(eventrouterSourceTypeYds),
				Description:   "Data Stream source of the connector",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"database": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Stream database",
						},
						"stream_name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Stream name, absolute or relative",
						},
						"consumer": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Consumer name",
						},
						"service_account_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Service account which has read permission on the stream",
						},
					},
				},
			},

			eventrouterSourceTypeYmq: {
				Type:          schema.TypeList,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: yandexEventrouterSourceConflictingTypes(eventrouterSourceTypeYmq),
				Description:   "Message Queue source of the connector",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"queue_arn": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Queue ARN. Example: yrn:yc:ymq:ru-central1:aoe***:test",
						},
						"service_account_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Service account which has read access to the queue",
						},
						"visibility_timeout": {
							Type:        schema.TypeString,
							Computed:    true,
							Optional:    true,
							Description: "Queue visibility timeout override",
						},
						"batch_size": {
							Type:        schema.TypeInt,
							Computed:    true,
							Optional:    true,
							Description: "Batch size for polling",
						},
						"polling_timeout": {
							Type:        schema.TypeString,
							Computed:    true,
							Optional:    true,
							Description: "Queue polling timeout",
						},
					},
				},
			},

			"deletion_protection": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Deletion protection",
			},
		},
	}
}

func resourceYandexEventrouterConnectorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return diag.Errorf("Error expanding labels while creating Event Router connector: %s", err)
	}

	busId := d.Get("bus_id").(string)

	source, err := constructYandexEventrouterSource(d)
	if err != nil {
		return diag.Errorf("Error constructing Event Router connector source: %s", err)
	}

	req := eventrouter.CreateConnectorRequest{
		BusId:              busId,
		Name:               d.Get("name").(string),
		Description:        d.Get("description").(string),
		Labels:             labels,
		Source:             source,
		DeletionProtection: d.Get("deletion_protection").(bool),
	}

	op, err := config.sdk.WrapOperation(config.sdk.Serverless().Eventrouter().Connector().Create(ctx, &req))
	if err != nil {
		return diag.Errorf("Error while requesting API to create Event Router connector: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return diag.Errorf("Error while requesting API to create Event Router connector: %s", err)
	}

	md, ok := protoMetadata.(*eventrouter.CreateConnectorMetadata)
	if !ok {
		return diag.Errorf("Could not get Event Router connector ID from create operation metadata")
	}

	d.SetId(md.ConnectorId)

	err = op.Wait(ctx)
	if err != nil {
		return diag.Errorf("Error while requesting API to create Event Router connector: %s", err)
	}

	return resourceYandexEventrouterConnectorRead(ctx, d, meta)
}

func resourceYandexEventrouterConnectorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	req := eventrouter.GetConnectorRequest{
		ConnectorId: d.Id(),
	}

	connector, err := config.sdk.Serverless().Eventrouter().Connector().Get(ctx, &req)
	if err != nil {
		return diag.FromErr(handleNotFoundError(err, d, fmt.Sprintf("Event Router connector %q", d.Id())))
	}

	flattenYandexEventrouterConnector(d, connector)
	return nil
}

func resourceYandexEventrouterConnectorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return diag.Errorf("Error expanding labels while updating Event Router connector: %s", err)
	}

	var updatePaths []string
	if d.HasChange("name") {
		updatePaths = append(updatePaths, "name")
	}

	if d.HasChange("description") {
		updatePaths = append(updatePaths, "description")
	}

	if d.HasChange("labels") {
		updatePaths = append(updatePaths, "labels")
	}

	if d.HasChange("deletion_protection") {
		updatePaths = append(updatePaths, "deletion_protection")
	}

	if len(updatePaths) != 0 {
		req := eventrouter.UpdateConnectorRequest{
			ConnectorId:        d.Id(),
			UpdateMask:         &field_mask.FieldMask{Paths: updatePaths},
			Name:               d.Get("name").(string),
			Description:        d.Get("description").(string),
			Labels:             labels,
			DeletionProtection: d.Get("deletion_protection").(bool),
		}

		op, err := config.sdk.Serverless().Eventrouter().Connector().Update(ctx, &req)
		err = waitOperation(ctx, config, op, err)
		if err != nil {
			return diag.Errorf("Error while requesting API to update Event Router connector: %s", err)
		}
	}

	return resourceYandexEventrouterConnectorRead(ctx, d, meta)
}

func resourceYandexEventrouterConnectorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	req := eventrouter.DeleteConnectorRequest{
		ConnectorId: d.Id(),
	}

	op, err := config.sdk.Serverless().Eventrouter().Connector().Delete(ctx, &req)
	err = waitOperation(ctx, config, op, err)
	if err != nil {
		return diag.FromErr(handleNotFoundError(err, d, fmt.Sprintf("Event Router connector %q", d.Id())))
	}

	return nil
}

func flattenYandexEventrouterConnector(
	d *schema.ResourceData,
	connector *eventrouter.Connector,
) {
	d.Set("name", connector.Name)
	d.Set("bus_id", connector.BusId)
	d.Set("folder_id", connector.FolderId)
	d.Set("cloud_id", connector.CloudId)
	d.Set("created_at", getTimestamp(connector.CreatedAt))
	d.Set("description", connector.Description)
	d.Set("labels", connector.Labels)
	flattenYandexEventrouterSource(d, connector.Source)
	d.Set("deletion_protection", connector.DeletionProtection)
}

func flattenYandexEventrouterSource(
	d *schema.ResourceData,
	source *eventrouter.Source,
) {
	switch s := source.Source.(type) {
	case *eventrouter.Source_DataStream:
		yds := s.DataStream
		d.Set(eventrouterSourceTypeYds, [1]map[string]interface{}{
			{
				"database":           yds.Database,
				"stream_name":        yds.StreamName,
				"consumer":           yds.Consumer,
				"service_account_id": yds.ServiceAccountId,
			},
		})
	case *eventrouter.Source_MessageQueue:
		ymq := s.MessageQueue
		d.Set(eventrouterSourceTypeYmq, [1]map[string]interface{}{
			{
				"queue_arn":          ymq.QueueArn,
				"service_account_id": ymq.ServiceAccountId,
				"visibility_timeout": formatDuration(ymq.VisibilityTimeout),
				"batch_size":         ymq.BatchSize,
				"polling_timeout":    formatDuration(ymq.PollingTimeout),
			},
		})
	}
}

func yandexEventrouterSourceConflictingTypes(sourceType string) []string {
	res := make([]string, 0, len(eventrouterSourceTypesList)-1)
	for _, sType := range eventrouterSourceTypesList {
		if sType != sourceType {
			res = append(res, sType)
		}
	}
	return res
}

func constructYandexEventrouterSource(d *schema.ResourceData) (*eventrouter.Source, error) {
	if _, ok := d.GetOk(eventrouterSourceTypeYds); ok {
		yds := &eventrouter.Source_DataStream{
			DataStream: &eventrouter.DataStream{
				Database:         d.Get("yds.0.database").(string),
				StreamName:       d.Get("yds.0.stream_name").(string),
				Consumer:         d.Get("yds.0.consumer").(string),
				ServiceAccountId: d.Get("yds.0.service_account_id").(string),
			},
		}
		return &eventrouter.Source{Source: yds}, nil
	} else if _, ok := d.GetOk(eventrouterSourceTypeYmq); ok {
		vt, err := parseDuration(d.Get("ymq.0.visibility_timeout").(string))
		if err != nil {
			return nil, fmt.Errorf("Incorrect Yandex Message Queue visibility timeout: %s", err)
		}
		pt, err := parseDuration(d.Get("ymq.0.polling_timeout").(string))
		if err != nil {
			return nil, fmt.Errorf("Incorrect Yandex Message Queue polling timeout: %s", err)
		}
		ymq := &eventrouter.Source_MessageQueue{
			MessageQueue: &eventrouter.MessageQueue{
				QueueArn:          d.Get("ymq.0.queue_arn").(string),
				ServiceAccountId:  d.Get("ymq.0.service_account_id").(string),
				VisibilityTimeout: vt,
				BatchSize:         int64(d.Get("ymq.0.batch_size").(int)),
				PollingTimeout:    pt,
			},
		}
		return &eventrouter.Source{Source: ymq}, nil
	}

	return nil, errors.New("Source not specified")
}
