package yandex

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

func dataSourceYandexOrganizationManagerUserSshKey() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about a Yandex Cloud User SSH Key.",

		ReadContext: dataSourceYandexOrganizationManagerUserSshKeyRead,
		Schema: map[string]*schema.Schema{
			"user_ssh_key_id": {
				Type:        schema.TypeString,
				Description: "ID of the user ssh key.",
				Required:    true,
			},
			"subject_id": {
				Type:        schema.TypeString,
				Description: resourceYandexOrganizationManagerUserSshKey().Schema["subject_id"].Description,
				Optional:    true,
			},
			"data": {
				Type:        schema.TypeString,
				Description: resourceYandexOrganizationManagerUserSshKey().Schema["data"].Description,
				Optional:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Optional:    true,
			},
			"fingerprint": {
				Type:        schema.TypeString,
				Description: resourceYandexOrganizationManagerUserSshKey().Schema["fingerprint"].Description,
				Computed:    true,
			},
			"organization_id": {
				Type:        schema.TypeString,
				Description: resourceYandexOrganizationManagerUserSshKey().Schema["organization_id"].Description,
				Optional:    true,
			},
			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},
			"expires_at": {
				Type:        schema.TypeString,
				Description: resourceYandexOrganizationManagerUserSshKey().Schema["expires_at"].Description,
				Optional:    true,
			},
		},
	}
}

func dataSourceYandexOrganizationManagerUserSshKeyRead(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return diag.FromErr(flattenUserSshKey(context, d, meta))
}
