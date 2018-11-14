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

const yandexVPCSubnetDefaultTimeout = 1 * time.Minute

func resourceYandexVPCSubnet() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexVPCSubnetCreate,
		Read:   resourceYandexVPCSubnetRead,
		Update: resourceYandexVPCSubnetUpdate,
		Delete: resourceYandexVPCSubnetDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexVPCSubnetDefaultTimeout),
			Update: schema.DefaultTimeout(yandexVPCSubnetDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexVPCSubnetDefaultTimeout),
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

			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},

			"zone": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},

			"network_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"v4_cidr_blocks": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateIPV4CidrBlocks,
				},
			},
			"v6_cidr_blocks": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}

}

func resourceYandexVPCSubnetCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return fmt.Errorf("Error creating subnet: %s", err)
	}

	folderID, err := getFolderID(d, config)
	if err != nil {
		return fmt.Errorf("Error creating subnet: %s", err)
	}

	zone, err := getZone(d, config)
	if err != nil {
		return err
	}

	rangesV4 := []string{}

	if v, ok := d.GetOk("v4_cidr_blocks"); ok {
		vS := v.([]interface{})
		for _, cidr := range vS {
			rangesV4 = append(rangesV4, cidr.(string))
		}
	}

	rangesV6 := []string{}

	if v, ok := d.GetOk("v6_cidr_blocks"); ok {
		vS := v.([]interface{})
		for _, cidr := range vS {
			rangesV6 = append(rangesV6, cidr.(string))
		}
	}

	req := vpc.CreateSubnetRequest{
		FolderId:     folderID,
		ZoneId:       zone,
		Name:         d.Get("name").(string),
		Description:  d.Get("description").(string),
		Labels:       labels,
		NetworkId:    d.Get("network_id").(string),
		V4CidrBlocks: rangesV4,
		V6CidrBlocks: rangesV6,
	}

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.VPC().Subnet().Create(ctx, &req))
	if err != nil {
		return fmt.Errorf("Error creating subnet: %s", err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error create subnet: %s", err)
	}

	resp, err := op.Response()
	if err != nil {
		return err
	}

	subnet, ok := resp.(*vpc.Subnet)
	if !ok {
		return errors.New("response doesn't contain Subnet")
	}

	d.SetId(subnet.Id)

	return resourceYandexVPCSubnetRead(d, meta)
}

func resourceYandexVPCSubnetRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	network, err := config.sdk.VPC().Subnet().Get(context.Background(), &vpc.GetSubnetRequest{
		SubnetId: d.Id(),
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Subnet %q", d.Get("name").(string)))
	}

	d.Set("name", network.Name)
	d.Set("folder_id", network.FolderId)
	d.Set("description", network.Description)
	d.Set("labels", network.Labels)
	d.Set("network_id", network.NetworkId)
	if err := d.Set("v4_cidr_blocks", convertStringArrToInterface(network.V4CidrBlocks)); err != nil {
		return err
	}
	if err := d.Set("v6_cidr_blocks", convertStringArrToInterface(network.V6CidrBlocks)); err != nil {
		return err
	}
	d.Set("zone", network.ZoneId)
	return nil
}

func resourceYandexVPCSubnetUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	d.Partial(true)

	req := &vpc.UpdateSubnetRequest{
		SubnetId:   d.Id(),
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

	op, err := config.sdk.WrapOperation(config.sdk.VPC().Subnet().Update(ctx, req))
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Subnet %q", d.Get("name").(string)))
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error updating Subnet %q: %s", d.Id(), err)
	}

	for _, v := range req.UpdateMask.Paths {
		d.SetPartial(v)
	}

	d.Partial(false)

	return resourceYandexVPCSubnetRead(d, meta)
}

func resourceYandexVPCSubnetDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Deleting Subnet %q", d.Id())

	req := &vpc.DeleteSubnetRequest{
		SubnetId: d.Id(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.VPC().Subnet().Delete(ctx, req))
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Subnet %q", d.Get("name").(string)))
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished deleting Subnet %q", d.Id())
	return nil
}
