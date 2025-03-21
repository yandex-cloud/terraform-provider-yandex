package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

func dataSourceYandexOrganizationManagerSamlFederation() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about a Yandex SAML Federation. For more information, see [the official documentation](https://yandex.cloud/docs/organization/add-federation).\n\n~> One of `federation_id` or `name` should be specified.\n",

		Read: dataSourceYandexOrganizationManagerSamlFederationRead,
		Schema: map[string]*schema.Schema{
			"federation_id": {
				Type:        schema.TypeString,
				Description: "ID of a SAML Federation.",
				Optional:    true,
				Computed:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Optional:    true,
				Computed:    true,
			},
			"organization_id": {
				Type:        schema.TypeString,
				Description: "Organization that the federation belongs to. If value is omitted, the default provider organization is used.",
				Optional:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["description"],
				Computed:    true,
			},
			"labels": {
				Type:        schema.TypeMap,
				Description: common.ResourceDescriptions["labels"],
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},
			"issuer": {
				Type:        schema.TypeString,
				Description: resourceYandexOrganizationManagerSamlFederation().Schema["issuer"].Description,
				Computed:    true,
			},
			"sso_binding": {
				Type:        schema.TypeString,
				Description: resourceYandexOrganizationManagerSamlFederation().Schema["sso_binding"].Description,
				Computed:    true,
			},
			"sso_url": {
				Type:        schema.TypeString,
				Description: resourceYandexOrganizationManagerSamlFederation().Schema["sso_url"].Description,
				Computed:    true,
			},
			"cookie_max_age": {
				Type:        schema.TypeString,
				Description: resourceYandexOrganizationManagerSamlFederation().Schema["cookie_max_age"].Description,
				Computed:    true,
			},
			"auto_create_account_on_login": {
				Type:        schema.TypeBool,
				Description: resourceYandexOrganizationManagerSamlFederation().Schema["auto_create_account_on_login"].Description,
				Computed:    true,
			},
			"case_insensitive_name_ids": {
				Type:        schema.TypeBool,
				Description: resourceYandexOrganizationManagerSamlFederation().Schema["case_insensitive_name_ids"].Description,
				Computed:    true,
			},
			"security_settings": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"encrypted_assertions": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"force_authn": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},
		},
	}
}

func dataSourceYandexOrganizationManagerSamlFederationRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := config.Context()

	err := checkOneOf(d, "federation_id", "name")
	if err != nil {
		return err
	}

	organizationID, err := getOrganizationID(d, config)
	if err != nil {
		return fmt.Errorf("Error getting organization ID while reading SAML Federation: %s", err)
	}

	federationID := d.Get("federation_id").(string)
	federationName, ok := d.GetOk("name")

	if ok {
		federationID, err = resolveFederationIDByName(ctx, config, federationName.(string), organizationID)
		if err != nil {
			return fmt.Errorf("failed to resolve data source SAML Federation by name: %v", err)
		}
	}

	err = flattenSamlFederation(federationID, d, meta)
	if err != nil {
		return err
	}
	d.SetId(federationID)
	return nil
}

func resolveFederationIDByName(ctx context.Context, config *Config, federationName, organizationID string) (string, error) {
	var objectID string
	resolver := sdkresolvers.OrganizationSamlFederationResolver(federationName, sdkresolvers.OrganizationID(organizationID), sdkresolvers.Out(&objectID))

	err := config.sdk.Resolve(ctx, resolver)
	if err != nil {
		return "", err
	}

	return objectID, nil
}
