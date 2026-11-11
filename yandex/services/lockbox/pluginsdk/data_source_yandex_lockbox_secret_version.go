package pluginsdk

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/lockbox/v1"
	lockboxsdk "github.com/yandex-cloud/go-sdk/services/lockbox/v1"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	"google.golang.org/protobuf/encoding/protojson"
)

func DataSourceYandexLockboxSecretVersion() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about Yandex Cloud Lockbox secret version. For more information, see [the official documentation](https://yandex.cloud/docs/lockbox/).\nIf you're creating the secret in the same project, then you should indicate `version_id`, since otherwise you may refer to a wrong version of the secret (e.g. the first version, when it is still empty).\n",

		ReadContext: dataSourceYandexLockboxSecretVersionRead,

		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"secret_id": {
				Type:         schema.TypeString,
				Description:  ResourceYandexLockboxSecretVersion().Schema["secret_id"].Description,
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
				Type:        schema.TypeString,
				Description: "The Yandex Cloud Lockbox secret version ID.",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func dataSourceYandexLockboxSecretVersionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*providerconfig.Config)

	id := d.Get("version_id").(string)
	req := &lockbox.GetPayloadRequest{
		SecretId:  d.Get("secret_id").(string),
		VersionId: id,
	}

	log.Printf("[INFO] reading Lockbox version: %s", protojson.Format(req))

	client := lockboxsdk.NewPayloadClient(config.SDK)

	payload, err := client.Get(ctx, req)
	if err != nil {
		return diag.FromErr(sdkcommon.HandleNotFoundError(err, d, fmt.Sprintf("secret version payload %q", id)))
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
