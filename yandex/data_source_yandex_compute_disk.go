package yandex

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
)

func dataSourceYandexComputeDisk() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexComputeDiskRead,
		Schema: map[string]*schema.Schema{
			"disk_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
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
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}

}

func dataSourceYandexComputeDiskRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := context.Background()
	var disk *compute.Disk

	diskID := d.Get("disk_id").(string)
	disk, err := config.sdk.Compute().Disk().Get(ctx, &compute.GetDiskRequest{
		DiskId: diskID,
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("disk with ID %q", diskID))
	}

	createdAt, err := getTimestamp(disk.CreatedAt)
	if err != nil {
		return err
	}

	d.Set("folder_id", disk.FolderId)
	d.Set("created_at", createdAt)
	d.Set("name", disk.Name)
	d.Set("description", disk.Description)
	d.Set("labels", disk.Labels)
	d.Set("type", disk.TypeId)
	d.Set("zone", disk.ZoneId)
	d.Set("size", toGigabytes(disk.Size))
	d.Set("product_ids", disk.ProductIds)
	d.Set("status", strings.ToLower(disk.Status.String()))
	d.Set("source_image_id", disk.GetSourceImageId())
	d.Set("source_snapshot_id", disk.GetSourceSnapshotId())
	d.Set("instance_ids", disk.InstanceIds)
	d.SetId(disk.Id)

	return nil
}
