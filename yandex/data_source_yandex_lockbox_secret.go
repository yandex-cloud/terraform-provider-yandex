package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

func dataSourceYandexLockboxSecret() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about Yandex Cloud Lockbox secret. For more information, see [the official documentation](https://yandex.cloud/docs/lockbox/).\n\n~> One of `secret_id` or `name` should be specified.\n",

		ReadContext: dataSourceYandexLockboxSecretRead,

		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"secret_id": {
				Type:         schema.TypeString,
				Description:  "The Yandex Cloud Lockbox secret ID.",
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 50),
			},
			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},

			"current_version": {
				Type:        schema.TypeList,
				Description: "Current secret version.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"created_at": {
							Type:        schema.TypeString,
							Description: "The version creation timestamp.",
							Computed:    true,
						},

						"description": {
							Type:        schema.TypeString,
							Description: "The version description.",
							Computed:    true,
						},

						"destroy_at": {
							Type:        schema.TypeString,
							Description: "The version destroy timestamp.",
							Computed:    true,
						},

						"id": {
							Type:        schema.TypeString,
							Description: "The version ID.",
							Computed:    true,
						},

						"payload_entry_keys": {
							Type:        schema.TypeList,
							Description: "List of keys that the version contains (doesn't include the values).",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Computed: true,
						},

						"secret_id": {
							Type:        schema.TypeString,
							Description: "The secret ID the version belongs to (it's the same as the `secret_id` argument indicated above)",
							Computed:    true,
						},

						"status": {
							Type:        schema.TypeString,
							Description: "The version status.",
							Computed:    true,
						},
					},
				},
				Computed: true,
			},

			"deletion_protection": {
				Type:        schema.TypeBool,
				Description: common.ResourceDescriptions["deletion_protection"],
				Computed:    true,
			},

			"description": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["description"],
				Computed:    true,
			},

			"folder_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["folder_id"],
				Optional:    true,
				Computed:    true,
			},

			"kms_key_id": {
				Type:        schema.TypeString,
				Description: resourceYandexLockboxSecret().Schema["kms_key_id"].Description,
				Computed:    true,
			},

			"labels": {
				Type:        schema.TypeMap,
				Description: common.ResourceDescriptions["labels"],
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Set:      schema.HashString,
				Computed: true,
			},

			"name": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Optional:    true,
				Computed:    true,
			},

			"status": {
				Type:        schema.TypeString,
				Description: resourceYandexLockboxSecret().Schema["status"].Description,
				Computed:    true,
			},

			"password_payload_specification": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"password_key": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"length": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"include_uppercase": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"include_lowercase": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"include_digits": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"include_punctuation": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"included_punctuation": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"excluded_punctuation": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceYandexLockboxSecretRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	err := checkOneOf(d, "secret_id", "name")
	if err != nil {
		return diag.FromErr(err)
	}
	secretId := d.Get("secret_id").(string)
	_, secretNameOk := d.GetOk("name")

	if secretNameOk {
		secretId, err = resolveSecretIDByName(ctx, d, config)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	return yandexLockboxSecretRead(secretId, true, ctx, d, meta)
}

func resolveSecretIDByName(ctx context.Context, d *schema.ResourceData, config *Config) (string, error) {
	secretId, err := resolveObjectID(ctx, config, d, sdkresolvers.SecretResolver)
	if err != nil {
		return "", fmt.Errorf("failed to resolve secret by name: %v ", err)
	}
	if err := d.Set("secret_id", secretId); err != nil {
		return "", fmt.Errorf("failed to set field 'secret_id': %v ", err)
	}
	return secretId, nil
}
