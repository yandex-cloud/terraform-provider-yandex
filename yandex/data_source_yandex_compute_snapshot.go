package yandex

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

func dataSourceYandexComputeSnapshot() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexComputeSnapshotRead,
		Schema: map[string]*schema.Schema{
			"snapshot_id": {
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
			"hardware_generation": {
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"legacy_features": {
							Type: schema.TypeList,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"pci_topology": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
							Computed: true,
						},

						"generation2_features": {
							Type: schema.TypeList,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{},
							},
							Computed: true,
						},
					},
				},
				Computed: true,
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}

}

func dataSourceYandexComputeSnapshotRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := config.Context()

	err := checkOneOf(d, "snapshot_id", "name")
	if err != nil {
		return err
	}

	snapshotID := d.Get("snapshot_id").(string)
	_, snapshotNameOk := d.GetOk("name")

	if snapshotNameOk {
		snapshotID, err = resolveObjectID(ctx, config, d, sdkresolvers.SnapshotResolver)
		if err != nil {
			return fmt.Errorf("failed to resolve data source snapshot by name: %v", err)
		}
	}

	snapshot, err := config.sdk.Compute().Snapshot().Get(ctx, &compute.GetSnapshotRequest{
		SnapshotId: snapshotID,
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("snapshot with ID %q", snapshotID))
	}

	hardwareGeneration, err := flattenComputeHardwareGeneration(snapshot.HardwareGeneration)
	if err != nil {
		return err
	}

	d.Set("snapshot_id", snapshot.Id)
	d.Set("folder_id", snapshot.FolderId)
	d.Set("created_at", getTimestamp(snapshot.CreatedAt))
	d.Set("name", snapshot.Name)
	d.Set("description", snapshot.Description)
	d.Set("storage_size", toGigabytes(snapshot.StorageSize))
	d.Set("disk_size", toGigabytes(snapshot.DiskSize))
	d.Set("status", strings.ToLower(snapshot.Status.String()))
	d.Set("source_disk_id", snapshot.GetSourceDiskId())

	if snapshot.KmsKey != nil {
		d.Set("kms_key_id", snapshot.KmsKey.KeyId)
	}

	if err := d.Set("labels", snapshot.Labels); err != nil {
		return err
	}

	if err := d.Set("product_ids", snapshot.ProductIds); err != nil {
		return err
	}

	if err := d.Set("hardware_generation", hardwareGeneration); err != nil {
		return err
	}

	d.SetId(snapshot.Id)

	return nil
}
