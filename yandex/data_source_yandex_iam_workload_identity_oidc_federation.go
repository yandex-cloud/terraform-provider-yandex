package yandex

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1/workload/oidc"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

func dataSourceYandexIAMWorkloadIdentityOidcFederation() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceYandexIAMWorkloadIdentityOidcFederationRead,

		Description: "Get information about a [Yandex Cloud IAM workload identity OIDC federation](https://yandex.cloud/docs/iam/concepts/workload-identity).",

		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"federation_id": {
				Type:          schema.TypeString,
				Computed:      true,
				Optional:      true,
				ConflictsWith: []string{"name"},
				Description:   "Id of the OIDC workload identity federation.",
			},

			"folder_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Id of the folder that the OIDC workload identity federation belongs to.",
			},

			"name": {
				Type:          schema.TypeString,
				Computed:      true,
				Optional:      true,
				ConflictsWith: []string{"federation_id"},
				Description:   "Name of the OIDC workload identity federation. The name is unique within the folder.",
			},

			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of the OIDC workload identity federation.",
			},

			"enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Enabled flag.",
			},

			"audiences": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed:    true,
				Description: "List of trusted values for aud claim.",
			},

			"issuer": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Issuer identifier of the external IdP server to be used for authentication.",
			},

			"jwks_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "URL reference to trusted keys in format of JSON Web Key Set.",
			},

			"labels": {
				Type: schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Set:         schema.HashString,
				Computed:    true,
				Description: "Resource labels as key-value pairs.",
			},

			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation timestamp.",
			},
		},
	}
}

func dataSourceYandexIAMWorkloadIdentityOidcFederationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	err := checkOneOf(d, "federation_id", "name")
	if err != nil {
		return diag.FromErr(err)
	}

	federationID := d.Get("federation_id").(string)
	_, federationNameOk := d.GetOk("name")

	if federationNameOk {
		federationID, err = resolveObjectID(ctx, config, d, sdkresolvers.WliFederationResolver)
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to resolve workload identity federation by name: %v", err))
		}
	}

	req := &oidc.GetFederationRequest{
		FederationId: federationID,
	}

	resp, err := config.sdk.WorkloadOidc().Federation().Get(ctx, req)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resp.Id)

	if err := d.Set("federation_id", resp.GetId()); err != nil {
		log.Printf("[ERROR] failed set field federation_id: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("folder_id", resp.GetFolderId()); err != nil {
		log.Printf("[ERROR] failed set field folder_id: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("name", resp.GetName()); err != nil {
		log.Printf("[ERROR] failed set field name: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("description", resp.GetDescription()); err != nil {
		log.Printf("[ERROR] failed set field description: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("enabled", resp.GetEnabled()); err != nil {
		log.Printf("[ERROR] failed set field enabled: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("audiences", resp.GetAudiences()); err != nil {
		log.Printf("[ERROR] failed set field audiences: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("issuer", resp.GetIssuer()); err != nil {
		log.Printf("[ERROR] failed set field issuer: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("jwks_url", resp.GetJwksUrl()); err != nil {
		log.Printf("[ERROR] failed set field jwks_url: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("labels", resp.GetLabels()); err != nil {
		log.Printf("[ERROR] failed set field labels: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("created_at", getTimestamp(resp.GetCreatedAt())); err != nil {
		log.Printf("[ERROR] failed set field created_at: %s", err)
		return diag.FromErr(err)
	}

	return nil
}
