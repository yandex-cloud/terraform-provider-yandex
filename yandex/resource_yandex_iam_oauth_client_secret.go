package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/encryption"
)

func resourceYandexIAMOAuthClientSecret() *schema.Resource {
	return &schema.Resource{
		Description: "Allows management of [Yandex Cloud IAM OAuth client secret](https://yandex.cloud/docs/iam/concepts/authorization/oauth-client-secret). The OAuth client secret is used for OAuth 2.0 client authentication.",
		Create:      resourceYandexIAMOAuthClientSecretCreate,
		Read:        resourceYandexIAMOAuthClientSecretRead,
		Update:      resourceYandexIAMOAuthClientSecretUpdate,
		Delete:      resourceYandexIAMOAuthClientSecretDelete,

		Schema: ExtendWithOutputToLockbox(map[string]*schema.Schema{
			"oauth_client_id": {
				Type:        schema.TypeString,
				Description: "ID of the OAuth client to create a secret for.",
				Required:    true,
				ForceNew:    true,
			},

			"description": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["description"],
				Optional:    true,
			},

			"pgp_key": {
				Type:        schema.TypeString,
				Description: "An optional PGP key to encrypt the resulting secret value. May either be a base64-encoded public key or a keybase username in the form `keybase:keybaseusername`.",
				Optional:    true,
				ForceNew:    true,
			},

			"secret_value": {
				Type:        schema.TypeString,
				Description: "The secret value. This is only populated when neither `pgp_key` nor `output_to_lockbox` are provided.",
				Computed:    true,
				Sensitive:   true,
			},

			"masked_secret": {
				Type:        schema.TypeString,
				Description: "The masked value of the OAuth client secret.",
				Computed:    true,
			},

			"key_fingerprint": {
				Type:        schema.TypeString,
				Description: "The fingerprint of the PGP key used to encrypt the secret value. This is only populated when `pgp_key` is supplied.",
				Computed:    true,
			},

			"encrypted_secret_value": {
				Type:        schema.TypeString,
				Description: "The encrypted secret value, base64 encoded. This is only populated when `pgp_key` is supplied.",
				Computed:    true,
			},

			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},
		}, resourceYandexIAMOAuthClientSecretSensitiveAttrs),
	}
}

var resourceYandexIAMOAuthClientSecretSensitiveAttrs = []string{"secret_value"}

func resourceYandexIAMOAuthClientSecretCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	req := &iam.CreateOAuthClientSecretRequest{
		OauthClientId: d.Get("oauth_client_id").(string),
		Description:   d.Get("description").(string),
	}

	op, err := config.sdk.WrapOperation(config.sdk.IAM().OAuthClientSecret().Create(ctx, req))
	if err != nil {
		return fmt.Errorf("error creating OAuth client secret: %s", err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while waiting operation to create OAuth client secret: %s", err)
	}

	resp, err := op.Response()
	if err != nil {
		return fmt.Errorf("OAuth client secret creation failed: %s", err)
	}

	createResponse := resp.(*iam.CreateOAuthClientSecretResponse)
	oauthClientSecret := createResponse.OauthClientSecret
	secretValue := createResponse.SecretValue

	d.SetId(oauthClientSecret.Id)

	// Data only available on create.
	if v, ok := d.GetOk("pgp_key"); ok {
		encryptionKey, err := encryption.RetrieveGPGKey(v.(string))
		if err != nil {
			return err
		}

		fingerprint, encrypted, err := encryption.EncryptValue(encryptionKey, secretValue, "Yandex OAuth Client Secret")
		if err != nil {
			return err
		}

		d.Set("key_fingerprint", fingerprint)
		d.Set("encrypted_secret_value", encrypted)
	} else {
		d.Set("secret_value", secretValue)
	}

	err = resourceYandexIAMOAuthClientSecretRead(d, meta)
	if err != nil {
		return err
	}

	return ManageOutputToLockbox(ctx, d, config, resourceYandexIAMOAuthClientSecretSensitiveAttrs)
}

func resourceYandexIAMOAuthClientSecretRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutRead))
	defer cancel()

	secret, err := config.sdk.IAM().OAuthClientSecret().Get(ctx, &iam.GetOAuthClientSecretRequest{
		OauthClientSecretId: d.Id(),
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("OAuth Client Secret %q", d.Id()))
	}

	d.Set("oauth_client_id", secret.OauthClientId)
	d.Set("created_at", getTimestamp(secret.CreatedAt))
	d.Set("description", secret.Description)
	d.Set("masked_secret", secret.MaskedSecret)

	return nil
}

// The update method was added because ExtendWithOutputToLockbox adds a new attribute output_to_lockbox that can change.
// Changes in output_to_lockbox are handled in ManageOutputToLockbox.
func resourceYandexIAMOAuthClientSecretUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	err := resourceYandexIAMOAuthClientSecretRead(d, meta)
	if err != nil {
		return err
	}

	return ManageOutputToLockbox(ctx, d, config, resourceYandexIAMOAuthClientSecretSensitiveAttrs)
}

func resourceYandexIAMOAuthClientSecretDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.IAM().OAuthClientSecret().Delete(ctx, &iam.DeleteOAuthClientSecretRequest{
		OauthClientSecretId: d.Id(),
	}))
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("OAuth Client Secret %q", d.Id()))
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while waiting operation to delete OAuth client secret: %s", err)
	}

	err = DestroyOutputToLockboxVersion(ctx, d, config)
	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}
