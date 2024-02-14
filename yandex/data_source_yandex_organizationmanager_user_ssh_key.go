package yandex

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceYandexOrganizationManagerUserSshKey() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceYandexOrganizationManagerUserSshKeyRead,
		Schema: map[string]*schema.Schema{
			"user_ssh_key_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"subject_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"data": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"organization_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"expires_at": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceYandexOrganizationManagerUserSshKeyRead(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return diag.FromErr(flattenUserSshKey(context, d, meta))
}
