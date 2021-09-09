package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/dns/v1"
)

func dataSourceYandexDnsZone() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexDnsZoneRead,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexDnsDefaultTimeout),
			Update: schema.DefaultTimeout(yandexDnsDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexDnsDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"dns_zone_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"zone": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"labels": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"public": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"private_networks": {
				Type:     schema.TypeSet,
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

func dataSourceYandexDnsZoneRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	sdk := getSDK(config)

	id := d.Get("dns_zone_id").(string)
	if id == "" {
		return fmt.Errorf("dns_zone_id should be provided")
	}

	dnsZone, err := sdk.DNS().DnsZone().Get(config.Context(), &dns.GetDnsZoneRequest{
		DnsZoneId: id,
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("DnsZone %q", d.Get("name").(string)))
	}

	d.Set("created_at", getTimestamp(dnsZone.CreatedAt))
	d.Set("name", dnsZone.Name)
	d.Set("folder_id", dnsZone.FolderId)
	d.Set("zone", dnsZone.Zone)
	d.Set("description", dnsZone.Description)

	d.Set("public", dnsZone.PublicVisibility != nil)
	d.SetId(dnsZone.Id)

	if dnsZone.PrivateVisibility != nil {
		if err := d.Set("private_networks", convertStringArrToInterface(dnsZone.PrivateVisibility.GetNetworkIds())); err != nil {
			return err
		}
	}
	return d.Set("labels", dnsZone.Labels)
}
