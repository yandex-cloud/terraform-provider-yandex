package yandex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
)

const yandexComputeSnapshotDefaultTimeout = 20 * time.Minute

func resourceYandexComputeSnapshot() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexComputeSnapshotCreate,
		Read:   resourceYandexComputeSnapshotRead,
		Update: resourceYandexComputeSnapshotUpdate,
		Delete: resourceYandexComputeSnapshotDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexComputeSnapshotDefaultTimeout),
			Update: schema.DefaultTimeout(yandexComputeSnapshotDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexComputeSnapshotDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"source_disk_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

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

			"disk_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"storage_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"hardware_generation": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"generation2_features": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{},
							},
							Optional: true,
							ForceNew: true,
							Computed: true,
						},

						"legacy_features": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"pci_topology": {
										Type:         schema.TypeString,
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
		},
	}

}

func resourceYandexComputeSnapshotCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	folderID, err := getFolderID(d, config)
	if err != nil {
		return fmt.Errorf("Error getting folder ID while creating snapshot: %s", err)
	}

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return fmt.Errorf("Error expanding labels while creating snapshot: %s", err)
	}

	hardwareGeneration, err := expandHardwareGeneration(d)
	if err != nil {
		return fmt.Errorf("Error expanding hardware generation while creating snapshot: %s", err)
	}

	req := compute.CreateSnapshotRequest{
		FolderId:           folderID,
		DiskId:             d.Get("source_disk_id").(string),
		Name:               d.Get("name").(string),
		Description:        d.Get("description").(string),
		Labels:             labels,
		HardwareGeneration: hardwareGeneration,
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Compute().Snapshot().Create(ctx, &req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create snapshot: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("Error while get snapshot create operation metadata: %s", err)
	}

	md, ok := protoMetadata.(*compute.CreateSnapshotMetadata)
	if !ok {
		return fmt.Errorf("could not get Snapshot ID from create operation metadata")
	}

	d.SetId(md.SnapshotId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while waiting operation to create snapshot: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("Snapshot creation failed: %s", err)
	}

	return resourceYandexComputeSnapshotRead(d, meta)
}

func resourceYandexComputeSnapshotRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	snapshot, err := config.sdk.Compute().Snapshot().Get(config.Context(), &compute.GetSnapshotRequest{
		SnapshotId: d.Id(),
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Snapshot %q", d.Get("name").(string)))
	}

	hardwareGeneration, err := flattenComputeHardwareGeneration(snapshot.HardwareGeneration)
	if err != nil {
		return err
	}

	d.Set("created_at", getTimestamp(snapshot.CreatedAt))
	d.Set("name", snapshot.Name)
	d.Set("folder_id", snapshot.FolderId)
	d.Set("description", snapshot.Description)
	d.Set("disk_size", toGigabytes(snapshot.DiskSize))
	d.Set("storage_size", toGigabytes(snapshot.StorageSize))
	d.Set("source_disk_id", snapshot.SourceDiskId)

	if err := d.Set("hardware_generation", hardwareGeneration); err != nil {
		return err
	}

	return d.Set("labels", snapshot.Labels)
}

func resourceYandexComputeSnapshotUpdate(d *schema.ResourceData, meta interface{}) error {
	d.Partial(true)

	labelPropName := "labels"
	if d.HasChange(labelPropName) {
		labelsProp, err := expandLabels(d.Get(labelPropName))
		if err != nil {
			return err
		}

		req := &compute.UpdateSnapshotRequest{
			SnapshotId: d.Id(),
			Labels:     labelsProp,
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{labelPropName},
			},
		}

		err = makeSnapshotUpdateRequest(req, d, meta)
		if err != nil {
			return err
		}

	}

	namePropName := "name"
	if d.HasChange(namePropName) {
		req := &compute.UpdateSnapshotRequest{
			SnapshotId: d.Id(),
			Name:       d.Get(namePropName).(string),
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{namePropName},
			},
		}

		err := makeSnapshotUpdateRequest(req, d, meta)
		if err != nil {
			return err
		}

	}

	descPropName := "description"
	if d.HasChange(descPropName) {
		req := &compute.UpdateSnapshotRequest{
			SnapshotId:  d.Id(),
			Description: d.Get(descPropName).(string),
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{descPropName},
			},
		}

		err := makeSnapshotUpdateRequest(req, d, meta)
		if err != nil {
			return err
		}

	}

	d.Partial(false)

	return resourceYandexComputeSnapshotRead(d, meta)
}

func resourceYandexComputeSnapshotDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Deleting Snapshot %q", d.Id())

	req := &compute.DeleteSnapshotRequest{
		SnapshotId: d.Id(),
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Compute().Snapshot().Delete(ctx, req))
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Snapshot %q", d.Get("name").(string)))
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished deleting Snapshot %q", d.Id())
	return nil
}

func makeSnapshotUpdateRequest(req *compute.UpdateSnapshotRequest, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Compute().Snapshot().Update(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to update Snapshot %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error updating Snapshot %q: %s", d.Id(), err)
	}

	return nil
}
