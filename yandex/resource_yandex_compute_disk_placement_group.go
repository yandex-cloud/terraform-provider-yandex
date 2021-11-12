package yandex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	"google.golang.org/genproto/protobuf/field_mask"
)

const yandexComputeDiskPlacementGroupDefaultTimeout = 1 * time.Minute

func resourceYandexComputeDiskPlacementGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexComputeDiskPlacementGroupCreate,
		Read:   resourceYandexComputeDiskPlacementGroupRead,
		Update: resourceYandexComputeDiskPlacementGroupUpdate,
		Delete: resourceYandexComputeDiskPlacementGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexComputeDiskPlacementGroupDefaultTimeout),
			Update: schema.DefaultTimeout(yandexComputeDiskPlacementGroupDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexComputeDiskPlacementGroupDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
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
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			// Conforms CLI behavior in regards to "zone_id"
			"zone": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "ru-central1-b",
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceYandexComputeDiskPlacementGroupCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	folderID, err := getFolderID(d, config)
	if err != nil {
		return fmt.Errorf("Error getting folder ID while creating Disk Placement Group: %s", err)
	}

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return fmt.Errorf("Error expanding labels while creating Disk Placement Group: %s", err)
	}

	req := compute.CreateDiskPlacementGroupRequest{
		FolderId:    folderID,
		Name:        d.Get("name").(string),
		Labels:      labels,
		Description: d.Get("description").(string),
		PlacementStrategy: &compute.CreateDiskPlacementGroupRequest_SpreadPlacementStrategy{
			SpreadPlacementStrategy: &compute.DiskSpreadPlacementStrategy{},
		},
		ZoneId: d.Get("zone").(string),
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Compute().DiskPlacementGroup().Create(ctx, &req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create Disk Placement Group: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("Error while get Disk Placement Group create operation metadata: %s", err)
	}

	md, ok := protoMetadata.(*compute.CreateDiskPlacementGroupMetadata)
	if !ok {
		return fmt.Errorf("could not get Disk Placement Group ID from create operation metadata")
	}

	d.SetId(md.GetDiskPlacementGroupId())

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while waiting operation to create Disk Placement Group: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("Disk Placement Group creation failed: %s", err)
	}

	return resourceYandexComputeDiskPlacementGroupRead(d, meta)
}

func resourceYandexComputeDiskPlacementGroupRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	placementGroup, err := config.sdk.Compute().DiskPlacementGroup().Get(context.Background(),
		&compute.GetDiskPlacementGroupRequest{
			DiskPlacementGroupId: d.Id(),
		})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Disk Placement Group %q", d.Id()))
	}

	d.Set("created_at", getTimestamp(placementGroup.CreatedAt))
	d.Set("name", placementGroup.Name)
	d.Set("folder_id", placementGroup.FolderId)
	d.Set("description", placementGroup.Description)
	d.Set("zone", placementGroup.ZoneId)
	d.Set("status", placementGroup.Status.String())

	return d.Set("labels", placementGroup.Labels)
}

func resourceYandexComputeDiskPlacementGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	req := &compute.UpdateDiskPlacementGroupRequest{
		DiskPlacementGroupId: d.Id(),
		UpdateMask:           &field_mask.FieldMask{},
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

	if len(req.UpdateMask.Paths) == 0 {
		return fmt.Errorf("No fields were updated for Disk Placement Group %s", d.Id())
	}

	err := makeDiskPlacementGroupUpdateRequest(req, d, meta)
	if err != nil {
		return err
	}

	return resourceYandexComputeDiskPlacementGroupRead(d, meta)
}

func resourceYandexComputeDiskPlacementGroupDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Deleting Disk Placement Group %q", d.Id())

	req := &compute.DeleteDiskPlacementGroupRequest{
		DiskPlacementGroupId: d.Id(),
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Compute().DiskPlacementGroup().Delete(ctx, req))
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Disk Placement Group %q", d.Id()))
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished deleting Disk Placement Group %q", d.Id())
	return nil
}

func makeDiskPlacementGroupUpdateRequest(req *compute.UpdateDiskPlacementGroupRequest, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Compute().DiskPlacementGroup().Update(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to update Disk Placement Group %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error updating Disk Placement Group %q: %s", d.Id(), err)
	}

	return nil
}
