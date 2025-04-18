package yandex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

const yandexVPCSubnetDefaultTimeout = 3 * time.Minute

func resourceYandexVPCSubnet() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a subnet within the Yandex Cloud. For more information, see [the official documentation](https://yandex.cloud/docs/vpc/concepts/network#subnet).\n\n* How-to Guides\n  * [Cloud Networking](https://yandex.cloud/docs/vpc/)\n  * [VPC Addressing](https://yandex.cloud/docs/vpc/concepts/address)\n",

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
			"network_id": {
				Type:        schema.TypeString,
				Description: "ID of the network this subnet belongs to. Only networks that are in the distributed mode can have subnets.",
				Required:    true,
				ForceNew:    true,
			},

			"v4_cidr_blocks": {
				Type:        schema.TypeList,
				Description: "A list of blocks of internal IPv4 addresses that are owned by this subnet. Provide this property when you create the subnet. For example, `10.0.0.0/22` or `192.168.0.0/16`. Blocks of addresses must be unique and non-overlapping within a network. Minimum subnet size is `/28`, and maximum subnet size is `/16`. Only IPv4 is supported.",
				Required:    true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateCidrBlocks,
				},
			},

			"name": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Optional:    true,
				Computed:    true,
			},

			"description": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["description"],
				Optional:    true,
			},

			"folder_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["folder_id"],
				Computed:    true,
				Optional:    true,
				ForceNew:    true,
			},

			"labels": {
				Type:        schema.TypeMap,
				Description: common.ResourceDescriptions["labels"],
				Optional:    true,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},

			"zone": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["zone"],
				Computed:    true,
				Optional:    true,
				ForceNew:    true,
			},

			"route_table_id": {
				Type:        schema.TypeString,
				Description: "The ID of the route table to assign to this subnet. Assigned route table should belong to the same network as this subnet.",
				Optional:    true,
			},

			"v6_cidr_blocks": {
				Type:        schema.TypeList,
				Description: "An optional list of blocks of IPv6 addresses that are owned by this subnet.",
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},

			"dhcp_options": {
				Type:        schema.TypeList,
				Description: "Options for DHCP client.",
				MaxItems:    1,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"domain_name": {
							Type:        schema.TypeString,
							Description: "Domain name.",
							Optional:    true,
						},
						"domain_name_servers": {
							Type:        schema.TypeList,
							Description: "Domain name server IP addresses.",
							Elem:        &schema.Schema{Type: schema.TypeString},
							Optional:    true,
						},
						"ntp_servers": {
							Type:        schema.TypeList,
							Description: "NTP server IP addresses.",
							Elem:        &schema.Schema{Type: schema.TypeString},
							Optional:    true,
						},
					},
				},
			},

			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},
		},
	}

}

func resourceYandexVPCSubnetCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	zone, err := getZone(d, config)
	if err != nil {
		return fmt.Errorf("Error getting zone while creating subnet: %s", err)
	}

	folderID, err := getFolderID(d, config)
	if err != nil {
		return fmt.Errorf("Error getting folder ID while creating subnet: %s", err)
	}

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return fmt.Errorf("Error expanding labels while creating subnet: %s", err)
	}

	rangesV4 := []string{}

	if v, ok := d.GetOk("v4_cidr_blocks"); ok {
		vS := v.([]interface{})
		for _, cidr := range vS {
			rangesV4 = append(rangesV4, cidr.(string))
		}
	}

	dhcpOptions, err := expandDhcpOptions(d)
	if err != nil {
		return fmt.Errorf("Error expanding dhcp options while creating subnet: %s", err)
	}

	req := vpc.CreateSubnetRequest{
		FolderId:     folderID,
		ZoneId:       zone,
		Name:         d.Get("name").(string),
		Description:  d.Get("description").(string),
		Labels:       labels,
		NetworkId:    d.Get("network_id").(string),
		RouteTableId: d.Get("route_table_id").(string),
		V4CidrBlocks: rangesV4,
		DhcpOptions:  dhcpOptions,
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.VPC().Subnet().Create(ctx, &req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create subnet: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("Error while get subnet create operation metadata: %s", err)
	}

	md, ok := protoMetadata.(*vpc.CreateSubnetMetadata)
	if !ok {
		return fmt.Errorf("could not get Subnet ID from create operation metadata")
	}

	d.SetId(md.SubnetId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while waiting operation to create subnet: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("Subnet creation failed: %s", err)
	}

	return resourceYandexVPCSubnetRead(d, meta)
}

func resourceYandexVPCSubnetRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	subnet, err := config.sdk.VPC().Subnet().Get(config.Context(), &vpc.GetSubnetRequest{
		SubnetId: d.Id(),
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Subnet %q", d.Get("name").(string)))
	}

	d.Set("created_at", getTimestamp(subnet.CreatedAt))
	d.Set("name", subnet.Name)
	d.Set("zone", subnet.ZoneId)
	d.Set("folder_id", subnet.FolderId)
	d.Set("description", subnet.Description)
	d.Set("network_id", subnet.NetworkId)
	d.Set("route_table_id", subnet.RouteTableId)

	if err := d.Set("labels", subnet.Labels); err != nil {
		return err
	}

	if err := d.Set("v4_cidr_blocks", convertStringArrToInterface(subnet.V4CidrBlocks)); err != nil {
		return err
	}

	if err := d.Set("v6_cidr_blocks", convertStringArrToInterface(subnet.V6CidrBlocks)); err != nil {
		return err
	}

	return d.Set("dhcp_options", flattenDhcpOptions(subnet.DhcpOptions))
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
		req.Description = d.Get("description").(string)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "description")
	}

	if d.HasChange("route_table_id") {
		req.RouteTableId = d.Get("route_table_id").(string)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "route_table_id")
	}

	if d.HasChange("dhcp_options") {
		dhcpOptions, err := expandDhcpOptions(d)
		if err != nil {
			return err
		}
		req.DhcpOptions = dhcpOptions
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "dhcp_options")
	}

	if d.HasChange("v4_cidr_blocks") {
		rangesV4 := []string{}

		if v, ok := d.GetOk("v4_cidr_blocks"); ok {
			vS := v.([]interface{})
			for _, cidr := range vS {
				rangesV4 = append(rangesV4, cidr.(string))
			}
		}

		req.V4CidrBlocks = rangesV4
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "v4_cidr_blocks")
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.VPC().Subnet().Update(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to update Subnet %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error updating Subnet %q: %s", d.Id(), err)
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

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutDelete))
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
