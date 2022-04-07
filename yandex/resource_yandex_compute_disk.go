package yandex

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
)

const (
	yandexComputeDiskDefaultTimeout = 5 * time.Minute
	yandexComputeDiskMoveTimeout    = 1 * time.Minute
)

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

			"block_size": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
				Default:  4096,
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

			"disk_placement_policy": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"disk_placement_group_id": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
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

			"allow_recreate": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}

}

func expandDiskPlacementPolicy(d *schema.ResourceData) (*compute.DiskPlacementPolicy, error) {
	sp := d.Get("disk_placement_policy").([]interface{})
	var placementPolicy *compute.DiskPlacementPolicy
	if len(sp) != 0 {
		placementPolicy = &compute.DiskPlacementPolicy{
			PlacementGroupId: d.Get("disk_placement_policy.0.disk_placement_group_id").(string),
		}
	}
	return placementPolicy, nil
}

func flattenDiskPlacementPolicy(disk *compute.Disk) ([]map[string]interface{}, error) {
	diskPlacementPolicy := make([]map[string]interface{}, 0, 1)
	diskPlacementMap := map[string]interface{}{
		"disk_placement_group_id": disk.DiskPlacementPolicy.PlacementGroupId,
	}
	diskPlacementPolicy = append(diskPlacementPolicy, diskPlacementMap)
	return diskPlacementPolicy, nil
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

	diskPlacementPolicy, err := expandDiskPlacementPolicy(d)
	if err != nil {
		return fmt.Errorf("Error expanding disk placement policy while creating disk: %s", err)
	}

	req := compute.CreateDiskRequest{
		FolderId:            folderID,
		Name:                d.Get("name").(string),
		Description:         d.Get("description").(string),
		Labels:              labels,
		TypeId:              d.Get("type").(string),
		ZoneId:              zone,
		Size:                toBytes(d.Get("size").(int)),
		BlockSize:           int64(d.Get("block_size").(int)),
		DiskPlacementPolicy: diskPlacementPolicy,
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

	diskPlacementPolicy, err := flattenDiskPlacementPolicy(disk)
	if err != nil {
		return err
	}

	d.Set("created_at", getTimestamp(disk.CreatedAt))
	d.Set("name", disk.Name)
	d.Set("folder_id", disk.FolderId)
	d.Set("zone", disk.ZoneId)
	d.Set("description", disk.Description)
	d.Set("status", strings.ToLower(disk.Status.String()))
	d.Set("type", disk.TypeId)
	d.Set("size", toGigabytes(disk.Size))
	d.Set("block_size", int(disk.BlockSize))
	d.Set("image_id", disk.GetSourceImageId())
	d.Set("snapshot_id", disk.GetSourceSnapshotId())
	d.Set("disk_placement_policy", diskPlacementPolicy)

	if err := d.Set("product_ids", disk.ProductIds); err != nil {
		return err
	}

	return d.Set("labels", disk.Labels)
}

func resourceYandexComputeDiskUpdate(d *schema.ResourceData, meta interface{}) error {
	d.Partial(true)

	folderPropName := "folder_id"
	if d.HasChange(folderPropName) {
		if !d.Get("allow_recreate").(bool) {
			req := &compute.MoveDiskRequest{
				DiskId:              d.Id(),
				DestinationFolderId: d.Get(folderPropName).(string),
			}

			if err := makeDiskMoveRequest(req, d, meta); err != nil {
				return err
			}
		} else {
			if err := resourceYandexComputeDiskDelete(d, meta); err != nil {
				return err
			}
			if err := resourceYandexComputeDiskCreate(d, meta); err != nil {
				return err
			}
		}
	}

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

	}

	placementPolicyPropName := "disk_placement_policy"
	if d.HasChange(placementPolicyPropName) {
		req := &compute.UpdateDiskRequest{
			DiskId: d.Id(),
			DiskPlacementPolicy: &compute.DiskPlacementPolicy{
				PlacementGroupId: d.Get("disk_placement_policy.0.disk_placement_group_id").(string),
			},
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{"disk_placement_policy.placement_group_id"},
			},
		}

		err := makeDiskUpdateRequest(req, d, meta)
		if err != nil {
			return err
		}

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

	}

	d.Partial(false)

	return resourceYandexComputeDiskRead(d, meta)
}

func resourceYandexComputeDiskDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	disk, err := config.sdk.Compute().Disk().Get(config.Context(), &compute.GetDiskRequest{
		DiskId: d.Id(),
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Disk %q", d.Get("name").(string)))
	}

	for _, instanceID := range disk.GetInstanceIds() {
		req := &compute.DetachInstanceDiskRequest{
			InstanceId: instanceID,
			Disk: &compute.DetachInstanceDiskRequest_DiskId{
				DiskId: disk.Id,
			},
		}
		if err := makeDetachDiskRequest(req, meta); err != nil {
			return err
		}
		log.Printf("[DEBUG] Successfully detached disk %s from instance %s", disk.Id, instanceID)
	}

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

	_, err = op.Response()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished deleting Disk %q", d.Id())
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

func makeDiskMoveRequest(req *compute.MoveDiskRequest, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), yandexComputeDiskMoveTimeout)
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Compute().Disk().Move(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to move Disk %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error moving Disk %q: %s", d.Id(), err)
	}

	return nil
}

func isDiskSizeDecreased(ctx context.Context, old, new, _ interface{}) bool {
	if old == nil || new == nil {
		return false
	}
	return new.(int) < old.(int)
}

func validateDiskSize(v interface{}, _ string) (warnings []string, errors []error) {
	value := v.(int)
	if value < 0 || value > 8192 {
		errors = append(errors, fmt.Errorf(
			"The `size` can only be between 0 and 8192"))
	}
	return warnings, errors
}
