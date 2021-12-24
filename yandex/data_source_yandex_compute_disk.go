package yandex

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

func dataSourceYandexComputeDisk() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexComputeDiskRead,
		Schema: map[string]*schema.Schema{
			"disk_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"block_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"image_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"snapshot_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"product_ids": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"instance_ids": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},

			"disk_placement_policy": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"disk_placement_group_id": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceYandexComputeDiskRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := config.Context()

	err := checkOneOf(d, "disk_id", "name")
	if err != nil {
		return err
	}

	diskID := d.Get("disk_id").(string)
	_, diskNameOk := d.GetOk("name")

	if diskNameOk {
		diskID, err = resolveObjectID(ctx, config, d, sdkresolvers.DiskResolver)
		if err != nil {
			return fmt.Errorf("failed to resolve data source disk by name: %v", err)
		}
	}

	disk, err := config.sdk.Compute().Disk().Get(ctx, &compute.GetDiskRequest{
		DiskId: diskID,
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("disk with ID %q", diskID))
	}

	diskPlacementPolicy, err := flattenDiskPlacementPolicy(disk)
	if err != nil {
		return err
	}

	d.Set("disk_id", disk.Id)
	d.Set("folder_id", disk.FolderId)
	d.Set("created_at", getTimestamp(disk.CreatedAt))
	d.Set("name", disk.Name)
	d.Set("description", disk.Description)
	d.Set("type", disk.TypeId)
	d.Set("zone", disk.ZoneId)
	d.Set("size", toGigabytes(disk.Size))
	d.Set("block_size", int(disk.BlockSize))
	d.Set("status", strings.ToLower(disk.Status.String()))
	d.Set("image_id", disk.GetSourceImageId())
	d.Set("snapshot_id", disk.GetSourceSnapshotId())
	d.Set("disk_placement_policy", diskPlacementPolicy)

	if err := d.Set("instance_ids", disk.InstanceIds); err != nil {
		return err
	}

	if err := d.Set("labels", disk.Labels); err != nil {
		return err
	}

	if err := d.Set("product_ids", disk.ProductIds); err != nil {
		return err
	}

	d.SetId(disk.Id)

	return nil
}
