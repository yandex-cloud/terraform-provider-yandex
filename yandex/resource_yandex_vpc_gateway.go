package yandex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
)

const yandexVPCGatewayDefaultTimeout = 1 * time.Minute

func resourceYandexVPCGateway() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexVPCGatewayCreate,
		Read:   resourceYandexVPCGatewayRead,
		Update: resourceYandexVPCGatewayUpdate,
		Delete: resourceYandexVPCGatewayDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexVPCGatewayDefaultTimeout),
			Update: schema.DefaultTimeout(yandexVPCGatewayDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexVPCGatewayDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},

			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"shared_egress_gateway": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{},
				},
				ForceNew: true,
			},
		},
	}

}

func resourceYandexVPCGatewayCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return fmt.Errorf("Error expanding labels while creating gateway: %s", err)
	}

	folderID, err := getFolderID(d, config)
	if err != nil {
		return fmt.Errorf("Error getting folder ID while creating gateway: %s", err)
	}

	req := vpc.CreateGatewayRequest{
		FolderId:    folderID,
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Labels:      labels,
	}

	if d.Get("shared_egress_gateway") != nil {
		req.Gateway = &vpc.CreateGatewayRequest_SharedEgressGatewaySpec{
			SharedEgressGatewaySpec: &vpc.SharedEgressGatewaySpec{},
		}
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.VPC().Gateway().Create(ctx, &req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create VPC Gateway: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("Error while get gateway create operation metadata: %s", err)
	}

	md, ok := protoMetadata.(*vpc.CreateGatewayMetadata)
	if !ok {
		return fmt.Errorf("could not get Gateway ID from create operation metadata")
	}

	d.SetId(md.GatewayId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while waiting operation to create gateway: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("Gateway creation failed: %s", err)
	}

	return resourceYandexVPCGatewayRead(d, meta)
}

func resourceYandexVPCGatewayRead(d *schema.ResourceData, meta interface{}) error {
	return yandexVPCGatewayRead(d, meta, d.Id())
}

func yandexVPCGatewayRead(d *schema.ResourceData, meta interface{}, id string) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutRead))
	defer cancel()

	gateway, err := config.sdk.VPC().Gateway().Get(ctx, &vpc.GetGatewayRequest{
		GatewayId: id,
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("VPC Gateway %q", d.Get("name").(string)))
	}

	d.Set("created_at", getTimestamp(gateway.CreatedAt))
	d.Set("name", gateway.Name)
	d.Set("folder_id", gateway.FolderId)
	d.Set("description", gateway.Description)

	sharedEgressGateway := flattenSharedEgressGateway(gateway.GetSharedEgressGateway())
	if err := d.Set("shared_egress_gateway", sharedEgressGateway); err != nil {
		return err
	}

	return d.Set("labels", gateway.Labels)
}

func resourceYandexVPCGatewayUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	d.Partial(true)

	req := &vpc.UpdateGatewayRequest{
		GatewayId:  d.Id(),
		UpdateMask: &field_mask.FieldMask{},
	}

	if d.HasChange("labels") {
		labelsProp, err := expandLabels(d.Get("labels"))
		if err != nil {
			return err
		}

		req.Labels = labelsProp
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "labels")
	}

	if d.HasChange("name") {
		req.Name = d.Get("name").(string)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "name")
	}

	if d.HasChange("description") {
		req.Description = d.Get("description").(string)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "description")
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.VPC().Gateway().Update(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to update VPC Gateway %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error updating VPC Gateway %q: %s", d.Id(), err)
	}

	d.Partial(false)

	return resourceYandexVPCGatewayRead(d, meta)
}

func resourceYandexVPCGatewayDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Deleting VPC Gateway %q", d.Id())

	req := &vpc.DeleteGatewayRequest{
		GatewayId: d.Id(),
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.VPC().Gateway().Delete(ctx, req))
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("VPC Gateway %q", d.Get("name").(string)))
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished deleting VPC Gateway %q", d.Id())
	return nil
}
