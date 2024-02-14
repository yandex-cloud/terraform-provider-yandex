package yandex

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceYandexOrganizationManagerOsLoginSettings() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceYandexOrganizationManagerOsLoginSettingsRead,
		Schema: map[string]*schema.Schema{
			"organization_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"user_ssh_key_settings": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"allow_manage_own_keys": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
			"ssh_certificate_settings": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceYandexOrganizationManagerOsLoginSettingsRead(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return diag.FromErr(flattenOsLoginSettings(context, d, meta))
}
