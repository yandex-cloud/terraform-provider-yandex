package yandex

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

func dataSourceYandexVPCAddress() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexVPCAddressRead,
		Schema: map[string]*schema.Schema{
			"address_id": {
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
			"external_ipv4_address": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"zone_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ddos_protection_provider": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"outgoing_smtp_capability": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"reserved": {
				Type:     schema.TypeBool,
				Computed: true,
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
				Computed: true,
			},
			"dns_record": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"dns_zone_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"fqdn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ttl": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"ptr": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceYandexVPCAddressRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	err := checkOneOf(d, "address_id", "name")
	if err != nil {
		return err
	}

	addressID := d.Get("address_id").(string)
	_, nameOk := d.GetOk("name")

	if nameOk {
		addressID, err = resolveObjectID(config.Context(), config, d, sdkresolvers.AddressResolver)
		if err != nil {
			return addressError("failed to resolve data source address by name: %v", err)
		}
	}

	if err := yandexVPCAddressRead(d, meta, addressID); err != nil {
		return err
	}

	d.SetId(addressID)

	return d.Set("address_id", addressID)
}
