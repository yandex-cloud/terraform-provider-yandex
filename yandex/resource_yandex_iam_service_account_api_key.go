package yandex

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/encryption"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
)

func resourceYandexIAMServiceAccountAPIKey() *schema.Resource {
	return &schema.Resource{
		Description: "Allows management of a [Yandex Cloud IAM service account API key](https://yandex.cloud/docs/iam/concepts/authorization/api-key). The API key is a private key used for simplified authorization in the Yandex Cloud API. API keys are only used for [service accounts](https://yandex.cloud/docs/iam/concepts/users/service-accounts).\n\nAPI keys do not expire. This means that this authentication method is simpler, but less secure. Use it if you can't automatically request an [IAM token](https://yandex.cloud/docs/iam/concepts/authorization/iam-token).",
		Create:      resourceYandexIAMServiceAccountAPIKeyCreate,
		Read:        resourceYandexIAMServiceAccountAPIKeyRead,
		Update:      resourceYandexIAMServiceAccountAPIKeyUpdate,
		Delete:      resourceYandexIAMServiceAccountAPIKeyDelete,

		Schema: ExtendWithOutputToLockbox(map[string]*schema.Schema{
			"service_account_id": {
				Type:        schema.TypeString,
				Description: "ID of the service account to an API key for.",
				Required:    true,
				ForceNew:    true,
			},

			"description": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["description"],
				Optional:    true,
			},

			"scopes": {
				Type:        schema.TypeList,
				Description: "The list of scopes of the key.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
				Computed: true,
			},

			"scope": {
				Type:        schema.TypeString,
				Description: "The scope of the key.",
				Optional:    true,
				Deprecated:  "Attribute `scope` deprecated and will be removed in the next major version of the provider. Use attribute `scopes` instead.",
			},

			"expires_at": {
				Type:        schema.TypeString,
				Description: "The key will be no longer valid after expiration timestamp.",
				Optional:    true,
			},

			"pgp_key": {
				Type:        schema.TypeString,
				Description: "An optional PGP key to encrypt the resulting secret key material. May either be a base64-encoded public key or a keybase username in the form `keybase:keybaseusername`.",
				Optional:    true,
				ForceNew:    true,
			},

			"secret_key": {
				Type:        schema.TypeString,
				Description: "The secret key. This is only populated when neither `pgp_key` nor `output_to_lockbox` are provided.",
				Computed:    true,
				Sensitive:   true,
			},

			"key_fingerprint": {
				Type:        schema.TypeString,
				Description: "The fingerprint of the PGP key used to encrypt the secret key. This is only populated when `pgp_key` is supplied.",
				Computed:    true,
			},

			"encrypted_secret_key": {
				Type:        schema.TypeString,
				Description: "The encrypted secret key, base64 encoded. This is only populated when `pgp_key` is supplied.",
				Computed:    true,
			},

			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},
		}, resourceYandexIAMServiceAccountAPIKeySensitiveAttrs),
	}
}

var resourceYandexIAMServiceAccountAPIKeySensitiveAttrs = []string{"secret_key"}

func resourceYandexIAMServiceAccountAPIKeyCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	serviceAccountID := d.Get("service_account_id").(string)
	req := iam.CreateApiKeyRequest{
		ServiceAccountId: serviceAccountID,
		Description:      d.Get("description").(string),
	}

	if v, ok := d.GetOk("scopes"); ok {
		req.SetScopes(expandStringSlice(v.([]interface{})))
	}

	if v, ok := d.GetOk("scope"); ok {
		req.SetScope(v.(string))
	}

	if v, ok := d.GetOk("expires_at"); ok {
		expiresAt, err := parseTimestamp(v.(string))
		if err != nil {
			return fmt.Errorf("Error during parsing field expires_at while creating API Key for Service Account %s: %s", serviceAccountID, err)
		}
		req.SetExpiresAt(expiresAt)
	}

	resp, err := config.sdk.IAM().ApiKey().Create(ctx, &req)
	if err != nil {
		return fmt.Errorf("error creating api key: %s", err)
	}

	d.SetId(resp.ApiKey.Id)
	// Data only available on create.
	if v, ok := d.GetOk("pgp_key"); ok {
		encryptionKey, err := encryption.RetrieveGPGKey(v.(string))
		if err != nil {
			return err
		}

		fingerprint, encrypted, err := encryption.EncryptValue(encryptionKey, resp.Secret, "Yandex Service Account API Key")
		if err != nil {
			return err
		}

		d.Set("key_fingerprint", fingerprint)
		d.Set("encrypted_secret_key", encrypted)
	} else {
		d.Set("secret_key", resp.Secret)
	}

	err = resourceYandexIAMServiceAccountAPIKeyRead(d, meta)
	if err != nil {
		return err
	}

	return ManageOutputToLockbox(ctx, d, config, resourceYandexIAMServiceAccountAPIKeySensitiveAttrs)
}

