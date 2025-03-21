package yandex

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1"

	"google.golang.org/genproto/protobuf/field_mask"
)

const yandexOrganizationManagerOsLoginSettingsDefaultTimeout = 1 * time.Minute

func resourceYandexOrganizationManagerOsLoginSettings() *schema.Resource {
	return &schema.Resource{
		Description: "Allows management of OsLogin Settings within an existing Yandex Cloud Organization.",

		CreateContext: resourceYandexOrganizationManagerOsLoginSettingsCreate,
		ReadContext:   resourceYandexOrganizationManagerOsLoginSettingsRead,
		UpdateContext: resourceYandexOrganizationManagerOsLoginSettingsUpdate,
		DeleteContext: resourceYandexOrganizationManagerOsLoginSettingsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexOrganizationManagerOsLoginSettingsDefaultTimeout),
			Read:   schema.DefaultTimeout(yandexOrganizationManagerOsLoginSettingsDefaultTimeout),
			Update: schema.DefaultTimeout(yandexOrganizationManagerOsLoginSettingsDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexOrganizationManagerOsLoginSettingsDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"organization_id": {
				Type:         schema.TypeString,
				Description:  "The organization to manage it's OsLogin Settings.",
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 50),
			},
			"user_ssh_key_settings": {
				Type:        schema.TypeList,
				Description: "Users SSH key settings.",
				Optional:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:        schema.TypeBool,
							Description: "Enables or disables usage of ssh keys assigned to a specific subject.",
							Optional:    true,
						},
						"allow_manage_own_keys": {
							Type:        schema.TypeBool,
							Description: "If set to true subject is allowed to manage own ssh keys without having to be assigned specific permissions.",
							Optional:    true,
						},
					},
				},
			},
			"ssh_certificate_settings": {
				Type:        schema.TypeList,
				Description: "SSH Certificate settings.",
				Optional:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:        schema.TypeBool,
							Description: "Enables or disables usage of SSH certificates signed by trusted Certification Authority (CA).",
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

func resourceYandexOrganizationManagerOsLoginSettingsCreate(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceYandexOrganizationManagerOsLoginSettingsUpdate(context, d, meta)
}

func resourceYandexOrganizationManagerOsLoginSettingsRead(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return diag.FromErr(flattenOsLoginSettings(context, d, meta))
}

var updateOsLoginSettingsFieldsMap = map[string]string{
	"user_ssh_key_settings":    "user_ssh_key_settings",
	"ssh_certificate_settings": "ssh_certificate_settings",
}

func resourceYandexOrganizationManagerOsLoginSettingsUpdate(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	organizationID := d.Get("organization_id").(string)
	req := &organizationmanager.UpdateOsLoginSettingsRequest{
		OrganizationId: organizationID,
		UpdateMask:     &field_mask.FieldMask{},
	}

	if _, ok := d.GetOk("user_ssh_key_settings"); ok {
		userSSHKeySettings, err := expandUserSshKeySettings(d)
		if err != nil {
			return diag.FromErr(err)
		}

		req.SetUserSshKeySettings(userSSHKeySettings)
	}

	if _, ok := d.GetOk("ssh_certificate_settings"); ok {
		sshCertificateSettings, err := expandSshCertificateSettings(d)
		if err != nil {
			return diag.FromErr(err)
		}

		req.SetSshCertificateSettings(sshCertificateSettings)
	}

	var updatePath []string
	for field, path := range updateOsLoginSettingsFieldsMap {
		if d.HasChange(field) {
			updatePath = append(updatePath, path)
		}
	}

	req.UpdateMask = &field_mask.FieldMask{Paths: updatePath}
	if len(req.UpdateMask.Paths) == 0 {
		return diag.Errorf("No fields were updated for OsLoginSettings %s", organizationID)
	}

	config := meta.(*Config)
	op, err := config.sdk.WrapOperation(config.sdk.OrganizationManager().OsLogin().UpdateSettings(context, req))
	if err != nil {
		return diag.Errorf("Error while requesting API to update OsLoginSettings %q: %s", organizationID, err)
	}

	err = op.Wait(context)
	if err != nil {
		return diag.Errorf("Error updating OsLoginSettings %q: %s", organizationID, err)
	}

	d.SetId(organizationID)

	return resourceYandexOrganizationManagerOsLoginSettingsRead(context, d, meta)
}

func resourceYandexOrganizationManagerOsLoginSettingsDelete(context.Context, *schema.ResourceData, interface{}) diag.Diagnostics {
	return nil
}
