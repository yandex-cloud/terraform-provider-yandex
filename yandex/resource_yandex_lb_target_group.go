package yandex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/loadbalancer/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

const yandexLBTargetGroupDefaultTimeout = 5 * time.Minute

func resourceYandexLBTargetGroup() *schema.Resource {
	return &schema.Resource{
		Description: "Creates a target group in the specified folder and adds the specified targets to it. For more information, see [the official documentation](https://yandex.cloud/docs/load-balancer/concepts/target-resources).",

		Create: resourceYandexLBTargetGroupCreate,
		Read:   resourceYandexLBTargetGroupRead,
		Update: resourceYandexLBTargetGroupUpdate,
		Delete: resourceYandexLBTargetGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexLBTargetGroupDefaultTimeout),
			Update: schema.DefaultTimeout(yandexLBTargetGroupDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexLBTargetGroupDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Optional:    true,
			},

			"region_id": {
				Type:        schema.TypeString,
				Description: "ID of the availability zone where the target group resides. If omitted, default region is being used.",
				Optional:    true,
				Computed:    true,
			},

			"target": {
				Type:        schema.TypeSet,
				Description: "A Target resource.",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subnet_id": {
							Type:        schema.TypeString,
							Description: "ID of the subnet that targets are connected to. All targets in the target group must be connected to the same subnet within a single availability zone.",
							Required:    true,
						},
						"address": {
							Type:        schema.TypeString,
							Description: "IP address of the target.",
							Required:    true,
						},
					},
				},
				Set: resourceLBTargetGroupTargetHash,
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
				ForceNew:    true,
			},

			"labels": {
				Type:        schema.TypeMap,
				Description: common.ResourceDescriptions["labels"],
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},

			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},
		},
	}

}

func resourceYandexLBTargetGroupCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return fmt.Errorf("Error expanding labels while creating target group: %s", err)
	}

	folderID, err := getFolderID(d, config)
	if err != nil {
		return fmt.Errorf("Error getting folder ID while creating target group: %s", err)
	}

	targets, err := expandLBTargets(d)
	if err != nil {
		return fmt.Errorf("Error expanding targets while creating target group: %s", err)
	}

	req := loadbalancer.CreateTargetGroupRequest{
		FolderId:    folderID,
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		RegionId:    d.Get("region_id").(string),
		Labels:      labels,
		Targets:     targets,
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.LoadBalancer().TargetGroup().Create(ctx, &req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create target group: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("Error while get target group create operation metadata: %s", err)
	}

	md, ok := protoMetadata.(*loadbalancer.CreateTargetGroupMetadata)
	if !ok {
		return fmt.Errorf("could not get TargetGroup ID from create operation metadata")
	}

	d.SetId(md.TargetGroupId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while waiting operation to create target group: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("TargetGroup creation failed: %s", err)
	}

	return resourceYandexLBTargetGroupRead(d, meta)
}

func resourceYandexLBTargetGroupRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutRead))
	defer cancel()

	tg, err := config.sdk.LoadBalancer().TargetGroup().Get(ctx, &loadbalancer.GetTargetGroupRequest{
		TargetGroupId: d.Id(),
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("TargetGroup %q", d.Get("name").(string)))
	}

	targets, err := flattenLBTargets(tg)
	if err != nil {
		return err
	}

	d.Set("created_at", getTimestamp(tg.CreatedAt))
	d.Set("name", tg.Name)
	d.Set("folder_id", tg.FolderId)
	d.Set("region_id", tg.RegionId)
	d.Set("description", tg.Description)

	if err := d.Set("target", targets); err != nil {
		return err
	}

	return d.Set("labels", tg.Labels)
}

func resourceYandexLBTargetGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return err
	}

	targets, err := expandLBTargets(d)
	if err != nil {
		return fmt.Errorf("Error expanding targets while creating target group: %s", err)
	}

	req := &loadbalancer.UpdateTargetGroupRequest{
		TargetGroupId: d.Id(),
		Name:          d.Get("name").(string),
		Description:   d.Get("description").(string),
		Labels:        labels,
		Targets:       targets,
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.LoadBalancer().TargetGroup().Update(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to update TargetGroup %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error updating TargetGroup %q: %s", d.Id(), err)
	}

	return resourceYandexLBTargetGroupRead(d, meta)
}

func resourceYandexLBTargetGroupDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Deleting TargetGroup %q", d.Id())

	req := &loadbalancer.DeleteTargetGroupRequest{
		TargetGroupId: d.Id(),
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.LoadBalancer().TargetGroup().Delete(ctx, req))
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("TargetGroup %q", d.Get("name").(string)))
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished deleting TargetGroup %q", d.Id())
	return nil
}
