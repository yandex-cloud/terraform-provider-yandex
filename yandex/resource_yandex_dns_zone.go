package yandex

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/dns/v1"
)

const yandexDnsDefaultTimeout = 5 * time.Minute

func resourceYandexDnsZone() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexDnsZoneCreate,
		Read:   resourceYandexDnsZoneRead,
		Update: resourceYandexDnsZoneUpdate,
		Delete: resourceYandexDnsZoneDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexDnsDefaultTimeout),
			Update: schema.DefaultTimeout(yandexDnsDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexDnsDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"zone": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateZoneName(),
			},

			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},

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

			"public": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"private_networks": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Set: schema.HashString,
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceYandexDnsZoneCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	sdk := getSDK(config)

	folderID, err := getFolderID(d, config)
	if err != nil {
		return fmt.Errorf("Error getting folder ID while creating DnsZone: %s", err)
	}

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return fmt.Errorf("Error expanding labels while creating DnsZone: %s", err)
	}

	req := dns.CreateDnsZoneRequest{
		FolderId:          folderID,
		Name:              d.Get("name").(string),
		Description:       d.Get("description").(string),
		Labels:            labels,
		Zone:              d.Get("zone").(string),
		PrivateVisibility: &dns.PrivateVisibility{},
	}

	if d.Get("public").(bool) {
		req.PublicVisibility = &dns.PublicVisibility{}
	}

	if n, ok := d.GetOk("private_networks"); ok {
		req.PrivateVisibility.NetworkIds = convertStringSet(n.(*schema.Set))
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := sdk.WrapOperation(sdk.DNS().DnsZone().Create(ctx, &req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create DnsZone: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("Error while get DnsZone create operation metadata: %s", err)
	}

	md, ok := protoMetadata.(*dns.CreateDnsZoneMetadata)
	if !ok {
		return fmt.Errorf("could not get DnsZone ID from create operation metadata")
	}

	d.SetId(md.DnsZoneId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while waiting operation to create DnsZone: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("DnsZone creation failed: %s", err)
	}

	return resourceYandexDnsZoneRead(d, meta)
}

func resourceYandexDnsZoneRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	sdk := getSDK(config)

	dnsZone, err := sdk.DNS().DnsZone().Get(config.Context(), &dns.GetDnsZoneRequest{
		DnsZoneId: d.Id(),
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("DnsZone %q", d.Get("name").(string)))
	}

	createdAt, err := getTimestamp(dnsZone.CreatedAt)
	if err != nil {
		return err
	}

	d.Set("created_at", createdAt)
	d.Set("name", dnsZone.Name)
	d.Set("folder_id", dnsZone.FolderId)
	d.Set("zone", dnsZone.Zone)
	d.Set("description", dnsZone.Description)

	d.Set("public", dnsZone.PublicVisibility != nil)

	if dnsZone.PrivateVisibility != nil {
		if err := d.Set("private_networks", convertStringArrToInterface(dnsZone.PrivateVisibility.GetNetworkIds())); err != nil {
			return err
		}
	}
	return d.Set("labels", dnsZone.Labels)
}

func resourceYandexDnsZoneUpdate(d *schema.ResourceData, meta interface{}) error {
	req, err := prepareDnsZoneUpdateRequest(d)
	if err != nil {
		return err
	}

	err = makeDnsZoneUpdateRequest(req, d, meta)
	if err != nil {
		return err
	}

	return resourceYandexDnsZoneRead(d, meta)
}

func resourceYandexDnsZoneDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	sdk := getSDK(config)

	log.Printf("[DEBUG] Deleting DnsZone %q", d.Id())

	req := &dns.DeleteDnsZoneRequest{
		DnsZoneId: d.Id(),
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := sdk.WrapOperation(sdk.DNS().DnsZone().Delete(ctx, req))
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("DnsZone %q", d.Get("name").(string)))
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	resp, err := op.Response()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished deleting DnsZone %q: %#v", d.Id(), resp)
	return nil
}

func prepareDnsZoneUpdateRequest(d *schema.ResourceData) (*dns.UpdateDnsZoneRequest, error) {
	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return nil, fmt.Errorf("Error expanding labels while creating instance: %s", err)
	}

	req := &dns.UpdateDnsZoneRequest{
		DnsZoneId:         d.Id(),
		Name:              d.Get("name").(string),
		Description:       d.Get("description").(string),
		Labels:            labels,
		PrivateVisibility: &dns.PrivateVisibility{},
	}

	if d.Get("public").(bool) {
		req.PublicVisibility = &dns.PublicVisibility{}
	}

	if n, ok := d.GetOk("private_networks"); ok {
		req.PrivateVisibility.NetworkIds = convertStringSet(n.(*schema.Set))
	}

	return req, nil
}

func makeDnsZoneUpdateRequest(req *dns.UpdateDnsZoneRequest, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	sdk := getSDK(config)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := sdk.WrapOperation(sdk.DNS().DnsZone().Update(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to update DnsZone %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error updating DnsZone %q: %s", d.Id(), err)
	}

	return nil
}

func validateZoneName() schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}
		if len(v) > 255 {
			es = append(es, fmt.Errorf("expected length of %s to be less than 256, got %s", k, v))
		}
		if !strings.HasSuffix(v, ".") {
			es = append(es, fmt.Errorf("zone name must ends with '.'"))
		}
		return
	}
}
