package yandex

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/dns/v1"
)

func resourceYandexDnsRecordSet() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexDnsRecordSetCreate,
		Read:   resourceYandexDnsRecordSetRead,
		Update: resourceYandexDnsRecordSetUpdate,
		Delete: resourceYandexDnsRecordSetDelete,
		Importer: &schema.ResourceImporter{
			State: resourceDnsRecordSetImportState,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexDnsDefaultTimeout),
			Update: schema.DefaultTimeout(yandexDnsDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexDnsDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"zone_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 254),
			},

			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 20),
			},

			"ttl": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(0, 2147483647),
			},

			"data": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				MaxItems: 100,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(1, 1024),
				},
				Set:              schema.HashString,
				DiffSuppressFunc: dataDiffSuppressFunc,
			},
		},
	}
}

func resourceYandexDnsRecordSetCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	sdk := getSDK(config)

	rs := &dns.RecordSet{
		Name: d.Get("name").(string),
		Type: d.Get("type").(string),
		Ttl:  int64(d.Get("ttl").(int)),
		Data: convertStringSet(d.Get("data").(*schema.Set)),
	}

	req := dns.UpdateRecordSetsRequest{
		DnsZoneId: d.Get("zone_id").(string),
		Additions: []*dns.RecordSet{rs},
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := sdk.WrapOperation(sdk.DNS().DnsZone().UpdateRecordSets(ctx, &req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create DnsRecordSet: %s", err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while waiting operation to create DnsRecordSet: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("DnsRecordSet creation failed: %s", err)
	}

	d.SetId(fmt.Sprintf("%s/%s/%s", d.Get("zone_id"), d.Get("name"), d.Get("type")))

	return resourceYandexDnsRecordSetRead(d, meta)
}

func resourceYandexDnsRecordSetRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	sdk := getSDK(config)

	req := &dns.GetDnsZoneRecordSetRequest{
		DnsZoneId: d.Get("zone_id").(string),
		Type:      d.Get("type").(string),
		Name:      d.Get("name").(string),
	}

	rs, err := sdk.DNS().DnsZone().GetRecordSet(config.Context(), req)

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("DnsRecordSet %s", rsId(d)))
	}

	d.Set("ttl", int(rs.Ttl))
	d.Set("data", convertStringArrToInterface(rs.Data))

	return nil
}

func resourceYandexDnsRecordSetUpdate(d *schema.ResourceData, meta interface{}) error {
	req, err := prepareDnsRecordSetUpdateRequest(d)
	if err != nil {
		return err
	}

	err = makeDnsRecordSetUpdateRequest(req, d, meta)
	if err != nil {
		return err
	}

	return resourceYandexDnsRecordSetRead(d, meta)
}

func resourceYandexDnsRecordSetDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	sdk := getSDK(config)

	rs := &dns.RecordSet{
		Name: d.Get("name").(string),
		Type: d.Get("type").(string),
		Ttl:  int64(d.Get("ttl").(int)),
		Data: convertStringSet(d.Get("data").(*schema.Set)),
	}

	req := dns.UpdateRecordSetsRequest{
		DnsZoneId: d.Get("zone_id").(string),
		Deletions: []*dns.RecordSet{rs},
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := sdk.WrapOperation(sdk.DNS().DnsZone().UpdateRecordSets(ctx, &req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create DnsRecordSet: %s", err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while waiting operation to create DnsRecordSet: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("DnsRecordSet creation failed: %s", err)
	}

	log.Printf("[DEBUG] Finished deleting DnsRecordSet %s", rsId(d))
	return nil
}

func prepareDnsRecordSetUpdateRequest(d *schema.ResourceData) (*dns.UpdateRecordSetsRequest, error) {
	name := d.Get("name").(string)

	oldTtl, newTtl := d.GetChange("ttl")
	oldType, newType := d.GetChange("type")

	oldData, _ := d.GetChange("data")

	req := &dns.UpdateRecordSetsRequest{
		DnsZoneId: d.Get("zone_id").(string),
		Deletions: []*dns.RecordSet{
			{
				Name: name,
				Type: oldType.(string),
				Ttl:  int64(oldTtl.(int)),
				Data: convertStringSet(oldData.(*schema.Set)),
			},
		},
		Additions: []*dns.RecordSet{
			{
				Name: name,
				Type: newType.(string),
				Ttl:  int64(newTtl.(int)),
				Data: convertStringSet(d.Get("data").(*schema.Set)),
			},
		},
	}

	return req, nil
}

func makeDnsRecordSetUpdateRequest(req *dns.UpdateRecordSetsRequest, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	sdk := getSDK(config)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := sdk.WrapOperation(sdk.DNS().DnsZone().UpdateRecordSets(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to update DnsRecordSet %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error updating DnsRecordSet %q: %s", d.Id(), err)
	}

	return nil
}

func resourceDnsRecordSetImportState(d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) == 3 {
		if err := d.Set("zone_id", parts[0]); err != nil {
			return nil, fmt.Errorf("Error setting zone_id: %s", err)
		}
		if err := d.Set("name", parts[1]); err != nil {
			return nil, fmt.Errorf("Error setting name: %s", err)
		}
		if err := d.Set("type", parts[2]); err != nil {
			return nil, fmt.Errorf("Error setting type: %s", err)
		}
	} else {
		return nil, fmt.Errorf("Invalid dns recordset specifier. Expecting {zone-id}/{record-name}/{record-type}.")
	}

	return []*schema.ResourceData{d}, nil
}

func rsId(d *schema.ResourceData) string {
	return fmt.Sprintf("%s %s", d.Get("type").(string), d.Get("name"))
}

func dataDiffSuppressFunc(_, _, _ string, d *schema.ResourceData) bool {
	if strings.ToUpper(d.Get("type").(string)) != "AAAA" {
		return false
	}
	o, n := d.GetChange("data")
	if o == nil || n == nil {
		return false
	}

	oldList := convertStringSet(o.(*schema.Set))
	newList := convertStringSet(n.(*schema.Set))

	if len(oldList) != len(newList) {
		return false
	}

	for i, oldIp := range oldList {
		log.Printf("compare %s and %s", oldIp, newList[i])
		if !ipv6Equal(oldIp, newList[i]) {
			return false
		}
	}

	return true
}

func ipv6Equal(old, new string) bool {
	ip1 := net.ParseIP(old)
	ip2 := net.ParseIP(new)
	return ip1.Equal(ip2)
}
