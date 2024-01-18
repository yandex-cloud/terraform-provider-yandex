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

const yandexComputePlacementGroupDefaultTimeout = 1 * time.Minute

func resourceYandexComputePlacementGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexComputePlacementGroupCreate,
		Read:   resourceYandexComputePlacementGroupRead,
		Update: resourceYandexComputePlacementGroupUpdate,
		Delete: resourceYandexComputePlacementGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexComputePlacementGroupDefaultTimeout),
			Update: schema.DefaultTimeout(yandexComputePlacementGroupDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexComputePlacementGroupDefaultTimeout),
		},

		SchemaVersion: 1,

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

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"placement_strategy_spread": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"placement_strategy_partitions": {
				Type:     schema.TypeInt,
				Optional: true,
			},
		},
	}
}

func resourceYandexComputePlacementGroupCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	folderID, err := getFolderID(d, config)
	if err != nil {
		return fmt.Errorf("Error getting folder ID while creating Placement Group: %s", err)
	}

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return fmt.Errorf("Error expanding labels while creating Placement Group: %s", err)
	}

	req := compute.CreatePlacementGroupRequest{
		FolderId:    folderID,
		Name:        d.Get("name").(string),
		Labels:      labels,
		Description: d.Get("description").(string),
		PlacementStrategy: &compute.CreatePlacementGroupRequest_SpreadPlacementStrategy{
			SpreadPlacementStrategy: &compute.SpreadPlacementStrategy{},
		},
	}

	spreadStrategy, spreadStrategySet := d.GetOk("placement_strategy_spread")
	partitions, partitionStrategySet := d.GetOk("placement_strategy_partitions")
	if partitionStrategySet && spreadStrategySet {
		return fmt.Errorf("Maximum one of 'placement_strategy_spread' or 'placement_strategy_partitions' should be set")
	}
	if partitionStrategySet {
		partitionCount := partitions.(int)
		req.PlacementStrategy = &compute.CreatePlacementGroupRequest_PartitionPlacementStrategy{
			PartitionPlacementStrategy: &compute.PartitionPlacementStrategy{
				Partitions: int64(partitionCount),
			},
		}
	}

	if spreadStrategySet {
		spreadStrategyVal := spreadStrategy.(bool)
		if !spreadStrategyVal {
			return fmt.Errorf("Invalid value for `placement_strategy_spread` should be true or unset")
		}
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Compute().PlacementGroup().Create(ctx, &req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create Placement Group: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("Error while get Placement Group create operation metadata: %s", err)
	}

	md, ok := protoMetadata.(*compute.CreatePlacementGroupMetadata)
	if !ok {
		return fmt.Errorf("could not get Placement Group ID from create operation metadata")
	}

	d.SetId(md.GetPlacementGroupId())

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while waiting operation to create Placement Group: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("Placement Group creation failed: %s", err)
	}

	return resourceYandexComputePlacementGroupRead(d, meta)
}

func resourceYandexComputePlacementGroupRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	placementGroup, err := config.sdk.Compute().PlacementGroup().Get(context.Background(),
		&compute.GetPlacementGroupRequest{
			PlacementGroupId: d.Id(),
		})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Placement Group %q", d.Id()))
	}

	d.Set("created_at", getTimestamp(placementGroup.CreatedAt))
	d.Set("name", placementGroup.Name)
	d.Set("folder_id", placementGroup.FolderId)
	d.Set("description", placementGroup.Description)

	return d.Set("labels", placementGroup.Labels)
}

func resourceYandexComputePlacementGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	req := &compute.UpdatePlacementGroupRequest{
		PlacementGroupId: d.Id(),
		UpdateMask:       &field_mask.FieldMask{},
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
		return fmt.Errorf("No fields were updated for Placement Group %s", d.Id())
	}

	err := makePlacementGroupUpdateRequest(req, d, meta)
	if err != nil {
		return err
	}

	return resourceYandexComputePlacementGroupRead(d, meta)
}

func resourceYandexComputePlacementGroupDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Deleting Placement Group %q", d.Id())

	req := &compute.DeletePlacementGroupRequest{
		PlacementGroupId: d.Id(),
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Compute().PlacementGroup().Delete(ctx, req))
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Placement Group %q", d.Id()))
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished deleting Placement Group %q", d.Id())
	return nil
}

func makePlacementGroupUpdateRequest(req *compute.UpdatePlacementGroupRequest, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Compute().PlacementGroup().Update(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to update Placement Group %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error updating Placement Group %q: %s", d.Id(), err)
	}

	return nil
}
