package yandex

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/kms/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func dataSourceYandexKMSSymmetricKey() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceYandexKMSSymmetricKeyRead,

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

			"default_algorithm": {
				Type:         schema.TypeString,
				Default:      "AES_128",
				Optional:     true,
				ValidateFunc: validateParsableValue(parseKmsDefaultAlgorithm),
			},

			"deletion_protection": {
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},

			"rotation_period": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validateParsableValue(parsePositiveDuration),
				DiffSuppressFunc: shouldSuppressDiffForTimeDuration,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"rotated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"symmetric_key_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 50),
			},
		},
	}
}

func dataSourceYandexKMSSymmetricKeyRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	err := checkOneOf(data, "symmetric_key_id", "name")
	if err != nil {
		return diag.FromErr(err)
	}
	keyID := data.Get("symmetric_key_id").(string)

	_, keyNameOk := data.GetOk("name")
	if keyNameOk {
		keyID, err = resolveObjectID(config.Context(), config, data, sdkresolvers.SymmetricKeyResolver)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	req := &kms.GetSymmetricKeyRequest{
		KeyId: keyID,
	}

	md := new(metadata.MD)
	resp, err := config.sdk.KMS().SymmetricKey().Get(ctx, req, grpc.Header(md))

	if err != nil {
		return diag.FromErr(handleNotFoundError(err, data, fmt.Sprintf("kms symmetric key %q", data.Get("symmetric_key_id").(string))))
	}
	data.SetId(resp.Id)

	createdAt := getTimestamp(resp.GetCreatedAt())
	rotatedAt := getTimestamp(resp.GetRotatedAt())

	data.Set("created_at", createdAt)
	data.Set("default_algorithm", resp.GetDefaultAlgorithm().String())
	data.Set("deletion_protection", resp.GetDeletionProtection())
	data.Set("description", resp.GetDescription())
	data.Set("folder_id", resp.GetFolderId())
	if err := data.Set("labels", resp.GetLabels()); err != nil {
		return diag.FromErr(err)
	}
	data.Set("name", resp.GetName())
	data.Set("rotated_at", rotatedAt)
	data.Set("status", resp.GetStatus().String())
	data.Set("symmetric_key_id", resp.GetId())

	return nil

}
