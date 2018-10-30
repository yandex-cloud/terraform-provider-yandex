package yandex

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
)

func dataSourceYandexComputeSnapshot() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexComputeSnapshotRead,
		Schema: map[string]*schema.Schema{
			"snapshot_id": {
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
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"storage_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"disk_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"source_disk_id": {
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
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}

}

func dataSourceYandexComputeSnapshotRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := context.Background()
	var snapshot *compute.Snapshot

	snapshotID := d.Get("snapshot_id").(string)
	snapshot, err := config.sdk.Compute().Snapshot().Get(ctx, &compute.GetSnapshotRequest{
		SnapshotId: snapshotID,
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("snapshot with ID %q", snapshotID))
	}

	createdAt, err := getTimestamp(snapshot.CreatedAt)
	if err != nil {
		return err
	}

	d.Set("folder_id", snapshot.FolderId)
	d.Set("created_at", createdAt)
	d.Set("name", snapshot.Name)
	d.Set("description", snapshot.Description)
	d.Set("labels", snapshot.Labels)
	d.Set("product_ids", snapshot.ProductIds)
	d.Set("storage_size", toGigabytes(snapshot.StorageSize))
	d.Set("disk_size", toGigabytes(snapshot.DiskSize))
	d.Set("status", strings.ToLower(snapshot.Status.String()))
	d.Set("source_disk_id", snapshot.GetSourceDiskId())
	d.SetId(snapshot.Id)

	return nil
}
