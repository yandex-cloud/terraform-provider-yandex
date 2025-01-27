package yandex

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1/workload/oidc"
)

const yandexIAMWorkloadIdentityOidcFederationDefaultTimeout = 1 * time.Minute

func resourceYandexIAMWorkloadIdentityOidcFederation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceYandexIAMWorkloadIdentityOidcFederationCreate,
		ReadContext:   resourceYandexIAMWorkloadIdentityOidcFederationRead,
		UpdateContext: resourceYandexIAMWorkloadIdentityOidcFederationUpdate,
		DeleteContext: resourceYandexIAMWorkloadIdentityOidcFederationDelete,

		Description: "Allows management of [Yandex Cloud IAM workload identity OIDC federations](https://yandex.cloud/docs/iam/concepts/workload-identity).",

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexIAMWorkloadIdentityOidcFederationDefaultTimeout),
			Read:   schema.DefaultTimeout(yandexIAMWorkloadIdentityOidcFederationDefaultTimeout),
			Update: schema.DefaultTimeout(yandexIAMWorkloadIdentityOidcFederationDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexIAMWorkloadIdentityOidcFederationDefaultTimeout),
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"federation_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "Id of the OIDC workload identity federation.",
			},

			"folder_id": {
				Type:        schema.TypeString,
				ForceNew:    true,
				Computed:    true,
				Optional:    true,
				Description: "Id of the folder that the OIDC workload identity federation belongs to.",
			},

			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the OIDC workload identity federation. The name is unique within the folder.",
			},

			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the OIDC workload identity federation.",
			},

			"disabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Disabled flag.",
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
				Optional:    true,
				Description: "List of trusted values for aud claim.",
			},

			"issuer": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Issuer identifier of the external IdP server to be used for authentication.",
			},

			"jwks_url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "URL reference to trusted keys in format of JSON Web Key Set.",
			},

			"labels": {
				Type: schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Set:         schema.HashString,
				Optional:    true,
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

func resourceYandexIAMWorkloadIdentityOidcFederationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	folderId, err := getFolderID(d, config)
	if err != nil {
		return diag.FromErr(err)
	}

	audiences := expandStringSlice(d.Get("audiences").([]interface{}))

	labels := expandStringStringMap(d.Get("labels").(map[string]interface{}))

	req := &oidc.CreateFederationRequest{
		FolderId:    folderId,
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Disabled:    d.Get("disabled").(bool),
		Audiences:   audiences,
		Issuer:      d.Get("issuer").(string),
		JwksUrl:     d.Get("jwks_url").(string),
		Labels:      labels,
	}

	op, err := config.sdk.WrapOperation(config.sdk.WorkloadOidc().Federation().Create(ctx, req))

	if err != nil {
		return diag.Errorf("error while requesting API to create workload identity OIDC federation: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return diag.Errorf("error while getting operation metadata of create workload identity OIDC federation: %s", err)
	}

	md, ok := protoMetadata.(*oidc.CreateFederationMetadata)
	if !ok {
		return diag.Errorf("could not get workload identity OIDC federation Id from create operation metadata")
	}

	d.SetId(md.FederationId)

	err = op.Wait(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceYandexIAMWorkloadIdentityOidcFederationRead(ctx, d, meta)
}

func resourceYandexIAMWorkloadIdentityOidcFederationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	req := &oidc.GetFederationRequest{
		FederationId: d.Id(),
	}

	resp, err := config.sdk.WorkloadOidc().Federation().Get(ctx, req)

	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("audiences", resp.GetAudiences()); err != nil {
		log.Printf("[ERROR] failed set field audiences: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("created_at", getTimestamp(resp.GetCreatedAt())); err != nil {
		log.Printf("[ERROR] failed set field created_at: %s", err)
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
	if err := d.Set("federation_id", resp.GetId()); err != nil {
		log.Printf("[ERROR] failed set field federation_id: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("folder_id", resp.GetFolderId()); err != nil {
		log.Printf("[ERROR] failed set field folder_id: %s", err)
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
	if err := d.Set("name", resp.GetName()); err != nil {
		log.Printf("[ERROR] failed set field name: %s", err)
		return diag.FromErr(err)
	}

	return nil
}

func resourceYandexIAMWorkloadIdentityOidcFederationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	req := &oidc.UpdateFederationRequest{
		FederationId: d.Id(),
		UpdateMask:   &field_mask.FieldMask{},
	}

	d.Partial(true)

	if d.HasChange("name") {
		req.Name = d.Get("name").(string)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "name")
	}

	if d.HasChange("description") {
		req.Description = d.Get("description").(string)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "description")
	}

	if d.HasChange("disabled") {
		req.Disabled = d.Get("disabled").(bool)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "disabled")
	}

	if d.HasChange("audiences") {
		audiencesProp := expandStringSlice(d.Get("audiences").([]interface{}))
		req.Audiences = audiencesProp
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "audiences")
	}

	if d.HasChange("jwks_url") {
		req.JwksUrl = d.Get("jwks_url").(string)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "jwks_url")
	}

	if d.HasChange("labels") {
		labelsProp := expandStringStringMap(d.Get("labels").(map[string]interface{}))
		req.Labels = labelsProp
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "labels")
	}

	if len(req.UpdateMask.Paths) > 0 {
		op, err := config.sdk.WrapOperation(config.sdk.WorkloadOidc().Federation().Update(ctx, req))
		if err != nil {
			return diag.Errorf("error while requesting API to update workload identity federation: %s", err)
		}

		err = op.Wait(ctx)
		if err != nil {
			return diag.Errorf("error while waiting operation to update workload identity federation: %s", err)

		}
		if _, err := op.Response(); err != nil {
			return diag.Errorf("workload identity federation update failed: %s", err)
		}
	}

	d.Partial(false)

	return resourceYandexIAMWorkloadIdentityOidcFederationRead(ctx, d, meta)
}

func resourceYandexIAMWorkloadIdentityOidcFederationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	req := &oidc.DeleteFederationRequest{
		FederationId: d.Id(),
	}

	op, err := config.sdk.WrapOperation(config.sdk.WorkloadOidc().Federation().Delete(ctx, req))
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
