package yandex

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
	"google.golang.org/genproto/protobuf/field_mask"
)

const yandexVPCAddressDefaultTimeout = 30 * time.Second

func addressError(format string, a ...interface{}) error {
	return fmt.Errorf("VPC address: "+format, a...)
}

func handleAddressNotFoundError(err error, d *schema.ResourceData, id string) error {
	return handleNotFoundError(err, d, fmt.Sprintf("VPC address %q", id))
}

func resourceYandexVPCAddress() *schema.Resource {
	return &schema.Resource{
		Read:          resourceYandexVPCAddressRead,
		Create:        resourceYandexVPCAddressCreate,
		UpdateContext: resourceYandexVPCAddressUpdateContext,
		Delete:        resourceYandexVPCAddressDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexVPCAddressDefaultTimeout),
			Update: schema.DefaultTimeout(yandexVPCAddressDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexVPCAddressDefaultTimeout),
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
				Optional: true,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"reserved": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"external_ipv4_address": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address": {
							Type:     schema.TypeString,
							Computed: true,
							ForceNew: true,
						},
						"zone_id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
						"ddos_protection_provider": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
						"outgoing_smtp_capability": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
					},
				},
			},
			"used": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deletion_protection": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"dns_record": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"dns_zone_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"fqdn": {
							Type:     schema.TypeString,
							Required: true,
						},
						"ttl": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"ptr": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func yandexVPCAddressRead(d *schema.ResourceData, meta interface{}, id string) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	req := &vpc.GetAddressRequest{AddressId: id}
	address, err := config.sdk.VPC().Address().Get(ctx, req)

	if err != nil {
		return handleAddressNotFoundError(err, d, id)
	}

	if err := d.Set("folder_id", address.GetFolderId()); err != nil {
		return err
	}
	if err := d.Set("created_at", getTimestamp(address.GetCreatedAt())); err != nil {
		return err
	}
	if err := d.Set("name", address.GetName()); err != nil {
		return err
	}
	if err := d.Set("description", address.GetDescription()); err != nil {
		return err
	}
	if err := d.Set("labels", address.GetLabels()); err != nil {
		return err
	}
	if err := d.Set("deletion_protection", address.GetDeletionProtection()); err != nil {
		return err
	}

	v4Addr := flattenExternalIpV4AddressSpec(address.GetExternalIpv4Address())
	if err := d.Set("external_ipv4_address", v4Addr); err != nil {
		return err
	}

	if err := d.Set("reserved", address.GetReserved()); err != nil {
		return err
	}

	dnsRecords := flattenVpcAddressDnsRecords(address.DnsRecords)
	if err := d.Set("dns_record", dnsRecords); err != nil {
		return err
	}

	return d.Set("used", address.GetUsed())
}

func resourceYandexVPCAddressRead(d *schema.ResourceData, meta interface{}) error {
	return yandexVPCAddressRead(d, meta, d.Id())
}

func resourceYandexVPCAddressCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return addressError("expanding labels while creating address: %s", err)
	}

	folderID, err := getFolderID(d, config)
	if err != nil {
		return addressError("expanding folder ID while creating address: %s", err)
	}

	spec, err := expandExternalIpv4Address(d)
	if err != nil {
		return addressError("expanding external ipv4 address while creating address: %s", err)
	}

	dnsSpecs, err := expandVpcAddressDnsRecords(d)
	if err != nil {
		return addressError("expanding dns record specs while creating address %s", err)
	}

	req := vpc.CreateAddressRequest{
		FolderId:    folderID,
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Labels:      labels,

		AddressSpec: &vpc.CreateAddressRequest_ExternalIpv4AddressSpec{
			ExternalIpv4AddressSpec: spec,
		},
		DeletionProtection: d.Get("deletion_protection").(bool),
		DnsRecordSpecs:     dnsSpecs,
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.VPC().Address().Create(ctx, &req))
	if err != nil {
		return addressError("while requesting API to create address: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return addressError("while get address create operation metadata: %s", err)
	}

	md, ok := protoMetadata.(*vpc.CreateAddressMetadata)
	if !ok {
		return addressError("could not get Address ID from create operation metadata")
	}

	d.SetId(md.AddressId)

	err = op.Wait(ctx)
	if err != nil {
		return addressError("while waiting operation to create address: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return addressError("creation failed: %s", err)
	}

	return resourceYandexVPCAddressRead(d, meta)
}

func resourceYandexVPCAddressUpdateContext(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	d.Partial(true)

	req := &vpc.UpdateAddressRequest{
		AddressId:  d.Id(),
		UpdateMask: &field_mask.FieldMask{},
	}

	const addrLabelsPropName = "labels"
	if d.HasChange(addrLabelsPropName) {
		labelsProp, err := expandLabels(d.Get(addrLabelsPropName))
		if err != nil {
			return diag.FromErr(err)
		}

		req.Labels = labelsProp
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, addrLabelsPropName)
	}

	const addrNamePropName = "name"
	if d.HasChange(addrNamePropName) {
		req.Name = d.Get(addrNamePropName).(string)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, addrNamePropName)
	}

	const addrDescPropName = "description"
	if d.HasChange(addrDescPropName) {
		req.Description = d.Get(addrDescPropName).(string)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, addrDescPropName)
	}

	const addrDeletionProtectionPropName = "deletion_protection"
	if d.HasChange(addrDeletionProtectionPropName) {
		req.DeletionProtection = d.Get(addrDeletionProtectionPropName).(bool)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, addrDeletionProtectionPropName)
	}

	const addrDnsRecords = "dns_record"
	if d.HasChange(addrDnsRecords) {
		specs, err := expandVpcAddressDnsRecords(d)
		if err != nil {
			return diag.FromErr(err)
		}
		req.DnsRecordSpecs = specs
		// differs in ycp and tf
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "dns_record_specs")
	}

	var diags diag.Diagnostics
	if d.HasChange("reserved") && !req.Reserved && len(req.DnsRecordSpecs) > 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "DNS records were copied to the network interface",
			Detail: "You changed the type of address to ephemeral. This copies DNS records to the network interface. " +
				"Don't forget to update it in Terraform specification!",
		})
	}

	ctx, cancel := context.WithTimeout(ctx, d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.VPC().Address().Update(ctx, req))
	if err != nil {
		return diag.FromErr(addressError("while requesting API to update Address %q: %s", d.Id(), err))
	}

	err = op.Wait(ctx)
	if err != nil {
		return diag.FromErr(addressError("updating Address %q: %s", d.Id(), err))
	}

	d.Partial(false)

	err = resourceYandexVPCAddressRead(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceYandexVPCAddressDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	req := &vpc.DeleteAddressRequest{
		AddressId: d.Id(),
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.VPC().Address().Delete(ctx, req))
	if err != nil {
		return handleAddressNotFoundError(err, d, d.Id())
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
