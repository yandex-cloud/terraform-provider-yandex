package yandex

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/dns/v1"
	"github.com/yandex-cloud/go-sdk/operation"
	"google.golang.org/grpc/status"
)

const yandexDnsDefaultTimeout = 5 * time.Minute

func resourceYandexDnsZone() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a DNS Zone.",
		Create:      resourceYandexDnsZoneCreate,
		Read:        resourceYandexDnsZoneRead,
		Update:      resourceYandexDnsZoneUpdate,
		Delete:      resourceYandexDnsZoneDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: zoneVisibilityChanged,

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
				Description:  "The DNS name of this zone, e.g. \"example.com.\". Must ends with dot.",
			},

			"folder_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				ForceNew:    true,
				Description: "ID of the folder to create a zone in. If it is not provided, the default provider folder is used.",
			},

			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "User assigned name of a specific resource. Must be unique within the folder.",
			},

			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the DNS zone.",
			},

			"labels": {
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Description: "A set of key/value label pairs to assign to the DNS zone.",
			},

			"public": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "The zone's visibility: public zones are exposed to the Internet, while private zones are visible only to Virtual Private Cloud resources.",
			},

			"private_networks": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Set:         schema.HashString,
				Description: "For privately visible zones, the set of Virtual Private Cloud resources that the zone is visible from.",
			},

			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The DNS zone creation timestamp.",
			},

			"deletion_protection": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Flag that protects the dns zone from accidental deletion.",
			},
		},
	}
}

func resourceYandexDnsZoneCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	folderID, err := getFolderID(d, config)
	if err != nil {
		return fmt.Errorf("Error getting folder ID while creating DnsZone: %s", err)
	}

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return fmt.Errorf("Error expanding labels while creating DnsZone: %s", err)
	}

	req := &dns.CreateDnsZoneRequest{
		FolderId:           folderID,
		Name:               d.Get("name").(string),
		Description:        d.Get("description").(string),
		Labels:             labels,
		Zone:               d.Get("zone").(string),
		DeletionProtection: d.Get("deletion_protection").(bool),
	}

	if d.Get("public").(bool) {
		req.PublicVisibility = &dns.PublicVisibility{}
	} else {
		req.PrivateVisibility = &dns.PrivateVisibility{}
	}

	if n, ok := d.GetOk("private_networks"); ok {
		if req.PrivateVisibility == nil {
			req.PrivateVisibility = &dns.PrivateVisibility{}
		}
		req.PrivateVisibility.NetworkIds = convertStringSet(n.(*schema.Set))
	}

	if err := makeDnsZoneCreateRequest(req, d, meta); err != nil {
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

	d.Set("created_at", getTimestamp(dnsZone.CreatedAt))
	d.Set("name", dnsZone.Name)
	d.Set("folder_id", dnsZone.FolderId)
	d.Set("zone", dnsZone.Zone)
	d.Set("description", dnsZone.Description)
	d.Set("deletion_protection", dnsZone.DeletionProtection)

	d.Set("public", dnsZone.PublicVisibility != nil)

	if dnsZone.PrivateVisibility != nil {
		if err := d.Set("private_networks", convertStringArrToInterface(dnsZone.PrivateVisibility.GetNetworkIds())); err != nil {
			return err
		}
	}
	return d.Set("labels", dnsZone.Labels)
}

func resourceYandexDnsZoneUpdate(d *schema.ResourceData, meta interface{}) error {
	if shouldReplaceZone(d.Id(), d.HasChange("public"), d.HasChange("private_networks"), d.Get("public").(bool), meta) {
		err := resourceYandexDnsZoneDelete(d, meta)
		if err != nil {
			return err
		}

		return resourceYandexDnsZoneCreate(d, meta)
	}

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
		DnsZoneId:          d.Id(),
		Name:               d.Get("name").(string),
		Description:        d.Get("description").(string),
		Labels:             labels,
		DeletionProtection: d.Get("deletion_protection").(bool),
	}

	if d.Get("public").(bool) {
		req.PublicVisibility = &dns.PublicVisibility{}
	} else {
		req.PrivateVisibility = &dns.PrivateVisibility{}
	}

	if n, ok := d.GetOk("private_networks"); ok {
		if req.PrivateVisibility == nil {
			req.PrivateVisibility = &dns.PrivateVisibility{}
		}
		req.PrivateVisibility.NetworkIds = convertStringSet(n.(*schema.Set))
	}

	return req, nil
}

func makeDnsZoneCreateRequest(req *dns.CreateDnsZoneRequest, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	sdk := getSDK(config)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	timeouts := []time.Duration{time.Millisecond * 500, time.Second * 2, time.Second * 10}

	op, err := retrySpecificError(timeouts, func() (*operation.Operation, error) {
		return sdk.WrapOperation(sdk.DNS().DnsZone().Create(ctx, req))
	}, isErrNetworkNotFound)

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

	return nil
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

func retrySpecificError(timeouts []time.Duration, fn func() (*operation.Operation, error), qualifier func(err error) bool) (*operation.Operation, error) {
	var op *operation.Operation
	var err error

	for i := 0; i < len(timeouts); i++ {
		op, err = fn()
		if err == nil {
			return op, nil
		}
		if !qualifier(err) {
			return op, err
		}

		log.Printf("[DEBUG] retry #%d, timeout %s\n", i, timeouts[i])
		time.Sleep(timeouts[i])
	}

	return fn()
}

func isErrNetworkNotFound(err error) bool {
	if grpcStatus, ok := status.FromError(err); ok {
		return strings.HasPrefix(grpcStatus.Message(), "Network not found:")
	}
	return false
}

func zoneVisibilityChanged(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
	if shouldReplaceZone(d.Id(), d.HasChange("public"), d.HasChange("private_networks"), d.Get("public").(bool), meta) {
		d.ForceNew("public")
		d.ForceNew("private_networks")
	}
	return nil
}

func shouldReplaceZone(id string, hasChangePublicAttr, hasChangePrivateNetworksAttr bool, isPublic bool, meta interface{}) bool {
	config := meta.(*Config)
	sdk := getSDK(config)

	dnsZone, _ := sdk.DNS().DnsZone().Get(config.Context(), &dns.GetDnsZoneRequest{
		DnsZoneId: id,
	})

	privateNetworkIds := dnsZone.GetPrivateVisibility().GetNetworkIds()
	networkIdsSet := hasChangePrivateNetworksAttr && privateNetworkIds == nil

	// "public-private -> public" transition is explicitly prohibited
	return (isPublic && networkIdsSet) || hasChangePublicAttr
}
