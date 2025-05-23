package yandex

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

const (
	yandexComputeDiskDefaultTimeout = 5 * time.Minute
	yandexComputeDiskMoveTimeout    = 1 * time.Minute
)

func resourceYandexComputeDisk() *schema.Resource {
	return &schema.Resource{
		Description: "Persistent disks are used for data storage and function similarly to physical hard and solid state drives.\n\nA disk can be attached or detached from the virtual machine and can be located locally. A disk can be moved between virtual machines within the same availability zone. Each disk can be attached to only one virtual machine at a time.\n\nFor more information about disks in Yandex Cloud, see:\n* [Documentation](https://yandex.cloud/docs/compute/concepts/disk)\n* How-to Guides:\n  * [Attach and detach a disk](https://yandex.cloud/docs/compute/concepts/disk#attach-detach)\n  * [Backup operation](https://yandex.cloud/docs/compute/concepts/disk#backup)\n\n~> Only one of `image_id` or `snapshot_id` can be specified.\n",

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
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Optional:    true,
				Default:     "",
			},

			"description": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["description"],
				Optional:    true,
			},

			"folder_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["folder_id"],
				Computed:    true,
				Optional:    true,
			},

			"labels": {
				Type:        schema.TypeMap,
				Description: common.ResourceDescriptions["labels"],
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},

			"zone": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["zone"],
				Computed:    true,
				Optional:    true,
				ForceNew:    true,
			},

			"size": {
				Type:         schema.TypeInt,
				Description:  "Size of the persistent disk, specified in GB. You can specify this field when creating a persistent disk using the `image_id` or `snapshot_id` parameter, or specify it alone to create an empty persistent disk. If you specify this field along with `image_id` or `snapshot_id`, the size value must not be less than the size of the source image or the size of the snapshot.",
				Optional:     true,
				Default:      150,
				ValidateFunc: validation.IntAtLeast(0),
			},

			"block_size": {
				Type:         schema.TypeInt,
				Description:  "Block size of the disk, specified in bytes.",
				Optional:     true,
				ForceNew:     true,
				Default:      4096,
				ValidateFunc: validation.IntAtLeast(0),
			},

			"image_id": {
				Type:          schema.TypeString,
				Description:   "The source image to use for disk creation.",
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"snapshot_id"},
			},

			"snapshot_id": {
				Type:          schema.TypeString,
				Description:   "The source snapshot to use for disk creation.",
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"image_id"},
			},

			"type": {
				Type:        schema.TypeString,
				Description: "Type of disk to create. Provide this when creating a disk.",
				Optional:    true,
				ForceNew:    true,
				Default:     "network-hdd",
			},

			"status": {
				Type:        schema.TypeString,
				Description: "The status of the disk.",
				Computed:    true,
			},

			"disk_placement_policy": {
				Type:        schema.TypeList,
				Description: "Disk placement policy configuration.",
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"disk_placement_group_id": {
							Type:        schema.TypeString,
							Description: "Specifies Disk Placement Group id.",
							Required:    true,
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
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},

			"allow_recreate": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"hardware_generation": {
				Type:        schema.TypeList,
				Description: "Hardware generation and its features, which will be applied to the instance when this disk is used as a boot disk. Provide this property if you wish to override this value, which otherwise is inherited from the source.",
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"generation2_features": {
							Type:        schema.TypeList,
							Description: "A newer hardware generation, which always uses `PCI_TOPOLOGY_V2` and UEFI boot.",
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{},
							},
							Optional: true,
							ForceNew: true,
							Computed: true,
						},

						"legacy_features": {
							Type:        schema.TypeList,
							Description: "Defines the first known hardware generation and its features.",
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"pci_topology": {
										Type:         schema.TypeString,
										Description:  "A variant of PCI topology, one of `PCI_TOPOLOGY_V1` or `PCI_TOPOLOGY_V2`.",
										Optional:     true,
										ForceNew:     true,
										Computed:     true,
										ValidateFunc: validateParsableValue(parseComputePCITopology),
									},
								},
							},
							Optional: true,
							ForceNew: true,
							Computed: true,
						},
					},
				},
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"kms_key_id": {
				Type:        schema.TypeString,
				Description: "ID of KMS symmetric key used to encrypt disk.",
				ForceNew:    true,
				Optional:    true,
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

	hardwareGeneration, err := expandHardwareGeneration(d)
	if err != nil {
		return fmt.Errorf("Error expanding hardware generation while creating disk: %s", err)
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
		HardwareGeneration:  hardwareGeneration,
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

	if v, ok := d.GetOk("kms_key_id"); ok {
		req.KmsKeyId = v.(string)
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

	hardwareGeneration, err := flattenComputeHardwareGeneration(disk.HardwareGeneration)
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

	if disk.KmsKey != nil {
		d.Set("kms_key_id", disk.KmsKey.KeyId)
	}

	if err := d.Set("product_ids", disk.ProductIds); err != nil {
		return err
	}
	if err := d.Set("hardware_generation", hardwareGeneration); err != nil {
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
