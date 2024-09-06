package yandex

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1/workload"
)

func dataSourceYandexIAMWorkloadIdentityFederatedCredential() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceYandexIAMWorkloadIdentityFederatedCredentialRead,

		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"federated_credential_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 50),
			},

			"service_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"federation_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"external_subject_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceYandexIAMWorkloadIdentityFederatedCredentialRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	fcId := d.Get("federated_credential_id").(string)

	req := &workload.GetFederatedCredentialRequest{
		FederatedCredentialId: fcId,
	}

	resp, err := config.sdk.Workload().FederatedCredential().Get(ctx, req)

	if err != nil {
		diag.FromErr(err)
	}

	d.SetId(resp.Id)

	if err := d.Set("federated_credential_id", resp.GetId()); err != nil {
		log.Printf("[ERROR] failed set field federated_credential_id: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("service_account_id", resp.GetServiceAccountId()); err != nil {
		log.Printf("[ERROR] failed set field service_account_id: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("federation_id", resp.GetFederationId()); err != nil {
		log.Printf("[ERROR] failed set field federation_id: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("external_subject_id", resp.GetExternalSubjectId()); err != nil {
		log.Printf("[ERROR] failed set field external_subject_id: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("created_at", getTimestamp(resp.GetCreatedAt())); err != nil {
		log.Printf("[ERROR] failed set field created_at: %s", err)
		return diag.FromErr(err)
	}

	return nil
}
