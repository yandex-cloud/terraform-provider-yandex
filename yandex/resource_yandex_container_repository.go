package yandex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"
)

const yandexContainerRepositoryDefaultTimeout = 5 * time.Minute

func resourceYandexContainerRepository() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexContainerRepositoryCreate,
		Read:   resourceYandexContainerRepositoryRead,
		Delete: resourceYandexContainerRepositoryDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexContainerRepositoryDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexContainerRepositoryDefaultTimeout),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceYandexContainerRepositoryCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	req := containerregistry.UpsertRepositoryRequest{
		Name: d.Get("name").(string),
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.ContainerRegistry().Repository().Upsert(ctx, &req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create Container Repository: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("Error while get Container Repository create operation metadata: %s", err)
	}

	md, ok := protoMetadata.(*containerregistry.UpsertRepositoryMetadata)
	if !ok {
		return fmt.Errorf("could not get Container Repository ID from create operation metadata")
	}

	d.SetId(md.RepositoryId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while waiting operation to create Container Repository: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("Container Repository creation failed: %s", err)
	}

	return resourceYandexContainerRepositoryRead(d, meta)
}

func resourceYandexContainerRepositoryRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	repository, err := config.sdk.ContainerRegistry().Repository().Get(context.Background(),
		&containerregistry.GetRepositoryRequest{
			RepositoryId: d.Id(),
		})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Container Repository %q", d.Id()))
	}

	return d.Set("name", repository.Name)
}

func resourceYandexContainerRepositoryDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Deleting Container Repository %q", d.Id())

	req := &containerregistry.DeleteRepositoryRequest{
		RepositoryId: d.Id(),
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.ContainerRegistry().Repository().Delete(ctx, req))
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Container Repository %q", d.Id()))
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished deleting Container Repository %q", d.Id())
	return nil
}
