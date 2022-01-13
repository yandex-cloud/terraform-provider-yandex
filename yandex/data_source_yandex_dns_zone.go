package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/dns/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
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
				Optional: true,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"folder_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"zone": {
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

	err := checkOneOf(d, "dns_zone_id", "name")
	if err != nil {
		return err
	}

	id := d.Get("dns_zone_id").(string)
	_, zoneNameOk := d.GetOk("name")

	if zoneNameOk {
		id, err = resolveObjectID(config.Context(), config, d, sdkresolvers.DNSZoneResolver)
		if err != nil {
			return fmt.Errorf("failed to resolve data source dns zone by name: %v", err)
		}
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
