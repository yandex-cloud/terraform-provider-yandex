package yandex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
				Computed: true,
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
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"subnet_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"default_security_group_id": {
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

func resourceYandexVPCNetworkCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return fmt.Errorf("Error expanding labels while creating network: %s", err)
	}

	folderID, err := getFolderID(d, config)
	if err != nil {
		return fmt.Errorf("Error getting folder ID while creating network: %s", err)
	}

	req := vpc.CreateNetworkRequest{
		FolderId:    folderID,
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Labels:      labels,
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.VPC().Network().Create(ctx, &req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create network: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("Error while get network create operation metadata: %s", err)
	}

	md, ok := protoMetadata.(*vpc.CreateNetworkMetadata)
	if !ok {
		return fmt.Errorf("could not get Network ID from create operation metadata")
	}

	d.SetId(md.NetworkId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while waiting operation to create network: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("Network creation failed: %s", err)
	}

	return resourceYandexVPCNetworkRead(d, meta)
}

func resourceYandexVPCNetworkRead(d *schema.ResourceData, meta interface{}) error {
	return yandexVPCNetworkRead(d, meta, d.Id())
}

func yandexVPCNetworkRead(d *schema.ResourceData, meta interface{}, id string) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutRead))
	defer cancel()

	network, err := config.sdk.VPC().Network().Get(ctx, &vpc.GetNetworkRequest{
		NetworkId: id,
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Network %q", d.Get("name").(string)))
	}

	subnets, err := config.sdk.VPC().Network().ListSubnets(ctx, &vpc.ListNetworkSubnetsRequest{
		NetworkId: id,
	})

	if err != nil {
		return err
	}

	subnetIds := make([]string, len(subnets.Subnets))
	for i, subnet := range subnets.Subnets {
		subnetIds[i] = subnet.Id
	}

	d.Set("created_at", getTimestamp(network.CreatedAt))
	d.Set("name", network.Name)
	d.Set("folder_id", network.FolderId)
	d.Set("description", network.Description)
	d.Set("default_security_group_id", network.DefaultSecurityGroupId)
	if err := d.Set("subnet_ids", subnetIds); err != nil {
		return err
	}

	return d.Set("labels", network.Labels)
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
		req.Description = d.Get("description").(string)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "description")
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.VPC().Network().Update(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to update Network %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error updating Network %q: %s", d.Id(), err)
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

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutDelete))
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
