package yandex

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/dns/v1"
)

// canonicalizeTXTRecordValue joins a TXT value that the DNS API may return as
// several quoted character-strings (RFC 1035 caps each at 255 bytes) into a
// single quoted string, so it compares equal to the single-string form used in
// configuration. Escapes are preserved and malformed input is returned as-is.
func canonicalizeTXTRecordValue(s string) string {
	if len(s) == 0 || s[0] != '"' {
		return s
	}
	var joined []byte
	i := 1
	for {
		closed := false
		for i < len(s) {
			b := s[i]
			if b == '\\' {
				if i+1 >= len(s) {
					return s
				}
				joined = append(joined, b, s[i+1])
				i += 2
			} else if b == '"' {
				i++
				closed = true
				break
			} else {
				joined = append(joined, b)
				i++
			}
			if i == len(s) {
				return s
			}
		}
		if !closed {
			return s
		}
		for i < len(s) && (s[i] == ' ' || s[i] == '\t' || s[i] == '\r' || s[i] == '\n') {
			i++
		}
		if i == len(s) {
			break
		}
		if s[i] != '"' {
			return s
		}
		i++
	}
	return `"` + string(joined) + `"`
}

func dnsRecordSetDataHash(v interface{}) int {
	return schema.HashString(canonicalizeTXTRecordValue(v.(string)))
}

func resourceYandexDnsRecordSet() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a DNS RecordSet within Yandex Cloud.",
		Create:      resourceYandexDnsRecordSetCreate,
		Read:        resourceYandexDnsRecordSetRead,
		Update:      resourceYandexDnsRecordSetUpdate,
		Delete:      resourceYandexDnsRecordSetDelete,
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
				Type:        schema.TypeString,
				Description: "The id of the zone in which this record set will reside.",
				Required:    true,
				ForceNew:    true,
			},

			"name": {
				Type:         schema.TypeString,
				Description:  "The DNS name this record set will apply to.",
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 254),
			},

			"description": {
				Type:        schema.TypeString,
				Description: "The DNS record set description.",
				Optional:    true,
			},

			"type": {
				Type:         schema.TypeString,
				Description:  "The DNS record set type.",
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 20),
			},

			"ttl": {
				Type:         schema.TypeInt,
				Description:  "The time-to-live of this record set (seconds).",
				Required:     true,
				ValidateFunc: validation.IntBetween(0, 2147483647),
			},

			"data": {
				Type:        schema.TypeSet,
				Description: "The string data for the records in this record set.",
				Required:    true,
				MinItems:    1,
				MaxItems:    100,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(1, 1024),
				},
				Set: dnsRecordSetDataHash,
			},
		},
	}
}

func resourceYandexDnsRecordSetCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	sdk := getSDK(config)

	rs := &dns.RecordSet{
		Name:        d.Get("name").(string),
		Type:        d.Get("type").(string),
		Description: d.Get("description").(string),
		Ttl:         int64(d.Get("ttl").(int)),
		Data:        convertStringSet(d.Get("data").(*schema.Set)),
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

	d.Set("description", rs.Description)
	d.Set("ttl", int(rs.Ttl))
	data := rs.Data
	if strings.EqualFold(rs.Type, "TXT") {
		normalized := make([]string, len(rs.Data))
		for i, v := range rs.Data {
			normalized[i] = canonicalizeTXTRecordValue(v)
		}
		data = normalized
	}
	d.Set("data", convertStringArrToInterface(data))

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
		Name:        d.Get("name").(string),
		Type:        d.Get("type").(string),
		Description: d.Get("description").(string),
		Ttl:         int64(d.Get("ttl").(int)),
		Data:        convertStringSet(d.Get("data").(*schema.Set)),
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
	oldDescription, newDescription := d.GetChange("description")
	oldData, _ := d.GetChange("data")

	req := &dns.UpdateRecordSetsRequest{
		DnsZoneId: d.Get("zone_id").(string),
		Deletions: []*dns.RecordSet{
			{
				Name:        name,
				Type:        oldType.(string),
				Description: oldDescription.(string),
				Ttl:         int64(oldTtl.(int)),
				Data:        convertStringSet(oldData.(*schema.Set)),
			},
		},
		Additions: []*dns.RecordSet{
			{
				Name:        name,
				Type:        newType.(string),
				Description: newDescription.(string),
				Ttl:         int64(newTtl.(int)),
				Data:        convertStringSet(d.Get("data").(*schema.Set)),
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
