package yandex

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
)

const yandexVPCNetworkDefaultTimeout = 1 * time.Minute

func resourceYandexVPCNetwork() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexVPCNetworkCreate,
		Read:   resourceYandexVPCNetworkRead,
		Update: resourceYandexVPCNetworkUpdate,
		Delete: resourceYandexVPCNetworkDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexVPCNetworkDefaultTimeout),
			Update: schema.DefaultTimeout(yandexVPCNetworkDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexVPCNetworkDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
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
		},
	}

}

func resourceYandexVPCNetworkCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return fmt.Errorf("Error creating network: %s", err)
	}

	folderID, err := getFolderID(d, config)
	if err != nil {
		return fmt.Errorf("Error creating subnet: %s", err)
	}

	req := vpc.CreateNetworkRequest{
		FolderId:    folderID,
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Labels:      labels,
	}

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.VPC().Network().Create(ctx, &req))
	if err != nil {
		return fmt.Errorf("Error creating network: %s", err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error create network: %s", err)
	}

	resp, err := op.Response()
	if err != nil {
		return err
	}

	network, ok := resp.(*vpc.Network)
	if !ok {
		return errors.New("response doesn't contain Network")
	}

	d.SetId(network.Id)

	return resourceYandexVPCNetworkRead(d, meta)
}

func resourceYandexVPCNetworkRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	network, err := config.sdk.VPC().Network().Get(context.Background(), &vpc.GetNetworkRequest{
		NetworkId: d.Id(),
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Network %q", d.Get("name").(string)))
	}

	d.Set("name", network.Name)
	d.Set("folder_id", network.FolderId)
	d.Set("description", network.Description)
	d.Set("labels", network.Labels)

	return nil
}

func resourceYandexVPCNetworkUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	d.Partial(true)

	req := &vpc.UpdateNetworkRequest{
		NetworkId:  d.Id(),
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
		req.Name = d.Get("description").(string)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "description")
	}

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.VPC().Network().Update(ctx, req))
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Network %q", d.Get("name").(string)))
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error updating Network %q: %s", d.Id(), err)
	}

	for _, v := range req.UpdateMask.Paths {
		d.SetPartial(v)
	}

	d.Partial(false)

	return resourceYandexVPCNetworkRead(d, meta)
}

func resourceYandexVPCNetworkDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Deleting Network %q", d.Id())

	req := &vpc.DeleteNetworkRequest{
		NetworkId: d.Id(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.VPC().Network().Delete(ctx, req))
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Network %q", d.Get("name").(string)))
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished deleting Network %q", d.Id())
	return nil
}
