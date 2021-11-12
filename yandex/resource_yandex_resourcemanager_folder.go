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
const yandexResourceManagerFolderDeleteTimeout = 30 * time.Minute

func resourceYandexResourceManagerFolder() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexResourceManagerFolderCreate,
		Read:   resourceYandexResourceManagerFolderRead,
		Update: resourceYandexResourceManagerFolderUpdate,
		Delete: resourceYandexResourceManagerFolderDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexResourceManagerFolderDefaultTimeout),
			Read:   schema.DefaultTimeout(yandexResourceManagerFolderDefaultTimeout),
			Update: schema.DefaultTimeout(yandexResourceManagerFolderDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexResourceManagerFolderDeleteTimeout),
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

			"cloud_id": {
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

func resourceYandexResourceManagerFolderCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	cloudID, err := getCloudID(d, config)
	if err != nil {
		return fmt.Errorf("Error getting cloud ID while creating Folder: %s", err)
	}

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return fmt.Errorf("Error expanding labels while creating Folder: %s", err)
	}

	req := resourcemanager.CreateFolderRequest{
		CloudId:     cloudID,
		Name:        d.Get("name").(string),
		Labels:      labels,
		Description: d.Get("description").(string),
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.ResourceManager().Folder().Create(ctx, &req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create Folder: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("Error while get Folder create operation metadata: %s", err)
	}

	md, ok := protoMetadata.(*resourcemanager.CreateFolderMetadata)
	if !ok {
		return fmt.Errorf("could not get Folder ID from create operation metadata")
	}

	d.SetId(md.FolderId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while waiting operation to create Folder: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("Folder creation failed: %s", err)
	}

	return resourceYandexResourceManagerFolderRead(d, meta)
}

func resourceYandexResourceManagerFolderRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	folder, err := config.sdk.ResourceManager().Folder().Get(context.Background(),
		&resourcemanager.GetFolderRequest{
			FolderId: d.Id(),
		})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Folder %q", d.Id()))
	}

	d.Set("created_at", getTimestamp(folder.CreatedAt))
	d.Set("name", folder.Name)
	d.Set("cloud_id", folder.CloudId)
	d.Set("description", folder.Description)

	return d.Set("labels", folder.Labels)
}

func resourceYandexResourceManagerFolderUpdate(d *schema.ResourceData, meta interface{}) error {
	req := &resourcemanager.UpdateFolderRequest{
		FolderId:   d.Id(),
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
		return fmt.Errorf("No fields were updated for Folder %s", d.Id())
	}

	err := makeFolderUpdateRequest(req, d, meta)
	if err != nil {
		return err
	}

	return resourceYandexResourceManagerFolderRead(d, meta)
}

func resourceYandexResourceManagerFolderDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Deleting Folder %q", d.Id())

	req := &resourcemanager.DeleteFolderRequest{
		FolderId:    d.Id(),
		DeleteAfter: timestamppb.Now(),
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.ResourceManager().Folder().Delete(ctx, req))
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Folder %q", d.Id()))
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished deleting Folder %q", d.Id())
	return nil
}

func makeFolderUpdateRequest(req *resourcemanager.UpdateFolderRequest, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.ResourceManager().Folder().Update(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to update Folder %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error updating Folder %q: %s", d.Id(), err)
	}

	return nil
}
