package yandex

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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

		CustomizeDiff: customdiff.ForceNewIfChange("size", isDiskSizeDecreased),

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

			"zone": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},

			"size": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      150,
				ValidateFunc: validateDiskSize,
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

			"product_ids": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}

}

func resourceYandexComputeDiskCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	zone, err := getZone(d, config)
	if err != nil {
		return fmt.Errorf("Error getting zone while creating disk: %s", err)
	}

	folderID, err := getFolderID(d, config)
	if err != nil {
		return fmt.Errorf("Error getting folder ID while creating disk: %s", err)
	}

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return fmt.Errorf("Error expanding labels while creating disk: %s", err)
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

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Compute().Disk().Create(ctx, &req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create disk: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("Error while get disk create operation metadata: %s", err)
	}

	md, ok := protoMetadata.(*compute.CreateDiskMetadata)
	if !ok {
		return fmt.Errorf("could not get Disk ID from create operation metadata")
	}

	d.SetId(md.DiskId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while waiting operation to create disk: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("Disk creation failed: %s", err)
	}

	return resourceYandexComputeDiskRead(d, meta)
}

func resourceYandexComputeDiskRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	disk, err := config.sdk.Compute().Disk().Get(config.Context(), &compute.GetDiskRequest{
		DiskId: d.Id(),
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Disk %q", d.Get("name").(string)))
	}

	createdAt, err := getTimestamp(disk.CreatedAt)
	if err != nil {
		return err
	}

	d.Set("created_at", createdAt)
	d.Set("name", disk.Name)
	d.Set("folder_id", disk.FolderId)
	d.Set("zone", disk.ZoneId)
	d.Set("description", disk.Description)
	d.Set("status", strings.ToLower(disk.Status.String()))
	d.Set("type", disk.TypeId)
	d.Set("size", toGigabytes(disk.Size))
	d.Set("image_id", disk.GetSourceImageId())
	d.Set("snapshot_id", disk.GetSourceSnapshotId())

	if err := d.Set("product_ids", disk.ProductIds); err != nil {
		return err
	}

	return d.Set("labels", disk.Labels)
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
			DiskId:      d.Id(),
			Description: d.Get(descPropName).(string),
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

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutDelete))
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

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Compute().Disk().Update(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to update Disk %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error updating Disk %q: %s", d.Id(), err)
	}

	return nil
}

func isDiskSizeDecreased(old, new, _ interface{}) bool {
	if old == nil || new == nil {
		return false
	}
	return new.(int) < old.(int)
}

func validateDiskSize(v interface{}, _ string) (warnings []string, errors []error) {
	value := v.(int)
	if value < 0 || value > 4096 {
		errors = append(errors, fmt.Errorf(
			"The `size` can only be between 0 and 4096"))
	}
	return warnings, errors
}
