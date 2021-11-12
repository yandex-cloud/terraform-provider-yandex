package yandex

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"
)

const yandexContainerRegistryDefaultTimeout = 15 * time.Minute

func resourceYandexContainerRegistry() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexContainerRegistryCreate,
		Read:   resourceYandexContainerRegistryRead,
		Update: resourceYandexContainerRegistryUpdate,
		Delete: resourceYandexContainerRegistryDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexContainerRegistryDefaultTimeout),
			Update: schema.DefaultTimeout(yandexContainerRegistryDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexContainerRegistryDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},

			"name": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},

			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceYandexContainerRegistryCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	folderID, err := getFolderID(d, config)
	if err != nil {
		return fmt.Errorf("Error getting folder ID while creating Container Registry: %s", err)
	}

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return fmt.Errorf("Error expanding labels while creating Container Registry: %s", err)
	}

	req := containerregistry.CreateRegistryRequest{
		FolderId: folderID,
		Name:     d.Get("name").(string),
		Labels:   labels,
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.ContainerRegistry().Registry().Create(ctx, &req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create Container Registry: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("Error while get Container Registry create operation metadata: %s", err)
	}

	md, ok := protoMetadata.(*containerregistry.CreateRegistryMetadata)
	if !ok {
		return fmt.Errorf("could not get Container Registry ID from create operation metadata")
	}

	d.SetId(md.RegistryId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while waiting operation to create Container Registry: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("Container Registry creation failed: %s", err)
	}

	return resourceYandexContainerRegistryRead(d, meta)
}

func resourceYandexContainerRegistryRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	registry, err := config.sdk.ContainerRegistry().Registry().Get(context.Background(),
		&containerregistry.GetRegistryRequest{
			RegistryId: d.Id(),
		})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Container Registry %q", d.Id()))
	}

	d.Set("created_at", getTimestamp(registry.CreatedAt))
	d.Set("name", registry.Name)
	d.Set("folder_id", registry.FolderId)
	d.Set("status", strings.ToLower(registry.Status.String()))

	return d.Set("labels", registry.Labels)
}

func resourceYandexContainerRegistryUpdate(d *schema.ResourceData, meta interface{}) error {

	req := &containerregistry.UpdateRegistryRequest{
		RegistryId: d.Id(),
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

	if len(req.UpdateMask.Paths) == 0 {
		return fmt.Errorf("No fields were updated for Container Registry %s", d.Id())
	}

	err := makeRegistryUpdateRequest(req, d, meta)
	if err != nil {
		return err
	}

	return resourceYandexContainerRegistryRead(d, meta)
}

func resourceYandexContainerRegistryDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Deleting Container Registry %q", d.Id())

	req := &containerregistry.DeleteRegistryRequest{
		RegistryId: d.Id(),
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.ContainerRegistry().Registry().Delete(ctx, req))
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Container Registry %q", d.Id()))
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished deleting Container Registry %q", d.Id())
	return nil
}

func makeRegistryUpdateRequest(req *containerregistry.UpdateRegistryRequest, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.ContainerRegistry().Registry().Update(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to update Container Registry %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error updating Container Registry %q: %s", d.Id(), err)
	}

	return nil
}
