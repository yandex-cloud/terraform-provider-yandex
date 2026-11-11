package pluginsdk

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/kms/v1"
	kmssdk "github.com/yandex-cloud/go-sdk/services/kms/v1"
	sdkresolversv2 "github.com/yandex-cloud/go-sdk/v2/pkg/sdkresolvers"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func DataSourceYandexKMSSymmetricKey() *schema.Resource {
	return &schema.Resource{
		Description: "Get data from Yandex KMS symmetric key.",

		ReadContext: dataSourceYandexKMSSymmetricKeyRead,

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

			"default_algorithm": {
				Type:         schema.TypeString,
				Description:  "Encryption algorithm to be used with a new key version, generated with the next rotation. The default value is `AES_128`.",
				Default:      "AES_128",
				Optional:     true,
				ValidateFunc: sdkcommon.ValidateParsableValue(parseKmsDefaultAlgorithm),
			},

			"deletion_protection": {
				Type:        schema.TypeBool,
				Description: common.ResourceDescriptions["deletion_protection"],
				Default:     false,
				Optional:    true,
			},

			"rotation_period": {
				Type:             schema.TypeString,
				Description:      "Interval between automatic rotations. To disable automatic rotation, omit this parameter.",
				Optional:         true,
				ValidateFunc:     sdkcommon.ValidateParsableValue(sdkcommon.ParsePositiveDuration),
				DiffSuppressFunc: sdkcommon.ShouldSuppressDiffForTimeDuration,
			},

			"status": {
				Type:        schema.TypeString,
				Description: "The status of the key.",
				Computed:    true,
			},

			"rotated_at": {
				Type:        schema.TypeString,
				Description: "Last rotation timestamp of the key.",
				Computed:    true,
			},

			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},
			"symmetric_key_id": {
				Type:         schema.TypeString,
				Description:  "The symmetric key ID.",
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 50),
			},
		},
	}
}

func dataSourceYandexKMSSymmetricKeyRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*providerconfig.Config)
	client := kmssdk.NewSymmetricKeyClient(config.SDK)

	err := sdkcommon.CheckOneOf(data, "symmetric_key_id", "name")
	if err != nil {
		return diag.FromErr(err)
	}
	keyID := data.Get("symmetric_key_id").(string)

	_, keyNameOk := data.GetOk("name")
	if keyNameOk {
		keyID, err = sdkcommon.ResolveObjectIDV2(config.Context(), config, data, func(name string, opts ...sdkresolversv2.ResolveOption) sdkresolversv2.Resolver {
			return kmssdk.SymmetricKeyResolver(name, client, opts...)
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	req := &kms.GetSymmetricKeyRequest{
		KeyId: keyID,
	}

	md := new(metadata.MD)
	resp, err := client.Get(ctx, req, grpc.Header(md))

	if err != nil {
		return diag.FromErr(sdkcommon.HandleNotFoundError(err, data, fmt.Sprintf("kms symmetric key %q", data.Get("symmetric_key_id").(string))))
	}
	data.SetId(resp.Id)

	createdAt := sdkcommon.GetTimestamp(resp.GetCreatedAt())
	rotatedAt := sdkcommon.GetTimestamp(resp.GetRotatedAt())

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
