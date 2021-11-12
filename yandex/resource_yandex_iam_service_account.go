package yandex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
)

const yandexIAMServiceAccountDefaultTimeout = 1 * time.Minute

func resourceYandexIAMServiceAccount() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexIAMServiceAccountCreate,
		Read:   resourceYandexIAMServiceAccountRead,
		Update: resourceYandexIAMServiceAccountUpdate,
		Delete: resourceYandexIAMServiceAccountDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexIAMServiceAccountDefaultTimeout),
			Update: schema.DefaultTimeout(yandexIAMServiceAccountDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexIAMServiceAccountDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
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

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceYandexIAMServiceAccountCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	folderID, err := getFolderID(d, config)
	if err != nil {
		return fmt.Errorf("Error getting folder ID while creating service account: %s", err)
	}

	// Lock cloud to prevent out IAM changes:
	// SA create operation adds 'resource-manager.clouds.member' role
	unlock, err := lockCloudByFolderID(config, folderID)
	if err != nil {
		return fmt.Errorf("could not lock cloud to prevent IAM changes: %s", err)
	}
	defer unlock()

	req := iam.CreateServiceAccountRequest{
		FolderId:    folderID,
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.IAM().ServiceAccount().Create(ctx, &req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create service account: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("Error while get service account create operation metadata: %s", err)
	}

	md, ok := protoMetadata.(*iam.CreateServiceAccountMetadata)
	if !ok {
		return fmt.Errorf("could not get Service Account ID from create operation metadata")
	}

	d.SetId(md.ServiceAccountId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while waiting operation to create service account: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("Service account creation failed: %s", err)
	}

	return resourceYandexIAMServiceAccountRead(d, meta)
}

func resourceYandexIAMServiceAccountRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	sa, err := config.sdk.IAM().ServiceAccount().Get(config.Context(), &iam.GetServiceAccountRequest{
		ServiceAccountId: d.Id(),
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Service Account %q", d.Get("name").(string)))
	}

	d.Set("created_at", getTimestamp(sa.CreatedAt))
	d.Set("name", sa.Name)
	d.Set("folder_id", sa.FolderId)
	d.Set("description", sa.Description)

	return nil
}

func resourceYandexIAMServiceAccountUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	d.Partial(true)

	req := &iam.UpdateServiceAccountRequest{
		ServiceAccountId: d.Id(),
		UpdateMask:       &field_mask.FieldMask{},
	}

	if d.HasChange("name") {
		req.Name = d.Get("name").(string)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "name")
	}

	if d.HasChange("description") {
		req.Description = d.Get("description").(string)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "description")
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.IAM().ServiceAccount().Update(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to update Service Account %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error updating Service Account %q: %s", d.Id(), err)
	}

	d.Partial(false)

	return resourceYandexIAMServiceAccountRead(d, meta)
}

func resourceYandexIAMServiceAccountDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	folderID, err := getFolderID(d, config)
	if err != nil {
		return fmt.Errorf("Error getting folder ID while deleting service account: %s", err)
	}

	// Lock cloud to prevent out IAM changes:
	// SA delete operation removes 'resource-manager.clouds.member' role
	unlock, err := lockCloudByFolderID(config, folderID)
	if err != nil {
		return fmt.Errorf("could not lock cloud to prevent IAM changes: %s", err)
	}
	defer unlock()

	log.Printf("[DEBUG] Deleting Service Account %q", d.Id())

	req := &iam.DeleteServiceAccountRequest{
		ServiceAccountId: d.Id(),
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.IAM().ServiceAccount().Delete(ctx, req))
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Service Account %q", d.Get("name").(string)))
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	resp, err := op.Response()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished deleting Service Account %q: %#v", d.Id(), resp)
	return nil
}
