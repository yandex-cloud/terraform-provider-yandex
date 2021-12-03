package yandex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/apploadbalancer/v1"
)

const yandexALBTargetGroupDefaultTimeout = 5 * time.Minute

func resourceYandexALBTargetGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexALBTargetGroupCreate,
		Read:   resourceYandexALBTargetGroupRead,
		Update: resourceYandexALBTargetGroupUpdate,
		Delete: resourceYandexALBTargetGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexALBTargetGroupDefaultTimeout),
			Update: schema.DefaultTimeout(yandexALBTargetGroupDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexALBTargetGroupDefaultTimeout),
		},

		SchemaVersion: 0,

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

			"target": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subnet_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"ip_address": {
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

func resourceYandexALBTargetGroupCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Creating Application Target Group %q", d.Id())
	labels, err := expandLabels(d.Get("labels"))

	if err != nil {
		return fmt.Errorf("Error expanding labels while creating Application Target Group: %w", err)
	}

	folderID, err := getFolderID(d, config)
	if err != nil {
		return fmt.Errorf("Error getting folder ID while creating Application Target Group: %w", err)
	}

	targets, err := expandALBTargets(d)
	if err != nil {
		return fmt.Errorf("Error expanding targets while creating Application Target Group: %w", err)
	}

	req := apploadbalancer.CreateTargetGroupRequest{
		FolderId:    folderID,
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Labels:      labels,
		Targets:     targets,
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.ApplicationLoadBalancer().TargetGroup().Create(ctx, &req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create Application Target Group: %w", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("Error while get Application Target Group create operation metadata: %w", err)
	}

	md, ok := protoMetadata.(*apploadbalancer.CreateTargetGroupMetadata)
	if !ok {
		return fmt.Errorf("could not get Application Target Group ID from create operation metadata")
	}

	d.SetId(md.TargetGroupId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while waiting operation to create Application Target Group: %w", err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("Application Target Group creation failed: %w", err)
	}

	log.Printf("[DEBUG] Finished creating Application Target Group %q", d.Id())
	return resourceYandexALBTargetGroupRead(d, meta)
}

func resourceYandexALBTargetGroupRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Reading Application Target Group %q", d.Id())

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutRead))
	defer cancel()

	tg, err := config.sdk.ApplicationLoadBalancer().TargetGroup().Get(ctx, &apploadbalancer.GetTargetGroupRequest{
		TargetGroupId: d.Id(),
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Application Target Group %q", d.Get("name").(string)))
	}

	targets := flattenALBTargets(tg)

	_ = d.Set("created_at", getTimestamp(tg.CreatedAt))
	_ = d.Set("name", tg.Name)
	_ = d.Set("folder_id", tg.FolderId)
	_ = d.Set("description", tg.Description)

	if err := d.Set("target", targets); err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished reading Application Target Group %q", d.Id())
	return d.Set("labels", tg.Labels)
}

func resourceYandexALBTargetGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Updating Application Target Group %q", d.Id())

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return err
	}

	targets, err := expandALBTargets(d)
	if err != nil {
		return fmt.Errorf("Error expanding targets while updating Application Target Group: %w", err)
	}

	req := &apploadbalancer.UpdateTargetGroupRequest{
		TargetGroupId: d.Id(),
		Name:          d.Get("name").(string),
		Description:   d.Get("description").(string),
		Labels:        labels,
		Targets:       targets,
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.ApplicationLoadBalancer().TargetGroup().Update(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to update Application Target Group %q: %w", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error updating Application Target Group %q: %w", d.Id(), err)
	}

	log.Printf("[DEBUG] Finished updating Application Target Group %q", d.Id())
	return resourceYandexALBTargetGroupRead(d, meta)
}

func resourceYandexALBTargetGroupDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Deleting Application Target Group %q", d.Id())

	req := &apploadbalancer.DeleteTargetGroupRequest{
		TargetGroupId: d.Id(),
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.ApplicationLoadBalancer().TargetGroup().Delete(ctx, req))
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Application Target Group %q", d.Get("name").(string)))
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished deleting Application Target Group %q", d.Id())
	return nil
}
