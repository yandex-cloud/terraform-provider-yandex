package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/encryption"
	"google.golang.org/genproto/protobuf/field_mask"
)

func resourceYandexIAMServiceAccountKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexIAMServiceAccountKeyCreate,
		Read:   resourceYandexIAMServiceAccountKeyRead,
		Update: resourceYandexIAMServiceAccountKeyUpdate,
		Delete: resourceYandexIAMServiceAccountKeyDelete,

		Schema: map[string]*schema.Schema{
			"service_account_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"format": {
				Type:         schema.TypeString,
				Default:      "PEM_FILE",
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateParsableValue(parseIamKeyFormat),
			},

			"key_algorithm": {
				Type:         schema.TypeString,
				Default:      "RSA_2048",
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateParsableValue(parseIamKeyAlgorithm),
			},

			"pgp_key": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"public_key": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"private_key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"key_fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"encrypted_private_key": {
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

func resourceYandexIAMServiceAccountKeyCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	format, err := parseIamKeyFormat(d.Get("format").(string))
	if err != nil {
		return err
	}

	algorithm, err := parseIamKeyAlgorithm(d.Get("key_algorithm").(string))
	if err != nil {
		return err
	}

	resp, err := config.sdk.IAM().Key().Create(ctx, &iam.CreateKeyRequest{
		ServiceAccountId: d.Get("service_account_id").(string),
		Description:      d.Get("description").(string),
		Format:           format,
		KeyAlgorithm:     algorithm,
	})
	if err != nil {
		return fmt.Errorf("error creating service account key: %s", err)
	}

	d.SetId(resp.Key.Id)
	// Data only available on create.
	if v, ok := d.GetOk("pgp_key"); ok {
		encryptionKey, err := encryption.RetrieveGPGKey(v.(string))
		if err != nil {
			return err
		}

		fingerprint, encrypted, err := encryption.EncryptValue(encryptionKey, resp.PrivateKey, "Yandex Service Account Key")
		if err != nil {
			return err
		}

		d.Set("key_fingerprint", fingerprint)
		d.Set("encrypted_private_key", encrypted)
	} else {
		d.Set("private_key", resp.PrivateKey)
	}

	return resourceYandexIAMServiceAccountKeyRead(d, meta)
}

func resourceYandexIAMServiceAccountKeyRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutRead))
	defer cancel()

	format, err := parseIamKeyFormat(d.Get("format").(string))
	if err != nil {
		return err
	}

	key, err := config.sdk.IAM().Key().Get(ctx, &iam.GetKeyRequest{
		KeyId:  d.Id(),
		Format: format,
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Service Account Key %q", d.Id()))
	}

	d.Set("service_account_id", key.GetServiceAccountId())
	d.Set("created_at", getTimestamp(key.CreatedAt))
	d.Set("description", key.Description)
	d.Set("key_algorithm", iam.Key_Algorithm_name[int32(key.KeyAlgorithm)])
	d.Set("public_key", key.PublicKey)

	return nil
}

func resourceYandexIAMServiceAccountKeyUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	req := &iam.UpdateKeyRequest{
		KeyId:       d.Id(),
		Description: d.Get("description").(string),
	}

	var updatedFields []string
	fields := []string{"description"}
	for _, field := range fields {
		if d.HasChange(field) {
			updatedFields = append(updatedFields, field)
		}
	}

	if len(updatedFields) != 0 {
		req.UpdateMask = &field_mask.FieldMask{Paths: updatedFields}
		_, err := config.sdk.IAM().Key().Update(ctx, req)
		if err != nil {
			return handleNotFoundError(err, d, fmt.Sprintf("Service Account Key %q", d.Id()))
		}
	}

	return resourceYandexIAMServiceAccountKeyRead(d, meta)
}

func resourceYandexIAMServiceAccountKeyDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutDelete))
	defer cancel()

	_, err := config.sdk.IAM().Key().Delete(ctx, &iam.DeleteKeyRequest{
		KeyId: d.Id(),
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Service Account Key %q", d.Id()))
	}

	d.SetId("")
	return nil
}
