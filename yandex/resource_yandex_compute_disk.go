package yandex

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
)

const yandexComputeDiskDefaultTimeout = 5 * time.Minute

func resourceYandexComputeDisk() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexComputeDiskCreate,
		Read:   resourceYandexComputeDiskRead,
		Update: resourceYandexComputeDiskUpdate,
		Delete: resourceYandexComputeDiskDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexComputeDiskDefaultTimeout),
			Update: schema.DefaultTimeout(yandexComputeDiskDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexComputeDiskDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"zone": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"size": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"image_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"snapshot_id"},
			},
			"snapshot_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"image_id"},
			},
			"type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "network-hdd",
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"source_image_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"source_snapshot_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"product_ids": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
		},
	}

}

func resourceYandexComputeDiskCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	zone, err := getZone(d, config)
	if err != nil {
		return fmt.Errorf("Error creating disk: %s", err)
	}

	folderID, err := getFolderID(d, config)
	if err != nil {
		return fmt.Errorf("Error creating disk: %s", err)
	}

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return fmt.Errorf("Error creating disk: %s", err)
	}

	req := compute.CreateDiskRequest{
		FolderId:    folderID,
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Labels:      labels,
		TypeId:      d.Get("type").(string),
		ZoneId:      zone,
		Size:        toBytes(d.Get("size").(int)),
	}

	if v, ok := d.GetOk("image_id"); ok {
		req.Source = &compute.CreateDiskRequest_ImageId{
			ImageId: v.(string),
		}
	} else if v, ok := d.GetOk("snapshot_id"); ok {
		req.Source = &compute.CreateDiskRequest_SnapshotId{
			SnapshotId: v.(string),
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Compute().Disk().Create(ctx, &req))
	if err != nil {
		return fmt.Errorf("Error creating disk: %s", err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error create disk: %s", err)
	}

	resp, err := op.Response()
	if err != nil {
		return err
	}

	disk, ok := resp.(*compute.Disk)
	if !ok {
		return errors.New("response doesn't contain Disk")
	}

	d.SetId(disk.Id)

	return resourceYandexComputeDiskRead(d, meta)
}

func resourceYandexComputeDiskRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	disk, err := config.sdk.Compute().Disk().Get(context.Background(), &compute.GetDiskRequest{
		DiskId: d.Id(),
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Disk %q", d.Get("name").(string)))
	}

	d.Set("name", disk.Name)
	d.Set("folder_id", disk.FolderId)
	d.Set("zone", disk.ZoneId)
	d.Set("description", disk.Description)
	d.Set("status", strings.ToLower(disk.Status.String()))
	d.Set("type", disk.TypeId)
	d.Set("size", toGigabytes(disk.Size))
	if err := d.Set("product_ids", disk.ProductIds); err != nil {
		return err
	}
	if err := d.Set("labels", disk.Labels); err != nil {
		return err
	}
	d.Set("source_image_id", disk.GetSourceImageId())
	d.Set("source_snapshot_id", disk.GetSourceSnapshotId())

	return nil
}

func resourceYandexComputeDiskUpdate(d *schema.ResourceData, meta interface{}) error {

	d.Partial(true)

	labelPropName := "labels"
	if d.HasChange(labelPropName) {
		labelsProp, err := expandLabels(d.Get(labelPropName))
		if err != nil {
			return err
		}

		req := &compute.UpdateDiskRequest{
			DiskId: d.Id(),
			Labels: labelsProp,
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{labelPropName},
			},
		}

		err = makeDiskUpdateRequest(req, d, meta)
		if err != nil {
			return err
		}

		d.SetPartial(labelPropName)
	}

	namePropName := "name"
	if d.HasChange(namePropName) {
		req := &compute.UpdateDiskRequest{
			DiskId: d.Id(),
			Name:   d.Get(namePropName).(string),
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{namePropName},
			},
		}

		err := makeDiskUpdateRequest(req, d, meta)
		if err != nil {
			return err
		}

		d.SetPartial(namePropName)
	}

	descPropName := "description"
	if d.HasChange(descPropName) {
		req := &compute.UpdateDiskRequest{
			DiskId: d.Id(),
			Name:   d.Get(descPropName).(string),
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{descPropName},
			},
		}

		err := makeDiskUpdateRequest(req, d, meta)
		if err != nil {
			return err
		}

		d.SetPartial(descPropName)
	}

	sizePropName := "size"
	if d.HasChange(sizePropName) {
		req := &compute.UpdateDiskRequest{
			DiskId: d.Id(),
			Size:   toBytes(d.Get(sizePropName).(int)),
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{sizePropName},
			},
		}

		err := makeDiskUpdateRequest(req, d, meta)
		if err != nil {
			return err
		}

		d.SetPartial(sizePropName)
	}

	d.Partial(false)

	return resourceYandexComputeDiskRead(d, meta)
}

func resourceYandexComputeDiskDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	// TODO: We need API to lookup Disk Usages to Attach/Detach disk properly before delete
	// if disks are attached, they must be detached before the disk can be deleted

	log.Printf("[DEBUG] Deleting Disk %q", d.Id())

	req := &compute.DeleteDiskRequest{
		DiskId: d.Id(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Compute().Disk().Delete(ctx, req))
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Disk %q", d.Get("name").(string)))
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	resp, err := op.Response()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished deleting Disk %q: %#v", d.Id(), resp)
	return nil
}

func makeDiskUpdateRequest(req *compute.UpdateDiskRequest, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Compute().Disk().Update(ctx, req))
	if err != nil {
		return err
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error updating Disk %q: %s", d.Id(), err)
	}

	return nil
}
