package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/encryption"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
)

func resourceYandexIAMServiceAccountAPIKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexIAMServiceAccountAPIKeyCreate,
		Read:   resourceYandexIAMServiceAccountAPIKeyRead,
		Delete: resourceYandexIAMServiceAccountAPIKeyDelete,

		Schema: map[string]*schema.Schema{
			"service_account_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			// There is no Update method for IAM API Key resource,
			// so "description" attr set as 'ForceNew:true'
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"pgp_key": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"secret_key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"key_fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"encrypted_secret_key": {
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

func resourceYandexIAMServiceAccountAPIKeyCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	resp, err := config.sdk.IAM().ApiKey().Create(ctx, &iam.CreateApiKeyRequest{
		ServiceAccountId: d.Get("service_account_id").(string),
		Description:      d.Get("description").(string),
	})
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

	return resourceYandexIAMServiceAccountAPIKeyRead(d, meta)
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

	createdAt, err := getTimestamp(ak.CreatedAt)
	if err != nil {
		return err
	}

	d.Set("service_account_id", ak.ServiceAccountId)
	d.Set("created_at", createdAt)
	d.Set("description", ak.Description)

	return nil
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

	d.SetId("")
	return nil
}
