package yandex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/resourcemanager/v1"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Delete can last up to 30 minutes, approved by IAM.
const yandexResourceManagerCloudDeleteTimeout = 30 * time.Minute

func resourceYandexResourceManagerCloud() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexResourceManagerCloudCreate,
		Read:   resourceYandexResourceManagerCloudRead,
		Update: resourceYandexResourceManagerCloudUpdate,
		Delete: resourceYandexResourceManagerCloudDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexResourceManagerCloudDefaultTimeout),
			Read:   schema.DefaultTimeout(yandexResourceManagerCloudDefaultTimeout),
			Update: schema.DefaultTimeout(yandexResourceManagerCloudDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexResourceManagerCloudDeleteTimeout),
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

			"organization_id": {
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

func resourceYandexResourceManagerCloudCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	organizationID, err := getOrganizationID(d, config)
	if err != nil {
		return fmt.Errorf("Error getting cloud ID while creating Cloud: %s", err)
	}

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return fmt.Errorf("Error expanding labels while creating Cloud: %s", err)
	}

	req := resourcemanager.CreateCloudRequest{
		OrganizationId: organizationID,
		Name:           d.Get("name").(string),
		Labels:         labels,
		Description:    d.Get("description").(string),
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.ResourceManager().Cloud().Create(ctx, &req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create Cloud: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("Error while get Cloud create operation metadata: %s", err)
	}

	md, ok := protoMetadata.(*resourcemanager.CreateCloudMetadata)
	if !ok {
		return fmt.Errorf("could not get Cloud ID from create operation metadata")
	}

	d.SetId(md.CloudId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while waiting operation to create Cloud: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("Cloud creation failed: %s", err)
	}

	return resourceYandexResourceManagerCloudRead(d, meta)
}

func resourceYandexResourceManagerCloudRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	cloud, err := config.sdk.ResourceManager().Cloud().Get(context.Background(),
		&resourcemanager.GetCloudRequest{
			CloudId: d.Id(),
		})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Cloud %q", d.Id()))
	}

	d.Set("created_at", getTimestamp(cloud.CreatedAt))
	d.Set("name", cloud.Name)
	d.Set("organization_id", cloud.OrganizationId)
	d.Set("description", cloud.Description)

	return d.Set("labels", cloud.Labels)
}

func resourceYandexResourceManagerCloudUpdate(d *schema.ResourceData, meta interface{}) error {
	req := &resourcemanager.UpdateCloudRequest{
		CloudId:    d.Id(),
		UpdateMask: &field_mask.FieldMask{},
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
		return fmt.Errorf("No fields were updated for Cloud %s", d.Id())
	}

	err := makeCloudUpdateRequest(req, d, meta)
	if err != nil {
		return err
	}

	return resourceYandexResourceManagerCloudRead(d, meta)
}

func resourceYandexResourceManagerCloudDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Deleting Cloud %q", d.Id())

	req := &resourcemanager.DeleteCloudRequest{
		CloudId:     d.Id(),
		DeleteAfter: timestamppb.Now(),
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.ResourceManager().Cloud().Delete(ctx, req))
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Cloud %q", d.Id()))
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished deleting Cloud %q", d.Id())
	return nil
}

func makeCloudUpdateRequest(req *resourcemanager.UpdateCloudRequest, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.ResourceManager().Cloud().Update(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to update Cloud %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error updating Cloud %q: %s", d.Id(), err)
	}

	return nil
}
