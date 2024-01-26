package yandex

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

func dataSourceYandexComputeSnapshotSchedule() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceYandexComputeSnapshotScheduleRead,
		Schema: map[string]*schema.Schema{
			"snapshot_schedule_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},

			"labels": {
				Type: schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
				Optional: true,
			},

			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"retention_period": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},

			"schedule_policy": {
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"expression": {
							Type:     schema.TypeString,
							Computed: true,
							Optional: true,
						},

						"start_at": {
							Type:     schema.TypeString,
							Computed: true,
							Optional: true,
						},
					},
				},
				Computed: true,
				Optional: true,
			},

			"snapshot_count": {
				Type:     schema.TypeInt,
				Computed: true,
				Optional: true,
			},

			"snapshot_spec": {
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"description": {
							Type:     schema.TypeString,
							Computed: true,
							Optional: true,
						},

						"labels": {
							Type: schema.TypeMap,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Set:      schema.HashString,
							Computed: true,
							Optional: true,
						},
					},
				},
				Computed: true,
				Optional: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"disk_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}

}

func dataSourceYandexComputeSnapshotScheduleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	err := checkOneOf(d, "snapshot_schedule_id", "name")
	if err != nil {
		return diag.FromErr(err)
	}

	scheduleID := d.Get("snapshot_schedule_id").(string)
	_, scheduleNameOk := d.GetOk("name")

	if scheduleNameOk {
		scheduleID, err = resolveObjectID(ctx, config, d, sdkresolvers.SnapshotScheduleResolver)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	schedule, err := config.sdk.Compute().SnapshotSchedule().Get(ctx, &compute.GetSnapshotScheduleRequest{
		SnapshotScheduleId: scheduleID,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	var diskIDs []string
	var token string
	for {
		resp, err := config.sdk.Compute().SnapshotSchedule().ListDisks(ctx, &compute.ListSnapshotScheduleDisksRequest{
			SnapshotScheduleId: scheduleID,
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

	d.Set("snapshot_schedule_id", schedule.Id)
	d.Set("folder_id", schedule.FolderId)
	d.Set("created_at", getTimestamp(schedule.CreatedAt))
	d.Set("name", schedule.Name)
	d.Set("description", schedule.Description)
	d.Set("status", strings.ToLower(schedule.Status.String()))

	d.Set("retention_period", schedule.GetRetentionPeriod().String())
	d.Set("snapshot_count", int(schedule.GetSnapshotCount()))

	d.Set("disk_ids", diskIDs)

	if err := d.Set("labels", schedule.Labels); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(schedule.Id)

	return nil
}
