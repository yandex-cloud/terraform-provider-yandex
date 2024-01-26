package yandex

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/protobuf/types/known/durationpb"
	fieldmaskpb "google.golang.org/protobuf/types/known/fieldmaskpb"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
)

const yandexComputeSnapshotScheduleDefaultTimeout = 5 * time.Minute

func resourceYandexComputeSnapshotSchedule() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceYandexComputeSnapshotScheduleCreate,
		ReadContext:   resourceYandexComputeSnapshotScheduleRead,
		UpdateContext: resourceYandexComputeSnapshotScheduleUpdate,
		DeleteContext: resourceYandexComputeSnapshotScheduleDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexComputeSnapshotScheduleDefaultTimeout),
			Update: schema.DefaultTimeout(yandexComputeSnapshotScheduleDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexComputeSnapshotScheduleDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"disk_ids": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
				Computed: true,
			},

			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},

			"labels": {
				Type: schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Set:      schema.HashString,
				Optional: true,
			},

			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"retention_period": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"schedule_policy": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"expression": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"start_at": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
				Optional: true,
				Computed: true,
			},

			"snapshot_count": {
				Type:     schema.TypeInt,
				Optional: true,
			},

			"snapshot_spec": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"labels": {
							Type: schema.TypeMap,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Set:      schema.HashString,
							Optional: true,
						},
					},
				},
				Optional: true,
				Computed: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}

}

func resourceYandexComputeSnapshotScheduleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	folderID, err := getFolderID(d, config)
	if err != nil {
		return diag.FromErr(err)
	}

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return diag.FromErr(err)
	}

	schedulePolicy, err := expandSnapshotScheduleSchedulePolicy(d)
	if err != nil {
		return diag.FromErr(err)
	}

	snapshotSpec, err := expandSnapshotScheduleSnapshotSpec(d)
	if err != nil {
		return diag.FromErr(err)
	}

	diskIDs := convertStringSet(d.Get("disk_ids").(*schema.Set))

	req := &compute.CreateSnapshotScheduleRequest{
		FolderId:       folderID,
		Name:           d.Get("name").(string),
		Description:    d.Get("description").(string),
		Labels:         labels,
		SchedulePolicy: schedulePolicy,
		SnapshotSpec:   snapshotSpec,
		DiskIds:        diskIDs,
	}

	if v, ok := d.GetOk("retention_period"); ok {
		retentionPeriod, err := time.ParseDuration(v.(string))
		if err != nil {
			return diag.FromErr(err)
		}

		req.SetRetentionPeriod(durationpb.New(retentionPeriod))
	}

	if v, ok := d.GetOk("snapshot_count"); ok {
		req.SetSnapshotCount(int64(v.(int)))
	}

	op, err := config.sdk.WrapOperation(config.sdk.Compute().SnapshotSchedule().Create(ctx, req))
	if err != nil {
		return diag.FromErr(err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return diag.FromErr(err)
	}

	md, ok := protoMetadata.(*compute.CreateSnapshotScheduleMetadata)
	if !ok {
		return diag.FromErr(fmt.Errorf("could not get Snapshot Schedule ID from create operation metadata"))
	}

	d.SetId(md.SnapshotScheduleId)

	err = op.Wait(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	if _, err := op.Response(); err != nil {
		return diag.FromErr(err)
	}

	return resourceYandexComputeSnapshotScheduleRead(ctx, d, meta)
}

func resourceYandexComputeSnapshotScheduleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	schedule, err := config.sdk.Compute().SnapshotSchedule().Get(config.Context(), &compute.GetSnapshotScheduleRequest{
		SnapshotScheduleId: d.Id(),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	var diskIDs []string
	var token string
	for {
		resp, err := config.sdk.Compute().SnapshotSchedule().ListDisks(config.Context(), &compute.ListSnapshotScheduleDisksRequest{
			SnapshotScheduleId: d.Id(),
			PageToken:          token,
		})
		if err != nil {
			return diag.FromErr(err)
		}
		for _, d := range resp.Disks {
			diskIDs = append(diskIDs, d.Id)
		}

		token = resp.NextPageToken
		if token == "" {
			break
		}
	}

	policy, err := flattenSnapshotScheduleSchedulePolicy(schedule.GetSchedulePolicy())
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("schedule_policy", policy); err != nil {
		return diag.FromErr(err)
	}

	snapshotSpec, err := flattenSnapshotScheduleSnapshotSpec(schedule.GetSnapshotSpec())
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("snapshot_spec", snapshotSpec); err != nil {
		return diag.FromErr(err)
	}

	d.Set("folder_id", schedule.FolderId)
	d.Set("created_at", getTimestamp(schedule.CreatedAt))
	d.Set("name", schedule.Name)
	d.Set("description", schedule.Description)
	d.Set("status", strings.ToLower(schedule.Status.String()))

	d.Set("retention_period", formatDuration(schedule.GetRetentionPeriod()))
	d.Set("snapshot_count", int(schedule.GetSnapshotCount()))

	d.Set("disk_ids", diskIDs)

	if err := d.Set("labels", schedule.Labels); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceYandexComputeSnapshotScheduleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var resourceYcpComputeSnapshotScheduleUpdateFieldsMap = map[string]string{
		"name":                         "name",
		"description":                  "description",
		"labels":                       "labels",
		"schedule_policy.0.start_at":   "schedule_policy.start_at",
		"schedule_policy.0.expression": "schedule_policy.expression",
		"retention_period":             "retention_period",
		"snapshot_count":               "snapshot_count",
		"snapshot_spec.0.description":  "snapshot_spec.description",
		"snapshot_spec.0.labels":       "snapshot_spec.labels",
	}

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return diag.FromErr(err)
	}

	schedulePolicy, err := expandSnapshotScheduleSchedulePolicy(d)
	if err != nil {
		return diag.FromErr(err)
	}

	snapshotSpec, err := expandSnapshotScheduleSnapshotSpec(d)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &compute.UpdateSnapshotScheduleRequest{
		SnapshotScheduleId: d.Id(),
		Name:               d.Get("name").(string),
		Description:        d.Get("description").(string),
		Labels:             labels,
		SchedulePolicy:     schedulePolicy,
		SnapshotSpec:       snapshotSpec,
	}

	if v, ok := d.GetOk("retention_period"); ok {
		updateSnapshotScheduleRequestRetentionPeriod, err := time.ParseDuration(v.(string))
		if err != nil {
			return diag.FromErr(err)
		}

		req.SetRetentionPeriod(durationpb.New(updateSnapshotScheduleRequestRetentionPeriod))
	}

	if v, ok := d.GetOk("snapshot_count"); ok {
		req.SetSnapshotCount(int64(v.(int)))
	}

	updatePath := generateFieldMasks(d, resourceYcpComputeSnapshotScheduleUpdateFieldsMap)
	req.UpdateMask = &fieldmaskpb.FieldMask{Paths: updatePath}

	if err := makeSnapshotScheduleUpdateRequest(ctx, req, d, meta); err != nil {
		return diag.FromErr(err)
	}

	return resourceYandexComputeSnapshotScheduleRead(ctx, d, meta)
}

func resourceYandexComputeSnapshotScheduleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	log.Printf("[DEBUG] Deleting SnapshotSchedule %q", d.Id())

	req := &compute.DeleteSnapshotScheduleRequest{
		SnapshotScheduleId: d.Id(),
	}

	op, err := config.sdk.WrapOperation(config.sdk.Compute().SnapshotSchedule().Delete(ctx, req))
	if err != nil {
		return diag.FromErr(err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = op.Response()
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Finished deleting SnapshotSchedule %q", d.Id())
	return nil
}

func makeSnapshotScheduleUpdateRequest(ctx context.Context, req *compute.UpdateSnapshotScheduleRequest, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Updating SnapshotSchedule %q", d.Id())

	op, err := config.sdk.WrapOperation(config.sdk.Compute().SnapshotSchedule().Update(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to update SnapshotSchedule %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error updating SnapshotSchedule %q: %s", d.Id(), err)
	}

	if err := updateSnapshotScheduleDisks(ctx, d, meta); err != nil {
		return fmt.Errorf("Error updating SnapshotScheduleDisks %q: %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Finished updating SnapshotSchedule %q", d.Id())
	return nil
}

func updateSnapshotScheduleDisks(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	oldDisks := make(map[string]bool)
	var token string
	for {
		resp, err := config.sdk.Compute().SnapshotSchedule().ListDisks(ctx, &compute.ListSnapshotScheduleDisksRequest{
			SnapshotScheduleId: d.Id(),
			PageToken:          token,
		})
		if err != nil {
			return fmt.Errorf("Failed to get snapshot schedule disks: %v", err)
		}
		for _, d := range resp.Disks {
			oldDisks[d.Id] = true
		}

		token = resp.NextPageToken
		if token == "" {
			break
		}
	}

	newDisks := make(map[string]bool)
	for _, d := range convertStringSet(d.Get("disk_ids").(*schema.Set)) {
		newDisks[d] = true
	}

	req := makeUpdateSnapshotScheduleDisksRequest(oldDisks, newDisks)
	req.SnapshotScheduleId = d.Id()

	if len(req.Add) == 0 && len(req.Remove) == 0 {
		return nil
	}

	log.Printf("[DEBUG] Updating SnapshotSchedule disks %q", d.Id())

	op, err := config.sdk.WrapOperation(config.sdk.Compute().SnapshotSchedule().UpdateDisks(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to update SnapshotSchedule disks %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error updating SnapshotSchedule disks %q: %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Finished updating SnapshotSchedule disks %q", d.Id())
	return nil
}

func makeUpdateSnapshotScheduleDisksRequest(oldDisks, newDisks map[string]bool) *compute.UpdateSnapshotScheduleDisksRequest {
	req := &compute.UpdateSnapshotScheduleDisksRequest{}

	// remove old disks
	for disk := range oldDisks {
		if newDisks[disk] {
			continue
		}
		req.Remove = append(req.Remove, disk)
	}

	// add new disks
	for disk := range newDisks {
		if oldDisks[disk] {
			continue
		}
		req.Add = append(req.Add, disk)
	}

	return req
}
