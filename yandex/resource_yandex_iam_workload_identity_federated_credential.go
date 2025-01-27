package yandex

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1/workload"
)

const yandexIAMWorkloadIdentityFederatedCredentialDefaultTimeout = 1 * time.Minute

func resourceYandexIAMWorkloadIdentityFederatedCredential() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceYandexIAMWorkloadIdentityFederatedCredentialCreate,
		ReadContext:   resourceYandexIAMWorkloadIdentityFederatedCredentialRead,
		DeleteContext: resourceYandexIAMWorkloadIdentityFederatedCredentialDelete,

		Description: "Allows management of [Yandex Cloud IAM federated credentials](https://yandex.cloud/docs/iam/concepts/workload-identity#federated-credentials).",

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexIAMWorkloadIdentityFederatedCredentialDefaultTimeout),
			Read:   schema.DefaultTimeout(yandexIAMWorkloadIdentityFederatedCredentialDefaultTimeout),
			Update: schema.DefaultTimeout(yandexIAMWorkloadIdentityFederatedCredentialDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexIAMWorkloadIdentityFederatedCredentialDefaultTimeout),
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"federation_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Id of the federated credential.",
			},

			"service_account_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				ForceNew:    true,
				Description: "Id of the service account that the federated credential belongs to.",
			},

			"external_subject_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Id of the workload identity federation which is used for authentication.",
			},

			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Id of the external subject.",
			},
		},
	}
}

func resourceYandexIAMWorkloadIdentityFederatedCredentialCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	req := &workload.CreateFederatedCredentialRequest{
		ServiceAccountId:  d.Get("service_account_id").(string),
		FederationId:      d.Get("federation_id").(string),
		ExternalSubjectId: d.Get("external_subject_id").(string),
	}

	op, err := config.sdk.WrapOperation(config.sdk.Workload().FederatedCredential().Create(ctx, req))

	if err != nil {
		return diag.Errorf("error while requesting API to create WLI federated credential: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return diag.Errorf("error while getting operation metadata of create WLI federated credential: %s", err)
	}

	md, ok := protoMetadata.(*workload.CreateFederatedCredentialMetadata)
	if !ok {
		return diag.Errorf("could not get WLI federated credential Id from create operation metadata")
	}

	d.SetId(md.FederatedCredentialId)

	err = op.Wait(ctx)
	if err != nil {
		return diag.Errorf("error while waiting operation to create WLI federated credential: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return diag.Errorf("WLI federated credential creation failed: %s", err)
	}

	return resourceYandexIAMWorkloadIdentityFederatedCredentialRead(ctx, d, meta)
}

func resourceYandexIAMWorkloadIdentityFederatedCredentialRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	req := &workload.GetFederatedCredentialRequest{
		FederatedCredentialId: d.Id(),
	}

	resp, err := config.sdk.Workload().FederatedCredential().Get(ctx, req)

	if err != nil {
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

func resourceYandexIAMWorkloadIdentityFederatedCredentialDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	req := &workload.DeleteFederatedCredentialRequest{
		FederatedCredentialId: d.Id(),
	}

	op, err := config.sdk.WrapOperation(config.sdk.Workload().FederatedCredential().Delete(ctx, req))
	if err != nil {
		return diag.FromErr(err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = op.Response()
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
