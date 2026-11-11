package pluginsdk

import (
	"context"
	"fmt"

	organizationmanagersdk "github.com/yandex-cloud/go-sdk/services/organizationmanager/v1"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1"

	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
)

func flattenUserSshKeySettings(v *organizationmanager.UserSshKeySettings) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["enabled"] = v.Enabled
	m["allow_manage_own_keys"] = v.AllowManageOwnKeys

	return []map[string]interface{}{m}, nil
}

func flattenSshCertificateSettings(v *organizationmanager.SshCertificateSettings) ([]map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}

	m := make(map[string]interface{})

	m["enabled"] = v.Enabled

	return []map[string]interface{}{m}, nil
}

func expandUserSshKeySettings(d *schema.ResourceData) (*organizationmanager.UpdateOsLoginSettingsRequest_UserSshKeySettings, error) {
	val := new(organizationmanager.UpdateOsLoginSettingsRequest_UserSshKeySettings)

	if v, ok := d.GetOk("user_ssh_key_settings.0.enabled"); ok {
		val.SetEnabled(v.(bool))
	}

	if v, ok := d.GetOk("user_ssh_key_settings.0.allow_manage_own_keys"); ok {
		val.SetAllowManageOwnKeys(v.(bool))
	}

	return val, nil
}

func expandSshCertificateSettings(d *schema.ResourceData) (*organizationmanager.UpdateOsLoginSettingsRequest_SshCertificateSettings, error) {
	val := new(organizationmanager.UpdateOsLoginSettingsRequest_SshCertificateSettings)

	if v, ok := d.GetOk("ssh_certificate_settings.0.enabled"); ok {
		val.SetEnabled(v.(bool))
	}

	return val, nil
}

func flattenOsLoginSettings(context context.Context, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*providerconfig.Config)
	client := organizationmanagersdk.NewOsLoginClient(config.SDK)

	organizationID := d.Get("organization_id").(string)

	osLoginSettings, err := client.GetSettings(context,
		&organizationmanager.GetOsLoginSettingsRequest{
			OrganizationId: organizationID,
		})

	if err != nil {
		return sdkcommon.HandleNotFoundError(err, d, fmt.Sprintf("OsLoginSettings %q", organizationID))
	}

	userSSHKeySettings, err := flattenUserSshKeySettings(osLoginSettings.UserSshKeySettings)
	if err != nil {
		return err
	}

	sshCertificateSettings, err := flattenSshCertificateSettings(osLoginSettings.SshCertificateSettings)
	if err != nil {
		return err
	}

	if err := d.Set("user_ssh_key_settings", userSSHKeySettings); err != nil {
		return err
	}
	d.SetId(organizationID)

	return d.Set("ssh_certificate_settings", sshCertificateSettings)
}
