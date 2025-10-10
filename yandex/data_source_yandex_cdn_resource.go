package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/cdn/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

func dataSourceYandexCDNResource() *schema.Resource {
	resourceSchema := dataSourceYandexCDNResourceSchema()

	// TODO: SA1019: resourceSchema.Read is deprecated: Use ReadContext or ReadWithoutTimeout instead. This implementation does not support request cancellation initiated by Terraform, such as a system or practitioner sending SIGINT (Ctrl-c). This implementation also does not support warning diagnostics. (staticcheck)
	resourceSchema.Read = dataSourceYandexCDNResourceRead

	return resourceSchema
}

func dataSourceYandexCDNResourceSchema() *schema.Resource {
	return &schema.Resource{
		Description: "Allows management of [Yandex Cloud CDN Resource](https://yandex.cloud/docs/cdn/concepts/resource).\n\n~> CDN provider must be activated prior usage of CDN resources, either via UI console or via yc cli command: `yc cdn provider activate --folder-id <folder-id> --type gcore`.",

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			// args
			"cname": {
				Type:        schema.TypeString,
				Description: "CDN endpoint CNAME, must be unique among resources.",

				Computed:     true,
				Optional:     true,
				ExactlyOneOf: []string{"cname", "resource_id"},
			},
			"resource_id": {
				Type:        schema.TypeString,
				Description: "The ID of a specific resource.",

				Computed:     true,
				Optional:     true,
				ExactlyOneOf: []string{"cname", "resource_id"},
			},
			// end args

			// computed data
			"folder_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["folder_id"],
				Computed:    true,
			},

			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},
			"updated_at": {
				Type:        schema.TypeString,
				Description: "Last update timestamp. Computed value for read and update operations.",
				Computed:    true,
			},

			"origin_protocol": {
				Type:        schema.TypeString,
				Description: "Protocol of origin resource. `http` or `https`.",
				Computed:    true,
			},
			"origin_group_id": {
				Type:        schema.TypeInt,
				Description: "The ID of a specific origin group.",
				Computed:    true,
			},
			"origin_group_name": {
				Type:        schema.TypeString,
				Description: "The name of a specific origin group.",
				Computed:    true,
			},

			"provider_cname": {
				Type:        schema.TypeString,
				Description: "Provider CNAME of CDN resource, computed value for read and update operations.",
				Computed:    true,
			},
			"provider_type": {
				Type:        schema.TypeString,
				Description: "CDN provider is a content delivery service provider.",
				Computed:    true,
			},

			"active": {
				Type:        schema.TypeBool,
				Description: "Flag to create Resource either in active or disabled state. `True` - the content from CDN is available to clients.",
				Computed:    true,
			},
			"labels": {
				Type:        schema.TypeMap,
				Description: common.ResourceDescriptions["labels"],
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"secondary_hostnames": {
				Type:        schema.TypeSet,
				Description: "List of secondary hostname strings.",
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"ssl_certificate": {
				Type:        schema.TypeSet,
				Description: "SSL certificate of CDN resource.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:             schema.TypeString,
							Description:      "SSL certificate type.",
							Required:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validateCDNResourceSSLCertType),
						},
						"status": {
							Type:        schema.TypeString,
							Description: "SSL certificate status.",
							Computed:    true,
						},
						"certificate_manager_id": {
							Type:        schema.TypeString,
							Description: "Certificate Manager ID.",
							Optional:    true,
						},
					},
				},
			},
			"options": {
				Type:        schema.TypeList,
				Description: "CDN Resource settings and options to tune CDN edge behavior.",
				Computed:    true,
				Optional:    true,
				MaxItems:    1,

				Elem: resourceYandexCDNResourceSchema_Options(),
			},
		},
	}
}

func resolveCDNResourceID(ctx context.Context, config *Config, d *schema.ResourceData) (string, error) {
	cname := d.Get("cname").(string)

	if cname == "" {
		return "", fmt.Errorf("empty cname for resource")
	}

	folderID, err := getFolderID(d, config)
	if err != nil {
		return "", err
	}

	iterator := config.sdk.CDN().Resource().ResourceIterator(ctx, &cdn.ListResourcesRequest{
		FolderId: folderID,
	})

	for iterator.Next() {
		cdnResource := iterator.Value()
		if cname == cdnResource.Cname {
			return cdnResource.Id, nil
		}
	}

	return "", fmt.Errorf("resource with cname %q not found", cname)
}

func dataSourceYandexCDNResourceRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := config.Context()

	err := checkOneOf(d, "resource_id", "cname")
	if err != nil {
		return err
	}

	resourceID := d.Get("resource_id").(string)
	_, resourceCNameOk := d.GetOk("cname")

	if resourceCNameOk {
		resourceID, err = resolveCDNResourceID(ctx, config, d)
		if err != nil {
			return fmt.Errorf("failed to resolve data source cdn resource by name: %v", err)
		}
	}

	resource, err := config.sdk.CDN().Resource().Get(ctx, &cdn.GetResourceRequest{
		ResourceId: resourceID,
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("cdn resource with ID %q", resourceID))
	}

	res, err := flattenCDNResource(resource)
	if err != nil {
		return err
	}

	for k, v := range res {
		if err := d.Set(k, v); err != nil {
			return fmt.Errorf("error setting %s for DataSource CDN Resource (%s): %s", k, d.Id(), err)
		}
	}

	_ = d.Set("resource_id", resource.Id)
	d.SetId(resource.Id)

	return nil
}