func resourceYandexIAMServiceAccountAPIKeyRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutRead))
	defer cancel()

	ak, err := config.sdk.IAM().ApiKey().Get(ctx, &iam.GetApiKeyRequest{
		ApiKeyId: d.Id(),
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Api Key %q", d.Id()))
	}

	d.Set("service_account_id", ak.ServiceAccountId)
	d.Set("created_at", getTimestamp(ak.CreatedAt))
	d.Set("description", ak.Description)

	if ak.Scopes != nil {
		d.Set("scopes", ak.Scopes)
	}

	if ak.Scope != "" {
		d.Set("scope", ak.Scope)
	}

	if ak.ExpiresAt != nil {
		d.Set("expires_at", getTimestamp(ak.ExpiresAt))
	}

	return nil
}

// The update method was added because ExtendWithOutputToLockbox adds a new attribute output_to_lockbox that can change.
// Changes in output_to_lockbox are handled in ManageOutputToLockbox.
func resourceYandexIAMServiceAccountAPIKeyUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	apiKeyId := d.Id()
	log.Printf("[INFO] Updating API key %q", apiKeyId)

	d.Partial(true)

	req := &iam.UpdateApiKeyRequest{
		ApiKeyId:   apiKeyId,
		UpdateMask: &field_mask.FieldMask{},
	}

	if d.HasChange("description") {
		req.SetDescription(d.Get("description").(string))
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "description")
	}

	if d.HasChange("scopes") {
		scopes := expandStringSlice(d.Get("scopes").([]interface{}))
		req.SetScopes(scopes)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "scopes")
	}

	if d.HasChange("expires_at") {
		expiresAt, err := parseTimestamp(d.Get("expires_at").(string))
		if err != nil {
			return fmt.Errorf("Error during parsing field expires_at while updating API key %s: %s", apiKeyId, err)
		}
		req.SetExpiresAt(expiresAt)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "expires_at")
	}

	if len(req.UpdateMask.Paths) > 0 {
		op, err := config.sdk.WrapOperation(config.sdk.IAM().ApiKey().Update(ctx, req))
		if err != nil {
			return fmt.Errorf("error while requesting API to update API Key %s: %s", apiKeyId, err)
		}

		err = op.Wait(ctx)
		if err != nil {
			return fmt.Errorf("error while waiting operation to update API key %s: %s", apiKeyId, err)

		}

		if _, err := op.Response(); err != nil {
			return fmt.Errorf("API Key %s update failed: %s", apiKeyId, err)
		}
	}

	d.Partial(false)

	err := resourceYandexIAMServiceAccountAPIKeyRead(d, meta)
	if err != nil {
		return err
	}

	return ManageOutputToLockbox(ctx, d, config, resourceYandexIAMServiceAccountAPIKeySensitiveAttrs)
}

func resourceYandexIAMServiceAccountAPIKeyDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutDelete))
	defer cancel()

	_, err := config.sdk.IAM().ApiKey().Delete(ctx, &iam.DeleteApiKeyRequest{
		ApiKeyId: d.Id(),
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Api Key %q", d.Id()))
	}

	err = DestroyOutputToLockboxVersion(ctx, d, config)
	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}
