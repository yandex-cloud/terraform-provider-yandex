package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

func dataSourceYandexLockboxSecret() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceYandexLockboxSecretRead,

		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"secret_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 50),
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"current_version": {
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"created_at": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"destroy_at": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"payload_entry_keys": {
							Type: schema.TypeList,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Computed: true,
						},

						"secret_id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
				Computed: true,
			},

			"deletion_protection": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"folder_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"kms_key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"labels": {
				Type: schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Set:      schema.HashString,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
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
