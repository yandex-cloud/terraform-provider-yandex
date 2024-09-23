package yandex

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1/privatelink"
	"google.golang.org/genproto/protobuf/field_mask"
)

const yandexVPCPrivateEndpointDefaultTimeout = 1 * time.Minute

func resourceYandexVPCPrivateEndpoint() *schema.Resource {
	return &schema.Resource{
		// TODO: Should we use *Context methods?
		Create: resourceYandexVPCPrivateEndpointCreate,
		Read:   resourceYandexVPCPrivateEndpointRead,
		Update: resourceYandexVPCPrivateEndpointUpdate,
		Delete: resourceYandexVPCPrivateEndpointDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexVPCPrivateEndpointDefaultTimeout),
			Update: schema.DefaultTimeout(yandexVPCPrivateEndpointDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexVPCPrivateEndpointDefaultTimeout),
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

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"network_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"object_storage": {
				Type:     schema.TypeList,
				MaxItems: 1,
				// NOTE: For now we require object storage, but later there will be more choices / services.
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{},
				},
			},

			"endpoint_address": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subnet_id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"address": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"address_id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
			},

			"dns_options": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"private_dns_records_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceYandexVPCPrivateEndpointCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return fmt.Errorf("Error expanding labels while creating gateway: %s", err)
	}

	folderID, err := getFolderID(d, config)
	if err != nil {
		return fmt.Errorf("Error getting folder ID while creating gateway: %s", err)
	}

	dnsOptions, err := expandPrivateEndpointDnsOptions(d)
	if err != nil {
		return fmt.Errorf("error getting dns options while creating private endpoint: %s", err)
	}

	addressSpec, err := expandPrivateEndpointAddressSpec(d)
	if err != nil {
		return fmt.Errorf("error getting address spec while creating private endpoint: %s", err)
	}

	req := privatelink.CreatePrivateEndpointRequest{
		FolderId:    folderID,
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Labels:      labels,
		NetworkId:   d.Get("network_id").(string),
		DnsOptions:  dnsOptions,
		AddressSpec: addressSpec,
	}

	if d.Get("object_storage") != nil {
		req.Service = &privatelink.CreatePrivateEndpointRequest_ObjectStorage{}
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.VPCPrivateLink().PrivateEndpoint().Create(ctx, &req))
	if err != nil {
		return fmt.Errorf("error while requesting API to create private endpoint: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("error while get private endpoint create operation metadata: %s", err)
	}

	md, ok := protoMetadata.(*privatelink.CreatePrivateEndpointMetadata)
	if !ok {
		return fmt.Errorf("could not get PrivateEndpoint ID from create operation metadata")
	}

	d.SetId(md.PrivateEndpointId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while waiting operation to create private endpoint: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("private endpoint creation failed: %s", err)
	}

	return resourceYandexVPCPrivateEndpointRead(d, meta)
}

func resourceYandexVPCPrivateEndpointRead(d *schema.ResourceData, meta interface{}) error {
	return yandexVPCPrivateEndpointRead(d, meta, d.Id())
}

func handlePrivateEndpointNotFoundError(err error, d *schema.ResourceData, id string) error {
	return handleNotFoundError(err, d, fmt.Sprintf("VPC Private Endpoint %q", id))
}

func yandexVPCPrivateEndpointRead(d *schema.ResourceData, meta interface{}, id string) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutRead))
	defer cancel()

	privateEndpoint, err := config.sdk.VPCPrivateLink().PrivateEndpoint().Get(ctx, &privatelink.GetPrivateEndpointRequest{
		PrivateEndpointId: id,
	})

	if err != nil {
		return handlePrivateEndpointNotFoundError(err, d, d.Id())
	}

	if err := d.Set("created_at", getTimestamp(privateEndpoint.GetCreatedAt())); err != nil {
		return err
	}
	if err := d.Set("name", privateEndpoint.GetName()); err != nil {
		return err
	}
	if err := d.Set("folder_id", privateEndpoint.GetFolderId()); err != nil {
		return err
	}
	if err := d.Set("network_id", privateEndpoint.GetNetworkId()); err != nil {
		return err
	}
	if err := d.Set("description", privateEndpoint.GetDescription()); err != nil {
		return err
	}
	if err := d.Set("status", privateEndpoint.GetStatus().String()); err != nil {
		return err
	}

	addr := flattenPrivateEndpointAddress(privateEndpoint.GetAddress())
	if err := d.Set("endpoint_address", addr); err != nil {
		return err
	}

	dnsOpt := flattenPrivateEndpointDnsOptions(privateEndpoint.GetDnsOptions())
	if err := d.Set("dns_options", dnsOpt); err != nil {
		return err
	}

	switch v := privateEndpoint.Service.(type) {
	case *privatelink.PrivateEndpoint_ObjectStorage_:
		objStorage := flattenPrivateEndpointObjectStorage(v.ObjectStorage)
		if err := d.Set("object_storage", objStorage); err != nil {
			return err
		}
	}

	return d.Set("labels", privateEndpoint.GetLabels())
}

func resourceYandexVPCPrivateEndpointUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	d.Partial(true)

	req := &privatelink.UpdatePrivateEndpointRequest{
		PrivateEndpointId: d.Id(),
		UpdateMask:        &field_mask.FieldMask{},
	}

	labelsPropName := "labels"
	if d.HasChange(labelsPropName) {
		labelsProp, err := expandLabels(d.Get(labelsPropName))
		if err != nil {
			return err
		}

		req.Labels = labelsProp
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, labelsPropName)
	}

	namePropName := "name"
	if d.HasChange(namePropName) {
		req.Name = d.Get(namePropName).(string)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, namePropName)
	}

	descPropName := "description"
	if d.HasChange(descPropName) {
		req.Description = d.Get(descPropName).(string)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, descPropName)
	}

	if d.HasChange("endpoint_address") {
		addressSpec, err := expandPrivateEndpointAddressSpec(d)
		if err != nil {
			return fmt.Errorf("error getting address spec while creating private endpoint: %s", err)
		}

		req.AddressSpec = addressSpec
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "address_spec")
	}

	dnsOptPropName := "dns_options"
	if d.HasChange(dnsOptPropName) {
		dnsOptions, err := expandPrivateEndpointDnsOptions(d)
		if err != nil {
			return fmt.Errorf("error getting dns options while updating private endpoint: %s", err)
		}

		req.DnsOptions = dnsOptions
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, dnsOptPropName)
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.VPCPrivateLink().PrivateEndpoint().Update(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to update Private Endpoint %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error updating Private Endpoint %q: %s", d.Id(), err)
	}

	d.Partial(false)

	return resourceYandexVPCPrivateEndpointRead(d, meta)
}

func resourceYandexVPCPrivateEndpointDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	req := &privatelink.DeletePrivateEndpointRequest{
		PrivateEndpointId: d.Id(),
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.VPCPrivateLink().PrivateEndpoint().Delete(ctx, req))
	if err != nil {
		return handlePrivateEndpointNotFoundError(err, d, d.Id())
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	if err != nil {
		return err
	}

	return nil
}

func expandPrivateEndpointDnsOptions(d *schema.ResourceData) (*privatelink.PrivateEndpoint_DnsOptions, error) {
	if d.Get("dns_options.#").(int) == 0 {
		return nil, nil
	}

	return &privatelink.PrivateEndpoint_DnsOptions{
		PrivateDnsRecordsEnabled: d.Get("dns_options.0.private_dns_records_enabled").(bool),
	}, nil
}

func expandPrivateEndpointAddressSpec(d *schema.ResourceData) (*privatelink.AddressSpec, error) {
	if v, ok := d.GetOk("endpoint_address.0.subnet_id"); ok {
		return &privatelink.AddressSpec{
			Address: &privatelink.AddressSpec_InternalIpv4AddressSpec{
				InternalIpv4AddressSpec: &privatelink.InternalIpv4AddressSpec{
					Address:  d.Get("endpoint_address.0.address").(string),
					SubnetId: v.(string),
				},
			},
		}, nil
	} else if v, ok := d.GetOk("endpoint_address.0.address_id"); ok {
		return &privatelink.AddressSpec{
			Address: &privatelink.AddressSpec_AddressId{
				AddressId: v.(string),
			},
		}, nil
	}

	return nil, nil
}

func flattenPrivateEndpointDnsOptions(dnsOpts *privatelink.PrivateEndpoint_DnsOptions) []map[string]interface{} {
	res := make(map[string]interface{})
	res["private_dns_records_enabled"] = dnsOpts.PrivateDnsRecordsEnabled
	return []map[string]interface{}{res}
}

func flattenPrivateEndpointAddress(endpointAddr *privatelink.PrivateEndpoint_EndpointAddress) []map[string]interface{} {
	res := make(map[string]interface{})
	res["address"] = endpointAddr.Address
	res["address_id"] = endpointAddr.AddressId
	res["subnet_id"] = endpointAddr.SubnetId
	return []map[string]interface{}{res}
}

func flattenPrivateEndpointObjectStorage(_ *privatelink.PrivateEndpoint_ObjectStorage) []map[string]interface{} {
	// NOTE: Just empty map for now.
	res := make(map[string]interface{})
	return []map[string]interface{}{res}
}
