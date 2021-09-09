package yandex

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/apploadbalancer/v1"
	"google.golang.org/genproto/protobuf/field_mask"
	"log"
	"time"
)

const yandexALBHTTPRouterDefaultTimeout = 5 * time.Minute

func resourceYandexALBHTTPRouter() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexALBHTTPRouterCreate,
		Read:   resourceYandexALBHTTPRouterRead,
		Update: resourceYandexALBHTTPRouterUpdate,
		Delete: resourceYandexALBHTTPRouterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexALBHTTPRouterDefaultTimeout),
			Update: schema.DefaultTimeout(yandexALBHTTPRouterDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexALBHTTPRouterDefaultTimeout),
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

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}

}

func resourceYandexALBHTTPRouterCreate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Creating Application Http Router %q", d.Id())

	config := meta.(*Config)

	labels, err := expandLabels(d.Get("labels"))

	if err != nil {
		return fmt.Errorf("Error expanding labels while creating Application : %w", err)
	}

	folderID, err := getFolderID(d, config)
	if err != nil {
		return fmt.Errorf("Error getting folder ID while creating Application Http Router: %w", err)
	}

	req := apploadbalancer.CreateHttpRouterRequest{
		FolderId:    folderID,
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Labels:      labels,
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.ApplicationLoadBalancer().HttpRouter().Create(ctx, &req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create Application Http Router: %w", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("Error while get Application Http Router create operation metadata: %w", err)
	}

	md, ok := protoMetadata.(*apploadbalancer.CreateHttpRouterMetadata)
	if !ok {
		return fmt.Errorf("could not get Application Http Router ID from create operation metadata")
	}

	d.SetId(md.HttpRouterId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while waiting operation to create Application Http Router: %w", err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("Application Http Router creation failed: %w", err)
	}

	log.Printf("[DEBUG] Finished creating Application Http Router %q", d.Id())
	return resourceYandexALBHTTPRouterRead(d, meta)

}

func resourceYandexALBHTTPRouterRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Reading Application Http Router %q", d.Id())
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutRead))
	defer cancel()

	bg, err := config.sdk.ApplicationLoadBalancer().HttpRouter().Get(ctx, &apploadbalancer.GetHttpRouterRequest{
		HttpRouterId: d.Id(),
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Application Http Router %q", d.Get("name").(string)))
	}

	d.Set("created_at", getTimestamp(bg.CreatedAt))
	d.Set("name", bg.Name)
	d.Set("folder_id", bg.FolderId)
	d.Set("description", bg.Description)

	log.Printf("[DEBUG] Finished reading Application Http Router %q", d.Id())
	return d.Set("labels", bg.Labels)
}

func resourceYandexALBHTTPRouterUpdate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Updating Application Http Router %q", d.Id())
	config := meta.(*Config)

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return err
	}

	req := &apploadbalancer.UpdateHttpRouterRequest{
		HttpRouterId: d.Id(),
		Name:         d.Get("name").(string),
		Description:  d.Get("description").(string),
		Labels:       labels,
	}

	var updatePath []string
	for field, path := range resourceALBHTTPRouterUpdateFieldsMap {
		if d.HasChange(field) {
			updatePath = append(updatePath, path)
		}
	}

	req.UpdateMask = &field_mask.FieldMask{Paths: updatePath}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.ApplicationLoadBalancer().HttpRouter().Update(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to update Application Http Router %q: %w", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error updating Application Http Router %q: %w", d.Id(), err)
	}

	log.Printf("[DEBUG] Finished updating Application Http Router %q", d.Id())

	return resourceYandexALBHTTPRouterRead(d, meta)
}

func resourceYandexALBHTTPRouterDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Deleting Application Http Router %q", d.Id())
	config := meta.(*Config)

	req := &apploadbalancer.DeleteHttpRouterRequest{
		HttpRouterId: d.Id(),
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.ApplicationLoadBalancer().HttpRouter().Delete(ctx, req))
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Application Http Router %q", d.Get("name").(string)))
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished deleting Application Http Router %q", d.Id())
	return nil
}

var resourceALBHTTPRouterUpdateFieldsMap = map[string]string{
	"name":        "name",
	"description": "description",
	"labels":      "labels",
}
