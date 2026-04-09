package yandex

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/lockbox/v1"
)

func dataSourceYandexLockboxSecretVersionEntry() *schema.Resource {
	return &schema.Resource{
		Description: "Get a single entry from a Yandex Cloud Lockbox secret version by key. For more information, see [the official documentation](https://yandex.cloud/docs/lockbox/).\n\nThis data source is a convenience wrapper around `yandex_lockbox_secret_version` that lets you retrieve a specific entry by key without having to filter the `entries` list yourself.\n\nIf you're creating the secret in the same project, then you should indicate `version_id`, since otherwise you may refer to a wrong version of the secret (e.g. the first version, when it is still empty).\n",

		ReadContext: dataSourceYandexLockboxSecretVersionEntryRead,

		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"secret_id": {
				Type:         schema.TypeString,
				Description:  "The Yandex Cloud Lockbox secret ID.",
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 50),
			},

			"version_id": {
				Type:        schema.TypeString,
				Description: "The Yandex Cloud Lockbox secret version ID. If omitted, the current (latest) version is used.",
				Optional:    true,
				Computed:    true,
			},

			"key": {
				Type:        schema.TypeString,
				Description: "The key of the entry to retrieve.",
				Required:    true,
			},

			"text_value": {
				Type:        schema.TypeString,
				Description: "The text value of the entry. Populated when the entry holds a UTF-8 string. Mutually exclusive with `binary_value`.",
				Computed:    true,
				Sensitive:   true,
			},

			"binary_value": {
				Type:        schema.TypeString,
				Description: "The binary value of the entry encoded as a base64 string. Populated when the entry holds binary data. Mutually exclusive with `text_value`.",
				Computed:    true,
				Sensitive:   true,
			},
		},
	}
}

func dataSourceYandexLockboxSecretVersionEntryRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	config := meta.(*Config)

	versionID := d.Get("version_id").(string)
	secretID := d.Get("secret_id").(string)
	key := d.Get("key").(string)

	req := &lockbox.GetPayloadRequest{
		SecretId:  secretID,
		VersionId: versionID,
	}

	payload, err := config.sdk.LockboxPayload().Payload().Get(ctx, req)
	if err != nil {
		return diag.FromErr(handleNotFoundError(err, d, fmt.Sprintf("secret version payload %q", versionID)))
	}

	var entry *lockbox.Payload_Entry
	for _, e := range payload.GetEntries() {
		if e.Key == key {
			entry = e
			break
		}
	}

	if entry == nil {
		return diag.Errorf("entry with key %q not found in secret %q version %q", key, secretID, payload.VersionId)
	}

	d.SetId(fmt.Sprintf("%s/%s/%s", secretID, payload.VersionId, key))

	if err := d.Set("version_id", payload.VersionId); err != nil {
		return diag.FromErr(err)
	}

	switch v := entry.GetValue().(type) {
	case *lockbox.Payload_Entry_TextValue:
		if err := d.Set("text_value", v.TextValue); err != nil {
			return diag.FromErr(err)
		}
	case *lockbox.Payload_Entry_BinaryValue:
		if err := d.Set("binary_value", base64.StdEncoding.EncodeToString(v.BinaryValue)); err != nil {
			return diag.FromErr(err)
		}
	default:
		return diag.Errorf("entry with key %q has unknown value type %T", key, entry.GetValue())
	}

	return nil
}
