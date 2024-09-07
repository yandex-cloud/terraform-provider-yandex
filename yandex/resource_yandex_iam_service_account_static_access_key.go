package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/encryption"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1/awscompatibility"
)

func resourceYandexIAMServiceAccountStaticAccessKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexIAMServiceAccountStaticAccessKeyCreate,
		Read:   resourceYandexIAMServiceAccountStaticAccessKeyRead,
		Update: resourceYandexIAMServiceAccountStaticAccessKeyUpdate,
		Delete: resourceYandexIAMServiceAccountStaticAccessKeyDelete,

		Schema: ExtendWithOutputToLockbox(map[string]*schema.Schema{
			"service_account_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			// There is no Update method for IAM SA Key resource,
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

			"access_key": {
				Type:     schema.TypeString,
				Computed: true,
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
		}, resourceYandexIAMServiceAccountStaticAccessKeySensitiveAttrs),
	}
}

// `access_key` is not Sensitive but, for convenience, we want to move both keys to the Lockbox secret.
var resourceYandexIAMServiceAccountStaticAccessKeySensitiveAttrs = []string{"secret_key", "access_key"}

func resourceYandexIAMServiceAccountStaticAccessKeyCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	resp, err := config.sdk.IAM().AWSCompatibility().AccessKey().Create(ctx, &awscompatibility.CreateAccessKeyRequest{
		ServiceAccountId: d.Get("service_account_id").(string),
		Description:      d.Get("description").(string),
	})
	if err != nil {
		return fmt.Errorf("error creating service account key: %s", err)
	}

	d.SetId(resp.AccessKey.Id)
	// Data only available on create.
	if v, ok := d.GetOk("pgp_key"); ok {
		encryptionKey, err := encryption.RetrieveGPGKey(v.(string))
		if err != nil {
			return err
		}

		fingerprint, encrypted, err := encryption.EncryptValue(encryptionKey, resp.Secret, "Yandex Service Account Static Access Key")
		if err != nil {
			return err
		}

		d.Set("key_fingerprint", fingerprint)
		d.Set("encrypted_secret_key", encrypted)
	} else {
		d.Set("secret_key", resp.Secret)
	}

	err = resourceYandexIAMServiceAccountStaticAccessKeyRead(d, meta)
	if err != nil {
		return err
	}

	return ManageOutputToLockbox(ctx, d, config, resourceYandexIAMServiceAccountStaticAccessKeySensitiveAttrs)
}

func resourceYandexIAMServiceAccountStaticAccessKeyRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutRead))
	defer cancel()

	sak, err := config.sdk.IAM().AWSCompatibility().AccessKey().Get(ctx, &awscompatibility.GetAccessKeyRequest{
		AccessKeyId: d.Id(),
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Service Account Static Access Key %q", d.Id()))
	}

	d.Set("service_account_id", sak.ServiceAccountId)
	d.Set("created_at", getTimestamp(sak.CreatedAt))
	d.Set("description", sak.Description)
	d.Set("access_key", sak.KeyId)

	return nil
}

// The update method was added because ExtendWithOutputToLockbox adds a new attribute output_to_lockbox that can change.
// Changes in output_to_lockbox are handled in ManageOutputToLockbox.
func resourceYandexIAMServiceAccountStaticAccessKeyUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	err := resourceYandexIAMServiceAccountStaticAccessKeyRead(d, meta)
	if err != nil {
		return err
	}

	return ManageOutputToLockbox(ctx, d, config, resourceYandexIAMServiceAccountStaticAccessKeySensitiveAttrs)
}

func resourceYandexIAMServiceAccountStaticAccessKeyDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutDelete))
	defer cancel()

	_, err := config.sdk.IAM().AWSCompatibility().AccessKey().Delete(ctx, &awscompatibility.DeleteAccessKeyRequest{
		AccessKeyId: d.Id(),
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Service Account Static Access Key %q", d.Id()))
	}

	err = DestroyOutputToLockboxVersion(ctx, d, config)
	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}
