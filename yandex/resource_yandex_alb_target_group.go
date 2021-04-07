package yandex

import (
	"context"
	"fmt"
	//"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

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

			"targets": {
				Type:     schema.TypeSet,
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
				Set: resourceLBTargetGroupTargetHash,
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

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return fmt.Errorf("Error expanding labels while creating target group: %s", err)
	}

	folderID, err := getFolderID(d, config)
	if err != nil {
		return fmt.Errorf("Error getting folder ID while creating target group: %s", err)
	}

	targets, err := expandALBTargets(d)
	if err != nil {
		return fmt.Errorf("Error expanding targets while creating target group: %s", err)
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
		return fmt.Errorf("Error while requesting API to create target group: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("Error while get target group create operation metadata: %s", err)
	}

	md, ok := protoMetadata.(*apploadbalancer.CreateTargetGroupMetadata)
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

	return resourceYandexALBTargetGroupRead(d, meta)
}

func resourceYandexALBTargetGroupRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutRead))
	defer cancel()

	tg, err := config.sdk.ApplicationLoadBalancer().TargetGroup().Get(ctx, &apploadbalancer.GetTargetGroupRequest{
		TargetGroupId: d.Id(),
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("TargetGroup %q", d.Get("name").(string)))
	}

	targets, err := flattenALBTargets(tg)
	if err != nil {
		return err
	}

	createdAt, err := getTimestamp(tg.CreatedAt)
	if err != nil {
		return err
	}

	d.Set("created_at", createdAt)
	d.Set("name", tg.Name)
	d.Set("folder_id", tg.FolderId)
	d.Set("description", tg.Description)

	if err := d.Set("target", targets); err != nil {
		return err
	}

	return d.Set("labels", tg.Labels)
}

func resourceYandexALBTargetGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return err
	}

	targets, err := expandALBTargets(d)
	if err != nil {
		return fmt.Errorf("Error expanding targets while creating target group: %s", err)
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
		return fmt.Errorf("Error while requesting API to update TargetGroup %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error updating TargetGroup %q: %s", d.Id(), err)
	}

	return resourceYandexALBTargetGroupRead(d, meta)
}

func resourceYandexALBTargetGroupDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	// log.Printf("[DEBUG] Deleting TargetGroup %q", d.Id())

	req := &apploadbalancer.DeleteTargetGroupRequest{
		TargetGroupId: d.Id(),
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.ApplicationLoadBalancer().TargetGroup().Delete(ctx, req))
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

	// log.Printf("[DEBUG] Finished deleting TargetGroup %q", d.Id())
	return nil
}
