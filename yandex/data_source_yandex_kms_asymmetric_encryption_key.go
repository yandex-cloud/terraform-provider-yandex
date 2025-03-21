package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	kms "github.com/yandex-cloud/go-genproto/yandex/cloud/kms/v1/asymmetricencryption"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func dataSourceYandexKMSAsymmetricEncryptionKey() *schema.Resource {
	return &schema.Resource{
		Description: "Get data from Yandex KMS asymmetric encryption key.",

		ReadContext: dataSourceYandexKMSAsymmetricEncryptionKeyRead,

		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Optional:    true,
			},

			"description": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["description"],
				Optional:    true,
			},

			"folder_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["folder_id"],
				Computed:    true,
				Optional:    true,
				ForceNew:    true,
			},

			"labels": {
				Type:        schema.TypeMap,
				Description: common.ResourceDescriptions["labels"],
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},

			"encryption_algorithm": {
				Type:         schema.TypeString,
				Description:  resourceYandexKMSAsymmetricEncryptionKey().Schema["encryption_algorithm"].Description,
				Default:      "RSA_2048_ENC_OAEP_SHA_256",
				Optional:     true,
				ValidateFunc: validateParsableValue(parseKmsAsymmetricEncryptionAlgorithm),
			},

			"deletion_protection": {
				Type:        schema.TypeBool,
				Description: common.ResourceDescriptions["deletion_protection"],
				Default:     false,
				Optional:    true,
			},

			"status": {
				Type:        schema.TypeString,
				Description: resourceYandexKMSAsymmetricEncryptionKey().Schema["status"].Description,
				Computed:    true,
			},

			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},

			"asymmetric_encryption_key_id": {
				Type:         schema.TypeString,
				Description:  "Asymmetric encryption key ID.",
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 50),
			},
		},
	}
}

func dataSourceYandexKMSAsymmetricEncryptionKeyRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	req := &kms.GetAsymmetricEncryptionKeyRequest{
		KeyId: data.Get("asymmetric_encryption_key_id").(string),
	}

	md := new(metadata.MD)
	resp, err := config.sdk.KMSAsymmetricEncryption().AsymmetricEncryptionKey().Get(ctx, req, grpc.Header(md))

	if err != nil {
		return diag.FromErr(handleNotFoundError(err, data, fmt.Sprintf("kms asymmetric encryption key %q", data.Get("asymmetric_encryption_key_id").(string))))
	}
	data.SetId(resp.Id)

	createdAt := getTimestamp(resp.GetCreatedAt())

	data.Set("created_at", createdAt)
	data.Set("encryption_algorithm", resp.GetEncryptionAlgorithm().String())
	data.Set("deletion_protection", resp.GetDeletionProtection())
	data.Set("description", resp.GetDescription())
	data.Set("folder_id", resp.GetFolderId())
	if err := data.Set("labels", resp.GetLabels()); err != nil {
		return diag.FromErr(err)
	}
	data.Set("name", resp.GetName())
	data.Set("status", resp.GetStatus().String())
	data.Set("asymmetric_encryption_key_id", resp.GetId())

	return nil

}
