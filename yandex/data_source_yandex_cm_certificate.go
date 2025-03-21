package yandex

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

func dataSourceYandexCMCertificate() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about a Yandex Certificate Manager Certificate. For more information, see [the official documentation](https://yandex.cloud/docs/certificate-manager/concepts/).\n\n~> One of `certificate_id` or `name` should be specified.\n",
		ReadContext: dataSourceYandexCMCertificateRead,

		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"certificate_id": {
				Type:        schema.TypeString,
				Description: "Certificate Id.",
				Optional:    true,
				Computed:    true,
			},
			"folder_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["folder_id"],
				Computed:    true,
				Optional:    true,
				ForceNew:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Optional:    true,
				Computed:    true,
			},
			"wait_validation": {
				Type:        schema.TypeBool,
				Description: "If `true`, the operation won't be completed while the certificate is in `VALIDATING`.",
				Optional:    true,
				Default:     false,
			},
			"description": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["description"],
				Computed:    true,
				Optional:    true,
			},
			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},
			"labels": {
				Type:        schema.TypeMap,
				Description: common.ResourceDescriptions["labels"],
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Optional:    true,
				Computed:    true,
			},
			"deletion_protection": {
				Type:        schema.TypeBool,
				Description: common.ResourceDescriptions["deletion_protection"],
				Computed:    true,
			},
			"type": {
				Type:        schema.TypeString,
				Description: resourceYandexCMCertificate().Schema["type"].Description,
				Computed:    true,
			},
			"domains": {
				Type:        schema.TypeList,
				Description: resourceYandexCMCertificate().Schema["domains"].Description,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
			},
			"status": {
				Type:        schema.TypeString,
				Description: resourceYandexCMCertificate().Schema["status"].Description,
				Computed:    true,
			},
			"issuer": {
				Type:        schema.TypeString,
				Description: resourceYandexCMCertificate().Schema["issuer"].Description,
				Computed:    true,
			},
			"subject": {
				Type:        schema.TypeString,
				Description: resourceYandexCMCertificate().Schema["subject"].Description,
				Computed:    true,
			},
			"serial": {
				Type:        schema.TypeString,
				Description: resourceYandexCMCertificate().Schema["serial"].Description,
				Computed:    true,
			},
			"updated_at": {
				Type:        schema.TypeString,
				Description: resourceYandexCMCertificate().Schema["updated_at"].Description,
				Computed:    true,
			},
			"issued_at": {
				Type:        schema.TypeString,
				Description: resourceYandexCMCertificate().Schema["issued_at"].Description,
				Computed:    true,
			},
			"not_after": {
				Type:        schema.TypeString,
				Description: resourceYandexCMCertificate().Schema["not_after"].Description,
				Computed:    true,
			},
			"not_before": {
				Type:        schema.TypeString,
				Description: resourceYandexCMCertificate().Schema["not_before"].Description,
				Computed:    true,
			},
			"challenges": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"domain": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"created_at": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"updated_at": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"message": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"dns_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"dns_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"dns_value": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"http_url": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"http_content": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceYandexCMCertificateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	err := checkOneOf(d, "certificate_id", "name")
	if err != nil {
		return diag.FromErr(err)
	}
	id := d.Get("certificate_id").(string)

	_, certificateNameOk := d.GetOk("name")
	if certificateNameOk {
		id, err = resolveObjectID(config.Context(), config, d, sdkresolvers.CertificateResolver)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	return yandexCMCertificateRead(
		id,
		ctx,
		d,
		meta,
		true,
	)
}
