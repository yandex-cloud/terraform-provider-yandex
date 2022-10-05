package yandex

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/lockbox/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

func dataSourceYandexLockboxSecretVersion() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceYandexLockboxSecretVersionRead,

		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"secret_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 50),
			},

			"entries": {
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"text_value": {
							Type:      schema.TypeString,
							Computed:  true,
							Sensitive: true,
						},
					},
				},
				Computed: true,
			},

			"version_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func dataSourceYandexLockboxSecretVersionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	id := d.Get("version_id").(string)
	req := &lockbox.GetPayloadRequest{
		SecretId:  d.Get("secret_id").(string),
		VersionId: id,
	}

	log.Printf("[INFO] reading Lockbox version: %s", protojson.Format(req))

	payload, err := config.sdk.LockboxPayload().Payload().Get(ctx, req)
	if err != nil {
		return diag.FromErr(handleNotFoundError(err, d, fmt.Sprintf("secret version payload %q", id)))
	}

	d.SetId(payload.VersionId)

	entries, err := flattenLockboxSecretVersionEntriesSlice(payload.GetEntries())
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("entries", entries); err != nil {
		log.Printf("[ERROR] failed set field entries: %s", err)
		return diag.FromErr(err)
	}

	log.Printf("[INFO] read Lockbox version with ID: %s", id)

	return diag.FromErr(err)
}
