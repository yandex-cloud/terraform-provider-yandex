package yandex

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/eventrouter/v1"
	"google.golang.org/genproto/protobuf/field_mask"
)

const (
	yandexEventrouterBusDefaultTimeout = 10 * time.Minute
)

func resourceYandexEventrouterBus() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceYandexEventrouterBusCreate,
		ReadContext:   resourceYandexEventrouterBusRead,
		UpdateContext: resourceYandexEventrouterBusUpdate,
		DeleteContext: resourceYandexEventrouterBusDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexEventrouterBusDefaultTimeout),
			Read:   schema.DefaultTimeout(yandexEventrouterBusDefaultTimeout),
			Update: schema.DefaultTimeout(yandexEventrouterBusDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexEventrouterBusDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the bus",
			},

			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the bus",
			},

			"folder_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				ForceNew:    true,
				Description: "ID of the folder that the bus belongs to",
			},

			"cloud_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of the cloud that the bus resides in",
			},

			"labels": {
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Description: "Bus labels",
			},

			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation timestamp",
			},

			"deletion_protection": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Deletion protection",
			},
		},
	}
}

func resourceYandexEventrouterBusCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return diag.Errorf("Error expanding labels while creating Event Router bus: %s", err)
	}

	folderID, err := getFolderID(d, config)
	if err != nil {
		return diag.Errorf("Error getting folder ID while creating Event Router bus: %s", err)
	}

	req := eventrouter.CreateBusRequest{
		FolderId:           folderID,
		Name:               d.Get("name").(string),
		Description:        d.Get("description").(string),
		Labels:             labels,
		DeletionProtection: d.Get("deletion_protection").(bool),
	}

	op, err := config.sdk.WrapOperation(config.sdk.Serverless().Eventrouter().Bus().Create(ctx, &req))
	if err != nil {
		return diag.Errorf("Error while requesting API to create Event Router bus: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return diag.Errorf("Error while requesting API to create Event Router bus: %s", err)
	}

	md, ok := protoMetadata.(*eventrouter.CreateBusMetadata)
	if !ok {
		return diag.Errorf("Could not get Event Router bus ID from create operation metadata")
	}

	d.SetId(md.BusId)

	err = op.Wait(ctx)
	if err != nil {
		return diag.Errorf("Error while requesting API to create Event Router bus: %s", err)
	}

	return resourceYandexEventrouterBusRead(ctx, d, meta)
}

func resourceYandexEventrouterBusUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return diag.Errorf("Error expanding labels while updating Event Router bus: %s", err)
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

	if len(updatePaths) != 0 {
		req := eventrouter.UpdateBusRequest{
			BusId:       d.Id(),
			UpdateMask:  &field_mask.FieldMask{Paths: updatePaths},
			Name:        d.Get("name").(string),
			Description: d.Get("description").(string),
			Labels:      labels,
		}

		op, err := config.sdk.Serverless().Eventrouter().Bus().Update(ctx, &req)
		err = waitOperation(ctx, config, op, err)
		if err != nil {
			return diag.Errorf("Error while requesting API to update Event Router bus: %s", err)
		}
	}

	return resourceYandexEventrouterBusRead(ctx, d, meta)
}

func resourceYandexEventrouterBusRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	req := eventrouter.GetBusRequest{
		BusId: d.Id(),
	}

	bus, err := config.sdk.Serverless().Eventrouter().Bus().Get(ctx, &req)
	if err != nil {
		return diag.FromErr(handleNotFoundError(err, d, fmt.Sprintf("Event Router bus %q", d.Id())))
	}

	return diag.FromErr(flattenYandexEventrouterBus(d, bus))
}

func resourceYandexEventrouterBusDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	req := eventrouter.DeleteBusRequest{
		BusId: d.Id(),
	}

	op, err := config.sdk.Serverless().Eventrouter().Bus().Delete(ctx, &req)
	err = waitOperation(ctx, config, op, err)
	if err != nil {
		return diag.FromErr(handleNotFoundError(err, d, fmt.Sprintf("Event Router bus %q", d.Id())))
	}

	return nil
}

func flattenYandexEventrouterBus(
	d *schema.ResourceData,
	bus *eventrouter.Bus,
) error {
	d.Set("name", bus.Name)
	d.Set("folder_id", bus.FolderId)
	d.Set("cloud_id", bus.CloudId)
	d.Set("created_at", getTimestamp(bus.CreatedAt))
	d.Set("description", bus.Description)
	d.Set("labels", bus.Labels)
	d.Set("deletion_protection", bus.DeletionProtection)

	return nil
}
