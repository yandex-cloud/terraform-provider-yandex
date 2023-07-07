package yandex

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/kms/v1/asymmetricsignature"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func dataSourceYandexKMSAsymmetricSignatureKey() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceYandexKMSAsymmetricSignatureKeyRead,

		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},

			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"signature_algorithm": {
				Type:         schema.TypeString,
				Default:      "RSA_2048_SIGN_PSS_SHA_256",
				Optional:     true,
				ValidateFunc: validateParsableValue(parseKmsAsymmetricSignatureAlgorithm),
			},

			"deletion_protection": {
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"asymmetric_signature_key_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 50),
			},
		},
	}
}

func dataSourceYandexKMSAsymmetricSignatureKeyRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	req := &kms.GetAsymmetricSignatureKeyRequest{
		KeyId: data.Get("asymmetric_signature_key_id").(string),
	}

	md := new(metadata.MD)
	resp, err := config.sdk.KMSAsymmetricSignature().AsymmetricSignatureKey().Get(ctx, req, grpc.Header(md))

	if err != nil {
		return diag.FromErr(handleNotFoundError(err, data, fmt.Sprintf("kms asymmetric signature key %q", data.Get("asymmetric_signature_key_id").(string))))
	}
	data.SetId(resp.Id)

	createdAt := getTimestamp(resp.GetCreatedAt())

	data.Set("created_at", createdAt)
	data.Set("signature_algorithm", resp.GetSignatureAlgorithm().String())
	data.Set("deletion_protection", resp.GetDeletionProtection())
	data.Set("description", resp.GetDescription())
	data.Set("folder_id", resp.GetFolderId())
	if err := data.Set("labels", resp.GetLabels()); err != nil {
		return diag.FromErr(err)
	}
	data.Set("name", resp.GetName())
	data.Set("status", resp.GetStatus().String())
	data.Set("asymmetric_signature_key_id", resp.GetId())

	return nil

}
